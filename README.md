# Muxt [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/muxt.svg)](https://pkg.go.dev/github.com/crhntr/muxt) [![Go](https://github.com/crhntr/muxt/actions/workflows/go.yml/badge.svg)](https://github.com/crhntr/muxt/actions/workflows/go.yml)

**Early WIP (not yet tested in prod)**

Sometimes as a developer it is nice to stay in an HTML headspace. This Go code generator helps you do that.
It also provides a nice test seam between your http and endpoint handlers.

Muxt generates Go code. It does not require you to add any dependencies outside the Go standard library. 

- It allows you to register HTTP routes from [HTML templates](https://pkg.go.dev/html/template)
- It generates handler functions and registers them on an [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux)
- It generates code in handler functions to parse path parameters and form fields
- It generates a receiver interface to represent the boundary between your app code and HTTP/HTML
  - Use this to mock out your server and test the view layer of your application

## Installation

You can install it using the Go toolchain.
```bash
cd # Change outside of your module (so you don't add muxt to your dependency chain)
go install github.com/crhntr/muxt@latest
cd -
```

You do not need to add this tool to your module ([unless you want to use the tools pattern](https://play-with-go.dev/tools-as-dependencies_go119_en/)).

## Usage

### Templates

`muxt generate` will read your HTML templates and generate and register [`http.HandlerFunc`](https://pkg.go.dev/net/http#HandlerFunc)
on a for templates with names that an expected patten.

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
  - we define the HTTP Method `GET`0,
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

```
mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
  ctx := request.Context()
  data := receiver.F(ctx, resposne, request)
  execute(response, request, false, "GET / F(ctx, response, request)", http.StatusOK, data)
})
```

You can also map path values from the path pattern to identifiers and pass them to your handler.


Given `{{define "GET /articles/:id ReadArticle(ctx, id)"}}{{end}}`,

You will get a handler generated like this:

```
mux.HandleFunc("GET /", func(response http.ResponseWriter, request *http.Request) {
  ctx := request.Context()
  id := request.PathValue("id")
  data := receiver.ReadArticle(ctx, id)
  execute(response, request, true, "GET /articles/:id ReadArticle(ctx, id)", http.StatusOK, data)
})
```

_TODO add more documentation on form and typed arguments_

## Examples

The [example directory](example) has a worked example.

For a more complete example, see: https://github.com/crhntr/muxt-template-module-htmx
