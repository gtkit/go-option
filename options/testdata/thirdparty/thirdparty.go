package thirdparty

import "github.com/gtkit/go-option/options/testdata/thirdparty/pkg"

// Service tests cross-package types (SelectorExpr), matching go-optioner's capability.
type Service struct {
	Name    string          `opt:"-"`
	Config  pkg.ThirdParty
	Timeout int
}
