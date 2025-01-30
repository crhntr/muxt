# Testing Output

I'd highly suggest using my other package

[`import "github.com/crhntr/dom"`](https://pkg.go.dev/github.com/crhntr/dom/domtest)

I generally write tests that look something like this (but usually table driven):

```go
package hypertext

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crhntr/dom/domtest"
	"github.com/stretchr/testify/require"

	"example.com/internal/fake"
)

func TestRoutes(t *testing.T) {
	mux := http.NewServeMux()
	srv := new(fake.Server)
	srv.GreetReturns("Greetings, Jimmy!")
	routes(mux, srv)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/32", nil)
	mux.ServeHTTP(rec, req)

	res := rec.Result()

	document := domtest.Response(t, res)

	require.NotNil(t, document.QuerySelector("#some-id"))

	_, id := srv.GreetArgsForCall()
	require.EqualValues(t, 32, id)
}
```

I generate my server fake using [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter).