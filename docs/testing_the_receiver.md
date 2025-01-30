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
If you really need to set particular headers, consider using a custom execute function and type assert the data
parameter using some interface.

## When You Really Need http.ResponseWriter

There are some cases (like streaming large file downloads or sending a specific header) where passing the response is
unavoidable.
In those situations remember to assert WriteHeader is called.
Dont' call `response.Write`. That's for execute.

 

