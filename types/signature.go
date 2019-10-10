package types

import crypto "github.com/tepleton/go-crypto"

// Standard Signature
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	crypto.Signature `json:"signature"`
	Sequence         int64 `json:"sequence"`
}
