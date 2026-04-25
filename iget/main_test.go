package main

import (
	"testing"
)

// TestParseSAMBridgeAddrIPv4 verifies that a standard IPv4 host:port is parsed correctly.
func TestParseSAMBridgeAddrIPv4(t *testing.T) {
	host, port, err := parseSAMBridgeAddr("127.0.0.1:7656")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "127.0.0.1" || port != "7656" {
		t.Errorf("expected 127.0.0.1:7656, got %s:%s", host, port)
	}
}

// TestParseSAMBridgeAddrIPv6 verifies that an IPv6 bracket-notation address is
// parsed correctly. This was previously broken by strings.Split(addr, ":").
func TestParseSAMBridgeAddrIPv6(t *testing.T) {
	host, port, err := parseSAMBridgeAddr("[::1]:7656")
	if err != nil {
		t.Fatalf("unexpected error for IPv6 address: %v", err)
	}
	if host != "::1" || port != "7656" {
		t.Errorf("expected ::1:7656, got %s:%s", host, port)
	}
}

// TestParseSAMBridgeAddrHostname verifies that a hostname:port pair works.
func TestParseSAMBridgeAddrHostname(t *testing.T) {
	host, port, err := parseSAMBridgeAddr("sam.local:7656")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "sam.local" || port != "7656" {
		t.Errorf("expected sam.local:7656, got %s:%s", host, port)
	}
}

// TestParseSAMBridgeAddrInvalid verifies that an invalid address returns an error.
func TestParseSAMBridgeAddrInvalid(t *testing.T) {
	_, _, err := parseSAMBridgeAddr("not-a-valid-addr")
	if err == nil {
		t.Error("expected error for invalid address, got nil")
	}
}
