package types

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgio"
	"github.com/jackc/pgtype"
)

type pguint64 struct {
	Uint   uint64
	Status pgtype.Status
}

func (p *pguint64) Set(src interface{}) error {
	switch value := src.(type) {
	case int64:
		if value < 0 {
			return fmt.Errorf("%d is less than minimum value for pguint64", value)
		}
		*p = pguint64{Uint: uint64(value), Status: pgtype.Present}
	case int32:
		if value < 0 {
			return fmt.Errorf("%d is less than minimum value for pguint64", value)
		}
		*p = pguint64{Uint: uint64(value), Status: pgtype.Present}
	case uint32:
		*p = pguint64{Uint: uint64(value), Status: pgtype.Present}
	case uint64:
		*p = pguint64{Uint: value, Status: pgtype.Present}
	default:
		return fmt.Errorf("cannot convert %v to pguint64", value)
	}

	return nil
}

func (p pguint64) Get() interface{} {
	switch p.Status {
	case pgtype.Present:
		return p.Uint
	case pgtype.Null:
		return nil
	default:
		return p.Status
	}
}

func (p *pguint64) AssignTo(dst interface{}) error {
	switch v := dst.(type) {
	case *uint64:
		if p.Status == pgtype.Present {
			*v = p.Uint
		} else {
			return fmt.Errorf("cannot assign %v into %T", p, dst)
		}
	case **uint64:
		if p.Status == pgtype.Present {
			n := p.Uint
			*v = &n
		} else {
			*v = nil
		}
	}

	return nil
}

func (p *pguint64) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*p = pguint64{Status: pgtype.Null}
		return nil
	}

	n, err := strconv.ParseUint(string(src), 10, 64)
	if err != nil {
		return err
	}

	*p = pguint64{Uint: n, Status: pgtype.Present}
	return nil
}

func (p *pguint64) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*p = pguint64{Status: pgtype.Null}
		return nil
	}

	if len(src) != 4 {
		return fmt.Errorf("invalid length: %v", len(src))
	}

	n := binary.BigEndian.Uint64(src)
	*p = pguint64{Uint: n, Status: pgtype.Present}
	return nil
}

func (p pguint64) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch p.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errors.New("encode text status undefined status")
	}

	return append(buf, strconv.FormatUint(p.Uint, 10)...), nil
}

func (p pguint64) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch p.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errors.New("encode text status undefined status")
	}

	return pgio.AppendUint64(buf, p.Uint), nil
}

// Scan implements the database/sql Scanner interface.
func (p *pguint64) Scan(src interface{}) error {
	if src == nil {
		*p = pguint64{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case uint32:
		*p = pguint64{Uint: uint64(src), Status: pgtype.Present}
		return nil
	case int64:
		*p = pguint64{Uint: uint64(src), Status: pgtype.Present}
		return nil
	case uint64:
		*p = pguint64{Uint: src, Status: pgtype.Present}
		return nil
	case string:
		return p.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return p.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (p pguint64) Value() (driver.Value, error) {
	switch p.Status {
	case pgtype.Present:
		return int64(p.Uint), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errors.New("encode text status undefined status")
	}
}
