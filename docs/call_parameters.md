# Method Parameter Field Sets

This is the (wip) "argument" documentation.

## The Method Call Scope

There are three parameters you can pass to a method that always generate the same code

- `ctx` -> `http.Request.Context`
- `request` -> `*http.Request`
- `response` -> `http.ResponseWriter` (if you use this, muxt won't generate code to call WriteHeader, you have to do
  this)

Using these three, the generated code will look something like this.
You can also map path values from the path pattern to identifiers and pass them to your handler.

Given `{{define "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)"}}Hello, world!{{end}}`,

You will get a handler generated like this:

```go
package main

import (
  "bytes"
  "context"
  "net/http"
)

type (
  RoutesReceiver interface {
    F(ctx context.Context, response http.ResponseWriter, request *http.Request, projectID string, taskID string) any
  }
  responseData[T any] struct {
    Request *http.Request
    Data    T
  }
)

func newResponseData[T any](data T, request *http.Request) responseData[T] {
  return responseData[T]{Data: data, Request: request}
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
  mux.HandleFunc("GET /project/{projectID}/task/{taskID}", func(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()
    projectID := request.PathValue("projectID")
    taskID := request.PathValue("taskID")
    data := receiver.F(ctx, response, request, projectID, taskID)
    buf := bytes.NewBuffer(nil)
    rd := newResponseData(data, request)
    if err := templates.ExecuteTemplate(buf, "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)", rd); err != nil {
      http.Error(response, err.Error(), http.StatusInternalServerError)
      return
    }
    _, _ = buf.WriteTo(response)
  })
}
```

## Parsing

Many basic Go types are supported.

Integer variants are most common.

If a type implements [`encoding.TextUmarshaler`](https://pkg.go.dev/encoding#TextUnmarshaler) we will use that.
