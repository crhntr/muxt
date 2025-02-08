# Making Template Source Files Discoverable

Muxt needs your template source files to be embedded in the package in the current directory for it to discover and
parse them (see the "Go Generate Comment Example" above).

```go
package server

import (
	"embed"
	"html/template"
)

//go:embed *.gohtml
var templatesSource embed.FS

var templates = template.Must(template.ParseFS(templatesSource, "*"))
```

You need to add a globally scoped variable with type `embed.FS` (like `templatesSource` in the example).
It should be passed into a call either the function `"html/template".ParseFS` or method
`"html/template".Template.ParseFS`.
Before it does so, it can call any of the following functions or methods in the right hand side of the `templates`
variable declaration.

Muxt will call any of the functions:

- [`Must`](https://pkg.go.dev/html/template#Must)
- [`Parse`](https://pkg.go.dev/html/template#Parse)
- [`New`](https://pkg.go.dev/html/template#New)
- [`ParseFS`](https://pkg.go.dev/html/template#ParseFS)

or methods:

- [`Template.Parse`](https://pkg.go.dev/html/template#Template.Parse)
- [`Template.New`](https://pkg.go.dev/html/template#Template.New)
- [`Template.ParseFS`](https://pkg.go.dev/html/template#Template.ParseFS)
- [`Template.Delims`](https://pkg.go.dev/html/template#Template.Delims)
- [`Template.Option`](https://pkg.go.dev/html/template#Template.Option)
- [`Template.Funcs`](https://pkg.go.dev/html/template#Template.Option)

Muxt iterates over the resulting parsed templates to discover templates matching the template name pattern documented in
the "Naming Templates" section below.

You can use a different variable name for the `*template.Template` just invoke `muxt generate` with the
`--templates-variable=someOtherName` flag
and ensure you have a globally scoped variable someOtherName where the right hand side of the expression is
`template.Must()` with some parse calls.  
