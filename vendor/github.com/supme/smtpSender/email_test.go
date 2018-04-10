package smtpSender

import "testing"

type emailField struct {
	input, name, email, domain string
}

var (
	rightEmail []emailField
	badEmail   []emailField
)

func init() {
	rightEmail = append(rightEmail, emailField{" My name   <  my+email@domain.tld.  > ", "My name", "my+email", "domain.tld"})
	rightEmail = append(rightEmail, emailField{"  < My+Email@doMain.tld.  >  ", "", "my+email", "domain.tld"})
	rightEmail = append(rightEmail, emailField{"  mY+eMail@Domain.Tld.   ", "", "my+email", "domain.tld"})

	badEmail = append(badEmail, emailField{input: "my+email@domain.t"})
	badEmail = append(badEmail, emailField{input: "< my+email[at]domain.tld>"})
	//badEmail = append(badEmail, emailField{input: "<my+email@domain.tld."})
}

func TestSplitEmail(t *testing.T) {
	for _, v := range rightEmail {
		name, email, domain, err := splitEmail(v.input)
		if err != nil {
			t.Errorf("Email '%s' has error: %s", v.input, err)
		}
		if v.name != name {
			t.Errorf("Email '%s' not valid name: want '%s', has '%s'", v.input, v.name, name)
		}
		if v.email != email {
			t.Errorf("Email '%s' not valid email: want '%s', has '%s'", v.input, v.email, email)
		}
		if v.domain != domain {
			t.Errorf("Email '%s' not valid name: want '%s', has '%s'", v.input, v.domain, domain)
		}
	}

	for _, v := range badEmail {
		_, _, _, err := splitEmail(v.input)
		if err == nil {
			t.Errorf("Email '%s' has bad format, but parsed without error", v.input)
		}
	}
}

func BenchmarkSplitEmailFullString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		splitEmail(rightEmail[0].input)
	}
}

func BenchmarkSplitEmailOnlyString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		splitEmail(rightEmail[0].input)
	}
}

func BenchmarkSplitEmail(b *testing.B) {
	for n := 0; n < b.N; n++ {
		splitEmail(rightEmail[0].input)
	}
}
