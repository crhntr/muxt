# Naming Templates

`muxt generate` will read your embedded HTML templates and generate/register an [
`http.HandlerFunc`](https://pkg.go.dev/net/http#HandlerFunc) for each template with a name that matches an expected
patten.

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

## [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux) Patterns

Here is an excerpt from [the standard libary documentation.](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)

> Patterns can match the method, host and path of a request. Some examples:
> - "/index.html" matches the path "/index.html" for any host and method.
> - "GET /static/" matches a GET request whose path begins with "/static/".
> - "example.com/" matches any request to the host "example.com".
> - "example.com/{$}" matches requests with host "example.com" and path "/".
> - "/b/{bucket}/o/{objectname...}" matches paths whose first segment is "b" and whose third segment is "o". The name "
    bucket" denotes the second segment and "objectname" denotes the remainder of the path.

_TODO add more documentation on form and typed arguments_
