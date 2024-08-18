package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"html/template"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func Templates(workingDirectory, templatesVariable string, fileSet *token.FileSet, files []*ast.File, embeddedAbsolutePath []string) (*template.Template, error) {
	for _, tv := range IterateValueSpecs(files) {
		i := slices.IndexFunc(tv.Names, func(e *ast.Ident) bool {
			return e.Name == templatesVariable
		})
		if i < 0 || i >= len(tv.Values) {
			continue
		}
		tn := "template"
		for _, im := range IterateImports(files) {
			if im.Path.Kind != token.STRING || im.Path.Value != `"html/template"` || im.Name == nil || im.Name.Name == "" {
				continue
			}
			tn = im.Name.Name
			break
		}
		embeddedPaths, err := relFilepaths(workingDirectory, embeddedAbsolutePath...)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate relative path for embedded files: %w", err)
		}
		ts, err := parseTemplates(workingDirectory, templatesVariable, tn, fileSet, files, tv.Values[i], embeddedPaths)
		if err != nil {
			return nil, fmt.Errorf("run template %s failed at %w", templatesVariable, err)
		}
		return ts, nil
	}
	return nil, fmt.Errorf("variable %s not found", templatesVariable)
}

func parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent string, fileSet *token.FileSet, files []*ast.File, expression ast.Expr, embeddedPaths []string) (*template.Template, error) {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return nil, contextError(workingDirectory, fileSet, expression.Pos(), fmt.Errorf("expected call expression"))
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, contextError(workingDirectory, fileSet, call.Fun.Pos(), fmt.Errorf("unexpected call: %s", Format(call.Fun)))
	}
	switch sel.Sel.Name {
	default:
		return nil, contextError(workingDirectory, fileSet, call.Fun.Pos(), fmt.Errorf("unsupported method %s", sel.Sel.Name))
	case "New":
		var templatesNew func(string) *template.Template
		switch x := sel.X.(type) {
		case *ast.Ident:
			if pkg := templatesPackageIdent; x.Name != pkg {
				return nil, contextError(workingDirectory, fileSet, call.Lparen, fmt.Errorf("expected %s got %s", pkg, x.Name))
			}
			templatesNew = template.New
		case *ast.CallExpr:
			ts, err := parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, x, embeddedPaths)
			if err != nil {
				return nil, err
			}
			templatesNew = ts.New
		default:
			return nil, contextError(workingDirectory, fileSet, sel.X.Pos(), fmt.Errorf("expected New to either be a call of function New from package template package or a call to method New on *template.Template"))
		}

		if len(call.Args) != 1 {
			return nil, contextError(workingDirectory, fileSet, call.Lparen, fmt.Errorf("expected exactly one string literal argument"))
		}

		switch arg := call.Args[0].(type) {
		case *ast.BasicLit:
			if arg.Kind != token.STRING {
				return nil, contextError(workingDirectory, fileSet, arg.Pos(), fmt.Errorf("expected argument to be a string literal got %s", Format(arg)))
			}
			name, _ := strconv.Unquote(arg.Value)
			return templatesNew(name), nil
		default:
			return nil, contextError(workingDirectory, fileSet, arg.Pos(), fmt.Errorf("expected argument to be a string literal got %s", Format(arg)))
		}
	case "Must":
		x, ok := sel.X.(*ast.Ident)
		if !ok {
			return nil, contextError(workingDirectory, fileSet, sel.X.Pos(), fmt.Errorf("expected package identifier %s got %s", templatesPackageIdent, Format(sel.X)))
		}
		if x.Name != templatesPackageIdent {
			return nil, contextError(workingDirectory, fileSet, sel.X.Pos(), fmt.Errorf("expected package identifier %s got %s", templatesPackageIdent, Format(sel.X)))
		}
		if len(call.Args) != 1 {
			return nil, contextError(workingDirectory, fileSet, call.Lparen, fmt.Errorf("expected exactly one argument %s got %d", Format(sel.X), len(call.Args)))
		}
		return parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, call.Args[0], embeddedPaths)
	case "ParseFS":
		var parseFiles func(files ...string) (*template.Template, error)
		switch x := sel.X.(type) {
		case *ast.Ident:
			if x.Name != templatesPackageIdent {
				return nil, contextError(workingDirectory, fileSet, sel.X.Pos(), fmt.Errorf("expected package identifier %s got %s", templatesPackageIdent, Format(sel.X)))
			}
			parseFiles = template.ParseFiles
		case *ast.CallExpr:
			ts, err := parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, x, embeddedPaths)
			if err != nil {
				return nil, err
			}
			parseFiles = ts.ParseFiles
		default:
			return nil, contextError(workingDirectory, fileSet, sel.X.Pos(), fmt.Errorf("unexpected method receiver %s", Format(sel.X)))
		}
		if len(call.Args) < 1 {
			return nil, contextError(workingDirectory, fileSet, call.Lparen, fmt.Errorf("missing required arguments"))
		}
		matches, err := embedFSFilepaths(workingDirectory, fileSet, files, call.Args[0], embeddedPaths)
		if err != nil {
			return nil, err
		}
		patterns, err := parseStringLiterals(workingDirectory, fileSet, call.Args[1:])
		if err != nil {
			return nil, err
		}
		filtered := matches[:0]
		for _, ef := range matches {
			for j, pattern := range patterns {
				match, err := filepath.Match(pattern, ef)
				if err != nil {
					return nil, contextError(workingDirectory, fileSet, call.Args[j+1].Pos(), fmt.Errorf("bad pattern %q: %w", pattern, err))
				}
				if !match {
					continue
				}
				filtered = append(filtered, ef)
				break
			}
		}
		return parseFiles(joinFilepaths(workingDirectory, filtered...)...)
	}
}

func parseStringLiterals(wd string, set *token.FileSet, list []ast.Expr) ([]string, error) {
	result := make([]string, 0, len(list))
	for _, a := range list {
		switch arg := a.(type) {
		case *ast.BasicLit:
			if arg.Kind != token.STRING {
				return nil, contextError(wd, set, arg.Pos(), fmt.Errorf("expected string literal got %s", Format(arg)))
			}
			s, _ := strconv.Unquote(arg.Value)
			result = append(result, s)
		}
	}
	return result, nil
}

func embedFSFilepaths(dir string, fileSet *token.FileSet, files []*ast.File, exp ast.Expr, embeddedFiles []string) ([]string, error) {
	fsIdent, ok := exp.(*ast.Ident)
	if !ok {
		return nil, contextError(dir, fileSet, exp.Pos(), fmt.Errorf("first argument to ParseFS must be an identifier"))
	}
	for _, decl := range IterateGenDecl(files, token.VAR) {
		for _, s := range decl.Specs {
			spec, ok := s.(*ast.ValueSpec)
			if !ok || !slices.ContainsFunc(spec.Names, func(e *ast.Ident) bool { return e.Name == fsIdent.Name }) {
				continue
			}
			var comment strings.Builder
			readComments(&comment, decl.Doc, spec.Doc)
			patterns, err := parsePatterns(comment.String())
			if err != nil {
				return nil, err
			}
			absMat, err := embeddedFilesMatchingPatternList(patterns, embeddedFiles)
			if err != nil {
				return nil, err
			}
			return absMat, nil
		}
	}
	return nil, fmt.Errorf("variable %s not found", fsIdent.Name)
}

func embeddedFilesMatchingPatternList(patterns, embeddedFiles []string) ([]string, error) {
	var matches []string
	for _, fp := range embeddedFiles {
		for _, pattern := range patterns {
			pat := filepath.FromSlash(pattern)
			if !strings.ContainsAny(pat, "*[]") {
				prefix := filepath.FromSlash(pat) + "/"
				if i := slices.IndexFunc(embeddedFiles, func(file string) bool {
					return strings.HasPrefix(file, prefix)
				}); i >= 0 {
					matches = append(matches, embeddedFiles[i])
					continue
				}
			}
			if matched, err := filepath.Match(pat, fp); err != nil {
				return nil, err
			} else if matched {
				matches = append(matches, fp)
			}
		}
	}
	return slices.Clip(matches), nil
}

const goEmbedCommentPrefix = "//go:embed"

func readComments(s *strings.Builder, groups ...*ast.CommentGroup) {
	for _, c := range groups {
		if c == nil {
			continue
		}
		for _, line := range c.List {
			if !strings.HasPrefix(line.Text, goEmbedCommentPrefix) {
				continue
			}
			s.WriteString(strings.TrimSpace(strings.TrimPrefix(line.Text, goEmbedCommentPrefix)))
			s.WriteByte(' ')
		}
	}
}

func parsePatterns(input string) ([]string, error) {
	// todo: refactor to use strconv.QuotedPrefix
	var (
		patterns       []string
		currentPattern strings.Builder
		inQuote        = false
		quoteChar      rune
	)

	for _, r := range input {
		switch {
		case r == '"' || r == '`':
			if !inQuote {
				inQuote = true
				quoteChar = r
				continue
			}
			if r != quoteChar {
				currentPattern.WriteRune(r)
				continue
			}
			patterns = append(patterns, currentPattern.String())
			currentPattern.Reset()
			inQuote = false
		case unicode.IsSpace(r):
			if inQuote {
				currentPattern.WriteRune(r)
				continue
			}
			if currentPattern.Len() > 0 {
				patterns = append(patterns, currentPattern.String())
				currentPattern.Reset()
			}
		default:
			currentPattern.WriteRune(r)
		}
	}

	// Add any remaining pattern
	if currentPattern.Len() > 0 {
		patterns = append(patterns, currentPattern.String())
	}

	return patterns, nil
}

func contextError(workingDirectory string, set *token.FileSet, pos token.Pos, err error) error {
	p := set.Position(pos)
	p.Filename, _ = filepath.Rel(workingDirectory, p.Filename)
	return fmt.Errorf("%s: %w", p, err)
}

func joinFilepaths(wd string, rel ...string) []string {
	result := slices.Clone(rel)
	for i := range result {
		result[i] = filepath.Join(wd, result[i])
	}
	return result
}

func relFilepaths(wd string, abs ...string) ([]string, error) {
	result := slices.Clone(abs)
	for i, p := range result {
		r, err := filepath.Rel(wd, p)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}
