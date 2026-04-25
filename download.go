package iget

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
)

// WriteCounter tracks the number of bytes written during a download and
// optionally prints progress messages to stdout at configurable intervals.
// Total is the running byte count, MarkSize is the interval between progress
// updates, and LastMark records the Total value at the last printed update.
type WriteCounter struct {
	Total    uint64
	MarkSize uint64
	LastMark uint64
}

// resolveOutputPath sets outputPath to the URL base name when markSize is set
// and the output destination is stdout. Should be called before writing or printing a response.
func (i *IGet) resolveOutputPath(rawURL string) {
	if i.markSize != 0 {
		if i.outputPath == "-" || i.outputPath == "stdout" {
			i.outputPath = path.Base(rawURL)
			fmt.Fprintf(i.verboseOut, "Saving to: %s\n", i.outputPath)
		}
	}
}

// saveToFile copies the response body to the configured output file with progress tracking.
// When continueDownload is true the file is opened for append so that a partial download
// can be extended; when false the file is truncated so that a fresh download always
// produces an intact file rather than appending to stale content.
func (i *IGet) saveToFile(body io.Reader) (err error) {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if i.continueDownload {
		flags = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	}
	f, err := os.OpenFile(i.outputPath, flags, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	var rangeBottom int64
	if i.continueDownload {
		rangeBottom = i.DownloadedFileSize()
	}
	counter := &WriteCounter{
		Total:    uint64(rangeBottom),
		MarkSize: uint64(i.markSize),
	}
	_, err = io.Copy(f, io.TeeReader(body, counter))
	return err
}

// truncateIfRangeIgnored truncates the output file when a resume Range request
// was sent but the server returned a full 200 response. This prevents the
// pre-existing partial bytes from being duplicated by the appended full body.
func (i *IGet) truncateIfRangeIgnored(req *http.Request, resp *http.Response) error {
	if !i.continueDownload || req.Header.Get("Range") == "" {
		return nil
	}
	if resp.StatusCode == http.StatusPartialContent {
		return nil
	}
	if err := os.Truncate(i.outputPath, 0); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("iget: server returned %s for range request; failed to truncate: %w", resp.Status, err)
	}
	return nil
}

// DownloadedFileSize returns the current size of the output file in bytes,
// used to set the Range header when resuming a partial download.
func (i *IGet) DownloadedFileSize() int64 {
	f, e := os.Stat(i.outputPath)
	if e != nil {
		return 0
	}
	return f.Size()
}

// PrintResponse routes the output to stdout, streaming the response body
// rather than buffering it entirely in memory. When lineLength is 0 the body
// is copied directly to stdout. When lineLength > 0 a bufio.Reader is used to
// read one rune at a time and a newline is inserted at each lineLength boundary
// without pre-loading the whole body.
func (i *IGet) PrintResponse(c *http.Response) string {
	defer c.Body.Close()
	i.resolveOutputPath(c.Request.URL.String())
	if i.outputPath != "-" && i.outputPath != "stdout" {
		return ""
	}
	if i.lineLength <= 0 {
		io.Copy(os.Stdout, c.Body) //nolint:errcheck
		return ""
	}
	br := bufio.NewReader(c.Body)
	col := 0
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			break
		}
		col++
		fmt.Printf("%c", r)
		if col%i.lineLength == 0 {
			fmt.Printf("\n")
		}
	}
	return ""
}

// Write implements io.Writer. It accumulates the byte count and calls
// PrintProgress whenever the total crosses a new MarkSize boundary.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	if wc.MarkSize > 0 && wc.Total/wc.MarkSize > wc.LastMark/wc.MarkSize {
		wc.PrintProgress()
		wc.LastMark = wc.Total
	}
	return n, nil
}

// PrintProgress writes a human-readable download progress line to stdout,
// overwriting the previous line in place via a carriage return.
func (wc WriteCounter) PrintProgress() {
	fmt.Fprintf(os.Stdout, "\r%s", strings.Repeat(" ", 35))
	fmt.Fprintf(os.Stdout, "\rDownloading... %s complete", humanize.Bytes(wc.Total))
}
