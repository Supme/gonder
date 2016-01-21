package dkim

import (
	"crypto"
	"testing"
)

func TestNewConf(t *testing.T) {
	conf, err := NewConf("", "selector")
	if err == nil {
		t.Fatal(err)
	}
	conf, err = NewConf("domain", "")
	if err == nil {
		t.Fatal(err)
	}
	conf, err = NewConf("domain", "selector")
	if err != nil {
		t.Fatal(err)
	}
	if x := len(conf); x != 10 {
		t.Fatal(x, conf)
	}
}

func TestValidate(t *testing.T) {
	conf := Conf{}
	if err := conf.Validate(); err == nil {
		t.Fatal(err)
	}
	conf, err := NewConf("domain", "selector")
	if err != nil {
		t.Fatal(err)
	}
	if err := conf.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAlgorithm(t *testing.T) {
	if x := (Conf{}).Algorithm(); x != AlgorithmSHA256 {
		t.Fatal(x)
	}
}

func TestHash(t *testing.T) {
	if x := (Conf{}).Hash(); x != crypto.SHA256 {
		t.Fatal(x)
	}
}

func TestRelaxedHeader(t *testing.T) {
	conf := Conf{}
	if x := conf.RelaxedHeader(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed/simple"}
	if x := conf.RelaxedHeader(); !x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple/relaxed"}
	if x := conf.RelaxedHeader(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed"}
	if x := conf.RelaxedHeader(); !x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple/simple"}
	if x := conf.RelaxedHeader(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple"}
	if x := conf.RelaxedHeader(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed/relaxed"}
	if x := conf.RelaxedHeader(); !x {
		t.Fatal(x)
	}
}

func TestRelaxedBody(t *testing.T) {
	conf := Conf{}
	if x := conf.RelaxedBody(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed/simple"}
	if x := conf.RelaxedBody(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple/relaxed"}
	if x := conf.RelaxedBody(); !x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed"}
	if x := conf.RelaxedBody(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple/simple"}
	if x := conf.RelaxedBody(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "simple"}
	if x := conf.RelaxedBody(); x {
		t.Fatal(x)
	}
	conf = Conf{CanonicalizationKey: "relaxed/relaxed"}
	if x := conf.RelaxedBody(); !x {
		t.Fatal(x)
	}
}

func TestString(t *testing.T) {
	conf := Conf{}
	if x := conf.String(); x != "" {
		t.Fatal(x)
	}
	conf, err := NewConf("domain", "selector")
	if err != nil {
		t.Fatal(err)
	}
	ts := conf[TimestampKey]
	join := conf.String()
	if join != "v=1; a=rsa-sha256; c=relaxed/simple; d=domain; q=dns/txt; s=selector; t="+ts+"; bh=; h=; b=" {
		t.Fatal(join)
	}
}
