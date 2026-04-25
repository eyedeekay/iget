package iget

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-i2p/onramp"
)

func (i IGet) samaddress() string {
	return i.samHost + ":" + i.samPort
}

// applyPortOptions returns a DialContext func that incorporates SAM virtual port
// settings when toPort or fromPort are configured, or when the URL scheme allows
// a virtual port to be inferred (https → 443, http → 80). When neither is set
// and no inference is possible it returns nil so the caller can fall back to the
// plain garlic.DialContext method.
func (i *IGet) applyPortOptions() func(ctx context.Context, network, addr string) (net.Conn, error) {
	effectiveToPort := i.toPort
	if effectiveToPort == "" && i.url != "" {
		// Infer virtual port from URL scheme so that multi-port I2P destinations
		// (e.g. serving HTTP on port 80 and HTTPS on port 443) are contacted on
		// the correct virtual port without requiring an explicit --to-port flag.
		if strings.HasPrefix(i.url, "https://") {
			effectiveToPort = "443"
		} else if strings.HasPrefix(i.url, "http://") {
			effectiveToPort = "80"
		}
	}
	if effectiveToPort == "" && i.fromPort == "" {
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
		if effectiveToPort != "" {
			// Encode destination virtual port into the address as "dest.i2p:port".
			addr = net.JoinHostPort(host, effectiveToPort)
		}
		// When effectiveToPort is empty, addr retains its original host:httpPort form so
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

// newGarlicWithTimeout creates a SAM garlic session, aborting if the
// configured timeout elapses before the session is established. This prevents
// NewIGet and Reset from blocking indefinitely when the SAM bridge is
// unreachable or slow.
func (i *IGet) newGarlicWithTimeout(samOpts []string) (*onramp.Garlic, error) {
	type result struct {
		g   *onramp.Garlic
		err error
	}
	ch := make(chan result, 1)
	go func() {
		var g *onramp.Garlic
		var err error
		if i.username != "" || i.password != "" {
			g, err = onramp.NewGarlicWithAuth(i.sessionName, i.samaddress(), i.username, i.password, samOpts)
		} else {
			g, err = onramp.NewGarlic(i.sessionName, i.samaddress(), samOpts)
		}
		ch <- result{g, err}
	}()
	timeout := time.Duration(i.timeoutTime) * time.Millisecond
	if timeout <= 0 {
		timeout = 6 * time.Minute
	}
	select {
	case r := <-ch:
		return r.g, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("iget: SAM session setup exceeded timeout (%v)", timeout)
	}
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

// Reset tears down the current SAM session and rebuilds it from scratch.
// It is intended for use in retry loops where the underlying session may have
// become unhealthy (e.g., after an I2P router restart or tunnel expiry).
// The IGet configuration (tunnel counts, timeout, ports, etc.) is preserved.
func (i *IGet) Reset() error {
	if err := i.Close(); err != nil {
		// Log but do not return — a close error should not block recreation.
		fmt.Fprintf(i.verboseOut, "[iget] warning: Close during Reset: %v\n", err)
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
	garlic, err := i.newGarlicWithTimeout(samOpts)
	if err != nil {
		return fmt.Errorf("iget: Reset: SAM session rebuild failed: %w", err)
	}
	i.garlic = garlic
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
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	i.client = &http.Client{
		Transport: i.transport,
	}
	return nil
}
