package iget

import (
	"log"
	"testing"
)

func TestIGet(t *testing.T) {
	if ic, e := NewIGet(Debug(true)); e != nil {
		t.Fatal(e.Error())
	} else {
		if b, e := ic.Get("http://i2p-projekt.i2p"); e != nil {
			t.Fatal(e.Error())
		} else {
			log.Println(b)
		}
	}
}
