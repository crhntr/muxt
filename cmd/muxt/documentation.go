package main

import (
	"io"

	"github.com/crhntr/muxt"
	"github.com/crhntr/muxt/internal/configuration"
)

func documentationCommand(wd string, args []string, stdout, stderr io.Writer) error {
	config, err := configuration.NewRoutesFileConfiguration(args, stderr)
	if err != nil {
		return err
	}
	return muxt.Documentation(stdout, wd, config)
}
