package dkim

type Signer interface {
	Sign(eml []byte) (signed []byte, err error)
}
