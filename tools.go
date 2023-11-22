//go:build tools
// +build tools

package stuff

import (
	_ "github.com/bokwoon95/wgo"
	_ "github.com/deepmap/oapi-codegen/pkg/codegen"
	_ "github.com/git-chglog/git-chglog/cmd/git-chglog"
	_ "github.com/golangci/golangci-lint/pkg/commands"
	_ "github.com/pressly/goose/v3/cmd/goose"
	_ "github.com/stephenafamo/bob/gen/bobgen-sqlite"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
