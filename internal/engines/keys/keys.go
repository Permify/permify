package keys

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckEngineWithKeys is a struct that holds an instance of a cache.Cache for managing engine keys.
type CheckEngineWithKeys struct {
	checker invoke.Check
	cache   cache.Cache
	l       *logger.Logger
}

// NewCheckEngineWithKeys creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithKeys(checker invoke.Check, cache cache.Cache, l *logger.Logger) invoke.Check {
	return &CheckEngineWithKeys{
		checker: checker,
		cache:   cache,
		l:       l,
	}
}

// Check performs a permission check for a given request, using the cached results if available.
func (c *CheckEngineWithKeys) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	// Try to get the cached result for the given request.
	res, found := c.getCheckKey(request)

	// If a cached result is found, handle exclusion and return the result.
	if found {
		// If the request doesn't have the exclusion flag set, return the cached result.
		return &base.PermissionCheckResponse{
			Can:      res.GetCan(),
			Metadata: &base.PermissionCheckResponseMetadata{},
		}, nil
	}

	// Perform the actual permission check using the provided request.
	res, err = c.checker.Check(ctx, request)

	// Check if there's an error or the response is nil, and return the result.
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	c.setCheckKey(request, &base.PermissionCheckResponse{
		Can:      res.GetCan(),
		Metadata: &base.PermissionCheckResponseMetadata{},
	})

	// Return the result of the permission check.
	return res, err
}

// GetCheckKey retrieves the value for the given key from the EngineKeys cache.
// It returns the PermissionCheckResponse if the key is found, and a boolean value
// indicating whether the key was found or not.
func (c *CheckEngineWithKeys) getCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	if key == nil {
		// If either the key or value is nil, return false
		return nil, false
	}

	// Initialize a new xxhash object
	h := xxhash.New()

	// Write the checkKey string to the hash object
	_, err := h.Write([]byte(GenerateKey(key)))
	if err != nil {
		// If there's an error, return nil and false
		return nil, false
	}

	// Generate the final cache key by encoding the hash object's sum as a hexadecimal string
	k := hex.EncodeToString(h.Sum(nil))

	// Get the value from the cache using the generated cache key
	resp, found := c.cache.Get(k)

	// If the key is found, return the value and true
	if found {
		// If permission is granted, return allowed response
		return &base.PermissionCheckResponse{
			Can: resp.(base.CheckResult),
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, true
	}

	// If the key is not found, return nil and false
	return nil, false
}

// setCheckKey is a function to set a check key in the cache of the CheckEngineWithKeys.
// It takes a permission check request as a key, a permission check response as a value,
// and returns a boolean value indicating if the operation was successful.
func (c *CheckEngineWithKeys) setCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
	// If either the key or the value is nil, return false.
	if key == nil || value == nil {
		return false
	}

	// Create a new xxhash object for hashing.
	h := xxhash.New()

	// Generate a key string from the permission check request and write it to the hash.
	// If there's an error while writing to the hash, return false.
	size, err := h.Write([]byte(GenerateKey(key)))
	if err != nil {
		return false
	}

	// Compute the hash sum and encode it as a hexadecimal string.
	k := hex.EncodeToString(h.Sum(nil))

	// Set the hashed key and the check result in the cache, using the size of the hashed key as an expiry.
	// The Set method should return true if the operation was successful, so return the result.
	return c.cache.Set(k, value.Can, int64(size))
}

// GenerateKey function takes a PermissionCheckRequest and generates a unique key
// Key format: check|{tenant_id}|{schema_version}|{snap_token}|{context}|{entity:id#permission(optional_arguments)@subject:id#optional_relation}
func GenerateKey(key *base.PermissionCheckRequest) string {
	// Create a new EntityAndRelation object with the entity and permission from the key
	entityRelation := &base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}

	// Initialize the parts slice with the string "check"
	parts := []string{"check"}

	// If tenantId is not empty, append it to parts
	if tenantId := key.GetTenantId(); tenantId != "" {
		parts = append(parts, tenantId)
	}

	// If Metadata exists, extract schema version and snap token and append them to parts if they are not empty
	if meta := key.GetMetadata(); meta != nil {
		if version := meta.GetSchemaVersion(); version != "" {
			parts = append(parts, version)
		}
		if token := meta.GetSnapToken(); token != "" {
			parts = append(parts, token)
		}
	}

	// If Context exists, convert it to string and append it to parts
	if ctx := key.GetContext(); ctx != nil {
		parts = append(parts, ContextToString(ctx))
	}

	// Convert entity and relation to string with any optional arguments and append to parts
	entityRelationString := tuple.EntityAndRelationToString(entityRelation, key.GetArguments()...)
	subjectString := tuple.SubjectToString(key.GetSubject())

	if entityRelationString != "" {
		parts = append(parts, fmt.Sprintf("%s@%s", entityRelationString, subjectString))
	}

	// Join all parts with "|" delimiter to generate the final key
	return strings.Join(parts, "|")
}

// ContextToString function takes a Context object and converts it into a string
func ContextToString(context *base.Context) string {
	// Initialize an empty slice to store parts of the context
	var parts []string

	// For each Tuple in the Context, convert it to a string and append to parts
	for _, tup := range context.GetTuples() {
		parts = append(parts, tuple.ToString(tup)) // replace with your function
	}

	// For each Attribute in the Context, convert it to a string and append to parts
	for _, attr := range context.GetAttributes() {
		parts = append(parts, attribute.ToString(attr)) // replace with your function
	}

	// If Data exists in the Context, convert it to JSON string and append to parts
	if data := context.GetData(); data != nil {
		parts = append(parts, mapToString(data.AsMap()))
	}

	// Join all parts with "," delimiter to generate the final context string
	return strings.Join(parts, ",")
}

// mapToString function takes a map[string]interface{} and converts it into a string
func mapToString(m map[string]interface{}) string {
	s := ""
	for key, value := range m {
		// This will work as long as `value` is a simple type;
		// it won't work correctly for nested maps, etc.
		s += fmt.Sprintf("%s:%v,", key, value)
	}
	// Remove the last comma and space
	if len(s) > 0 {
		s = s[:len(s)-2]
	}
	return s
}
