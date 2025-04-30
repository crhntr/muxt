package configuration

import (
	"flag"
	"fmt"
	"go/token"
	"io"
	"path/filepath"

	"github.com/typelate/muxt/internal/muxt"
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
	ReceiverStaticType     = "receiver-type"

	receiverStaticTypePackageHelp = `The package path to use when looking for receiver-type. If not set, the package in the current directory is used.`
	receiverStaticTypePackage     = "receiver-type-package"

	receiverInterfaceNameHelp = `The interface name in the generated output-file listing the methods used by the handler routes in routes-func.`
	receiverInterfaceName     = "receiver-interface"

	errIdentSuffix = " value must be a well-formed Go identifier"
)

func NewRoutesFileConfiguration(args []string, stderr io.Writer) (muxt.RoutesFileConfiguration, error) {
	var g muxt.RoutesFileConfiguration
	flagSet := RoutesFileConfigurationFlagSet(&g)
	flagSet.SetOutput(stderr)
	if err := flagSet.Parse(args); err != nil {
		return g, err
	}
	if g.TemplatesVariable != "" && !token.IsIdentifier(g.TemplatesVariable) {
		return muxt.RoutesFileConfiguration{}, fmt.Errorf(templatesVariable + errIdentSuffix)
	}
	if g.RoutesFunction != "" && !token.IsIdentifier(g.RoutesFunction) {
		return muxt.RoutesFileConfiguration{}, fmt.Errorf(routesFunc + errIdentSuffix)
	}
	if g.ReceiverType != "" && !token.IsIdentifier(g.ReceiverType) {
		return muxt.RoutesFileConfiguration{}, fmt.Errorf(ReceiverStaticType + errIdentSuffix)
	}
	if g.ReceiverInterface != "" && !token.IsIdentifier(g.ReceiverInterface) {
		return muxt.RoutesFileConfiguration{}, fmt.Errorf(receiverInterfaceName + errIdentSuffix)
	}
	if g.OutputFileName != "" && filepath.Ext(g.OutputFileName) != ".go" {
		return muxt.RoutesFileConfiguration{}, fmt.Errorf("output filename must use .go extension")
	}
	return g, nil
}

func RoutesFileConfigurationFlagSet(g *muxt.RoutesFileConfiguration) *flag.FlagSet {
	flagSet := flag.NewFlagSet("generate", flag.ContinueOnError)
	flagSet.StringVar(&g.OutputFileName, outputFlagName, muxt.DefaultOutputFileName, outputFlagNameHelp)
	flagSet.StringVar(&g.TemplatesVariable, templatesVariable, muxt.DefaultTemplatesVariableName, templatesVariableHelp)
	flagSet.StringVar(&g.RoutesFunction, routesFunc, muxt.DefaultRoutesFunctionName, routesFuncHelp)
	flagSet.StringVar(&g.ReceiverType, ReceiverStaticType, "", receiverStaticTypeHelp)
	flagSet.StringVar(&g.ReceiverPackage, receiverStaticTypePackage, "", receiverStaticTypePackageHelp)
	flagSet.StringVar(&g.ReceiverInterface, receiverInterfaceName, muxt.DefaultReceiverInterfaceName, receiverInterfaceNameHelp)
	return flagSet
}
