package iget

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/eyedeekay/goSam"
)

// IGet is an IGet client
type IGet struct {
	samHost string
	samPort string
	debug   bool
	verb    bool

	username string
	password string

	outputPath string

	method string
	url    string
	body   string

	destLifespan    int
	timeoutTime     int
	tunnelLength    int
	inboundTunnels  int
	outboundTunnels int
	keepAlives      bool
	idleConns       int
	inboundBackups  int
	outboundBackups int

	toPort   string
	fromPort string

	client    *http.Client
	transport *http.Transport
	samClient *goSam.Client

	lineLength       int
	continueDownload bool
	markSize         int
}

func (i IGet) samaddress() string {
	return i.samHost + ":" + i.samPort
}

// applyPortOptions applies optional SAM virtual-port settings to the SAM client.
func (i *IGet) applyPortOptions() error {
	if i.toPort != "" {
		if err := goSam.SetToPort(i.toPort)(i.samClient); err != nil {
			return err
		}
	}
	if i.fromPort != "" {
		if err := goSam.SetFromPort(i.fromPort)(i.samClient); err != nil {
			return err
		}
	}
	return nil
}

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
			fmt.Fprintf(os.Stderr, "Saving to: %s\n", i.outputPath)
		}
	}
}

// saveToFile copies the response body to the configured output file with progress tracking.
func (i *IGet) saveToFile(body io.Reader) error {
	f, err := os.OpenFile(i.outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	var rangeBottom int64
	if i.continueDownload {
		rangeBottom = i.DownloadedFileSize()
	}
	counter := &WriteCounter{
		Total:    uint64(rangeBottom),
		MarkSize: uint64(i.markSize),
	}
	_, copyErr := io.Copy(f, io.TeeReader(body, counter))
	return copyErr
}

// Do "does" a request and returns a response. It's just a wrapper for http/client.Do
func (i *IGet) Do(req *http.Request) (*http.Response, error) {
	i.resolveOutputPath(req.URL.String())
	if i.verb {
		fmt.Fprintf(os.Stderr, "[iget] %s %s\n", req.Method, req.URL)
	}
	c, e := i.client.Do(req)
	if e != nil {
		return nil, e
	}
	if i.verb {
		fmt.Fprintf(os.Stderr, "[iget] response: %s\n", c.Status)
	}
	if i.outputPath != "-" && i.outputPath != "stdout" {
		if err := i.saveToFile(c.Body); err != nil {
			return nil, err
		}
	}
	return c, e
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	if wc.MarkSize > 0 && wc.Total/wc.MarkSize > wc.LastMark/wc.MarkSize {
		wc.PrintProgress()
		wc.LastMark = wc.Total
	}
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Fprintf(os.Stdout, "\r%s", strings.Repeat(" ", 35))
	fmt.Fprintf(os.Stdout, "\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

// DoBytes does a request and returns the body of a response as bytes.
// it is unused in iget, it's here for a hypothetical API and testing.
func (i *IGet) DoBytes(req *http.Request) ([]byte, error) {
	var b []byte
	var err error
	c, e := i.Do(req)
	if e != nil {
		return b, e
	}
	defer c.Body.Close()
	b, err = io.ReadAll(c.Body)
	if err != nil {
		return b, err
	}
	return b, nil
}

// DoString does a request, converts the body to bytes, then returns the body of a response as a string.
// it is unused in iget, it's here for a hypothetical API and testing.
func (i *IGet) DoString(req *http.Request) (string, error) {
	b, e := i.DoBytes(req)
	if e != nil {
		return "", e
	}
	return string(b), nil
}

// Request generates an *http.Request
func (i *IGet) Request(setters ...RequestOption) (*http.Request, error) {
	r, err := http.NewRequest(i.method, i.url, io.NopCloser(strings.NewReader(i.body)))
	if err != nil {
		return nil, err
	}
	if i.continueDownload {
		RangeBottom := i.DownloadedFileSize()
		if RangeBottom > 0 {
			r.Header.Add("Range", fmt.Sprintf("bytes=%d-", RangeBottom))
		}
	}
	for _, setter := range setters {
		setter(r)
	}
	r.Header.Set("User-Agent", "wget/1.11.4")
	return r, nil
}

func (i *IGet) DownloadedFileSize() int64 {
	f, e := os.Stat(i.outputPath)
	if e != nil {
		return 0
	}
	return f.Size()
}

// PrintResponse routes the output
func (i *IGet) PrintResponse(c *http.Response) string {
	i.resolveOutputPath(c.Request.URL.String())
	if i.outputPath == "-" || i.outputPath == "stdout" {
		b, err := io.ReadAll(c.Body)
		if err != nil {
			return ""
		}
		a := []rune(string(b))
		for index, r := range a {
			if index > 0 && (index+1)%i.lineLength == 0 {
				fmt.Printf("%c\n", r)
			} else {
				fmt.Printf("%c", r)
			}
		}
		// fmt.Printf("%s", b)
		return string(b)
	}
	return ""
}

// NewIGet is an IGet Client
func NewIGet(setters ...Option) (*IGet, error) {
	i := &IGet{
		samHost:    "localhost",
		samPort:    "7656",
		verb:       false,
		debug:      false,
		outputPath: "-",

		destLifespan:     6 * 60 * 1000,
		timeoutTime:      3000000,
		tunnelLength:     3,
		inboundTunnels:   2,
		outboundTunnels:  2,
		keepAlives:       false,
		idleConns:        4,
		inboundBackups:   1,
		outboundBackups:  1,
		continueDownload: true,
		lineLength:       80,
		transport:        nil,
	}
	var err error
	for _, setter := range setters {
		setter(i)
	}
	i.samClient, err = goSam.NewClientFromOptions(
		goSam.SetCloseIdle(true),
		goSam.SetCloseIdleTime(uint(i.destLifespan)),
		goSam.SetHost(i.samHost),
		goSam.SetPort(i.samPort),
		goSam.SetDebug(i.debug),
		goSam.SetInLength(uint(i.tunnelLength)),
		goSam.SetOutLength(uint(i.tunnelLength)),
		goSam.SetInQuantity(uint(i.inboundTunnels)),
		goSam.SetOutQuantity(uint(i.outboundTunnels)),
		goSam.SetInBackups(uint(i.inboundBackups)),
		goSam.SetOutBackups(uint(i.outboundBackups)),
		goSam.SetUser(i.username),
		goSam.SetPass(i.password),
	)
	if err != nil {
		return nil, err
	}
	if err = i.applyPortOptions(); err != nil {
		return nil, err
	}
	if i.verb {
		fmt.Fprintf(os.Stderr, "[iget] SAM bridge: %s  tunnels in=%d out=%d length=%d\n",
			i.samaddress(), i.inboundTunnels, i.outboundTunnels, i.tunnelLength)
	}
	i.transport = &http.Transport{
		Dial:                  i.samClient.Dial,
		MaxIdleConns:          i.idleConns,
		MaxIdleConnsPerHost:   i.idleConns,
		DisableKeepAlives:     i.keepAlives,
		ResponseHeaderTimeout: time.Duration(i.timeoutTime) * time.Millisecond,
		ExpectContinueTimeout: time.Duration(i.timeoutTime) * time.Millisecond,
		IdleConnTimeout:       time.Duration(i.timeoutTime) * time.Millisecond,
		TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		// InsecureSkipVerify is intentional: i2p eepsites commonly use self-signed certificates.
		// All traffic is routed through the i2p network via SAM, so there is no clearnet exposure.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	i.client = &http.Client{
		Transport: i.transport,
	}
	return i, nil
}
