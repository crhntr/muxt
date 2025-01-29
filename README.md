# Muxt [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/muxt.svg)](https://pkg.go.dev/github.com/crhntr/muxt) [![Go](https://github.com/crhntr/muxt/actions/workflows/go.yml/badge.svg)](https://github.com/crhntr/muxt/actions/workflows/go.yml)

Since Go 1.22, the standard library route **mu**ltiple**x**er [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux) uses http methods, hosts, and path parameters.
Muxt extends this syntax to add method signatures and type static analysis based template type safety to make it faster to write and test server side rendered hypermedia web applications.

Muxt generates Go code. It does not require you to add any dependencies outside the Go standard library.

- It allows you to register HTTP routes from [HTML templates](https://pkg.go.dev/html/template)
- It generates handler functions and registers them on an [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux)
- It generates code in handler functions to parse path parameters and form fields
- It generates a receiver interface to represent the boundary between your app code and HTTP/HTML
  - Use this to mock out your server and test the view layer of your application

### Used By
- [portfoliotree.com](https://portfoliotree.com)

## Installation

You can install it using the Go toolchain.
```bash
cd # Change outside of your module (so you don't add muxt to your dependency chain)
go install github.com/crhntr/muxt@latest
cd -
```
You do not need to add this tool to your module ([unless you want to use the tools pattern](https://play-with-go.dev/tools-as-dependencies_go119_en/)).

## Usage

Commands:
- `muxt generate` generate a routes function and receiver interface
- `muxt version` writes muxt version to standard out
- `muxt check` static type check your templates
- `muxt documentation` (wip) template documentation

### Generate Command

This command is how you use muxt. 

You can call it either by invoking it from your terminal or by adding a generate comment to a Go source file.

<details>
<summary>Flags</summary>
<pre>
Usage of generate:
  -output-file string
    	The generated file name containing the routes function and receiver interface. (default "template_routes.go")
  -receiver-interface string
    	The interface name in the generated output-file listing the methods used by the handler routes in routes-func. (default "RoutesReceiver")
  -receiver-type string
    	The type name for a named type to use for looking up method signatures. If not set, all methods added to the receiver interface will have inferred signatures with argument types based on the argument identifier names. The inferred method signatures always return a single result of type any.
  -receiver-type-package string
    	The package path to use when looking for receiver-type. If not set, the package in the current directory is used.
  -routes-func string
    	The function name for the package registering handler functions on an *"net/http".ServeMux.
    	This function also receives an argument with a type matching the name given by receiver-interface. (default "routes")
  -templates-variable string
    	the name of the global variable with type *"html/template".Template in the working directory package. (default "templates")

</pre>
</details>

#### Shell Example

If you invoke it from the shell, it expects to find a Go source package in the current directory where it can find a templates variable.

```shell
muxt generate
```

#### Go Generate Comment Example

If you do the generate comment, make sure you need to write the comment in the same package as your (globally scoped) `templates` variable.

```go
package main

import (
    "embed"
    "html/template"
)

//go:generate muxt generate

var (
    //go:embed *.gohtml
    templatesSource embed.FS

    templates = template.Must(template.ParseFS(templatesSource, "*.gohtml"))
)

```

### Making Template Source Files Discoverable

Muxt needs your template source files to be embedded in the package in the current directory for it to discover and parse them (see the "Go Generate Comment Example" above).

You need to add a globally scoped variable with type `embed.FS` (like `templatesSource` in the example).
It should be passed into a call either the function `"html/template".ParseFS` or method `"html/template".Template.ParseFS`.
Before it does so, it can call any of the following functions or methods in the right hand side of the `templates` variable declaration.

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

Muxt iterates over the resulting parsed templates to discover templates matching the template name pattern documented in the "Naming Templates" section below. 

### Naming Templates

`muxt generate` will read your embedded HTML templates and generate/register an [`http.HandlerFunc`](https://pkg.go.dev/net/http#HandlerFunc) for each template with a name that matches an expected patten.

If the template name does not match the pattern, it is ignored by muxt.

Since Go 1.22, the standard library route **mu**ltiple**x**er can parse path parameters.

It has expects strings like this

`[METHOD ][HOST]/[PATH]`

Muxt extends this a little bit.

`[METHOD ][HOST]/[PATH ][HTTP_STATUS ][CALL]`

A template name that muxt understands looks like this:

```gotemplate
{{define "GET /greet/{language} 200 Greeting(ctx, language)" }}
    <h1>{{.Hello}}</h1>
{{end}}
```

In this template name
- Passed through to `http.ServeMux`
  - we define the HTTP Method `GET`,
  - the path prefix `/greet/`
  - the path parameter called `language` (available in the call scope as `language`)
- Used by muxt to generate a `http.HandlerFunc`
  - the status code to use when muxt calls WriteHeader is `200` aka `http.StatusOK`
  - the method name on the configured receiver to call is `Greeting`
  - the parameters to pass to `Greeting` are `ctx` and `language`

#### [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux) Patterns¶

Here is an excerpt from [the standard libary documentation.](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)

> Patterns can match the method, host and path of a request. Some examples:
> - "/index.html" matches the path "/index.html" for any host and method.
> - "GET /static/" matches a GET request whose path begins with "/static/".
> - "example.com/" matches any request to the host "example.com".
> - "example.com/{$}" matches requests with host "example.com" and path "/".
> - "/b/{bucket}/o/{objectname...}" matches paths whose first segment is "b" and whose third segment is "o". The name "bucket" denotes the second segment and "objectname" denotes the remainder of the path.

#### The Method Call Scope

There are three parameters you can pass to a method that always generate the same code

- `ctx` -> `http.Request.Context`
- `request` -> `*http.Request`
- `response` -> `http.ResponseWriter` (if you use this, muxt won't generate code to call WriteHeader, you have to do this)

Using these three, the generated code will look something like this.

Given `{{define "GET / F(ctx, response, request)"}}Hello{{end}}`,

You will get a handler generated like this:

```go
package main

import (
    "context"
    "net/http"
)

type RoutesReceiver interface {
  F(ctx context.Context, response http.ResponseWriter, request *http.Request) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
  mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()
    data := receiver.F(ctx, response, request)
    execute(response, request, false, "GET / F(ctx, response, request)", http.StatusOK, data)
  })
}

func execute(http.ResponseWriter, *http.Request, bool, string, int, any) {}
```

You can also map path values from the path pattern to identifiers and pass them to your handler.


Given `{{define "GET /articles/{id} ReadArticle(ctx, id)"}}{{end}}`,

You will get a handler generated like this:

```go
package main

import (
  "context"
  "net/http"
)

type RoutesReceiver interface {
  ReadArticle(ctx context.Context, id string) any
}

func routes(mux *http.ServeMux, receiver RoutesReceiver) {
  mux.HandleFunc("GET /articles/{id}", func(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()
    id := request.PathValue("id")
    data := receiver.ReadArticle(ctx, id)
    execute(response, request, true, "GET /articles/{id} ReadArticle(ctx, id)", http.StatusOK, data)
  })
}

func execute(http.ResponseWriter, *http.Request, bool, string, int, any) {}
```

_TODO add more documentation on form and typed arguments_

## Examples

The [example directory](example) has a worked example.

For a more complete example, see: https://github.com/crhntr/muxt-template-module-htmx
