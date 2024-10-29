package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

func commandTest(t *testing.T, pattern string) {
	e := script.NewEngine()
	e.Quiet = true
	e.Cmds = scripttest.DefaultCmds()
	e.Cmds["muxt"] = scriptCommand()
	testFiles, err := filepath.Glob(filepath.FromSlash(pattern))
	if err != nil {
		t.Fatal(err)
	}
	for _, filePath := range testFiles {
		name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			scripttest.Test(t, ctx, e, nil, filePath)
		})
	}
}
