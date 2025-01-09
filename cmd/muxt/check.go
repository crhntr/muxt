package main

import (
	"fmt"
	"io"

	"github.com/crhntr/muxt"
	"github.com/crhntr/muxt/internal/configuration"
)

func checkCommand(workingDirectory string, args []string, stderr io.Writer) error {
	config, err := configuration.NewRoutesFileConfiguration(args, stderr)
	if err != nil {
		return err
	}
	if err := muxt.CheckTemplates(workingDirectory, config); err != nil {
		return fmt.Errorf("fail")
	}
	return nil
}
