package postgres

import (
	"time"
)

const (
	_defaultMaxOpenConnections    = 20
	_defaultMinOpenConnections    = 1
	_defaultMaxConnectionLifetime = time.Hour
	_defaultMaxConnectionIdleTime = time.Minute * 30
	_defaultConnectionTimeout     = time.Second
	_defaultHealthCheckPeriod     = time.Minute
)
