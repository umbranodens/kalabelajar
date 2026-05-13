# Google SSO and RBAC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the auth foundation with Google SSO compatibility and backend RBAC middleware.

**Architecture:** The auth service owns provider-neutral session creation and Google-profile upsert behavior. The handler owns OAuth redirects/callbacks and user-info fetch. Middleware reads the existing auth session cookie, loads the user into request context, and enforces deny-by-default role checks.

**Tech Stack:** Go 1.23, Gin, GORM, OAuth2, Google OAuth endpoint, server-side sessions.

---

### Task 1: Google Auth Service

**Files:**
- Modify: `internal/services/auth_service_test.go`
- Modify: `internal/services/auth_service.go`
- Modify: `internal/repositories/core_data.go`

- [x] **Step 1: Write failing tests**

```go
func TestLoginWithGoogleCreatesPendingUserAndSession(t *testing.T) {
	result, err := service.LoginWithGoogleProfile(ctx, services.GoogleProfile{Email: "new@example.com", Name: "New User", ProviderID: "gid"})
	require.NoError(t, err)
	assert.Equal(t, models.ProviderGoogle, result.User.Provider)
	assert.Nil(t, result.User.RoleID)
	assert.NotEmpty(t, result.Session.Token)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/services`
Expected: FAIL because `LoginWithGoogleProfile` and repository support do not exist.

- [x] **Step 3: Implement Google profile login**

Add `GoogleProfile`, update existing users by email, create new pending users, reject inactive users, and create a normal server-side session.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/services`
Expected: PASS.

### Task 2: RBAC Middleware

**Files:**
- Create: `internal/middleware/rbac_test.go`
- Create: `internal/middleware/rbac.go`

- [x] **Step 1: Write failing tests**

```go
func TestRequireRolesAllowsMatchingRole(t *testing.T) {
	router.GET("/admin", middleware.RequireRoles(models.RoleIDSuperAdmin), handler)
	requestWithSessionCookie(router, "token")
	assert.Equal(t, http.StatusOK, recorder.Code)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware`
Expected: FAIL because middleware is missing.

- [x] **Step 3: Implement middleware**

Add `LoadSession`, `RequireAuth`, `RequireRoles`, and `CurrentUser`. Unauthenticated requests redirect to login for browser routes; authenticated wrong-role requests return 403.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware`
Expected: PASS.

### Task 3: OAuth Routes and Login UI

**Files:**
- Modify: `internal/handlers/auth.go`
- Modify: `cmd/server/main.go`
- Modify: `web/templates/login.html`

- [x] **Step 1: Add OAuth routes**

Expose `GET /auth/google` and `GET /auth/google/callback`. Use a state cookie and Google userinfo endpoint.

- [x] **Step 2: Add login button**

Add `Masuk dengan Google` on the login page with a clear fallback if config is missing.

- [x] **Step 3: Verify route behavior**

Run local server and verify `/auth/google` redirects to Google when env values exist or returns a clear error when missing.

### Task 4: Docs, Commit, Push

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `implementation-summary.md`

- [x] **Step 1: Run full verification**

Run `go test ./...`, seed, server/browser checks for login page and Google link behavior.

- [x] **Step 2: Update docs**

Record verified test cases and remaining limitations for full external OAuth completion.

- [ ] **Step 3: Commit and push**

Commit message: `feat: add google sso and rbac`
