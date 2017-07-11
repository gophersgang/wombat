package gen

import (
	"plugin"
	"os"
	"io/ioutil"
	"os/exec"
	"bytes"
	"sync"
	"github.com/v2pro/plz"
	"text/template"
	"reflect"
	"github.com/v2pro/plz/logging"
	"strings"
	"strconv"
)

var logger = plz.LoggerOf("package", "gen")

func init() {
	logging.Providers = append(logging.Providers, func(loggerKv []interface{}) logging.Logger {
		for i := 0; i < len(loggerKv); i += 2 {
			key := loggerKv[i].(string)
			if key == "package" && "gen" == loggerKv[i+1] {
				return logging.NewStderrLogger(loggerKv, logging.LEVEL_DEBUG)
			}
		}
		return nil
	})
}

type FuncTemplate struct {
	Variables    map[string]string
	Source       string
	FuncName     string
	Dependencies map[string]*FuncTemplate
}

type generator struct {
	generatedTypes map[reflect.Type]bool
}

func (g *generator) gen(fTmpl *FuncTemplate, args ...interface{}) (string, string) {
	generatedSource := ""
	data := map[string]interface{}{}
	variables := map[string]string{}
	for k, v := range fTmpl.Variables {
		variables[k] = v
	}
	for i := 0; i < len(args); i += 2 {
		varName := args[i].(string)
		_, isDeclared := variables[varName]
		if !isDeclared {
			logger.Error("variable not declared", "varName", varName, "args", args)
			panic("variable " + varName + " is not declared")
		}
		delete(variables, varName)
		data[varName] = args[i+1]
		typ, _ := args[i+1].(reflect.Type)
		if typ != nil && typ.Kind() == reflect.Struct {
			if !g.generatedTypes[typ] {
				g.generatedTypes[typ] = true
				generatedSource += generateStruct(typ)
			}
		}
	}
	for k, v := range variables {
		logger.Error("missing variable", "varName", k, "varDescription", v, "args", args)
		panic("missing variable " + k + ": " + v)
	}
	funcName := genFuncName(fTmpl.FuncName, data)
	data["funcName"] = funcName
	tmpl, err := template.New(NewID().String()).Funcs(map[string]interface{}{
		"gen": func(depName string, newKv ...interface{}) interface{} {
			funcName, source := g.gen(fTmpl.Dependencies[depName], newKv...)
			return struct {
				FuncName string
				Source   string
			}{FuncName: funcName, Source: source}
		},
		"is_one_ptr_struct_or_array": func_is_one_ptr_struct_or_array,
		"field_of": func_field_of,
		"elem":     func_elem,
		"is_ptr":   func_is_ptr,
		"name":     func_name,
		"symbol":   func_symbol,
		"cast":     func_cast,
	}).Parse(fTmpl.Source)
	panicOnError(err)
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	panicOnError(err)
	return funcName, generatedSource + out.String()
}

func Gen(fTmpl *FuncTemplate, kv ...interface{}) (string, string) {
	return (&generator{
		generatedTypes: map[reflect.Type]bool{},
	}).gen(fTmpl, kv...)
}

func genFuncName(funcNameTmpl string, data interface{}) string {
	tmpl, err := template.New(NewID().String()).Funcs(map[string]interface{}{
		"symbol": func_symbol,
		"name":   func_name,
	}).Parse(funcNameTmpl)
	panicOnError(err)
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	panicOnError(err)
	return out.String()
}

var compilerMutex = &sync.Mutex{}

func Compile(template *FuncTemplate, kv ...interface{}) plugin.Symbol {
	compilerMutex.Lock()
	defer compilerMutex.Unlock()
	funcName, source := Gen(template, kv...)
	//fmt.Println(source)
	source = `
package main
import "unsafe"

type emptyInterface struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}
	` + source
	srcFileName := "/tmp/" + NewID().String() + ".go"
	soFileName := "/tmp/" + NewID().String() + ".so"
	err := ioutil.WriteFile(srcFileName, []byte(source), 0666)
	if err != nil {
		panic("failed to generate source code: " + err.Error())
	}
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soFileName, srcFileName)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	err = cmd.Run()
	if err != nil {
		logger.Error("compile failed", "source", annotateLines(source))
		panic("failed to compile generated plugin: " + err.Error() + ", " + errBuf.String())
	}
	generatedPlugin, err := plugin.Open(soFileName)
	if err != nil {
		panic("failed to load generated plugin: " + err.Error())
	}
	compareObj, err := generatedPlugin.Lookup(funcName)
	if err != nil {
		panic("failed to lookup symbol from generated plugin: " + err.Error())
	}
	err = os.Remove(srcFileName)
	if err != nil {
		logger.Error("failed to remove generated source", "srcFileName", srcFileName)
	}
	err = os.Remove(soFileName)
	if err != nil {
		logger.Error("failed to remove generated plugin", "soFileName", soFileName)
	}
	return compareObj
}

func annotateLines(source string) string {
	var buf bytes.Buffer
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lineNo := strconv.FormatInt(int64(i+1), 10)
		buf.WriteString(lineNo)
		buf.WriteString(": ")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return buf.String()
}
