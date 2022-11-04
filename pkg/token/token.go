package token

import (
	`crypto/rand`
	`fmt`
	"time"
	"unsafe"
)

type SnapToken [rawLen]byte

const (
	encodedLen = 20
	rawLen     = 15
	encoding   = "0123456789abcdefghijklmnopqrstuv" // 32
)

//var (
//	objectIDCounter = randInt()
//)

func New(value uint64) (token SnapToken) {
	return NewWithTime(time.Now(), value)
}

// NewWithTime - creates a new SnapToken with a decimal
func NewWithTime(t time.Time, value uint64) (token SnapToken) {
	// TODO
	return token
}

func (id SnapToken) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

func encode(dst, token []byte) {
	// TODO
}

// Value - returns the value of the token.
func (id SnapToken) Value() uint64 {
	// TODO
	return 0
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

// randInt generates a random uint32
func randInt() uint32 {
	b := make([]byte, 3)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("snaphot: cannot generate random number: %v", err))
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}
