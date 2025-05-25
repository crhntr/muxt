# Method Parameter Field Sets

In template names, you may add a call expression.
The names of these arguments and the parameter types for the method being called will be used to generate a handler func.

## Default Argument Types
- `ctx` -> `http.Request.Context`
- `request` -> `*http.Request`
- `response` -> `http.ResponseWriter`
- `form` -> `url.Values`
- `someID` with corresponding path identifier `/{someID}` -> `string`

The types for `form` and `someID` can be overridden by providing a `--receiver-type` flag to `muxt generate`.
When you do this, `muxt` will generate a method that finds the method parameter type and generates a parser from string to that type.

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

Note, the default result type is `any` if you do not provide a `--receiver-type` flag to `muxt generate`.

### Example with Receiver Type

Now, when you provide `--receiver-type=Server`, `muxt` now will generate parsers in the handler and the generated interface
and use the method signature result in the generated interface method signature.

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

## String Parsing
 
`muxt` can generate form field and path parameter string parsers for most basic Go types.

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
`muxt` will use that.
