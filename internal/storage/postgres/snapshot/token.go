package snapshot

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"slices"
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

// NewToken creates a new snapshot token with proper MVCC visibility.
// It automatically processes the snapshot to ensure each transaction has a unique snapshot.
func NewToken(value postgres.XID8, snapshot string) token.SnapToken {
	t := Token{
		Value:    value,
		Snapshot: snapshot,
	}
	t.Snapshot = t.createFinalSnapshot()
	return t
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
			Snapshot: "",
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

// createFinalSnapshot creates a final snapshot by adding the current transaction ID to the XIP list.
// This ensures each concurrent transaction gets a unique snapshot for proper MVCC visibility.
func (t Token) createFinalSnapshot() string {
	if t.Snapshot == "" {
		return ""
	}

	parts := strings.SplitN(strings.TrimSpace(t.Snapshot), ":", 3)
	if len(parts) < 2 {
		return t.Snapshot
	}

	xmin, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return t.Snapshot
	}
	xmax, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return t.Snapshot
	}

	txid := t.Value.Uint

	// Parse existing XIPs
	xipList := []uint64{}
	if len(parts) == 3 && parts[2] != "" {
		for _, xipStr := range strings.Split(parts[2], ",") {
			xipStr = strings.TrimSpace(xipStr)
			if xipStr == "" {
				continue
			}
			xip, err := strconv.ParseUint(xipStr, 10, 64)
			if err != nil {
				return t.Snapshot
			}
			xipList = append(xipList, xip)
		}
	}

	// Sort XIPs to ensure deterministic snapshot encoding and maintain PostgreSQL invariants
	// Per PostgreSQL semantics, xip[] must contain only XIDs with xmin ≤ xip < xmax
	slices.Sort(xipList)

	// Add current txid to make snapshot unique for this transaction
	// Insert in sorted order to maintain consistency
	inserted := false
	for i, xip := range xipList {
		if xip == txid {
			// Already in list, don't add again
			inserted = true
			break
		}
		if xip > txid {
			// Insert at position i
			xipList = append(xipList[:i], append([]uint64{txid}, xipList[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		// Append to end
		xipList = append(xipList, txid)
	}

	// Adjust xmax if necessary
	newXmax := xmax
	if txid >= newXmax {
		newXmax = txid + 1
	}

	// Adjust xmin if current txid is smaller
	newXmin := xmin
	if txid < newXmin {
		newXmin = txid
	}

	// Build the result snapshot string efficiently using strings.Builder
	var xipStrs []string
	for _, xip := range xipList {
		xipStrs = append(xipStrs, fmt.Sprintf("%d", xip))
	}

	return fmt.Sprintf("%d:%d:%s", newXmin, newXmax, strings.Join(xipStrs, ","))
}
