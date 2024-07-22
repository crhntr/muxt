package source_test

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/template/parse"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crhntr/template/source"
)

func TestParseTree(t *testing.T) {
	t.Run("simple template", func(t *testing.T) {
		set := make(map[string]*parse.Tree)
		tree, err := source.CreateParseTree("simple", "Hello, World!", "{{", "}}", set)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Len(t, set, 1)
	})
	t.Run("simple with actions", func(t *testing.T) {
		set := make(map[string]*parse.Tree)
		tree, err := source.CreateParseTree("greeting", `(.Hello), (.World)!`, "(", ")", set)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Len(t, set, 1)
	})
	t.Run("with templates", func(t *testing.T) {
		set := make(map[string]*parse.Tree)
		tree, err := source.CreateParseTree("template", `(define "greeting")(.Hello), (.World)!(end -) outside defined template`, "(", ")", set)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Len(t, set, 2)
	})
	t.Run("with standard delimiters", func(t *testing.T) {
		set := make(map[string]*parse.Tree)
		tree, err := source.CreateParseTree("template", `{{.Greeting}}`, "", "", set)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Len(t, set, 1)
	})
}

func TestGoTypeComments(t *testing.T) {
	t.Run("template with templates", func(t *testing.T) {
		setup := sync.OnceValues(func() (map[string]*parse.Tree, string) {
			filePath := filepath.FromSlash("testdata/test_go_type_comments.gohtml")
			buf, err := os.ReadFile(filepath.FromSlash(filePath))
			if err != nil {
				log.Fatal(err)
			}
			set := make(map[string]*parse.Tree)
			_, err = source.CreateParseTree(path.Base(filePath), string(buf), "", "", set)
			if err != nil {
				log.Fatal(err)
			}
			return set, filePath
		})

		t.Run("with file containing definitions", func(t *testing.T) {
			set, filePath := setup()

			file, ok := set[filepath.Base(filePath)]
			require.True(t, ok)

			comments, err := source.FindTypeComments(filePath, file)
			require.NoError(t, err)
			require.Len(t, comments, 0)
		})
		t.Run("with no space after colon", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.NoError(t, err)
			require.Len(t, comments, 1)
		})
		t.Run("with space after colon", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.NoError(t, err)
			require.Len(t, comments, 1)
		})
		t.Run("with a standard library type", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.NoError(t, err)
			require.Len(t, comments, 1)
		})
		t.Run("with no gotype comment", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.NoError(t, err)
			require.Len(t, comments, 0)
		})
		t.Run("with wrong cased gotype comment", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.NoError(t, err)
			require.Len(t, comments, 1)
		})
		t.Run("with just the identifier", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.Error(t, err)
			require.Len(t, comments, 0)
		})
		t.Run("with just the package path", func(t *testing.T) {
			set, filePath := setup()
			tree := templateTree(t, set)

			comments, err := source.FindTypeComments(filePath, tree)
			require.Error(t, err)
			require.Len(t, comments, 0)
		})
	})
}

func TestResolveCommentTypes(t *testing.T) {
	setup := sync.OnceValue(func() []source.TypeComment {
		filePath := filepath.FromSlash("testdata/resolve_go_comment_types.gohtml")
		buf, err := os.ReadFile(filepath.FromSlash(filePath))
		if err != nil {
			log.Fatal(err)
		}
		set := make(map[string]*parse.Tree)
		_, err = source.CreateParseTree(path.Base(filePath), string(buf), "", "", set)
		if err != nil {
			log.Fatal(err)
		}
		var comments []source.TypeComment
		for _, tree := range set {
			results, err := source.FindTypeComments(filePath, tree)
			if err != nil {
				t.Fatal(err)
			}
			comments = append(comments, results...)
		}
		return comments
	})

	t.Run("with missing GOROOT", func(t *testing.T) {
		comments := setup()
		comments = []source.TypeComment{findComment(t, comments)}
		t.Setenv("GOROOT", "/tmp/missing-"+strconv.Itoa(int(time.Now().Unix())))
		err := source.ResolveCommentTypes(comments, nil)
		require.ErrorContains(t, err, `load packages failed`)
	})

	t.Run("with malformed package identifier", func(t *testing.T) {
		comments := setup()
		comments = []source.TypeComment{findComment(t, comments)}
		err := source.ResolveCommentTypes(comments, nil)
		require.ErrorContains(t, err, `malformed import path`)
	})

	t.Run("with no comments", func(t *testing.T) {
		err := source.ResolveCommentTypes(nil, nil)
		require.NoError(t, err)
	})

	t.Run("with no packages on comments", func(t *testing.T) {
		err := source.ResolveCommentTypes(make([]source.TypeComment, 5), nil)
		require.NoError(t, err)
	})

	t.Run("nil resolver", func(t *testing.T) {
		t.Run("with a findable type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.NoError(t, err)
		})

		t.Run("with a standard library type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.NoError(t, err)
		})
		t.Run("with an unknown package", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.ErrorContains(t, err, `load package "github.com/crhntr/template/internal/missing" failed`)
		})
		t.Run("with an unknown identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.ErrorContains(t, err, `lookup of MicroService failed in package`)
		})
		t.Run("with function identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.ErrorContains(t, err, `gotype comment error: unexpected kind for identifier Function (got func) in package`)
		})
		t.Run("with variable identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, nil)
			require.ErrorContains(t, err, `gotype comment error: unexpected kind for identifier Variable (got var) in package`)
		})
	})

	t.Run("return err resolver", func(t *testing.T) {
		t.Run("with a findable type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.NoError(t, err)
		})

		t.Run("with a standard library type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.NoError(t, err)
		})
		t.Run("with an unknown package", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.ErrorContains(t, err, `load package "github.com/crhntr/template/internal/missing" failed`)
		})
		t.Run("with an unknown identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.ErrorContains(t, err, `lookup of MicroService failed in package`)
		})
		t.Run("with function identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.ErrorContains(t, err, `gotype comment error: unexpected kind for identifier Function (got func) in package`)
		})
		t.Run("with variable identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return err
			})
			require.ErrorContains(t, err, `gotype comment error: unexpected kind for identifier Variable (got var) in package`)
		})
	})

	t.Run("return errors ignored", func(t *testing.T) {
		t.Run("with a findable type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})

		t.Run("with a standard library type", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})
		t.Run("with an unknown package", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})
		t.Run("with an unknown identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})
		t.Run("with function identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})
		t.Run("with variable identifier", func(t *testing.T) {
			comments := setup()
			comments = []source.TypeComment{findComment(t, comments)}
			err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
				return nil
			})
			require.NoError(t, err)
		})
	})

	t.Run("with resolver error", func(t *testing.T) {
		comments := setup()
		comments = []source.TypeComment{findComment(t, comments)}
		err := source.ResolveCommentTypes(comments, func(comment source.TypeComment, resolvedType types.Type, err error) error {
			return fmt.Errorf("banana")
		})
		require.ErrorContains(t, err, "banana")
	})
}

func templateTree(t *testing.T, set map[string]*parse.Tree) *parse.Tree {
	t.Helper()
	n := templateName(t)
	tree, ok := set[n]
	require.True(t, ok, "missing %q", n)
	return tree
}

func templateName(t *testing.T) string {
	return strings.Replace(path.Base(t.Name()), "_", " ", -1)
}

func findComment(t *testing.T, comments []source.TypeComment) source.TypeComment {
	t.Helper()
	for _, comment := range comments {
		if comment.Tree.Name == templateName(t) {
			return comment
		}
	}
	t.Fatalf("missing %q", templateName(t))
	return source.TypeComment{}
}
