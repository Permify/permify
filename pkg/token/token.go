package token

type EncodedSnapToken interface {
	// String returns the string representation of the token.
	String() string
	// Decode decodes the token from a string
	Decode() (SnapToken, error)
}

type SnapToken interface {
	// Encode encodes the token to a string.
	Encode() EncodedSnapToken
	// Eg snapshot is equal to given snapshot
	Eg(token SnapToken) bool
	// Gt snapshot is greater than given snapshot
	Gt(token SnapToken) bool
	// Lt snapshot is less than given snapshot
	Lt(token SnapToken) bool
}
