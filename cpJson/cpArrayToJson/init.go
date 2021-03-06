package cpArrayToJson

import (
	"github.com/v2pro/wombat/gen"
	"github.com/v2pro/wombat/cp/cpAnything"
	"reflect"
)

func init() {
	cpAnything.F.AddDependency(F)
}

// F the function definition
var F = &gen.FuncTemplate{
	TemplateName: "cpArrayToJson",
	Dependencies: []*gen.FuncTemplate{cpAnything.F},
	TemplateParams: map[string]string{
		"DT": "the dst type to copy into",
		"ST": "the src type to copy from",
	},
	FuncName: `cp_into_{{ .DT|symbol }}_from_{{ .ST|symbol }}`,
	Source: `
{{ $cpElem := gen "cpAnything" "DT" .DT "ST" (.ST|elem) }}
func {{ .funcName }}(
	err *error,
	dst {{ .DT|name }},
	src {{ .ST|name }}) {
	// end of signature
	dst.WriteArrayStart()
	{{ range $index, $_ := .ST|elems }}
		{{ if ne $index 0 }}
		dst.WriteMore()
		{{ end }}
		{{ $cpElem }}(err, dst, src[{{ $index }}])
	{{ end }}
	dst.WriteArrayEnd()
}
`,
	GenMap: map[string]interface{}{
		"elems": genElems,
	},
}

func genElems(typ reflect.Type) []bool {
	return make([]bool, typ.Len())
}
