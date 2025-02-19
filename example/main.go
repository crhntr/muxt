package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/crhntr/muxt/example/hypertext"
)

type Backend struct {
	sync.RWMutex
	data []hypertext.Row
}

func (b *Backend) SubmitFormEditRow(fruitID int, form hypertext.EditRow) hypertext.EditRowPage {
	b.Lock()
	defer b.Unlock()
	if fruitID < 0 || fruitID >= len(b.data) {
		return hypertext.EditRowPage{Error: fmt.Errorf("fruit not found")}
	}
	row := b.data[fruitID]
	row.Value = form.Value
	b.data[fruitID] = row
	return hypertext.EditRowPage{Error: nil, Row: row}
}

func (b *Backend) GetFormEditRow(fruitID int) hypertext.EditRowPage {
	b.RLock()
	defer b.RUnlock()
	if fruitID < 0 || fruitID >= len(b.data) {
		return hypertext.EditRowPage{Error: fmt.Errorf("fruit not found")}
	}
	return hypertext.EditRowPage{Error: nil, Row: b.data[fruitID]}
}

func (b *Backend) List(_ context.Context) []hypertext.Row {
	b.RLock()
	defer b.RUnlock()
	return slices.Clone(b.data)
}

func main() {
	backend := &Backend{
		data: []hypertext.Row{
			{ID: 0, Name: "Peach", Value: 10},
			{ID: 1, Name: "Plum", Value: 20},
			{ID: 2, Name: "Pineapple", Value: 2},
		},
	}
	mux := http.NewServeMux()
	hypertext.TemplateRoutes(mux, backend)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
