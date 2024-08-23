package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
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

func (b *Backend) SubmitFormEditRow(request *http.Request, fruitID int) EditRowPage {
	if fruitID < 0 || fruitID >= len(b.data) {
		return EditRowPage{Error: fmt.Errorf("fruit not found")}
	}
	count, err := strconv.Atoi(request.FormValue("count"))
	if err != nil {
		return EditRowPage{Error: err, Row: b.data[fruitID]}
	}
	row := b.data[fruitID]
	row.Value = count
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
