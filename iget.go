// Package iget provides a highly-configurable curl/wget-like HTTP client that
// works exclusively over I2P via the SAM API. It prevents destination reuse
// across different sites, never forwards traffic through a clearnet HTTP proxy,
// and exposes inline I2CP configuration (tunnel counts, lengths, and lifespans)
// for each invocation. The package is also used as the library backend for the
// iget CLI binary.
package iget

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-i2p/onramp"
	"github.com/valyala/fasthttp"
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

	sessionName string

	client    *http.Client
	transport *http.Transport
	garlic    *onramp.Garlic

	lineLength       int
	continueDownload bool
	markSize         int

	// verboseOut receives all library diagnostic output. It is set to
	// io.Discard when verb==false and to os.Stderr when verb==true, so that
	// the process-level os.Stderr is never redirected and cobra's error
	// handler always has a valid file descriptor.
	verboseOut io.Writer
}

// Do "does" a request and returns a response. It's just a wrapper for http/client.Do
func (i *IGet) Do(req *http.Request) (*http.Response, error) {
	i.resolveOutputPath(req.URL.String())
	if i.verb {
		fmt.Fprintf(i.verboseOut, "[iget] %s %s\n", req.Method, req.URL)
	}
	c, e := i.client.Do(req)
	if e != nil {
		return nil, e
	}
	if i.debug {
		fmt.Fprintf(i.verboseOut, "[iget debug] response status: %s\n", c.Status)
		for k, vals := range c.Header {
			for _, v := range vals {
				fmt.Fprintf(i.verboseOut, "[iget debug] header: %s: %s\n", k, v)
			}
		}
	}
	if i.verb {
		fmt.Fprintf(i.verboseOut, "[iget] response: %s\n", c.Status)
	}
	if i.outputPath != "-" && i.outputPath != "stdout" {
		if err := i.truncateIfRangeIgnored(req, c); err != nil {
			c.Body.Close()
			return nil, err
		}
		if err := i.saveToFile(c.Body); err != nil {
			c.Body.Close()
			return nil, err
		}
	}
	return c, e
}

// DoBytes does a request and returns the body of a response as bytes.
// It must be used with the default stdout output path ("-" or "stdout");
// if an explicit file output path has been set via Output(), DoBytes returns
// an error because Do() will have already drained the response body into the
// file before DoBytes can read it.
// It is unused in iget itself; it is part of the library API for callers and testing.
func (i *IGet) DoBytes(req *http.Request) ([]byte, error) {
	if i.outputPath != "-" && i.outputPath != "stdout" {
		return nil, fmt.Errorf("iget: DoBytes requires stdout output (got outputPath=%q); use Do() when writing to a file", i.outputPath)
	}
	c, e := i.Do(req)
	if e != nil {
		return nil, e
	}
	defer c.Body.Close()
	body, err := io.ReadAll(c.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// DoString does a request, converts the body to bytes, then returns the body of a response as a string.
// It must be used with the default stdout output path; see DoBytes for the constraint details.
// It is unused in iget itself; it is part of the library API for callers and testing.
func (i *IGet) DoString(req *http.Request) (string, error) {
	b, e := i.DoBytes(req)
	if e != nil {
		return "", e
	}
	return string(b), nil
}

// Request generates an *http.Request
func (i *IGet) Request(setters ...RequestOption) (*http.Request, error) {
	return i.RequestWithContext(context.Background(), setters...)
}

// RequestWithContext generates an *http.Request with the provided context.
// The context is passed to http.NewRequestWithContext so callers can cancel
// in-flight I2P requests (e.g. for graceful daemon shutdown or pipeline
// cancellation). Use Request() for the common background-context case.
func (i *IGet) RequestWithContext(ctx context.Context, setters ...RequestOption) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, i.method, i.url, io.NopCloser(strings.NewReader(i.body)))
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

// HTTPClient returns the underlying *http.Client used by IGet. It is exposed for advanced users who need to bypass the IGet Request and Do methods and interact directly with the HTTP client, for example to use a custom http.RoundTripper or to inspect client state.
func (i *IGet) HTTPClient() *http.Client {
	return i.client
}

// RoundTrip implements the http.RoundTripper interface by delegating to the IGet's http.Transport. This allows IGet to be used as a custom RoundTripper in contexts that require it, while still applying all of IGet's SAM and tunnel configurations.
func (i *IGet) RoundTrip(req *http.Request) (*http.Response, error) {
	return i.transport.RoundTrip(req)
}

// FasthttpDial returns a fasthttp.DialFunc that dials through the IGet's http.Transport. This allows IGet to be used as a custom dialer in fasthttp clients, enabling SAM and tunnel configurations for fasthttp requests.
func (i *IGet) FasthttpDial() fasthttp.DialFunc {
	return func(addr string) (net.Conn, error) {
		return i.transport.Dial("tcp", addr)
	}
}

// Compile-time assertion that *IGet implements io.Closer.
var _ io.Closer = (*IGet)(nil)

// NewIGet is an IGet Client
func NewIGet(setters ...Option) (*IGet, error) {
	i := &IGet{
		samHost:    "localhost",
		samPort:    "7656",
		verb:       false,
		debug:      false,
		outputPath: "-",

		destLifespan:     6 * 60 * 1000,
		timeoutTime:      6 * 60 * 1000, // 6 minutes — matches --timeout default
		tunnelLength:     3,
		inboundTunnels:   8,
		outboundTunnels:  8,
		keepAlives:       false,
		idleConns:        4,
		inboundBackups:   3,
		outboundBackups:  3,
		continueDownload: true,
		lineLength:       80,
		transport:        nil,
	}
	var err error
	for _, setter := range setters {
		setter(i)
	}
	// Default to an ephemeral, per-invocation session name so that no two runs
	// share the same I2P destination. Users who need a persistent identity (e.g.
	// for session-cookie continuity) can override this via SessionName().
	if i.sessionName == "" {
		i.sessionName = fmt.Sprintf("iget-%d", time.Now().UnixNano())
	}
	samOpts := []string{
		fmt.Sprintf("inbound.length=%d", i.tunnelLength),
		fmt.Sprintf("outbound.length=%d", i.tunnelLength),
		fmt.Sprintf("inbound.quantity=%d", i.inboundTunnels),
		fmt.Sprintf("outbound.quantity=%d", i.outboundTunnels),
		fmt.Sprintf("inbound.backupQuantity=%d", i.inboundBackups),
		fmt.Sprintf("outbound.backupQuantity=%d", i.outboundBackups),
		fmt.Sprintf("i2cp.closeIdleTime=%d", i.destLifespan),
	}
	i.garlic, err = i.newGarlicWithTimeout(samOpts)
	if err != nil {
		return nil, err
	}
	if i.verb || i.debug {
		i.verboseOut = os.Stderr
	} else {
		i.verboseOut = io.Discard
	}
	if i.debug {
		fmt.Fprintf(i.verboseOut, "[iget debug] SAM bridge: %s\n", i.samaddress())
		fmt.Fprintf(i.verboseOut, "[iget debug] SAM opts: %v\n", samOpts)
		fmt.Fprintf(i.verboseOut, "[iget debug] toPort=%q fromPort=%q\n", i.toPort, i.fromPort)
	}
	if i.verb {
		fmt.Fprintf(i.verboseOut, "[iget] SAM bridge: %s  tunnels in=%d out=%d length=%d\n",
			i.samaddress(), i.inboundTunnels, i.outboundTunnels, i.tunnelLength)
	}
	dialContext := i.applyPortOptions()
	if dialContext == nil {
		dialContext = i.garlic.DialContext
	}
	i.transport = &http.Transport{
		DialContext:           dialContext,
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
