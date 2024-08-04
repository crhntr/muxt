package templatehandler_test

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
	"github.com/crhntr/template/templatehandler"
)

//go:generate counterfeiter -generate
//counterfeiter:generate -o ../internal/fake/article_service.go --fake-name ArticleService . articleService

type (
	articleService interface {
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
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
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
			`{{define "GET /articles a.ListArticles(ctx)" }}<ul>{{range .}}<li data-id="{{.ID}}">{{.Title}}</li>{{end}}</ul>{{end}}`,
		))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		as := new(fake.ArticleService)
		articles := []example.Article{
			{ID: 1, Title: "Hello"},
			{ID: 2, Title: "Goodbye"},
		}
		as.ListArticlesReturns(articles, nil)
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"a": as,
		})
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

	t.Run("bad service name", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(""))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"$": nil,
		})
		require.ErrorContains(t, err, `service name not permitted: "$" is not a Go identifier`)
	})

	t.Run("unexpected method", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(`{{define "CONNECT /articles" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
		require.ErrorContains(t, err, `failed to parse endpoint for template "CONNECT /articles": CONNECT method not allowed`)
	})

	t.Run("no method", func(t *testing.T) {
		//
		ts := template.Must(template.New("simple path").Parse(`{{define "/x/y" }}{{.Method}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
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
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
		require.ErrorContains(t, err, "failed to parse handler expression")
	})

	t.Run("selector must be a service", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / func().Method(req)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
		require.ErrorContains(t, err, "unexpected method receiver")
	})

	t.Run("no services provided", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / x.Method(req)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, nil)
		require.ErrorContains(t, err, "no services provided")
	})

	t.Run("unknown service name", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / x.Method(req)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"y": nil})
		require.ErrorContains(t, err, `service with identifier "x" not found`)
	})

	t.Run("nil service", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / s.Method(req)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": nil})
		require.ErrorContains(t, err, `service "s" must not be nil`)
	})

	t.Run("method not found on basic type", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / s.Foo(req)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": 100})
		require.ErrorContains(t, err, `method Foo not found on int`)
	})

	t.Run("ellipsis not allowed", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{name} s.ToUpper(name...)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `ellipsis call not allowed`)
	})

	t.Run("duplicate path param identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /articles/{id}/comment/{id} s.GetComment(ctx, id, id)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `identifier already declared`)
	})

	t.Run("duplicate path param and service identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /articles/privacy/{s} s.SomeString(ctx, s)" }}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `identifier already declared`)
	})

	t.Run("path param is not an identifier ", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{key-id} s.SomeString(ctx, key-id)"}}KEY{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `path parameter name not permitted: "key-id" is not a Go identifier`)
	})

	t.Run("path param is not an identifier ", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /{key-id} s.SomeString(ctx, key-id)"}}KEY{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `path parameter name not permitted: "key-id" is not a Go identifier`)
	})

	t.Run("ctx as service name", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / request.ListArticles(ctx)"}}#{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"ctx": s})
		require.ErrorContains(t, err, `identifier already declared`)
	})

	t.Run("request as service name", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / request.ListArticles(ctx)"}}#{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"request": s})
		require.ErrorContains(t, err, `identifier already declared`)
	})
	t.Run("template execution fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Funcs(template.FuncMap{
			"errorNow": func() (string, error) { return "", fmt.Errorf("BANANA") },
		}).Parse(`{{define "GET / s.ListArticles(ctx)"}}{{ errorNow }}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
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
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(errorWriter{ResponseWriter: rec}, req)
		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, logBuffer.String(), "failed to write full response")
	})

	t.Run("write fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /"}}{{printf "%d" 199}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(errorWriter{ResponseWriter: rec}, req)
		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, logBuffer.String(), "failed to write full response")
	})

	t.Run("too many results", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / s.TooManyResults()" }}{{.}}{{end}}"`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, `method must either return (T) or (T, error)`)
	})

	t.Run("call fails", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-articles s.ListArticles(ctx)"}}{{len .}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.ListArticlesReturns(nil, fmt.Errorf("banana"))
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
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

	t.Run("non selector expression", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-articles panic(500)"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, "expected method call on some service")
	})
	t.Run("not a function call", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-articles <-c"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, "unexpected handler expression")
	})

	t.Run("single return", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /number-of-authors s.NumAuthors()"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
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
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /auth s.CheckAuth(request)"}}OK{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
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
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /site-owner s.GetComment(ctx, 3, 1+2)"}}OK{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{"s": s})
		require.ErrorContains(t, err, "method arguments must be identifiers: argument 1 is not an identifier got 3")
	})

	t.Run("passing value from scope to function", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /parsed-domain s.Parse(domain)"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s":      s,
			"domain": "example.com",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/parsed-domain", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		v := s.ParseArgsForCall(0)
		assert.Equal(t, "example.com", v)
	})

	t.Run("query param", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET /input/{in} s.Parse(in)"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s":      s,
			"domain": "example.com",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/input/peach", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		v := s.ParseArgsForCall(0)
		assert.Equal(t, "peach", v)
	})

	t.Run("unknown identifier", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / s.Parse(enemy)"}}@{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)
		s.NumAuthorsReturns(234)
		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s":      s,
			"domain": "example.com",
		})
		require.ErrorContains(t, err, "unknown argument 0 enemy")
	})

	t.Run("full handler func signature", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "GET / s.Handler(response, request)"}}{{.}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)

		s.HandlerStub = func(writer http.ResponseWriter, request *http.Request) template.HTML {
			writer.WriteHeader(http.StatusCreated)

			return "<strong>Progressive</strong>"
		}

		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s": s,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/input/peach", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("handler uses a logger", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "POST /stdin s.LogLines(logger)"}}{{printf "lines: %d" .}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()
		s := new(fake.ArticleService)

		s.LogLinesStub = func(logger *slog.Logger) int {
			logger.Info("some message")
			return 42
		}

		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s": s,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/stdin", strings.NewReader(""))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, logBuffer.String(), "some message")
	})

	t.Run("wrong number of arguments", func(t *testing.T) {
		ts := template.Must(template.New("simple path").Parse(`{{define "POST /stdin s.LogLines(ctx, logger)"}}{{printf "lines: %d" .}}{{end}}`))
		logBuffer := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		mux := http.NewServeMux()

		err := templatehandler.Routes(mux, ts, logger, map[string]any{
			"s": new(fake.ArticleService),
		})
		require.ErrorContains(t, err, "wrong number of arguments")
	})
}

type errorWriter struct {
	http.ResponseWriter
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("banna")
}
