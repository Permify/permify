package validation

// Shape - Is a file shape.
type Shape struct {
	// Schema - String of authorization model.
	Schema string `yaml:"schema"`

	// Relationships - Authorization data
	Relationships []string `yaml:"relationships"`

	// Assertions -
	// can user:1 push repository:2 => true
	Assertions []map[string]bool `yaml:"assertions"`
}
