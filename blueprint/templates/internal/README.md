# internal Template Guidance

A new service should normally own:

- `internal/<service>/config.go`,
- `internal/<service>/server.go`,
- `internal/<service>/handle_<action>.go`,
- `internal/<service>/service.go` when orchestration deserves a helper,
- `internal/<service>/database` for concrete DB adapters,
- `internal/<service>/model` for domain structs and pure rules.

Shared internal runtime packages in the starter:

- `internal/setup`,
- `internal/serverenv`,
- `internal/middleware`,
- `internal/jsonutil`,
- `internal/buildinfo`.

These mirror the source repo's package boundaries but are intentionally smaller.
