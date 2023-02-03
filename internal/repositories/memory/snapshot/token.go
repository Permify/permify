package snapshot

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/Permify/permify/pkg/token"
)

type (
	// Token - Structure for Token
	Token struct {
		Value uint64
	}
	// EncodedToken - Structure for EncodedToken
	EncodedToken struct {
		Value string
	}
)

// NewToken - Creates a new snapshot token
func NewToken(value time.Time) token.SnapToken {
	return Token{
		Value: uint64(value.UnixNano()),
	}
}

// Encode - Encodes the token to a string
func (t Token) Encode() token.EncodedSnapToken {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, t.Value)
	return EncodedToken{
		Value: base64.StdEncoding.EncodeToString(b),
	}
}

// Eg token is equal to given token
func (t Token) Eg(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value == ct.Value
}

// Gt snapshot is greater than given snapshot
func (t Token) Gt(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value > ct.Value
}

// Lt snapshot is less than given snapshot
func (t Token) Lt(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value < ct.Value
}

// Decode decodes the token from a string
func (t EncodedToken) Decode() (token.SnapToken, error) {
	b, err := base64.StdEncoding.DecodeString(t.Value)
	if err != nil {
		return nil, err
	}
	return Token{
		Value: binary.LittleEndian.Uint64(b),
	}, nil
}

// Decode decodes the token from a string
func (t EncodedToken) String() string {
	return t.Value
}
