# Muxt [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/muxt.svg)](https://pkg.go.dev/github.com/crhntr/muxt) [![Go](https://github.com/crhntr/muxt/actions/workflows/go.yml/badge.svg)](https://github.com/crhntr/muxt/actions/workflows/go.yml)

**Muxt** is a Go code generator that helps you build server-side rendered web apps with minimal overhead, leveraging Go 1.22’s improved `http.ServeMux` and standard `html/template` features.
No extra runtime dependencies are required—just plain Go code.

- It allows you to register HTTP routes from [HTML templates](https://pkg.go.dev/html/template)
- It generates handler functions and registers them on an [`*http.ServeMux`](https://pkg.go.dev/net/http#ServeMux)
- It generates code in handler functions to parse path parameters and form fields
- It generates a receiver interface to represent the boundary between your app code and HTTP/HTML
  - Use this to mock out your server and test the view layer of your application

### Used By
- [portfoliotree.com](https://portfoliotree.com)

## Installation

You can install it using the Go toolchain.
```bash
cd # Change outside of your module (so you don't add muxt to your dependency chain)
go install github.com/crhntr/muxt@latest
cd -
```
You do not need to add this tool to your module ([unless you want to use the tools pattern](https://play-with-go.dev/tools-as-dependencies_go119_en/)).

## Usage

Commands:
- `muxt generate` generate a routes function and receiver interface
- `muxt version` writes muxt version to standard out
- `muxt check` static type check your templates
- `muxt documentation` (wip) template documentation

### Generate Command

This command is how you use muxt. 

You can call it either by invoking it from your terminal or by adding a generate comment to a Go source file.

<details>
<summary>Flags</summary>
<pre>
Usage of generate:
  -output-file string
    	The generated file name containing the routes function and receiver interface. (default "template_routes.go")
  -receiver-interface string
    	The interface name in the generated output-file listing the methods used by the handler routes in routes-func. (default "RoutesReceiver")
  -receiver-type string
    	The type name for a named type to use for looking up method signatures. If not set, all methods added to the receiver interface will have inferred signatures with argument types based on the argument identifier names. The inferred method signatures always return a single result of type any.
  -receiver-type-package string
    	The package path to use when looking for receiver-type. If not set, the package in the current directory is used.
  -routes-func string
    	The function name for the package registering handler functions on an *"net/http".ServeMux.
    	This function also receives an argument with a type matching the name given by receiver-interface. (default "routes")
  -templates-variable string
    	the name of the global variable with type *"html/template".Template in the working directory package. (default "templates")

</pre>
</details>

#### Shell Example

If you invoke it from the shell, it expects to find a Go source package in the current directory where it can find a templates variable.

```shell
muxt generate
```

#### Go Generate Comment Example

If you do the generate comment, make sure you need to write the comment in the same package as your (globally scoped) `templates` variable.

```go
package main

import (
    "embed"
    "html/template"
)

//go:generate muxt generate

var (
    //go:embed *.gohtml
    templatesSource embed.FS

    templates = template.Must(template.ParseFS(templatesSource, "*.gohtml"))
)

```

## Examples

The [example directory](example) has a worked example.

For a more complete example, see: https://github.com/crhntr/muxt-template-module-htmx
