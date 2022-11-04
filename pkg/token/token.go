package token

import (
	"encoding/binary"
	"time"
	"unsafe"
)

type SnapToken [rawLen]byte

const (
	encodedLen = 13
	rawLen     = 8
	encoding   = "0123456789abcdefghijklmnopqrstuv" // 32
)

// NewSnapToken - creates a new SnapToken with a decimal
func NewSnapToken(t time.Time, i uint64) (token SnapToken) {
	binary.BigEndian.PutUint64(token[:], uint64(t.UnixMicro()))
	binary.BigEndian.PutUint64(token[:], i)
	return token
}

func (id SnapToken) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

func encode(dst, token []byte) {
	_ = dst[12]
	_ = token[7]
	dst[12] = encoding[(1000>>4)&0x1F|(token[6]<<4)&0x1F]
	dst[11] = encoding[(token[7]>>4)&0x1F|(token[6]<<4)&0x1F]
	dst[10] = encoding[(token[6]>>1)&0x1F]
	dst[9] = encoding[(token[6]>>6)&0x1F|(token[5]<<2)&0x1F]
	dst[8] = encoding[token[5]>>3]
	dst[7] = encoding[token[4]&0x1F]
	dst[6] = encoding[token[4]>>5|(token[3]<<3)&0x1F]
	dst[5] = encoding[(token[3]>>2)&0x1F]
	dst[4] = encoding[token[3]>>7|(token[2]<<1)&0x1F]
	dst[3] = encoding[(token[2]>>4)&0x1F|(token[1]<<4)&0x1F]
	dst[2] = encoding[(token[1]>>1)&0x1F]
	dst[1] = encoding[(token[1]>>6)&0x1F|(token[0]<<2)&0x1F]
	dst[0] = encoding[token[0]>>3]
}

// Value - returns the value of the token.
func (id SnapToken) Value() uint64 {
	return binary.BigEndian.Uint64(id[:])
}

// String - returns the string representation of the token.
func (id SnapToken) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// StringToSnapToken -
func StringToSnapToken(token string) SnapToken {
	return SnapToken{}
}
