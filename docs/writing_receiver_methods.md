# Writing Receiver Methods

When writing your `muxt` receiver methods, **avoid** passing the raw `request *http.Request` or `response http.ResponseWriter` whenever you can.
Instead, use on domain-oriented structs.
This guideline clarifies your application code and your tests easier to maintain.

```go
func (MyReceiver) CreateUser(ctx context.Context, form CreateUserForm) (User, error) {
// Implementation that doesn’t worry about HTTP details
}
```

As a Go proverb says, `return static types`. Doing this will with `muxt check`.
`muxt` will try to generate new methods on the `RoutesReciever` interface and have them return `any` (I might change this initial result to `struct{}`).
Once you are updating the receiver method signatures, run `muxt generate` to get an updated interface.

## Handling Responses Without http.ResponseWriter


### Happy Path

If the "happy path" of your endpoint should be something other than `http.StatusOK`, you can set the status code in the template name.

```gotemplate
{{define "POST /user 201 CreateUser(ctx, userID)"}}
<!-- User Profile -->
{{end}}
```

### Result Data Method

If your result type has information that is useful to determine the correct status code, implement the ` interface { StatusCode() int }` interface and muxt will call it.

```go
type Data struct{
	code int
}

func (d Data) StatusCode() int {
    return d.code
}
```

### Result Data Field

If your result type is a struct and it has a field called `StatusCode int`, muxt will use that value as the status code.

```go
type Data struct{
	StatusCode int
}
```

### Change The Status Code in a Template Action

If you need to change the status code in a template action, you can do so by using the `SetStatusCode` method on the `http.ResponseWriter`.

```gotemplate
{{define "GET /user/{id} ReadUser(ctx, id)"}}
  {{- if .Err }}
    {{- with and (.StatusCode 400) (.Header "HX-Retarget" "#error") (.Header "HX-Reswap" "outerHTML")}}
      <div id='error'>{{.Err.Error}}</div>
    {{- end}}
  {{- else}}
    {{- template "user profile" .}}
  {{- end}}
{{end}}
```

### When You Really Need http.ResponseWriter

There are some cases (like streaming large file downloads or sending a specific header) where passing the response is unavoidable.
In those situations remember to assert WriteHeader is called.
Make sure to set relevant headers and set the status code but don't call `response.Write`. The handler will do that.

If you need to have more control, just register your own handler outside the generated routes function on the `http.ServeMux`.
