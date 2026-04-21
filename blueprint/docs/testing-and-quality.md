# Testing And Quality

## Observed Test Structure

- Tests live next to packages: `internal/publish/publish_test.go`, `internal/export/database/export_test.go`, `pkg/database/lock_test.go`.
- Table-driven tests are common.
- `t.Parallel()` is used heavily when safe.
- HTTP handlers use `httptest`.
- Constructor tests verify `NewServer` dependency checks using `serverenv.New(...)`.
- Database tests use `pkg/database.TestInstance`, which starts Postgres through `dockertest`, runs migrations once, and clones a template database for each test.
- `go test -short ./...` skips database tests through `testing.Short()`.
- `SKIP_DATABASE_TESTS` also skips database integration tests.
- Some orchestration tests use fakes or function fields rather than broad interfaces, especially in `federationin`.

## Quality Gates

Observed `Makefile` targets:

- `make test`: `go test -shuffle=on -count=1 -short -timeout=5m ./...`
- `make test-acc`: race-enabled full test suite with coverage.
- `make lint`: golangci-lint.
- `make zapcheck`: structured logging check.
- `make generate-check`, `make protoc-check`, `make tabcheck`.

Guidance: **preserve as-is** where possible.

## Handler Tests

Preserve:

- Construct a `Server` with fake dependencies where possible.
- Call the handler or router directly with `httptest.NewRecorder`.
- Assert HTTP status and response JSON.

The starter includes a handler test with fake service behavior.

## Service / Use-Case Tests

Observed:

- Some use-case logic is tested by calling private or package-level helpers.
- Function injection is used where it makes tests simple.
- Model/domain logic has focused unit tests.

Guidance:

- Keep pure business rules in `model` or helper functions and test them directly.
- Use narrow fakes for repository/external dependencies.
- Avoid exporting interfaces solely for mocks unless another package really consumes them.

## Repository Tests

Observed:

- Repository tests use real Postgres via `dockertest`.
- Driver errors are normalized to sentinel errors and asserted with `errors.Is`.
- Transaction behavior is tested by exercising real DB state.

Guidance:

- Repository tests should be integration tests by default, using the DB harness.
- Unit-test SQL string builders only when they contain meaningful branching.
- Keep transaction boundaries inside repository methods.

## Error Handling And Validation

Observed:

- JSON validation happens in `jsonutil.Unmarshal`: content type, max body size, unknown fields, syntax/type errors, single-object body.
- Config validation is done in service config `Validate` methods or service constructors.
- Domain validation often happens in model transformation code.
- Repository errors are wrapped with operation context and sentinel errors are used for expected branches.
- Transport handlers map domain/repository errors to HTTP/gRPC responses.

Guidance:

- Parse and validate request shape at the handler boundary.
- Validate domain rules in model/service helpers.
- Let repositories validate persistence invariants and return sentinel/typed errors.
- Map errors to transport responses at the handler/use-case boundary.
- Log server-side causes, but return stable public messages.

## Anti-Patterns To Avoid

- Broad interfaces with one implementation.
- Repositories returning HTTP statuses.
- Handlers directly constructing SQL.
- Long-running work without context timeout.
- Logging request bodies or secrets.
- Returning `err.Error()` to public clients for internal failures.
- Spawning goroutines from request paths without cancellation/ownership.
