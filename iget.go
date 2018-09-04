package iget

import (
	"io/ioutil"
	"net/http"
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

// Get does a request
func (i *IGet) Get(addr string) ([]byte, error) {
	var b []byte
	var err error
	if c, e := i.client.Get(addr); e != nil {
		return b, e
	} else {
		if b, err = ioutil.ReadAll(c.Body); err != nil {
			return b, err
		} else {
			return b, nil
		}
	}
}

// GetString returns a response as a string
func (i *IGet) GetString(addr string) (string, error) {
	b, e := i.Get(addr)
	if e != nil {
		return "", e
	}
	return string(b), nil
}

// NewIGet is an IGet Client
func NewIGet(setters ...Option) (*IGet, error) {
	i := &IGet{
		samHost: "localhost",
		samPort: "7656",
		verb:    false,
		debug:   true,

		destLifespan:    3000000,
		timeoutTime:     3000000,
		tunnelLength:    3,
		inboundTunnels:  4,
		outboundTunnels: 4,
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
			Dial: i.samClient.Dial,
		}
		i.client = &http.Client{
			Transport: i.transport,
		}
		return i, nil
	}
}
