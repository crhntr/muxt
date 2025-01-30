package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/crhntr/muxt/internal/configuration"
	"github.com/crhntr/muxt/internal/muxt"
)

//go:generate go run .

var (
	//go:embed README.md.template
	templateSource string
	templates      = template.Must(template.New("README.md.template").Delims("{{{", "}}}").Parse(templateSource))
)

func main() {
	var out bytes.Buffer
	gf := configuration.RoutesFileConfigurationFlagSet(new(muxt.RoutesFileConfiguration))
	gf.SetOutput(&out)
	gf.Usage()
	generateUsage := out.Bytes()
	out.Reset()

	err := templates.Execute(&out, struct {
		GenerateUsage string
	}{
		GenerateUsage: string(generateUsage),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.FromSlash("../../README.md"), out.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}
