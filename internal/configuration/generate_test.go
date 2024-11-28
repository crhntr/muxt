package configuration

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerate(t *testing.T) {
	t.Run("unknown flag", func(t *testing.T) {
		_, err := NewGenerate([]string{
			"--unknown",
		}, func(s string) string { return "" }, io.Discard)
		assert.ErrorContains(t, err, "flag provided but not defined")
	})
	t.Run(receiverStaticType+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := NewGenerate([]string{
			"--" + receiverStaticType, "123",
		}, func(s string) string { return "" }, io.Discard)
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(routesFunc+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := NewGenerate([]string{
			"--" + routesFunc, "123",
		}, func(s string) string { return "" }, io.Discard)
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(templatesVariable+" flag value is an invalid identifier", func(t *testing.T) {
		_, err := NewGenerate([]string{
			"--" + templatesVariable, "123",
		}, func(s string) string { return "" }, io.Discard)
		assert.ErrorContains(t, err, errIdentSuffix)
	})
	t.Run(outputFlagName+" flag value is not a go file", func(t *testing.T) {
		_, err := NewGenerate([]string{
			"--" + outputFlagName, "output.txt",
		}, func(s string) string { return "" }, io.Discard)
		assert.ErrorContains(t, err, "filename must use .go extension")
	})
}
