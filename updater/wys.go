// Parser for wys files
// File ID: IUSDFV2
// Compressed File ID: = { 0x50, 0x4b, 0x03, 0x04 } = { 'P', 'K', 0x03, 0x04 }
// File Extension: wys

package updater

import (
	"archive/zip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WYS tags
const (
	DSTRING_WYS_CURRENT_LAST_VERSION      = 0x01
	DSTRING_WYS_SERVER_FILE_SITE          = 0x02
	DSTRING_WYS_MIN_CLIENT_VERSION        = 0x07
	INT_WYS_DUMMY_VAR_LEN                 = 0x0F
	DSTRING_WYS_VERSION_TO_UPDATE         = 0x0B
	DSTRING_WYS_UPDATE_FILE_SITE          = 0x03
	BYTE_WYS_RTF                          = 0x80
	DSTRING_WYS_LATEST_CHANGES            = 0x04
	LONG_WYS_UPDATE_FILE_SIZE             = 0x09
	LONG_WYS_UPDATE_FILE_ADLER32_CHECKSUM = 0x08
	BYTE_WYS_FILE_SHA1                    = 0x14
	INT_WYS_FOLDER                        = 0x0A
	DSTRING_WYS_UPDATE_ERROR_TEXT         = 0x20
	DSTRING_WYS_UPDATE_ERROR_LINK         = 0x21
	END_WYS                               = 0xFF
)

// WYSTags is a mapping of WYS tags to strings
var WYSTags = map[uint8]string{
	BYTE_WYS_FILE_SHA1:                    "BYTE_WYS_FILE_SHA1",
	BYTE_WYS_RTF:                          "BYTE_WYS_RTF",
	DSTRING_WYS_CURRENT_LAST_VERSION:      "DSTRING_WYS_CURRENT_LAST_VERSION",
	DSTRING_WYS_LATEST_CHANGES:            "DSTRING_WYS_LATEST_CHANGES",
	DSTRING_WYS_MIN_CLIENT_VERSION:        "DSTRING_WYS_MIN_CLIENT_VERSION",
	DSTRING_WYS_SERVER_FILE_SITE:          "DSTRING_WYS_SERVER_FILE_SITE",
	DSTRING_WYS_UPDATE_ERROR_LINK:         "DSTRING_WYS_UPDATE_ERROR_LINK",
	DSTRING_WYS_UPDATE_ERROR_TEXT:         "DSTRING_WYS_UPDATE_ERROR_TEXT",
	DSTRING_WYS_UPDATE_FILE_SITE:          "DSTRING_WYS_UPDATE_FILE_SITE",
	DSTRING_WYS_VERSION_TO_UPDATE:         "DSTRING_WYS_VERSION_TO_UPDATE",
	END_WYS:                               "END_WYS",
	INT_WYS_DUMMY_VAR_LEN:                 "INT_WYS_DUMMY_VAR_LEN",
	INT_WYS_FOLDER:                        "INT_WYS_FOLDER",
	LONG_WYS_UPDATE_FILE_ADLER32_CHECKSUM: "LONG_WYS_UPDATE_FILE_ADLER32_CHECKSUM",
	LONG_WYS_UPDATE_FILE_SIZE:             "LONG_WYS_UPDATE_FILE_SIZE",
}

// ConfigWYS contains the server file (WYS) details
type ConfigWYS struct {
	FileSha1           []byte
	RTF                []byte
	CurrentLastVersion string
	LatestChanges      string
	MinClientVersion   string
	ServerFileSite     string
	UpdateErrorLink    string
	UpdateErrorText    string
	UpdateFileSite     []string // hosts the WYU file
	VersionToUpdate    string
	DummyVarLen        int
	WYSFolder          int
	UpdateFileAdler32  int64
	UpdateFileSize     int64
}

func ReadWYSTLV(r io.Reader) *TLV {
	var record TLV

	err := binary.Read(r, binary.BigEndian, &record.Tag)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return nil
	}

	if record.Tag == END_WYS {
		return nil
	}

	record.TagString = WYSTags[record.Tag]

	// handle d. strings with the data length
	switch record.Tag {
	case DSTRING_WYS_CURRENT_LAST_VERSION,
		DSTRING_WYS_LATEST_CHANGES,
		DSTRING_WYS_MIN_CLIENT_VERSION,
		DSTRING_WYS_SERVER_FILE_SITE,
		DSTRING_WYS_UPDATE_ERROR_LINK,
		DSTRING_WYS_UPDATE_ERROR_TEXT,
		DSTRING_WYS_UPDATE_FILE_SITE,
		DSTRING_WYS_VERSION_TO_UPDATE:
		err = binary.Read(r, binary.LittleEndian, &record.DataLength)
		if err != nil {
			return nil
		}
	default:
	}

	err = binary.Read(r, binary.LittleEndian, &record.Length)
	if err != nil {
		return nil
	}

	// there is no value for the dummy var
	if record.Tag == INT_WYS_DUMMY_VAR_LEN {
		return &record
	}

	record.Value = make([]byte, record.Length)
	_, err = io.ReadFull(r, record.Value)
	if err != nil {
		return nil
	}

	return &record
}

// ParseWYSFromReader returns the parsed contents, wys, as read from the reader, which are assumed to be a compresseed WYS file,
// annotated with the content size.  wys will always be zero value when err is not nil.
func (wysInfo Info) ParseWYSFromReader(reader io.ReaderAt, size int64) (wys ConfigWYS, err error) {
	zipr, err := zip.NewReader(reader, size)
	if err != nil {
		return wys, err
	}

	return wysInfo.parseWYSFromZipReader(zipr)
}

// ParseWYSFromReader returns the parsed contents, wys, as read from the compressedWYSFilePath.
// wys will always be zero value when err is not nil.
func (wysInfo Info) ParseWYSFromFilePath(compressedWYSFilePath string, _ Args) (wys ConfigWYS, err error) {
	zipr, err := zip.OpenReader(compressedWYSFilePath)
	if err != nil {
		return wys, err
	}
	defer zipr.Close()

	return wysInfo.parseWYSFromZipReader(&zipr.Reader)
}

// parseWYSFromZipReader returns the parsed contents, wys, as read from zipr.
// wys will always be zero value when err is not nil.
func (wysInfo Info) parseWYSFromZipReader(zipr *zip.Reader) (wys ConfigWYS, err error) {
	for _, f := range zipr.File {
		// there is only one file in the archive
		// "0" is the name of the uncompressed wys file
		if f.FileHeader.Name == "0" {
			fh, err := f.Open()
			if err != nil {
				return wys, err
			}
			defer fh.Close()

			// read HEADER
			header := make([]byte, 7)
			fh.Read(header)

			if string(header) != WYS_HEADER {
				err = fmt.Errorf("invalid wys header")
				return wys, err
			}

			for {
				tlv := ReadWYSTLV(fh)
				if tlv == nil {
					break
				}

				switch tlv.Tag {
				case BYTE_WYS_FILE_SHA1:
					wys.FileSha1 = ValueToByteSlice(tlv)
				case BYTE_WYS_RTF:
					wys.RTF = ValueToByteSlice(tlv)
				case DSTRING_WYS_CURRENT_LAST_VERSION:
					wys.CurrentLastVersion = ValueToString(tlv)
				case DSTRING_WYS_LATEST_CHANGES:
					wys.LatestChanges = ValueToString(tlv)
				case DSTRING_WYS_MIN_CLIENT_VERSION:
					wys.MinClientVersion = ValueToString(tlv)
				case DSTRING_WYS_SERVER_FILE_SITE:
					wys.ServerFileSite = ValueToString(tlv)
				case DSTRING_WYS_UPDATE_ERROR_LINK:
					wys.UpdateErrorLink = ValueToString(tlv)
				case DSTRING_WYS_UPDATE_ERROR_TEXT:
					wys.UpdateErrorText = ValueToString(tlv)
				case DSTRING_WYS_UPDATE_FILE_SITE:
					wys.UpdateFileSite = append(wys.UpdateFileSite, ValueToString(tlv))
				case DSTRING_WYS_VERSION_TO_UPDATE:
					wys.VersionToUpdate = ValueToString(tlv)
				case INT_WYS_DUMMY_VAR_LEN:
					// do nothing
				case INT_WYS_FOLDER:
					wys.WYSFolder = ValueToInt(tlv)
				case LONG_WYS_UPDATE_FILE_ADLER32_CHECKSUM:
					wys.UpdateFileAdler32 = ValueToLong(tlv)
				case LONG_WYS_UPDATE_FILE_SIZE:
					wys.UpdateFileSize = ValueToLong(tlv)
				default:
					err := fmt.Errorf("wys tag %x not implemented", tlv.Tag)
					return wys, err
				}
			}

			return wys, nil
		}
	}

	// wys not parsed
	err = fmt.Errorf("wys not parsed")
	return wys, err
}

// GetWYUURLs returns the UpdateFileSite(s) included in the WYS file associated with config and populates
// urls with the site URLs.
// args are used to inject any CLI provided URL arguments and allow for overriding of the site URL.
func (wys ConfigWYS) GetWYUURLs(args Args) (urls []string) {
	urlsToConsider := wys.UpdateFileSite
	// This can only be specified in tests
	if len(args.WYUTestServer) > 0 {
		urlsToConsider = []string{args.WYUTestServer}
	}

	// we want to allow injection of URL args on an overidden URL as well
	// as the one(s) in the WYC file
	for _, s := range urlsToConsider {
		u := strings.Replace(s, "%urlargs%", args.Urlargs, 1)
		urls = append(urls, u)
	}
	return urls
}

var lastWyuFilePath string

// lastWyuDownload returns the pathname for the wyu cache. This file
// will contain the most recently downloaded wyu file
func (wys ConfigWYS) lastWyuDownload() string {
	// small efficiency hack that allows us to easily test
	if lastWyuFilePath != "" {
		return lastWyuFilePath
	}
	const lastWyuFileName = "last_wyu_download"

	instDir := GetExeDir()
	lastWyuFilePath = filepath.Join(instDir, lastWyuFileName)

	return lastWyuFilePath
}

// copyFile is a utility function to copy one file to another
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// getWyuFile returns the wyu file identified in the ConfigWYS into
// the fp location. It checks to see if we have a previously
// downloaded wyu file and verifies that it matches the adler32
// checksum present in the ConfigWYS struct
func (wys ConfigWYS) getWyuFile(args Args, fp string) error {
	lastWyuDownload := wys.lastWyuDownload()

	_, err := os.Stat(lastWyuDownload)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// if there was a stat error and it is anything but an
		// ErrNotExist then return that error
		return err
	}

	// err == nil means the file exists. If the UpdateFileAdler32
	// matches then just copy the cached file
	if err == nil && VerifyAdler32Checksum(wys.UpdateFileAdler32, lastWyuDownload) {
		if err := copyFile(lastWyuDownload, fp); err == nil {
			LogOutputInfoMsg(args, "Reusing cached WYU file")
			return nil
		}
		// if the copy file fails fall through to downloading
		// the file
	}

	// if we get here lastWyuDownload does not exist, the adler32
	// mismatched, or we could not copy the cached file.  Download
	// the wyu file and copy it to the lastWyuDownload (cached
	// location)
	urls := wys.GetWYUURLs(args)
	err = DownloadFileToDisk(urls, fp)

	// check to make sure the downloaded file matches the adler32
	// checksum
	if err == nil {
		if VerifyAdler32Checksum(wys.UpdateFileAdler32, fp) {
			// if this copy fails log the error message
			// but still return success (no error).
			if err := copyFile(fp, lastWyuDownload); err != nil {
				LogOutputInfoMsg(args, fmt.Sprintf("Error caching WYU file: %v", err))
			}
		} else {
			err = fmt.Errorf(`The downloaded file "%s" failed the Adler32 validation.`, fp)
		}

	}

	return err
}
