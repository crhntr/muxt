package configuration

import (
	"flag"
	"fmt"
	"go/token"
	"io"
	"path/filepath"

	"github.com/crhntr/muxt"
)

const (
	outputFlagNameHelp = `The generated file name containing the routes function and receiver interface.`
	outputFlagName     = "output-file"

	templatesVariableHelp = `the name of the global variable with type *"html/template".Template in the working directory package.`
	templatesVariable     = "templates-variable"

	routesFuncHelp = `The function name for the package registering handler functions on an *"net/http".ServeMux.
This function also receives an argument with a type matching the name given by receiver-interface.`
	routesFunc = "routes-func"

	receiverStaticTypeHelp = `The type name for a named type to use for looking up method signatures. If not set, all methods added to the receiver interface will have inferred signatures with argument types based on the argument identifier names. The inferred method signatures always return a single result of type any.`
	receiverStaticType     = "receiver-type"

	receiverStaticTypePackageHelp = `The package path to use when looking for receiver-type. If not set, the package in the current directory is used.`
	receiverStaticTypePackage     = "receiver-type-package"

	receiverInterfaceNameHelp = `The interface name in the generated output-file listing the methods used by the handler routes in routes-func.`
	receiverInterfaceName     = "receiver-interface"

	errIdentSuffix = " value must be a well-formed Go identifier"
)

type Generate struct {
	GoFile string
	GoLine string

	TemplatesVariable         string
	OutputFilename            string
	RoutesFunction            string
	ReceiverIdent             string
	ReceiverStaticTypePackage string

	ReceiverInterfaceIdent string
}

func NewGenerate(args []string, getEnv func(string) string, stderr io.Writer) (Generate, error) {
	g := Generate{
		GoFile: getEnv("GOFILE"),
		GoLine: getEnv("GOLINE"),
	}
	flagSet := g.FlagSet()
	flagSet.SetOutput(stderr)
	if err := flagSet.Parse(args); err != nil {
		return g, err
	}
	if g.TemplatesVariable != "" && !token.IsIdentifier(g.TemplatesVariable) {
		return Generate{}, fmt.Errorf(templatesVariable + errIdentSuffix)
	}
	if g.RoutesFunction != "" && !token.IsIdentifier(g.RoutesFunction) {
		return Generate{}, fmt.Errorf(routesFunc + errIdentSuffix)
	}
	if g.ReceiverIdent != "" && !token.IsIdentifier(g.ReceiverIdent) {
		return Generate{}, fmt.Errorf(receiverStaticType + errIdentSuffix)
	}
	if g.ReceiverInterfaceIdent != "" && !token.IsIdentifier(g.ReceiverInterfaceIdent) {
		return Generate{}, fmt.Errorf(receiverInterfaceName + errIdentSuffix)
	}
	if g.OutputFilename != "" && filepath.Ext(g.OutputFilename) != ".go" {
		return Generate{}, fmt.Errorf("output filename must use .go extension")
	}
	return g, nil
}

func (g *Generate) FlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("generate", flag.ContinueOnError)
	flagSet.StringVar(&g.OutputFilename, outputFlagName, muxt.DefaultOutputFileName, outputFlagNameHelp)
	flagSet.StringVar(&g.TemplatesVariable, templatesVariable, muxt.DefaultTemplatesVariableName, templatesVariableHelp)
	flagSet.StringVar(&g.RoutesFunction, routesFunc, muxt.DefaultRoutesFunctionName, routesFuncHelp)
	flagSet.StringVar(&g.ReceiverIdent, receiverStaticType, "", receiverStaticTypeHelp)
	flagSet.StringVar(&g.ReceiverStaticTypePackage, receiverStaticTypePackage, "", receiverStaticTypePackageHelp)
	flagSet.StringVar(&g.ReceiverInterfaceIdent, receiverInterfaceName, muxt.DefaultReceiverInterfaceName, receiverInterfaceNameHelp)
	return flagSet
}
