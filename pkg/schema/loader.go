package schema // Schema loading utilities
import (       // Package imports
	"errors"        // Error handling
	"fmt"           // Formatting
	"io"            // IO operations
	"net/http"      // HTTP client
	"net/url"       // URL parsing
	"os"            // File system
	"path/filepath" // Path utilities
) // End imports
// Type defines an enumeration for different schema types.
type Type int // Schema type enum
const (       // Schema type constants
	URL    Type = iota // URL represents a schema type for URLs
	File               // File represents a schema type for file paths
	Inline             // Inline represents a schema type for inline data
) // End constants
// Loader is a struct that holds a map of loader functions, each corresponding to a Type.
type Loader struct {
	// loaders is a map where each Type is associated with a corresponding function
	// that takes a string as input and returns a string and an error.
	// These functions are responsible for loading data based on the Type.
	loaders map[Type]func(string) (string, error)
} // End Loader struct
// NewSchemaLoader initializes and returns a new Loader instance.
// It sets up the map of loader functions for each Type.
func NewSchemaLoader() *Loader {
	return &Loader{
		loaders: map[Type]func(string) (string, error){ // Initialize loaders
			URL:    loadFromURL,  // URL loader function
			File:   loadFromFile, // File loader function
			Inline: loadInline,   // Inline loader function
		}, // End loaders map
	} // End return
} // End NewSchemaLoader
// LoadSchema loads a schema based on its type
func (s *Loader) LoadSchema(input string) (string, error) {
	schemaType, err := determineSchemaType(input)
	if err != nil {
		return "", fmt.Errorf("error determining schema type: %w", err)
	}
	loaderFunc, exists := s.loaders[schemaType] // Get loader function
	if !exists {
		return "", fmt.Errorf("loader function not found for schema type: %v", schemaType)
	}
	return loaderFunc(input) // Execute loader
} // End LoadSchema
// determineSchemaType determines the type of schema based on the input string
func determineSchemaType(input string) (Type, error) {
	if isURL(input) { // Check URL first
		return URL, nil // Return URL type
	} // Not URL
	valid, err := isFilePath(input) // Check if file path
	if err != nil {
		return Inline, nil
	}
	if valid {
		return File, nil
	}
	return Inline, nil // Default to inline
} // End determineSchemaType
func isURL(input string) bool { // Check if input is URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return false
	}
	// Check if the URL has a valid scheme and host
	return parsedURL.Scheme != "" && parsedURL.Host != ""
} // End isURL
func isFilePath(input string) (bool, error) { // Check if input is file
	_, err := os.Stat(input) // Get file info
	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.New("file does not exist")
		}
		if os.IsPermission(err) {
			return false, errors.New("permission denied")
		}
		return false, err
	}
	return true, nil // File exists
} // End isFilePath
func loadFromURL(inputURL string) (string, error) { // Load schema from URL
	// Parse and validate the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}
	// Add checks here to validate the scheme, host, etc., of parsedURL
	// For example, ensure the scheme is either http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("invalid URL scheme")
	}
	// Perform the HTTP GET request
	resp, err := http.Get(parsedURL.String())
	if err != nil { // HTTP request failed
		return "", err // Return error
	} // Request succeeded
	defer resp.Body.Close() // Ensure cleanup
	// Read the response body
	body, err := io.ReadAll(resp.Body) // Read response
	if err != nil {                    // Read failed
		return "", err // Return error
	} // Read succeeded
	return string(body), nil // Return body
} // End loadFromURL
func loadFromFile(path string) (string, error) { // Load schema from file
	// Clean the path
	cleanPath := filepath.Clean(path)
	// Check if the cleaned path is trying to traverse directories
	if filepath.IsAbs(cleanPath) || filepath.HasPrefix(cleanPath, "..") {
		return "", errors.New("invalid file path")
	}
	content, err := os.ReadFile(path) // Read file
	if err != nil {                   // Read failed
		return "", err // Return error
	} // Read succeeded
	return string(content), nil // Return content
} // End loadFromFile
// loadInline is a function that handles inline schema types.
func loadInline(schema string) (string, error) { // Load inline schema
	// Add validation if necessary. For example:
	if schema == "" {
		return "", errors.New("schema is empty")
	}
	return schema, nil // Return inline schema
} // End loadInline
