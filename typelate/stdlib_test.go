package typelate_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	typelate2 "github.com/crhntr/muxt/typelate"
)

// bigInt and bigUint are hex string representing numbers either side
// of the max int boundary.
// We do it this way so the test doesn't depend on ints being 32 bits.
var (
	bigInt  = fmt.Sprintf("0x%x", int(1<<uint(reflect.TypeFor[int]().Bits()-1)-1))
	bigUint = fmt.Sprintf("0x%x", uint(1<<uint(reflect.TypeFor[int]().Bits()-1)))
)

// TestExec is a modified version of the same test from text/template.
// These are the commented out tests:
// - "with empty interface, struct field"
//   Accessing fields on any is not supported because it is not type-safe.
// -
// The following tests from stdlib are over-written by tests tree_test.go
// - "field on interface"
// - "field on parenthesized interface"

func TestExec(t *testing.T) {
	const testFuncName = "TestExec"
	type execTest struct {
		name   string
		input  string
		output string
		data   any
		ok     bool
	}

	testPkg := find(t, loadPkg(), func(p *packages.Package) bool {
		return p.Name == "templatetype_test"
	})

	funcExec := template.FuncMap{
		"add":         add,
		"count":       count,
		"dddArg":      dddArg,
		"die":         die,
		"echo":        echo,
		"makemap":     makemap,
		"mapOfThree":  mapOfThree,
		"oneArg":      oneArg,
		"returnInt":   returnInt,
		"stringer":    stringer,
		"twoArgs":     twoArgs,
		"typeOf":      typeOf,
		"valueString": valueString,
		"vfunc":       vfunc,
		"zeroArgs":    zeroArgs,
	}
	funcSource := typelate2.DefaultFunctions(testPkg.Types)
	for k := range funcExec {
		v, ok := testPkg.Types.Scope().Lookup(k).Type().(*types.Signature)
		require.True(t, ok)
		funcSource[k] = v

	}
	fileIndex := slices.IndexFunc(testPkg.Syntax, func(file *ast.File) bool {
		pos := testPkg.Fset.Position(file.Pos())
		return file.Name.Name == "templatetype_test" && filepath.Base(pos.Filename) == "stdlib_test.go"
	})
	if fileIndex < 0 {
		t.Fatal("no stdlib_test.go found")
	}
	file := testPkg.Syntax[fileIndex]
	var ttRows *ast.CompositeLit
	for _, decl := range file.Decls {
		testFunc, ok := decl.(*ast.FuncDecl)
		if !ok || testFunc.Name.Name != testFuncName {
			continue
		}
		for _, stmt := range testFunc.Body.List {
			rangeStatement, ok := stmt.(*ast.RangeStmt)
			if !ok {
				continue
			}
			tests, ok := rangeStatement.X.(*ast.CompositeLit)
			if !ok {
				continue
			}
			arr, ok := tests.Type.(*ast.ArrayType)
			if !ok {
				continue
			}

			if testType, ok := arr.Elt.(*ast.Ident); !ok || testType.Name != "execTest" {
				continue
			}
			ttRows = tests
		}
	}

	for _, tt := range []execTest{
		// Trivial cases.
		{"empty", "", "", nil, true},
		{"text", "some text", "some text", nil, true},
		{"nil action", "{{nil}}", "", nil, false},

		// Ideal constants.
		{"ideal int", "{{typeOf 3}}", "int", 0, true},
		{"ideal float", "{{typeOf 1.0}}", "float64", 0, true},
		{"ideal exp float", "{{typeOf 1e1}}", "float64", 0, true},
		{"ideal complex", "{{typeOf 1i}}", "complex128", 0, true},
		{"ideal int", "{{typeOf " + bigInt + "}}", "int", 0, true},
		{"ideal too big", "{{typeOf " + bigUint + "}}", "", 0, false},
		{"ideal nil without type", "{{nil}}", "", 0, false},

		// Fields of structs.
		{".X", "-{{.X}}-", "-x-", tVal, true},
		{".U.V", "-{{.U.V}}-", "-v-", tVal, true},
		{".unexported", "{{.unexported}}", "", tVal, false},

		// Fields on maps.
		{"map .one", "{{.MSI.one}}", "1", tVal, true},
		{"map .two", "{{.MSI.two}}", "2", tVal, true},
		{"map .NO", "{{.MSI.NO}}", "<no value>", tVal, true},
		{"map .one interface", "{{.MXI.one}}", "1", tVal, true},
		{"map .WRONG args", "{{.MSI.one 1}}", "", tVal, false},
		{"map .WRONG type", "{{.MII.one}}", "", tVal, false},

		// Dots of all kinds to test basic evaluation.
		{"dot int", "<{{.}}>", "<13>", 13, true},
		{"dot uint", "<{{.}}>", "<14>", uint(14), true},
		{"dot float", "<{{.}}>", "<15.1>", 15.1, true},
		{"dot bool", "<{{.}}>", "<true>", true, true},
		{"dot complex", "<{{.}}>", "<(16.2-17i)>", 16.2 - 17i, true},
		{"dot string", "<{{.}}>", "<hello>", "hello", true},
		{"dot slice", "<{{.}}>", "<[-1 -2 -3]>", []int{-1, -2, -3}, true},
		{"dot map", "<{{.}}>", "<map[two:22]>", map[string]int{"two": 22}, true},
		{"dot struct", "<{{.}}>", "<{7 seven}>", struct {
			a int
			b string
		}{7, "seven"}, true},

		// Variables.
		{"$ int", "{{$}}", "123", 123, true},
		{"$.I", "{{$.I}}", "17", tVal, true},
		{"$.U.V", "{{$.U.V}}", "v", tVal, true},
		{"declare in action", "{{$x := $.U.V}}{{$x}}", "v", tVal, true},
		{"simple assignment", "{{$x := 2}}{{$x = 3}}{{$x}}", "3", tVal, true},
		{
			"nested assignment",
			"{{$x := 2}}{{if true}}{{$x = 3}}{{end}}{{$x}}",
			"3", tVal, true,
		},
		{
			"nested assignment changes the last declaration",
			"{{$x := 1}}{{if true}}{{$x := 2}}{{if true}}{{$x = 3}}{{end}}{{end}}{{$x}}",
			"1", tVal, true,
		},

		// Type with String method.
		{"V{6666}.String()", "-{{.V0}}-", "-<6666>-", tVal, true},
		{"&V{7777}.String()", "-{{.V1}}-", "-<7777>-", tVal, true},
		{"(*V)(nil).String()", "-{{.V2}}-", "-nilV-", tVal, true},

		// Type with Error method.
		{"W{888}.Error()", "-{{.W0}}-", "-[888]-", tVal, true},
		{"&W{999}.Error()", "-{{.W1}}-", "-[999]-", tVal, true},
		{"(*W)(nil).Error()", "-{{.W2}}-", "-nilW-", tVal, true},

		// Pointers.
		{"*int", "{{.PI}}", "23", tVal, true},
		{"*string", "{{.PS}}", "a string", tVal, true},
		{"*[]int", "{{.PSI}}", "[21 22 23]", tVal, true},
		{"*[]int[1]", "{{index .PSI 1}}", "22", tVal, true},
		{"NIL", "{{.NIL}}", "<nil>", tVal, true},

		// Empty interfaces holding values.
		{"empty nil", "{{.Empty0}}", "<no value>", tVal, true},
		{"empty with int", "{{.Empty1}}", "3", tVal, true},
		{"empty with string", "{{.Empty2}}", "empty2", tVal, true},
		{"empty with slice", "{{.Empty3}}", "[7 8]", tVal, true},
		{"empty with struct", "{{.Empty4}}", "{UinEmpty}", tVal, true},
		// empty interface not supported - {"empty with struct, field", "{{.Empty4.V}}", "UinEmpty", tVal, true},

		// Edge cases with <no value> with an interface value
		// See tree_test.go - {"field on interface", "{{.foo}}", "<no value>", nil, true},
		// See tree_test.go - {"field on parenthesized interface", "{{(.).foo}}", "<no value>", nil, true},

		// Issue 31810: Parenthesized first element of pipeline with arguments.
		// See also TestIssue31810.
		{"unparenthesized non-function", "{{1 2}}", "", nil, false},
		{"parenthesized non-function", "{{(1) 2}}", "", nil, false},
		{"parenthesized non-function with no args", "{{(1)}}", "1", nil, true}, // This is fine.

		// Method calls.
		{".Method0", "-{{.Method0}}-", "-M0-", tVal, true},
		{".Method1(1234)", "-{{.Method1 1234}}-", "-1234-", tVal, true},
		{".Method1(.I)", "-{{.Method1 .I}}-", "-17-", tVal, true},
		{".Method2(3, .X)", "-{{.Method2 3 .X}}-", "-Method2: 3 x-", tVal, true},
		{".Method2(.U16, `str`)", "-{{.Method2 .U16 `str`}}-", "-Method2: 16 str-", tVal, true},
		{".Method2(.U16, $x)", "{{if $x := .X}}-{{.Method2 .U16 $x}}{{end}}-", "-Method2: 16 x-", tVal, true},
		{".Method3(nil constant)", "-{{.Method3 nil}}-", "-Method3: <nil>-", tVal, true},
		{".Method3(nil value)", "-{{.Method3 .MXI.unset}}-", "-Method3: <nil>-", tVal, true},
		{"method on var", "{{if $x := .}}-{{$x.Method2 .U16 $x.X}}{{end}}-", "-Method2: 16 x-", tVal, true},
		{
			"method on chained var",
			"{{range .MSIone}}{{if $.U.TrueFalse $.True}}{{$.U.TrueFalse $.True}}{{else}}WRONG{{end}}{{end}}",
			"true", tVal, true,
		},
		{
			"chained method",
			"{{range .MSIone}}{{if $.GetU.TrueFalse $.True}}{{$.U.TrueFalse $.True}}{{else}}WRONG{{end}}{{end}}",
			"true", tVal, true,
		},
		{
			"chained method on variable",
			"{{with $x := .}}{{with .SI}}{{$.GetU.TrueFalse $.True}}{{end}}{{end}}",
			"true", tVal, true,
		},
		{".NilOKFunc not nil", "{{call .NilOKFunc .PI}}", "false", tVal, true},
		{".NilOKFunc nil", "{{call .NilOKFunc nil}}", "true", tVal, true},
		{"method on nil value from slice", "-{{range .}}{{.Method1 1234}}{{end}}-", "-1234-", tSliceOfNil, true},
		{"method on typed nil interface value", "{{.NonEmptyInterfaceTypedNil.Method0}}", "M0", tVal, true},

		// Function call builtin.
		{".BinaryFunc", "{{call .BinaryFunc `1` `2`}}", "[1=2]", tVal, true},
		{".VariadicFunc0", "{{call .VariadicFunc}}", "<>", tVal, true},
		{".VariadicFunc2", "{{call .VariadicFunc `he` `llo`}}", "<he+llo>", tVal, true},
		{".VariadicFuncInt", "{{call .VariadicFuncInt 33 `he` `llo`}}", "33=<he+llo>", tVal, true},
		{"if .BinaryFunc call", "{{ if .BinaryFunc}}{{call .BinaryFunc `1` `2`}}{{end}}", "[1=2]", tVal, true},
		{"if not .BinaryFunc call", "{{ if not .BinaryFunc}}{{call .BinaryFunc `1` `2`}}{{else}}No{{end}}", "No", tVal, true},
		// any not permitted - {"Interface Call", `{{stringer .S}}`, "foozle", map[string]any{"S": bytes.NewBufferString("foozle")}, true},
		{".ErrFunc", "{{call .ErrFunc}}", "bla", tVal, true},
		{"call nil", "{{call nil}}", "", tVal, false},

		// Erroneous function calls (check args).
		{".BinaryFuncTooFew", "{{call .BinaryFunc `1`}}", "", tVal, false},
		{".BinaryFuncTooMany", "{{call .BinaryFunc `1` `2` `3`}}", "", tVal, false},
		{".BinaryFuncBad0", "{{call .BinaryFunc 1 3}}", "", tVal, false},
		{".BinaryFuncBad1", "{{call .BinaryFunc `1` 3}}", "", tVal, false},
		{".VariadicFuncBad0", "{{call .VariadicFunc 3}}", "", tVal, false},
		{".VariadicFuncIntBad0", "{{call .VariadicFuncInt}}", "", tVal, false},
		{".VariadicFuncIntBad`", "{{call .VariadicFuncInt `x`}}", "", tVal, false},
		{".VariadicFuncNilBad", "{{call .VariadicFunc nil}}", "", tVal, false},

		// Pipelines.
		{"pipeline", "-{{.Method0 | .Method2 .U16}}-", "-Method2: 16 M0-", tVal, true},
		{"pipeline func", "-{{call .VariadicFunc `llo` | call .VariadicFunc `he` }}-", "-<he+<llo>>-", tVal, true},

		// Nil values aren't missing arguments.
		// should fail type check - {"nil pipeline", "{{ .Empty0 | call .NilOKFunc }}", "true", tVal, true},
		// Empty0 is any, this should fail type check {"nil call arg", "{{ call .NilOKFunc .Empty0 }}", "true", tVal, true},
		{"bad nil pipeline", "{{ .Empty0 | .VariadicFunc }}", "", tVal, false},

		// Parenthesized expressions
		{"parens in pipeline", "{{printf `%d %d %d` (1) (2 | add 3) (add 4 (add 5 6))}}", "1 5 15", tVal, true},

		// Parenthesized expressions with field accesses
		{"parens: $ in paren", "{{($).X}}", "x", tVal, true},
		{"parens: $.GetU in paren", "{{($.GetU).V}}", "v", tVal, true},
		// echo changes type $ to any this should fail the type checker - {"parens: $ in paren in pipe", "{{($ | echo).X}}", "x", tVal, true},
		{"parens: spaces and args", `{{(makemap "up" "down" "left" "right").left}}`, "right", tVal, true},

		// If.
		{"if true", "{{if true}}TRUE{{end}}", "TRUE", tVal, true},
		{"if false", "{{if false}}TRUE{{else}}FALSE{{end}}", "FALSE", tVal, true},
		{"if nil", "{{if nil}}TRUE{{end}}", "", tVal, false},
		{"if on typed nil interface value", "{{if .NonEmptyInterfaceTypedNil}}TRUE{{ end }}", "", tVal, true},
		{"if 1", "{{if 1}}NON-ZERO{{else}}ZERO{{end}}", "NON-ZERO", tVal, true},
		{"if 0", "{{if 0}}NON-ZERO{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"if 1.5", "{{if 1.5}}NON-ZERO{{else}}ZERO{{end}}", "NON-ZERO", tVal, true},
		{"if 0.0", "{{if .FloatZero}}NON-ZERO{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"if 1.5i", "{{if 1.5i}}NON-ZERO{{else}}ZERO{{end}}", "NON-ZERO", tVal, true},
		{"if 0.0i", "{{if .ComplexZero}}NON-ZERO{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"if emptystring", "{{if ``}}NON-EMPTY{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"if string", "{{if `notempty`}}NON-EMPTY{{else}}EMPTY{{end}}", "NON-EMPTY", tVal, true},
		{"if emptyslice", "{{if .SIEmpty}}NON-EMPTY{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"if slice", "{{if .SI}}NON-EMPTY{{else}}EMPTY{{end}}", "NON-EMPTY", tVal, true},
		{"if emptymap", "{{if .MSIEmpty}}NON-EMPTY{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"if map", "{{if .MSI}}NON-EMPTY{{else}}EMPTY{{end}}", "NON-EMPTY", tVal, true},
		{"if map unset", "{{if .MXI.none}}NON-ZERO{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"if map not unset", "{{if not .MXI.none}}ZERO{{else}}NON-ZERO{{end}}", "ZERO", tVal, true},
		{"if $x with $y int", "{{if $x := true}}{{with $y := .I}}{{$x}},{{$y}}{{end}}{{end}}", "true,17", tVal, true},
		{"if $x with $x int", "{{if $x := true}}{{with $x := .I}}{{$x}},{{end}}{{$x}}{{end}}", "17,true", tVal, true},
		{"if else if", "{{if false}}FALSE{{else if true}}TRUE{{end}}", "TRUE", tVal, true},
		{"if else chain", "{{if eq 1 3}}1{{else if eq 2 3}}2{{else if eq 3 3}}3{{end}}", "3", tVal, true},

		// Print etc.
		{"print", `{{print "hello, print"}}`, "hello, print", tVal, true},
		{"print 123", `{{print 1 2 3}}`, "1 2 3", tVal, true},
		{"print nil", `{{print nil}}`, "<nil>", tVal, true},
		{"println", `{{println 1 2 3}}`, "1 2 3\n", tVal, true},
		{"printf int", `{{printf "%04x" 127}}`, "007f", tVal, true},
		{"printf float", `{{printf "%g" 3.5}}`, "3.5", tVal, true},
		{"printf complex", `{{printf "%g" 1+7i}}`, "(1+7i)", tVal, true},
		{"printf string", `{{printf "%s" "hello"}}`, "hello", tVal, true},
		{"printf function", `{{printf "%#q" zeroArgs}}`, "`zeroArgs`", tVal, true},
		{"printf field", `{{printf "%s" .U.V}}`, "v", tVal, true},
		{"printf method", `{{printf "%s" .Method0}}`, "M0", tVal, true},
		{"printf dot", `{{with .I}}{{printf "%d" .}}{{end}}`, "17", tVal, true},
		{"printf var", `{{with $x := .I}}{{printf "%d" $x}}{{end}}`, "17", tVal, true},
		{"printf lots", `{{printf "%d %s %g %s" 127 "hello" 7-3i .Method0}}`, "127 hello (7-3i) M0", tVal, true},

		// HTML.
		{
			"html", `{{html "<script>alert(\"XSS\");</script>"}}`,
			"&lt;script&gt;alert(&#34;XSS&#34;);&lt;/script&gt;", nil, true,
		},
		{
			"html pipeline", `{{printf "<script>alert(\"XSS\");</script>" | html}}`,
			"&lt;script&gt;alert(&#34;XSS&#34;);&lt;/script&gt;", nil, true,
		},
		{"html PS", `{{html .PS}}`, "a string", tVal, true}, // test renamed, added " PS" suffix
		{"html typed nil", `{{html .NIL}}`, "&lt;nil&gt;", tVal, true},
		{"html untyped nil", `{{html .Empty0}}`, "&lt;no value&gt;", tVal, true},

		// JavaScript.
		{"js", `{{js .}}`, `It\'d be nice.`, `It'd be nice.`, true},

		// URL query.
		{"urlquery", `{{"http://www.example.org/"|urlquery}}`, "http%3A%2F%2Fwww.example.org%2F", nil, true},

		// Booleans
		{"not", "{{not true}} {{not false}}", "false true", nil, true},
		{"and", "{{and false 0}} {{and 1 0}} {{and 0 true}} {{and 1 1}}", "false 0 0 1", nil, true},
		{"or", "{{or 0 0}} {{or 1 0}} {{or 0 true}} {{or 1 1}}", "0 1 true 1", nil, true},
		// type check should not get short-circuit - {"or short-circuit", "{{or 0 1 (die)}}", "1", nil, true},
		// type check should not get short-circuit - {"and short-circuit", "{{and 1 0 (die)}}", "0", nil, true},
		{"or short-circuit2", "{{or 0 0 (die)}}", "", nil, false},
		{"and short-circuit2", "{{and 1 1 (die)}}", "", nil, false},
		{"and pipe-true", "{{1 | and 1}}", "1", nil, true},
		{"and pipe-false", "{{0 | and 1}}", "0", nil, true},
		{"or pipe-true", "{{1 | or 0}}", "1", nil, true},
		{"or pipe-false", "{{0 | or 0}}", "0", nil, true},
		// Should fail type check - {"and undef", "{{and 1 .Unknown}}", "<no value>", nil, true},
		// Should fail type check - {"or undef", "{{or 0 .Unknown}}", "<no value>", nil, true},
		{"boolean if", "{{if and true 1 `hi`}}TRUE{{else}}FALSE{{end}}", "TRUE", tVal, true},
		{"boolean if not", "{{if and true 1 `hi` | not}}TRUE{{else}}FALSE{{end}}", "FALSE", nil, true},
		{"boolean if pipe", "{{if true | not | and 1}}TRUE{{else}}FALSE{{end}}", "FALSE", nil, true},

		// Indexing.
		{"slice[0]", "{{index .SI 0}}", "3", tVal, true},
		{"slice[1]", "{{index .SI 1}}", "4", tVal, true},
		// this is a runtime error {"slice[HUGE]", "{{index .SI 10}}", "", tVal, false},
		{"slice[WRONG]", "{{index .SI `hello`}}", "", tVal, false},
		{"slice[nil]", "{{index .SI nil}}", "", tVal, false},
		{"map[one]", "{{index .MSI `one`}}", "1", tVal, true},
		{"map[two]", "{{index .MSI `two`}}", "2", tVal, true},
		{"map[NO]", "{{index .MSI `XXX`}}", "0", tVal, true},
		{"map[nil]", "{{index .MSI nil}}", "", tVal, false},
		{"map[``]", "{{index .MSI ``}}", "0", tVal, true},
		{"map[WRONG]", "{{index .MSI 10}}", "", tVal, false},
		{"double index", "{{index .SMSI 1 `eleven`}}", "11", tVal, true},
		{"nil[1]", "{{index nil 1}}", "", tVal, false},
		{"map MI64S", "{{index .MI64S 2}}", "i642", tVal, true},
		{"map MI32S", "{{index .MI32S 2}}", "two", tVal, true},
		{"map MUI64S", "{{index .MUI64S 3}}", "ui643", tVal, true},
		{"map MI8S", "{{index .MI8S 3}}", "i83", tVal, true},
		{"map MUI8S", "{{index .MUI8S 2}}", "u82", tVal, true},
		// This should fail the type checker - {"index of an interface field", "{{index .Empty3 0}}", "7", tVal, true},

		// Slicing.
		{"slice[:]", "{{slice .SI}}", "[3 4 5]", tVal, true},
		{"slice[1:]", "{{slice .SI 1}}", "[4 5]", tVal, true},
		{"slice[1:2]", "{{slice .SI 1 2}}", "[4]", tVal, true},
		{"slice[-1:]", "{{slice .SI -1}}", "", tVal, false},
		{"slice[1:-2]", "{{slice .SI 1 -2}}", "", tVal, false},
		{"slice[1:2:-1]", "{{slice .SI 1 2 -1}}", "", tVal, false},
		// need to figure out const value passing - {"slice[2:1]", "{{slice .SI 2 1}}", "", tVal, false},
		// need to figure out const value passing - {"slice[2:2:1]", "{{slice .SI 2 2 1}}", "", tVal, false},
		// need to figure out const value passing - {"out of range", "{{slice .SI 4 5}}", "", tVal, false},
		// need to figure out const value passing - {"out of range", "{{slice .SI 2 2 5}}", "", tVal, false},
		{"len(s) < indexes < cap(s)", "{{slice .SICap 6 10}}", "[0 0 0 0]", tVal, true},
		{"len(s) < indexes < cap(s)", "{{slice .SICap 6 10 10}}", "[0 0 0 0]", tVal, true},
		// need to figure out const value passing - {"indexes > cap(s)", "{{slice .SICap 10 11}}", "", tVal, false},
		// need to figure out const value passing - {"indexes > cap(s)", "{{slice .SICap 6 10 11}}", "", tVal, false},
		{"array[:]", "{{slice .AI}}", "[3 4 5]", tVal, true},
		{"array[1:]", "{{slice .AI 1}}", "[4 5]", tVal, true},
		{"array[1:2]", "{{slice .AI 1 2}}", "[4]", tVal, true},
		{"string[:]", "{{slice .S}}", "xyz", tVal, true},
		{"string[0:1]", "{{slice .S 0 1}}", "x", tVal, true},
		{"string[1:]", "{{slice .S 1}}", "yz", tVal, true},
		{"string[1:2]", "{{slice .S 1 2}}", "y", tVal, true},
		// need to figure out const value passing - {"out of range", "{{slice .S 1 5}}", "", tVal, false},
		{"3-index slice of string", "{{slice .S 1 2 2}}", "", tVal, false},
		// This should fail the type checker - {"slice of an interface field", "{{slice .Empty3 0 1}}", "[7]", tVal, true},

		// Len.
		{"slice", "{{len .SI}}", "3", tVal, true},
		{"map", "{{len .MSI }}", "3", tVal, true},
		{"len of int", "{{len 3}}", "", tVal, false},
		{"len of nothing", "{{len .Empty0}}", "", tVal, false},
		// This should fail the type checker - {"len of an interface field", "{{len .Empty3}}", "2", tVal, true},

		// With.
		{"with true", "{{with true}}{{.}}{{end}}", "true", tVal, true},
		{"with false", "{{with false}}{{.}}{{else}}FALSE{{end}}", "FALSE", tVal, true},
		{"with 1", "{{with 1}}{{.}}{{else}}ZERO{{end}}", "1", tVal, true},
		{"with 0", "{{with 0}}{{.}}{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"with 1.5", "{{with 1.5}}{{.}}{{else}}ZERO{{end}}", "1.5", tVal, true},
		{"with 0.0", "{{with .FloatZero}}{{.}}{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"with 1.5i", "{{with 1.5i}}{{.}}{{else}}ZERO{{end}}", "(0+1.5i)", tVal, true},
		{"with 0.0i", "{{with .ComplexZero}}{{.}}{{else}}ZERO{{end}}", "ZERO", tVal, true},
		{"with emptystring", "{{with ``}}{{.}}{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"with string", "{{with `notempty`}}{{.}}{{else}}EMPTY{{end}}", "notempty", tVal, true},
		{"with emptyslice", "{{with .SIEmpty}}{{.}}{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"with slice", "{{with .SI}}{{.}}{{else}}EMPTY{{end}}", "[3 4 5]", tVal, true},
		{"with emptymap", "{{with .MSIEmpty}}{{.}}{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"with map", "{{with .MSIone}}{{.}}{{else}}EMPTY{{end}}", "map[one:1]", tVal, true},
		// {"with empty interface, struct field", "{{with .Empty4}}{{.V}}{{end}}", "UinEmpty", tVal, true},
		{"with $x int", "{{with $x := .I}}{{$x}}{{end}}", "17", tVal, true},
		{"with $x struct.U.V", "{{with $x := $}}{{$x.U.V}}{{end}}", "v", tVal, true},
		{"with variable and action", "{{with $x := $}}{{$y := $.U.V}}{{$y}}{{end}}", "v", tVal, true},
		{"with on typed nil interface value", "{{with .NonEmptyInterfaceTypedNil}}TRUE{{ end }}", "", tVal, true},
		{"with else with", "{{with 0}}{{.}}{{else with true}}{{.}}{{end}}", "true", tVal, true},
		{"with else with chain", "{{with 0}}{{.}}{{else with false}}{{.}}{{else with `notempty`}}{{.}}{{end}}", "notempty", tVal, true},

		// Range.
		{"range []int", "{{range .SI}}-{{.}}-{{end}}", "-3--4--5-", tVal, true},
		{"range empty no else", "{{range .SIEmpty}}-{{.}}-{{end}}", "", tVal, true},
		{"range []int else", "{{range .SI}}-{{.}}-{{else}}EMPTY{{end}}", "-3--4--5-", tVal, true},
		{"range empty else", "{{range .SIEmpty}}-{{.}}-{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		{"range []int break else", "{{range .SI}}-{{.}}-{{break}}NOTREACHED{{else}}EMPTY{{end}}", "-3-", tVal, true},
		{"range []int continue else", "{{range .SI}}-{{.}}-{{continue}}NOTREACHED{{else}}EMPTY{{end}}", "-3--4--5-", tVal, true},
		{"range []bool", "{{range .SB}}-{{.}}-{{end}}", "-true--false-", tVal, true},
		{"range []int method", "{{range .SI | .MAdd .I}}-{{.}}-{{end}}", "-20--21--22-", tVal, true},
		{"range map", "{{range .MSI}}-{{.}}-{{end}}", "-1--3--2-", tVal, true},
		{"range empty map no else", "{{range .MSIEmpty}}-{{.}}-{{end}}", "", tVal, true},
		{"range map else", "{{range .MSI}}-{{.}}-{{else}}EMPTY{{end}}", "-1--3--2-", tVal, true},
		{"range empty map else", "{{range .MSIEmpty}}-{{.}}-{{else}}EMPTY{{end}}", "EMPTY", tVal, true},
		// {"range empty interface", "{{range .Empty3}}-{{.}}-{{else}}EMPTY{{end}}", "-7--8-", tVal, true},
		// {"range empty nil", "{{range .Empty0}}-{{.}}-{{end}}", "", tVal, true},
		{"range $x SI", "{{range $x := .SI}}<{{$x}}>{{end}}", "<3><4><5>", tVal, true},
		{"range $x $y SI", "{{range $x, $y := .SI}}<{{$x}}={{$y}}>{{end}}", "<0=3><1=4><2=5>", tVal, true},
		{"range $x MSIone", "{{range $x := .MSIone}}<{{$x}}>{{end}}", "<1>", tVal, true},
		{"range $x $y MSIone", "{{range $x, $y := .MSIone}}<{{$x}}={{$y}}>{{end}}", "<one=1>", tVal, true},
		{"range $x PSI", "{{range $x := .PSI}}<{{$x}}>{{end}}", "<21><22><23>", tVal, true},
		{"declare in range", "{{range $x := .PSI}}<{{$foo:=$x}}{{$x}}>{{end}}", "<21><22><23>", tVal, true},
		{"range count", `{{range $i, $x := count 5}}[{{$i}}]{{$x}}{{end}}`, "[0]a[1]b[2]c[3]d[4]e", tVal, true},
		{"range nil count", `{{range $i, $x := count 0}}{{else}}empty{{end}}`, "empty", tVal, true},

		// Cute examples.
		{"or as if true", `{{or .SI "slice is empty"}}`, "[3 4 5]", tVal, true},
		{"or as if false", `{{or .SIEmpty "slice is empty"}}`, "slice is empty", tVal, true},

		// Error handling.
		// The types are cromulent. This test shall pass - {"error method, error", "{{.MyError true}}", "", tVal, false},
		{"error method, no error", "{{.MyError false}}", "false", tVal, true},

		// Numbers
		{"decimal", "{{print 1234}}", "1234", tVal, true},
		{"decimal _", "{{print 12_34}}", "1234", tVal, true},
		{"binary", "{{print 0b101}}", "5", tVal, true},
		{"binary _", "{{print 0b_1_0_1}}", "5", tVal, true},
		{"BINARY", "{{print 0B101}}", "5", tVal, true},
		{"octal0", "{{print 0377}}", "255", tVal, true},
		{"octal", "{{print 0o377}}", "255", tVal, true},
		{"octal _", "{{print 0o_3_7_7}}", "255", tVal, true},
		{"OCTAL", "{{print 0O377}}", "255", tVal, true},
		{"hex", "{{print 0x123}}", "291", tVal, true},
		{"hex _", "{{print 0x1_23}}", "291", tVal, true},
		{"HEX", "{{print 0X123ABC}}", "1194684", tVal, true},
		{"float", "{{print 123.4}}", "123.4", tVal, true},
		{"float _", "{{print 0_0_1_2_3.4}}", "123.4", tVal, true},
		{"hex float", "{{print +0x1.ep+2}}", "7.5", tVal, true},
		{"hex float _", "{{print +0x_1.e_0p+0_2}}", "7.5", tVal, true},
		{"HEX float", "{{print +0X1.EP+2}}", "7.5", tVal, true},
		{"print multi", "{{print 1_2_3_4 7.5_00_00_00}}", "1234 7.5", tVal, true},
		{"print multi2", "{{print 1234 0x0_1.e_0p+02}}", "1234 7.5", tVal, true},

		// Fixed bugs.
		// Must separate dot and receiver; otherwise args are evaluated with dot set to variable.
		{"bug0", "{{range .MSIone}}{{if $.Method1 .}}X{{end}}{{end}}", "X", tVal, true},
		// Do not loop endlessly in indirect for non-empty interfaces.
		// The bug appears with *interface only; looped forever.
		{"bug1", "{{.Method0}}", "M0", &iVal, true},
		// Was taking address of interface field, so method set was empty.
		{"bug2", "{{$.NonEmptyInterface.Method0}}", "M0", tVal, true},
		// Struct values were not legal in with - mere oversight.
		{"bug3", "{{with $}}{{.Method0}}{{end}}", "M0", tVal, true},
		// Nil interface values in if.
		{"bug4", "{{if .Empty0}}non-nil{{else}}nil{{end}}", "nil", tVal, true},
		// Stringer.
		{"bug5", "{{.Str}}", "foozle", tVal, true},
		{"bug5a", "{{.Err}}", "erroozle", tVal, true},
		// Args need to be indirected and dereferenced sometimes.
		{"bug6a", "{{vfunc .V0 .V1}}", "vfunc", tVal, true},
		{"bug6b", "{{vfunc .V0 .V0}}", "vfunc", tVal, true},
		{"bug6c", "{{vfunc .V1 .V0}}", "vfunc", tVal, true},
		{"bug6d", "{{vfunc .V1 .V1}}", "vfunc", tVal, true},
		// Legal parse but illegal execution: non-function should have no arguments.
		{"bug7a", "{{3 2}}", "", tVal, false},
		{"bug7b", "{{$x := 1}}{{$x 2}}", "", tVal, false},
		{"bug7c", "{{$x := 1}}{{3 | $x}}", "", tVal, false},
		// Pipelined arg was not being type-checked.
		{"bug8a", "{{3|oneArg}}", "", tVal, false},
		{"bug8b", "{{4|dddArg 3}}", "", tVal, false},
		// A bug was introduced that broke map lookups for lower-case names.
		{"bug9", "{{.cause}}", "neglect", map[string]string{"cause": "neglect"}, true},
		// Field chain starting with function did not work.
		{"bug10", "{{mapOfThree.three}}-{{(mapOfThree).three}}", "3-3", 0, true}, // note type change
		// Dereferencing nil pointer while evaluating function arguments should not panic. Issue 7333.
		// this is an exec error, type checking should not fail - {"bug11", "{{valueString .PS}}", "", T{}, false},
		// 0xef gave constant type float64. Issue 8622.
		{"bug12xe", "{{printf `%T` 0xef}}", "int", T{}, true},
		{"bug12xE", "{{printf `%T` 0xEE}}", "int", T{}, true},
		{"bug12Xe", "{{printf `%T` 0Xef}}", "int", T{}, true},
		{"bug12XE", "{{printf `%T` 0XEE}}", "int", T{}, true},
		// Chained nodes did not work as arguments. Issue 8473.
		{"bug13", "{{print (.Copy).I}}", "17", tVal, true},
		// Didn't protect against nil or literal values in field chains.
		{"bug14a", "{{(nil).True}}", "", tVal, false},
		{"bug14b", "{{$x := nil}}{{$x.anything}}", "", tVal, false},
		{"bug14c", `{{$x := (1.0)}}{{$y := ("hello")}}{{$x.anything}}{{$y.true}}`, "", tVal, false},
		// Didn't call validateType on function results. Issue 10800.
		{"bug15", "{{valueString returnInt}}", "", tVal, false},
		// Variadic function corner cases. Issue 10946.
		{"bug16a", "{{true|printf}}", "", tVal, false},
		{"bug16b", "{{1|printf}}", "", tVal, false},
		{"bug16c", "{{1.1|printf}}", "", tVal, false},
		{"bug16d", "{{'x'|printf}}", "", tVal, false},
		{"bug16e", "{{0i|printf}}", "", tVal, false},
		{"bug16f", "{{true|twoArgs \"xxx\"}}", "", tVal, false},
		{"bug16g", "{{\"aaa\" |twoArgs \"bbb\"}}", "twoArgs=bbbaaa", tVal, true},
		{"bug16h", "{{1|oneArg}}", "", tVal, false},
		{"bug16i", "{{\"aaa\"|oneArg}}", "oneArg=aaa", tVal, true},
		{"bug16j", "{{1+2i|printf \"%v\"}}", "(1+2i)", tVal, true},
		{"bug16k", "{{\"aaa\"|printf }}", "aaa", tVal, true},
		// bug17 not relevant, type checking does not eval the static type under an interface.
		//{"bug17a", "{{.NonEmptyInterface.X}}", "x", tVal, true},
		//{"bug17b", "-{{.NonEmptyInterface.Method1 1234}}-", "-1234-", tVal, true},
		//{"bug17c", "{{len .NonEmptyInterfacePtS}}", "2", tVal, true},
		//{"bug17d", "{{index .NonEmptyInterfacePtS 0}}", "a", tVal, true},
		//{"bug17e", "{{range .NonEmptyInterfacePtS}}-{{.}}-{{end}}", "-a--b-", tVal, true},

		// More variadic function corner cases. Some runes would get evaluated
		// as constant floats instead of ints. Issue 34483.
		{"bug18a", "{{eq . '.'}}", "true", '.', true},
		{"bug18b", "{{eq . 'e'}}", "true", 'e', true},
		{"bug18c", "{{eq . 'P'}}", "true", 'P', true},

		{"issue56490", "{{$i := 0}}{{$x := 0}}{{range $i = .AI}}{{end}}{{$i}}", "5", tVal, true},
		{"issue60801", "{{$k := 0}}{{$v := 0}}{{range $k, $v = .AI}}{{$k}}={{$v}} {{end}}", "0=3 1=4 2=5 ", tVal, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New(tt.name).Funcs(funcExec).Parse(tt.input)
			if err != nil {
				t.Errorf("%s: parse error: %s", tt.name, err)
				return
			}
			execErr := tmpl.Execute(io.Discard, tt.data)

			dataType := stdlibTestRowType(t, testPkg, ttRows, tt.name)
			require.NotNil(t, dataType)

			checkErr := typelate2.Check(tmpl.Tree, dataType, testPkg.Types, testPkg.Fset, findTextTree(tmpl), MortalFunctions(funcSource))
			switch {
			case !tt.ok && checkErr == nil:
				t.Logf("exec error: %s", execErr)
				t.Errorf("%s: expected error; got none", tt.name)
				return
			case tt.ok && checkErr != nil:
				t.Errorf("%s: unexpected execute error: %s", tt.name, checkErr)
				return
			case !tt.ok && checkErr != nil:
				// expected error, got one
				if testing.Verbose() {
					fmt.Printf("%s: %s\n\t%s\n", tt.name, tt.input, checkErr)
				}
			}
		})
	}
}

func stdlibTestRowType(t *testing.T, p *packages.Package, ttRows *ast.CompositeLit, name string) types.Type {
	t.Helper()
	for _, r := range ttRows.Elts {
		row, ok := r.(*ast.CompositeLit)
		if !ok {
			continue
		}
		if len(row.Elts) < 1 {
			continue
		}
		lit, ok := row.Elts[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		n, err := strconv.Unquote(lit.Value)
		if err != nil {
			continue
		}
		if name != n {
			continue
		}
		var buf bytes.Buffer
		require.NoError(t, format.Node(&buf, p.Fset, row.Elts[3]))
		result, err := types.Eval(p.Fset, p.Types, row.Elts[3].Pos(), buf.String())
		require.NoError(t, err)
		tp := result.Type
		require.NotNil(t, tp)
		return tp
	}
	t.Fatalf("failed to load row name %q", name)
	return nil
}

type MortalFunctions typelate2.Functions

func (fn MortalFunctions) CheckCall(name string, nodes []parse.Node, args []types.Type) (types.Type, error) {
	switch name {
	case "die":
		return nil, fmt.Errorf("exec error die")
	default:
		return typelate2.Functions(fn).CheckCall(name, nodes, args)
	}
}
