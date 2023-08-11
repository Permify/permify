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

type (
	NoopToken struct {
		Value string
	}
	NoopEncodedToken struct {
		Value string
	}
)

// NewNoopToken - Creates a new noop snapshot token
func NewNoopToken() SnapToken {
	return NoopToken{
		Value: "noop",
	}
}

// Encode - encodes the token to a string
func (t NoopToken) Encode() EncodedSnapToken {
	return NoopEncodedToken{
		Value: "noop",
	}
}

// Eg - Snapshot is equal to given snapshot
func (t NoopToken) Eg(token SnapToken) bool {
	_, ok := token.(NoopToken)
	return ok
}

// Gt - Snapshot is greater than given snapshot
func (t NoopToken) Gt(SnapToken) bool {
	return false
}

// Lt - Snapshot is less than given snapshot
func (t NoopToken) Lt(SnapToken) bool {
	return false
}

// Decode - Decodes the token from a string
func (t NoopEncodedToken) Decode() (SnapToken, error) {
	return NoopToken{
		Value: "noop",
	}, nil
}

// Decode - Decodes the token from a string
func (t NoopEncodedToken) String() string {
	return t.Value
}
