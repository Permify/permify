package snapshot

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/jackc/pgtype"

	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/pkg/token"
)

type (
	// Token - Structure for Token
	Token struct {
		Value types.XID8
	}
	// EncodedToken - Structure for EncodedToken
	EncodedToken struct {
		Value string
	}
)

// NewToken - Creates a new snapshot token
func NewToken(value types.XID8) token.SnapToken {
	return Token{
		Value: value,
	}
}

// Encode - Encodes the token to a string
func (t Token) Encode() token.EncodedSnapToken {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, t.Value.Uint)
	return EncodedToken{
		Value: base64.StdEncoding.EncodeToString(b),
	}
}

// Eg snapshot is equal to given snapshot
func (t Token) Eg(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value.Uint == ct.Value.Uint
}

// Gt snapshot is greater than given snapshot
func (t Token) Gt(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value.Uint > ct.Value.Uint
}

// Lt snapshot is less than given snapshot
func (t Token) Lt(token token.SnapToken) bool {
	ct, ok := token.(Token)
	return ok && t.Value.Uint < ct.Value.Uint
}

// Decode decodes the token from a string
func (t EncodedToken) Decode() (token.SnapToken, error) {
	b, err := base64.StdEncoding.DecodeString(t.Value)
	if err != nil {
		return nil, err
	}
	return Token{
		Value: types.XID8{
			Uint:   binary.LittleEndian.Uint64(b),
			Status: pgtype.Present,
		},
	}, nil
}

// Decode decodes the token from a string
func (t EncodedToken) String() string {
	return t.Value
}
