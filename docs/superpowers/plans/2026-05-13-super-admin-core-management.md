# Super Admin Core Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Super Admin dashboard and core data management screens for users, tutors, students, tutor assignment, and activity logs.

**Architecture:** Seed service creates realistic local data. Repositories expose dashboard/list/update queries. An Admin handler renders server-side operational pages protected by RBAC middleware and delegates tutor assignment to the existing assignment service.

**Tech Stack:** Go 1.23, Gin, GORM, PostgreSQL, Go templates, server-side sessions.

---

### Task 1: Admin Services and Seed Data

**Files:**
- Create: `internal/services/admin_service_test.go`
- Create: `internal/services/admin_service.go`
- Modify: `internal/services/seed_service.go`
- Modify: `internal/repositories/core_data.go`

- [x] **Step 1: Write failing tests**

```go
func TestAdminServiceVerifyTutorLogsActivity(t *testing.T) {
	err := service.VerifyTutor(ctx, actorID, tutorID, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, models.ActivityVerifyTutor, logs.entries[0].Action)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/services`
Expected: FAIL because AdminService is missing.

- [x] **Step 3: Implement service/repository methods**

Add dashboard counts, user/tutor/student/activity list queries, user active toggle, tutor verification, local seed data, and activity log creation.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/services`
Expected: PASS.

### Task 2: Super Admin Handlers and Templates

**Files:**
- Create: `internal/handlers/admin.go`
- Create: `web/templates/admin_users.html`
- Create: `web/templates/admin_tutors.html`
- Create: `web/templates/admin_students.html`
- Create: `web/templates/admin_activity.html`
- Modify: `web/templates/dashboard.html`
- Modify: `cmd/server/main.go`
- Modify: `web/static/css/app.css`

- [x] **Step 1: Add RBAC-protected routes**

Add `/admin/users`, `/admin/tutors`, `/admin/students`, `/admin/activity`, and POST actions for active toggle, tutor verify, and tutor assignment.

- [x] **Step 2: Add operational templates**

Use tables/cards, filters, clear empty states, and Indonesian labels.

- [x] **Step 3: Browser QA**

Login as Super Admin and verify dashboard metrics, filters, tutor approval, assignment, activity log, and status changes.

### Task 3: Docs, Commit, Push

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `implementation-summary.md`

- [x] **Step 1: Run full verification**

Run `go test ./...`, seed, server, and browser cases for `TC-SA-1`, `TC-SA-2`, `TC-SA-3`, `TC-SA-4`, `TC-SA-5`, `TC-SA-6`, and `TC-SA-8`.

- [x] **Step 2: Update docs**

Record exact checks and any partial cases.

- [ ] **Step 3: Commit and push**

Commit message: `feat: add super admin core management`
