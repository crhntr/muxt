# LLM Architecture Prompt: HTML Templates for Muxt

## Purpose

Templates must define HTTP routes, bind data returned from Go domain methods, and be parsed with `html/template`.  
Muxt generates handlers by reading template names and type-checking template actions against Go code.

## Template Naming

Each `.gohtml` template must define a top-level template with a name following this exact structure:

```
METHOD [HOST]/path [HTTP_STATUS] MethodName(ctx, param1, param2, ...)
```

Examples:

```gotemplate
{{define "GET /authors/{id} 200 ShowAuthor(ctx, id)"}}
{{end}}

{{define "POST /authors 303 CreateAuthor(ctx, form)"}}
{{end}}
```

Rules:
- `METHOD` is the HTTP method (GET, POST, PATCH, etc.).
- `[HOST]` is optional.
- `/path` can include `{param}` path variables.
- `[HTTP_STATUS]` is the HTTP status code sent if handler succeeds.
- `MethodName` must match a method on the Go receiver passed to `TemplateRoutes`.
- Parameters must exactly match Go method parameters.

## Template Body

Templates must:
- Be valid, well-formed HTML5.
- Use Go template actions (`{{.Field}}`, `{{template "name"}}`, etc.) consistent with the type returned by the handler.
- Prefer semantic HTML elements (`<main>`, `<section>`, `<form>`, etc.).
- Only use safe inline scripts or styles if necessary.
- Avoid dynamic element attributes unless compatible with Muxt type checking.

Example:

```gotemplate
{{define "GET /authors/{id} 200 ShowAuthor(ctx, id)"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Result.Name}}</title>
</head>
<body>
    <h1>{{.Result.Name}}</h1>
    <p>Born in {{.Result.BirthYear}}</p>
    {{if .Result.Bio}}
    <section>
        <h2>Biography</h2>
        <p>{{.Result.Bio}}</p>
    </section>
    {{end}}
</body>
</html>
{{end}}
```

## Context Available in Templates

Each template receives a `TemplateData[T]` struct:

```go
type TemplateData[T any] struct {
	Request *http.Request
	Result  T
}
```

Use `.Request` for limited HTTP metadata (e.g., headers or URL parameters).  
Use `.Result` fields populated from the Go domain method's return value.

## Strict Constraints

- All template actions must match Go types at compile time where possible (`muxt check` must pass).
- All HTML must be syntactically correct to allow static parsing and Muxt type-checking.
- Template names must match exactly the corresponding domain method names and parameters.

## Exclusions

- Do not assume JavaScript frameworks or dynamic frontend behavior outside standard HTML + Go template syntax.
- Do not define templates without valid Muxt-compliant names.
- Do not reference untyped dynamic objects (e.g., `interface{}`) without explicit structure.

## Enforcement

- Templates must generate a complete HTML document unless intentionally returning fragments.
- All placeholders must match generated Go types.
- Route declarations in template names must match the intended application router paths exactly.
