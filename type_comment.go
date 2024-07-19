package templatesource

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"
	"text/template/parse"

	"golang.org/x/tools/go/packages"
)

// TypeComment represents a gotype comment in a template.
// These comments look something like this: {{- /*gotype: example.com/package-name.TypeName*/ -}}
type TypeComment struct {
	Package    string
	Identifier string

	Filepath    string
	CommentNode *parse.CommentNode
	Tree        *parse.Tree
}

func CreateParseTree(templateName, templateSource, leftDelim, rightDelim string, treeSet map[string]*parse.Tree) (*parse.Tree, error) {
	tr := parse.New(templateName)
	tr.Mode = parse.ParseComments | parse.SkipFuncCheck
	return tr.Parse(templateSource, leftDelim, rightDelim, treeSet)
}

func FindTypeComments(filePath string, tree *parse.Tree) ([]TypeComment, error) {
	var (
		comments []TypeComment
		err      error
	)
	for _, n := range tree.Root.Nodes {
		cn, ok := n.(*parse.CommentNode)
		if !ok {
			continue
		}
		text := strings.TrimSpace(cn.Text)
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, "gotype:") {
			continue
		}
		text = strings.TrimPrefix(text, "gotype:")
		text = strings.TrimSpace(text)

		dotIndex := strings.LastIndexByte(text, '.')
		if dotIndex < 0 {
			err = errors.Join(err, fmt.Errorf("malformed gotype comment: %q", text))
			continue
		}
		ident := strings.TrimSpace(text[dotIndex+1:])
		if !token.IsIdentifier(ident) {
			err = errors.Join(err, fmt.Errorf("malformed gotype comment identifier %q: %w", ident, err))
			continue
		}

		comments = append(comments, TypeComment{
			Package:     text[:dotIndex],
			Identifier:  ident,
			Filepath:    filePath,
			CommentNode: cn,
			Tree:        tree,
		})
	}
	return comments, err
}

func ResolveCommentTypes(comments []TypeComment, resolved func(comment TypeComment, resolvedType types.Type, err error) error) error {
	var packageIDs []string
	for _, comment := range comments {
		if comment.Package == "" {
			continue
		}
		packageIDs = append(packageIDs, comment.Package)
	}
	slices.Sort(packageIDs)
	packageIDs = slices.Compact(packageIDs)

	if len(packageIDs) == 0 {
		return nil
	}

	loadedPackages, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedFiles,
	}, packageIDs...)
	if err != nil {
		return fmt.Errorf("gotype comment error: load packages failed: %w", err)
	}

	for _, comment := range comments {
		i := slices.IndexFunc(loadedPackages, func(pkg *packages.Package) bool { return pkg.ID == comment.Package })
		pkg := loadedPackages[i]

		if len(pkg.Errors) != 0 {
			err := fmt.Errorf("template %q gotype comment error: load package %q failed: %w", comment.Tree.Name, comment.Package, errors.New(pkg.Errors[0].Error()))
			if resolved != nil {
				if err := resolved(comment, nil, err); err != nil {
					return err
				}
				continue
			} else {
				return err
			}
		}

		obj := pkg.Types.Scope().Lookup(comment.Identifier)
		if obj == nil {
			err := fmt.Errorf("template %q gotype comment error: lookup of %s failed in package %q", comment.Tree.Name, comment.Identifier, comment.Package)
			if resolved != nil {
				if err := resolved(comment, nil, err); err != nil {
					return err
				}
				continue
			} else {
				return err
			}
		}
		tp, ok := obj.(*types.TypeName)
		if !ok {
			fo := pkg.Fset.File(obj.Pos()).Pos(0)
			f := pkg.Syntax[slices.IndexFunc(pkg.Syntax, func(f *ast.File) bool {
				return f.Pos() == fo
			})]
			decl := f.Scope.Lookup(comment.Identifier)
			err := fmt.Errorf("template %q gotype comment error: unexpected kind for identifier %s (got %s) in package %s", comment.Tree.Name, comment.Identifier, decl.Kind, comment.Package)
			if resolved != nil {
				if err := resolved(comment, nil, err); err != nil {
					return err
				}
			} else {
				return err
			}
			continue
		}
		if resolved != nil {
			if err := resolved(comment, tp.Type(), nil); err != nil {
				return err
			}
		}
	}
	return err
}
