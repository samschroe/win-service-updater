package updater

import (
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Infoer interface used to make testing easier
type Infoer interface {
	ParseWYSFromReader(reader io.ReaderAt, size int64) (wys ConfigWYS, err error)
	ParseWYSFromFilePath(compressedWYSFilePath string, _ Args) (wys ConfigWYS, err error)
	ParseWYC(string) (ConfigIUC, error)
}

// Info struct for the Infoer interface
type Info struct{}

// Handler is the "main" called by cmd/main.go
func Handler() int {
	args, err := ParseArgs(os.Args)
	if err != nil {
		if args.Debug {
			log.Println(err.Error())
		}
		LogErrorMsg(args, err.Error())
		LogOutputInfoMsg(args, err.Error())
		return EXIT_ERROR
	}

	// check for updates
	info := Info{}
	if args.Quickcheck && args.Justcheck {
		// Quickcheck
		if args.Debug {
			log.Println("Quick Check and Just Check checking for Updates...")
		}

		rc, err := CheckForUpdateHandler(info, args)
		if nil != err {
			if args.Debug {
				log.Println(err.Error())
			}
			LogErrorMsg(args, err.Error())
			LogOutputInfoMsg(args, err.Error())
		}

		if args.Debug {
			switch rc {
			case EXIT_NO_UPDATE:
				log.Println("No update available")
			case EXIT_UPDATE_AVALIABLE:
				log.Println("Update available")
			}
		}

		// End Quickcheck
		return rc
	}

	// update
	if args.Fromservice {
		if args.Debug {
			log.Println("Updating...")
		}

		rc, err := UpdateHandler(info, (args))
		if err != nil {
			if args.Debug {
				log.Println(err.Error())
			}
			LogErrorMsg(args, err.Error())
			LogOutputInfoMsg(args, err.Error())
		}

		if args.Debug && rc == 0 {
			log.Println("Update successful")
		}
		return rc
	}

	return EXIT_ERROR
}

// UpdateHandler performs the update. Returns int exit code and error.
func UpdateHandler(infoer Infoer, args Args) (int, error) {
	candidateUpdateReq, err := NewCandidateUpdateRequest(args, infoer)
	if err != nil {
		return EXIT_ERROR, err
	}

	tmpDir, err := CreateTempDir()
	if nil != err {
		err = fmt.Errorf("failed to create temp dir; %w", err)
		return EXIT_ERROR, err
	}
	defer DeleteDirectory(tmpDir)

	// write the contents of the wys file to disk (contains details about the available update)
	wysFilePath := filepath.Join(tmpDir, "wys")
	err = os.WriteFile(wysFilePath, candidateUpdateReq.CandidateWysFileContent.Bytes(), 0644)
	if err != nil {
		err = fmt.Errorf("failed to write WYS file to: %v; %w", wysFilePath, err)
		return EXIT_ERROR, err
	}

	// download WYU (this is the archive with the updated files)
	wys := candidateUpdateReq.ConfigWYS
	wyuFilePath := filepath.Join(tmpDir, "wyu")
	if err := wys.getWyuFile(args, wyuFilePath); err != nil {
		return EXIT_ERROR, err
	}

	iuc := candidateUpdateReq.ConfigIUC
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
		sha1hash, err := GenerateSHA1HashFromFilePath(wyuFilePath)
		if nil != err {
			err = fmt.Errorf("The downloaded file \"%s\" failed the signature validation: %w", wyuFilePath, err)
			return EXIT_ERROR, err
		}

		// verify the signature of the WYU file (the signed hash is included in the WYS file)
		err = VerifyHash(&rsa, sha1hash, wys.FileSha1)
		if nil != err {
			err = fmt.Errorf("The downloaded file \"%s\" is not signed. %w", wyuFilePath, err)
			return EXIT_ERROR, err
		}
	}

	// extract the WYU to tmpDir
	_, files, err := Unzip(wyuFilePath, tmpDir)
	if nil != err {
		err = fmt.Errorf("error unzipping %s; %w", wyuFilePath, err)
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
	instDir := GetExeDir()
	backupDir, err := BackupFiles(updates, instDir)
	defer DeleteDirectory(backupDir)
	if nil != err {
		// Errors from rollback may occur from missing expected files - ignore
		RollbackFiles(backupDir, instDir)
		return EXIT_ERROR, err
	}

	// TODO is there a way to clean this up
	err = InstallUpdate(udt, updates, instDir)
	if nil != err {
		err = fmt.Errorf("error applying update; %w", err)
		// TODO rollback should restore client.wyc
		RollbackFiles(backupDir, instDir)

		err = os.Rename(wysFilePath, filepath.Join(instDir, INSTALL_FAILED_SENTINAL_WYS_FILE_NAME))
		if err != nil {
			err = fmt.Errorf("error renaming %s to failed install sentinel; %w", wysFilePath, err)
		}

		// start services, best effort
		for _, s := range udt.ServiceToStartAfterUpdate {
			svc := ValueToString(&s)
			_ = StartService(svc)
		}
		return EXIT_ERROR, err
	}

	// we haven't erred, write latest version number and exit
	// Newest version is recorded and we wipe out all temp files
	UpdateWYCWithNewVersionNumber(iuc, args.Cdata, wys.VersionToUpdate)
	return EXIT_SUCCESS, nil
}

// CheckForUpdateHandler checks to see if an update is availible. Returns int
// exit code and error.
func CheckForUpdateHandler(infoer Info, args Args) (int, error) {
	candidateUpdateReq, err := NewCandidateUpdateRequest(args, infoer)
	if err != nil {
		return EXIT_ERROR, err
	}

	iuc := candidateUpdateReq.ConfigIUC
	wys := candidateUpdateReq.ConfigWYS
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
