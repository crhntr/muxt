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

	pathValueNames []string
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

	names, err := p.parsePathValueNames()
	if err != nil {
		return Pattern{}, err, true
	}
	p.pathValueNames = names
	if err := checkPathValueNames(p.pathValueNames); err != nil {
		return Pattern{}, err, true
	}

	return p, parseHandler(&p), true
}

var (
	pathSegmentPattern = regexp.MustCompile(`/\{([^}]*)}`)
	templateNameMux    = regexp.MustCompile(`^(?P<Route>(((?P<Method>[A-Z]+)\s+)?)(?P<Host>([^/])*)(?P<Path>(/(\S)*)))(?P<Handler>.*)$`)
)

func (def Pattern) parsePathValueNames() ([]string, error) {
	var result []string
	for _, match := range pathSegmentPattern.FindAllStringSubmatch(def.Path, strings.Count(def.Path, "/")) {
		n := match[1]
		if n == "$" && strings.Count(def.Path, "$") == 1 && strings.HasSuffix(def.Path, "{$}") {
			continue
		}
		n = strings.TrimSuffix(n, "...")
		result = append(result, n)
	}
	return result, nil
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
			return fmt.Errorf("the name %s is not allowed as a path paramenter it is alredy in scope", n)
		}
	}
	return nil
}

func (def Pattern) String() string           { return def.name }
func (def Pattern) PathValueNames() []string { return def.pathValueNames }
func (def Pattern) CallExpr() *ast.CallExpr  { return def.call }
func (def Pattern) FunIdent() *ast.Ident     { return def.fun }
func (def Pattern) sameRoute(p Pattern) bool { return def.Route == p.Route }

func (def Pattern) byPathThenMethod(d Pattern) int {
	if n := cmp.Compare(def.Path, d.Path); n != 0 {
		return n
	}
	if m := cmp.Compare(def.Method, d.Method); m != 0 {
		return m
	}
	return cmp.Compare(def.Handler, d.Handler)
}

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
	args := make([]*ast.Ident, len(call.Args))
	scope := append(patternScope(), def.pathValueNames...)
	slices.Sort(scope)
	for i, a := range call.Args {
		arg, ok := a.(*ast.Ident)
		if !ok {
			return fmt.Errorf("expected only argument expressions as arguments, argument at index %d is: %s", i, source.Format(a))
		}
		if _, ok := slices.BinarySearch(scope, arg.Name); !ok {
			return fmt.Errorf("unknown argument %s at index %d", arg.Name, i)
		}
		args[i] = arg
	}
	def.fun = fun
	def.call = call
	return nil
}

const (
	PatternScopeIdentifierHTTPRequest  = "request"
	PatternScopeIdentifierHTTPResponse = "response"
	PatternScopeIdentifierContext      = "ctx"
	PatternScopeIdentifierTemplate     = "template"
	PatternScopeIdentifierLogger       = "logger"
)

func patternScope() []string {
	return []string{
		PatternScopeIdentifierHTTPRequest,
		PatternScopeIdentifierHTTPResponse,
		PatternScopeIdentifierContext,
		PatternScopeIdentifierTemplate,
		PatternScopeIdentifierLogger,
	}
}
