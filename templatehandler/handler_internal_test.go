package templatehandler

import (
	"go/ast"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_endpoint(t *testing.T) {
	for _, tt := range []struct {
		Name         string
		TemplateName string
		ExpMatch     bool
		Pattern      func(t *testing.T, pat Pattern)
		Error        func(t *testing.T, err error)
	}{
		{
			Name:         "get root",
			TemplateName: "GET /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, Pattern{
					Method:  http.MethodGet,
					Host:    "",
					Path:    "/",
					Pattern: "GET /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "post root",
			TemplateName: "POST /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, Pattern{
					Method:  http.MethodPost,
					Host:    "",
					Path:    "/",
					Pattern: "POST /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "patch root",
			TemplateName: "PATCH /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, Pattern{
					Method:  http.MethodPatch,
					Host:    "",
					Path:    "/",
					Pattern: "PATCH /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "delete root",
			TemplateName: "DELETE /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, Pattern{
					Method:  http.MethodDelete,
					Host:    "",
					Path:    "/",
					Pattern: "DELETE /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "put root",
			TemplateName: "PUT /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, Pattern{
					Method:  http.MethodPut,
					Host:    "",
					Path:    "/",
					Pattern: "PUT /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "put root",
			TemplateName: "OPTIONS /",
			ExpMatch:     true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "OPTIONS method not allowed")
			},
		},
		{
			Name:         "handle method call with path parameter",
			TemplateName: "GET /article/{articleID} x.ReadArticle(req, articleID)",
			ExpMatch:     true,
			Error: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			Pattern: func(t *testing.T, pat Pattern) {
				assert.Equal(t, "GET", pat.Method)
				assert.Equal(t, "/article/{articleID}", pat.Path)
				assert.Equal(t, "x.ReadArticle(req, articleID)", pat.Handler)
				assert.Equal(t, &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "x", NamePos: 1},
						Sel: &ast.Ident{NamePos: 3, Name: "ReadArticle"},
					},
					Lparen: 14,
					Args: []ast.Expr{
						&ast.Ident{Name: "req", NamePos: 15},
						&ast.Ident{Name: "articleID", NamePos: 20},
					},
					Rparen: 29,
				}, pat.handler)
			},
		},
		{
			Name:         "handler is not an expression",
			TemplateName: "GET /article/{articleID} T int",
			ExpMatch:     true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "failed to parse handler expression")
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			pat, err, match := endpoint(tt.TemplateName)
			require.Equal(t, tt.ExpMatch, match)
			if tt.Error != nil {
				tt.Error(t, err)
			} else if tt.Pattern != nil {
				assert.NoError(t, err)
				tt.Pattern(t, pat)
			}
		})
	}

}
