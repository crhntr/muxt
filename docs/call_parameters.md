# Method Parameter Field Sets

This is the (wip) "argument" documentation.

## The Method Call Scope

There are three parameters you can pass to a method that always generate the same code

### Default Mapping

If you don't provide a named type with `--receiver-type`, muxt will try use the following default types for the
generated interface methods.

- `ctx` -> `http.Request.Context`
- `request` -> `*http.Request`
- `response` -> `http.ResponseWriter`
- `form` -> `url.Values`
- (named path values) -> `string` (i.e. "/some/{value}" where the identifier "value" now a variable name in the call
  scope)

The method will return `any`.
This result type does not play well with `muxt check`.
You should set a `--receiver-type`.

#### Example without Receiver Type

Using some of the above, the generated code will look something like this.

Given
`{{define "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)"}}Hello, world!{{end}}`,
then you will get this:

```go 
type RoutesReceiver interface {
  F(ctx context.Context, response http.ResponseWriter, request *http.Request, projectID string, taskID string) any
}
```

### Example with Receiver Type

Now, say you provide `--receiver-type=Server`, muxt now will generate parsers in the handler and the generated interface
will look like this

```go
package server

import (
  "context"
  "net/http"
)

type Server struct{}

type Data struct{}

func (_ Server) F(ctx context.Context, response http.ResponseWriter, request *http.Request, projectID uint32, taskID int8) Data {
	return Data{}
}

```

Given (the same as above)
`{{define "GET /project/{projectID}/task/{taskID} F(ctx, response, request, projectID, taskID)"}}Hello, world!{{end}}`,
then you will get this:

```go 
type RoutesReceiver interface {
    F(ctx context.Context, response http.ResponseWriter, request *http.Request, projectID uint32, taskID int8) Data
}
```

## Parsing

Muxt can generate form field and path parameter parsers for most basic Go types.

### Basic Kinds

- `int`
- `int64`
- `int32`
- `int16`
- `int8`
- `uint`
- `uint64`
- `uint32`
- `uint16`
- `uint8`
- `bool`
- `string` _(passed through with no parsing)_

If a type implements [`encoding.TextUmarshaler`](https://pkg.go.dev/encoding#TextUnmarshaler),
we will use that.

