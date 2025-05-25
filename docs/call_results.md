# Method Result Field Sets

How `muxt` handles the results of method calls, is similar to how template.ExecuteTemplate handles functions.
You must not call a function without a result.
You can return a single value, a value and an error, or a value and a boolean.
When you return a single value, the error and boolean are assumed to be zero values.
The type of the first result will be used as the type parameter for the `TemplateData[T]` struct.
The struct will be passed to the template as the data argument.

The following methods on T would be acceptable method result signatures.

```go
package domain

type T struct {
	MissingDataReason error
}

type Server struct{}

func (Server) F1() T          { return T{} }
func (Server) F2() (T, bool)  { return T{}, false } // if the boolean is true, the handler will return without writing a response.
func (Server) F3() (T, error) { return T{}, nil }
func (Server) F4() error { return nil }

```

Before the left most value is passed to the template, it is boxed in a struct that also includes the `*http.Request`.

So in your template you will receive a struct like this.
T will be the left most return from your method:

```go
package hypertext

type TemplateData[T any] struct {}
```

The `TemplateData` type will have accessors for
- `Request` (the `*http.Request`)
- `Result` (the left most return value from your method)

Other methods on TemplateData exist. These are in active development and are likely to change.