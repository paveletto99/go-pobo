# Package Boundaries

This note is a concise checklist for reviewing new services against repository style.

## Boundary Rules

- `cmd/<service>` may import `internal/<service>`, `internal/setup`, `pkg/logging`, and `pkg/server`.
- `internal/<service>` may import shared runtime packages and its own `database`/`model`.
- `internal/<service>/database` may import `pkg/database`, pgx, and its own model package.
- `internal/<service>/model` should avoid importing transport or database packages unless the repo already has a concrete reason.
- `pkg/*` should not import service-specific `internal/<service>` packages.
- Sibling feature imports are allowed only when the target package is deliberately reused, not for incidental helpers.

## Review Questions

- Does `NewServer` make all runtime dependencies explicit?
- Does `Routes` own middleware and routes?
- Does context flow from request to every IO call?
- Is SQL contained in `database` packages?
- Are errors mapped at the transport boundary?
- Are repository interfaces narrow and consumer-owned, if present?
- Are tests able to exercise use-case logic without running the full HTTP server?
