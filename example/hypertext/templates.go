package hypertext

import (
	"embed"
	"html/template"
)

//go:generate go run ../../cmd/muxt generate --receiver-type Backend --receiver-type-package github.com/crhntr/muxt/example --routes-func TemplateRoutes
//go:generate counterfeiter -generate
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
	Value int `name:"count" template:"count"`
}
