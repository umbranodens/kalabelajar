# Changelog

## 2026-05-13 - Foundation scaffold

### Added
- Initialized the Go module for `github.com/umbranodens/kalabelajar`.
- Added config loading from `.env` with validation and safe defaults.
- Added PostgreSQL DSN, connect, and ping helpers using GORM.
- Added Gin server scaffold with `/healthz` and static asset mounting.
- Added seed command placeholder for the upcoming Super Admin seed slice.
- Added the requested project directory structure with placeholders.
- Added `.env.example` and `.gitignore` so local `.env` credentials stay untracked.
- Added first-slice implementation plan under `docs/superpowers/plans`.

### Tested
- Ran config and database TDD red checks before implementation.
- Ran `go test ./...` successfully.
- Started the local server with the provided `.env`.
- Verified `GET /healthz` returns `200 ok` via local HTTP request.

### Notes
- The in-app browser client blocked local `localhost` and `127.0.0.1` URLs with `ERR_BLOCKED_BY_CLIENT`; the health endpoint was verified with `Invoke-WebRequest` instead.
- Auth, models, migrations, RBAC, and seed Super Admin are intentionally left for the next Phase 1 slices.
