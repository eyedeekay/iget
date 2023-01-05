package iget

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
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
	//headers []string

	destLifespan    int
	timeoutTime     int
	tunnelLength    int
	inboundTunnels  int
	outboundTunnels int
	keepAlives      bool
	idleConns       int
	inboundBackups  int
	outboundBackups int

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

type WriteCounter struct {
	Total uint64
}

// Do "does" a request and returns a response. It's just a wrapper for http/client.Do
func (i *IGet) Do(req *http.Request) (*http.Response, error) {
	if i.markSize != 0 {
		if i.outputPath == "-" || i.outputPath == "stdout" {
			i.outputPath = path.Base(req.URL.String())
		}

	}
	c, e := i.client.Do(req)
	if e != nil {
		return nil, e
	}
	if i.outputPath != "-" && i.outputPath != "stdout" {
		tempDestinationPath := i.outputPath
		f, _ := os.OpenFile(tempDestinationPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		var RangeBottom int64
		if i.continueDownload {
			RangeBottom = i.DownloadedFileSize()
		}
		counter := &WriteCounter{
			Total: uint64(RangeBottom),
		}
		io.Copy(f, io.TeeReader(c.Body, counter))
	}
	return c, e
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
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
	b, err = ioutil.ReadAll(c.Body)
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
	r, err := http.NewRequest(i.method, i.url, ioutil.NopCloser(strings.NewReader(i.body)))
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
	if i.markSize != 0 {
		if i.outputPath == "-" || i.outputPath == "stdout" {
			i.outputPath = path.Base(c.Request.URL.String())
		}
	}
	if i.outputPath == "-" || i.outputPath == "stdout" {
		b, err := ioutil.ReadAll(c.Body)
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
		//fmt.Printf("%s", b)
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

		destLifespan:    3000000,
		timeoutTime:     3000000,
		tunnelLength:    3,
		inboundTunnels:  2,
		outboundTunnels: 2,
		keepAlives:      true,
		idleConns:       4,
		inboundBackups:  1,
		outboundBackups: 1,
		transport:       nil,
	}
	var err error
	for _, setter := range setters {
		setter(i)
	}
	i.samClient, err = goSam.NewClientFromOptions(
		goSam.SetCloseIdle(true),
		goSam.SetCloseIdleTime(3000000),
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
	i.transport = &http.Transport{
		Dial:                  i.samClient.Dial,
		MaxIdleConns:          i.idleConns,
		MaxIdleConnsPerHost:   i.idleConns,
		DisableKeepAlives:     i.keepAlives,
		ResponseHeaderTimeout: time.Duration(i.timeoutTime) * time.Millisecond,
		ExpectContinueTimeout: time.Duration(i.timeoutTime) * time.Millisecond,
		IdleConnTimeout:       time.Duration(i.timeoutTime) * time.Millisecond,
		TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	i.client = &http.Client{
		Transport: i.transport,
	}
	return i, nil
}
