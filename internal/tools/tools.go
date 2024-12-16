// Package tools is used to track the versions of the tools in the go.mod file.
package tools

import (
	_ "github.com/boumenot/gocover-cobertura"  // Code coverage
	_ "github.com/mgechev/revive"              // Linter
	_ "github.com/securego/gosec/v2/cmd/gosec" // Security checker
	_ "github.com/segmentio/golines"           // Line formatter
	_ "mvdan.cc/gofumpt"                       // Code formatter
)
