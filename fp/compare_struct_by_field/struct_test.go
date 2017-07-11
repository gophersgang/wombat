package compare_struct_by_field

import (
	"testing"
	"github.com/stretchr/testify/require"
	"reflect"
	"github.com/v2pro/wombat/gen"
	"fmt"
)

func Test_src_struct(t *testing.T) {
	t.Skip()
	type TestObject struct {
		Field int
	}
	_, src := gen.Gen(F, "S", reflect.TypeOf(TestObject{}), "F", "Field", "T", reflect.TypeOf(int(0)))
	fmt.Println(src)
}

func Test_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		Field int
	}
	should.Equal(-1, Call(
		TestObject{1}, TestObject{2},
		"Field"))
}

func by_reflect(obj1 interface{}, obj2 interface{}, fieldName string) int {
	field1 := reflect.ValueOf(obj1).FieldByName(fieldName).Int()
	field2 := reflect.ValueOf(obj2).FieldByName(fieldName).Int()
	if field1 < field2 {
		return -1
	} else if field1 == field2 {
		return 0
	} else {
		return 1
	}
}

func Benchmark_struct(b *testing.B) {
	type TestObject struct {
		Field int
	}
	Call(
		TestObject{1}, TestObject{2},
		"Field")
	b.Run("plz", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			Call(
				TestObject{1}, TestObject{2},
				"Field")
		}
	})
	b.Run("reflect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.ReportAllocs()
			by_reflect(
				TestObject{1}, TestObject{2},
				"Field")
		}
	})
}
