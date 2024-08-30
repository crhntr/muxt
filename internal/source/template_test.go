package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
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
		require.ErrorContains(t, err, "run template templatesIdent failed at template.go:32:19: expected call expression")
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

	t.Run("call ParseFS with assets dir", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/assets_dir.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.Join(dir, filepath.FromSlash("assets/index.gohtml")),
			filepath.Join(dir, filepath.FromSlash("assets/form.gohtml")),
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
		require.ErrorContains(t, err, "template.go:20:19: expected template got UNKNOWN")
	})

	t.Run("call New with wrong arg count", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateWrongArgCount", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:22:38: expected exactly one string literal argument")
	})

	t.Run("call New on unexpected X", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewOnIndexed", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:24:25: expected exactly one argument ts[0] got 2")
	})

	t.Run("call New with non string literal arg", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewArg42", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:26:34: expected string literal got 42")
	})

	t.Run("call New with non literal arg", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewArgIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:28:37: expected string literal got TemplateName")
	})

	t.Run("call New with upstream error", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewErrUpstream", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateNewErrUpstream failed at template.go:30:40: expected string literal got fail")
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
		require.ErrorContains(t, err, "run template unsupportedMethod failed at template.go:34:22: unsupported function Unknown")
	})

	t.Run("call Must with unexpected function expression", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "unexpectedFunExpression", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template unexpectedFunExpression failed at template.go:36:28: unexpected expression *ast.IndexExpr: x[3]")
	})

	t.Run("call Must on non ident receiver", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateMustNonIdentReceiver", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateMustNonIdentReceiver failed at template.go:38:33: unexpected expression *ast.Ident: f")
	})
	t.Run("call Must with two arguments", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateMustCalledWithTwoArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateMustCalledWithTwoArgs failed at template.go:40:47: expected exactly one argument template got 2")
	})
	t.Run("call Must with one argument", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateMustCalledWithNoArg", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateMustCalledWithNoArg failed at template.go:42:47: expected exactly one argument template got 0")
	})
	t.Run("call Must wrong template package ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateMustWrongPackageIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateMustWrongPackageIdent failed at template.go:44:34: expected template got wrong")
	})
	t.Run("call ParseFS wrong template package ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSWrongPackageIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateParseFSWrongPackageIdent failed at template.go:46:37: expected template got wrong")
	})
	t.Run("call ParseFS receiver errored", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSReceiverErr", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateParseFSReceiverErr failed at template.go:48:43: expected exactly one string literal argument")
	})
	t.Run("call ParseFS unexpected receiver", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSUnexpectedReceiver", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "run template templateParseFSUnexpectedReceiver failed at template.go:50:38: expected exactly one argument x[0] got 2")
	})
	t.Run("call ParseFS with no arguments", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSNoArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:52:42: missing required arguments")
	})
	t.Run("call ParseFS with first arg non ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSFirstArgNonIdent", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:54:53: first argument to ParseFS must be an identifier")
	})
	t.Run("call ParseFS with first arg non ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSNonStringLiteralGlob", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, "template.go:56:78: expected string literal got 42")
	})
	t.Run("call ParseFS with bad glob", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSWithBadGlob", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:58:64: bad pattern "[fail": syntax error in pattern`)
	})
	t.Run("call ParseFS and fail to get relative template path", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/template_ParseFS.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.FromSlash("\x00/index.gohtml"), // null must not be in a path
		})
		require.ErrorContains(t, err, `failed to calculate relative path for embedded files: Rel: can't make`)
	})
	t.Run("call ParseFS and filter filepaths by globs", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/template_ParseFS.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		tsHTML, err := source.Templates(dir, "templatesHTML", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
			filepath.Join(dir, "script.html"),
		})
		require.NoError(t, err)
		tsGoHTML, err := source.Templates(dir, "templatesGoHTML", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
			filepath.Join(dir, "script.html"),
		})

		assert.NotNil(t, tsHTML.Lookup("script.html"))
		assert.NotNil(t, tsHTML.Lookup("console_log"))
		assert.Nil(t, tsGoHTML.Lookup("script.html"))
		assert.Nil(t, tsGoHTML.Lookup("console_log"))
	})
	t.Run("call bad embed pattern", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/bad_embed_pattern.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.Join(dir, "greeting.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:9:2: embed comment malformed: syntax error in pattern`)
	})
	t.Run("call bad embed pattern", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/template_ParseFS.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateEmbedVariableNotFound", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:22:65: variable hiding not found`)
	})
	t.Run("multiple delimiter types", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/delims.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		templates, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.Join(dir, "default.gohtml"),
			filepath.Join(dir, "triple_parens.gohtml"),
			filepath.Join(dir, "double_square.gohtml"),
		})
		require.NoError(t, err)
		var names []string
		for _, ts := range templates.Templates() {
			names = append(names, ts.Name())
		}
		assert.ElementsMatch(t, []string{"triple_parens.gohtml", "parens", "double_square.gohtml", "square", "", "default.gohtml", "default"}, names)
	})
	t.Run("Run method call gets no args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewHasWrongNumberOfArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:60:101: expected exactly one string literal argument`)
	})
	t.Run("Run method call gets wrong type of args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewHasWrongTypeOfArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:62:56: expected string literal got 9000`)
	})
	t.Run("Run method call gets too many args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateNewHasTooManyArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:64:51: expected exactly one string literal argument`)
	})
	t.Run("Delims method call gets no args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateDelimsGetsNoArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:66:53: expected exactly two string literal arguments`)
	})
	t.Run("Delims method call gets too many args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateDelimsGetsTooMany", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:68:54: expected exactly two string literal arguments`)
	})
	t.Run("Delims have wrong type of argument expressions", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateDelimsWrongExpressionArg", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:70:67: expected string literal got y`)
	})
	t.Run("ParseFS method fails", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateParseFSMethodFails", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:72:73: expected string literal got fail`)
	})
	t.Run("Options method requires string literals", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateOptionsRequiresStringLiterals", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:74:67: expected string literal got fail`)
	})
	t.Run("unknown method", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateUnknownMethod", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.ErrorContains(t, err, `template.go:76:26: unsupported method Unknown`)
	})
	t.Run("Option call", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templateOptionCall", fileSet, goFiles, []string{
			filepath.Join(dir, "index.gohtml"),
		})
		require.NoError(t, err)
	})
	t.Run("Option call wrong argument", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/templates.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		assert.Panics(t, func() {
			_, _ = source.Templates(dir, "templateOptionCallUnknownArg", fileSet, goFiles, []string{
				filepath.Join(dir, "index.gohtml"),
			})
		})
	})
	t.Run("Funcs call", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templates", fileSet, goFiles, []string{
			filepath.Join(dir, "greet.gohtml"),
		})
		require.NoError(t, err)
	})
	t.Run("Func not defined", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesFuncNotDefined", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `missing_func.gohtml:1: function "enemy" not defined`)
	})
	t.Run("Func wrong parameter kind", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongArg", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected a composite literal with type template.FuncMap got wrong`)
	})

	t.Run("Func wrong too many args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesTwoArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected exactly 1 template.FuncMap composite literal argument`)
	})
	t.Run("Func wrong too no args", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesNoArgs", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected exactly 1 template.FuncMap composite literal argument`)
	})
	t.Run("Func wrong package ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongTypePackageName", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected a composite literal with type template.FuncMap got wrong.FuncMap{}`)
	})
	t.Run("Func wrong Type ident", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongTypeName", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected a composite literal with type template.FuncMap got template.Wrong{}`)
	})
	t.Run("Func wrong Type", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongTypeExpression", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected a composite literal with type template.FuncMap got wrong{}`)
	})
	t.Run("Func wrong elem", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongTypeElem", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected element at index 0 to be a key value pair got wrong`)
	})
	t.Run("Func wrong elem key", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/funcs.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "templatesWrongElemKey", fileSet, goFiles, []string{
			filepath.Join(dir, "missing_func.gohtml"),
			filepath.Join(dir, "greet.gohtml"),
		})
		require.ErrorContains(t, err, `expected string literal got wrong`)
	})
	t.Run("Parse template name from new", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/parse.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "templates", fileSet, goFiles, nil)
		require.NoError(t, err)
		assert.NotNil(t, ts.Lookup("GET /"))
	})
	t.Run("Parse string has multiple routes", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/parse.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		ts, err := source.Templates(dir, "multiple", fileSet, goFiles, nil)
		require.NoError(t, err)
		assert.NotNil(t, ts.Lookup("GET /"))
		assert.NotNil(t, ts.Lookup("GET /{name}"))
	})
	t.Run("Parse is missing argument", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/parse.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "noArg", fileSet, goFiles, nil)
		require.ErrorContains(t, err, "run template noArg failed at parse.go:12:35: expected exactly one string literal argument")
	})
	t.Run("Parse gets wrong argument type", func(t *testing.T) {
		dir := createTestDir(t, filepath.FromSlash("testdata/parse.txtar"))
		goFiles, fileSet := parseGo(t, dir)
		_, err := source.Templates(dir, "wrongArg", fileSet, goFiles, nil)
		require.ErrorContains(t, err, "run template wrongArg failed at parse.go:14:40: expected string literal got 500")
	})
}

func createTestDir(t *testing.T, filename string) string {
	t.Helper()
	dir := t.TempDir()
	archive, err := txtar.ParseFile(filepath.FromSlash(filename))
	if err != nil {
		t.Fatal(err)
	}
	tfs, err := txtar.FS(archive)
	if err := os.CopyFS(dir, tfs); err != nil {
		log.Fatal(err)
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
