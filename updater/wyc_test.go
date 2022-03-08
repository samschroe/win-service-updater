package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func appendFiles(filename string, zipw *zip.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := zipw.Create(path.Base(filename))
	if err != nil {
		msg := "Failed to create entry for %s in zip file: %s"
		return fmt.Errorf(msg, filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Failed to write %s to zip: %s", filename, err)
	}

	return nil
}

func Zip(archive string, files []string) {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(archive, flags, 0644)
	if err != nil {
		log.Fatalf("Failed to open zip for writing: %s", err)
	}
	defer file.Close()

	zipw := zip.NewWriter(file)
	defer zipw.Close()

	for _, filename := range files {
		if err := appendFiles(filename, zipw); err != nil {
			log.Fatalf("Failed to add file %s to zip: %s", filename, err)
		}
	}
}

func TestWYC(t *testing.T) {
	origFile := "./testdata/client.1.0.0.wyc"

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)
	assert.Equal(t, wyc.IucServerFileSite[0].Value, []byte("http://127.0.0.1/updates/wyserver.wys?key=%urlargs%"))
}

func TestWYC_GetWYSURLs(t *testing.T) {
	origFile := "./testdata/client.1.0.0.wyc"

	args := Args{}

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)
	urls := wyc.GetWYSURLs(args)
	assert.Equal(t, 1, len(urls))
}

func TestWYC_UpdateWYCWithNewVersion(t *testing.T) {
	origFile := "./testdata/client.1.0.0.wyc"

	info := Info{}
	wyc, err := info.ParseWYC(origFile)
	assert.Nil(t, err)

	new, err := UpdateWYCWithNewVersionNumber(wyc, origFile, "1.2.3.4")
	assert.Nil(t, err)
	// fmt.Println(new)

	newConfig, err := info.ParseWYC(new)
	assert.Nil(t, err)
	assert.Equal(t, string(newConfig.IucInstalledVersion.Value), "1.2.3.4")
}

func TestWYC_WriteIUC(t *testing.T) {
	origClientWYC := "./testdata/client.1.0.0.wyc"

	// create a new uiclient.iuc and compare it to the one in the archive
	info := Info{}
	wyc, err := info.ParseWYC(origClientWYC)
	assert.Nil(t, err)

	tmpIUC := GenerateTempFile()
	defer os.Remove(tmpIUC)

	err = writeIuc(wyc, tmpIUC)
	assert.Nil(t, err)

	tmpDir := GetTempDir()
	defer os.RemoveAll(tmpDir)

	newHash, err := GetSHA256(tmpIUC)
	assert.Nil(t, err)

	found := false
	_, files, err := Unzip(origClientWYC, tmpDir)
	for _, f := range files {
		// fmt.Println(f)
		if filepath.Base(f) == IUCLIENT_IUC {
			origHash, err := GetSHA256(f)
			assert.Nil(t, err)
			assert.Equal(t, origHash, newHash)
			found = true
		}
	}
	assert.True(t, found)
}

// TestWyc_setWycUrls ...
func TestWyc_setWycUrls(t *testing.T) {
	origClientWYC := "./testdata/client.1.0.0.wyc"

	// create a new uiclient.iuc and compare it to the one in the archive
	info := Info{}
	wyc, err := info.ParseWYC(origClientWYC)
	assert.Nil(t, err)

	args := Args{}
	url := "http://floob.com/blorf"
	wyc.setWysUrls(url)

	tmpIUC := GenerateTempFile()
	defer os.Remove(tmpIUC)

	err = writeIuc(wyc, tmpIUC)
	assert.Nil(t, err)

	tmpDir := GetTempDir()
	defer os.RemoveAll(tmpDir)

	fh, err := os.Open(tmpIUC)
	if err != nil {
		t.Fatalf("could not open %s", tmpIUC)
	}
	defer fh.Close()
	config, err := readIuc(fh)
	if err != nil {
		t.Fatal(err)
	}
	savedUrls := config.GetWYSURLs(args)
	assert.Equal(t, 1, len(savedUrls))
	assert.Equal(t, url, savedUrls[0])

}

// TestWyc_generateFile ...
func TestWyc_generateFile(t *testing.T) {
	origClientWYC := "./testdata/client.1.0.0.wyc"

	// create a new uiclient.iuc and compare it to the one in the archive
	info := Info{}
	wyc, err := info.ParseWYC(origClientWYC)
	assert.Nil(t, err)

	tmpDir := GetTempDir()
	defer os.RemoveAll(tmpDir)

	tmpIUC := GenerateTempFile()
	defer os.Remove(tmpIUC)

	if err := copyIucToFile(tmpIUC, wyc); err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	fh, err := os.Open(tmpIUC)
	if err != nil {
		t.Fatal(err)
	}

	newConfig, err := readIuc(fh)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, wyc, newConfig)
}

// copyIucToFile ...
func copyIucToFile(tmpIUC string, wyc ConfigIUC) error {
	tmpIucFile, err := os.Create(tmpIUC)
	if err != nil {
		return err
	}

	defer tmpIucFile.Close()

	// write HEADER
	if _, err := tmpIucFile.Write([]byte(IUC_HEADER)); err != nil {
		return err
	}

	// DSTRING_IUC_COMPANY_NAME:
	// writeTlv(f, wyc.IucCompanyName)

	if err := writeTagAsTlv(tmpIucFile, wyc.IucCompanyName.Tag, ValueToString(&wyc.IucCompanyName)); err != nil {
		return err
	}
	// DSTRING_IUC_PRODUCT_NAME:
	// writeTlv(f, wyc.IucProductName)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucProductName.Tag, ValueToString(&wyc.IucProductName)); err != nil {
		return err
	}

	// STRING_IUC_GUID:
	// writeTlv(f, wyc.IucGUID)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucGUID.Tag, ValueToString(&wyc.IucGUID)); err != nil {
		return err
	}

	// DSTRING_IUC_INSTALLED_VERSION:
	// writeTlv(f, wyc.IucInstalledVersion)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucInstalledVersion.Tag, ValueToString(&wyc.IucInstalledVersion)); err != nil {
		return err
	}

	// DSTRING_IUC_SERVER_FILE_SITE
	for _, s := range wyc.IucServerFileSite {
		// writeTlv(f, s)
		if err := writeTagAsTlv(tmpIucFile, s.Tag, ValueToString(&s)); err != nil {
			return err
		}
	}

	// DSTRING_IUC_WYUPDATE_SERVER_SITE - NOT USED
	for _, s := range wyc.IucWyupdateServerSite {
		// writeTlv(f, s)
		if err := writeTagAsTlv(tmpIucFile, s.Tag, ValueToString(&s)); err != nil {
			return err
		}
	}

	// DSTRING_IUC_HEADER_IMAGE_ALIGNMENT
	// writeTlv(f, wyc.IucHeaderImageAlignment)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucHeaderImageAlignment.Tag, ValueToString(&wyc.IucHeaderImageAlignment)); err != nil {
		return err
	}

	// INT_IUC_HEADER_TEXT_INDENT
	// writeTlv(f, wyc.IucHeaderTextIndent)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucHeaderTextIndent.Tag, ValueToInt(&wyc.IucHeaderTextIndent)); err != nil {
		return err
	}

	// DSTRING_IUC_HEADER_TEXT_COLOR
	// writeTlv(f, wyc.IucHeaderTextColor)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucHeaderTextColor.Tag, ValueToString(&wyc.IucHeaderTextColor)); err != nil {
		return err
	}

	// DSTRING_IUC_HEADER_FILENAME
	// writeTlv(f, wyc.IucHeaderFilename)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucHeaderFilename.Tag, ValueToString(&wyc.IucHeaderFilename)); err != nil {
		return err
	}

	// DSTRING_IUC_SIDE_IMAGE_FILENAME:
	// writeTlv(f, wyc.IucSideImageFilename)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucSideImageFilename.Tag, ValueToString(&wyc.IucSideImageFilename)); err != nil {
		return err
	}

	// DSTRING_IUC_LANGUAGE_CULTURE:
	// writeTlv(f, wyc.IucLanguageCulture)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucLanguageCulture.Tag, ValueToString(&wyc.IucLanguageCulture)); err != nil {
		return err
	}

	// BOOL_IUC_HIDE_HEADER_DIVIDER:
	// writeTlv(f, wyc.IucHideHeaderDivider)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucHideHeaderDivider.Tag, ValueToBool(&wyc.IucHideHeaderDivider)); err != nil {
		return err
	}

	// STRING_IUC_PUBLIC_KEY:
	// writeTlv(f, wyc.IucPublicKey)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucPublicKey.Tag, ValueToString(&wyc.IucPublicKey)); err != nil {
		return err
	}

	// DSTRING_IUC_LANGUAGE_FILENAME - NOT USED
	// writeTlv(f, wyc.IucLanguageFilename)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucLanguageFilename.Tag, ValueToString(&wyc.IucLanguageFilename)); err != nil {
		return err
	}

	// STRING_IUC_CUSTOM_TITLE_BAR - NOT USED
	// writeTlv(f, wyc.IucCustomTitleBar)
	if err := writeTagAsTlv(tmpIucFile, wyc.IucCustomTitleBar.Tag, ValueToString(&wyc.IucCustomTitleBar)); err != nil {
		return err
	}

	// BOOL_IUC_CLOSE_WYUPDATE:
	// writeTlv(f, wyc.IucCloseWyupate)
	if wyc.IucCloseWyupate.Tag != 0 {
		if err := writeTagAsTlv(tmpIucFile, wyc.IucCloseWyupate.Tag, ValueToBool(&wyc.IucCloseWyupate)); err != nil {
			return err
		}
	}
	return nil
}
