package check_test

import (
	"fmt"
	"go/types"
	"html/template"
	"reflect"
	"slices"
	"testing"
	"text/template/parse"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	"github.com/crhntr/muxt/internal/check"
)

func TestTree(t *testing.T) {
	packageList, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedTypes,
		Tests: true,
	}, ".")
	if err != nil {
		t.Fatal(err)
	}

	var checkTestPackage *packages.Package
	if i := slices.IndexFunc(packageList, func(p *packages.Package) bool {
		return p.Name == "check_test"
	}); i > 0 {
		checkTestPackage = packageList[i]
	} else {
		t.Fatal("no check_test package")
	}

	for _, tt := range []struct {
		Name     string
		Template string
		Type     types.Type
		Error    func(t *testing.T, err error, tp types.Type)
	}{
		{
			Name:     "on an empty template",
			Template: ``,
			Type:     typeFor[EmptyStruct](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when accessing nil on an empty struct",
			Template: `{{.Field}}`,
			Type:     typeFor[EmptyStruct](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.EqualError(t, err, fmt.Sprintf(`type check failed: template:1:2: Field not found on %s`, tp))
			},
		},
		{
			Name:     "when accessing the dot",
			Template: `{{.}}`,
			Type:     typeFor[EmptyStruct](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when a method does not any results",
			Template: `{{.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureNoResultMethod](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				method, _, _ := types.LookupFieldOrMethod(tp, true, checkTestPackage.Types, "Method")
				require.NotNil(t, method)
				methodPos := checkTestPackage.Fset.Position(method.Pos())

				require.EqualError(t, err, fmt.Sprintf(`type check failed: template:1:2: function Method has 0 return values; should be 1 or 2: incorrect signature at %s`, methodPos))
			},
		},
		{
			Name:     "when a method does has a result",
			Template: `{{.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureResult](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when a method also has an error",
			Template: `{{.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureResultAndError](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when a method has a second result that is not an error",
			Template: `{{.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureResultAndNonError](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				method, _, _ := types.LookupFieldOrMethod(tp, true, checkTestPackage.Types, "Method")
				require.NotNil(t, method)
				methodPos := checkTestPackage.Fset.Position(method.Pos())

				require.EqualError(t, err, fmt.Sprintf(`type check failed: template:1:2: invalid function signature for Method: second return value should be error; is int: incorrect signature at %s`, methodPos))
			},
		},
		{
			Name:     "when a method with too many results",
			Template: `{{.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureThreeResults](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				method, _, _ := types.LookupFieldOrMethod(tp, true, checkTestPackage.Types, "Method")
				require.NotNil(t, method)
				methodPos := checkTestPackage.Fset.Position(method.Pos())

				require.EqualError(t, err, fmt.Sprintf(`type check failed: template:1:2: function Method has 3 return values; should be 1 or 2: incorrect signature at %s`, methodPos))
			},
		},
		{
			Name:     "when a method is part of a field node list",
			Template: `{{.Method.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureResultHasMethod](checkTestPackage),
			Error: func(t *testing.T, err error, _ types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when result method does not have a method",
			Template: `{{.Method.Method}}`,
			Type:     typeFor[TypeWithMethodSignatureResultHasMethodWithNoResults](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				m1, _, _ := types.LookupFieldOrMethod(tp, true, checkTestPackage.Types, "Method")
				require.NotNil(t, m1)
				m2, _, _ := types.LookupFieldOrMethod(m1.Type().(*types.Signature).Results().At(0).Type(), true, checkTestPackage.Types, "Method")
				require.NotNil(t, m2)
				methodPos := checkTestPackage.Fset.Position(m2.Pos())

				require.EqualError(t, err, fmt.Sprintf(`type check failed: template:1:9: function Method has 0 return values; should be 1 or 2: incorrect signature at %s`, methodPos))
			},
		},
		{
			Name:     "when the struct has the field",
			Template: `{{.Field}}`,
			Type:     typeFor[StructWithField](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when the struct has the field and the field has a method",
			Template: `{{.Field.Method}}`,
			Type:     typeFor[StructWithFieldWithMethod](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when the struct has the field and the field has a method",
			Template: `{{.Field}}`,
			Type:     typeFor[StructWithFieldWithMethod](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				require.NoError(t, err)
			},
		},
		{
			Name:     "when the struct has the field of kind func",
			Template: `{{.Field.Method}}`,
			Type:     typeFor[StructWithFuncFieldWithResultWithMethod](checkTestPackage),
			Error: func(t *testing.T, err error, tp types.Type) {
				field, _, _ := types.LookupFieldOrMethod(tp, true, checkTestPackage.Types, "Field")
				require.NotNil(t, field)
				require.ErrorContains(t, err, fmt.Sprintf("type check failed: template:1:8: can't evaluate field Field in type %s", field.Type()))
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			templates, err := template.New("template").Parse(tt.Template)
			require.NoError(t, err)

			trees := make(map[string]*parse.Tree)
			for _, ts := range templates.Templates() {
				trees[ts.Tree.Name] = ts.Tree
			}
			fns := make(map[string]*types.Signature)

			if err := check.Tree(templates.Tree, tt.Type, checkTestPackage.Types, checkTestPackage.Fset, trees, fns); tt.Error != nil {
				tt.Error(t, err, tt.Type)
			}
		})
	}
}

func typeFor[T any](pkg *packages.Package) types.Type {
	return pkg.Types.Scope().Lookup(reflect.TypeFor[T]().Name()).Type()
}

type EmptyStruct struct{}

type TypeWithMethodSignatureNoResultMethod struct{}

func (TypeWithMethodSignatureNoResultMethod) Method() {}

type TypeWithMethodSignatureResult struct{}

func (TypeWithMethodSignatureResult) Method() struct{} { return struct{}{} }

type TypeWithMethodSignatureResultAndError struct{}

func (TypeWithMethodSignatureResultAndError) Method() (struct{}, error) { return struct{}{}, nil }

type TypeWithMethodSignatureResultAndNonError struct{}

func (TypeWithMethodSignatureResultAndNonError) Method() (struct{}, int) { return struct{}{}, 0 }

type TypeWithMethodSignatureThreeResults struct{}

func (TypeWithMethodSignatureThreeResults) Method() (struct{}, struct{}, error) {
	return struct{}{}, struct{}{}, nil
}

type TypeWithMethodSignatureResultHasMethod struct{}

func (TypeWithMethodSignatureResultHasMethod) Method() (_ TypeWithMethodSignatureResult) {
	return
}

type TypeWithMethodSignatureResultHasMethodWithNoResults struct{}

func (TypeWithMethodSignatureResultHasMethodWithNoResults) Method() (_ TypeWithMethodSignatureNoResultMethod) {
	return
}

type StructWithField struct {
	Field struct{}
}

type StructWithFieldWithMethod struct {
	Field TypeWithMethodSignatureResultAndError
}

type StructWithFuncFieldWithResultWithMethod struct {
	Field func() TypeWithMethodSignatureResult
}