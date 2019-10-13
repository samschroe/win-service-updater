package updater

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
)

const (
	TimeoutDial         = 30
	TimeoutTLSHandshake = 30
	TimeoutClient       = 60
)

// DownloadFile will download a file, saving it to a local file. It
// will try all URLs until one succeeds or all fail (error).
func DownloadFile(urls []string, localpath string) error {
	if len(urls) == 0 {
		err := fmt.Errorf("No download urls are specified.")
		return err
	}

	// Create the local output file
	out, err := os.Create(localpath)
	if nil != err {
		if len(localpath) == 0 {
			err = fmt.Errorf("Error trying to save file: %w", err)
		} else {
			err = fmt.Errorf("Error trying to save file \"%s\": %w", localpath, err)
		}
		return err
	}
	defer out.Close()

	// attempt each URL in the slice until download succeeds
	var result error
	for _, url := range urls {
		//  GET file, if we fail try next URL, otherwise return success (nil)
		err = HTTPGetFile(url, out)
		if nil != err {
			result = multierror.Append(result, err)
			continue
		} else {
			return nil
		}
	}

	return result
}

// HTTPGetFile GETs a file and saves it locally
func HTTPGetFile(URL string, file *os.File) error {
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second * TimeoutDial,
		}).Dial,
		TLSHandshakeTimeout: time.Second * TimeoutTLSHandshake,
	}

	httpClient := &http.Client{
		Timeout:   time.Second * TimeoutClient,
		Transport: httpTransport,
	}

	resp, err := httpClient.Get(URL)
	if nil != err {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || strings.Contains(resp.Header.Get("Content-type"), "text/html") {
		err = fmt.Errorf("Could not download \"%s\" - a web page was returned from the web server.", URL)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error downloading \"%s\": %s", URL, http.StatusText(resp.StatusCode))
		return err
	}

	// Write the body to file
	_, err = io.Copy(file, resp.Body)
	if nil != err {
		return err
	}

	return nil

}
