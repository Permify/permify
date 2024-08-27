package schema

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Type defines an enumeration for different schema types.
type Type int

const (
	// URL represents a schema type for URLs.
	URL Type = iota

	// File represents a schema type for file paths.
	File

	// Inline represents a schema type for inline data.
	Inline
)

// Loader is a struct that holds a map of loader functions, each corresponding to a Type.
type Loader struct {
	// loaders is a map where each Type is associated with a corresponding function
	// that takes a string as input and returns a string and an error.
	// These functions are responsible for loading data based on the Type.
	loaders map[Type]func(string) (string, error)
}

// NewSchemaLoader initializes and returns a new Loader instance.
// It sets up the map of loader functions for each Type.
func NewSchemaLoader() *Loader {
	return &Loader{
		loaders: map[Type]func(string) (string, error){
			// URL loader function for handling URL type schemas.
			URL: loadFromURL,

			// File loader function for handling file path type schemas.
			File: loadFromFile,

			// Inline loader function for handling inline type schemas.
			Inline: loadInline,
		},
	}
}

// LoadSchema loads a schema based on its type
func (s *Loader) LoadSchema(input string) (string, error) {
	schemaType, err := determineSchemaType(input)
	if err != nil {
		return "", fmt.Errorf("error determining schema type: %w", err)
	}

	loaderFunc, exists := s.loaders[schemaType]
	if !exists {
		return "", fmt.Errorf("loader function not found for schema type: %v", schemaType)
	}

	return loaderFunc(input)
}

// determineSchemaType determines the type of schema based on the input string
func determineSchemaType(input string) (Type, error) {
	if isURL(input) {
		return URL, nil
	}

	valid, err := isFilePath(input)
	if err != nil {
		return Inline, nil
	}
	if valid {
		return File, nil
	}

	return Inline, nil
}

func isURL(input string) bool {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return false
	}

	// Check if the URL has a valid scheme and host
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isFilePath(input string) (bool, error) {
	_, err := os.Stat(input)
	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.New("file does not exist")
		}
		if os.IsPermission(err) {
			return false, errors.New("permission denied")
		}
		return false, err
	}

	return true, nil
}

func loadFromURL(inputURL string) (string, error) {
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
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func loadFromFile(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check if the cleaned path is trying to traverse directories
	if filepath.IsAbs(cleanPath) || filepath.HasPrefix(cleanPath, "..") {
		return "", errors.New("invalid file path")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// loadInline is a function that handles inline schema types.
func loadInline(schema string) (string, error) {
	// Add validation if necessary. For example:
	if schema == "" {
		return "", errors.New("schema is empty")
	}

	return schema, nil
}
