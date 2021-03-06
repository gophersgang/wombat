package cpJsonToSlice

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
	TemplateName: "cpJsonToSlice",
	Dependencies: []*gen.FuncTemplate{cpAnything.F},
	TemplateParams: map[string]string{
		"DT": "the dst type to copy into",
		"ST": "the src type to copy from",
	},
	FuncName: `cp_into_{{ .DT|symbol }}_from_{{ .ST|symbol }}`,
	Source: `
{{ $cpElem := gen "cpAnything" "DT" (.DT|ptrSliceElem) "ST" .ST }}
func {{ .funcName }}(
	err *error,
	dst {{ .DT|name }},
	src {{ .ST|name }}) {
	// end of signature
	if src.ReadNil() {
		*dst = nil
		return
	}
	index := 0
	originalLen := len(*dst)
	src.ReadArrayCB(func(iter *jsoniter.Iterator) bool {
		if index < originalLen {
			elem := &(*dst)[index]
			{{ $cpElem }}(err, elem, iter)
		} else {
			elem := new({{ .DT|elem|elem|name }})
			{{ $cpElem }}(err, elem, iter)
			*dst = append(*dst, *elem)
		}
		index++
		return true
	})
}
`,
	GenMap: map[string]interface{}{
		"ptrSliceElem": genPtrSliceElem,
	},
}

func genPtrSliceElem(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		panic("unexpected")
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Slice && typ.Kind() != reflect.Array {
		panic("unexpected")
	}
	return reflect.PtrTo(typ.Elem())
}