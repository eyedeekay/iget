package iget

import (
	"log"
	"testing"
)

func TestIGet(t *testing.T) {
	if ic, e := NewIGet(Length(1), URL("http://i2p-projekt.i2p"), Inbound(15), Debug(true)); e != nil {
		t.Fatal(e.Error())
	} else {
		if r, e := ic.Request(); e != nil {
			t.Fatal(e.Error())
		} else {
			if b, e := ic.Do(r); e != nil {
				t.Fatal(e.Error())
			} else {
                c := ic.PrintResponse(b)
			}
		}
	}
}
