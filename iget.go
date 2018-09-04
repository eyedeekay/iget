package iget

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

import (
	"github.com/eyedeekay/gosam"
)

//IGet is an IGet client
type IGet struct {
	samHost string
	samPort string
	debug   bool
	verb    bool

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

	client    *http.Client
	transport *http.Transport
	samClient *goSam.Client
}

func (i IGet) samaddress() string {
	return i.samHost + ":" + i.samPort
}

// Do "does" a request and returns a response. It's just a wrapper for http/client.Do
func (i *IGet) Do(req *http.Request) (*http.Response, error) {
	c, e := i.client.Do(req)
	if e != nil {
		return nil, e
	}
	return c, e
}

// DoBytes does a request and returns the body of a response as bytes.
// Eventually, it will be unused in iget, it's here for a hypothetical API and testing.
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
// Eventually, it will be unused in iget, it's here for a hypothetical API and testing.
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
	for _, setter := range setters {
		setter(r)
	}
	return r, nil
}

// PrintResponse routes the output
func (i *IGet) PrintResponse(c *http.Response) string {
	if i.verb {

	}
	if i.outputPath == "-" || i.outputPath == "stdout" {
		b, err := ioutil.ReadAll(c.Body)
		if err != nil {
			return ""
		}
		fmt.Printf("%s", b)
		return string(b)
	} else {
		b, err := ioutil.ReadAll(c.Body)
		if err != nil {
			return ""
		}
		err = ioutil.WriteFile(i.outputPath, b, 0644)
		if err != nil {
			return ""
		}
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
		inboundTunnels:  8,
		outboundTunnels: 2,
		keepAlives:      true,
		idleConns:       4,
		inboundBackups:  2,
		outboundBackups: 2,
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
