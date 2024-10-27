package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	commandTest(t, "testdata/generate/*.txtar")
}

func Test_newGenerate(t *testing.T) {
	t.Run("unknown flag", func(t *testing.T) {
		_, err := newGenerate([]string{
			"--unknown",
		}, func(s string) string { return "" })
		assert.ErrorContains(t, err, "flag provided but not defined")
	})
	t.Run(receiverStaticType+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := newGenerate([]string{
			"--" + receiverStaticType, "123",
		}, func(s string) string { return "" })
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(routesFunc+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := newGenerate([]string{
			"--" + routesFunc, "123",
		}, func(s string) string { return "" })
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(templatesVariable+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := newGenerate([]string{
			"--" + templatesVariable, "123",
		}, func(s string) string { return "" })
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(outputFlagFlagName+" flag value is not a go file", func(t *testing.T) {
		_, err := newGenerate([]string{
			"--" + outputFlagFlagName, "output.txt",
		}, func(s string) string { return "" })
		assert.ErrorContains(t, err, "filename must use .go extension")
	})
}
