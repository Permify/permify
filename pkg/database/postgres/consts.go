package postgres

const (
	_defaultMaxConnections              = 0 // 0 = use pgxpool default (unlimited). Set explicitly to override.
	_defaultMaxIdleConnections          = 0 // Deprecated: Use _defaultMinConnections instead. Kept for backward compatibility (maps to MinConnections if MinConnections is not set).
	_defaultMinConnections              = 0 // 0 = use pgxpool default (no minimum). Set explicitly to override.
	_defaultMinIdleConnections          = 1 // Maintain at least one idle connection by default (matches internal/config DefaultConfig). Set explicitly to override if different.
	_defaultMaxDataPerWrite             = 1000
	_defaultMaxRetries                  = 10
	_defaultWatchBufferSize             = 100
	_defaultHealthCheckPeriod           = 0 // 0 = use pgxpool default (1 minute). Set explicitly to override.
	_defaultMaxConnectionLifetimeJitter = 0 // 0 = will default to 20% of MaxConnLifetime if MaxConnLifetime is set. Set explicitly to override.
	_defaultConnectTimeout              = 0 // 0 = use pgx default (no timeout). Set explicitly to override.
)
