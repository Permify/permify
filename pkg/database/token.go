package database

// EncodedContinuousToken -
type EncodedContinuousToken interface {
	// String returns the string representation of the continuous token.
	String() string
	// Decode decodes the continuous token from a string
	Decode() (ContinuousToken, error)
}

// ContinuousToken -
type ContinuousToken interface {
	// Encode encodes the continuous token to a string.
	Encode() EncodedContinuousToken
}

type (
	NoopContinuousToken struct {
		Value string
	}
	NoopEncodedContinuousToken struct {
		Value string
	}
)

// NewNoopContinuousToken - Creates a new continuous token
func NewNoopContinuousToken() ContinuousToken {
	return &NoopContinuousToken{
		Value: "",
	}
}

// Encode - Encodes the token to a string
func (t NoopContinuousToken) Encode() EncodedContinuousToken {
	return NoopEncodedContinuousToken{
		Value: "",
	}
}

// Decode decodes the token from a string
func (t NoopEncodedContinuousToken) Decode() (ContinuousToken, error) {
	return NoopContinuousToken{
		Value: "",
	}, nil
}

// Decode decodes the token from a string
func (t NoopEncodedContinuousToken) String() string {
	return ""
}
