# Kala Belajar Manual Test Walkthrough

This guide is for running the local MVP and checking the main browser flows.

## 1. Prerequisites

- PostgreSQL is running and reachable with the values in `.env`.
- Do not commit `.env`.
- From this folder: `C:\Users\armiariyan\Desktop\kalabelajar`

## 2. Prepare the App

```powershell
go test ./...
go run ./cmd/seed
go run ./cmd/server
```

Open:

```text
http://localhost:8080/login
```

## 3. Seeded Accounts

| Role | Email | Password |
| --- | --- | --- |
| Super Admin | `admin@kalabelajar.com` | `admin123` |
| Tutor | `tutor.math@kalabelajar.com` | `password123` |
| Tutor | `tutor.english@kalabelajar.com` | `password123` |
| Parent | `parent.satu@kalabelajar.com` | `password123` |
| Parent | `parent.dua@kalabelajar.com` | `password123` |

## 4. Super Admin Checks

1. Login as `admin@kalabelajar.com`.
2. Open `/dashboard`.
3. Confirm dashboard metrics include tutor, murid, parent, jadwal aktif, and pending approval.
4. Open `/admin/users` and test role/search/status filters.
5. Open `/admin/tutors` and verify/unverify state display.
6. Open `/admin/students` and assign a verified active tutor.
7. Open `/admin/schedules`.
8. Create a valid schedule.
9. Try invalid schedule time where end time is earlier than start time and confirm the error appears.
10. Toggle a schedule inactive and confirm Tutor/Parent active schedule views do not show it.
11. Open `/admin/lesson-sessions`.
12. Filter `completed`.
13. Open a published report detail.
14. Open a draft report detail.
15. Confirm completed sessions without reports show `Laporan belum tersedia`.
16. Open `/admin/activity` and confirm important actions are logged.

## 5. Tutor Checks

1. Login as `tutor.math@kalabelajar.com`.
2. Open `/dashboard`.
3. Open `/tutor/schedules` and confirm only that tutor's active schedules appear.
4. Open `/tutor/sessions`.
5. Open `Isi Laporan` for one session.
6. Fill status, materi, progres, PR, kendala, and saran latihan.
7. Keep `Terbitkan untuk Parent` checked and submit.
8. Confirm the session list shows the report as `Terbit`.

## 6. Parent Checks

1. Login as `parent.satu@kalabelajar.com`.
2. Open `/dashboard`.
3. Open `/parent/schedules`.
4. Confirm only active schedules for that parent's children appear.
5. Open `/parent/reports`.
6. Confirm published reports are visible.
7. Confirm draft report text such as `Sesi batal karena jadwal keluarga` is not visible.
8. Use the child filter and confirm reports stay scoped to that parent's child.

## 7. Mobile Checks

Use browser devtools responsive mode around `390 x 844`.

1. Login as Tutor and check `/tutor/sessions`.
2. Open the report form and confirm fields are comfortable to type in.
3. Login as Parent and check `/parent/reports`.
4. Confirm report cards are readable without horizontal scrolling.

## 8. Stop the App

Press `Ctrl+C` in the terminal running `go run ./cmd/server`.

