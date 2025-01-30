# Integrating Muxt into an Existing Project

If you already have a Go application and want to introduce **Muxt** for server-rendered templates, consider placing your
`.gohtml` files and hypermedia-related code into a dedicated package
A convenient choice is something like `hypertext/`, which keeps template and routing logic isolated from the rest of
your application.
Http routes and hypertext tend to be highly coupled in SSR sites (for example via anchor tags `<a href="/some-route">`
and form actions `<form method="GET" action="/hello">`)
This coupling increases when you use something epic like [htmx](https://htmx.org).

## 1. Create a `hypertext` Package

In your existing repository, create a new folder

`mkdir -p internal/hypertext`

You likely want to hide this package from external callers.
(I've dreamed of using generated code for shared hypertext libraries, but that has yet to be explored.)

### Why a Separate Package?

- **Separation of Concerns**: Keeps all template-related files and the generated route code in one place, making it
  easier to maintain and test.
- **Avoid Pollution**: If your main or other packages are large, this keeps the new code and routes from cluttering
  existing files.

I've found having methods used in template actions close to the implementation really nice.

For example if you need to do more complicated control flow than would be maintainable in Go templates,
then should consider having a custom type that has a method returning [
`template.HTML`](https://pkg.go.dev/html/template#HTML).
*Remember the standard library's admonition "use of this type presents a security risk..."*

## 2. Adding Templates & an Entry Point

It is common to keep template source in a separate directory, you can leave your html there and just add a file in the
parent dir.
The "templates.go" file does not need to be colocated with source as seen in the Getting Started example.

```go
package hypertext

import (
	"embed"
	"html/template"
)

//go:embed templates/*.gohtml
var templatesDir embed.FS

//go:generate muxt generate --receiver-type=Server --receiver-type-package=example.com/internal/domain --routes-func=Routes
var templates = template.Must(template.ParseFS(templatesDir, "*.gohtml"))
```

Once you get to this step, consider running `muxt generate && muxt check` to see if your templates have any issues that
Muxt can detect before you go too far.
If the command fails see the known issues document or consider filing an issue (if you do, many thanks).

Register your routes on an existing ServeMux.

```go
package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"example.com/internal/api"
	"example.com/internal/hypertext"
	"example.com/internal/domain"
)

func main() {
	mux := http.NewServeMux()

	srv := domain.New()

	api.Routes(mux, srv) // 
	hypertext.Routes(mux, srv)

	log.Fatal(http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), mux))
}
```

