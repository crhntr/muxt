package main

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	commandTest(t, "testdata/generate/*.txtar")
}
