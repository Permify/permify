//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/securego/gosec/cmd/gosec"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
