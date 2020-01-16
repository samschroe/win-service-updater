package updater

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path"
)

// Infoer interface used to make testing easier
type Infoer interface {
	ParseWYS(string, Args) (ConfigWYS, error)
	ParseWYC(string) (ConfigIUC, error)
}

// Info struct for the Infoer interface
type Info struct{}

// Handler is the "main" called by cmd/main.go
func Handler() int {
	args, err := ParseArgs(os.Args)
	if nil != err {
		if args.Debug {
			fmt.Println(err.Error())
		}
		LogErrorMsg(args, err.Error())
		LogOutputInfoMsg(args, err.Error())
	}

	info := Info{}

	// check for updates
	if args.Quickcheck && args.Justcheck {
		if args.Debug {
			fmt.Println("Checking for updates...")
		}

		rc, err := CheckForUpdateHandler(info, args)
		if nil != err {
			if args.Debug {
				fmt.Println(err.Error())
			}
			LogErrorMsg(args, err.Error())
			LogOutputInfoMsg(args, err.Error())
		}

		if args.Debug {
			switch rc {
			case EXIT_NO_UPDATE:
				fmt.Println("No update available")
			case EXIT_UPDATE_AVALIABLE:
				fmt.Println("Update available")
			}
		}

		return rc
	}

	// update
	if args.Fromservice {
		if args.Debug {
			fmt.Println("Updating...")
		}

		rc, err := UpdateHandler(info, (args))
		if nil != err {
			if args.Debug {
				fmt.Println(err.Error())
			}
			LogErrorMsg(args, err.Error())
			LogOutputInfoMsg(args, err.Error())
		}

		if args.Debug && rc == 0 {
			fmt.Println("Update successful")
		}
		return rc
	}

	return EXIT_ERROR
}

// UpdateHandler performs the update. Returns int exit code and error.
func UpdateHandler(infoer Infoer, args Args) (int, error) {
	instDir := GetExeDir()

	tmpDir, err := CreateTempDir()
	if nil != err {
		err = fmt.Errorf("no temp dir; %w", err)
		return EXIT_ERROR, err
	}
	defer os.RemoveAll(tmpDir)

	// parse the WYC file for get update site, installed version, etc.
	iuc, err := infoer.ParseWYC(args.Cdata)
	if nil != err {
		err = fmt.Errorf("error reading %s; %w", args.Cdata, err)
		return EXIT_ERROR, err
	}

	// download the wys file (contains details about the availiable update)
	fp := fmt.Sprintf("%s/wys", tmpDir)
	urls := iuc.GetWYSURLs(args)
	err = DownloadFile(urls, fp)
	if nil != err {
		return EXIT_ERROR, err
	}

	// parse the WYS file (contains the version number of the update and the link to the update)
	wys, err := infoer.ParseWYS(fp, args)
	if nil != err {
		err = fmt.Errorf("error reading wys file (%s); %w", fp, err)
		return EXIT_ERROR, err
	}

	// fmt.Println("installed ", string(iuc.IucInstalledVersion.Value))
	// fmt.Println("new ", wys.VersionToUpdate)

	// download WYU (this is the archive with the updated files)
	fp = fmt.Sprintf("%s/wyu", tmpDir)
	urls = wys.GetWYUURLs(args)
	err = DownloadFile(urls, fp)
	if nil != err {
		return EXIT_ERROR, err
	}

	if iuc.IucPublicKey.Value != nil {
		if len(wys.FileSha1) == 0 {
			err = fmt.Errorf("The update is not signed. All updates must be signed in order to be installed.")
			return EXIT_ERROR, err
		}

		// convert the public key from the WYC file to an rsa.PublicKey
		key, err := ParsePublicKey(string(iuc.IucPublicKey.Value))
		var rsa rsa.PublicKey
		rsa.N = key.Modulus
		rsa.E = key.Exponent

		// hash the downloaded WYU file
		sha1hash, err := SHA1Hash(fp)
		if nil != err {
			err = fmt.Errorf("The downloaded file \"%s\" failed the signature validation: %w", fp, err)
			return EXIT_ERROR, err
		}

		// verify the signature of the WYU file (the signed hash is included in the WYS file)
		err = VerifyHash(&rsa, sha1hash, wys.FileSha1)
		if nil != err {
			err = fmt.Errorf("The downloaded file \"%s\" is not signed. %w", fp, err)
			return EXIT_ERROR, err
		}
	}

	// adler32 checksum
	if wys.UpdateFileAdler32 != 0 {
		v := VerifyAdler32Checksum(wys.UpdateFileAdler32, fp)
		if v != true {
			err = fmt.Errorf("The downloaded file \"%s\" failed the Adler32 validation.", fp)
			return EXIT_ERROR, err
		}
	}

	// extract the WYU to tmpDir
	_, files, err := Unzip(fp, tmpDir)
	if nil != err {
		err = fmt.Errorf("error unzipping %s; %w", fp, err)
		return EXIT_ERROR, err
	}

	// get the details of the update
	// the update "config" is "updtdetails.udt"
	// the "files" are the updated files
	udt, updates, err := GetUpdateDetails(files)
	if nil != err {
		return EXIT_ERROR, err
	}

	// backup the existing files that will be overwritten by the update
	backupDir, err := BackupFiles(updates, instDir)
	if nil != err {
		return EXIT_ERROR, err
	}

	// TODO is there a way to clean this up
	err = InstallUpdate(udt, updates, instDir)
	if nil != err {
		err = fmt.Errorf("error applying update; %w", err)
		// TODO rollback should restore client.wyc
		RollbackFiles(backupDir, instDir)
		// start services, best effort
		for _, s := range udt.ServiceToStartAfterUpdate {
			svc := ValueToString(&s)
			_ = StartService(svc)
		}
		return EXIT_ERROR, err
	}

	// we haven't erred, write latest version number and exit
	UpdateWYCWithNewVersionNumber(iuc, args.Cdata, wys.VersionToUpdate)
	return EXIT_SUCCESS, nil
}

// CheckForUpdateHandler checks to see if an update is availible. Returns int
// exit code and error.
func CheckForUpdateHandler(infoer Infoer, args Args) (int, error) {
	// read WYC
	iuc, err := infoer.ParseWYC(args.Cdata)
	if nil != err {
		err = fmt.Errorf("error reading %s; %w", args.Cdata, err)
		return EXIT_ERROR, err
	}

	tmpDir, err := CreateTempDir()
	if nil != err {
		err = fmt.Errorf("no temp dir; %w", err)
		return EXIT_ERROR, err
	}
	defer os.RemoveAll(tmpDir)

	wysTmpFile := path.Join(tmpDir, "wysTemp")
	urls := iuc.GetWYSURLs(args)

	err = DownloadFile(urls, wysTmpFile)
	if nil != err {
		return EXIT_ERROR, err
	}

	wys, err := infoer.ParseWYS(wysTmpFile, args)
	if nil != err {
		err = fmt.Errorf("error reading %s; %w", wysTmpFile, err)
		return EXIT_ERROR, err
	}

	// compare versions
	rc := CompareVersions(string(iuc.IucInstalledVersion.Value), wys.VersionToUpdate)
	switch rc {
	case A_LESS_THAN_B:
		// need update
		err = fmt.Errorf(wys.VersionToUpdate)
		return EXIT_UPDATE_AVALIABLE, err
	case A_EQUAL_TO_B:
		// no update
		err = fmt.Errorf(wys.VersionToUpdate)
		return EXIT_NO_UPDATE, err
	case A_GREATER_THAN_B:
		// no update
		err = fmt.Errorf(string(iuc.IucInstalledVersion.Value))
		return EXIT_NO_UPDATE, err
	default:
		// unknown case
		return EXIT_ERROR, fmt.Errorf("unknown case")
	}
}
