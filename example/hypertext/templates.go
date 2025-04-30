package hypertext

import (
	"embed"
	"html/template"
)

//go:generate go run ../../cmd/muxt generate --receiver-type Backend --receiver-type-package github.com/typelate/muxt/example --routes-func TemplateRoutes
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o internal/fake/routes_receiver.go -fake-name Backend . RoutesReceiver

//go:embed *.gohtml
var templateSource embed.FS

var templates = template.Must(template.ParseFS(templateSource, "*"))

type Row struct {
	ID    int
	Name  string
	Value int
}

type EditRowPage struct {
	Row   Row
	Error error
}

type EditRow struct {
	Value int `name:"count" template:"count-input"`
}
