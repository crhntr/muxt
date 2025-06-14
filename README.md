# Muxt [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/muxt.svg)](https://pkg.go.dev/github.com/crhntr/muxt) [![Go](https://github.com/crhntr/muxt/actions/workflows/go.yml/badge.svg)](https://github.com/crhntr/muxt/actions/workflows/go.yml)

**Muxt** generates and registers HTTP Handler functions specified in HTML templates.
It increases locality of behavior when creating server side rendered hypermedia web applications.

Muxt looks for templates with names that match an extended version of the `http.ServeMux` pattern syntax.

The standard `http.ServeMux` pattern syntax looks like this:

> `[METHOD ][HOST]/[PATH]`

Muxt extends this by adding optional fields for an HTTP status and a call:

> `[METHOD ][HOST]/[PATH][ HTTP_STATUS][ CALL]`

### Route Registration Example

You tell `muxt` how to generate the handler functions by defining templates like this `{{define "GET / F()" -}}`.
`muxt` will generate a handler function that calls F and pass the result to the template.
The template result will then be written to the HTTP response.

```html
{{define "GET / F()" -}}
<!DOCTYPE html>
<html lang='en'>
<head>
    <meta charset='UTF-8'/>
    <title>Hello!</title>
</head>
<body>
<h1>Number {{.Result}}</h1>
</body>
</html>
{{- end}}
```

### Tiny Examples

`muxt` routes HTML templates to Go methods and handles common web plumbing:

* `{{define "GET /{id} F(id)"}}{{end}}` — Parses `{id}` as `int` and passes to `F(int)`.
* `{{define "GET / F(ctx)"}}{{end}}` — Injects `request.Context()` if `ctx` is used.
* `{{define "GET / F(request)"}}{{end}}` — Injects the `*http.Request` when `request` is named.
* `{{define "GET / F(response)"}}{{end}}` — Injects `http.ResponseWriter` if `response` is used.
* `{{define "POST / F(form)"}}{{end}}` — Parses form data into a struct from `url.Values` if the `form` parameter is a struct.
* `{{define "POST / F(form)"}}{{end}}` — Parses form data into a struct if the `form` parameter is a `url.Values`.

The result of the call is wrapped in a `TemplateData[T]` struct and passed to `ExecuteTemplate`.

### Bigger Examples

For a small runnable, see: [./example/hypertext/index.gohtml](./example/hypertext/index.gohtml)
The Go package documentation for the example shows what is generated `https://pkg.go.dev/github.com/crhntr/muxt/example/hypertext`.

For larger complete examples, see:
- [muxt-example-htmx-sortable](http://github.com/crhntr/muxt-example-htmx-sortable) _**(NEW)**_
- [muxt-template-module-htmx](https://github.com/crhntr/muxt-template-module-htmx)

## License

Muxt is licensed under the [GNU AGPLv3](LICENSE).

However, the Go code generated by `muxt` is **not** covered by the AGPL.
It is licensed under the MIT license (see `https://choosealicense.com/licenses/mit/`).
The code generated by `muxt` is provided as-is, without a warranty of any kind.
The `muxt` author disclaims all liability for any bugs, regressions, or defects in generated output.
## Documentation

### Introduction

- [Getting Started](./docs/getting_started.md)
- [Notes on Integration with Existing Projects](./docs/integrating.md)
- [Writing Template Names](./docs/template_names.md)

### Reference

- [Call Parameters](./docs/call_parameters.md)
- [Call Results](./docs/call_results.md)
- [Writing Receiver Methods](./docs/writing_receiver_methods.md)
- [Templates Variable](./docs/templates_variable.md)
- [Template Action Type-Checking](./docs/action_type_checking.md)
- [Known Issues](./docs/known_issues.md)

### Testing

- [Testing Hypertext](./docs/testing_hypertext.md)

### Philosophy & Vision

- [Manifesto](./docs/manifesto.md)
- [Motivation](./docs/motivation.md)
- Goals:
  [see enhancement issues](https://github.com/crhntr/muxt/issues?q=is%3Aissue%20state%3Aopen%20label%3Aenhancement)


## Prompting

- [Prompting Helpers](./docs/prompts)