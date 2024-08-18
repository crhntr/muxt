package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"

	"github.com/crhntr/muxt/internal/source"
)

func TestTemplates(t *testing.T) {
	t.Run("non call", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templatesIdent failed at template.go:32:20: expected call expression")
	})

	t.Run("call ParseFS", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/template_ParseFS.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
			filepath.Join(dir, "form.gohtml"),
		})
		require.NoError(t, err)
		var names []string
		for _, t := range ts.Templates() {
			names = append(names, t.Name())
		}
		slices.Sort(names)
		assert.Equal(t, []string{"create", "form.gohtml", "home", "index.gohtml", "update"}, names)
	})

	t.Run("call New", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templateNew", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.NoError(t, err)
		var names []string
		for _, t := range ts.Templates() {
			names = append(names, t.Name())
		}
		slices.Sort(names)
		assert.Equal(t, []string{"some-name"}, names)
	})

	t.Run("call New after calling ParseFS", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templateParseFSNew", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.NoError(t, err)
		var names []string
		for _, t := range ts.Templates() {
			names = append(names, t.Name())
		}
		slices.Sort(names)
		assert.Equal(t, []string{"greetings", "index.gohtml"}, names)
	})

	t.Run("call New before calling ParseFS", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templateNewParseFS", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.NoError(t, err)
		var names []string
		for _, t := range ts.Templates() {
			names = append(names, t.Name())
		}
		slices.Sort(names)
		assert.Equal(t, []string{"greetings", "index.gohtml"}, names)
	})

	t.Run("call new with non args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewMissingArg", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "expected exactly one string literal argument")
	})

	t.Run("call New on unknown X", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateWrongX", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:20:31: expected template got UNKNOWN")
	})

	t.Run("call New with wrong arg count", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateWrongArgCount", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:22:39: expected exactly one string literal argument")
	})

	t.Run("call New on unexpected X", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewOnIndexed", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:24:26: expected New to either be a call of function New from package template package or a call to method New on *template.Template")
	})

	t.Run("call New with non string literal arg", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewArg42", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:26:35: expected argument to be a string literal got 42")
	})

	t.Run("call New with non literal arg", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewArgIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:28:38: expected argument to be a string literal got TemplateName")
	})

	t.Run("call New with upstream error", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewErrUpstream", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateNewErrUpstream failed at template.go:30:41: expected argument to be a string literal got fail")
	})

	t.Run("unknown templates variable", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "variableDoesNotExist", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.NotNil(t, err)
		require.Equal(t, "variable variableDoesNotExist not found", err.Error())
	})

	t.Run("unknown templates variable", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "unsupportedMethod", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template unsupportedMethod failed at template.go:34:23: unsupported method Unknown")
	})

	t.Run("unexpected function expression", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "unexpectedFunExpression", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template unexpectedFunExpression failed at template.go:36:29: unexpected call: x[3]")
	})
}

func createTestDir(t *testing.T, filename string) string {
	t.Helper()
	dir := t.TempDir()
	archive, err := txtar.ParseFile(filepath.FromSlash(filename))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range archive.Files {
		output := filepath.Join(dir, filepath.FromSlash(file.Name))
		if err := os.WriteFile(output, file.Data, 0o666); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func parseGo(t *testing.T, dir string) ([]*ast.File, *token.FileSet) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	require.NoError(t, err)
	set := token.NewFileSet()
	var files []*ast.File
	for _, match := range matches {
		file, err := parser.ParseFile(set, match, nil, parser.ParseComments|parser.AllErrors|parser.SkipObjectResolution)
		require.NoError(t, err)
		files = append(files, file)
	}
	return files, set
}
