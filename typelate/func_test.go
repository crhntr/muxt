package typelate_test

import (
	"text/template"
	"text/template/parse"

	"github.com/crhntr/muxt/typelate"
)

func findTextTree(tmpl *template.Template) typelate.FindTreeFunc {
	return func(name string) (*parse.Tree, bool) {
		ts := tmpl.Lookup(name)
		if ts == nil {
			return nil, false
		}
		return ts.Tree, true
	}
}
