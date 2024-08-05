package muxt_test

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/crhntr/dom/domtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html/atom"

	"github.com/crhntr/template/internal/example"
	"github.com/crhntr/template/internal/fake"
	"github.com/crhntr/template/muxt"
)

//go:generate counterfeiter -generate
//counterfeiter:generate -o ../internal/fake/receiver.go --fake-name Receiver . receiver

var _ receiver = (*fake.Receiver)(nil)

type (
	receiver interface {
		ListArticles(ctx context.Context) ([]example.Article, error)
		ToUpper(in ...rune) string
		Parse(string) []string
		GetComment(ctx context.Context, articleID, commentID int) (string, error)
		SomeString(ctx context.Context, x string) (string, error)
		TooManyResults() (int, int, int)
		NumAuthors() int
		CheckAuth(req *http.Request) (string, error)
		Handler(http.ResponseWriter, *http.Request) template.HTML
		LogLines(*slog.Logger) int
	}
)

func TestRoutes(t *testing.T) {
	t.Run("GET index", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(
			/* language=gotemplate */
			`{{define "GET /" }}<h1>Hello, friend!</h1>{{end}}`,
		))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("when a handler is registered", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(
			/* language=gotemplate */
			`{{define "GET /articles ListArticles(ctx)" }}<ul>{{range .}}<li data-id="{{.ID}}">{{.Title}}</li>{{end}}</ul>{{end}}`,
		))
		as := new(fake.Receiver)
		articles := []example.Article{
			{ID: 1, Title: "Hello"},
			{ID: 2, Title: "Goodbye"},
		}
		as.ListArticlesReturns(articles, nil)
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(as))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, 1, as.ListArticlesCallCount())

		res := rec.Result()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		fragment := domtest.DocumentFragmentResponse(t, res, atom.Body)
		listItems := fragment.QuerySelectorAll(`[data-id]`)
		assert.Equal(t, len(articles), listItems.Length())
		for i := 0; i < listItems.Length(); i++ {
			li := listItems.Item(i)
			assert.Equal(t, articles[i].Title, li.TextContent())
			assert.Equal(t, strconv.Itoa(articles[i].ID), li.GetAttribute("data-id"))
		}
	})

	t.Run("unexpected method", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(`{{define "CONNECT /articles" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts)
		require.ErrorContains(t, err, `failed to parse NewPattern for template "CONNECT /articles": CONNECT method not allowed`)
	})

	t.Run("no method", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(`{{define "/x/y" }}{{.Method}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts)
		require.NoError(t, err)

		for _, method := range []string{http.MethodGet, http.MethodPost} {
			req := httptest.NewRequest(method, "/x/y", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)
			res := rec.Result()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			body, _ := io.ReadAll(res.Body)
			assert.Equal(t, method, string(body))
		}
	})

	t.Run("selector must be an expression", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / var x int" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(new(fake.Receiver)))
		require.ErrorContains(t, err, "failed to parse handler expression")
	})

	t.Run("function must be an identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / func().Method(req)" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		rec := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(rec))
		require.ErrorContains(t, err, `expected method call on receiver`)
	})

	t.Run("receiver is nil and a method is expected", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / Method(req)" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts)
		require.ErrorContains(t, err, "receiver is nil")
	})

	t.Run("method not found on basic type", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / Foo(req)" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(100))
		require.ErrorContains(t, err, `method Foo not found on int`)
	})

	t.Run("ellipsis not allowed", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{name} ToUpper(name...)" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, `ellipsis call not allowed`)
	})

	t.Run("duplicate path param identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /articles/{id}/comment/{id} GetComment(ctx, id, id)" }}{{.}}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, `identifier already declared`)
	})

	t.Run("path param is not an identifier ", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{key-id} SomeString(ctx, key-id)"}}KEY{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, `path parameter name not permitted: "key-id" is not a Go identifier`)
	})

	t.Run("path param is not an identifier ", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{key-id} SomeString(ctx, key-id)"}}KEY{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(new(fake.Receiver)))
		require.ErrorContains(t, err, `path parameter name not permitted: "key-id" is not a Go identifier`)
	})

	t.Run("template execution fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Funcs(template.FuncMap{
			"errorNow": func() (string, error) { return "", fmt.Errorf("BANANA") },
		}).Parse(`{{define "GET / ListArticles(ctx)"}}{{ errorNow }}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		res := rec.Result()
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})

	t.Run("write fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /"}}{{printf "%d" 199}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithStructuredLogger(logger))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(errorWriter{ResponseWriter: rec}, req)
		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, logBuffer.String(), "failed to write full response")
	})

	t.Run("too many results", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / TooManyResults()" }}{{.}}{{end}}"`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(new(fake.Receiver)))
		require.ErrorContains(t, err, `method must either return (T) or (T, error)`)
	})

	t.Run("call fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-articles ListArticles(ctx)"}}{{len .}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.ListArticlesReturns(nil, fmt.Errorf("banana"))
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s).WithStructuredLogger(logger))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/number-of-articles", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		res := rec.Result()
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		body, _ := io.ReadAll(res.Body)
		assert.NotContains(t, string(body), "banana")

		assert.Contains(t, logBuffer.String(), "service call failed")
	})

	t.Run("not a function call", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-articles <-c"}}{{.}}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, "unexpected handler expression")
	})

	t.Run("single return", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-authors NumAuthors()"}}{{.}}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.NumAuthorsReturns(234)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/number-of-authors", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, _ := io.ReadAll(res.Body)

		assert.Equal(t, string(body), "234")
	})

	t.Run("request as a parameter", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /auth CheckAuth(request)"}}OK{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.NumAuthorsReturns(234)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/auth", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, _ := io.ReadAll(res.Body)

		assert.Equal(t, string(body), "OK")
	})

	t.Run("non identifier params", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /site-owner GetComment(ctx, 3, 1+2)"}}OK{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.NumAuthorsReturns(234)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, `method argument at index 1: argument is not an identifier got 3`)
	})

	t.Run("query param", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /input/{in} Parse(in)"}}{{.}}{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.NumAuthorsReturns(234)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/input/peach", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		v := s.ParseArgsForCall(0)
		assert.Equal(t, "peach", v)
	})

	t.Run("unknown identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / Parse(enemy)"}}@{{end}}`))
		mux := http.NewServeMux()
		s := new(fake.Receiver)
		s.NumAuthorsReturns(234)
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s))
		require.ErrorContains(t, err, `method argument at index 0: unknown argument type for enemy`)
	})

	t.Run("full handler func signature", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / Handler(response, request)"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.Receiver)

		s.HandlerStub = func(writer http.ResponseWriter, request *http.Request) template.HTML {
			writer.WriteHeader(http.StatusCreated)

			return "<strong>Progressive</strong>"
		}

		err := muxt.Handlers(mux, ts, muxt.WithReceiver(s).WithStructuredLogger(logger))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/input/peach", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("handler uses a logger", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "POST /stdin LogLines(logger)"}}{{printf "lines: %d" .}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.Receiver)

		s.LogLinesStub = func(logger *slog.Logger) int {
			logger.Info("some message")
			return 42
		}

		err := muxt.Handlers(mux, ts, muxt.WithStructuredLogger(logger).WithReceiver(s))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/stdin", strings.NewReader(""))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, logBuffer.String(), "some message")
	})

	t.Run("wrong number of arguments", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "POST /stdin LogLines(ctx, logger)"}}{{printf "lines: %d" .}}{{end}}`))
		mux := http.NewServeMux()
		err := muxt.Handlers(mux, ts, muxt.WithReceiver(new(fake.Receiver)))
		require.ErrorContains(t, err, "wrong number of arguments")
	})
}

type errorWriter struct {
	http.ResponseWriter
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("banna")
}

func Test_endpoint(t *testing.T) {
	for _, tt := range []struct {
		Name         string
		TemplateName string
		ExpMatch     bool
		Pattern      func(t *testing.T, pat muxt.EndpointDefinition)
		Error        func(t *testing.T, err error)
	}{
		{
			Name:         "get root",
			TemplateName: "GET /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodGet,
					Host:    "",
					Path:    "/",
					Pattern: "GET /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "multiple spaces after method",
			TemplateName: "GET  /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodGet,
					Host:    "",
					Path:    "/",
					Pattern: "GET  /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "post root",
			TemplateName: "POST /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodPost,
					Host:    "",
					Path:    "/",
					Pattern: "POST /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "patch root",
			TemplateName: "PATCH /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodPatch,
					Host:    "",
					Path:    "/",
					Pattern: "PATCH /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "delete root",
			TemplateName: "DELETE /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodDelete,
					Host:    "",
					Path:    "/",
					Pattern: "DELETE /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "put root",
			TemplateName: "PUT /",
			ExpMatch:     true,
			Pattern: func(t *testing.T, pat muxt.EndpointDefinition) {
				assert.Equal(t, muxt.EndpointDefinition{
					Method:  http.MethodPut,
					Host:    "",
					Path:    "/",
					Pattern: "PUT /",
					Handler: "",
				}, pat)
			},
		},
		{
			Name:         "put root",
			TemplateName: "OPTIONS /",
			ExpMatch:     true,
			Error: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "OPTIONS method not allowed")
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			pat, err, match := muxt.NewEndpointDefinition(tt.TemplateName)
			require.Equal(t, tt.ExpMatch, match)
			if tt.Error != nil {
				tt.Error(t, err)
			} else if tt.Pattern != nil {
				assert.NoError(t, err)
				tt.Pattern(t, pat)
			}
		})
	}
}
