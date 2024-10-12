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
		Name     string
		In       string
		ExpErr   string
		ExpMatch bool
	}{
		{
			Name:     "no arg post",
			In:       "GET / F()",
			ExpMatch: true,
		},
		{
			Name:     "no arg get",
			In:       "POST / F()",
			ExpMatch: true,
		},
		{
			Name:     "float64 as handler",
			In:       "POST / 1.2",
			ExpMatch: true,
			ExpErr:   `failed to parse status code: strconv.Atoi: parsing "1.2": invalid syntax`,
		},
		{
			Name:     "status constant",
			In:       "POST / http.StatusOK",
			ExpMatch: true,
		},
		{
			Name:     "not an expression",
			In:       "GET / package main",
			ExpMatch: true,
			ExpErr:   "failed to parse handler expression: ",
		},
		{
			Name:     "function literal",
			In:       "GET / func() {} ()",
			ExpMatch: true,
			ExpErr:   "expected function identifier",
		},
		{
			Name:     "call ellipsis",
			In:       "GET /{fileName} F(fileName...)",
			ExpMatch: true,
			ExpErr:   "unexpected ellipsis",
		},
		{
			Name:     "response arg and status code",
			In:       "GET /{fileName} 201 F(response)",
			ExpMatch: true,
			ExpErr:   "you can not use response as an argument and specify an HTTP status code",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			tn, err, ok := muxt.NewTemplateName(tt.In)
			if assert.Equal(t, tt.ExpMatch, ok) {
				if tt.ExpErr != "" {
					assert.ErrorContains(t, err, tt.ExpErr)
				} else {
					assert.Equal(t, tt.In, tn.String())
					assert.NoError(t, err)
				}
			}
		})
	}
}
