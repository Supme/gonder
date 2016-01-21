package dkim

import (
	"testing"
)

var headerListSample []byte = []byte("A: X\r\n" +
	"B : Y\t\r\n" +
	"\tZ  \r\n" +
	"\r\n" +
	" C \r\n" +
	"D \t E\r\n" +
	"\r\n" +
	"\r\n")

func TestParseHeaderList(t *testing.T) {
	header, _, err := splitEML(headerListSample)
	if err != nil {
		t.Fatal("error not nil", err)
	}
	list := ParseHeaderList(header)
	if x := len(list); x != 2 {
		t.Fatal(x)
	}
	if x := list[0]; x != "A: X" {
		t.Fatal(x)
	}
	if x := list[1]; x != "B : Y\t\r\n\tZ  " {
		t.Fatal(x)
	}
}

func TestGet(t *testing.T) {
	list := HeaderList{
		Header(" A\t:X y"),
		Header(" b\t: z"),
	}
	if x := list.Get("A").Component(1); x != "X y" {
		t.Fatal(x)
	}
	if x := list.Get("B").Component(1); x != " z" {
		t.Fatal(x)
	}
	if x := list.Get("C"); x != "" {
		t.Fatal(x)
	}
}

func TestFields(t *testing.T) {
	list := HeaderList{
		Header(" A\t:X y"),
		Header(" b\t: z"),
	}

	if x := list.Fields(); x != "A:b" {
		t.Fatal(x)
	}
}

func TestCanonical2(t *testing.T) {
	header, _, err := splitEML(headerListSample)
	if err != nil {
		t.Fatal(err)
	}
	list := ParseHeaderList(header)
	if x := len(list); x != 2 {
		t.Fatal(x)
	}
	simple := list.Canonical(false)
	if simple != "A: X\r\nB : Y\t\r\n\tZ  \r\n" {
		t.Fatal(simple)
	}
	relaxed := list.Canonical(true)
	if relaxed != "a:X\r\nb:Y Z\r\n" {
		t.Fatal(relaxed)
	}
}
