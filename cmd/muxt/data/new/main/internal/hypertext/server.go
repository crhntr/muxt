package hypertext

import (
	"context"
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.gohtml
var templateFiles embed.FS

var templates = template.Must(template.ParseFS(templateFiles, "*"))

//go:generate go run github.com/crhntr/muxt/cmd/muxt generate --receiver-static-type=Server --receiver-interface-name=serverInterface
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o=../fake/server.go --fake-name=Server . serverInterface

type Server struct{}

func (srv *Server) RegisterRoutes(mux *http.ServeMux) {
	routes(mux, srv)
}

type IndexData struct {
	Name string
}

func (srv *Server) Index(_ context.Context) IndexData {
	return IndexData{
		Name: "friend",
	}
}
