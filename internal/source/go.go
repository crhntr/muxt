package source

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"net/http"
	"strconv"
	"strings"
)

func IterateGenDecl(files []*ast.File, tok token.Token) func(func(*ast.File, *ast.GenDecl) bool) {
	return func(yield func(*ast.File, *ast.GenDecl) bool) {
		for _, file := range files {
			for _, decl := range file.Decls {
				d, ok := decl.(*ast.GenDecl)
				if !ok || d.Tok != tok {
					continue
				}
				if !yield(file, d) {
					return
				}
			}
		}
	}
}

func IterateValueSpecs(files []*ast.File) func(func(*ast.File, *ast.ValueSpec) bool) {
	return func(yield func(*ast.File, *ast.ValueSpec) bool) {
		for file, decl := range IterateGenDecl(files, token.VAR) {
			for _, s := range decl.Specs {
				if !yield(file, s.(*ast.ValueSpec)) {
					return
				}
			}
		}
	}
}

//func IterateTypes(files []*ast.File) func(func(*ast.File, *ast.TypeSpec) bool) {
//	return func(yield func(*ast.File, *ast.TypeSpec) bool) {
//		for _, file := range files {
//			for _, decl := range file.Decls {
//				spec, ok := decl.(*ast.GenDecl)
//				if !ok || spec.Tok != token.TYPE {
//					continue
//				}
//				for _, s := range spec.Specs {
//					t, ok := s.(*ast.TypeSpec)
//					if !ok {
//						continue
//					}
//					if !yield(file, t) {
//						return
//					}
//				}
//			}
//		}
//	}
//}

func IterateFunctions(files []*ast.File) func(func(*ast.File, *ast.FuncDecl) bool) {
	return func(yield func(*ast.File, *ast.FuncDecl) bool) {
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if !yield(file, fn) {
					return
				}
			}
		}
	}
}

//func IterateImports(files []*ast.File) func(func(*ast.File, *ast.ImportSpec) bool) {
//	return func(yield func(*ast.File, *ast.ImportSpec) bool) {
//		for _, file := range files {
//			for _, decl := range file.Decls {
//				genDecl, ok := decl.(*ast.GenDecl)
//				if !ok || genDecl.Tok != token.IMPORT {
//					continue
//				}
//				for _, s := range genDecl.Specs {
//					if !yield(file, s.(*ast.ImportSpec)) {
//						return
//					}
//				}
//			}
//		}
//	}
//}

func Format(node ast.Node) string {
	var buf strings.Builder
	if err := printer.Fprint(&buf, token.NewFileSet(), node); err != nil {
		return fmt.Sprintf("formatting error: %v", err)
	}
	return buf.String()
}

func evaluateStringLiteralExpressionList(wd string, set *token.FileSet, list []ast.Expr) ([]string, error) {
	result := make([]string, 0, len(list))
	for _, a := range list {
		s, err := evaluateStringLiteralExpression(wd, set, a)
		if err != nil {
			return result, err
		}
		result = append(result, s)
	}
	return result, nil
}

func evaluateStringLiteralExpression(wd string, set *token.FileSet, exp ast.Expr) (string, error) {
	arg, ok := exp.(*ast.BasicLit)
	if !ok || arg.Kind != token.STRING {
		return "", contextError(wd, set, exp.Pos(), fmt.Errorf("expected string literal got %s", Format(exp)))
	}
	return strconv.Unquote(arg.Value)
}

func IterateFieldTypes(list []*ast.Field) func(func(int, ast.Expr) bool) {
	return func(yield func(int, ast.Expr) bool) {
		i := 0
		for _, field := range list {
			if len(field.Names) == 0 {
				if !yield(i, field.Type) {
					return
				}
				i++
			} else {
				for range field.Names {
					if !yield(i, field.Type) {
						return
					}
					i++
				}
			}
		}
	}
}

var httpCodes = map[int]string{
	http.StatusContinue:           "StatusContinue",
	http.StatusSwitchingProtocols: "StatusSwitchingProtocols",
	http.StatusProcessing:         "StatusProcessing",
	http.StatusEarlyHints:         "StatusEarlyHints",

	http.StatusOK:                   "StatusOK",
	http.StatusCreated:              "StatusCreated",
	http.StatusAccepted:             "StatusAccepted",
	http.StatusNonAuthoritativeInfo: "StatusNonAuthoritativeInfo",
	http.StatusNoContent:            "StatusNoContent",
	http.StatusResetContent:         "StatusResetContent",
	http.StatusPartialContent:       "StatusPartialContent",
	http.StatusMultiStatus:          "StatusMultiStatus",
	http.StatusAlreadyReported:      "StatusAlreadyReported",
	http.StatusIMUsed:               "StatusIMUsed",

	http.StatusMultipleChoices:   "StatusMultipleChoices",
	http.StatusMovedPermanently:  "StatusMovedPermanently",
	http.StatusFound:             "StatusFound",
	http.StatusSeeOther:          "StatusSeeOther",
	http.StatusNotModified:       "StatusNotModified",
	http.StatusUseProxy:          "StatusUseProxy",
	http.StatusTemporaryRedirect: "StatusTemporaryRedirect",
	http.StatusPermanentRedirect: "StatusPermanentRedirect",

	http.StatusBadRequest:                   "StatusBadRequest",
	http.StatusUnauthorized:                 "StatusUnauthorized",
	http.StatusPaymentRequired:              "StatusPaymentRequired",
	http.StatusForbidden:                    "StatusForbidden",
	http.StatusNotFound:                     "StatusNotFound",
	http.StatusMethodNotAllowed:             "StatusMethodNotAllowed",
	http.StatusNotAcceptable:                "StatusNotAcceptable",
	http.StatusProxyAuthRequired:            "StatusProxyAuthRequired",
	http.StatusRequestTimeout:               "StatusRequestTimeout",
	http.StatusConflict:                     "StatusConflict",
	http.StatusGone:                         "StatusGone",
	http.StatusLengthRequired:               "StatusLengthRequired",
	http.StatusPreconditionFailed:           "StatusPreconditionFailed",
	http.StatusRequestEntityTooLarge:        "StatusRequestEntityTooLarge",
	http.StatusRequestURITooLong:            "StatusRequestURITooLong",
	http.StatusUnsupportedMediaType:         "StatusUnsupportedMediaType",
	http.StatusRequestedRangeNotSatisfiable: "StatusRequestedRangeNotSatisfiable",
	http.StatusExpectationFailed:            "StatusExpectationFailed",
	http.StatusTeapot:                       "StatusTeapot",
	http.StatusMisdirectedRequest:           "StatusMisdirectedRequest",
	http.StatusUnprocessableEntity:          "StatusUnprocessableEntity",
	http.StatusLocked:                       "StatusLocked",
	http.StatusFailedDependency:             "StatusFailedDependency",
	http.StatusTooEarly:                     "StatusTooEarly",
	http.StatusUpgradeRequired:              "StatusUpgradeRequired",
	http.StatusPreconditionRequired:         "StatusPreconditionRequired",
	http.StatusTooManyRequests:              "StatusTooManyRequests",
	http.StatusRequestHeaderFieldsTooLarge:  "StatusRequestHeaderFieldsTooLarge",
	http.StatusUnavailableForLegalReasons:   "StatusUnavailableForLegalReasons",

	http.StatusInternalServerError:           "StatusInternalServerError",
	http.StatusNotImplemented:                "StatusNotImplemented",
	http.StatusBadGateway:                    "StatusBadGateway",
	http.StatusServiceUnavailable:            "StatusServiceUnavailable",
	http.StatusGatewayTimeout:                "StatusGatewayTimeout",
	http.StatusHTTPVersionNotSupported:       "StatusHTTPVersionNotSupported",
	http.StatusVariantAlsoNegotiates:         "StatusVariantAlsoNegotiates",
	http.StatusInsufficientStorage:           "StatusInsufficientStorage",
	http.StatusLoopDetected:                  "StatusLoopDetected",
	http.StatusNotExtended:                   "StatusNotExtended",
	http.StatusNetworkAuthenticationRequired: "StatusNetworkAuthenticationRequired",
}

func HTTPStatusName(name string) (int, error) {
	n := strings.TrimPrefix(name, "http.")
	for code, constName := range httpCodes {
		if constName == n {
			return code, nil
		}
	}
	return 0, fmt.Errorf("unknown %s", name)
}

func HTTPStatusCode(imports *Imports, n int) ast.Expr {
	ident, ok := httpCodes[n]
	if !ok {
		return &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(n)}
	}
	return &ast.SelectorExpr{
		X:   ast.NewIdent(imports.Add("", "net/http")),
		Sel: ast.NewIdent(ident),
	}
}

func Ident(n string) *ast.Ident { return &ast.Ident{Name: n} }

func Int(n int) *ast.BasicLit { return &ast.BasicLit{Value: strconv.Itoa(n), Kind: token.INT} }

func String(s string) *ast.BasicLit {
	return &ast.BasicLit{Value: strconv.Quote(s), Kind: token.STRING}
}

func Nil() *ast.Ident { return ast.NewIdent("nil") }

func ErrorCheckReturn(errVarIdent string, body ...ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{X: ast.NewIdent(errVarIdent), Op: token.NEQ, Y: Nil()},
		Body: &ast.BlockStmt{List: append(body, &ast.ReturnStmt{})},
	}
}

func FieldIndex(fields []*ast.Field, i int) (*ast.Ident, ast.Expr, bool) {
	n := 0
	for _, field := range fields {
		if len(field.Names) == 0 {
			if n != i {
				n++
				continue
			}
			return nil, field.Type, true
		}
		for _, name := range field.Names {
			if n != i {
				n++
				continue
			}
			return name, field.Type, true
		}
	}
	return nil, nil, false
}

func CallError(errIdent string) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(errIdent),
			Sel: ast.NewIdent("Error"),
		},
		Args: []ast.Expr{},
	}
}

func FindFieldWithName(list *ast.FieldList, name string) (*ast.Field, bool) {
	for _, field := range list.List {
		for _, ident := range field.Names {
			if ident.Name == name {
				return field, true
			}
		}
	}
	return nil, false
}

func HasFieldWithName(list *ast.FieldList, name string) bool {
	_, ok := FindFieldWithName(list, name)
	return ok
}

func StaticTypeMethods(files []*ast.File, typeName string) *ast.FieldList {
	methods := new(ast.FieldList)
	for _, funcDecl := range IterateFunctions(files) {
		if !isMethodForType(typeName, funcDecl) {
			continue
		}
		methods.List = append(methods.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(funcDecl.Name.Name)},
			Type:  funcDecl.Type,
		})
	}
	return methods
}

func isMethodForType(receiverTypeIdent string, funcDecl *ast.FuncDecl) bool {
	if funcDecl == nil || funcDecl.Name == nil || funcDecl.Recv == nil || len(funcDecl.Recv.List) < 1 {
		return false
	}
	exp := funcDecl.Recv.List[0].Type
	if star, ok := exp.(*ast.StarExpr); ok {
		exp = star.X
	}
	ident, ok := exp.(*ast.Ident)
	return ok && ident.Name == receiverTypeIdent
}
