# Changelog

## 2026-05-13 - Google SSO and RBAC

### Added
- Added Google OAuth start and callback routes.
- Added Google profile login/upsert support that creates pending-role users or updates existing accounts by email.
- Added the `Masuk dengan Google` login action when OAuth environment values are configured.
- Added backend RBAC middleware for loading sessions, requiring authentication, and enforcing allowed role IDs.
- Added automated coverage for Google profile login and RBAC allow/deny behavior.

### Changed
- Server now loads session context middleware for authenticated routes.

### Tested
- Ran `go test ./...` successfully.
- Ran `go run ./cmd/seed` successfully.
- Verified `/login` renders with the Google login action.
- Verified `/auth/google` returns a redirect to `accounts.google.com` without printing OAuth secrets.
- Verified seeded Super Admin login still reaches `/dashboard` in the in-app browser.

### Notes
- Full `TC-AUTH-3.1` external Google consent flow still requires completing real Google OAuth in a browser session. The backend callback and Google profile session creation path are implemented and covered at service level.

## 2026-05-13 - Local auth foundation

### Added
- Added auth `sessions` model and migration registration.
- Added bcrypt password hashing and local email-password register/login service.
- Added server-side session creation and lookup through the `kb_session` cookie.
- Added role and Super Admin seed service with default local account `admin@kalabelajar.com`.
- Added seed command execution for roles and Super Admin.
- Added server-rendered login, register, pending-role, and dashboard pages in Bahasa Indonesia.
- Added minimal responsive Kala Belajar styling with brand colors and fonts.

### Changed
- Server startup now runs migrations, seeds roles/Super Admin, loads templates, and registers auth routes.
- Database logging now suppresses expected record-not-found noise.

### Tested
- Ran `go test ./...` successfully.
- Ran `go run ./cmd/seed` successfully.
- Started the local server and verified `/login`, local admin login, `/dashboard`, register success, duplicate email validation, password confirmation validation, and wrong-password login validation via HTTP.
- Verified the local admin login flow in the in-app browser.

### Notes
- Google SSO is still pending and should be the next auth slice.
- Email fields use `type="text"` with `inputmode="email"` because server-side validation owns the email rule and this keeps local browser automation compatible.

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
