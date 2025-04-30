package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/typelate/muxt/internal/source"
)

func TestIterateFieldTypes(t *testing.T) {
	t.Run("multiple names per param", func(t *testing.T) {
		exp, err := parser.ParseExpr(`func (a, b, c int, x, y, z float64) {}`)
		require.NoError(t, err)
		expIndex := 0
		for gotIndex, tp := range source.IterateFieldTypes(exp.(*ast.FuncLit).Type.Params.List) {
			assert.NotNil(t, tp)
			assert.Equal(t, expIndex, gotIndex)
			expIndex++
		}
		assert.Equal(t, 6, expIndex)
	})
	t.Run("just types", func(t *testing.T) {
		exp, err := parser.ParseExpr(`func (int, float64) {}`)
		require.NoError(t, err)
		expIndex := 0
		for gotIndex, tp := range source.IterateFieldTypes(exp.(*ast.FuncLit).Type.Params.List) {
			assert.NotNil(t, tp)
			assert.Equal(t, expIndex, gotIndex)
			expIndex++
		}
		assert.Equal(t, 2, expIndex)
	})
}

func TestHTTPStatusCode(t *testing.T) {
	imports := source.NewImports(nil)
	exp := source.HTTPStatusCode(imports, 600)
	require.NotNil(t, exp)
	lit, ok := exp.(*ast.BasicLit)
	require.True(t, ok)
	require.Equal(t, token.INT, lit.Kind)
	require.Equal(t, "600", lit.Value)
	require.Nil(t, imports.GenDecl, "it should not add the import if it is not needed")
}
