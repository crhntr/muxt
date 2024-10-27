package main

import (
	"fmt"
	"io"
)

func newCommand(_ []string, _ string, _ func(string) string, stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, "Coming soon...")
	return err
}
