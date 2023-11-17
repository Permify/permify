package schema

import (
	"io"
	"net/http"
	"os"
	"strings"
)

type SchemaType int

const (
	URL SchemaType = iota
	File
	Inline
)

type SchemaLoader struct {
	loaders map[SchemaType]func(string) (string, error)
}

func NewSchemaLoader() *SchemaLoader {
	return &SchemaLoader{
		loaders: map[SchemaType]func(string) (string, error){
			URL:    loadFromURL,
			File:   loadFromFile,
			Inline: loadInline,
		},
	}
}

func (s *SchemaLoader) LoadSchema(input string) (string, error) {
	return s.loaders[determineSchemaType(input)](input)
}

func determineSchemaType(input string) SchemaType {
	if isURL(input) {
		return URL
	} else if isFilePath(input) {
		return File
	} else {
		return Inline
	}
}

func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

func isFilePath(input string) bool {
	_, err := os.Stat(input)
	return err == nil
}

func loadFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func loadFromFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func loadInline(schema string) (string, error) {
	return schema, nil
}
