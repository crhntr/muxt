package muxt_test

import (
	"html/template"
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
			_, err, ok := muxt.NewTemplateName(tt.In)
			require.True(t, ok)
			if tt.ExpErr != "" {
				assert.ErrorContains(t, err, tt.ExpErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
