package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/txtar"

	"github.com/crhntr/txtarfmt"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("txtarfmt: ")
	var (
		config txtarfmt.Configuration
		ext    string
	)
	flag.BoolVar(&config.SkipGo, "skip-go", false, "skip formatting Go code")
	flag.BoolVar(&config.SkipJSON, "skip-json", false, "skip formatting JSON files")
	flag.StringVar(&ext, "ext", ".txtar", "file extension filter")
	flag.Parse()
	count := 0
	for i, arg := range flag.Args() {
		matches, err := filepath.Glob(arg)
		if err != nil {
			log.Fatalf("bad command argument %d: %s", i, err)
		}
		for _, match := range matches {
			if ext != "" && filepath.Ext(match) != ext {
				continue
			}
			log.Printf("glob match: %q\n", match)
			count++
			archive, err := txtar.ParseFile(match)
			if err != nil {
				log.Fatalf("%s: %s", match, err)
			}
			info, err := os.Stat(match)
			if err != nil {
				log.Fatalf("%s: %s", match, err)
			}
			if err := txtarfmt.Archive(archive, config); err != nil {
				log.Fatalf("%s: %s", match, err)
			}
			if err := os.WriteFile(match, txtar.Format(archive), info.Mode()); err != nil {
				log.Fatalf("%s: %s", match, err)
			}
		}
	}
	if count == 0 {
		log.Fatal("no files matched")
	}
}
