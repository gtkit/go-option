package templates

import (
	_ "embed"
)

var (
	//go:embed tpl/option.tmpl
	OptionTmpl string

	//go:embed tpl/additional.tmpl
	AdditionalTmpl string

	//go:embed tpl/header.tmpl
	HeaderTmpl string
)
