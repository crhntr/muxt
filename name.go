package muxt

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/crhntr/muxt/internal/source"
)

func TemplateNames(ts *template.Template) ([]TemplateName, error) {
	var templateNames []TemplateName
	routes := make(map[string]struct{})
	for _, t := range ts.Templates() {
		templateName, err, ok := NewTemplateName(t.Name())
		if !ok {
			continue
		}
		if err != nil {
			return templateNames, err
		}
		if _, exists := routes[templateName.method+templateName.path]; exists {
			return templateNames, fmt.Errorf("duplicate route pattern: %s", templateName.endpoint)
		}
		templateName.template = t
		routes[templateName.method+templateName.path] = struct{}{}
		templateNames = append(templateNames, templateName)
	}
	slices.SortFunc(templateNames, TemplateName.byPathThenMethod)
	return templateNames, nil
}

type TemplateName struct {
	// name has the full unaltered template name
	name string

	// method, host, path, and endpoint are parsed sub-parts of the string passed to mux.Handle
	method, host, path, endpoint string

	// handler is used to generate the method interface
	handler string

	// statusCode is the status code to use in the response header
	statusCode int

	fun  *ast.Ident
	call *ast.CallExpr

	fileSet *token.FileSet

	template *template.Template
}

func NewTemplateName(in string) (TemplateName, error, bool) { return newTemplate(in) }

func newTemplate(in string) (TemplateName, error, bool) {
	if !templateNameMux.MatchString(in) {
		return TemplateName{}, nil, false
	}
	matches := templateNameMux.FindStringSubmatch(in)
	p := TemplateName{
		name:       in,
		method:     matches[templateNameMux.SubexpIndex("method")],
		host:       matches[templateNameMux.SubexpIndex("host")],
		path:       matches[templateNameMux.SubexpIndex("path")],
		handler:    strings.TrimSpace(matches[templateNameMux.SubexpIndex("handler")]),
		endpoint:   matches[templateNameMux.SubexpIndex("endpoint")],
		fileSet:    token.NewFileSet(),
		statusCode: http.StatusOK,
	}
	httpStatusCode := matches[templateNameMux.SubexpIndex("code")]
	if httpStatusCode != "" {
		if strings.HasPrefix(httpStatusCode, "http.Status") {
			code, err := source.HTTPStatusName(httpStatusCode)
			if err != nil {
				return TemplateName{}, fmt.Errorf("failed to parse status code: %w", err), true
			}
			p.statusCode = code
		} else {
			code, err := strconv.Atoi(strings.TrimSpace(httpStatusCode))
			if err != nil {
				return TemplateName{}, fmt.Errorf("failed to parse status code: %w", err), true
			}
			p.statusCode = code
		}
	}

	switch p.method {
	default:
		return p, fmt.Errorf("%s method not allowed", p.method), true
	case "", http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	}

	pathParameterNames := p.parsePathValueNames()
	if err := checkPathValueNames(pathParameterNames); err != nil {
		return TemplateName{}, err, true
	}

	err := parseHandler(p.fileSet, &p, pathParameterNames)
	if err != nil {
		return p, err, true
	}

	if httpStatusCode != "" && !p.callWriteHeader(nil) {
		return p, fmt.Errorf("you can not use %s as an argument and specify an HTTP status code", TemplateNameScopeIdentifierHTTPResponse), true
	}

	return p, nil, true
}

var (
	pathSegmentPattern = regexp.MustCompile(`/\{([^}]*)}`)
	templateNameMux    = regexp.MustCompile(`^(?P<endpoint>(((?P<method>[A-Z]+)\s+)?)(?P<host>([^/])*)(?P<path>(/(\S)*)))(\s+(?P<code>(\d|http\.Status)\S+))?(?P<handler>.*)?$`)
)

func (tn TemplateName) parsePathValueNames() []string {
	var result []string
	for _, match := range pathSegmentPattern.FindAllStringSubmatch(tn.path, strings.Count(tn.path, "/")) {
		n := match[1]
		if n == "$" && strings.Count(tn.path, "$") == 1 && strings.HasSuffix(tn.path, "{$}") {
			continue
		}
		n = strings.TrimSuffix(n, "...")
		result = append(result, n)
	}
	return result
}

func checkPathValueNames(in []string) error {
	for i, n := range in {
		if !token.IsIdentifier(n) {
			return fmt.Errorf("path parameter name not permitted: %q is not a Go identifier", n)
		}
		if slices.Contains(in[:i], n) {
			return fmt.Errorf("forbidden repeated path parameter names: found at least 2 path parameters with name %q", n)
		}
		if slices.Contains(patternScope(), n) {
			return fmt.Errorf("the name %s is not allowed as a path parameter it is already in scope", n)
		}
	}
	return nil
}

func (tn TemplateName) String() string { return tn.name }

func (tn TemplateName) byPathThenMethod(d TemplateName) int {
	if n := cmp.Compare(tn.path, d.path); n != 0 {
		return n
	}
	if m := cmp.Compare(tn.method, d.method); m != 0 {
		return m
	}
	return cmp.Compare(tn.handler, d.handler)
}

func parseHandler(fileSet *token.FileSet, def *TemplateName, pathParameterNames []string) error {
	if def.handler == "" {
		return nil
	}
	e, err := parser.ParseExprFrom(fileSet, "template_name.go", []byte(def.handler), 0)
	if err != nil {
		return fmt.Errorf("failed to parse handler expression: %v", err)
	}
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("expected call expression, got: %s", source.Format(e))
	}
	fun, ok := call.Fun.(*ast.Ident)
	if !ok {
		return fmt.Errorf("expected function identifier, got got: %s", source.Format(call.Fun))
	}
	if call.Ellipsis != token.NoPos {
		return fmt.Errorf("unexpected ellipsis")
	}

	scope := append(patternScope(), pathParameterNames...)
	slices.Sort(scope)
	if err := checkArguments(scope, call); err != nil {
		return err
	}

	def.fun = fun
	def.call = call
	return nil
}

func (tn TemplateName) callWriteHeader(receiverInterfaceType *ast.InterfaceType) bool {
	if tn.call == nil {
		return true
	}
	return !hasIdentArgument(tn.call.Args, TemplateNameScopeIdentifierHTTPResponse, receiverInterfaceType, 1, 1)
}

func hasIdentArgument(args []ast.Expr, ident string, receiverInterfaceType *ast.InterfaceType, depth, maxDepth int) bool {
	if depth > maxDepth {
		return false
	}
	for _, arg := range args {
		switch exp := arg.(type) {
		case *ast.Ident:
			if exp.Name == ident {
				return true
			}
		case *ast.CallExpr:
			methodIdent, ok := exp.Fun.(*ast.Ident)
			if ok && receiverInterfaceType != nil {
				field, ok := source.FindFieldWithName(receiverInterfaceType.Methods, methodIdent.Name)
				if ok {
					funcType, ok := field.Type.(*ast.FuncType)
					if ok {
						if funcType.Results.NumFields() == 1 {
							if hasIdentArgument(exp.Args, ident, receiverInterfaceType, depth+1, maxDepth+1) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func checkArguments(identifiers []string, call *ast.CallExpr) error {
	for i, a := range call.Args {
		switch exp := a.(type) {
		case *ast.Ident:
			if _, ok := slices.BinarySearch(identifiers, exp.Name); !ok {
				return fmt.Errorf("unknown argument %s at index %d", exp.Name, i)
			}
		case *ast.CallExpr:
			if err := checkArguments(identifiers, exp); err != nil {
				return fmt.Errorf("call %s argument error: %w", source.Format(call.Fun), err)
			}
		default:
			return fmt.Errorf("expected only identifier or call expressions as arguments, argument at index %d is: %s", i, source.Format(a))
		}
	}
	return nil
}

const (
	TemplateNameScopeIdentifierHTTPRequest  = "request"
	TemplateNameScopeIdentifierHTTPResponse = "response"
	TemplateNameScopeIdentifierContext      = "ctx"
	TemplateNameScopeIdentifierForm         = "form"
)

func patternScope() []string {
	return []string{
		TemplateNameScopeIdentifierHTTPRequest,
		TemplateNameScopeIdentifierHTTPResponse,
		TemplateNameScopeIdentifierContext,
		TemplateNameScopeIdentifierForm,
	}
}
