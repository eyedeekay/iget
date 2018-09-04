package iget

import (
	"crypto/tls"
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

func (i *IGet) Do(req *http.Request) (*http.Response, error) {
	if c, e := i.client.Do(req); e != nil {
		return nil, e
	} else {
		return c, e
	}
}

func (i *IGet) DoBytes(req *http.Request) ([]byte, error) {
	var b []byte
	var err error
	if c, e := i.Do(req); e != nil {
		return b, e
	} else {
		if b, err = ioutil.ReadAll(c.Body); err != nil {
			return b, err
		} else {
			return b, nil
		}
	}
}

func (i *IGet) DoString(req *http.Request) (string, error) {
	if b, e := i.DoBytes(req); e != nil {
		return "", e
	} else {
		return string(b), nil
	}
}

// Request generates an *http.Request
func (i *IGet) Request(setters ...RequestOption) (*http.Request, error) {
	if r, err := http.NewRequest(i.method, i.url, ioutil.NopCloser(strings.NewReader(i.body))); err != nil {
		return nil, err
	} else {
		for _, setter := range setters {
			setter(r)
		}
		return r, nil
	}
}

// NewIGet is an IGet Client
func NewIGet(setters ...Option) (*IGet, error) {
	i := &IGet{
		samHost: "localhost",
		samPort: "7656",
		verb:    false,
		debug:   false,

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
	if i.samClient, err = goSam.NewClientFromOptions(
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
	); err != nil {
		return nil, err
	} else {
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
}
