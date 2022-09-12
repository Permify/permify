package authn

import (
	"github.com/labstack/echo/v4"
)

// KeyAuthenticator -
type KeyAuthenticator interface {
	SetKeys(keys []string)
	Validator() func(key string, c echo.Context) (bool, error)
}

// KeyAuthn -
type KeyAuthn struct {
	Keys []string
}

func NewKeyAuthn(keys ...string) *KeyAuthn {
	return &KeyAuthn{
		Keys: keys,
	}
}

// SetKeys -
func (a *KeyAuthn) SetKeys(keys []string) {
	a.Keys = keys
}

// Validator -
func (a *KeyAuthn) Validator() func(key string, c echo.Context) (bool, error) {
	return func(key string, c echo.Context) (bool, error) {
		for _, k := range a.Keys {
			if key == k {
				return true, nil
			}
		}
		return false, nil
	}
}
