package source

import (
	"go/ast"
	"go/token"
	"log"
	"path"
	"slices"
	"strconv"
	"strings"
)

type Imports struct {
	*ast.GenDecl
}

func NewImports(decl *ast.GenDecl) *Imports {
	if decl != nil {
		if got := decl.Tok; got != token.IMPORT {
			log.Panicf("expected decl to have token.IMPORT Tok got %s", got)
		}
	}
	return &Imports{GenDecl: decl}
}

func (imports *Imports) Add(pkgIdent, pkgPath string) string {
	if imports.GenDecl == nil {
		imports.GenDecl = new(ast.GenDecl)
		imports.GenDecl.Tok = token.IMPORT
	}
	if pkgIdent == "" {
		pkgIdent = path.Base(pkgPath)
	}
	for _, s := range imports.GenDecl.Specs {
		spec := s.(*ast.ImportSpec)
		pp, _ := strconv.Unquote(spec.Path.Value)
		if pp == pkgPath {
			if spec.Name != nil && spec.Name.Name != "" && spec.Name.Name != pkgIdent {
				return spec.Name.Name
			}
			return path.Base(pp)
		}
	}
	var pi *ast.Ident
	if path.Base(pkgPath) != pkgIdent {
		pi = Ident(pkgIdent)
	}
	imports.GenDecl.Specs = append(imports.GenDecl.Specs, &ast.ImportSpec{
		Path: String(pkgPath),
		Name: pi,
	})
	slices.SortFunc(imports.GenDecl.Specs, func(a, b ast.Spec) int {
		return strings.Compare(a.(*ast.ImportSpec).Path.Value, b.(*ast.ImportSpec).Path.Value)
	})
	return pkgIdent
}

func (imports *Imports) Ident(pkgPath string) string {
	if imports != nil && imports.GenDecl != nil {
		for _, s := range imports.GenDecl.Specs {
			spec := s.(*ast.ImportSpec)
			pp, _ := strconv.Unquote(spec.Path.Value)
			if pp == pkgPath {
				if spec.Name != nil && spec.Name.Name != "" {
					return spec.Name.Name
				}
				return path.Base(pp)
			}
		}
	}
	return path.Base(pkgPath)
}

func (imports *Imports) Call(pkgName, pkgPath, funcIdent string, args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(imports.Add(pkgName, pkgPath)),
			Sel: ast.NewIdent(funcIdent),
		},
		Args: args,
	}
}

func (imports *Imports) ImportSpecs() []*ast.ImportSpec {
	result := make([]*ast.ImportSpec, 0, len(imports.GenDecl.Specs))
	for _, spec := range imports.GenDecl.Specs {
		result = append(result, spec.(*ast.ImportSpec))
	}
	slices.SortFunc(result, func(a, b *ast.ImportSpec) int { return strings.Compare(a.Path.Value, b.Path.Value) })
	return slices.CompactFunc(result, func(a, b *ast.ImportSpec) bool { return a.Path.Value == b.Path.Value })
}

func (imports *Imports) SortImports() {
	sorted := imports.GenDecl.Specs[:0]
	for _, spec := range imports.ImportSpecs() {
		sorted = append(sorted, spec)
	}
	imports.GenDecl.Specs = sorted
}

func (imports *Imports) AddNetHTTP() string      { return imports.Add("", "net/http") }
func (imports *Imports) AddHTMLTemplate() string { return imports.Add("", "html/template") }
func (imports *Imports) AddContext() string      { return imports.Add("", "context") }

func (imports *Imports) HTTPErrorCall(response, message ast.Expr, code int) *ast.CallExpr {
	return imports.Call("", "net/http", "Error", []ast.Expr{
		response,
		message,
		HTTPStatusCode(imports, code),
	})
}

func (imports *Imports) StrconvAtoiCall(expr ast.Expr) *ast.CallExpr {
	return imports.Call("", "strconv", "Atoi", []ast.Expr{expr})
}

func (imports *Imports) StrconvParseIntCall(expr ast.Expr, base, size int) *ast.CallExpr {
	return imports.Call("", "strconv", "ParseInt", []ast.Expr{expr, Int(base), Int(size)})
}

func (imports *Imports) StrconvParseUintCall(expr ast.Expr, base, size int) *ast.CallExpr {
	return imports.Call("", "strconv", "ParseUint", []ast.Expr{expr, Int(base), Int(size)})
}

func (imports *Imports) StrconvParseFloatCall(expr ast.Expr, size int) *ast.CallExpr {
	return imports.Call("", "strconv", "ParseFloat", []ast.Expr{expr, Int(size)})
}

func (imports *Imports) StrconvParseBoolCall(expr ast.Expr) *ast.CallExpr {
	return imports.Call("", "strconv", "ParseBool", []ast.Expr{expr})
}

func (imports *Imports) TimeParseCall(layout string, expr ast.Expr) *ast.CallExpr {
	return imports.Call("", "time", "Parse", []ast.Expr{String(layout), expr})
}
