package file

// Shape is a struct that represents an authorization configuration.
type Shape struct {
	TenantID string `yaml:"tenant_id"`

	// Schema is a string that represents the authorization model schema.
	Schema string `yaml:"schema"`

	// Relationships is a slice of strings that represent the authorization relationships.
	Relationships []string `yaml:"relationships"`

	// Attributes is a slice of strings that represent the authorization attributes.
	Attributes []string `yaml:"attributes"`

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

	// SubjectFilters is a slice of Filter structs that represent the filters to be applied during the checks.
	SubjectFilters []SubjectFilter `yaml:"subject_filters"`
}

// Context represents a structure with context data.
type Context struct {
	// Tuples is a slice of strings, each representing a tuple in the context.
	Tuples []string `yaml:"tuples"`

	// Attributes is a slice of strings, each representing an attribute in the context.
	Attributes []string `yaml:"attributes"`

	// Data is a map where each key-value pair represents additional context data.
	Data map[string]interface{} `yaml:"data"`
}

// Check is a struct that represents an individual authorization check.
type Check struct {
	// Context is a struct that represents the context of the authorization check.
	Context Context `yaml:"context"`

	// Entity is a string that represents the entity type involved in the authorization check.
	Entity string `yaml:"entity"`

	// Subject is a string that represents the subject of the authorization check.
	Subject string `yaml:"subject"`

	// Assertions is a map that contains the authorization assertions to be evaluated.
	Assertions map[string]bool `yaml:"assertions"`
}

// EntityFilter is a struct that represents a filter to be applied during an authorization check.
type EntityFilter struct {
	// Context is a struct that represents the context of the authorization entity filter.
	Context Context `yaml:"context"`

	// EntityType is a string that represents the type of entity the filter applies to.
	EntityType string `yaml:"entity_type"`

	// Subject is a string that represents the subject of the filter.
	Subject string `yaml:"subject"`

	// Assertions is a map that contains the filter assertions to be applied.
	Assertions map[string][]string `yaml:"assertions"`
}

// SubjectFilter is a struct that represents a filter to be applied during an authorization check.
type SubjectFilter struct {
	// Context is a struct that represents the context of the authorization subject filter.
	Context Context `yaml:"context"`

	// EntityType is a string that represents the type of entity the filter applies to.
	SubjectReference string `yaml:"subject_reference"`

	// Entity is a string that represents the entity type involved in the authorization check.
	Entity string `yaml:"entity"`

	// Assertions is a map that contains the filter assertions to be applied.
	Assertions map[string][]string `yaml:"assertions"`
}
