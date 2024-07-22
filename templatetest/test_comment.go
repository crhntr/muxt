package templatetest

import (
	"errors"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"text/template/parse"

	"github.com/stretchr/testify/assert"

	"github.com/crhntr/template/source"
)

func AssertTypeCommentsAreFound(t assert.TestingT, leftDelim, rightDelim string, patterns ...string) bool {
	if h, ok := t.(interface {
		Helper()
	}); ok {
		h.Helper()
	}
	var filePaths []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if !assert.NoError(t, err) {
			return false
		}
		filePaths = append(filePaths, matches...)
	}
	slices.Sort(filePaths)
	filePaths = slices.Compact(filePaths)

	var comments []source.TypeComment
	set := make(map[string]*parse.Tree)
	for _, filePath := range filePaths {
		buf, err := os.ReadFile(filePath)
		if !assert.NoError(t, err) {
			return false
		}
		_, err = source.CreateParseTree(filepath.Base(filePath), string(buf), leftDelim, rightDelim, set)
		if !assert.NoError(t, err) {
			return false
		}
		for _, tree := range set {
			results, err := source.FindTypeComments(filePath, tree)
			if !assert.NoError(t, err) {
				return false
			}
			comments = append(comments, results...)
		}
	}

	var list []error
	err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
		if err != nil {
			list = append(list, err)
		}
		if l, ok := t.(interface {
			Logf(format string, args ...interface{})
		}); ok {
			if assert.NoError(t, err) && testing.Verbose() {
				l.Logf("%q: %s", comment.Tree.Name, resolvedType.String())
			}
		}
		return nil
	})
	return assert.NoError(t, errors.Join(append(list, err)...))
}
