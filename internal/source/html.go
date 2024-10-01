package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/crhntr/dom/spec"
	"golang.org/x/net/html/atom"
)

type ValidationGenerator interface {
	GenerateValidation(imports *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt
}

func ParseInputValidations(name string, input spec.Element, tp ast.Expr) ([]ValidationGenerator, error) {
	if tag := strings.ToLower(input.TagName()); tag != atom.Input.String() {
		return nil, fmt.Errorf("expected element to have tag <input> got <%s>", tag)
	}
	var result []ValidationGenerator
	if input.HasAttribute("min") {
		val := input.GetAttribute("min")
		_, err := ParseStringWithType(val, tp)
		if err != nil {
			return nil, err
		}
		result = append(result, MinValidation{
			Name:   name,
			MinExp: &ast.BasicLit{Value: val, Kind: token.INT},
		})
	}
	return result, nil
}
