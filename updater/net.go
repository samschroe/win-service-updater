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
	"github.com/huntresslabs/win-service-updater/updater/useragent"
)

const (
	TimeoutDial         = 30
	TimeoutTLSHandshake = 30
	TimeoutClient       = 60
)

// DownloadFileToDisk will download the content linked by one of the provided urls and save it locally to localpath. It
// will try all URLs in order until one succeeds. If all fail it will return an error.
func DownloadFileToDisk(urls []string, localpath string) error {
	if len(localpath) == 0 {
		return fmt.Errorf("Error trying to save file: no file path provide")
	}

	// Create the local output file
	out, err := os.Create(localpath)
	if nil != err {
		return fmt.Errorf("Error trying to save file \"%s\": %w", localpath, err)
	}
	defer out.Close()

	return DownloadFileToWriter(urls, out)
}

// DownloadFileToWriter will download the content linked by one of the provided urls and write it to the provided writer. It
// will try all URLs in order until one succeeds. If all fail it will return an error.
func DownloadFileToWriter(urls []string, writer io.Writer) error {
	if len(urls) == 0 {
		err := fmt.Errorf("No download urls are specified.")
		return err
	}

	// attempt each URL in the slice until download succeeds
	var result error
	for _, url := range urls {
		//  GET file, if we fail try next URL, otherwise return success (nil)
		err := HTTPGetFile(url, writer)
		if nil == err {
			return nil
		}

		result = multierror.Append(result, err)
	}

	return result
}

// HTTPGetFile GETs the contented linked by the URL and writes it to the writer and
// returns an error if the content is HTML or the HTTP request doesn't respond with 200 (OK).
func HTTPGetFile(URL string, writer io.Writer) error {
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

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if nil != err {
		return err
	}

	req.Header.Set("User-Agent", useragent.GetUserAgentString())

	resp, err := httpClient.Do(req)
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
	_, err = io.Copy(writer, resp.Body)
	if nil != err {
		return err
	}

	return nil

}
