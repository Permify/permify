package postgres

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

// XID8 represents a PostgreSQL xid8 (64-bit transaction ID) type
type XID8 pguint64

// pgSnapshot represents a PostgreSQL snapshot type
type pgSnapshot struct {
	pgtype.Value
}

// SnapshotCodec handles encoding/decoding of PostgreSQL snapshot values
type SnapshotCodec struct{}

// Uint64Codec handles encoding/decoding of PostgreSQL xid8 values
type Uint64Codec struct{}

// Set sets the XID8 value from various input types
func (x *XID8) Set(src interface{}) error {
	return (*pguint64)(x).Set(src)
}

// Get returns the underlying value
func (x XID8) Get() interface{} {
	return (pguint64)(x).Get()
}

// AssignTo assigns the value to the destination
func (x *XID8) AssignTo(dst interface{}) error {
	return (*pguint64)(x).AssignTo(dst)
}

// DecodeText decodes text format
func (x *XID8) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	return (*pguint64)(x).DecodeText(ci, src)
}

// DecodeBinary decodes binary format
func (x *XID8) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return (*pguint64)(x).DecodeBinary(ci, src)
}

// EncodeText encodes to text format
func (x XID8) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return (pguint64)(x).EncodeText(ci, buf)
}

// EncodeBinary encodes to binary format
func (x XID8) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return (pguint64)(x).EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface
func (x *XID8) Scan(src interface{}) error {
	return (*pguint64)(x).Scan(src)
}

// Value implements the database/sql/driver Valuer interface
func (x XID8) Value() (driver.Value, error) {
	return (pguint64)(x).Value()
}

// FormatCode returns the format code for snapshot values
func (c SnapshotCodec) FormatCode() int16 {
	return pgtype.BinaryFormatCode
}

// FormatCode returns the format code for xid8 values
func (c Uint64Codec) FormatCode() int16 {
	return pgtype.BinaryFormatCode
}
