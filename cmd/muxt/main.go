package main

import (
	"fmt"
	"io"
	"os"

	"github.com/crhntr/muxt/internal/configuration"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(handleError(err))
	}
	os.Exit(handleError(command(wd, os.Args[1:], os.Getenv, os.Stdout, os.Stderr)))
}

func command(wd string, args []string, getEnv func(string) string, stdout, stderr io.Writer) error {
	var err error
	wd, args, err = configuration.Global(wd, args, stderr)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		switch cmd, cmdArgs := args[0], args[1:]; cmd {
		case "generate", "gen", "g":
			return generateCommand(wd, cmdArgs, getEnv, stdout, stderr)
		case "version", "v":
			return versionCommand(stdout)
		case "check", "c":
			return checkCommand(wd, cmdArgs, stderr)
		case "documentation", "docs", "d":
			return documentationCommand(wd, cmdArgs, stdout, stderr)
		default:
			return fmt.Errorf("unknown command")
		}
	}

	_, _ = fmt.Fprintf(stdout, `muxt - Generate HTTP Endpoints from HTML Templates

muxt check

	Do some static analysis on the templates. 

muxt documentation

	This work in progress command will 

muxt generate

	Use this command to generate template_routes.go
	
	Consider using a Go generate comment where your templates variable is declared.

	  //go:generate muxt generate %s=Server
      var templates = templates = template.Must(template.ParseFS(templatesSource, "*.gohtml"))

muxt version

	Print the version of muxt to standard out.

`, configuration.ReceiverStaticType)

	return fmt.Errorf("no arguments provided")
}

func handleError(err error) int {
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return 1
	}
	return 0
}
