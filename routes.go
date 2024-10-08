package muxt

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/crhntr/dom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/crhntr/muxt/internal/source"
)

const (
	executeIdentName = "execute"
	receiverIdent    = "receiver"

	dataVarIdent = "data"
	muxVarIdent  = "mux"

	requestPathValue         = "PathValue"
	httpRequestContextMethod = "Context"
	httpResponseWriterIdent  = "ResponseWriter"
	httpServeMuxIdent        = "ServeMux"
	httpRequestIdent         = "Request"
	httpHandleFuncIdent      = "HandleFunc"

	contextContextTypeIdent = "Context"

	defaultPackageName           = "main"
	DefaultTemplatesVariableName = "templates"
	DefaultRoutesFunctionName    = "routes"
	DefaultOutputFileName        = "template_routes.go"
	receiverInterfaceIdent       = "RoutesReceiver"

	InputAttributeNameStructTag     = "name"
	InputAttributeTemplateStructTag = "template"

	errIdent = "err"
)

func Generate(templateNames []TemplateName, ts *template.Template, packageName, templatesVariableName, routesFunctionName, receiverTypeIdent, output string, fileSet *token.FileSet, receiverPackage, templatesPackage []*ast.File, log *log.Logger) (string, error) {
	packageName = cmp.Or(packageName, defaultPackageName)
	templatesVariableName = cmp.Or(templatesVariableName, DefaultTemplatesVariableName)
	routesFunctionName = cmp.Or(routesFunctionName, DefaultRoutesFunctionName)
	file := &ast.File{
		Name: ast.NewIdent(packageName),
	}
	importsDecl := &ast.GenDecl{
		Tok: token.IMPORT,
	}
	imports := source.NewImports(importsDecl)
	routes := &ast.FuncDecl{
		Name: ast.NewIdent(routesFunctionName),
		Type: routesFuncType(imports, ast.NewIdent(receiverInterfaceIdent)),
		Body: &ast.BlockStmt{},
	}
	receiverInterface := &ast.InterfaceType{
		Methods: &ast.FieldList{},
	}
	for _, name := range templateNames {
		t := ts.Lookup(name.name)
		var method *ast.FuncType
		if name.fun != nil {
			for _, funcDecl := range source.IterateFunctions(receiverPackage) {
				if !name.matchReceiver(funcDecl, receiverTypeIdent) {
					continue
				}
				method = funcDecl.Type
				break
			}
			if method == nil {
				method = name.funcType(imports)
			}
			receiverInterface.Methods.List = append(receiverInterface.Methods.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(name.fun.Name)},
				Type:  method,
			})
		}

		handlerFunc, err := name.funcLit(imports, method, t, receiverPackage)
		if err != nil {
			return "", err
		}

		routes.Body.List = append(routes.Body.List, name.callHandleFunc(handlerFunc))
		log.Printf("%s has route for %s", routesFunctionName, name.String())
	}
	imports.SortImports()
	file.Decls = append(file.Decls, importsDecl)
	file.Decls = append(file.Decls, &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{Name: ast.NewIdent(receiverInterfaceIdent), Type: receiverInterface}},
	})
	file.Decls = append(file.Decls, routes)
	hasExecuteFunc := false
	for _, fn := range source.IterateFunctions(templatesPackage) {
		if fn.Recv == nil && fn.Name.Name == executeIdentName {
			p := fileSet.Position(fn.Pos())
			if filepath.Base(p.Filename) != output {
				hasExecuteFunc = true
			}
			break
		}
	}
	if !hasExecuteFunc {
		file.Decls = append(file.Decls, executeFuncDecl(imports, templatesVariableName))
	}
	return source.Format(file), nil
}

func (def TemplateName) callHandleFunc(handlerFuncLit *ast.FuncLit) *ast.ExprStmt {
	return &ast.ExprStmt{X: &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(muxVarIdent),
			Sel: ast.NewIdent(httpHandleFuncIdent),
		},
		Args: []ast.Expr{source.String(def.endpoint), handlerFuncLit},
	}}
}

func (def TemplateName) funcLit(imports *source.Imports, method *ast.FuncType, t *template.Template, files []*ast.File) (*ast.FuncLit, error) {
	if method == nil {
		return def.httpRequestReceiverTemplateHandlerFunc(imports, def.statusCode), nil
	}
	lit := &ast.FuncLit{
		Type: httpHandlerFuncType(imports),
		Body: &ast.BlockStmt{},
	}
	call := &ast.CallExpr{Fun: callReceiverMethod(def.fun)}
	if method.Params.NumFields() != len(def.call.Args) {
		return nil, errWrongNumberOfArguments(def, method)
	}
	var formStruct *ast.StructType
	for pi, pt := range fieldListTypes(method.Params) {
		if err := checkArgument(imports, method, pi, def.call.Args[pi], pt, files); err != nil {
			return nil, err
		}
		if s, ok := findFormStruct(pt, files); ok {
			formStruct = s
		}
	}
	writeHeader := true
	for i, a := range def.call.Args {
		arg := a.(*ast.Ident)
		switch arg.Name {
		case TemplateNameScopeIdentifierHTTPResponse:
			writeHeader = false
			fallthrough
		case TemplateNameScopeIdentifierHTTPRequest:
			call.Args = append(call.Args, ast.NewIdent(arg.Name))
			imports.AddNetHTTP()
		case TemplateNameScopeIdentifierContext:
			lit.Body.List = append(lit.Body.List, contextAssignment())
			call.Args = append(call.Args, ast.NewIdent(TemplateNameScopeIdentifierContext))
			imports.AddContext()
		case TemplateNameScopeIdentifierForm:
			_, tp, _ := source.FieldIndex(method.Params.List, i)
			lit.Body.List = append(lit.Body.List,
				&ast.ExprStmt{X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
						Sel: ast.NewIdent("ParseForm"),
					},
					Args: []ast.Expr{},
				}},
				formDeclaration(imports, arg.Name, tp))
			if formStruct != nil {
				for _, field := range formStruct.Fields.List {
					for _, name := range field.Names {
						fieldExpr := &ast.SelectorExpr{
							X:   ast.NewIdent(arg.Name),
							Sel: ast.NewIdent(name.Name),
						}

						fieldTemplate := formInputTemplate(field, t)

						errCheck := func(exp ast.Expr) ast.Stmt {
							return &ast.ExprStmt{
								X: imports.HTTPErrorCall(ast.NewIdent(httpResponseField(imports).Names[0].Name), source.CallError(errIdent), http.StatusBadRequest),
							}
						}

						const parsedVariableName = "value"
						if fieldType, ok := field.Type.(*ast.ArrayType); ok {
							inputName := formInputName(field, name)
							const valVar = "val"
							assignment := appendAssignment(token.ASSIGN, &ast.SelectorExpr{
								X:   ast.NewIdent(arg.Name),
								Sel: ast.NewIdent(name.Name),
							})
							var templateNodes []*html.Node
							if fieldTemplate != nil {
								templateNodes, _ = html.ParseFragment(strings.NewReader(fieldTemplate.Tree.Root.String()), &html.Node{
									Type:     html.ElementNode,
									DataAtom: atom.Body,
									Data:     atom.Body.String(),
								})
							}
							validations, err, ok := source.GenerateValidations(imports, ast.NewIdent(parsedVariableName), fieldType.Elt, fmt.Sprintf("[name=%q]", inputName), inputName, httpResponseField(imports).Names[0].Name, dom.NewDocumentFragment(templateNodes))
							if ok && err != nil {
								return nil, err
							}
							statements, err := source.GenerateParseValueFromStringStatements(imports, parsedVariableName, ast.NewIdent(valVar), fieldType.Elt, errCheck, validations, assignment)
							if err != nil {
								return nil, fmt.Errorf("failed to generate parse statements for form field %s: %w", name.Name, err)
							}

							forLoop := &ast.RangeStmt{
								Key:   ast.NewIdent("_"),
								Value: ast.NewIdent(valVar),
								Tok:   token.DEFINE,
								X: &ast.IndexExpr{
									X: &ast.SelectorExpr{
										X:   ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
										Sel: ast.NewIdent("Form"),
									},
									Index: source.String(inputName),
								},
								Body: &ast.BlockStmt{
									List: statements,
								},
							}

							lit.Body.List = append(lit.Body.List, forLoop)
						} else {
							assignment := singleAssignment(token.ASSIGN, fieldExpr)
							inputName := formInputName(field, name)
							str := &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
									Sel: ast.NewIdent("FormValue"),
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: strconv.Quote(inputName),
									},
								},
							}
							var templateNodes []*html.Node
							if fieldTemplate != nil {
								templateNodes, _ = html.ParseFragment(strings.NewReader(fieldTemplate.Tree.Root.String()), &html.Node{
									Type:     html.ElementNode,
									DataAtom: atom.Body,
									Data:     atom.Body.String(),
								})
							}
							validations, err, ok := source.GenerateValidations(imports, ast.NewIdent(parsedVariableName), field.Type, fmt.Sprintf("[name=%q]", inputName), inputName, httpResponseField(imports).Names[0].Name, dom.NewDocumentFragment(templateNodes))
							if ok && err != nil {
								return nil, err
							}
							statements, err := source.GenerateParseValueFromStringStatements(imports, parsedVariableName, str, field.Type, errCheck, validations, assignment)
							if err != nil {
								return nil, fmt.Errorf("failed to generate parse statements for form field %s: %w", name.Name, err)
							}

							if len(statements) > 1 {
								statements = []ast.Stmt{&ast.BlockStmt{
									List: statements,
								}}
							}

							lit.Body.List = append(lit.Body.List, statements...)
						}
					}
				}
			} else {
				imports.Add("", "net/url")
			}
			call.Args = append(call.Args, ast.NewIdent(arg.Name))
		default:
			errCheck := func(msg ast.Expr) ast.Stmt {
				return &ast.ExprStmt{
					X: imports.HTTPErrorCall(ast.NewIdent(httpResponseField(imports).Names[0].Name), msg, http.StatusBadRequest),
				}
			}
			src := &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
					Sel: ast.NewIdent(requestPathValue),
				},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(arg.Name)}},
			}
			statements, err := httpPathValueAssignment(imports, method, i, arg, src, token.DEFINE, errCheck)
			if err != nil {
				return nil, err
			}
			lit.Body.List = append(lit.Body.List, statements...)
			call.Args = append(call.Args, ast.NewIdent(arg.Name))
		}
	}

	const dataVarIdent = "data"
	if method.Results == nil || len(method.Results.List) == 0 {
		return lit, fmt.Errorf("method for endpoint %q has no results it should have one or two", def)
	} else if len(method.Results.List) > 1 {
		lit.Body.List = append(lit.Body.List,
			&ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(dataVarIdent), ast.NewIdent(errIdent)}, Tok: token.DEFINE, Rhs: []ast.Expr{call}},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: ast.NewIdent(errIdent), Op: token.NEQ, Y: source.Nil()},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{X: imports.HTTPErrorCall(ast.NewIdent(httpResponseField(imports).Names[0].Name), source.CallError(errIdent), http.StatusInternalServerError)},
						&ast.ReturnStmt{},
					},
				},
			},
		)
	} else {
		lit.Body.List = append(lit.Body.List, &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(dataVarIdent)}, Tok: token.DEFINE, Rhs: []ast.Expr{call}})
	}
	lit.Body.List = append(lit.Body.List, def.executeCall(source.HTTPStatusCode(imports, def.statusCode), ast.NewIdent(dataVarIdent), writeHeader))
	return lit, nil
}

func formInputName(field *ast.Field, name *ast.Ident) string {
	if field.Tag != nil {
		v, _ := strconv.Unquote(field.Tag.Value)
		tags := reflect.StructTag(v)
		n, hasInputTag := tags.Lookup(InputAttributeNameStructTag)
		if hasInputTag {
			return n
		}
	}
	return name.Name
}

func formInputTemplate(field *ast.Field, t *template.Template) *template.Template {
	if field.Tag != nil {
		v, _ := strconv.Unquote(field.Tag.Value)
		tags := reflect.StructTag(v)
		n, hasInputTag := tags.Lookup(InputAttributeTemplateStructTag)
		if hasInputTag {
			return t.Lookup(n)
		}
	}
	return t
}

func (def TemplateName) funcType(imports *source.Imports) *ast.FuncType {
	method := &ast.FuncType{
		Params:  &ast.FieldList{},
		Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("any")}}},
	}
	for _, a := range def.call.Args {
		arg := a.(*ast.Ident)
		switch arg.Name {
		case TemplateNameScopeIdentifierHTTPRequest:
			method.Params.List = append(method.Params.List, httpRequestField(imports))
		case TemplateNameScopeIdentifierHTTPResponse:
			method.Params.List = append(method.Params.List, httpResponseField(imports))
		case TemplateNameScopeIdentifierContext:
			method.Params.List = append(method.Params.List, contextContextField(imports))
		case TemplateNameScopeIdentifierForm:
			method.Params.List = append(method.Params.List, urlValuesField(imports, arg.Name))
		default:
			method.Params.List = append(method.Params.List, pathValueField(arg.Name))
		}
	}
	return method
}

func fieldListTypes(fieldList *ast.FieldList) func(func(int, ast.Expr) bool) {
	return func(yield func(int, ast.Expr) bool) {
		paramIndex := 0
		for _, param := range fieldList.List {
			if len(param.Names) == 0 {
				if !yield(paramIndex, param.Type) {
					return
				}
				paramIndex++
				continue
			}
			for range param.Names {
				if !yield(paramIndex, param.Type) {
					return
				}
				paramIndex++
			}
		}
	}
}

func errWrongNumberOfArguments(def TemplateName, method *ast.FuncType) error {
	return fmt.Errorf("handler %s expects %d arguments but call %s has %d", source.Format(&ast.FuncDecl{Name: ast.NewIdent(def.fun.Name), Type: method}), method.Params.NumFields(), def.handler, len(def.call.Args))
}

func checkArgument(imports *source.Imports, method *ast.FuncType, argIndex int, exp ast.Expr, argType ast.Expr, files []*ast.File) error {
	// TODO: rewrite to "cannot use 32 (untyped int constant) as string value in argument to strings.ToUpper"
	arg := exp.(*ast.Ident)
	switch arg.Name {
	case TemplateNameScopeIdentifierHTTPRequest:
		if !matchSelectorIdents(argType, imports.AddNetHTTP(), httpRequestIdent, true) {
			return fmt.Errorf("method expects type %s but %s is *%s.%s", source.Format(argType), arg.Name, imports.AddNetHTTP(), httpRequestIdent)
		}
		return nil
	case TemplateNameScopeIdentifierHTTPResponse:
		if !matchSelectorIdents(argType, imports.AddNetHTTP(), httpResponseWriterIdent, false) {
			return fmt.Errorf("method expects type %s but %s is %s.%s", source.Format(argType), arg.Name, imports.AddNetHTTP(), httpResponseWriterIdent)
		}
		return nil
	case TemplateNameScopeIdentifierContext:
		if !matchSelectorIdents(argType, imports.AddContext(), contextContextTypeIdent, false) {
			return fmt.Errorf("method expects type %s but %s is %s.%s", source.Format(argType), arg.Name, imports.AddContext(), contextContextTypeIdent)
		}
		return nil
	case TemplateNameScopeIdentifierForm:
		if matchSelectorIdents(argType, "url", "Values", false) {
			return nil
		}
		_, ok := findFormStruct(argType, files)
		if !ok {
			return fmt.Errorf("method expects form to have type url.Values or T (where T is some struct type)")
		}
		return nil
	default:
		for paramIndex, paramType := range source.IterateFieldTypes(method.Params.List) {
			if argIndex != paramIndex {
				continue
			}
			if err := compareTypes(paramType, argType); err != nil {
				return fmt.Errorf("method argument and param mismatch: %w", err)
			}
			break
		}
		return nil
	}
}

func compareTypes(expA, expB ast.Expr) error {
	if a, b, ok := matchExpressionType[*ast.Ident](expA, expB); ok && a.Name == b.Name {
		return nil
	}
	if a, b, ok := matchExpressionType[*ast.SelectorExpr](expA, expB); ok && a.Sel == b.Sel {
		if _, _, ok = matchExpressionType[*ast.Ident](a.X, b.X); ok {
			return nil
		}
	}
	return fmt.Errorf("type %s is not assignable to %s", source.Format(expA), source.Format(expB))
}

func matchExpressionType[T ast.Expr](a, b ast.Expr) (T, T, bool) {
	ax, aOk := a.(T)
	bx, bOk := b.(T)
	return ax, bx, aOk && bOk
}

func findFormStruct(argType ast.Expr, files []*ast.File) (*ast.StructType, bool) {
	if argTypeIdent, ok := argType.(*ast.Ident); ok {
		for _, file := range files {
			for _, d := range file.Decls {
				decl, ok := d.(*ast.GenDecl)
				if !ok || decl.Tok != token.TYPE {
					continue
				}
				for _, s := range decl.Specs {
					spec := s.(*ast.TypeSpec)
					structType, isStruct := spec.Type.(*ast.StructType)
					if isStruct && spec.Name.Name == argTypeIdent.Name {
						return structType, true
					}
				}
			}
		}
	}
	return nil, false
}

func matchSelectorIdents(expr ast.Expr, pkg, name string, star bool) bool {
	if star {
		st, ok := expr.(*ast.StarExpr)
		if !ok {
			return false
		}
		expr = st.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && sel.Sel.Name == name && id.Name == pkg
}

func pathValueField(name string) *ast.Field {
	return &ast.Field{
		Type:  ast.NewIdent("string"),
		Names: []*ast.Ident{ast.NewIdent(name)},
	}
}

func contextContextField(imports *source.Imports) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(TemplateNameScopeIdentifierContext)},
		Type:  contextContextType(imports.AddContext()),
	}
}

func httpResponseField(imports *source.Imports) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(TemplateNameScopeIdentifierHTTPResponse)},
		Type:  &ast.SelectorExpr{X: ast.NewIdent(imports.AddNetHTTP()), Sel: ast.NewIdent(httpResponseWriterIdent)},
	}
}

func routesFuncType(imports *source.Imports, receiverType ast.Expr) *ast.FuncType {
	return &ast.FuncType{Params: &ast.FieldList{
		List: []*ast.Field{
			{Names: []*ast.Ident{ast.NewIdent(muxVarIdent)}, Type: &ast.StarExpr{
				X: &ast.SelectorExpr{X: ast.NewIdent(imports.AddNetHTTP()), Sel: ast.NewIdent(httpServeMuxIdent)},
			}},
			{Names: []*ast.Ident{ast.NewIdent(receiverIdent)}, Type: receiverType},
		},
	}}
}

func urlValuesField(imports *source.Imports, ident string) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(ident)},
		Type:  &ast.SelectorExpr{X: ast.NewIdent(imports.Add("", "net/url")), Sel: ast.NewIdent("Values")},
	}
}

func httpRequestField(imports *source.Imports) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest)},
		Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: ast.NewIdent(imports.AddNetHTTP()), Sel: ast.NewIdent(httpRequestIdent)}},
	}
}

func httpHandlerFuncType(imports *source.Imports) *ast.FuncType {
	return &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{httpResponseField(imports), httpRequestField(imports)}}}
}

func callReceiverMethod(fun *ast.Ident) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: ast.NewIdent(receiverIdent), Sel: ast.NewIdent(fun.Name)}
}

func contextContextType(contextPackageIdent string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: ast.NewIdent(contextPackageIdent), Sel: ast.NewIdent(contextContextTypeIdent)}
}

func contextAssignment() *ast.AssignStmt {
	return &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{ast.NewIdent(TemplateNameScopeIdentifierContext)},
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
				Sel: ast.NewIdent(httpRequestContextMethod),
			},
		}},
	}
}

func formDeclaration(imports *source.Imports, ident string, typeExp ast.Expr) *ast.DeclStmt {
	if matchSelectorIdents(typeExp, imports.Ident("net/url"), "Values", false) {
		imports.Add("", "net/url")
		return &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(ident)},
						Type:  typeExp,
						Values: []ast.Expr{
							&ast.SelectorExpr{X: ast.NewIdent(httpResponseField(imports).Names[0].Name), Sel: ast.NewIdent("Form")},
						},
					},
				},
			},
		}
	}
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent(ident)},
					Type:  typeExp,
				},
			},
		},
	}
}

func httpPathValueAssignment(imports *source.Imports, method *ast.FuncType, i int, arg *ast.Ident, str ast.Expr, assignTok token.Token, errCheck func(stmt ast.Expr) ast.Stmt) ([]ast.Stmt, error) {
	for typeIndex, typeExp := range source.IterateFieldTypes(method.Params.List) {
		if typeIndex != i {
			continue
		}
		assignment := singleAssignment(assignTok, ast.NewIdent(arg.Name))
		return source.GenerateParseValueFromStringStatements(imports, arg.Name+"Parsed", str, typeExp, errCheck, nil, assignment)
	}
	return nil, fmt.Errorf("type for argumement %d not found", i)
}

func singleAssignment(assignTok token.Token, result ast.Expr) func(exp ast.Expr) ast.Stmt {
	return func(exp ast.Expr) ast.Stmt {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{result},
			Tok: assignTok,
			Rhs: []ast.Expr{exp},
		}
	}
}

func appendAssignment(assignTok token.Token, result ast.Expr) func(exp ast.Expr) ast.Stmt {
	return func(exp ast.Expr) ast.Stmt {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{result},
			Tok: assignTok,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun:  ast.NewIdent("append"),
				Args: []ast.Expr{result, exp},
			}},
		}
	}
}

func (def TemplateName) executeCall(status, data ast.Expr, writeHeader bool) *ast.ExprStmt {
	return &ast.ExprStmt{X: &ast.CallExpr{
		Fun: ast.NewIdent(executeIdentName),
		Args: []ast.Expr{
			ast.NewIdent(TemplateNameScopeIdentifierHTTPResponse),
			ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest),
			ast.NewIdent(strconv.FormatBool(writeHeader)),
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(def.name)},
			status,
			data,
		},
	}}
}

func (def TemplateName) httpRequestReceiverTemplateHandlerFunc(imports *source.Imports, statusCode int) *ast.FuncLit {
	return &ast.FuncLit{
		Type: httpHandlerFuncType(imports),
		Body: &ast.BlockStmt{List: []ast.Stmt{def.executeCall(source.HTTPStatusCode(imports, statusCode), ast.NewIdent(TemplateNameScopeIdentifierHTTPRequest), true)}},
	}
}

func (def TemplateName) matchReceiver(funcDecl *ast.FuncDecl, receiverTypeIdent string) bool {
	if funcDecl == nil || funcDecl.Name == nil || funcDecl.Name.Name != def.fun.Name ||
		funcDecl.Recv == nil || len(funcDecl.Recv.List) < 1 {
		return false
	}
	exp := funcDecl.Recv.List[0].Type
	if star, ok := exp.(*ast.StarExpr); ok {
		exp = star.X
	}
	ident, ok := exp.(*ast.Ident)
	return ok && ident.Name == receiverTypeIdent
}

func executeFuncDecl(imports *source.Imports, templatesVariableIdent string) *ast.FuncDecl {
	const writeHeaderIdent = "writeHeader"
	return &ast.FuncDecl{
		Name: ast.NewIdent(executeIdentName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					httpResponseField(imports),
					httpRequestField(imports),
					{Names: []*ast.Ident{ast.NewIdent(writeHeaderIdent)}, Type: ast.NewIdent("bool")},
					{Names: []*ast.Ident{ast.NewIdent("name")}, Type: ast.NewIdent("string")},
					{Names: []*ast.Ident{ast.NewIdent("code")}, Type: ast.NewIdent("int")},
					{Names: []*ast.Ident{ast.NewIdent(dataVarIdent)}, Type: ast.NewIdent("any")},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("buf")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent(imports.Add("", "bytes")),
							Sel: ast.NewIdent("NewBuffer"),
						},
						Args: []ast.Expr{source.Nil()},
					}},
				},
				&ast.IfStmt{
					Init: &ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent(errIdent)},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent(templatesVariableIdent),
								Sel: ast.NewIdent("ExecuteTemplate"),
							},
							Args: []ast.Expr{ast.NewIdent("buf"), ast.NewIdent("name"), ast.NewIdent(dataVarIdent)},
						}},
					},
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent(errIdent),
						Op: token.NEQ,
						Y:  source.Nil(),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{X: imports.HTTPErrorCall(ast.NewIdent(httpResponseField(imports).Names[0].Name), source.CallError(errIdent), http.StatusInternalServerError)},
							&ast.ReturnStmt{},
						},
					},
				},
				&ast.IfStmt{
					Cond: ast.NewIdent(writeHeaderIdent),
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.ExprStmt{X: &ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent(TemplateNameScopeIdentifierHTTPResponse), Sel: ast.NewIdent("Header")}, Args: []ast.Expr{}}, Sel: ast.NewIdent("Set")},
							Args: []ast.Expr{source.String("content-type"), source.String("text/html; charset=utf-8")},
						}},
						&ast.ExprStmt{X: &ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: ast.NewIdent(httpResponseField(imports).Names[0].Name), Sel: ast.NewIdent("WriteHeader")},
							Args: []ast.Expr{ast.NewIdent("code")},
						}},
					}}},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("_"), ast.NewIdent("_")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("buf"),
							Sel: ast.NewIdent("WriteTo"),
						},
						Args: []ast.Expr{ast.NewIdent(httpResponseField(imports).Names[0].Name)},
					}},
				},
			},
		},
	}
}
