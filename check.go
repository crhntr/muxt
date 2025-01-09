package muxt

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/crhntr/muxt/internal/source"
	"github.com/crhntr/muxt/internal/templatetype"
)

func CheckTemplates(wd string, config RoutesFileConfiguration) error {
	config = config.applyDefaults()
	if !token.IsIdentifier(config.PackageName) {
		return fmt.Errorf("package name %q is not an identifier", config.PackageName)
	}
	imports := source.NewImports(&ast.GenDecl{Tok: token.IMPORT})

	patterns := []string{
		wd, "encoding", "fmt", "net/http",
	}

	if config.ReceiverPackage != "" {
		patterns = append(patterns, config.ReceiverPackage)
	}

	pl, err := packages.Load(&packages.Config{
		Fset: imports.FileSet(),
		Mode: packages.NeedModule | packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedEmbedPatterns | packages.NeedEmbedFiles,
		Dir:  wd,
	}, patterns...)
	if err != nil {
		return err
	}
	imports.AddPackages(pl...)

	routesPkg, ok := imports.PackageAtFilepath(wd)
	if !ok {
		return fmt.Errorf("could not find package in working directory %q", wd)
	}

	ts, fm, err := source.Templates(wd, config.TemplatesVariable, routesPkg)
	if err != nil {
		return err
	}
	templates, err := Templates(ts)
	if err != nil {
		return err
	}

	receiverPkgPath := cmp.Or(config.ReceiverPackage, config.PackagePath, routesPkg.PkgPath)
	receiverPkg, ok := imports.Package(receiverPkgPath)
	if !ok {
		return fmt.Errorf("could not determine receiver package %s", receiverPkgPath)
	}
	obj := receiverPkg.Types.Scope().Lookup(config.ReceiverType)
	if config.ReceiverType != "" && obj == nil {
		return fmt.Errorf("could not find receiver type %s in %s", config.ReceiverType, receiverPkg.PkgPath)
	}
	receiver, ok := obj.Type().(*types.Named)
	if !ok {
		return fmt.Errorf("expected receiver %s to be a named type", config.ReceiverType)
	}

	for _, t := range templates {
		methodObj, _, _ := types.LookupFieldOrMethod(receiver, true, receiver.Obj().Pkg(), t.fun.Name)
		if methodObj == nil {
			return fmt.Errorf("failed to generate method %s", t.fun.Name)
		}
		sig := methodObj.Type().(*types.Signature)
		if sig.Results().Len() == 0 {
			return fmt.Errorf("method for pattern %q has no results it should have one or two", t.name)
		}
		dataVar := sig.Results().At(0)
		if types.Identical(dataVar.Type(), types.Universe.Lookup("any").Type()) {
			continue
		}
		fns := templatetype.DefaultFunctions(routesPkg.Types)
		fns.Add(templatetype.Functions(fm))
		if err := templatetype.Check(t.template.Tree, dataVar.Type(), dataVar.Pkg(), routesPkg.Fset, newForrest(ts), fns); err != nil {
			return err
		}
	}

	return nil
}
