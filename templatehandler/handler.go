package templatehandler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

func Routes(mux *http.ServeMux, ts *template.Template, logger *slog.Logger, services map[string]any) error {
	for name := range services {
		if !token.IsIdentifier(name) {
			return fmt.Errorf("service name not permitted: %q is not a Go identifier", name)
		}
	}
	for _, t := range ts.Templates() {
		pattern, err, match := NewEndpointDefinition(t.Name())
		if !match {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to parse NewPattern for template %q: %w", t.Name(), err)
		}

		if pattern.Handler == "" {
			mux.HandleFunc(pattern.Pattern, simpleTemplateHandler(t, logger))
			continue
		}
		ex, err := parser.ParseExpr(pattern.Handler)
		if err != nil {
			return fmt.Errorf("failed to parse handler expression: %w", err)
		}
		switch exp := ex.(type) {
		case *ast.CallExpr:
			h, err := callMethodHandler(t, logger, services, pattern, exp)
			if err != nil {
				return fmt.Errorf("failed to create handler for %q: %w", pattern.Pattern, err)
			}
			mux.HandleFunc(pattern.Pattern, h)
		default:
			return fmt.Errorf("unexpected handler expression %v", pattern.Handler)
		}
	}
	return nil
}

var pathSegmentPattern = regexp.MustCompile(`/\{([^}]*)}`)

func callMethodHandler(t *template.Template, logger *slog.Logger, services map[string]any, pattern EndpointDefinition, call *ast.CallExpr) (http.HandlerFunc, error) {
	scope, _, pathParams, err := generateScope(pattern, services)
	if err != nil {
		return nil, err
	}
	switch function := call.Fun.(type) {
	default:
		return nil, fmt.Errorf("expected method call on some service %s known service names are %v", pattern.Handler, keys(services))
	case *ast.SelectorExpr:
		return createSelectorHandler(t, call, function, services, scope, pathParams, logger)
	}
}

func generateScope(pattern EndpointDefinition, services map[string]any) ([]string, []string, []string, error) {
	scope := append([]string{"ctx", "request"})
	var serviceNames []string
	for n := range services {
		if slices.Contains(scope, n) {
			return nil, nil, nil, fmt.Errorf("identifier already declared %q", n)
		}
		scope = append(scope, n)
		serviceNames = append(serviceNames, n)
	}
	var pathParams []string
	for _, matches := range pathSegmentPattern.FindAllStringSubmatch(pattern.Path, strings.Count(pattern.Path, "/")) {
		n := matches[1]
		n = strings.TrimSuffix(n, "...")
		if slices.Contains(scope, n) {
			return nil, nil, nil, fmt.Errorf("identifier already declared %q", n)
		}
		if !token.IsIdentifier(n) {
			return nil, nil, nil, fmt.Errorf("path parameter name not permitted: %q is not a Go identifier", n)
		}
		scope = append(scope, n)
		pathParams = append(pathParams, n)
	}
	return scope, serviceNames, pathParams, nil
}

func createSelectorHandler(t *template.Template, call *ast.CallExpr, function *ast.SelectorExpr, services map[string]any, scope, pathParams []string, logger *slog.Logger) (http.HandlerFunc, error) {
	method, err := serviceMethod(call, function, services, scope)
	if err != nil {
		return nil, err
	}
	inputs, err := generateInputsFunction(method.Type(), call, logger, services, pathParams)
	if err != nil {
		return nil, err
	}
	return generateOutputsFunction(t, logger, method, inputs)
}

type inputsFunc = func(res http.ResponseWriter, req *http.Request) []reflect.Value

func generateOutputsFunction(t *template.Template, logger *slog.Logger, method reflect.Value, inputs inputsFunc) (http.HandlerFunc, error) {
	switch num := method.Type().NumOut(); num {
	case 1:
		return valueResultHandler(t, method, inputs, logger), nil
	case 2:
		return valuesResultHandler(t, method, inputs, logger), nil
	default:
		return nil, fmt.Errorf("method must either return (T) or (T, error)")
	}
}

func valueResultHandler(t *template.Template, method reflect.Value, inputs inputsFunc, logger *slog.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		in := inputs(res, req)
		out := method.Call(in)
		execute(res, req, t, logger, out[0].Interface())
	}
}

func valuesResultHandler(t *template.Template, method reflect.Value, inputs inputsFunc, logger *slog.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		in := inputs(res, req)
		out := method.Call(in)
		callRes, callErr := out[0], out[1]
		if !callErr.IsNil() {
			err := callErr.Interface().(error)
			logger.Error("service call failed", "method", req.Method, "path", req.URL.Path, "error", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		execute(res, req, t, logger, callRes.Interface())
	}
}

func serviceMethod(call *ast.CallExpr, function *ast.SelectorExpr, services map[string]any, scope []string) (reflect.Value, error) {
	if call.Ellipsis != token.NoPos {
		return reflect.Value{}, fmt.Errorf("ellipsis call not allowed")
	}
	receiver, ok := function.X.(*ast.Ident)
	if !ok {
		return reflect.Value{}, fmt.Errorf("unexpected method receiver expected one of %q but got: %s", scope, printNode(function.X))
	}
	if len(services) == 0 {
		return reflect.Value{}, fmt.Errorf("no services provided")
	}
	s, ok := services[receiver.Name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("service with identifier %q not found", receiver.Name)
	}
	if s == nil {
		return reflect.Value{}, fmt.Errorf("service %q must not be nil", receiver.Name)
	}
	service := reflect.ValueOf(s)
	method := service.MethodByName(function.Sel.Name)
	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("method %s not found on %s", function.Sel.Name, service.Type())
	}

	return method, nil
}

func printNode(node ast.Node) string {
	buf := bytes.NewBuffer(nil)
	_ = format.Node(buf, token.NewFileSet(), node)
	return buf.String()
}

func keys[K comparable, V any, M ~map[K]V](s M) []K {
	sn := make([]K, 0, len(s))
	for n := range s {
		sn = append(sn, n)
	}
	return sn
}

const (
	requestArgumentIdentifier  = "request"
	contextArgumentIdentifier  = "ctx"
	responseArgumentIdentifier = "response"
	loggerArgumentIdentifier   = "logger"
)

func generateInputsFunction(method reflect.Type, call *ast.CallExpr, logger *slog.Logger, services map[string]any, pathParams []string) (inputsFunc, error) {
	if method.NumIn() != len(call.Args) {
		return nil, fmt.Errorf("wrong number of arguments")
	}
	if len(call.Args) == 0 {
		return func(http.ResponseWriter, *http.Request) []reflect.Value {
			return nil
		}, nil
	}
	var args []string
	for i, exp := range call.Args {
		arg, err := typeCheckMethodParameters(pathParams, services, i, method.In(i), exp)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return func(res http.ResponseWriter, req *http.Request) []reflect.Value {
		var in []reflect.Value
		for _, arg := range args {
			switch arg {
			case requestArgumentIdentifier:
				in = append(in, reflect.ValueOf(req))
			case contextArgumentIdentifier:
				in = append(in, reflect.ValueOf(req.Context()))
			case responseArgumentIdentifier:
				in = append(in, reflect.ValueOf(res))
			case loggerArgumentIdentifier:
				in = append(in, reflect.ValueOf(logger))
			default:
				s, ok := services[arg]
				if ok {
					in = append(in, reflect.ValueOf(s))
					continue
				}
				if slices.Index(pathParams, arg) >= 0 {
					in = append(in, reflect.ValueOf(req.PathValue(arg)))
				}
			}
		}
		return in
	}, nil
}

func typeCheckMethodParameters(pathParams []string, services map[string]any, i int, tp reflect.Type, exp ast.Expr) (string, error) {
	arg, ok := exp.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("method arguments must be identifiers: argument %d is not an identifier got %s", i, printNode(exp))
	}
	switch an := arg.Name; an {
	case requestArgumentIdentifier:
		return an, nil
	case contextArgumentIdentifier:
		return an, nil
	case responseArgumentIdentifier:
		return an, nil
	case loggerArgumentIdentifier:
		return an, nil
	default:
		if _, found := services[arg.Name]; found {
			return an, nil
		}
		if slices.Contains(pathParams, arg.Name) {
			return an, nil
		}
		return "", fmt.Errorf("unknown argument %d %s", i, an)
	}
}

func simpleTemplateHandler(t *template.Template, logger *slog.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		execute(res, req, t, logger, req)
	}
}

func execute(res http.ResponseWriter, req *http.Request, t *template.Template, logger *slog.Logger, data any) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		logger.Error("failed to render page", "method", req.Method, "path", req.URL.Path, "error", err)
		http.Error(res, "failed to render page", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(res); err != nil {
		logger.Error("failed to write full response", "method", req.Method, "path", req.URL.Path, "error", err)
		return
	}
}

type EndpointDefinition struct {
	Method, Host, Path, Pattern string
	Handler                     string
}

var templateNameMux = regexp.MustCompile(`^(?P<Pattern>(?P<Method>([A-Z]+\s+)?)(?P<Host>([^/])*)(?P<Path>(/(\S)*)))(?P<Handler>.*)$`)

func NewEndpointDefinition(in string) (EndpointDefinition, error, bool) {
	if !templateNameMux.MatchString(in) {
		return EndpointDefinition{}, nil, false
	}
	matches := templateNameMux.FindStringSubmatch(in)
	p := EndpointDefinition{
		Method:  strings.TrimSpace(matches[templateNameMux.SubexpIndex("Method")]),
		Host:    strings.TrimSpace(matches[templateNameMux.SubexpIndex("Host")]),
		Path:    strings.TrimSpace(matches[templateNameMux.SubexpIndex("Path")]),
		Handler: strings.TrimSpace(matches[templateNameMux.SubexpIndex("Handler")]),
		Pattern: strings.TrimSpace(matches[templateNameMux.SubexpIndex("Pattern")]),
	}

	switch p.Method {
	default:
		return p, fmt.Errorf("%s method not allowed", p.Method), true
	case "", http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	}

	return p, nil, true
}
