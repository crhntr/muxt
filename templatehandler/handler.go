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
		pattern, err, match := endpoint(t.Name())
		if !match {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to parse endpoint for template %q: %w", t.Name(), err)
		}

		if pattern.Handler == "" {
			mux.HandleFunc(pattern.Pattern, simpleTemplateHandler(ts, logger, t.Name()))
			continue
		}
		switch exp := pattern.handler.(type) {
		case *ast.CallExpr:
			h, err := callMethodHandler(ts, logger, services, pattern, t.Name(), exp)
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

func callMethodHandler(ts *template.Template, logger *slog.Logger, services map[string]any, pattern Pattern, templateName string, call *ast.CallExpr) (http.HandlerFunc, error) {
	scope := append([]string{"ctx", "request"})
	var serviceNames []string
	for n := range services {
		if slices.Contains(scope, n) {
			return nil, fmt.Errorf("identifier already declared %q", n)
		}
		scope = append(scope, n)
		serviceNames = append(serviceNames, n)
	}
	var pathParams []string
	for _, matches := range pathSegmentPattern.FindAllStringSubmatch(pattern.Path, strings.Count(pattern.Path, "/")) {
		n := matches[1]
		n = strings.TrimSuffix(n, "...")
		if slices.Contains(scope, n) {
			return nil, fmt.Errorf("identifier already declared %q", n)
		}
		if !token.IsIdentifier(n) {
			return nil, fmt.Errorf("path parameter name not permitted: %q is not a Go identifier", n)
		}
		scope = append(scope, n)
		pathParams = append(pathParams, n)
	}

	switch function := call.Fun.(type) {
	default:
		return nil, fmt.Errorf("expected method call on some service %s known service names are %v", pattern.Handler, keys(services))
	case *ast.SelectorExpr:
		receiver, ok := function.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("unexpected method receiver expected one of %q but got: %s", scope, printNode(function.X))
		}
		if len(services) == 0 {
			return nil, fmt.Errorf("no services provided")
		}
		s, ok := services[receiver.Name]
		if !ok {
			return nil, fmt.Errorf("service with identifier %q not found", receiver.Name)
		}
		if s == nil {
			return nil, fmt.Errorf("service %q must not be nil", receiver.Name)
		}
		service := reflect.ValueOf(s)
		method := service.MethodByName(function.Sel.Name)
		if !method.IsValid() {
			return nil, fmt.Errorf("method %s not found on %s", function.Sel.Name, service.Type())
		}

		if call.Ellipsis != token.NoPos {
			return nil, fmt.Errorf("ellipsis call not allowed")
		}

		inputs, err := generateInputsFunction(call, logger, services, pathParams)
		if err != nil {
			return nil, err
		}

		switch on := method.Type().NumOut(); on {
		case 1:
			return func(res http.ResponseWriter, req *http.Request) {
				in := inputs(res, req)
				out := method.Call(in)
				execute(res, req, logger, ts, templateName, out[0].Interface())
			}, nil
		case 2:
			return func(res http.ResponseWriter, req *http.Request) {
				in := inputs(res, req)
				out := method.Call(in)
				callRes, callErr := out[0], out[1]
				if !callErr.IsNil() {
					logger.Error("service call failed", "method", req.Method, "path", req.URL.Path, "error", err)
					http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				execute(res, req, logger, ts, templateName, callRes.Interface())
			}, nil
		default:
			return nil, fmt.Errorf("method must either return (T) or (T, error)")
		}
	}
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

func generateInputsFunction(call *ast.CallExpr, logger *slog.Logger, services map[string]any, pathParams []string) (func(res http.ResponseWriter, req *http.Request) []reflect.Value, error) {
	const (
		requestIdentifier  = "request"
		contextIdentifier  = "ctx"
		responseIdentifier = "response"
		loggerIdentifier   = "logger"
	)

	if len(call.Args) == 0 {
		return func(http.ResponseWriter, *http.Request) []reflect.Value {
			return nil
		}, nil
	}

	var args []string
	for i, exp := range call.Args {
		arg, ok := exp.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("method arguments must be identifiers: argument %d is not an identifier got %s", i, printNode(exp))
		}
		switch an := arg.Name; an {
		case requestIdentifier, contextIdentifier, responseIdentifier, loggerIdentifier:
			args = append(args, arg.Name)
		default:
			if _, found := services[arg.Name]; found {
				args = append(args, arg.Name)
				continue
			}
			if slices.Contains(pathParams, arg.Name) {
				args = append(args, arg.Name)
				continue
			}
			return nil, fmt.Errorf("unknown argument %d %s", i, an)
		}
	}
	return func(res http.ResponseWriter, req *http.Request) []reflect.Value {
		var in []reflect.Value
		for _, arg := range args {
			switch arg {
			case requestIdentifier:
				in = append(in, reflect.ValueOf(req))
			case contextIdentifier:
				in = append(in, reflect.ValueOf(req.Context()))
			case responseIdentifier:
				in = append(in, reflect.ValueOf(res))
			case loggerIdentifier:
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

func simpleTemplateHandler(ts *template.Template, logger *slog.Logger, name string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		execute(res, req, logger, ts, name, req)
	}
}

func execute(res http.ResponseWriter, req *http.Request, logger *slog.Logger, ts *template.Template, name string, data any) {
	var buf bytes.Buffer
	if err := ts.ExecuteTemplate(&buf, name, data); err != nil {
		logger.Error("failed to render page", "method", req.Method, "path", req.URL.Path, "error", err)
		http.Error(res, "failed to render page", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(res); err != nil {
		logger.Error("failed to write full response", "method", req.Method, "path", req.URL.Path, "error", err)
		return
	}
}

type Pattern struct {
	Method, Host, Path, Pattern string
	Handler                     string

	handler ast.Expr
}

var templateNameMux = regexp.MustCompile(`^(?P<Pattern>(?P<Method>([A-Z]+ )?)(?P<Host>([^/])*)(?P<Path>(/(\S)*)))(?P<Handler>.*)$`)

func endpoint(in string) (Pattern, error, bool) {
	if !templateNameMux.MatchString(in) {
		return Pattern{}, nil, false
	}
	matches := templateNameMux.FindStringSubmatch(in)
	p := Pattern{
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

	if p.Handler != "" {
		ex, err := parser.ParseExpr(p.Handler)
		if err != nil {
			return p, fmt.Errorf("failed to parse handler expression: %w", err), true
		}
		p.handler = ex
	}

	return p, nil, true
}
