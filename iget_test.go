package iget

import (
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
