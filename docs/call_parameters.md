# Method Parameter Field Sets

This is the (wip) "argument" documentation.

## The Method Call Scope

There are three parameters you can pass to a method that always generate the same code

- `ctx` -> `http.Request.Context`
- `request` -> `*http.Request`
- `response` -> `http.ResponseWriter` (if you use this, muxt won't generate code to call WriteHeader, you have to do
  this)

Using these three, the generated code will look something like this.

Given `{{define "GET / F(ctx, response, request)"}}Hello{{end}}`,

You will get a handler generated like this:

```go
package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	F(ctx context.Context, response http.ResponseWriter, request *http.Request) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		data := receiver.F(ctx, response, request)
		execute(response, request, false, "GET / F(ctx, response, request)", http.StatusOK, data)
	})
}

func execute(http.ResponseWriter, *http.Request, bool, string, int, any) {}

```

You can also map path values from the path pattern to identifiers and pass them to your handler.

Given `{{define "GET /articles/{id} ReadArticle(ctx, id)"}}{{end}}`,

You will get a handler generated like this:

```go
package main

import (
	"context"
	"net/http"
)

type RoutesReceiver interface {
	ReadArticle(ctx context.Context, id string) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
	mux.HandleFunc("GET /articles/{id}", func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		id := request.PathValue("id")
		data := receiver.ReadArticle(ctx, id)
		execute(response, request, true, "GET /articles/{id} ReadArticle(ctx, id)", http.StatusOK, data)
	})
}

func execute(http.ResponseWriter, *http.Request, bool, string, int, any) {}
```

## Parsing

Many basic Go types are supported.

Integer variants are most common.

If a type implements [`encoding.TextUmarshaler`](https://pkg.go.dev/encoding#TextUnmarshaler) we will use that.