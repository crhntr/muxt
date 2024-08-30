package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

//go:embed *.gohtml
var templateSource embed.FS

var templates = template.Must(template.ParseFS(templateSource, "*"))

type Backend struct {
	data []Row
}

type EditRowPage struct {
	Row   Row
	Error error
}

type EditRow struct {
	Value int `name:"count"`
}

func (b *Backend) SubmitFormEditRow(fruitID int, form EditRow) EditRowPage {
	if fruitID < 0 || fruitID >= len(b.data) {
		return EditRowPage{Error: fmt.Errorf("fruit not found")}
	}
	row := b.data[fruitID]
	row.Value = form.Value
	return EditRowPage{Error: nil, Row: row}
}

func (b *Backend) GetFormEditRow(fruitID int) EditRowPage {
	if fruitID < 0 || fruitID >= len(b.data) {
		return EditRowPage{Error: fmt.Errorf("fruit not found")}
	}
	return EditRowPage{Error: nil, Row: b.data[fruitID]}
}

type Row struct {
	Name  string
	Value int
}

func (b *Backend) List(_ context.Context) []Row { return b.data }

//go:generate go run ../cmd/muxt generate --receiver-static-type Backend

func main() {
	backend := &Backend{
		data: []Row{
			{Name: "Peach", Value: 10},
			{Name: "Plum", Value: 20},
			{Name: "Pineapple", Value: 2},
		},
	}
	mux := http.NewServeMux()
	routes(mux, backend)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
