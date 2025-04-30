package source_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/typelate/muxt/internal/source"
)

func TestImports(t *testing.T) {
	t.Run("it returns the package path base", func(t *testing.T) {
		assert.Panics(t, func() {
			source.NewImports(&ast.GenDecl{Tok: token.VAR})
		})
	})
	t.Run("initial add", func(t *testing.T) {
		imports := source.NewImports(nil)
		assert.Equal(t, "http", imports.Add("http", "net/http"))
		assert.Equal(t, source.Format(imports.GenDecl), `import "net/http"`)
	})
	t.Run("initial with pkg ident", func(t *testing.T) {
		imports := source.NewImports(nil)
		assert.Equal(t, "p", imports.Add("p", "net/http"))
		assert.Equal(t, source.Format(imports.GenDecl), `import p "net/http"`)
	})
	t.Run("initial with empty ident", func(t *testing.T) {
		imports := source.NewImports(nil)
		assert.Equal(t, "http", imports.Add("", "net/http"))
		assert.Equal(t, source.Format(imports.GenDecl), `import "net/http"`)
	})
	t.Run("initial with empty ident", func(t *testing.T) {
		imports := source.NewImports(nil)
		_ = imports.Add("", "net/http")
		_ = imports.Add("", "html/template")
		assert.Equal(t, source.Format(imports.GenDecl), `import (
	"html/template"
	"net/http"
)`)
	})
	t.Run("it respects order", func(t *testing.T) {
		imports := source.NewImports(nil)
		_ = imports.Add("", "html/template")
		_ = imports.Add("", "net/http")
		assert.Equal(t, source.Format(imports.GenDecl), `import (
	"html/template"
	"net/http"
)`)
	})
	t.Run("it returns the registered identifier", func(t *testing.T) {
		imports := source.NewImports(nil)
		_ = imports.Add("t", "html/template")
		assert.Equal(t, "t", imports.Add("", "html/template"))
	})
	t.Run("it returns the package path base", func(t *testing.T) {
		imports := source.NewImports(nil)
		_ = imports.Add("", "html/template")
		assert.Equal(t, "template", imports.Add("", "html/template"))
	})
}
