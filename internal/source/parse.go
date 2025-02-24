package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"regexp"
	"strconv"

	"github.com/crhntr/dom/spec"
)

func GenerateValidations(imports *Imports, variable ast.Expr, variableType types.Type, inputQuery, inputName, responseIdent string, fragment spec.DocumentFragment) ([]ast.Stmt, error, bool) {
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
		statements = append(statements, validation.GenerateValidation(imports, variable, func(message string) ast.Stmt {
			return &ast.ExprStmt{X: imports.HTTPErrorCall(ast.NewIdent(responseIdent), &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(message),
			}, http.StatusBadRequest)}
		}))
	}
	return statements, nil, true
}

type MinValidation struct {
	Name   string
	MinExp ast.Expr
}

func (val MinValidation) GenerateValidation(_ *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  variable,
			Op: token.LSS, // value < 13
			Y:  val.MinExp,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				handleError(fmt.Sprintf("%s must not be less than %s", val.Name, Format(val.MinExp))),
				&ast.ReturnStmt{},
			},
		},
	}
}

type MaxValidation struct {
	Name   string
	MinExp ast.Expr
}

func (val MaxValidation) GenerateValidation(_ *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  variable,
			Op: token.GTR, // value > 13
			Y:  val.MinExp,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				handleError(fmt.Sprintf("%s must not be more than %s", val.Name, Format(val.MinExp))),
				&ast.ReturnStmt{},
			},
		},
	}
}

type PatternValidation struct {
	Name string
	Exp  *regexp.Regexp
}

func (val PatternValidation) GenerateValidation(imports *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt {
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
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				handleError(fmt.Sprintf("%s must match %q", val.Name, val.Exp.String())),
				&ast.ReturnStmt{},
			},
		},
	}
}

type MaxLengthValidation struct {
	Name      string
	MaxLength int
}

func (val MaxLengthValidation) GenerateValidation(_ *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{variable}},
			Op: token.GTR,
			Y:  Int(val.MaxLength),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				handleError(fmt.Sprintf("%s is too long (the max length is %d)", val.Name, val.MaxLength)),
				&ast.ReturnStmt{},
			},
		},
	}
}

type MinLengthValidation struct {
	Name      string
	MinLength int
}

func (val MinLengthValidation) GenerateValidation(_ *Imports, variable ast.Expr, handleError func(string) ast.Stmt) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{variable}},
			Op: token.LSS,
			Y:  Int(val.MinLength),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				handleError(fmt.Sprintf("%s is too short (the min length is %d)", val.Name, val.MinLength)),
				&ast.ReturnStmt{},
			},
		},
	}
}
