package engines

import (
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

const (
	_defaultConcurrencyLimit = 100
)

// CheckOption - a functional option type for configuring the CheckEngine.
type CheckOption func(engine *CheckEngine)

// CheckConcurrencyLimit - a functional option that sets the concurrency limit for the CheckEngine.
func CheckConcurrencyLimit(limit int) CheckOption {
	return func(c *CheckEngine) {
		c.concurrencyLimit = limit
	}
}

// LookupEntityOption - a functional option type for configuring the LookupEntityEngine.
type LookupEntityOption func(engine *LookupEntityEngine)

// LookupEntityConcurrencyLimit - a functional option that sets the concurrency limit for the LookupEntityEngine.
func LookupEntityConcurrencyLimit(limit int) LookupEntityOption {
	return func(c *LookupEntityEngine) {
		c.concurrencyLimit = limit
	}
}

// LookupSubjectOption - a functional option type for configuring the LookupSubjectEngine.
type LookupSubjectOption func(engine *LookupSubjectEngine)

// LookupSubjectConcurrencyLimit - a functional option that sets the concurrency limit for the LookupSubjectEngine.
func LookupSubjectConcurrencyLimit(limit int) LookupSubjectOption {
	return func(c *LookupSubjectEngine) {
		c.concurrencyLimit = limit
	}
}

// SubjectPermissionOption - a functional option type for configuring the SubjectPermissionEngine.
type SubjectPermissionOption func(engine *SubjectPermissionEngine)

// SubjectPermissionConcurrencyLimit - a functional option that sets the concurrency limit for the SubjectPermissionEngine.
func SubjectPermissionConcurrencyLimit(limit int) SubjectPermissionOption {
	return func(c *SubjectPermissionEngine) {
		c.concurrencyLimit = limit
	}
}

// joinResponseMetas - a helper function that merges multiple PermissionCheckResponseMetadata structs into one.
func joinResponseMetas(meta ...*base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	response := &base.PermissionCheckResponseMetadata{}
	for _, m := range meta {
		response.CheckCount += m.CheckCount
	}
	return response
}

// SubjectPermissionResponse - a struct that holds a SubjectPermissionResponse and an error for a single subject permission check result.
type SubjectPermissionResponse struct {
	permission string
	result     base.CheckResult
	err        error
}

// CheckResponse - a struct that holds a PermissionCheckResponse and an error for a single check function.
type CheckResponse struct {
	resp *base.PermissionCheckResponse
	err  error
}

// ERMap - a thread-safe map of ENR records.
type ERMap struct {
	value sync.Map
}

func (s *ERMap) Add(onr *base.EntityAndRelation) bool {
	key := tuple.EntityAndRelationToString(onr)
	_, existed := s.value.LoadOrStore(key, struct{}{})
	return !existed
}

// LookupSubjectResponse -
type LookupSubjectResponse struct {
	resp *base.PermissionLookupSubjectResponse
	err  error
}

// getDuplicates is a function that accepts a slice of strings and returns a slice of duplicated strings in the input slice
func getDuplicates(s []string) []string {
	// "seen" will keep track of all strings we have encountered in the slice
	seen := make(map[string]bool)

	// "duplicates" will keep track of all strings that are duplicated in the slice
	duplicates := make(map[string]bool)

	// Iterate over every string in the input slice
	for _, str := range s {
		// If we have seen the string before, then it is a duplicate.
		// So, we add it to our duplicates map.
		if _, value := seen[str]; value {
			duplicates[str] = true
		} else {
			// If we haven't seen the string before, add it to the seen map.
			seen[str] = true
		}
	}

	// "duplicatesSlice" will eventually hold our results: all strings that are duplicated in the input slice
	duplicatesSlice := make([]string, 0, len(duplicates))

	// Now, duplicates map contains all the duplicated strings.
	// We iterate over it and add each string to our result slice.
	for str := range duplicates {
		duplicatesSlice = append(duplicatesSlice, str)
	}

	// Return the slice that contains all the duplicated strings
	return duplicatesSlice
}
