package dkim

import (
	"testing"
)

var utilSampleEML []byte = []byte("A: X\r\n" +
	"B : Y\t\r\n" +
	"\tZ  \r\n" +
	"\r\n" +
	" C \r\n" +
	"D \t E\r\n" +
	"\r\n" +
	"\r\n")

func TestSplitEML(t *testing.T) {
	header, body, err := splitEML(utilSampleEML)
	if err != nil {
		t.Fatal("error not nil", err)
	}
	if len(header) == 0 {
		t.Fatal("wrong header length", len(header))
	}
	if len(body) == 0 {
		t.Fatal("wrong body length", len(body))
	}
	if x := string(header); x != "A: X\r\nB : Y\t\r\n\tZ  " {
		t.Fatal("wrong header", x)
	}
	if x := string(body); x != " C \r\nD \t E\r\n\r\n\r\n" {
		t.Fatal("wrong body", x)
	}
}
