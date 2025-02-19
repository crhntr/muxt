# Method Result Field Sets

It's kinda similar to template functions.
You *should* just return one struct value (including maybe one or more error fields).
However, you can also return a value and an error or a boolean.

The following would be acceptable result sets.

```go
package domain

type T struct {
	MissingDataReason error
}

type Server struct{}

func (Server) F1() T          { return T{} }
func (Server) F2() (T, bool)  { return T{}, false }
func (Server) F3() (T, error) { return T{}, nil }
func (Server) F4() error { return nil } // not sure why you'd do this

```

Before the left most value is passed to the template, it is boxed in a struct that also includes the `*http.Request`.

So in your template you will receive a struct like this and T will be the left most return from your method:

```go
package hypertext

import "net/http"

type TemplateData[T any] struct {
	Request *http.Request
	Result    T
}
```

### Roadmap Notes

I'd like to add methods on this type to generate URLs based on the template routes.
This will make using URLs in your templates more type safe.   