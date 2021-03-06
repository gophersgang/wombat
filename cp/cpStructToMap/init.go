package cpStructToMap

import (
	"github.com/v2pro/wombat/cp/cpAnything"
	"github.com/v2pro/wombat/gen"
	"reflect"
)

func init() {
	cpAnything.F.AddDependency(F)
}

// F the function definition
var F = &gen.FuncTemplate{
	TemplateName: "cpStructToMap",
	Dependencies: []*gen.FuncTemplate{cpAnything.F},
	TemplateParams: map[string]string{
		"DT": "the dst type to copy into",
		"ST": "the src type to copy from",
	},
	FuncName: `cp_into_{{ .DT|symbol }}_from_{{ .ST|symbol }}`,
	Source: `
{{ $bindings := calcBindings .DT .ST }}
{{ range $_, $binding := $bindings}}
	{{ $cp := gen "cpAnything" "DT" $binding.dstFieldType "ST" $binding.srcFieldType }}
	{{ assignCp $binding $cp }}
{{ end }}
func {{ .funcName }}(
	err *error,
	dst {{ .DT|name }},
	src {{ .ST|name }}) {
	// end of signature
	var existingElem {{ .DT|elem|name }}
	var found bool
	{{ range $_, $binding := $bindings }}
		existingElem, found = dst["{{ $binding.dstFieldName }}"]
		if found {
			{{ $binding.cp }}(err, &existingElem, src.{{ $binding.srcFieldName }})
			dst["{{ $binding.dstFieldName }}"] = existingElem
		} else {
			newElem := new({{ $binding.dstFieldType|elem|name }})
			{{ $binding.cp }}(err, newElem, src.{{ $binding.srcFieldName }})
			dst["{{ $binding.dstFieldName }}"] = *newElem
		}
	{{ end }}
}`,
	GenMap: map[string]interface{}{
		"calcBindings": genCalcBindings,
		"assignCp":     genAssignCp,
	},
}

func genCalcBindings(dstType, srcType reflect.Type) interface{} {
	bindings := []interface{}{}
	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		bindings = append(bindings, map[string]interface{}{
			"srcFieldName": srcField.Name,
			"srcFieldType": srcField.Type,
			"dstFieldName": srcField.Name,
			"dstFieldType": reflect.PtrTo(dstType.Elem()),
		})
	}
	return bindings
}

func genAssignCp(binding map[string]interface{}, cpFuncName string) string {
	binding["cp"] = cpFuncName
	return ""
}
