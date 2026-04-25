// Package iget provides a highly-configurable curl/wget-like HTTP client that
// works exclusively over I2P via the SAM API. It prevents destination reuse
// across different sites, never forwards traffic through a clearnet HTTP proxy,
// and exposes inline I2CP configuration (tunnel counts, lengths, and lifespans)
// for each invocation. The package is also used as the library backend for the
// iget CLI binary.
package iget

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-i2p/onramp"
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

func (i IGet) samaddress() string {
	return i.samHost + ":" + i.samPort
}

// applyPortOptions returns a DialContext func that incorporates SAM virtual port
// settings when toPort or fromPort are configured. When neither is set it returns
// nil so the caller can fall back to the plain garlic.DialContext method.
func (i *IGet) applyPortOptions() func(ctx context.Context, network, addr string) (net.Conn, error) {
	if i.toPort == "" && i.fromPort == "" {
		return nil
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// addr arrives as "dest.b32.i2p:httpPort" from http.Transport.
		// Strip the HTTP port so we can substitute the SAM virtual port.
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// addr has no port component; use it as-is.
			host = addr
		}
		if i.toPort != "" {
			// Encode destination virtual port into the address as "dest.i2p:port".
			addr = net.JoinHostPort(host, i.toPort)
		}
		// When toPort is empty, addr retains its original host:httpPort form so
		// that DialContextToPort receives a well-formed address string.
		if i.fromPort != "" {
			fp, err := net.LookupPort("tcp", i.fromPort)
			if err != nil {
				return nil, fmt.Errorf("iget: invalid FROM_PORT %q: %w", i.fromPort, err)
			}
			return i.garlic.DialContextToPort(ctx, network, addr, fp)
		}
		return i.garlic.DialContextToPort(ctx, network, addr)
	}
}

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

// HTTPClient returns the underlying *http.Client used by IGet. It is exposed for advanced users who need to bypass the IGet Request and Do methods and interact directly with the HTTP client, for example to use a custom http.RoundTripper or to inspect client state.
func (i *IGet) HTTPClient() *http.Client {
	return i.client
}

// RoundTrip implements the http.RoundTripper interface by delegating to the IGet's http.Transport. This allows IGet to be used as a custom RoundTripper in contexts that require it, while still applying all of IGet's SAM and tunnel configurations.
func (i *IGet) RoundTrip(req *http.Request) (*http.Response, error) {
	return i.transport.RoundTrip(req)
}

// Close releases the underlying SAM session and closes all idle transport connections.
// It must be called when the IGet client is no longer needed.
func (i *IGet) Close() error {
	if i.transport != nil {
		i.transport.CloseIdleConnections()
	}
	if i.garlic != nil {
		return i.garlic.Close()
	}
	return nil
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
	if i.username != "" || i.password != "" {
		i.garlic, err = onramp.NewGarlicWithAuth(i.sessionName, i.samaddress(), i.username, i.password, samOpts)
	} else {
		i.garlic, err = onramp.NewGarlic(i.sessionName, i.samaddress(), samOpts)
	}
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
		Timeout:   time.Duration(i.timeoutTime) * time.Millisecond,
	}
	return i, nil
}
