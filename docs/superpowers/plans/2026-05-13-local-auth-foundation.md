# Local Auth Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add email-password register/login, server-side auth sessions, role seed, and Super Admin seed.

**Architecture:** Models add the auth `sessions` table, services own registration/login/password/session rules, repositories persist users and sessions, seed code creates roles and the local Super Admin account, and Gin handlers render minimal server-side pages for auth flows.

**Tech Stack:** Go 1.23, Gin, GORM, PostgreSQL, bcrypt, Go HTML templates.

---

### Task 1: Auth Models and Migration

**Files:**
- Modify: `internal/models/core_data.go`
- Modify: `internal/models/core_data_test.go`
- Modify: `internal/database/database_test.go`

- [x] **Step 1: Write failing tests**

```go
func TestSessionAssignsUUIDBeforeCreate(t *testing.T) {
	session := models.Session{UserID: uuid.New(), Token: "token", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, session.BeforeCreate(nil))
	assert.NotEqual(t, uuid.Nil, session.ID)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/models ./internal/database`
Expected: FAIL because `models.Session` is missing from models and migration registration.

- [x] **Step 3: Implement session model**

Add `Session` with UUID ID, user relation, token, refresh token, expiration, IP/user-agent, and created timestamp. Add to migration registration.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/models ./internal/database`
Expected: PASS.

### Task 2: Local Auth Service and Seed Service

**Files:**
- Create: `internal/services/auth_service_test.go`
- Create: `internal/services/auth_service.go`
- Create: `internal/services/seed_service_test.go`
- Create: `internal/services/seed_service.go`
- Modify: `internal/repositories/core_data.go`
- Modify: `cmd/seed/main.go`

- [x] **Step 1: Write failing tests**

```go
func TestRegisterCreatesLocalUserWithPasswordHash(t *testing.T) {
	result, err := service.Register(ctx, services.RegisterInput{Name: "Nia", Email: "nia@example.com", Password: "password123", PasswordConfirmation: "password123"})
	require.NoError(t, err)
	assert.Equal(t, models.ProviderLocal, result.User.Provider)
	assert.NotEqual(t, "password123", *result.User.PasswordHash)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/services`
Expected: FAIL because auth service and seed service are missing.

- [x] **Step 3: Implement services**

Implement register validation, bcrypt hashing, login validation, inactive/OAuth-only rejection, session creation, role seeding, and Super Admin seeding.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/services`
Expected: PASS.

### Task 3: Server Routes and Templates

**Files:**
- Create: `internal/handlers/auth.go`
- Create: `web/templates/login.html`
- Create: `web/templates/register.html`
- Create: `web/templates/dashboard.html`
- Create: `web/templates/pending.html`
- Create: `web/static/css/app.css`
- Modify: `cmd/server/main.go`

- [x] **Step 1: Add handlers and routes**

Expose `GET /login`, `POST /login`, `GET /register`, `POST /register`, `POST /logout`, `GET /dashboard`, and `GET /pending` with server-side sessions stored in the `kb_session` cookie.

- [x] **Step 2: Add minimal responsive templates**

Use Bahasa Indonesia and the Kala Belajar color/typography tokens. Keep dashboard operational, not marketing-style.

- [x] **Step 3: Run manual HTTP checks**

Run the server, seed Super Admin, request login/register pages, POST invalid login, POST valid admin login, and request dashboard with the auth cookie.

### Task 4: Docs, Commit, Push

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `implementation-summary.md`

- [x] **Step 1: Run full verification**

Run: `go test ./...`, `go run ./cmd/seed`, server smoke checks.

- [x] **Step 2: Update docs**

Record tests, manual checks, verified auth test cases, and Google OAuth as pending.

- [ ] **Step 3: Commit and push**

Commit message: `feat: add local auth foundation`
