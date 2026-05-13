# Changelog

## 2026-05-13 - Core data foundation

### Added
- Added core GORM models for roles, users, tutors, students, and activity logs.
- Added UUID creation hooks and role/provider/activity constants aligned with the tech spec.
- Added database migration registration and startup migration execution for core data tables.
- Added GORM repositories for users, tutors, students, and activity logs.
- Added assignment service for assigning a verified active tutor to a student.
- Added backend activity logging for tutor assignment actions.
- Added a Phase 2 core data implementation plan.

### Tested
- Ran TDD red checks before implementing model, migration, and assignment-service APIs.
- Ran `go test ./...` successfully.
- Started the local server with migrations enabled and verified `GET /healthz` returns `200 ok`.

### Notes
- This is a backend core-data foundation slice only; Super Admin UI, auth-gated handlers, filters, and browser CRUD flows are still pending.
- Browser QA remains limited by the in-app browser local URL blocking noted in the previous slice, so the local health check used `Invoke-WebRequest`.

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
