# Database+Domain+Hypertext Prompt: Muxt + SQLC Application

## Code Generation Tools

**Muxt**  
Generates Go HTTP handlers and HTML templates using `html/template`. Routing is based on structured template names.

**sqlc**  
Generates type-safe Go code from SQL queries. SQLC defines the database access layer, avoiding ORM complexity.

## Internal Package Structure

**`internal/database`**
- Generated: SQLC query methods from `.sql` files.
- Manual: Hand-written transaction helpers, DDL scripts, SQL queries, and the `sqlc.yaml` configuration.

**`internal/domain`**
- Manual: Domain models and the `Server` type containing public business methods invoked by hypertext handlers.

**`internal/hypertext`**
- Generated: HTTP handlers based on `.gohtml` templates parsed by Muxt.
- Manual: Hand-crafted HTML templates with optional embedded scripts and styles.

Each package contains a mix of generated and manually written code. All packages include table-driven `_test.go` tests.

## Testing Strategy: Table-Driven Tests

Tests in the domain package follow a strict Given-When-Then table-driven structure.  
Mocks and fakes for dependencies (such as the database) are generated using `counterfeiter`.

### Table-Driven Test Structure Example

```go
func TestServer_Endpoint(t *testing.T) {
	type (
		Given struct {
			db *fake.Database
			// additional fakes if needed
		}
		When struct {
			// request parameters
		}
		Then struct {
			db  *fake.Database
			res *http.Response
			// endpoint results or external service fakes
		}
		Case struct {
			Name  string
			Given func(*testing.T, Given)
			When  func(*testing.T, When) *http.Request
			Then  func(*testing.T, Then)
		}
	)

	for _, tc := range []Case{
		// cases defined here
	} {
		t.Run(tc.Name, func(t *testing.T) {
			db := new(fake.Database)
			// initialize additional fakes

			tc.Given(t, Given{db: db})
			req := tc.When(t, When{})

			rec := httptest.NewRecorder()
			mux := http.NewServeMux()
			srv := endpoint.Server{db: db}
			hypertext.TemplateRoutes(mux, srv)
			mux.ServeHTTP(rec, req)

			tc.Then(t, Then{
				db:  db,
				res: rec.Result(),
			})
		})
	}
}
```

Key characteristics:
- `Given` sets up mocks and initial state.
- `When` defines the HTTP request for the test.
- `Then` asserts HTTP response, database call expectations, and domain method behavior.
- Tests simulate full HTTP flows using `httptest.NewRecorder` and `ServeMux`.

Tests in the domain or the hypertext package should only test hand-written code. Generated code should be tested in the `internal/domain` package with the Given-When-Then pattern.

## Design Constraints

- Handlers must remain thin: extract parameters, call domain methods, render templates.
- Domain methods must accept `context.Context` and typed parameters, and return typed results (structs or structs with error).
- Domain logic must not leak HTTP or database details.
- SQLC-generated code is trusted for database operations; domain logic should not reimplement queries.
- Muxt template names must strictly match the format `[METHOD] /path [HTTP_STATUS] [Method(ctx, param, ...)]`.
- Template action type-checking (`muxt check`) must pass.

## Exclusions

- Do not include redundant boilerplate or general Go idioms unless critical to Muxt or SQLC behavior.
- Do not speculate about dependencies outside the standard library or Muxt/sqlc unless stated.
- Do not invent new flows, architectural patterns, or types not described here.

## Enforcement

- Strictly respect type signatures and separation of concerns.
- Generated code must align with package ownership: database queries in `internal/database`, business logic in `internal/domain`, routing and rendering in `internal/hypertext`.
- Tests must simulate real HTTP request/response flows using fakes for external dependencies.
