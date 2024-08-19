package muxt_test

import (
	"html/template"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crhntr/muxt"
)

func TestTemplateNames(t *testing.T) {
	t.Run("when one of the template names is a malformed pattern", func(t *testing.T) {
		ts := template.Must(template.New("").Parse(`{{define "HEAD /"}}{{end}}`))
		_, err := muxt.TemplateNames(ts)
		require.Error(t, err)
	})
	t.Run("when the pattern is not unique", func(t *testing.T) {
		ts := template.Must(template.New("").Parse(`{{define "GET  / F1()"}}a{{end}} {{define "GET /  F2()"}}b{{end}}`))
		_, err := muxt.TemplateNames(ts)
		require.Error(t, err)
	})
}

func TestNewTemplateName(t *testing.T) {
	for _, tt := range []struct {
		Name         string
		In           string
		ExpMatch     bool
		TemplateName func(t *testing.T, pat muxt.TemplateName)
		Error        func(t *testing.T, err error)
	}{
		{
			Name:     "get root",
			In:       "GET /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodGet,
					Host:    "",
					Path:    "/",
					Route:   "GET /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "multiple spaces after method",
			In:       "GET  /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodGet,
					Host:    "",
					Path:    "/",
					Route:   "GET  /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "post root",
			In:       "POST /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodPost,
					Host:    "",
					Path:    "/",
					Route:   "POST /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "patch root",
			In:       "PATCH /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodPatch,
					Host:    "",
					Path:    "/",
					Route:   "PATCH /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "delete root",
			In:       "DELETE /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodDelete,
					Host:    "",
					Path:    "/",
					Route:   "DELETE /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "put root",
			In:       "PUT /",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodPut,
					Host:    "",
					Path:    "/",
					Route:   "PUT /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "with end of path wildcard",
			In:       "PUT /ping/pong/{$}",
			ExpMatch: true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {
				assert.EqualExportedValues(t, muxt.TemplateName{
					Method:  http.MethodPut,
					Host:    "",
					Path:    "/ping/pong/{$}",
					Route:   "PUT /ping/pong/{$}",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:     "put root",
			In:       "OPTIONS /",
			ExpMatch: true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "OPTIONS method not allowed")
			},
		},
		{
			Name:     "path parameter is not an identifier",
			In:       "GET /{123} F(123)",
			ExpMatch: true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `path parameter name not permitted: "123" is not a Go identifier`)
			},
		},
		{
			Name:     "path end sentential in the middle is not permitted",
			In:       "GET /x/{$}/y F()",
			ExpMatch: true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `path parameter name not permitted: "$" is not a Go identifier`)
			},
		},
		{
			Name:         "path end sentential in the middle is not permitted",
			In:           "GET /x/{$} F()",
			ExpMatch:     true,
			TemplateName: func(t *testing.T, pat muxt.TemplateName) {},
		},
		{
			Name:     "duplicate path parameter name",
			In:       "GET /{name}/{name} F()",
			ExpMatch: true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `forbidden repeated path parameter names: found at least 2 path parameters with name "name"`)
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			pat, err, match := muxt.NewTemplateName(tt.In)
			require.Equal(t, tt.ExpMatch, match)
			if tt.Error != nil {
				tt.Error(t, err)
			} else if tt.TemplateName != nil {
				assert.NoError(t, err)
				tt.TemplateName(t, pat)
			}
		})
	}
}

func TestPattern_parseHandler(t *testing.T) {
	for _, tt := range []struct {
		Name   string
		In     string
		ExpErr string
	}{
		{
			Name: "no arg post",
			In:   "GET / F()",
		},
		{
			Name: "no arg get",
			In:   "POST / F()",
		},
		{
			Name:   "int as handler",
			In:     "POST / 1",
			ExpErr: "expected call, got: 1",
		},
		{
			Name:   "not an expression",
			In:     "GET / package main",
			ExpErr: "failed to parse handler expression: ",
		},
		{
			Name:   "function literal",
			In:     "GET / func() {} ()",
			ExpErr: "expected function identifier",
		},
		{
			Name:   "call ellipsis",
			In:     "GET /{fileName} F(fileName...)",
			ExpErr: "unexpected ellipsis",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			p, err, ok := muxt.NewTemplateName(tt.In)
			require.True(t, ok)
			require.NotZero(t, p.Handler)
			if tt.ExpErr != "" {
				assert.ErrorContains(t, err, tt.ExpErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
