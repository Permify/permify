package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Permify/permify/internal/storage/postgres/utils"
)

func TestContinuousToken(t *testing.T) {
	tokenValue := "test_token"
	token := utils.NewContinuousToken(tokenValue)

	// Test Encode
	encodedToken := token.Encode()
	assert.NotEmpty(t, encodedToken.String())

	// Test Decode
	decodedToken, err := encodedToken.Decode()
	assert.NoError(t, err)
	assert.Equal(t, tokenValue, decodedToken.(utils.ContinuousToken).Value)

	// Test Encode and Decode
	assert.Equal(t, tokenValue, decodedToken.(utils.ContinuousToken).Value)
}
