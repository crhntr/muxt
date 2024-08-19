package muxt

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateName_ByPathThenMethod(t *testing.T) {
	for _, tt := range []struct {
		Name    string
		In, Exp []TemplateName
	}{
		{
			Name: "sort by path then method",
			In: mustNewTemplateName(
				"GET /b",
				"POST /a",
				"GET /a",
			),
			Exp: mustNewTemplateName(
				"GET /a",
				"POST /a",
				"GET /b",
			),
		},
		{
			Name: "sort just paths",
			In: mustNewTemplateName(
				"/b",
				"/c",
				"/a",
			),
			Exp: mustNewTemplateName(
				"/a",
				"/b",
				"/c",
			),
		},
		{
			Name: "sort just methods",
			In: mustNewTemplateName(
				"DELETE /",
				"POST /",
				"GET /",
				"PATCH /",
			),
			Exp: mustNewTemplateName(
				"DELETE /",
				"GET /",
				"PATCH /",
				"POST /",
			),
		},
		{
			// this is blocked higher up in parsing templates but this is lower down so if a
			// caller does not use TemplatePatterns they get consistent results
			Name: "method and path are the same",
			In: mustNewTemplateName(
				"GET / F2()",
				"GET / F3()",
				"GET / F1()",
			),
			Exp: mustNewTemplateName(
				"GET / F1()",
				"GET / F2()",
				"GET / F3()",
			),
		},
		{
			// this is blocked higher up in parsing templates but this is lower down so if a
			// caller does not use TemplatePatterns they get consistent results
			Name: "method and path are the same",
			In: mustNewTemplateName(
				"GET / F2()",
				"GET / F3()",
				"GET / F1()",
			),
			Exp: mustNewTemplateName(
				"GET / F1()",
				"GET / F2()",
				"GET / F3()",
			),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			slices.SortFunc(tt.In, TemplateName.byPathThenMethod)
			assert.Equal(t, tt.Exp, tt.In)
		})
	}
}

func mustNewTemplateName(in ...string) []TemplateName {
	var result []TemplateName
	for _, n := range in {
		p, err, _ := NewTemplateName(n)
		if err != nil {
			panic(err)
		}
		result = append(result, p)
	}
	return result
}
