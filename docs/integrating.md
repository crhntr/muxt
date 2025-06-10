# Integrating Muxt into an Existing Project

`muxt` is designed bring value to existing projects quickly.

If you are already using `"html/template"`, you can do a couple quick changes to get static analysis of that source.
1. Make your `*template.Template` variable is a initialized as a global declaration and that the source is provided via embedded files
  ```go
  package server
  
  import (
    "embed"
    "html/template"
  )
  
  //go:embed *.gohtml
  var templateSource embed.FS
  
  var templates = template.Must(template.ParseFS(templateSource, "*"))
  ```
2. Make sure all your calls to `templates.ExecuteTemplate` use string literals for the name and static types for the data argument.
That is all you need to do to have `muxt check` check your templates.
This should make refactoring your templates safer.

To have `muxt` map HTTP requests to method calls, you can now start integrating `muxt generate` 

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

