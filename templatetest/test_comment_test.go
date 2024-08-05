package templatetest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/crhntr/muxt/internal/fake"
	"github.com/crhntr/muxt/templatetest"
)

//go:generate counterfeiter -generate

//counterfeiter:generate --fake-name T -o ../internal/fake/t.go . testingT

type testingT interface {
	assert.TestingT
	Logf(string, ...interface{})
	Helper()
}

func TestAssertTypeCommentsAreFound(t *testing.T) {
	t.Run("when glob fails", func(t *testing.T) {
		ft := new(fake.T)

		pass := templatetest.AssertTypeCommentsAreFound(ft, "", "", "[]a]")
		assert.False(t, pass)
		assert.NotZero(t, ft.HelperCallCount())
	})
	t.Run("when the template file is not well formed", func(t *testing.T) {
		// this was hard to figure out, I was able to do it by just reading the directory
		ft := new(fake.T)

		pass := templatetest.AssertTypeCommentsAreFound(ft, "", "", filepath.FromSlash("testdata/malformed.gohtml"))
		assert.False(t, pass)
		assert.NotZero(t, ft.HelperCallCount())
	})
}
