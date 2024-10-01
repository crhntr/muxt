package source

import (
	"go/ast"
	"go/token"

	"github.com/crhntr/dom/spec"
)

type ValidationGenerator interface {
	GenerateValidation(imports *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt
}

func ParseInputValidations(name string, input spec.Element, tp ast.Expr) ([]ValidationGenerator, error) {
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
