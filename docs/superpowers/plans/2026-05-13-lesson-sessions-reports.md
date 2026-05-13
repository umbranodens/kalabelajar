# Phase 4 - Lesson Sessions and Reports Slice

## Goal

Implement the MVP lesson session and report flow:

- `lesson_sessions` and `lesson_reports` models and migration.
- Seed all lesson session statuses plus published, draft, and missing-report cases.
- Tutor report form for own sessions.
- Parent progress reports that only show published reports for their children.
- Super Admin monitoring for all sessions and report detail, including draft reports.
- Backend validation for session ownership, status enum, and published material summary.

## Test First

- Add service tests for published report material validation.
- Add service tests for report save updating session status.
- Add service tests rejecting reports against sessions owned by another tutor.

## Implementation Steps

1. Add model constants and structs.
2. Add migration registration and model test coverage.
3. Add repository methods for report/session write flow.
4. Add report service validation and activity logging.
5. Seed lesson sessions and reports.
6. Add Tutor sessions/report routes and templates.
7. Add Parent reports route and template with published-only filtering.
8. Add Super Admin session/report monitoring routes and templates.
9. Run automated tests, seed, and browser QA.
10. Update docs, commit, and push.

