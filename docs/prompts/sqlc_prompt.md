# LLM Architecture Prompt: SQLC Usage for Application

## SQLC Purpose

SQLC is used to generate type-safe Go code from raw SQL queries.  
It defines the database access layer. SQLC-generated code must be treated as the source of truth for interacting with the database.

## SQL Files and Code Generation

- SQL files contain **schemas** (DDL) and **queries**.
- Queries must be annotated with the `-- name: QueryName :result_type` syntax where `:result_type` is one of `:one`, `:many`, `:exec`, or `:execrows`.
- SQLC parses the SQL and generates corresponding Go interfaces, methods, and types.

Example:

```sql
-- name: GetAuthor :one
SELECT id, name, birth_year FROM authors WHERE id = $1;
```

Generates:

```go
type GetAuthorRow struct {
	ID        int64
	Name      string
	BirthYear int
}

func (q *Queries) GetAuthor(ctx context.Context, id int64) (GetAuthorRow, error)
```

## SQLC-Generated Code Expectations

- Each SQL file corresponds to a Go package, typically placed in `internal/database`.
- SQLC-generated methods must use `context.Context` and primitive types (e.g., `int64`, `string`, `bool`) or simple Go structs.
- Nullable database fields map to `sql.NullX` types or pointers, depending on configuration.
- Struct field names in Go match column names in the SQL but are PascalCase.

## Design Constraints

- SQL query design must minimize runtime parsing complexity (e.g., no dynamic SQL generation inside queries).
- Query parameters must be positional (`$1`, `$2`, etc.).
- Avoid joins that result in difficult-to-type or ambiguous return sets unless carefully controlled with explicit query names and expected results.
- Write simple, composable queries. Complex aggregation should be moved into the database via views or server-side aggregation functions if necessary.

## Error Handling

- SQLC methods that fetch one row (`:one`) return a `(RowType, error)`.  
  Applications must handle `sql.ErrNoRows` cleanly in the domain layer.
- SQLC `:exec` methods return `(sql.Result, error)` for inserts, updates, and deletes.

## Separation of Concerns

- Domain methods in `internal/domain` must call SQLC-generated methods, never executing raw SQL directly.
- Application logic must not bypass the `internal/database` layer.
- Business rules based on SQLC query results must be defined only in `internal/domain`.

## Exclusions

- Do not generate new types or methods outside of what the SQLC configuration allows.
- Do not introduce dynamic query generation in Go code.
- Do not wrap or reformat SQLC-generated code except inside domain logic.

## Enforcement

- Generated code must precisely match the SQL query expectations.
- Domain services must not reinterpret raw database types; type mapping must be explicit at the SQLC generation boundary.
- Testing must simulate the `Queries` interface or equivalent using fakes when isolating domain tests.
