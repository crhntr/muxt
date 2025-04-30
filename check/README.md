# Typelate

**Typelate** is a Go library for statically type-checking your Go text/templates and html/templates.
It was originally developed to be used with [muxt check](https://github.com/typelate/muxt), which provides a higher-level
CLI for verifying template correctness.
If you’re just looking for command-line usage or are are willing to shell out to a tool in your tests,
consider using [muxt check](https://github.com/typelate/muxt) directly instead.

> **Disclaimer:** The Typelate package’s API may change in the future, and no guarantees of API stability are currently
> provided.

## Why Typelate?

Do you type up templates late at night wishing you had a type checker for your templates? Try typelate.
Typelate leverages Go’s type checker to help you identify mismatched function calls, incorrect field accesses, and other
type errors in your templates—before you ship your code.

- **Reduce runtime errors:** Catch mistakes in function calls or template fields (e.g., `{{.NonExistentField}}`) during
  development.
- **Improve maintainability:** Type-checking your templates makes refactoring less risky.
- **Compliments the Go toolchain:** Uses Go’s own `go/types` to analyze your code.

## How It Works

1. **Load Packages**
   Typelate requires type information about your Go code, so you’ll typically load your packages with [
   `golang.org/x/tools/go/packages`](https://pkg.go.dev/golang.org/x/tools/go/packages).

2. **Provide Template Trees**
   You give Typelate the parsed template (via its internal `*parse.Tree`), along with a function (`TreeFinder`) to
   locate other templates by name.

3. **Specify Available Functions**  
   You provide a custom or default set of allowed functions (via `DefaultFunctions` or your own `Functions` map) that
   your templates can call.

4. **Call `Check`**  
   Typelate’s core entry point is
   `Check(*parse.Tree, types.Type, *types.Package, *token.FileSet, TreeFinder, CallChecker) error`. This inspects each
   node in the template parse tree, validating that fields, variables, and function calls match the expected types.

## Basic Usage Example

Below is a simplified outline for how you might integrate Typelate into your project. This assumes you already have
parsed templates, a `*types.Package`, and a `*token.FileSet`.

```go
package main_test

import (
	"fmt"
	"go/token"
	"go/types"
	"log"
	"testing"
	"text/template"
	"text/template/parse"

	"golang.org/x/tools/go/packages"

	"github.com/typelate/muxt/typelate"
)

func Test(t *testing.T) {
	// 1. Load packages to gather type information.
	patterns := []string{
		".", "encoding", "fmt", "net/http",
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedModule | packages.NeedTypesInfo | packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedEmbedPatterns | packages.NeedEmbedFiles,
		Dir:  ".",
	}, patterns...)
	if err != nil {
		t.Fatal(err)
	}

	// Pick a package that contains the data your templates rely on.
	// This is just an example — you'll need to locate your actual package.
	var myPkg *packages.Package
	if len(pkgs) > 0 {
		myPkg = pkgs[0]
	} else {
		t.Fatalf("No packages found")
	}

	// 2. Parse or load your templates.
	tmpl, err := template.New("example").Parse(`Hello, {{.Name}}!`)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// 3. Build a function set (either default or custom).
	functions := typelate.DefaultFunctions(myPkg.Types)

	// 4. Create a TreeFinder. For single-template usage, a trivial one:
	treeFinder := typelate.FindTreeFunc(func(name string) (*parse.Tree, bool) {
		named := tmpl.Lookup(name)
		if named == nil {
			return nil, false
		}
		return named.Tree, true
	})

	// 5. Type-check your template. Suppose .Name is a string in your data type.
	//    We'll need the type from our loaded package that represents .Name.
	//    For demonstration, assume you identified a struct type named "Person".
	personObj := myPkg.Types.Scope().Lookup("Person") // e.g., type Person struct { Name string }
	if personObj == nil {
		t.Fatalf("Could not find type Person in package %s", myPkg.PkgPath)
	}

	// 6. Finally, call Check.
	err = typelate.Check(
		tmpl.Tree,
		personObj.Type(), // The type that your template's dot should match
		myPkg.Types,
		fset,
		treeFinder,
		functions,
	)
	if err != nil {
		t.Logf("Template type-check error: %v\n", err)
	} else {
		t.Logf("Template is type-correct!")
	}
}
```

## Known Issues / Caveats

1. **Static Types Must Be Provided**  
   You must explicitly provide the Go `types.Type` that matches your template’s data. Without that, the checker can’t
   validate fields.

2. **Distinguishing `text/template` vs. `html/template`**  
   Typelate doesn’t yet differentiate between standard text vs. HTML template built-ins. We plan to split out
   `DefaultFunctions` into dedicated `TextDefaultFunctions` and `HTMLDefaultFunctions`.

3. **Google Safe HTML Templates**  
   It’s unclear whether Typelate works seamlessly with [safehtml](https://pkg.go.dev/github.com/google/safehtml) or other
   specialized template libraries.
   If you find it works, PR a removal of this warning. If it fails, send me a link to your code and I'll see if the
   typelate API can be improved to
   let it work with your template library.

4. **Exec-Time Panics vs. Compile-Time Checks**  
   Some template errors only appear at runtime (e.g., out-of-range indexing)
   Typelate can catch type-level issues but can’t detect certain execution-time conditions (like an index value that ends
   up being negative).

## Contributing

Contributions or issue reports are welcome! Please open a pull request or file an issue if you find a bug or think of an
enhancement.
