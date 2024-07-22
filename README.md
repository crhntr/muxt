# Template [![Go Reference](https://pkg.go.dev/badge/github.com/crhntr/template.svg)](https://pkg.go.dev/github.com/crhntr/template)


## "github.com/crhntr/template/templatetext"
Given the following three files in the package "hypertext" in the module "example.com":
- template_test.go
- template.go
- templates.gohtml

the function `templatetest.AssertTypeCommentsAreFound` will ensure package and identifier in the `{{- /* gotype: ... */ -}}` comment are found.

```go
package hypertext_test

import (
	"testing"

	"github.com/crhntr/template/templatetest"
)

func TestSource(t *testing.T) {
	templatetest.AssertTypeCommentsAreFound(t, "", "", "*.gohtml")
}
```

```go
package hypertext

type Website struct {
  Name string
}
```

```
{{- define "with no space after colon" -}}
  {{- /* gotype: example.com/hypertext.Website */ -}}
  <header>
    {{- .Name -}}
  </header>
{{- end -}}
```
