package types

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestPgUint64(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		var p pguint64

		err := p.Set(int64(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Set(int64(-1))
		assert.Error(t, err)

		err = p.Set(int32(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Set(int32(-1))
		assert.Error(t, err)

		err = p.Set(uint32(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Set(uint64(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Set("invalid")
		assert.Error(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		var p pguint64

		p.Status = pgtype.Present
		p.Uint = 42
		assert.Equal(t, uint64(42), p.Get())

		p.Status = pgtype.Null
		assert.Nil(t, p.Get())

		p.Status = pgtype.Undefined
		assert.Equal(t, pgtype.Undefined, p.Get())
	})

	t.Run("AssignTo", func(t *testing.T) {
		p := pguint64{Uint: 42, Status: pgtype.Present}

		var u64 uint64
		err := p.AssignTo(&u64)
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), u64)

		var pu64 *uint64
		err = p.AssignTo(&pu64)
		assert.NoError(t, err)
		assert.NotNil(t, pu64)
		assert.Equal(t, uint64(42), *pu64)

		p.Status = pgtype.Null
		err = p.AssignTo(&pu64)
		assert.NoError(t, err)
		assert.Nil(t, pu64)

		p.Status = pgtype.Undefined
		err = p.AssignTo(&pu64)
		assert.NoError(t, err)
	})

	t.Run("DecodeText", func(t *testing.T) {
		var p pguint64

		err := p.DecodeText(nil, []byte("42"))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.DecodeText(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, pgtype.Null, p.Status)
	})

	t.Run("EncodeText", func(t *testing.T) {
		p := pguint64{Uint: 42, Status: pgtype.Present}

		buf, err := p.EncodeText(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "42", string(buf))

		p.Status = pgtype.Null
		buf, err = p.EncodeText(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, buf)

		p.Status = pgtype.Undefined
		buf, err = p.EncodeText(nil, nil)
		assert.Error(t, err)
	})

	t.Run("EncodeBinary", func(t *testing.T) {
		p := pguint64{Uint: 42, Status: pgtype.Present}

		buf, err := p.EncodeBinary(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 42}, buf)

		p.Status = pgtype.Null
		buf, err = p.EncodeBinary(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, buf)

		p.Status = pgtype.Undefined
		buf, err = p.EncodeBinary(nil, nil)
		assert.Error(t, err)
	})

	t.Run("Scan", func(t *testing.T) {
		var p pguint64

		err := p.Scan(uint32(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Scan(int64(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Scan(uint64(42))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Scan("42")
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Scan([]byte("42"))
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), p.Uint)
		assert.Equal(t, pgtype.Present, p.Status)

		err = p.Scan(nil)
		assert.NoError(t, err)
		assert.Equal(t, pgtype.Null, p.Status)
	})

	t.Run("Value", func(t *testing.T) {
		p := pguint64{Uint: 42, Status: pgtype.Present}

		value, err := p.Value()
		assert.NoError(t, err)
		assert.Equal(t, int64(42), value)

		p.Status = pgtype.Null
		value, err = p.Value()
		assert.NoError(t, err)
		assert.Nil(t, value)

		p.Status = pgtype.Undefined
		value, err = p.Value()
		assert.Error(t, err)
	})
}
