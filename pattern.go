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
	"strings"

	"github.com/crhntr/muxt/internal/source"
)

func TemplatePatterns(ts *template.Template) ([]Pattern, error) {
	var patterns []Pattern
	routes := make(map[string]struct{})
	for _, t := range ts.Templates() {
		pat, err, ok := NewPattern(t.Name())
		if !ok {
			continue
		}
		if err != nil {
			return patterns, err
		}
		if _, exists := routes[pat.Method+pat.Path]; exists {
			return patterns, fmt.Errorf("duplicate route pattern: %s", pat.Route)
		}
		routes[pat.Method+pat.Path] = struct{}{}
		patterns = append(patterns, pat)
	}
	slices.SortFunc(patterns, Pattern.byPathThenMethod)
	return patterns, nil
}

type Pattern struct {
	name                      string
	Method, Host, Path, Route string
	Handler                   string

	fun  *ast.Ident
	call *ast.CallExpr
	args []*ast.Ident
}

func NewPattern(in string) (Pattern, error, bool) {
	if !templateNameMux.MatchString(in) {
		return Pattern{}, nil, false
	}
	matches := templateNameMux.FindStringSubmatch(in)
	p := Pattern{
		name:    in,
		Method:  matches[templateNameMux.SubexpIndex("Method")],
		Host:    matches[templateNameMux.SubexpIndex("Host")],
		Path:    matches[templateNameMux.SubexpIndex("Path")],
		Handler: strings.TrimSpace(matches[templateNameMux.SubexpIndex("Handler")]),
		Route:   matches[templateNameMux.SubexpIndex("Route")],
	}

	switch p.Method {
	default:
		return p, fmt.Errorf("%s method not allowed", p.Method), true
	case "", http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	}

	return p, parseHandler(&p), true
}

var (
	pathSegmentPattern = regexp.MustCompile(`/\{([^}]*)}`)
	templateNameMux    = regexp.MustCompile(`^(?P<Route>(((?P<Method>[A-Z]+)\s+)?)(?P<Host>([^/])*)(?P<Path>(/(\S)*)))(?P<Handler>.*)$`)
)

func (def Pattern) PathParameters() ([]string, error) {
	var result []string
	for _, matches := range pathSegmentPattern.FindAllStringSubmatch(def.Path, strings.Count(def.Path, "/")) {
		n := matches[1]
		if n == "$" {
			continue
		}
		n = strings.TrimSuffix(n, "...")
		if !token.IsIdentifier(n) {
			return nil, fmt.Errorf("path parameter name not permitted: %q is not a Go identifier", n)
		}
		result = append(result, n)
	}
	for i, n := range result {
		if slices.Contains(result[:i], n) {
			return nil, fmt.Errorf("path parameter %s defined at least twice", n)
		}
	}
	return result, nil
}

func (def Pattern) CallExpr() *ast.CallExpr { return def.call }
func (def Pattern) ArgIdents() []*ast.Ident { return def.args }
func (def Pattern) FunIdent() *ast.Ident    { return def.fun }

func (def Pattern) String() string {
	return def.name
}

func (def Pattern) byPathThenMethod(d Pattern) int {
	if n := cmp.Compare(def.Path, d.Path); n != 0 {
		return n
	}
	if m := cmp.Compare(def.Method, d.Method); m != 0 {
		return m
	}
	return cmp.Compare(def.Handler, d.Handler)
}

func (def Pattern) sameRoute(p Pattern) bool { return def.Route == p.Route }

func parseHandler(def *Pattern) error {
	if def.Handler == "" {
		return nil
	}
	e, err := parser.ParseExpr(def.Handler)
	if err != nil {
		return fmt.Errorf("failed to parse handler expression: %v", err)
	}
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("expected call, got: %s", source.Format(e))
	}
	fun, ok := call.Fun.(*ast.Ident)
	if !ok {
		return fmt.Errorf("expected function identifier, got got: %s", source.Format(call.Fun))
	}
	if call.Ellipsis != token.NoPos {
		return fmt.Errorf("unexpected ellipsis")
	}
	pathParams, err := def.PathParameters()
	if err != nil {
		return err
	}
	args := make([]*ast.Ident, len(call.Args))
	for i, a := range call.Args {
		arg, ok := a.(*ast.Ident)
		if !ok {
			return fmt.Errorf("expected only argument expressions as arguments, argument at index %d is: %s", i, source.Format(a))
		}
		switch name := arg.Name; name {
		case PatternScopeIdentifierHTTPRequest,
			PatternScopeIdentifierHTTPResponse,
			PatternScopeIdentifierContext,
			PatternScopeIdentifierTemplate,
			PatternScopeIdentifierLogger:
			if slices.Contains(pathParams, name) {
				return fmt.Errorf("the name %s is not allowed as a path paramenter it is alredy in scope", name)
			}
		default:
			if !slices.Contains(pathParams, name) {
				return fmt.Errorf("unknown argument %s at index %d", name, i)
			}
		}
		args[i] = arg
	}
	def.fun = fun
	def.call = call
	def.args = args
	return nil
}

const (
	PatternScopeIdentifierHTTPRequest  = "request"
	PatternScopeIdentifierHTTPResponse = "response"
	PatternScopeIdentifierContext      = "ctx"
	PatternScopeIdentifierTemplate     = "template"
	PatternScopeIdentifierLogger       = "logger"
)
