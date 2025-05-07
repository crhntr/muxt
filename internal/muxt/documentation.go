package muxt

import (
	"cmp"
	"fmt"
	"go/token"
	"go/types"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/crhntr/muxt/internal/source"
)

func Documentation(w io.Writer, wd string, config RoutesFileConfiguration) error {
	config = config.applyDefaults()
	if !token.IsIdentifier(config.PackageName) {
		return fmt.Errorf("package name %q is not an identifier", config.PackageName)
	}

	patterns := []string{wd, "net/http"}
	if config.ReceiverPackage != "" {
		patterns = append(patterns, config.ReceiverPackage)
	}

	fileSet := token.NewFileSet()
	pl, err := packages.Load(&packages.Config{
		Fset: fileSet,
		Mode: packages.NeedModule | packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedEmbedPatterns | packages.NeedEmbedFiles,
		Dir:  wd,
	}, patterns...)
	if err != nil {
		return err
	}
	file, err := source.NewFile(filepath.Join(wd, config.OutputFileName), fileSet, pl)
	if err != nil {
		return err
	}
	routesPkg := file.OutputPackage()

	config.PackagePath = routesPkg.PkgPath
	config.PackageName = routesPkg.Name
	var receiver *types.Named
	if config.ReceiverType != "" {
		receiverPkgPath := cmp.Or(config.ReceiverPackage, config.PackagePath)
		receiverPkg, ok := file.Package(receiverPkgPath)
		if !ok {
			return fmt.Errorf("could not determine receiver package %s", receiverPkgPath)
		}
		obj := receiverPkg.Types.Scope().Lookup(config.ReceiverType)
		if config.ReceiverType != "" && obj == nil {
			return fmt.Errorf("could not find receiver type %s in %s", config.ReceiverType, receiverPkg.PkgPath)
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return fmt.Errorf("expected receiver %s to be a named type", config.ReceiverType)
		}
		receiver = named
	} else {
		receiver = types.NewNamed(types.NewTypeName(0, routesPkg.Types, "Receiver", nil), types.NewStruct(nil, nil), nil)
	}

	ts, functions, err := source.Templates(wd, config.TemplatesVariable, routesPkg)
	if err != nil {
		return err
	}
	templates, err := Templates(ts)
	if err != nil {
		return err
	}

	writeOutput(w, functions, templates, receiver)

	return nil
}

func writeOutput(w io.Writer, functions source.Functions, templates []Template, receiver *types.Named) {
	_, _ = fmt.Fprintf(w, "functions:\n")
	names := slices.Collect(maps.Keys(functions))
	for _, name := range names {
		s := strings.TrimPrefix(functions[name].String(), "func")
		_, _ = fmt.Fprintf(w, "  - func %s%s\n", name, s)
	}

	_, _ = fmt.Fprintf(w, "routes:\n")
	for _, t := range templates {
		_, _ = fmt.Fprintf(w, "  - %s\n", t.String())
	}

	_, _ = fmt.Fprintf(w, "reciever: %s\n", receiver.String())
	if receiver.NumMethods() > 0 {
		_, _ = fmt.Fprintf(w, "reciever_methods:\n")
	}
	for i := 0; i < receiver.NumMethods(); i++ {
		m := receiver.Method(i)
		_, _ = fmt.Fprintf(w, "  - func %s%s\n", m.Name(), strings.TrimPrefix(m.Signature().String(), "func"))
	}
}
