package main

import (
	"bytes"
	"cmp"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/crhntr/muxt"
)

//go:embed *.gohtml
var formHTML embed.FS

func main() {
	s := &Server{
		templates: template.Must(template.ParseFS(formHTML, "*")),
		table: []Row{
			{Fruit: "Peach", Count: 1},
			{Fruit: "Pear", Count: 2},
			{Fruit: "Plum", Count: 3},
			{Fruit: "Pineapple", Count: 4},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Index)
	if err := muxt.Handlers(mux, s.templates, muxt.WithReceiver(s).WithErrorFunc(noopErr)); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), mux))
}

func noopErr(http.ResponseWriter, *http.Request, *template.Template, *slog.Logger, error) {}

type Row struct {
	Fruit string
	Count int
}

type Server struct {
	table     []Row
	templates *template.Template
}

func (s *Server) Index(res http.ResponseWriter, _ *http.Request) {
	buf := bytes.NewBuffer(nil)
	if err := s.templates.ExecuteTemplate(buf, "form.gohtml", s.table); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	_, _ = io.Copy(res, buf)
}

func (s *Server) EditRow(res http.ResponseWriter, req *http.Request, fruit string) (Row, error) {
	count, err := strconv.Atoi(req.FormValue("count"))
	if err != nil {
		http.Error(res, "failed to parse count: "+err.Error(), http.StatusBadRequest)
		return Row{}, err
	}
	if count > 9000 {
		err = fmt.Errorf("count must not exceed 9000")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return Row{}, err
	}
	for i, row := range s.table {
		if row.Fruit != fruit {
			continue
		}
		res.Header().Set("HX-Retarget", "closest tr")
		res.Header().Set("HX-Reswap", "outerHTML")
		s.table[i].Count = count
		res.WriteHeader(http.StatusOK)
		return Row{
			Fruit: fruit,
			Count: count,
		}, nil
	}
	err = fmt.Errorf("row not found")
	http.Error(res, err.Error(), http.StatusNotFound)
	return Row{}, err
}
