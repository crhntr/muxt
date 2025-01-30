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

## Examples

The [example directory](example) has a worked example.

For a more complete example, see: https://github.com/crhntr/muxt-template-module-htmx

## Documentation

### Introduction
- [Getting_Started](./docs/getting_started.md)
- [Notes on Integration with Existing Projects](./docs/integrating.md)
- [Writing Template_Names](./docs/template_names.md)

### Reference
- [call_parameters](./docs/call_parameters.md)
- [call_results](./docs/call_results.md)
- [custom_execute_func](./docs/custom_execute_func.md)
- [templates_variable](./docs/templates_variable.md)
- [action type checking](./docs/action_type_checking.md)
- [known_issues](./docs/known_issues.md)

### Testing
- [testing_hypertext](./docs/testing_hypertext.md)
- [testing_the_receiver](./docs/testing_the_receiver.md)

### Philosophy & Vision
- [manifesto](./docs/manifesto.md)
- [motivation](./docs/motivation.md)
