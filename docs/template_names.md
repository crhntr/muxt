# Naming Templates

`muxt generate` will read your HTML templates and generate/register an [
`http.HandlerFunc`](https://pkg.go.dev/net/http#HandlerFunc) for each template with a name that matches an expected
patten.

If a template name does not match an expected pattern, the template is ignored by `muxt`.

Since Go 1.22, the standard library route **mu**ltiple**x**er can parse path parameters.

It has expects strings like this

`[METHOD ][HOST]/[PATH]`

Muxt extends this by adding optional fields for the status code and a method call.

`[METHOD ][HOST]/[PATH ][HTTP_STATUS ][CALL]`

A template name pattern that `muxt` understands looks like this:

```gotemplate
{{define "GET /greet/{language} 200 Greeting(ctx, language)" }}
<h1>{{.Hello}}</h1>
{{end}}
```

## [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux) Patterns

Here is an excerpt from [the standard libary documentation.](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)

> Patterns can match the method, host and path of a request. Some examples:
> - "/index.html" matches the path "/index.html" for any host and method.
> - "GET /static/" matches a GET request whose path begins with "/static/".
> - "example.com/" matches any request to the host "example.com".
> - "example.com/{$}" matches requests with host "example.com" and path "/".
> - "/b/{bucket}/o/{objectname...}" matches paths whose first segment is "b" and whose third segment is "o". The name "
    bucket" denotes the second segment and "objectname" denotes the remainder of the path.

## More Precise Template Name Specification

```bnf
<route> ::= [ <method> <space> ] [ <host> ] <path> [ <space> <http_status> ] [ <space> <call_expr> ]

<method> ::= "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS"

<host> ::= <hostname> | <ip_address>

<hostname> ::= <label> { "." <label> }
<label> ::= <letter> { <letter> | <digit> | "-" }
<ip_address> ::= <digit>+ "." <digit>+ "." <digit>+ "." <digit>+

<path> ::= "/" [ <path_segment> { "/" <path_segment> } [ "/" ] ]
<path_segment> ::= <unreserved_characters>+

<http_status> ::= <integer> | <qualified_identifier>
<integer> ::= <digit> { <digit> }
<qualified_identifier> ::= <identifier> "." <identifier>

<call_expr> ::= <identifier> "(" [ <identifier> { "," <identifier> } ] ")"

<identifier> ::= <letter> { <letter> | <digit> | "_" }

<space> ::= " "

<letter> ::= "a" | ... | "z" | "A" | ... | "Z"
<digit> ::= "0" | ... | "9"
<unreserved_characters> ::= <letter> | <digit> | "-" | "_" | "." | "~"
```

_TODO add more documentation on form and typed arguments_
