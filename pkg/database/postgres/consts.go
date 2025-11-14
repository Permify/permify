package postgres

const (
	_defaultMaxConns              = 0 // 0 = use pgxpool default (unlimited). Set explicitly to override.
	_defaultMaxIdleConnections    = 0 // Deprecated: Use _defaultMinConns or _defaultMinIdleConns instead. Kept for backward compatibility.
	_defaultMinConns              = 0 // 0 = use pgxpool default (no minimum). Set explicitly to override.
	_defaultMinIdleConns          = 0 // 0 = use pgxpool default (no minimum idle). Set explicitly to override.
	_defaultMaxDataPerWrite       = 1000
	_defaultMaxRetries            = 10
	_defaultWatchBufferSize       = 100
	_defaultHealthCheckPeriod     = 0 // 0 = use pgxpool default (1 minute). Set explicitly to override.
	_defaultMaxConnLifetimeJitter = 0 // 0 = will default to 20% of MaxConnLifetime if MaxConnLifetime is set. Set explicitly to override.
	_defaultConnectTimeout        = 0 // 0 = use pgx default (no timeout). Set explicitly to override.
)
