package updater

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// >unzip client.wyc
// Archive:  client.wyc
//   inflating: iuclient.iuc
//   inflating: s.png
//   inflating: t.png

const (
	DSTRING_IUC_COMPANY_NAME           = 0x01
	DSTRING_IUC_PRODUCT_NAME           = 0x02
	DSTRING_IUC_INSTALLED_VERSION      = 0x03
	STRING_IUC_GUID                    = 0x0A
	DSTRING_IUC_SERVER_FILE_SITE       = 0x04
	DSTRING_IUC_WYUPDATE_SERVER_SITE   = 0x09
	DSTRING_IUC_HEADER_IMAGE_ALIGNMENT = 0x11
	INT_IUC_HEADER_TEXT_INDENT         = 0x12
	DSTRING_IUC_HEADER_TEXT_COLOR      = 0x13
	DSTRING_IUC_HEADER_FILENAME        = 0x14
	DSTRING_IUC_SIDE_IMAGE_FILENAME    = 0x15
	DSTRING_IUC_LANGUAGE_CULTURE       = 0x18 // e.g., en-US
	DSTRING_IUC_LANGUAGE_FILENAME      = 0x16
	BOOL_IUC_HIDE_HEADER_DIVIDER       = 0x17
	BOOL_IUC_CLOSE_WYUPDATE            = 0x19
	STRING_IUC_CUSTOM_TITLE_BAR        = 0x1A
	STRING_IUC_PUBLIC_KEY              = 0x1B
	END_IUC                            = 0xFF
)

var IUCTags = map[uint8]string{
	BOOL_IUC_CLOSE_WYUPDATE:            "BOOL_IUC_CLOSE_WYUPDATE",
	BOOL_IUC_HIDE_HEADER_DIVIDER:       "BOOL_IUC_HIDE_HEADER_DIVIDER",
	DSTRING_IUC_COMPANY_NAME:           "DSTRING_IUC_COMPANY_NAME",
	DSTRING_IUC_HEADER_FILENAME:        "DSTRING_IUC_HEADER_FILENAME",
	DSTRING_IUC_HEADER_IMAGE_ALIGNMENT: "DSTRING_IUC_HEADER_IMAGE_ALIGNMENT",
	DSTRING_IUC_HEADER_TEXT_COLOR:      "DSTRING_IUC_HEADER_TEXT_COLOR",
	DSTRING_IUC_INSTALLED_VERSION:      "DSTRING_IUC_INSTALLED_VERSION",
	DSTRING_IUC_LANGUAGE_CULTURE:       "DSTRING_IUC_LANGUAGE_CULTURE", // e.g., en-US
	DSTRING_IUC_LANGUAGE_FILENAME:      "DSTRING_IUC_LANGUAGE_FILENAME",
	DSTRING_IUC_PRODUCT_NAME:           "DSTRING_IUC_PRODUCT_NAME",
	DSTRING_IUC_SERVER_FILE_SITE:       "DSTRING_IUC_SERVER_FILE_SITE",
	DSTRING_IUC_SIDE_IMAGE_FILENAME:    "DSTRING_IUC_SIDE_IMAGE_FILENAME",
	DSTRING_IUC_WYUPDATE_SERVER_SITE:   "DSTRING_IUC_WYUPDATE_SERVER_SITE",
	INT_IUC_HEADER_TEXT_INDENT:         "INT_IUC_HEADER_TEXT_INDENT",
	STRING_IUC_CUSTOM_TITLE_BAR:        "STRING_IUC_CUSTOM_TITLE_BAR",
	STRING_IUC_GUID:                    "STRING_IUC_GUID",
	STRING_IUC_PUBLIC_KEY:              "STRING_IUC_PUBLIC_KEY",
	END_IUC:                            "END_IUC",
}

type ConfigIUC struct {
	IucCompanyName          TLV
	IucProductName          TLV
	IucInstalledVersion     TLV
	IucGUID                 TLV
	IucServerFileSite       []TLV
	IucWyupdateServerSite   []TLV
	IucHeaderImageAlignment TLV
	IucHeaderTextIndent     TLV
	IucHeaderTextColor      TLV
	IucHeaderFilename       TLV
	IucSideImageFilename    TLV
	IucLanguageCulture      TLV
	IucLanguageFilename     TLV
	IucHideHeaderDivider    TLV
	IucCloseWyupate         TLV
	IucCustomTitleBar       TLV
	IucPublicKey            TLV
}

type wycConfig struct {
	CompanyName          string
	ProductName          string
	InstalledVersions    string
	Guid                 string
	ServerFileSite       []string
	WyupdateServierSite  []string
	HeaderImageAlignment string
	HeaderTextIndent     int
	HeaderTextColor      string
	HeaderFilename       string
	SideImageFilename    string
	LanguageCulture      string
	LanguageFilename     string
	HideHeaderDivider    bool
	CloseWyupate         bool
	// strings not d. strings.Replace
	CustomTitleBar string
	PublicKey      string
}

// GetWYSURLs returns the ServerFileSite(s) listed in the WYC file.
func (config ConfigIUC) GetWYSURLs(args Args) (urls []string) {
	// WYS URL specified on the command line
	if len(args.Server) > 0 {
		urls = append(urls, args.Server)
		return urls
	}

	for _, s := range config.IucServerFileSite {
		u := strings.Replace(string(s.Value), "%urlargs%", args.Urlargs, 1)
		urls = append(urls, u)
	}
	return urls
}

// setWysUrls sets the URLs uses for fetching the .wys file
func (config *ConfigIUC) setWysUrls(urls ...string) {
	config.IucServerFileSite = make([]TLV, 0, len(urls))
	for _, url := range urls {
		entry := TLV{
			Tag:        DSTRING_IUC_SERVER_FILE_SITE,
			Type:       TLV_DSTRING,
			DataLength: uint32(len(url)) + 4,
			Length:     uint32(len(url)),
			Value:      []byte(url),
		}
		config.IucServerFileSite = append(config.IucServerFileSite, entry)
	}
}

// readIuc reads a .iuc file into a ConfigUIC structure
func readIuc(iucReader io.Reader) (ConfigIUC, error) {
	var config ConfigIUC

	// read HEADER
	header := make([]byte, len(IUC_HEADER))
	iucReader.Read(header)

	if string(header) != IUC_HEADER {
		err := fmt.Errorf("invalid iuclient.iuc file")
		return config, err
	}

	// read each TLV record until EOF or a TLV indicating the end
	// of the .iuc file
	for {
		tlv, err := readTlv(iucReader)
		if err != nil {
			return config, err
		}

		if tlv == nil {
			break
		}
		switch tlv.Tag {
		case DSTRING_IUC_COMPANY_NAME:
			tlv.Type = TLV_DSTRING
			config.IucCompanyName = *tlv
		case DSTRING_IUC_PRODUCT_NAME:
			tlv.Type = TLV_DSTRING
			config.IucProductName = *tlv
		case DSTRING_IUC_INSTALLED_VERSION:
			tlv.Type = TLV_DSTRING
			config.IucInstalledVersion = *tlv
		case DSTRING_IUC_SERVER_FILE_SITE:
			tlv.Type = TLV_DSTRING
			config.IucServerFileSite = append(config.IucServerFileSite, *tlv)
		case DSTRING_IUC_WYUPDATE_SERVER_SITE:
			tlv.Type = TLV_DSTRING
			config.IucWyupdateServerSite = append(config.IucWyupdateServerSite, *tlv)
		case DSTRING_IUC_HEADER_IMAGE_ALIGNMENT:
			tlv.Type = TLV_DSTRING
			config.IucHeaderImageAlignment = *tlv
		case DSTRING_IUC_HEADER_TEXT_COLOR:
			tlv.Type = TLV_DSTRING
			config.IucHeaderTextColor = *tlv
		case DSTRING_IUC_HEADER_FILENAME:
			tlv.Type = TLV_DSTRING
			config.IucHeaderFilename = *tlv
		case DSTRING_IUC_SIDE_IMAGE_FILENAME:
			tlv.Type = TLV_DSTRING
			config.IucSideImageFilename = *tlv
		case DSTRING_IUC_LANGUAGE_CULTURE:
			tlv.Type = TLV_DSTRING
			config.IucLanguageCulture = *tlv
		case DSTRING_IUC_LANGUAGE_FILENAME:
			tlv.Type = TLV_DSTRING
			config.IucLanguageFilename = *tlv
		case INT_IUC_HEADER_TEXT_INDENT:
			tlv.Type = TLV_INT
			config.IucHeaderTextIndent = *tlv
		case BOOL_IUC_HIDE_HEADER_DIVIDER:
			tlv.Type = TLV_BOOL
			config.IucHideHeaderDivider = *tlv
		case BOOL_IUC_CLOSE_WYUPDATE:
			tlv.Type = TLV_BOOL
			config.IucCloseWyupate = *tlv
		case STRING_IUC_CUSTOM_TITLE_BAR:
			tlv.Type = TLV_STRING
			config.IucCustomTitleBar = *tlv
		case STRING_IUC_PUBLIC_KEY:
			tlv.Type = TLV_STRING
			config.IucPublicKey = *tlv
		case STRING_IUC_GUID:
			tlv.Type = TLV_STRING
			config.IucGUID = *tlv
		default:
			err := fmt.Errorf("malformed .iuc file in .wyc archive")
			return config, err
		}
	}
	return config, nil
}

// ParseWYC parses a compress WYC file, returning the details as a ConfigIUC struct
func (wycInfo Info) ParseWYC(compressedWYC string) (ConfigIUC, error) {
	var config ConfigIUC

	zipr, err := zip.OpenReader(compressedWYC)
	if err != nil {
		return config, err
	}
	defer zipr.Close()

	for _, f := range zipr.File {
		// "iuclient.iuc" is the name of the uncompressed wyc file
		if f.FileHeader.Name == IUCLIENT_IUC {
			fh, err := f.Open()
			if err != nil {
				return config, err
			}
			defer fh.Close()
			config, err = readIuc(fh)
			if err != nil {
				return config, err
			}
		}
	}
	return config, nil
}

// writeIuc writes a IUC file
func writeIuc(config ConfigIUC, path string) error {
	f, err := os.Create(path)
	if nil != err {
		return err
	}
	defer f.Close()

	// write HEADER
	f.Write([]byte(IUC_HEADER))

	// DSTRING_IUC_COMPANY_NAME:
	writeTlv(f, config.IucCompanyName)

	// DSTRING_IUC_PRODUCT_NAME:
	writeTlv(f, config.IucProductName)

	// STRING_IUC_GUID:
	writeTlv(f, config.IucGUID)

	// DSTRING_IUC_INSTALLED_VERSION:
	writeTlv(f, config.IucInstalledVersion)

	// DSTRING_IUC_SERVER_FILE_SITE
	for _, s := range config.IucServerFileSite {
		writeTlv(f, s)
	}

	// DSTRING_IUC_WYUPDATE_SERVER_SITE - NOT USED
	for _, s := range config.IucWyupdateServerSite {
		writeTlv(f, s)
	}

	// DSTRING_IUC_HEADER_IMAGE_ALIGNMENT
	writeTlv(f, config.IucHeaderImageAlignment)

	// INT_IUC_HEADER_TEXT_INDENT
	writeTlv(f, config.IucHeaderTextIndent)

	// DSTRING_IUC_HEADER_TEXT_COLOR
	writeTlv(f, config.IucHeaderTextColor)

	// DSTRING_IUC_HEADER_FILENAME
	writeTlv(f, config.IucHeaderFilename)

	// DSTRING_IUC_SIDE_IMAGE_FILENAME:
	writeTlv(f, config.IucSideImageFilename)

	// DSTRING_IUC_LANGUAGE_CULTURE:
	writeTlv(f, config.IucLanguageCulture)

	// BOOL_IUC_HIDE_HEADER_DIVIDER:
	writeTlv(f, config.IucHideHeaderDivider)

	// STRING_IUC_PUBLIC_KEY:
	writeTlv(f, config.IucPublicKey)

	// DSTRING_IUC_LANGUAGE_FILENAME - NOT USED
	writeTlv(f, config.IucLanguageFilename)

	// STRING_IUC_CUSTOM_TITLE_BAR - NOT USED
	writeTlv(f, config.IucCustomTitleBar)

	// BOOL_IUC_CLOSE_WYUPDATE:
	writeTlv(f, config.IucCloseWyupate)

	err = binary.Write(f, binary.BigEndian, byte(END_IUC))
	if nil != err {
		return err
	}

	return nil
}

// UpdateWYCWithNewVersionNumber updates a WYC file with a new version number
func UpdateWYCWithNewVersionNumber(config ConfigIUC, origWYCFile string, version string) (newWYCFile string, err error) {
	// Unzip the archive. We'll create a new iuclient.iuc, but we need the
	// other files.
	tmpDir, err := CreateTempDir()
	if nil != err {
		err = fmt.Errorf("no temp dir; %v", err)
		return "", err
	}

	_, files, err := Unzip(origWYCFile, tmpDir)
	if nil != err {
		return "", err
	}

	config.IucInstalledVersion.Value = []byte(version)
	config.IucInstalledVersion.DataLength = uint32(len(config.IucInstalledVersion.Value) + 4)
	config.IucInstalledVersion.Length = uint32(len(config.IucInstalledVersion.Value))

	for _, f := range files {
		if filepath.Base(f) == IUCLIENT_IUC {
			// overwrite this file with new IUC
			err := writeIuc(config, f)
			if nil != err {
				return "", err
			}
		}
	}

	newWYCFile = filepath.Join(tmpDir, CLIENT_WYC)
	err = CreateWYCArchive(newWYCFile, files)
	if nil != err {
		return "", err
	}
	return newWYCFile, nil
}

// CreateWYCArchive compresses files into a .wyc archive
func CreateWYCArchive(filename string, files []string) error {

	wycHandle, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer wycHandle.Close()

	zipWriter := zip.NewWriter(wycHandle)
	defer zipWriter.Close()

	// Add files
	for _, file := range files {
		if err = AddFileToWYCArchive(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

// AddFileToWYCArchive adds a file to the archive
func AddFileToWYCArchive(zipWriter *zip.Writer, filename string) error {
	fileToCompress, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToCompress.Close()

	// get file info for the header
	info, err := fileToCompress.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate

	// all files in the WYC archive are at the root (Base())
	header.Name = filepath.Base(filename)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToCompress)
	return err
}
