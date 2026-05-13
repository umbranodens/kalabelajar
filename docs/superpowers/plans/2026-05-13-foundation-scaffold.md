# Foundation Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the initial Go/Gin/PostgreSQL foundation for the Kala Belajar MVP without implementing auth yet.

**Architecture:** The server entrypoint loads environment configuration, opens a PostgreSQL connection through a dedicated database package, and exposes a health endpoint for local verification. Config and database helpers are tested independently so later auth and model work can reuse stable foundations.

**Tech Stack:** Go 1.23, Gin, GORM, PostgreSQL driver, godotenv.

---

### Task 1: Config Loader

**Files:**
- Create: `internal/config/config_test.go`
- Create: `internal/config/config.go`

- [x] **Step 1: Write failing tests**

```go
func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_URL", "http://localhost:9090")
	t.Setenv("APP_SECRET", "secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "kb")
	t.Setenv("DB_PASS", "password")
	t.Setenv("DB_NAME", "kalabelajar_test")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("SESSION_TTL_HOURS", "48")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.App.Port)
	assert.Equal(t, 48*time.Hour, cfg.Session.TTL)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config`
Expected: FAIL because `internal/config` implementation is missing.

- [x] **Step 3: Implement config loader**

Implement `Config`, `AppConfig`, `DatabaseConfig`, `GoogleConfig`, and `SessionConfig`. Load `.env` if present, read environment values, default `APP_PORT` to `8080`, `DB_PORT` to `5432`, `DB_SSLMODE` to `disable`, and `SESSION_TTL_HOURS` to `24`. Return validation errors for missing required app/database values.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config`
Expected: PASS.

### Task 2: Database Foundation

**Files:**
- Create: `internal/database/database_test.go`
- Create: `internal/database/database.go`

- [x] **Step 1: Write failing tests**

```go
func TestBuildDSNUsesPostgresConnectionFields(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host: "localhost", Port: "5432", User: "kb",
		Password: "secret", Name: "kalabelajar", SSLMode: "disable",
	}

	dsn := database.BuildDSN(cfg)

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=kb")
	assert.Contains(t, dsn, "dbname=kalabelajar")
	assert.Contains(t, dsn, "sslmode=disable")
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/database`
Expected: FAIL because `internal/database` implementation is missing.

- [x] **Step 3: Implement database helper**

Implement `BuildDSN`, `Connect`, and `Ping` using GORM with the PostgreSQL driver. Keep migration orchestration ready but leave model migration for the next slice.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/database`
Expected: PASS.

### Task 3: Server Scaffold

**Files:**
- Create: `cmd/server/main.go`
- Create: `cmd/seed/main.go`
- Create: `web/templates/.gitkeep`
- Create: `web/static/css/.gitkeep`
- Create: `web/static/js/.gitkeep`
- Create: `web/static/images/.gitkeep`
- Create: `internal/models/.gitkeep`
- Create: `internal/repositories/.gitkeep`
- Create: `internal/services/.gitkeep`
- Create: `internal/middleware/.gitkeep`
- Create: `internal/handlers/.gitkeep`
- Create: `.env.example`
- Create: `.gitignore`

- [x] **Step 1: Add server entrypoint**

`cmd/server/main.go` loads config, connects to PostgreSQL, pings the DB, serves static files, and exposes `GET /healthz` returning `ok`.

- [x] **Step 2: Add seed placeholder**

`cmd/seed/main.go` loads config and prints a non-secret placeholder message. Actual Super Admin seed belongs to the auth/model slice.

- [x] **Step 3: Add repo structure placeholders**

Create the expected project directories with `.gitkeep` files for empty folders.

- [x] **Step 4: Run verification**

Run: `go test ./...`, `go run ./cmd/server`, and browser/manual `GET /healthz`.
Expected: tests pass; app starts if `.env` database is reachable.

### Task 4: Slice Documentation and Git

**Files:**
- Create: `CHANGELOG.md`
- Create: `implementation-summary.md`

- [x] **Step 1: Update changelog**

Add `2026-05-13 - Foundation scaffold` with added/tested/notes.

- [x] **Step 2: Update implementation summary**

Record branch, commit hash after commit, push result, tests, browser/manual checks, known issues, and next steps.

- [ ] **Step 3: Commit and push**

Run: `git status`, stage only relevant files excluding `.env`, commit `chore: scaffold go app`, push to GitHub.
