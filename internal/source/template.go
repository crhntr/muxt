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
		ts, err := parseTemplates(workingDirectory, templatesVariable, tn, fileSet, files, tv.Values[i], embeddedAbsolutePath)
		if err != nil {
			return nil, fmt.Errorf("run template %s failed at %w", templatesVariable, err)
		}
		return ts, nil
	}
	return nil, fmt.Errorf("variable %s not found", templatesVariable)
}

func parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent string, fileSet *token.FileSet, files []*ast.File, expression ast.Expr, embeddedAbsolutePath []string) (*template.Template, error) {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return nil, contextError(workingDirectory, fileSet, expression.Pos(), fmt.Errorf("expected call expression"))
	}

	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		switch sel.Sel.Name {
		case "New":
			var templatesNew func(string) *template.Template
			switch x := sel.X.(type) {
			case *ast.Ident:
				if pkg := templatesPackageIdent; x.Name != pkg {
					return nil, contextError(workingDirectory, fileSet, call.Lparen, fmt.Errorf("expected %s got %s", pkg, x.Name))
				}
				templatesNew = template.New
			case *ast.CallExpr:
				ts, err := parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, x, embeddedAbsolutePath)
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
				return nil, fmt.Errorf("expected %s.%s", templatesPackageIdent, sel.Sel.Name)
			}
			if x.Name != "template" {
				return nil, fmt.Errorf("expected %s.%s", templatesPackageIdent, sel.Sel.Name)
			}
			if len(call.Args) != 1 {
				return nil, fmt.Errorf("expected %s.%s", templatesPackageIdent, sel.Sel.Name)
			}
			return parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, call.Args[0], embeddedAbsolutePath)
		case "ParseFS":
			var parseFiles func(files ...string) (*template.Template, error)
			switch x := sel.X.(type) {
			case *ast.Ident:
				if x.Name != templatesPackageIdent {
					return nil, fmt.Errorf("expected %s.%s", templatesPackageIdent, sel.Sel.Name)
				}
				parseFiles = template.ParseFiles
			case *ast.CallExpr:
				ts, err := parseTemplates(workingDirectory, templatesVariable, templatesPackageIdent, fileSet, files, x, embeddedAbsolutePath)
				if err != nil {
					return nil, err
				}
				parseFiles = ts.ParseFiles
			default:
				return nil, fmt.Errorf("expected %s.%s", templatesPackageIdent, sel.Sel.Name)
			}
			if len(call.Args) < 1 {
				return nil, fmt.Errorf("%s.%s is missing required fs.FS argument", templatesPackageIdent, sel.Sel.Name)
			}
			fsIdent, ok := call.Args[0].(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("%s.%s expected a variable with type embed.FS as the first argument", templatesPackageIdent, sel.Sel.Name)
			}
			embeddedFiles, err := embedFSFilepaths(workingDirectory, files, fsIdent, embeddedAbsolutePath)
			if err != nil {
				return nil, err
			}
			globs := make([]string, 0, len(call.Args[1:]))
			for _, a := range call.Args[1:] {
				switch arg := a.(type) {
				case *ast.BasicLit:
					if arg.Kind != token.STRING {
						return nil, fmt.Errorf("expected string literal")
					}
					value, err := strconv.Unquote(arg.Value)
					if err != nil {
						return nil, err
					}
					globs = append(globs, value)
				}
			}
			filtered := embeddedFiles[:0]
			for _, ef := range embeddedFiles {
				rel, err := filepath.Rel(workingDirectory, ef)
				if err != nil {
					return nil, err
				}
				for _, pattern := range globs {
					match, err := filepath.Match(pattern, rel)
					if err != nil || !match {
						continue
					}
					filtered = append(filtered, ef)
					break
				}
			}
			embeddedFiles = slices.Clip(filtered)
			return parseFiles(embeddedFiles...)
		}
	}
	return nil, fmt.Errorf("no templates found")
}

func embedFSFilepaths(dir string, files []*ast.File, fsIdent *ast.Ident, embeddedFiles []string) ([]string, error) {
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
			efs, err := embeddedFilesMatchingPatternList(dir, patterns, embeddedFiles)
			if err != nil {
				return nil, err
			}
			return efs, nil
		}
	}
	return nil, fmt.Errorf("variable %s not found", fsIdent.Name)
}

func embeddedFilesMatchingPatternList(dir string, patterns, embeddedFiles []string) ([]string, error) {
	var matches []string
	for _, fp := range embeddedFiles {
		rel, err := filepath.Rel(dir, fp)
		if err != nil {
			return nil, err
		}
		for _, pattern := range patterns {
			pat := filepath.FromSlash(pattern)
			fullPat := filepath.Join(dir, filepath.FromSlash(pat)) + "/"
			if i := slices.IndexFunc(embeddedFiles, func(file string) bool {
				return strings.HasPrefix(file, fullPat)
			}); i >= 0 {
				matches = append(matches, embeddedFiles[i])
				continue
			}
			if matched, err := filepath.Match(pat, rel); err != nil {
				return nil, err
			} else if matched {
				matches = append(matches, fp)
			}
		}
	}
	return matches, nil
}

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
