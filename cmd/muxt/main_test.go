package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func Test_example(t *testing.T) {
	require.NoError(t, os.Remove(filepath.FromSlash("../../example/template_routes.go")))

	ctx := context.TODO()
	cmd := exec.CommandContext(ctx, "go", "generate")
	cmd.Dir = filepath.FromSlash("../../example")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	require.NoError(t, cmd.Run())
}
