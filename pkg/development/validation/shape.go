package validation

// Shape is a file shape.
type Shape struct {
	// Schema string of authorization model.
	Schema string `yaml:"schema"`

	// Tuples authorization data
	Tuples []string `yaml:"relationships"`

	// Assertions -
	// can user:1 push repository:2 => true
	Assertions []map[string]bool `yaml:"assertions"`
}
