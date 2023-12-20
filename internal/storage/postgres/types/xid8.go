package types

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

type XID8 pguint64

func (xid8 *XID8) Set(src interface{}) error {
	return (*pguint64)(xid8).Set(src)
}

func (xid8 XID8) Get() interface{} {
	return (pguint64)(xid8).Get()
}

func (xid8 *XID8) AssignTo(dst interface{}) error {
	return (*pguint64)(xid8).AssignTo(dst)
}

func (xid8 *XID8) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	return (*pguint64)(xid8).DecodeText(ci, src)
}

func (xid8 *XID8) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return (*pguint64)(xid8).DecodeBinary(ci, src)
}

func (xid8 XID8) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return (pguint64)(xid8).EncodeText(ci, buf)
}

func (xid8 XID8) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return (pguint64)(xid8).EncodeBinary(ci, buf)
}

func (xid8 *XID8) Scan(src interface{}) error {
	return (*pguint64)(xid8).Scan(src)
}

func (xid8 XID8) Value() (driver.Value, error) {
	return (pguint64)(xid8).Value()
}
