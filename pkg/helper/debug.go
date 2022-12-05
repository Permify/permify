package helper

import (
	"os"

	"github.com/davecgh/go-spew/spew"
)

// Pre - Exit running project.
func Pre(x interface{}, y ...interface{}) {
	spew.Dump(x)
	os.Exit(1)
}
