package cpIntoInterface

import (
	"github.com/v2pro/wombat/cp/cpAnything"
	"github.com/v2pro/wombat/gen"
)

func init() {
	cpAnything.F.AddDependency(F)
}

// F the function definition
var F = &gen.FuncTemplate{
	TemplateName: "cpIntoInterface",
	TemplateParams: map[string]string{
		"DT": "the dst type to copy into",
		"ST": "the src type to copy from",
	},
	FuncName: `cp_into_{{ .DT|symbol }}_from_{{ .ST|symbol }}`,
	Source: `
func {{ .funcName }}(
	err *error,
	dst {{ .DT|name }},
	src {{ .ST|name }}) {
	// end of signature
	if dst == nil {
		return
	}
	if *dst == nil {
		newDst := new({{ .ST|name }})
		newErr := cpDynamically(newDst, src)
		if newErr != nil && *err == nil {
			*err = newErr
		}
		*dst = *newDst
	} else {
		newErr := cpDynamically(*dst, src)
		if newErr != nil && *err == nil {
			*err = newErr
		}
	}
}
`,
}
