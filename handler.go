package muxt

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type Options struct {
	logger   *slog.Logger
	receiver any
}

func newOptions() Options {
	return Options{
		logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
		receiver: nil,
	}
}

func WithStructuredLogger(log *slog.Logger) Options {
	return newOptions().WithStructuredLogger(log)
}

func WithReceiver(r any) Options {
	return newOptions().WithReceiver(r)
}

func (o Options) WithStructuredLogger(log *slog.Logger) Options {
	o.logger = log
	return o
}

func (o Options) WithReceiver(r any) Options {
	o.receiver = r
	return o
}

func applyOptions(options []Options) *Options {
	result := newOptions()
	for _, o := range options {
		if o.logger != nil {
			result.logger = o.logger
		}
		if o.receiver != nil {
			result.receiver = o.receiver
		}
	}
	return &result
}

func Handlers(mux *http.ServeMux, ts *template.Template, options ...Options) error {
	o := applyOptions(options)
	for _, t := range ts.Templates() {
		pattern, err, match := NewEndpointDefinition(t.Name())
		if !match {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to parse NewPattern for template %q: %w", t.Name(), err)
		}
		if pattern.Handler == "" {
			mux.HandleFunc(pattern.Pattern, simpleTemplateHandler(t, o.logger))
			continue
		}
		ex, err := parser.ParseExpr(pattern.Handler)
		if err != nil {
			return fmt.Errorf("failed to parse handler expression: %w", err)
		}
		switch exp := ex.(type) {
		case *ast.CallExpr:
			h, err := callMethodHandler(o, t, pattern, exp)
			if err != nil {
				return fmt.Errorf("failed to create handler for %q: %w", pattern.Pattern+" "+pattern.Handler, err)
			}
			mux.HandleFunc(pattern.Pattern, h)
		default:
			return fmt.Errorf("unexpected handler expression %v", pattern.Handler)
		}
	}
	return nil
}

func callMethodHandler(o *Options, t *template.Template, pattern EndpointDefinition, call *ast.CallExpr) (http.HandlerFunc, error) {
	pathParams, err := pattern.pathParams()
	if err != nil {
		return nil, err
	}
	switch function := call.Fun.(type) {
	default:
		return nil, fmt.Errorf("expected method call on receiver")
	case *ast.Ident:
		return createSelectorHandler(o, t, call, function, pathParams)
	}
}

func createSelectorHandler(o *Options, t *template.Template, call *ast.CallExpr, method *ast.Ident, pathParams []string) (http.HandlerFunc, error) {
	m, err := serviceMethod(o, call, method)
	if err != nil {
		return nil, err
	}
	inputs, err := generateInputsFunction(o, t, m.Type(), call, pathParams)
	if err != nil {
		return nil, err
	}
	return generateOutputsFunction(o, t, m, inputs)
}

type inputsFunc = func(res http.ResponseWriter, req *http.Request) []reflect.Value

func generateOutputsFunction(o *Options, t *template.Template, method reflect.Value, inputs inputsFunc) (http.HandlerFunc, error) {
	switch num := method.Type().NumOut(); num {
	case 1:
		return valueResultHandler(t, method, inputs, o.logger), nil
	case 2:
		return valuesResultHandler(t, method, inputs, o.logger), nil
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

func serviceMethod(o *Options, call *ast.CallExpr, method *ast.Ident) (reflect.Value, error) {
	if o.receiver == nil {
		return reflect.Value{}, fmt.Errorf("receiver is nil")
	}
	if call.Ellipsis != token.NoPos {
		return reflect.Value{}, fmt.Errorf("ellipsis call not allowed")
	}
	r := reflect.ValueOf(o.receiver)
	m := r.MethodByName(method.Name)
	if !m.IsValid() {
		return reflect.Value{}, fmt.Errorf("method %s not found on %s", method.Name, r.Type())
	}
	return m, nil
}

func printNode(node ast.Node) string {
	buf := bytes.NewBuffer(nil)
	_ = format.Node(buf, token.NewFileSet(), node)
	return buf.String()
}

const (
	// requestArgumentIdentifier identifies an *http.Request
	requestArgumentIdentifier = "request"

	// contextArgumentIdentifier identifies a context.Context off of *http.Request
	contextArgumentIdentifier = "ctx"

	// responseArgumentIdentifier identifies an http.ResponseWriter
	responseArgumentIdentifier = "response"

	// loggerArgumentIdentifier identifies an *slog.Logger
	loggerArgumentIdentifier = "logger"

	// loggerArgumentTemplate identifies a *template.Template
	loggerArgumentTemplate = "template"
)

func rootScope() []string {
	return []string{
		requestArgumentIdentifier,
		contextArgumentIdentifier,
		responseArgumentIdentifier,
		loggerArgumentIdentifier,
		loggerArgumentTemplate,
	}
}

func generateInputsFunction(o *Options, t *template.Template, method reflect.Type, call *ast.CallExpr, pathParams []string) (inputsFunc, error) {
	if method.NumIn() != len(call.Args) {
		return nil, fmt.Errorf("wrong number of arguments")
	}
	if len(call.Args) == 0 {
		return func(http.ResponseWriter, *http.Request) []reflect.Value {
			return nil
		}, nil
	}
	for _, pp := range pathParams {
		if slices.Contains(rootScope(), pp) {
			return nil, fmt.Errorf("identfier %s is already defined", pp)
		}
	}
	var args []string
	for i, exp := range call.Args {
		arg, err := typeCheckMethodParameters(pathParams, method.In(i), exp)
		if err != nil {
			return nil, fmt.Errorf("method argument at index %d: %w", i, err)
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
				in = append(in, reflect.ValueOf(o.logger))
			case loggerArgumentTemplate:
				in = append(in, reflect.ValueOf(t))
			default:
				if slices.Index(pathParams, arg) >= 0 {
					in = append(in, reflect.ValueOf(req.PathValue(arg)))
				}
			}
		}
		return in
	}, nil
}

var argumentType = sync.OnceValue(func() func(argName string, pathParams []string) (reflect.Type, error) {
	requestType := reflect.TypeFor[*http.Request]()
	contextType := reflect.TypeFor[context.Context]()
	responseType := reflect.TypeFor[http.ResponseWriter]()
	loggerType := reflect.TypeFor[*slog.Logger]()
	templateType := reflect.TypeFor[*template.Template]()
	stringType := reflect.TypeFor[string]()
	return func(argName string, pathParams []string) (reflect.Type, error) {
		var argType reflect.Type
		switch argName {
		case requestArgumentIdentifier:
			argType = requestType
		case contextArgumentIdentifier:
			argType = contextType
		case responseArgumentIdentifier:
			argType = responseType
		case loggerArgumentIdentifier:
			argType = loggerType
		case loggerArgumentTemplate:
			argType = templateType
		default:
			if !slices.Contains(pathParams, argName) {
				return nil, fmt.Errorf("unknown argument type for %s", argName)
			}
			argType = stringType
		}
		return argType, nil
	}
})

func typeCheckMethodParameters(pathParams []string, paramType reflect.Type, exp ast.Expr) (string, error) {
	arg, ok := exp.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("argument is not an identifier got %s", printNode(exp))
	}
	argType, err := argumentType()(arg.Name, pathParams)
	if err != nil {
		return arg.Name, err
	}
	if !argType.AssignableTo(paramType) {
		return arg.Name, fmt.Errorf("argument %s %s is not assignable to parameter type %s", arg.Name, argType, paramType)
	}
	return arg.Name, nil
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

var pathSegmentPattern = regexp.MustCompile(`/\{([^}]*)}`)

func (def EndpointDefinition) pathParams() ([]string, error) {
	var result []string
	for _, matches := range pathSegmentPattern.FindAllStringSubmatch(def.Path, strings.Count(def.Path, "/")) {
		n := matches[1]
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
