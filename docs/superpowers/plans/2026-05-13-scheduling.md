# Phase 3 - Scheduling Slice

## Goal

Implement recurring lesson schedules from the tech spec:

- `schedules` model and migration.
- Backend validation for tutor, student, day, time range, and inactive visibility.
- Super Admin schedule management page with create and inactive actions.
- Tutor-only schedule view for own active schedules.
- Parent-only schedule view for active schedules belonging to their children.
- Activity logging for schedule create/inactivate.

## Test First

- Add `ScheduleService` tests for valid create, invalid time range, inactive action, and active visibility filtering.
- Run the new tests before implementation to confirm the missing service/model failure.

## Implementation Steps

1. Add schedule constants/model and migration registration.
2. Add schedule repository methods in `internal/repositories`.
3. Add schedule service validation and activity logging in `internal/services`.
4. Seed local active and inactive schedules.
5. Add admin schedule routes/templates.
6. Add tutor and parent schedule views with backend ownership filters.
7. Update dashboard links/metrics and shared admin navigation.
8. Verify automated tests and browser QA for `TC-SA-7.*`, tutor schedule ACs, and parent schedule ACs.
9. Update `CHANGELOG.md` and `implementation-summary.md`.
10. Commit and push.

