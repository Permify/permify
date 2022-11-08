package snapshot

import (
	"time"

	"github.com/Permify/permify/pkg/token"
)

type (
	Token struct {
		Value time.Time
	}
	EncodedToken struct {
		Value string
	}
)

// Encode encodes the token to a string
func (t Token) Encode() token.EncodedSnapToken {
	return nil
}

// Eg token is equal to given token
func (t Token) Eg(token token.SnapToken) bool {
	return false
}

// Gt snapshot is greater than given snapshot
func (t Token) Gt(token token.SnapToken) bool {
	return false
}

// Lt snapshot is less than given snapshot
func (t Token) Lt(token token.SnapToken) bool {
	return false
}

// Decode decodes the token from a string
func (t EncodedToken) Decode() (token.SnapToken, error) {
	return nil, nil
}

// Decode decodes the token from a string
func (t EncodedToken) String() string {
	return t.Value
}
