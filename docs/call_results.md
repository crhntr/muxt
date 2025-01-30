# Method Result Field Sets

This is the (wip) "returns" documentation.

It's kinda similar to template functions.

You *should* just return one struct value (including maybe one or more error fields).

However, you can also return a value and an error or a boolean.

The following would be acceptable result sets.

```go
package domain

type T struct{
	MissingDataReason error
}

type Server struct{}

func (Server) F1() T { return T{}}
func (Server) F2() (T, bool) { return T{}, false}
func (Server) F3() (T, error) { return T{}, nil}
```