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
		Imports         []string
		Method          *ast.FuncType

		ExpectedError string
		ExpectedFile  string
	}{
		{
			Name:      "simple",
			Templates: `{{define "GET /"}}Hello, world!{{end}}`,
			ExpectedFile: `package main

import (
	"net/http"
	"bytes"
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
	"net/http"
	"bytes"
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
	"context"
	"net/http"
	"bytes"
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
	"net/http"
	"bytes"
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
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "method receiver is a pointer",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `s
-- receiver.go --
package main

type T struct{}

func (*T) F(username string) int { return 30 }
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"bytes"
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
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
		{
			Name:      "execute function defined",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `s
-- receiver.go --
package main

type T struct{}

func (*T) F(username string) int { return 30 }

func execute(response http.ResponseWriter, request *http.Request, t *template.Template, code int, data any) {
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
			Name:      "call method with two returns",
			Templates: `{{define "GET /age/{username} F(username)"}}Hello, {{.}}!{{end}}`,
			ReceiverPackage: `
-- receiver.go --
package main

type T struct{}

func (T) F(username string) (int, error) { return 30, error }
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"net/http"
	"bytes"
)

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
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
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
`,
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
`,
			Receiver: "T",
			ExpectedFile: `package main

import (
	"context"
	"net/http"
	"bytes"
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
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
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
`,
			ExpectedFile: `package main

import (
	"net/http"
	"strconv"
	"bytes"
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
		value, err := strconv.ParseBool(request.PathValue("value"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		data := receiver.PassBool(value)
		execute(response, request, true, "GET /bool/{value}   PassBool(value)", http.StatusOK, data)
	})
	mux.HandleFunc("GET /int/{value}", func(response http.ResponseWriter, request *http.Request) {
		valueParsed, err := strconv.ParseInt(request.PathValue("value"), 10, 64)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		value := int(valueParsed)
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
		value, err := strconv.ParseInt(request.PathValue("value"), 10, 64)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
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
		valueParsed, err := strconv.ParseUint(request.PathValue("value"), 10, 64)
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
		value, err := strconv.ParseUint(request.PathValue("value"), 10, 64)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
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
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {
	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if writeHeader {
		response.WriteHeader(code)
	}
	_, _ = buf.WriteTo(response)
}
`,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			ts := template.Must(template.New(tt.Name).Parse(tt.Templates))
			templateNames, err := muxt.TemplateNames(ts)
			require.NoError(t, err)
			logs := log.New(io.Discard, "", 0)
			set := token.NewFileSet()
			goFiles := methodFuncTypeLoader(t, set, tt.ReceiverPackage)
			out, err := muxt.Generate(templateNames, tt.PackageName, tt.TemplatesVar, tt.RoutesFunc, tt.Receiver, set, goFiles, goFiles, logs)
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
