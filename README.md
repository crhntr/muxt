# MUXT lets you register HTTP routes from your Go HTML Templates [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/muxt.svg)](https://pkg.go.dev/github.com/crhntr/muxt)

This is especially helpful when you are writing HTMX.

## Example

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <title>Hello, world!</title>
    <script src="https://unpkg.com/htmx.org@2.0.1"
            integrity="sha384-QWGpdj554B4ETpJJC9z+ZHJcA/i59TyjxEPXiiUgN2WmTyV5OEZWCD6gQhgkdpB/"
            crossorigin="anonymous"></script>
    <script src="https://unpkg.com/htmx-ext-response-targets@2.0.0/response-targets.js"></script>

    <style>
        #error {
            background: lightcoral;
            font-weight: bold;
            padding: .25rem;
            border-radius: .25rem;
        }

        #error:empty {
            display: none;
        }
    </style>

</head>
<body hx-ext="response-targets">
<table>
    <thead>
    <tr>
        <th>Fruit</th>
        <th>Count</th>
    </tr>
    </thead>
    <tbody>

    {{- range . -}}
    {{- block "fruit row" . -}}
    <tr>
        <td>{{ .Fruit }}</td>
        <td hx-get="/fruits/{{.Fruit}}/edit" hx-include="this" hx-swap="outerHTML" hx-target="closest tr">{{ .Count }}
            <input type="hidden" name="count" value="{{.Count}}">
        </td>
    </tr>
    {{- end -}}
    {{- end -}}


    {{- define "GET /fruits/{fruit}/edit" -}}
    <tr>
        <td>{{ .PathValue "fruit" }}</td>
        <td>
            <form hx-patch="/fruits/{{.PathValue " fruit
            "}}" hx-target-error="#error">
            <input aria-label="Count" type="number" name="count" value="{{ .FormValue " count" }}" step="1" min="0">
            <input type="submit" value="Update">
            </form>
            <p id="error"></p>
        </td>
    </tr>
    {{- end -}}

    {{- define "PATCH /fruits/{fruit} EditRow(response, request, fruit)" }}
    {{template "fruit row" .}}
    {{ end -}}

    </tbody>
</table>
</body>
</html>
```

```go
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
```