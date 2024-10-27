package main

import "testing"

func TestNew(t *testing.T) {
	commandTest(t, "testdata/new/*.txtar")
}
