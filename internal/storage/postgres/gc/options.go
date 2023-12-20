package gc

import (
	"time"
)

// Option represents a function that configures a GC (Garbage Collector) instance.
type Option func(gc *GC)

// Interval is an option that sets the interval duration for the GC.
func Interval(n time.Duration) Option {
	return func(gc *GC) {
		gc.interval = n
	}
}

// Window is an option that sets the window duration for the GC.
func Window(n time.Duration) Option {
	return func(gc *GC) {
		gc.window = n
	}
}

// Timeout is an option that sets the timeout duration for the GC.
func Timeout(n time.Duration) Option {
	return func(gc *GC) {
		gc.timeout = n
	}
}
