package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"golang.org/x/tools/txtar"
	"rsc.io/script"
)

//go:embed data/new/*.txtar
var projectTemplates embed.FS

func newCommand(args []string, workingDirectory string, _ func(string) string, stdout, stderr io.Writer) error {
	templateFilePaths, err := fs.Glob(projectTemplates, "data/new/*.txtar")
	if err != nil {
		return fmt.Errorf("failed to load new project templates")
	}
	var newProjectTemplateNames []string
	for _, filePath := range templateFilePaths {
		name := strings.TrimSuffix(path.Base(filePath), path.Ext(filePath))
		newProjectTemplateNames = append(newProjectTemplateNames, name)
	}
	var (
		templateName string
	)
	flagSet := flag.NewFlagSet("new", flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	flagSet.StringVar(&templateName, "template", "main", fmt.Sprintf("new project template name one of: [%s]", strings.Join(newProjectTemplateNames, ", ")))
	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("failed to parse arguments for new command: %w", err)
	}

	i := slices.Index(newProjectTemplateNames, templateName)
	if i < 0 {
		return fmt.Errorf("unknown new project tamplate name: %q", templateName)
	}
	selectedTemplateName := templateFilePaths[i]

	buf, err := fs.ReadFile(projectTemplates, selectedTemplateName)
	if err != nil {
		return fmt.Errorf("failed to read new project template: %w", err)
	}

	archive := txtar.Parse(buf)
	dir, err := txtar.FS(archive)
	if err != nil {
		return fmt.Errorf("failed to use new project template as directory: %w", err)
	}
	if err := os.CopyFS(workingDirectory, dir); err != nil {
		return fmt.Errorf("failed to copy new project template files to output directory %q: %w", workingDirectory, err)
	}

	eng := script.NewEngine()
	eng.Cmds["muxt"] = scriptCommand()
	ctx := context.Background()
	state, err := script.NewState(ctx, workingDirectory, os.Environ())
	if err != nil {
		return fmt.Errorf("failed to setup project template script: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "running script:\n")
	_, _ = stdout.Write(archive.Comment)
	_, _ = io.WriteString(stdout, "\n")
	if err := eng.Execute(state, templateName, bufio.NewReader(bytes.NewReader(archive.Comment)), stderr); err != nil {
		return fmt.Errorf("failed while running project template script: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "new project generated\nnow run:\n\n\tgo generate")
	return err
}
