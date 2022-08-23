package postgres

import (
	"time"
)

const (
	_defaultMaxPoolSize  = 20
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)
