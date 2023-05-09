package file

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Decoder - Decoder interface
type Decoder interface {
	Decode(out interface{}) error
}

// NewDecoderFromURL - Creates new decoder
func NewDecoderFromURL(url *url.URL) (Decoder, error) {
	switch url.Scheme {
	case "file":
		return NewFileDecoder(url.Path), nil
	case "http", "https":
		if url.Hostname() == "gist.github.com" {
			url.Host = "gist.githubusercontent.com"
			url.Path = path.Join(url.Path, "/raw")
		}
		return NewHTTPDecoder(url.String()), nil
	case "":
		return NewFileDecoder(url.Path), nil
	default:
		return nil, errors.New("unknown decoder type")
	}
}

// FILE

type FileDecoder struct {
	path string
}

func NewFileDecoder(path string) *FileDecoder {
	return &FileDecoder{
		path: path,
	}
}

// Decode - Decode a file
func (d FileDecoder) Decode(out interface{}) (err error) {
	file, err := os.Open(d.path)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

// HTTP

type HTTPDecoder struct {
	url string
}

// NewHTTPDecoder - Creates new HTTP decoder
func NewHTTPDecoder(url string) *HTTPDecoder {
	return &HTTPDecoder{
		url: url,
	}
}

// Decode - decode HTTP
func (d HTTPDecoder) Decode(out interface{}) (err error) {
	r, err := http.Get(d.url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}
