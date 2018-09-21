package gopl

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestDataFrameBody(t *testing.T) {
	df := NewDataFrame()
	df.SetBody(strings.NewReader("ABC123"))

	txt := func() string {
		r := df.GetBody()
		b := bytes.NewBuffer(make([]byte, 0))
		io.Copy(b, r)
		return b.String()
	}
	if "ABC123" != txt() {
		t.Fatal("Not match, was: " + txt())
	}
	if "ABC123" != txt() {
		t.Fatal("Not match, was: " + txt())
	}
	if "ABC123" != txt() {
		t.Fatal("Not match, was: " + txt())
	}
}
