package muxt_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"

	"github.com/crhntr/muxt"
)

func TestGenerate(t *testing.T) {
	for _, tt := range []struct {
		Name            string
		Templates       string
		Receiver        string
		ReceiverPackage string
		PackageName     string
		TemplatesVar    string
		RoutesFunc      string
		Interface       string
		Imports         []string

		ExpectedError string
		ExpectedFile  string
	}{
		{
			Name:      "simple",
			Templates: `{{define "GET /"}}Hello, world!{{end}}`,
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type RoutesReceiver interface {
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		execute(response, request, true, "GET /", http.StatusOK, request)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "simple call",
			Templates: `{{define "GET / F()"}}Hello, world!{{end}}`,
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type RoutesReceiver interface {
	F() any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		data := receiver.F()
		execute(response, request, true, "GET / F()", http.StatusOK, data)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "multiple arguments no static receiver",
			Templates: `{{define "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)"}}Hello, world!{{end}}`,
			ExpectedFile: `package main

import (
	"bytes"
	"context"
	"net/http"
)

type RoutesReceiver interface {
	F(ctx context.Context, response http.ResponseWriter, request *http.Request, projectID string, taskID string) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /project/{projectID}/task/{taskID}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		projectID := request.PathValue("projectID")
		taskID := request.PathValue("taskID")
		data := receiver.F(ctx, response, request, projectID, taskID)
		execute(response, request, false, "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)", http.StatusOK, data)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "simple call with static one argument",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F(username string) int { return 30 }
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type RoutesReceiver interface {
	F(username string) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		data := receiver.F(username)
		execute(response, request, true, "GET /age/{username} F(username)", http.StatusOK, data)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "when the default interface name is overwritten",
			Templates: `{{define "GET / F()"}}Hello{{end}}`,
			Receiver:  "T",
			Interface: "Server",
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type Server interface {
	F() any
}

func routes(mux *http.ServeMux, receiver Server) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		data := receiver.F()
		execute(response, request, true, "GET / F()", http.StatusOK, data)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "F returns a value and a boolean",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F(username string) (int, bool) { return 30, true }
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type RoutesReceiver interface {
	F(username string) (int, bool)
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		data, ok := receiver.F(username)
		if !ok {
			return
		}
		execute(response, request, true, "GET /age/{username} F(username)", http.StatusOK, data)
	})
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "F returns a value and an unsupported type",
			Templates: `{{define "GET /{$} F()"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F() (int, float64) { return 30, true }
`,
			Receiver:      "T",
			ExpectedError: "expected last result to be either an error or a bool",
		},
		{
			Name:      "F returns a value and an unsupported type",
			Templates: `{{define "GET /{$} F()"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F() (int, []error) { return 30, nil }
`,
			Receiver:      "T",
			ExpectedError: "expected last result to be either an error or a bool",
		},
		{
			Name:      "method receiver is a pointer",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `s
-- receiver.go --
package main

type T struct{}

func (*T) F(username string) int { return 30 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(username string) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		data := receiver.F(username)
		execute(response, request, true, "GET /age/{username} F(username)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "execute function defined in receiver file",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `s
-- receiver.go --
package main

type T struct{}

func (*T) F(username string) int { return 30 }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	response.WriteHeader(code)
	_ = templates.ExecuteTemplate(response, name, data)
}
`,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(username string) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		data := receiver.F(username)
		execute(response, request, true, "GET /age/{username} F(username)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "execute function already defined in output file",
			Templates: ``,
			ReceiverPackage: `
-- receiver.go --
package main

-- template_routes.go --
package main

import(
	"html/template"
	"net/http"
)

func routes(mux *http.ServeMux, receiver RoutesReceiver) {}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	response.WriteHeader(code)
	_ = templates.ExecuteTemplate(response, name, data)
}
`,
			ExpectedFile: `package main

import (
	"bytes"
	"net/http"
)

type RoutesReceiver interface {
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
}
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.Header().Set("content-type", "text/html; charset=utf-8")
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "call method with two returns",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

import "net/http"

type T struct{}

func (T) F(username string) (int, error) { return 30, error }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(username string) (int, error)
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		data, err := receiver.F(username)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		execute(response, request, true, "GET /age/{username} F(username)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "error wrong argument type",
			Templates: `{{define "GET / F(request)"}}Hello, world!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F(ctx context.Context) int { return 30 }

` + executeGo,
			Receiver:      "T",
			ExpectedError: "method expects type context.Context but request is *http.Request",
		},
		{
			Name:      "simple call larger receiver with larger package",
			Templates: `{{define "GET /age/{username} F(ctx, username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type (
	T0 struct{}

	T struct{}
)

-- f.go --
package main

import "context"

func F(string) int { return 20 }

func (T0) F(ctx context.Context) int { return 30 }

func (T) F1(ctx context.Context, username string) int { return 30 }

func (T) F(ctx context.Context, username string) int { return 30 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	F(ctx context.Context, username string) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /age/{username}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		username := request.PathValue("username")
		data := receiver.F(ctx, username)
		execute(response, request, true, "GET /age/{username} F(ctx, username)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:     "when using param parsers",
			Receiver: "T",
			Templates: `
{{- define "GET /bool/{value}   PassBool(value)"   -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /int/{value}    PassInt(value)"    -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /int16/{value}  PassInt16(value)"  -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /int32/{value}  PassInt32(value)"  -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /int64/{value}  PassInt64(value)"  -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /int8/{value}   PassInt8(value)"   -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /uint/{value}   PassUint(value)"   -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /uint16/{value} PassUint16(value)" -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /uint32/{value} PassUint32(value)" -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /uint64/{value} PassUint64(value)" -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
{{- define "GET /uint8/{value}  PassUint8(value)"  -}} <p>{{- printf "%[1]#v %[1]T" . -}}</p> {{- end -}}
`,
			ReceiverPackage: `
-- t.go --
package main

import (
	"embed"
)

//go:embed *.gohtml
var formHTML embed.FS

var templates = template.Must(template.ParseFS(formHTML, "*"))

type T struct{}

func (T) PassInt(in int) int          { return in }
func (T) PassInt64(in int64) int64    { return in }
func (T) PassInt32(in int32) int32    { return in }
func (T) PassInt16(in int16) int16    { return in }
func (T) PassInt8(in int8) int8       { return in }
func (T) PassUint(in uint) uint       { return in }
func (T) PassUint64(in uint64) uint64 { return in }
func (T) PassUint16(in uint16) uint16 { return in }
func (T) PassUint32(in uint32) uint32 { return in }
func (T) PassUint64(in uint16) uint16 { return in }
func (T) PassUint8(in uint8) uint8    { return in }
func (T) PassBool(in bool) bool       { return in }
func (T) PassByte(in byte) byte       { return in }
func (T) PassRune(in rune) rune       { return in }
` + executeGo,
			ExpectedFile: `package main

import (
	"net/http"
	"strconv"
)

type RoutesReceiver interface {
	PassBool(in bool) bool
	PassInt(in int) int
	PassInt16(in int16) int16
	PassInt32(in int32) int32
	PassInt64(in int64) int64
	PassInt8(in int8) int8
	PassUint(in uint) uint
	PassUint16(in uint16) uint16
	PassUint32(in uint32) uint32
	PassUint64(in uint64) uint64
	PassUint8(in uint8) uint8
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /bool/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseBool(request.PathValue("value"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := valueParsed
		data := receiver.PassBool(value)
		execute(response, request, true, "GET /bool/{value}   PassBool(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.Atoi(request.PathValue("value"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := valueParsed
		data := receiver.PassInt(value)
		execute(response, request, true, "GET /int/{value}    PassInt(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int16/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseInt(request.PathValue("value"), 10, 16)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := int16(valueParsed)
		data := receiver.PassInt16(value)
		execute(response, request, true, "GET /int16/{value}  PassInt16(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int32/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseInt(request.PathValue("value"), 10, 32)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := int32(valueParsed)
		data := receiver.PassInt32(value)
		execute(response, request, true, "GET /int32/{value}  PassInt32(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int64/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseInt(request.PathValue("value"), 10, 64)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := valueParsed
		data := receiver.PassInt64(value)
		execute(response, request, true, "GET /int64/{value}  PassInt64(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int8/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseInt(request.PathValue("value"), 10, 8)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := int8(valueParsed)
		data := receiver.PassInt8(value)
		execute(response, request, true, "GET /int8/{value}   PassInt8(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /uint/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 0)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := uint(valueParsed)
		data := receiver.PassUint(value)
		execute(response, request, true, "GET /uint/{value}   PassUint(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /uint16/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 16)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := uint16(valueParsed)
		data := receiver.PassUint16(value)
		execute(response, request, true, "GET /uint16/{value} PassUint16(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /uint32/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 32)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := uint32(valueParsed)
		data := receiver.PassUint32(value)
		execute(response, request, true, "GET /uint32/{value} PassUint32(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /uint64/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 64)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := valueParsed
		data := receiver.PassUint64(value)
		execute(response, request, true, "GET /uint64/{value} PassUint64(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /uint8/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 8)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := uint8(valueParsed)
		data := receiver.PassUint8(value)
		execute(response, request, true, "GET /uint8/{value}  PassUint8(value)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "form has no fields",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

type T struct{}

type In struct{}

func (T) F(form In) any { return nil }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(form In) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is not defined and a form field is passed",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

type T struct{}

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"net/url"
)

type RoutesReceiver interface {
	F(form url.Values) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form url.Values = response.Form
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form type is a struct",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

type (
	T struct{}
	In struct{
		field string
	}
)

func (T) F(form In) int { return 0 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		form.field = request.FormValue("field")
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "form html has a cromulent min attribute",
			Templates: `{{define "GET / F(form)"}}<input type="number" name="field" min="13">{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

type (
	T struct{}
	In struct{
		field int
	}
)

func (T) F(form In) int { return 0 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"strconv"
)

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		{
			value, err := strconv.Atoi(request.FormValue("field"))
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			if value < 13 {
				http.Error(response, "field must not be less than 13", http.StatusBadRequest)
				return
			}
			form.field = value
		}
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form field has an input tag",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

type (
	T struct{}
	In struct{
		field string ` + "`name:\"some-field\"`" + `
	}
)

func (T) F(form In) int { return 0 }
` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		form.field = request.FormValue("some-field")
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form has two string fields",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

import "net/http"

type (
	T struct{}
	In struct{
		fieldInt    int
		fieldInt64  int64
		fieldInt32  int32
		fieldInt16  int16
		fieldInt8   int8
		fieldUint   uint
		fieldUint64 uint64
		fieldUint16 uint16
		fieldUint32 uint32
		fieldUint16 uint16
		fieldUint8  uint8
		fieldBool   bool
		fieldTime   time.Time
	}
)

func (T) F(form In) int { return 0 }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"strconv"
	"time"
)

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		{
			value, err := strconv.Atoi(request.FormValue("fieldInt"))
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt = value
		}
		{
			value, err := strconv.ParseInt(request.FormValue("fieldInt64"), 10, 64)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt64 = value
		}
		{
			value, err := strconv.ParseInt(request.FormValue("fieldInt32"), 10, 32)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt32 = int32(value)
		}
		{
			value, err := strconv.ParseInt(request.FormValue("fieldInt16"), 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt16 = int16(value)
		}
		{
			value, err := strconv.ParseInt(request.FormValue("fieldInt8"), 10, 8)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt8 = int8(value)
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint"), 10, 0)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint = uint(value)
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint64"), 10, 64)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint64 = value
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint16"), 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint16 = uint16(value)
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint32"), 10, 32)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint32 = uint32(value)
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint16"), 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint16 = uint16(value)
		}
		{
			value, err := strconv.ParseUint(request.FormValue("fieldUint8"), 10, 8)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint8 = uint8(value)
		}
		{
			value, err := strconv.ParseBool(request.FormValue("fieldBool"))
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldBool = value
		}
		{
			value, err := time.Parse("2006-01-02", request.FormValue("fieldTime"))
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldTime = value
		}
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form has two two names for a single field",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

import "net/http"

type (
	T struct{}
	In struct{
		field1, field2 string
	}
)

func (T) F(form In) int { return 0 }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		form.field1 = request.FormValue("field1")
		form.field2 = request.FormValue("field2")
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form slice field",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

import "net/http"

type (
	T struct{}
	In struct{
		field []string
	}
)

func (T) F(form In) int { return 0 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		for _, val := range request.Form["field"] {
			form.field = append(form.field, val)
		}
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form has typed slice fields",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

import "time"

type (
	T struct{}
	In struct{
		fieldInt    []int
		fieldInt64  []int64
		fieldInt32  []int32
		fieldInt16  []int16
		fieldInt8   []int8
		fieldUint   []uint
		fieldUint64 []uint64
		fieldUint16 []uint16
		fieldUint32 []uint32
		fieldUint16 []uint16
		fieldUint8  []uint8
		fieldBool   []bool
		fieldTime   []time.Time
	}
)

func (T) F(form In) int { return 0 }

` + executeGo,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"strconv"
	"time"
)

type RoutesReceiver interface {
	F(form In) int
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form In
		for _, val := range request.Form["fieldInt"] {
			value, err := strconv.Atoi(val)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt = append(form.fieldInt, value)
		}
		for _, val := range request.Form["fieldInt64"] {
			value, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt64 = append(form.fieldInt64, value)
		}
		for _, val := range request.Form["fieldInt32"] {
			value, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt32 = append(form.fieldInt32, int32(value))
		}
		for _, val := range request.Form["fieldInt16"] {
			value, err := strconv.ParseInt(val, 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt16 = append(form.fieldInt16, int16(value))
		}
		for _, val := range request.Form["fieldInt8"] {
			value, err := strconv.ParseInt(val, 10, 8)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldInt8 = append(form.fieldInt8, int8(value))
		}
		for _, val := range request.Form["fieldUint"] {
			value, err := strconv.ParseUint(val, 10, 0)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint = append(form.fieldUint, uint(value))
		}
		for _, val := range request.Form["fieldUint64"] {
			value, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint64 = append(form.fieldUint64, value)
		}
		for _, val := range request.Form["fieldUint16"] {
			value, err := strconv.ParseUint(val, 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint16 = append(form.fieldUint16, uint16(value))
		}
		for _, val := range request.Form["fieldUint32"] {
			value, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint32 = append(form.fieldUint32, uint32(value))
		}
		for _, val := range request.Form["fieldUint16"] {
			value, err := strconv.ParseUint(val, 10, 16)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint16 = append(form.fieldUint16, uint16(value))
		}
		for _, val := range request.Form["fieldUint8"] {
			value, err := strconv.ParseUint(val, 10, 8)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldUint8 = append(form.fieldUint8, uint8(value))
		}
		for _, val := range request.Form["fieldBool"] {
			value, err := strconv.ParseBool(val)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldBool = append(form.fieldBool, value)
		}
		for _, val := range request.Form["fieldTime"] {
			value, err := time.Parse("2006-01-02", val)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			form.fieldTime = append(form.fieldTime, value)
		}
		data := receiver.F(form)
		execute(response, request, true, "GET / F(form)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "F is defined and form has unsupported field type",
			Templates: `{{define "GET / F(form)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- in.go --
package main

import (
	"net/http"
	"net/url"
)

type (
	T struct{}
	In struct{
		href url.URL
	}
)

func (T) F(form In) int { return 0 }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			Receiver:      "T",
			ExpectedError: "failed to generate parse statements for form field href: unsupported type: url.URL",
		},
		{
			Name:        "call F",
			Templates:   `{{define "GET / F()"}}Hello, world!{{end}}`,
			Receiver:    "T",
			PackageName: "main",
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F() any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		data := receiver.F()
		execute(response, request, true, "GET / F()", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:        "no handler",
			Templates:   `{{define "GET /"}}Hello, world!{{end}}`,
			Receiver:    "T",
			PackageName: "main",
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		execute(response, request, true, "GET /", http.StatusOK, request)
	})
}
`,
		},
		{
			Name:        "no handler",
			Templates:   `{{define "GET /"}}Hello, world!{{end}}`,
			Receiver:    "T",
			PackageName: "main",
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		execute(response, request, true, "GET /", http.StatusOK, request)
	})
}
`,
		},
		{
			Name:      "call F with argument response",
			Templates: `{{define "GET / F(response)"}}{{end}}`,
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func (T) F(http.ResponseWriter) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(response http.ResponseWriter) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		data := receiver.F(response)
		execute(response, request, false, "GET / F(response)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call F with argument context",
			Templates: `{{define "GET / F(ctx)"}}{{end}}`,
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func (T) F(ctx context.Context) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	F(ctx context.Context) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		data := receiver.F(ctx)
		execute(response, request, true, "GET / F(ctx)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call F with argument path param",
			Templates: `{{define "GET /{param} F(param)"}}{{end}}`,
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func (T) F(param string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import "net/http"

type RoutesReceiver interface {
	F(param string) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /{param}", func(response http.ResponseWriter, request *http.Request) {
		param := request.PathValue("param")
		data := receiver.F(param)
		execute(response, request, true, "GET /{param} F(param)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call F with multiple arguments",
			Templates: `{{define "GET /{userName} F(ctx, userName)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import "context"

type T struct{}

func (T) F(ctx context.Context, userName string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	F(ctx context.Context, userName string) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /{userName}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userName := request.PathValue("userName")
		data := receiver.F(ctx, userName)
		execute(response, request, true, "GET /{userName} F(ctx, userName)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "missing arguments",
			Templates: `{{define "GET / F()"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

type T struct{}

func (T) F(string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,

			ExpectedError: "handler func F(string) any expects 1 arguments but call F() has 0",
		},
		{
			Name:      "extra arguments",
			Templates: `{{define "GET /{name} F(ctx, name)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import ( 
	"context"
	"net/http"
)

type T struct{}

func (T) F(context.Context) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "handler func F(context.Context) any expects 1 arguments but call F(ctx, name) has 2",
		},
		{
			Name:      "wrong argument type request",
			Templates: `{{define "GET / F(request)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import ( 
	"context"
	"net/http"
)

type T struct{}

func (T) F(string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method expects type string but request is *http.Request",
		},
		{
			Name:      "wrong argument type ctx",
			Templates: `{{define "GET / F(ctx)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import "net/http"

type T struct{}

func (T) F(string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method expects type string but ctx is context.Context",
		},
		{
			Name:      "wrong argument type response",
			Templates: `{{define "GET / F(response)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import "net/http"

type T struct{}

func (T) F(string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method expects type string but response is http.ResponseWriter",
		},
		{
			Name:      "method missing a result",
			Templates: `{{define "GET / F()"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- t.go --
package main

type T struct{}

func (T) F() {}

` + executeGo,
			ExpectedError: `method for endpoint "GET / F()" has no results it should have one or two`,
		},
		{
			Name:      "wrong argument type path value",
			Templates: `{{define "GET /{name} F(name)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import "net/http"

type T struct{}

func (T) F(float64) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method param type float64 not supported",
		},
		{
			Name:      "wrong argument type request ptr",
			Templates: `{{define "GET / F(request)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import "net/http"

type T struct{}

func (T) F(*T) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method expects type *T but request is *http.Request",
		},
		{
			Name:      "wrong argument type in field list",
			Templates: `{{define "GET /post/{postID}/comment/{commentID} F(ctx, request, commentID)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import (
	"context"
	"net/http"
)

type T struct{}

func (T) F(context.Context, string, string) any {return nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedError: "method expects type string but request is *http.Request",
		},
		{
			Name:      "call expression argument with bool last result",
			Templates: `{{define "GET /{id} F(ctx, Session(response, request), id)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import (
	"context"
	"net/http"
)

type (
	T struct{}
	S struct{}
)

func (T) F(context.Context, S, int) any {return nil}

func (T) Session(http.ResponseWriter, *http.Request) (S, bool) {return Session{}, false}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
	"strconv"
)

type RoutesReceiver interface {
	Session(http.ResponseWriter, *http.Request) (S, bool)
	F(context.Context, S, int) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /{id}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		result0, ok := receiver.Session(response, request)
		if !ok {
			return
		}
		idParsed, err := strconv.Atoi(request.PathValue("id"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		id := idParsed
		data := receiver.F(ctx, result0, id)
		execute(response, request, true, "GET /{id} F(ctx, Session(response, request), id)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call expression argument with error last result",
			Templates: `{{define "GET /{id} F(ctx, Author(id), id)"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import (
	"context"
	"net/http"
)

type (
	T struct{}
	User struct{}
)

func (T) F(context.Context, S, int) any {return nil}

func (T) Author(int) (User, error) {return Session{}, nil}

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
	"strconv"
)

type RoutesReceiver interface {
	Author(int) (User, error)
	F(context.Context, S, int) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /{id}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		idParsed, err := strconv.Atoi(request.PathValue("id"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		id := idParsed
		result0, err := receiver.Author(id)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		data := receiver.F(ctx, result0, id)
		execute(response, request, true, "GET /{id} F(ctx, Author(id), id)", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call expression argument",
			Templates: `{{define "GET / F(ctx, LoadConfiguration())"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import (
	"context"
	"net/http"
)

type (
	T struct{}
	Configuration struct{}
)

func (T) F(context.Context, Configuration) any {return nil}

func (T) LoadConfiguration() (_ Configuration) { return }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	LoadConfiguration() (_ Configuration)
	F(context.Context, Configuration) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		result0 := receiver.LoadConfiguration()
		data := receiver.F(ctx, result0)
		execute(response, request, true, "GET / F(ctx, LoadConfiguration())", http.StatusOK, data)
	})
}
`,
		},
		{
			Name:      "call expression argument",
			Templates: `{{define "GET / F(ctx, Headers(response))"}}{{end}}`,
			Receiver:  "T",
			ReceiverPackage: `-- in.go --
package main

import (
	"context"
	"net/http"
)

type (
	T struct{}
	Configuration struct{}
)

func (T) F(context.Context, any) any {return nil}

func (T) Headers(response http.ResponseWriter) any { return }

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`,
			ExpectedFile: `package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	Headers(response http.ResponseWriter) any
	F(context.Context, any) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		result0 := receiver.Headers(response)
		data := receiver.F(ctx, result0)
		execute(response, request, false, "GET / F(ctx, Headers(response))", http.StatusOK, data)
	})
}
`,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			ts := template.Must(template.New(tt.Name).Parse(tt.Templates))
			templateNames, err := muxt.Templates(ts)
			require.NoError(t, err)
			logs := log.New(io.Discard, "", 0)
			set := token.NewFileSet()
			goFiles := methodFuncTypeLoader(t, set, tt.ReceiverPackage)
			out, err := muxt.Generate(templateNames, tt.PackageName, tt.TemplatesVar, tt.RoutesFunc, tt.Receiver, tt.Interface, muxt.DefaultOutputFileName, set, goFiles, logs)
			if tt.ExpectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedFile, out)
			} else {
				assert.ErrorContains(t, err, tt.ExpectedError)
			}
		})
	}
}

func methodFuncTypeLoader(t *testing.T, set *token.FileSet, in string) []*ast.File {
	t.Helper()
	archive := txtar.Parse([]byte(in))
	var files []*ast.File
	for _, file := range archive.Files {
		f, err := parser.ParseFile(set, file.Name, file.Data, parser.AllErrors)
		require.NoError(t, err)
		files = append(files, f)
	}
	return files
}

const executeGo = `-- execute.go --
package main

import "net/http"

func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
`
