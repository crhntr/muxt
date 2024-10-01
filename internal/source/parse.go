package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"net/http"
	"regexp"
	"slices"
	"strconv"

	"github.com/crhntr/dom/spec"
)

func GenerateParseValueFromStringStatements(imports *Imports, tmp string, errVarIdent string, str, typeExp ast.Expr, errCheck *ast.IfStmt, validations []ast.Stmt, assignment func(ast.Expr) ast.Stmt) ([]ast.Stmt, error) {
	paramTypeIdent, ok := typeExp.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("unsupported type: %s", Format(typeExp))
	}
	switch paramTypeIdent.Name {
	default:
		return nil, fmt.Errorf("method param type %s not supported", Format(typeExp))
	case "bool":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseBoolCall(str)},
		}

		assign := assignment(ast.NewIdent(tmp))
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "int":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvAtoiCall(str)},
		}

		assign := assignment(ast.NewIdent(tmp))
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "int8":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseIntCall(str, 10, 8)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "int16":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseIntCall(str, 10, 16)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "int32":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseIntCall(str, 10, 32)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})

		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "int64":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseIntCall(str, 10, 64)},
		}

		assign := assignment(ast.NewIdent(tmp))
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "uint":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseUintCall(str, 10, 64)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "uint8":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseUintCall(str, 10, 8)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "uint16":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseUintCall(str, 10, 16)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "uint32":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseUintCall(str, 10, 32)},
		}

		assign := assignment(&ast.CallExpr{
			Fun:  ast.NewIdent(paramTypeIdent.Name),
			Args: []ast.Expr{ast.NewIdent(tmp)},
		})
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "uint64":
		parse := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp), ast.NewIdent(errVarIdent)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{imports.StrconvParseUintCall(str, 10, 64)},
		}

		assign := assignment(ast.NewIdent(tmp))
		statements := slices.Concat([]ast.Stmt{parse, errCheck}, validations, []ast.Stmt{assign})
		return statements, nil
	case "string":
		if len(validations) == 0 {
			assign := assignment(str)
			statements := slices.Concat(validations, []ast.Stmt{assign})
			return statements, nil
		}
		statements := slices.Concat([]ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tmp)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{str},
		}}, validations, []ast.Stmt{assignment(ast.NewIdent(tmp))})
		return statements, nil
	}
}

func GenerateValidations(imports *Imports, variable, variableType ast.Expr, inputQuery, inputName, responseIdent string, fragment spec.DocumentFragment) ([]ast.Stmt, error, bool) {
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
			return imports.HTTPErrorCall(ast.NewIdent(responseIdent), &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(message),
			}, http.StatusBadRequest)
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
