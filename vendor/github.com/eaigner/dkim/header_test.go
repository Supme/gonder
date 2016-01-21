package dkim

import (
	"testing"
)

func TestParseHeader(t *testing.T) {
	h := Header("\tX: A b ")
	if x := h.Component(0); x != "\tX" {
		t.Fatal(x)
	}
	if x := h.Component(1); x != " A b " {
		t.Fatal(x)
	}
}

func TestKey(t *testing.T) {
	h := Header(" C\t:A b")
	if x := h.Key(false); x != "C" {
		t.Fatal(x)
	}
}

func TestValue(t *testing.T) {
	h := Header(" C\t:\t A B ")
	if x := h.Value(false); x != "A B" {
		t.Fatal(x)
	}
}

func TestRelaxedKey(t *testing.T) {
	h := Header(" C\t:A b")
	if x := h.Key(true); x != "c" {
		t.Fatal(x)
	}
}

func TestRelaxedValue(t *testing.T) {
	h := Header(" C\t: \tA \t   b ")
	if x := h.Value(true); x != "A b" {
		t.Fatal(x)
	}
}

func TestCanonical(t *testing.T) {
	h := Header(" C\t: \tA \t   b ")
	if x := h.Canonical(true); x != "c:A b" {
		t.Fatal(x)
	}
	k := " C\t"
	v := " \tA \t   b "
	h = Header(k + ":" + v)
	if x := h.Canonical(false); x != k+":"+v {
		t.Fatal(x)
	}
}
