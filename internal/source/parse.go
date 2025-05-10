package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"

	"github.com/crhntr/dom/spec"
)

func GenerateValidations(imports *File, variable ast.Expr, variableType types.Type, inputQuery, inputName, responseIdent string, fragment spec.DocumentFragment, validationFailureBlock ValidationErrorBlock) ([]ast.Stmt, error, bool) {
	input := fragment.QuerySelector(inputQuery)
	if input == nil {
		return nil, nil, false
	}

	validations, err := ParseInputValidations(inputName, input, variableType)
	if err != nil {
		return nil, err, true
	}

	var statements []ast.Stmt
	for _, validation := range validations {
		statements = append(statements, validation.GenerateValidation(imports, variable, validationFailureBlock))
	}
	return statements, nil, true
}

type MinValidation struct {
	Name   string
	MinExp ast.Expr
}

func (val MinValidation) GenerateValidation(_ *File, variable ast.Expr, handleError ValidationErrorBlock) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  variable,
			Op: token.LSS, // value < 13
			Y:  val.MinExp,
		},
		Body: handleError(fmt.Sprintf("%s must not be less than %s", val.Name, Format(val.MinExp))),
	}
}

type MaxValidation struct {
	Name   string
	MinExp ast.Expr
}

func (val MaxValidation) GenerateValidation(_ *File, variable ast.Expr, handleError ValidationErrorBlock) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  variable,
			Op: token.GTR, // value > 13
			Y:  val.MinExp,
		},
		Body: handleError(fmt.Sprintf("%s must not be more than %s", val.Name, Format(val.MinExp))),
	}
}

type PatternValidation struct {
	Name string
	Exp  *regexp.Regexp
}

func (val PatternValidation) GenerateValidation(imports *File, variable ast.Expr, handleError ValidationErrorBlock) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   imports.Call("", "regexp", "MustCompile", []ast.Expr{String(val.Exp.String())}),
					Sel: ast.NewIdent("MatchString"),
				},
				Args: []ast.Expr{variable},
			},
		},
		Body: handleError(fmt.Sprintf("%s must match %q", val.Name, val.Exp.String())),
	}
}

type MaxLengthValidation struct {
	Name      string
	MaxLength int
}

func (val MaxLengthValidation) GenerateValidation(_ *File, variable ast.Expr, handleError ValidationErrorBlock) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{variable}},
			Op: token.GTR,
			Y:  Int(val.MaxLength),
		},
		Body: handleError(fmt.Sprintf("%s is too long (the max length is %d)", val.Name, val.MaxLength)),
	}
}

type MinLengthValidation struct {
	Name      string
	MinLength int
}

func (val MinLengthValidation) GenerateValidation(_ *File, variable ast.Expr, handleError ValidationErrorBlock) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{variable}},
			Op: token.LSS,
			Y:  Int(val.MinLength),
		},
		Body: handleError(fmt.Sprintf("%s is too short (the min length is %d)", val.Name, val.MinLength)),
	}
}
