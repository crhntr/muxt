package hypertext

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crhntr/dom/domtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crhntr/muxt/cmd/muxt/data/new/main/internal/fake"
)

func TestTemplates(t *testing.T) {
	for _, tt := range []struct {
		Name     string
		Request  func(rsv *fake.Server) *http.Request
		Response func(rsv *fake.Server, res *http.Response)
	}{
		{
			Name: "the header has the name",
			Request: func(srv *fake.Server) *http.Request {
				srv.IndexReturns(IndexData{
					Name: "somebody",
				})
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			Response: func(rsv *fake.Server, res *http.Response) {
				if assert.Equal(t, 1, rsv.IndexCallCount()) {
					ctx := rsv.IndexArgsForCall(0)
					require.NotNil(t, ctx)
				}
				assert.Equal(t, http.StatusOK, res.StatusCode)
				doc := domtest.Response(t, res)
				if el := doc.QuerySelector(`h1`); assert.NotNil(t, el) {
					assert.Equal(t, "Hello, somebody!", strings.TrimSpace(el.TextContent()))
				}
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			srv := new(fake.Server)
			mux := http.NewServeMux()
			routes(mux, srv)
			rec := httptest.NewRecorder()
			req := tt.Request(srv)
			mux.ServeHTTP(rec, req)
			tt.Response(srv, rec.Result())
		})
	}
}
