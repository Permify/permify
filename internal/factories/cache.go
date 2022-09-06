package factories

import (
	"errors"
	"fmt"

	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
)

// CacheFactory -
func CacheFactory(engine cache.Engine) (c cache.Cache, err error) {
	switch engine {
	case cache.RISTRETTO:
		c, err = ristretto.New()
		if err != nil {
			return nil, err
		}
		return
	default:
		return nil, errors.New(fmt.Sprintf("%s connection is unsupported", engine.String()))
	}
}
