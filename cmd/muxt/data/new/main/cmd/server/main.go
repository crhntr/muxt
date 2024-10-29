package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"github.com/crhntr/muxt/cmd/muxt/data/new/main/internal/hypertext"
)

func main() {
	srv := hypertext.Server{}
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	log.Fatal(http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), mux))
}
