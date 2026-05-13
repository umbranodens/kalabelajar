# Core Data Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the Phase 2 backend data foundation for users, tutors, students, tutor assignment, and activity logging.

**Architecture:** GORM models represent the core data tables from the tech spec, database migration registration keeps startup schema creation centralized, repositories wrap persistence, and services enforce business rules before writing data. This slice stays server-side and test-first so later dashboard handlers can call stable services.

**Tech Stack:** Go 1.23, GORM, PostgreSQL, testify.

---

### Task 1: Core Data Models and Migration Registration

**Files:**
- Create: `internal/models/core_data_test.go`
- Create: `internal/models/core_data.go`
- Modify: `internal/database/database_test.go`
- Modify: `internal/database/database.go`
- Modify: `cmd/server/main.go`

- [x] **Step 1: Write failing tests**

```go
func TestCoreDataModelsAssignUUIDsBeforeCreate(t *testing.T) {
	user := models.User{Email: "parent@example.com", Name: "Parent", Provider: models.ProviderLocal, IsActive: true}
	require.NoError(t, user.BeforeCreate(nil))
	require.NotEqual(t, uuid.Nil, user.ID)
}

func TestMigrationModelsIncludesCoreDataTables(t *testing.T) {
	modelTypes := database.MigrationModelNames()
	assert.Contains(t, modelTypes, "Role")
	assert.Contains(t, modelTypes, "User")
	assert.Contains(t, modelTypes, "Tutor")
	assert.Contains(t, modelTypes, "Student")
	assert.Contains(t, modelTypes, "ActivityLog")
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/models ./internal/database`
Expected: FAIL because the core data model and migration APIs do not exist.

- [x] **Step 3: Implement minimal model and migration code**

Define role constants, provider constants, `Role`, `User`, `Tutor`, `Student`, and `ActivityLog` GORM models. Add UUID assignment hooks. Add `database.MigrationModels`, `database.MigrationModelNames`, and `database.Migrate`.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/models ./internal/database`
Expected: PASS.

### Task 2: Repositories and Assignment Service

**Files:**
- Create: `internal/services/assignment_service_test.go`
- Create: `internal/services/assignment_service.go`
- Create: `internal/repositories/core_data.go`

- [x] **Step 1: Write failing tests**

```go
func TestAssignTutorToStudentUpdatesStudentAndLogsActivity(t *testing.T) {
	service := services.NewAssignmentService(students, tutors, logs)
	student, err := service.AssignTutorToStudent(ctx, services.AssignTutorInput{ActorUserID: actorID, StudentID: studentID, TutorID: tutorID})
	require.NoError(t, err)
	assert.Equal(t, tutorID, *student.AssignedTutorID)
	assert.Equal(t, models.ActivityAssignTutor, logs.entries[0].Action)
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/services`
Expected: FAIL because assignment service and repository contracts are missing.

- [x] **Step 3: Implement service and GORM repositories**

Create repository interfaces used by `AssignmentService` and GORM-backed repositories for users, tutors, students, and activity logs. Enforce that assigned tutors have role Tutor, active account status, and verified tutor status.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/services ./internal/repositories`
Expected: PASS.

### Task 3: Documentation, Verification, Commit, Push

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `implementation-summary.md`

- [x] **Step 1: Run full verification**

Run: `go test ./...`, start the server, and request `GET /healthz`.
Expected: tests pass and health endpoint returns `ok`.

- [x] **Step 2: Update docs**

Add a newest changelog entry for the core data foundation and a newest implementation-summary entry with actual commit/push details.

- [ ] **Step 3: Commit and push**

Commit message: `feat: add core data foundation`
