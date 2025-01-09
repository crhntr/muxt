package main

import (
	"io"

	"github.com/crhntr/muxt"
	"github.com/crhntr/muxt/internal/configuration"
)

func checkCommand(workingDirectory string, args []string, stderr io.Writer) error {
	config, err := configuration.NewRoutesFileConfiguration(args, stderr)
	if err != nil {
		return err
	}
	return muxt.CheckTemplates(workingDirectory, config)
}
