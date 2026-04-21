# pkg Template Guidance

Copy only the shared runtime packages that the new service actually needs.

Usually copy:

- `pkg/logging` for zap context logging,
- `pkg/server` for graceful HTTP serving, health, and optional metrics,
- `pkg/database` for pgx config and `InTx`,
- `pkg/render` for simple JSON responses.

Do not copy all of `pkg` by default. In the source repository, `pkg` contains reusable utilities shared by many binaries, but a new service should keep its surface small.
