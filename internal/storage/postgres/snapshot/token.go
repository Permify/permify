package snapshot

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"

	"github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/token"
)

type (
	// Token - Structure for Token
	Token struct {
		Value    postgres.XID8
		Snapshot string
	}
	// EncodedToken - Structure for EncodedToken
	EncodedToken struct {
		Value string
	}
)

// NewToken - Creates a new snapshot token
func NewToken(value postgres.XID8, snapshot string) token.SnapToken {
	return Token{
		Value:    value,
		Snapshot: snapshot,
	}
}

// Encode - Encodes the token to a string
func (t Token) Encode() token.EncodedSnapToken {
	if t.Snapshot != "" {
		// New format: "xid:snapshot" as single base64
		combined := fmt.Sprintf("%d:%s", t.Value.Uint, t.Snapshot)
		encoded := base64.StdEncoding.EncodeToString([]byte(combined))
		return EncodedToken{
			Value: encoded,
		}
	}

	// Legacy format: binary encoded xid (for backward compatibility)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, t.Value.Uint)
	valueEncoded := base64.StdEncoding.EncodeToString(b)

	return EncodedToken{
		Value: valueEncoded,
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
	// Decode the base64 string
	b, err := base64.StdEncoding.DecodeString(t.Value)
	if err != nil {
		return nil, err
	}

	// Check if it's legacy binary format (8 bytes)
	if len(b) == 8 {
		// Legacy format: binary encoded xid
		return Token{
			Value: postgres.XID8{
				Uint:   binary.LittleEndian.Uint64(b),
				Status: pgtype.Present,
			},
			Snapshot: "", // Empty for backward compatibility
		}, nil
	}

	// New format: "xid:snapshot" as string
	decodedStr := string(b)
	parts := strings.Split(decodedStr, ":")
	if len(parts) >= 2 {
		// New format: "xid:snapshot"
		xidStr := parts[0]
		snapshot := strings.Join(parts[1:], ":") // Rejoin in case snapshot contains colons

		xid, err := strconv.ParseUint(xidStr, 10, 64)
		if err != nil {
			return nil, err
		}

		return Token{
			Value: postgres.XID8{
				Uint:   xid,
				Status: pgtype.Present,
			},
			Snapshot: snapshot,
		}, nil
	}

	// This should never happen with current formats, but handle gracefully
	return nil, fmt.Errorf("invalid token format")
}

// Decode decodes the token from a string
func (t EncodedToken) String() string {
	return t.Value
}
