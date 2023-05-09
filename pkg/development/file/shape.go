package file

// Shape is a struct that represents an authorization configuration.
type Shape struct {
	// Schema is a string that represents the authorization model schema.
	Schema string `yaml:"schema"`

	// Relationships is a slice of strings that represent the authorization relationships.
	Relationships []string `yaml:"relationships"`

	// Scenarios is a slice of Scenario structs that represent the different authorization scenarios.
	Scenarios []Scenario `yaml:"scenarios"`
}

// Scenario is a struct that represents a specific authorization scenario.
type Scenario struct {
	// Name is a string that represents the name of the scenario.
	Name string `yaml:"name"`

	// Description is a string that provides a brief explanation of the scenario.
	Description string `yaml:"description"`

	// Checks is a slice of Check structs that represent the authorization checks to be performed.
	Checks []Check `yaml:"checks"`

	// EntityFilters is a slice of Filter structs that represent the filters to be applied during the checks.
	EntityFilters []EntityFilter `yaml:"entity_filters"`
}

// Check is a struct that represents an individual authorization check.
type Check struct {
	// Entity is a string that represents the entity type involved in the authorization check.
	Entity string `yaml:"entity"`

	// Subject is a string that represents the subject of the authorization check.
	Subject string `yaml:"subject"`

	// Assertions is a map that contains the authorization assertions to be evaluated.
	Assertions map[string]bool `yaml:"assertions"`
}

// EntityFilter is a struct that represents a filter to be applied during an authorization check.
type EntityFilter struct {
	// EntityType is a string that represents the type of entity the filter applies to.
	EntityType string `yaml:"entity_type"`

	// Subject is a string that represents the subject of the filter.
	Subject string `yaml:"subject"`

	// Assertions is a map that contains the filter assertions to be applied.
	Assertions map[string][]string `yaml:"assertions"`
}
