# Method Result Field Sets

It's kinda similar to template functions.
You can also return a value and an error or an "ok" boolean.

The following would be acceptable result sets.

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

So in your template you will receive a struct like this and T will be the left most return from your method:

```go
package hypertext

type TemplateData[T any] struct {}
```

The `TemplateData` type will have accessors for
- `Request` (the `*http.Request`)
- `Result` (the left most return value from your method)

Other methods on TemplateData exist. These are in active development and are likely to change.