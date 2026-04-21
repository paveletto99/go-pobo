# Runbook

## Repository Discovery Commands

Useful commands for future archaeology:

```sh
rg -n "package main|func main\\("
rg -n "NewServer|ServeHTTPHandler|ServeGRPC|Routes\\("
rg -n "type .*DB|func New\\(db \\*database.DB\\)"
rg -n "InTx\\(|Lock\\(|MultiLock\\("
rg -n "errors\\.Is|errors\\.As|ErrNotFound|http\\.Status"
go list ./...
go test -short ./...
```

## Starter Commands

From `blueprint`:

```sh
make test-template
make skaffold-render
make yaml-check
```

To create a new service:

```sh
./scripts/scaffold-service.sh my-service ../my-service
```

To run with Skaffold:

```sh
skaffold dev
```

## Local DB Notes

The original repo uses `scripts/dev` to run a local Postgres container with TLS and `cmd/migrate` for migrations. The starter keeps DB configuration compatible with env vars:

- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_HOST`
- `DB_PORT`
- `DB_SSLMODE`

For Kubernetes, these are split between ConfigMap and Secret examples.

## Operational Expectations

- `/health` should be wired before shipping.
- `PORT` defaults to `8080`.
- Shutdown is driven by SIGINT/SIGTERM context cancellation.
- Logs should include request id and structured fields.
- Scheduled job endpoints should use DB locks when concurrent execution would be harmful.

## Validation Log

Record validation results in `docs/validation.md` after changing the blueprint:

- repo `go list ./...`,
- repo `go test -short ./...` or reason not run,
- starter `go test ./...`,
- `skaffold render` if Skaffold is installed,
- YAML parse/apply dry-run if available.
