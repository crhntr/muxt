# Testing the Receiver Type

When writing your Muxt receiver methods, **avoid** passing the raw `request *http.Request` or
`response http.ResponseWriter` whenever you can.
Instead, focus on domain-oriented signatures—think `context.Context` plus arguments or structures that represent your
business data.
This makes your application code clearer and your tests easier to maintain.

```go
func (MyReceiver) CreateUser(ctx context.Context, form CreateUserForm) (User, error) {
// Implementation that doesn’t worry about HTTP details
}
```

As a Go proverb says, `return static types`. It helps with `muxt check`.

Muxt will try to generate new methods on the RoutesReciever interface and have them return any.
Replace that. It will trip up `muxt check`.

## Handling Responses Without http.ResponseWriter

If you need specific HTTP behavior—like a custom status code—consider these approaches:

### Use a Specialized Return Type

Return a struct or type that includes error details or status information.
Your template can then display relevant user-facing messages.

```go
type UpdateArticle {
  Error     error // even when this is set, return http.StatusOK (200)
  PlainText string
  Markdown  template.HTML
}
```

## When You Really Need http.ResponseWriter

There are some cases (like streaming large file downloads or sending a specific header) where passing the response is unavoidable.
In those situations remember to assert WriteHeader is called.
Make sure to set relevant headers and set the status code but dont' call `response.Write`. That's for execute.

### Example

This type of example sometimes makes sense for parsing authentication inforamtion from the `*http.Request`.
You can set this up like so:

```gotemplate
{{define "GET /user/settings/{userID} UserSettings(ctx, SessionRedirectUnauthenticated(res, req), userID)"}}
<!-- User Settings -->
{{end}}
```

```go
func (MyReceiver) SessionRedirectUnauthenticated(http.ResponseWriter, *http.Request) (SessionClaims, bool) {
  // ...
}

func (MyReceiver) UserSettings(ctx context.Context, SessionClaims, userID int) UserSettings { /* ... */ }

```

When the second result from `SessionRedirectUnauthenticated` returns `true`, the next function is called.
When the second result from `SessionRedirectUnauthenticated` returns `false` the handler returns early and `UserSettings` is not called.
