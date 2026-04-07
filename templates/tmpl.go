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

	//go:embed tpl/option_closure.tmpl
	OptionClosureTmpl string

	//go:embed tpl/additional_closure.tmpl
	AdditionalClosureTmpl string
)
