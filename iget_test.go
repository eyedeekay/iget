package iget

import (
	"os"
	"testing"
)

func TestIGet(t *testing.T) {
	if ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	); e != nil {
		t.Fatal(e.Error())
	} else {
		defer ic.Close()
		if r, e := ic.Request(); e != nil {
			t.Fatal(e.Error())
		} else {
			if b, e := ic.Do(r); e != nil {
				t.Fatal(e.Error())
			} else {
				ic.PrintResponse(b)
			}
		}
	}
}

func TestIGetFile(t *testing.T) {
	if ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
		Output("file.html"),
	); e != nil {
		t.Fatal(e.Error())
	} else {
		defer ic.Close()
		t.Cleanup(func() { os.Remove("file.html") })
		if r, e := ic.Request(); e != nil {
			t.Fatal(e.Error())
		} else {
			if b, e := ic.Do(r); e != nil {
				t.Fatal(e.Error())
			} else {
				ic.PrintResponse(b)
			}
		}
	}
}

func TestDoBytes(t *testing.T) {
	ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	r, e := ic.Request()
	if e != nil {
		t.Fatal(e.Error())
	}
	b, e := ic.DoBytes(r)
	if e != nil {
		t.Fatal(e.Error())
	}
	if len(b) == 0 {
		t.Error("DoBytes returned empty response body")
	}
}

func TestDoString(t *testing.T) {
	ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	r, e := ic.Request()
	if e != nil {
		t.Fatal(e.Error())
	}
	s, e := ic.DoString(r)
	if e != nil {
		t.Fatal(e.Error())
	}
	if s == "" {
		t.Error("DoString returned empty response body")
	}
}

func TestLineLengthZero(t *testing.T) {
	ic, e := NewIGet(
		LineLength(0),
		URL("http://i2p-projekt.i2p"),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	// Verify that constructing with LineLength(0) succeeds (no panic expected in PrintResponse).
	if ic.lineLength != 0 {
		t.Errorf("expected lineLength=0, got %d", ic.lineLength)
	}
}
