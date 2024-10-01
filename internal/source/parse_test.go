package source_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crhntr/dom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/crhntr/muxt/internal/source"
)

func Test_inputValidations(t *testing.T) {
	for _, tt := range []struct {
		Name     string
		Type     string
		Template string
		Result   string
		Error    string
	}{
		{
			Type:     "int",
			Name:     "no attributes",
			Template: `<input type="number" name="field">`,
			Result: `{
}`,
		},
		{
			Type:     "int",
			Name:     "min",
			Template: `<input type="number" name="field" min="100">`,
			Result: `{
	if v < 100 {
		http.Error(response, "field must not be less than 100", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int",
			Name:     "negative min",
			Template: `<input type="number" name="field" min="-5">`,
			Result: `{
	if v < -5 {
		http.Error(response, "field must not be less than -5", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int8",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int16",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int32",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int64",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "uint",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "uint8",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "uint16",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "uint32",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "uint64",
			Name:     "zero min",
			Template: `<input type="number" name="field" min="0">`,
			Result: `{
	if v < 0 {
		http.Error(response, "field must not be less than 0", http.StatusBadRequest)
		return
	}
}`,
		},
		{
			Type:     "int",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="18446744073709551616">`,
			Error:    `strconv.ParseInt: parsing "18446744073709551616": value out of range`,
		},
		{
			Type:     "int8",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="256">`,
			Error:    `strconv.ParseInt: parsing "256": value out of range`,
		},
		{
			Type:     "int16",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="32768">`,
			Error:    `strconv.ParseInt: parsing "32768": value out of range`,
		},
		{
			Type:     "int32",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="2147483648">`,
			Error:    `strconv.ParseInt: parsing "2147483648": value out of range`,
		},
		{
			Type:     "int64",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="9223372036854775808">`,
			Error:    `strconv.ParseInt: parsing "9223372036854775808": value out of range`,
		},
		{
			Type:     "uint",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="-10">`,
			Error:    `strconv.ParseUint: parsing "-10": invalid syntax`,
		},
		{
			Type:     "uint8",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="256">`,
			Error:    `strconv.ParseUint: parsing "256": value out of range`,
		},
		{
			Type:     "uint16",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="65536">`,
			Error:    `strconv.ParseUint: parsing "65536": value out of range`,
		},
		{
			Type:     "uint32",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="4294967296">`,
			Error:    `strconv.ParseUint: parsing "4294967296": value out of range`,
		},
		{
			Type:     "uint64",
			Name:     "out of range",
			Template: `<input type="number" name="field" min="18446744073709551616">`,
			Error:    `strconv.ParseUint: parsing "18446744073709551616": value out of range`,
		},
		{
			Type:     "*T",
			Name:     "unsupported type",
			Template: `<input type="number" name="field" min="1">`,
			Error:    `type *T is not supported`,
		},
		{
			Type:     "T",
			Name:     "type unknown",
			Template: `<input type="number" name="field" min="1">`,
			Error:    `type T unknown`,
		},
		{
			Type:     "int",
			Name:     "not a number",
			Template: `<input type="number" name="field" min="NaN">`,
			Error:    `strconv.ParseInt: parsing "NaN": invalid syntax`,
		},
		{
			Type:     "int",
			Name:     "wrong tag",
			Template: `<form type="number" name="field" min="32"></form>`,
			Error:    `expected element to have tag <input> got <form>`,
		},
		{
			Type:     "uint32",
			Name:     "zero max",
			Template: `<input type="number" name="field" max="0">`,
			Result: `{
	if v > 0 {
		http.Error(response, "field must not be more than 0", http.StatusBadRequest)
		return
	}
}`,
		},
	} {
		t.Run(fmt.Sprintf("cromulent attribute type %s %s", tt.Type, tt.Name), func(t *testing.T) {
			v := ast.NewIdent("v")
			tp, err := parser.ParseExpr(tt.Type)
			require.NoError(t, err)
			ts := template.Must(template.New("").Parse(tt.Template))
			nodes, err := html.ParseFragment(strings.NewReader(ts.Tree.Root.String()), &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Body,
				Data:     atom.Body.String(),
			})
			fragment := dom.NewDocumentFragment(nodes)
			imports := source.NewImports(nil)
			statements, err, ok := source.GenerateValidations(imports, v, tp, `[name="field"]`, "field", "response", fragment)
			require.True(t, ok)
			if tt.Error != "" {
				require.Error(t, err)
				assert.Equal(t, tt.Error, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.Result, source.Format(&ast.BlockStmt{List: statements}))
			}
		})
	}
}
