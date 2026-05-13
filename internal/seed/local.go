package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
	"gorm.io/gorm"
)

func LocalDevelopmentData(ctx context.Context, db *gorm.DB) error {
	passwordHash, err := services.HashPassword("password123")
	if err != nil {
		return err
	}

	tutorRole := models.RoleIDTutor
	parentRole := models.RoleIDParent
	users := []models.User{
		{Email: "tutor.math@kalabelajar.com", Name: "Tutor Matematika", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &tutorRole, IsActive: true},
		{Email: "tutor.english@kalabelajar.com", Name: "Tutor Bahasa Inggris", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &tutorRole, IsActive: true},
		{Email: "tutor.inactive@kalabelajar.com", Name: "Tutor Nonaktif", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &tutorRole, IsActive: false},
		{Email: "parent.satu@kalabelajar.com", Name: "Parent Satu", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &parentRole, IsActive: true},
		{Email: "parent.dua@kalabelajar.com", Name: "Parent Dua", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &parentRole, IsActive: true},
	}
	for i := range users {
		if err := firstOrCreateUser(ctx, db, &users[i]); err != nil {
			return err
		}
	}

	mathTutor, err := ensureTutor(ctx, db, "tutor.math@kalabelajar.com", []string{"Matematika"}, true)
	if err != nil {
		return err
	}
	englishTutor, err := ensureTutor(ctx, db, "tutor.english@kalabelajar.com", []string{"Bahasa Inggris"}, true)
	if err != nil {
		return err
	}
	if _, err := ensureTutor(ctx, db, "tutor.inactive@kalabelajar.com", []string{"Fisika"}, false); err != nil {
		return err
	}

	parentOne, err := findUser(ctx, db, "parent.satu@kalabelajar.com")
	if err != nil {
		return err
	}
	parentTwo, err := findUser(ctx, db, "parent.dua@kalabelajar.com")
	if err != nil {
		return err
	}
	students := []models.Student{
		{ParentID: parentOne.ID, Name: "Alya Pratama", Grade: stringPtr("5 SD"), School: stringPtr("SD Harapan"), AssignedTutorID: &mathTutor.ID},
		{ParentID: parentTwo.ID, Name: "Bima Santoso", Grade: stringPtr("7 SMP"), School: stringPtr("SMP Nusantara"), AssignedTutorID: &englishTutor.ID},
		{ParentID: parentTwo.ID, Name: "Citra Lestari", Grade: stringPtr("4 SD"), School: stringPtr("SD Nusantara")},
	}
	for i := range students {
		if err := firstOrCreateStudent(ctx, db, &students[i]); err != nil {
			return err
		}
	}

	mathSchedule, err := firstOrCreateSchedule(ctx, db, models.Schedule{
		TutorID:   mathTutor.ID,
		StudentID: students[0].ID,
		Subject:   stringPtr("Matematika"),
		DayOfWeek: 1,
		StartTime: "15:00",
		EndTime:   "16:30",
		Location:  stringPtr("Rumah Alya"),
		IsActive:  true,
	})
	if err != nil {
		return err
	}
	englishSchedule, err := firstOrCreateSchedule(ctx, db, models.Schedule{
		TutorID:   englishTutor.ID,
		StudentID: students[1].ID,
		Subject:   stringPtr("Bahasa Inggris"),
		DayOfWeek: 3,
		StartTime: "18:00",
		EndTime:   "19:30",
		Location:  stringPtr("Online"),
		IsActive:  true,
	})
	if err != nil {
		return err
	}
	if _, err := firstOrCreateSchedule(ctx, db, models.Schedule{
		TutorID:   mathTutor.ID,
		StudentID: students[0].ID,
		Subject:   stringPtr("Matematika"),
		DayOfWeek: 6,
		StartTime: "10:00",
		EndTime:   "11:00",
		Location:  stringPtr("Rumah Alya"),
		IsActive:  false,
	}); err != nil {
		return err
	}

	if err := seedLessonSessionsAndReports(ctx, db, mathTutor, englishTutor, &students[0], &students[1], mathSchedule, englishSchedule); err != nil {
		return err
	}

	return nil
}

func firstOrCreateUser(ctx context.Context, db *gorm.DB, user *models.User) error {
	var existing models.User
	err := db.WithContext(ctx).Where("email = ?", user.Email).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed user %s: %w", user.Email, err)
	}
	return db.WithContext(ctx).Create(user).Error
}

func ensureTutor(ctx context.Context, db *gorm.DB, email string, subjects []string, verified bool) (*models.Tutor, error) {
	user, err := findUser(ctx, db, email)
	if err != nil {
		return nil, err
	}
	var tutor models.Tutor
	err = db.WithContext(ctx).Where("user_id = ?", user.ID).First(&tutor).Error
	if err == nil {
		tutor.Subjects = subjects
		tutor.IsVerified = verified
		return &tutor, db.WithContext(ctx).Save(&tutor).Error
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tutor = models.Tutor{UserID: user.ID, Subjects: subjects, IsVerified: verified}
	return &tutor, db.WithContext(ctx).Create(&tutor).Error
}

func findUser(ctx context.Context, db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func firstOrCreateStudent(ctx context.Context, db *gorm.DB, student *models.Student) error {
	var existing models.Student
	err := db.WithContext(ctx).Where("parent_id = ? AND name = ?", student.ParentID, student.Name).First(&existing).Error
	if err == nil {
		if err := db.WithContext(ctx).Model(&existing).Update("assigned_tutor_id", student.AssignedTutorID).Error; err != nil {
			return err
		}
		student.ID = existing.ID
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return db.WithContext(ctx).Create(student).Error
}

func firstOrCreateSchedule(ctx context.Context, db *gorm.DB, schedule models.Schedule) (*models.Schedule, error) {
	var existing models.Schedule
	err := db.WithContext(ctx).
		Where("tutor_id = ? AND student_id = ? AND day_of_week = ? AND start_time = ?", schedule.TutorID, schedule.StudentID, schedule.DayOfWeek, schedule.StartTime).
		First(&existing).Error
	if err == nil {
		updates := map[string]any{
			"subject":   schedule.Subject,
			"end_time":  schedule.EndTime,
			"location":  schedule.Location,
			"is_active": schedule.IsActive,
		}
		if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
			return nil, err
		}
		existing.Subject = schedule.Subject
		existing.EndTime = schedule.EndTime
		existing.Location = schedule.Location
		existing.IsActive = schedule.IsActive
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &schedule, db.WithContext(ctx).Create(&schedule).Error
}

func seedLessonSessionsAndReports(ctx context.Context, db *gorm.DB, mathTutor *models.Tutor, englishTutor *models.Tutor, alya *models.Student, bima *models.Student, mathSchedule *models.Schedule, englishSchedule *models.Schedule) error {
	base := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	sessions := []models.LessonSession{
		lessonSession(mathSchedule, mathTutor.ID, alya.ID, "Matematika", base.Add(15*time.Hour), base.Add(16*time.Hour+30*time.Minute), models.LessonSessionCompleted, "Rumah Alya"),
		lessonSession(mathSchedule, mathTutor.ID, alya.ID, "Matematika", base.AddDate(0, 0, 1).Add(15*time.Hour), base.AddDate(0, 0, 1).Add(16*time.Hour+30*time.Minute), models.LessonSessionScheduled, "Rumah Alya"),
		lessonSession(mathSchedule, mathTutor.ID, alya.ID, "Matematika", base.AddDate(0, 0, 2).Add(15*time.Hour), base.AddDate(0, 0, 2).Add(16*time.Hour+30*time.Minute), models.LessonSessionCanceled, "Rumah Alya"),
		lessonSession(englishSchedule, englishTutor.ID, bima.ID, "Bahasa Inggris", base.AddDate(0, 0, 3).Add(18*time.Hour), base.AddDate(0, 0, 3).Add(19*time.Hour+30*time.Minute), models.LessonSessionStudentAbsent, "Online"),
		lessonSession(englishSchedule, englishTutor.ID, bima.ID, "Bahasa Inggris", base.AddDate(0, 0, 4).Add(18*time.Hour), base.AddDate(0, 0, 4).Add(19*time.Hour+30*time.Minute), models.LessonSessionFollowUpRequired, "Online"),
		lessonSession(mathSchedule, mathTutor.ID, alya.ID, "Matematika", base.AddDate(0, 0, 5).Add(10*time.Hour), base.AddDate(0, 0, 5).Add(11*time.Hour), models.LessonSessionRescheduled, "Rumah Alya"),
		lessonSession(mathSchedule, mathTutor.ID, alya.ID, "Matematika", base.AddDate(0, 0, 6).Add(15*time.Hour), base.AddDate(0, 0, 6).Add(16*time.Hour+30*time.Minute), models.LessonSessionCompleted, "Rumah Alya"),
	}
	for i := range sessions {
		created, err := firstOrCreateLessonSession(ctx, db, &sessions[i])
		if err != nil {
			return err
		}
		sessions[i].ID = created.ID
	}
	publishedAt := base.AddDate(0, 0, 1)
	if err := firstOrCreateLessonReport(ctx, db, models.LessonReport{
		LessonSessionID:        sessions[0].ID,
		TutorID:                mathTutor.ID,
		StudentID:              alya.ID,
		MaterialSummary:        "Pecahan campuran dan penyederhanaan.",
		ProgressSummary:        stringPtr("Alya makin percaya diri menyelesaikan soal bertahap."),
		Homework:               stringPtr("Latihan 5 soal pecahan."),
		IssueNotes:             stringPtr("Masih perlu pelan-pelan saat pembagian."),
		HomePracticeSuggestion: stringPtr("Review 15 menit sebelum sesi berikutnya."),
		PublishedAt:            &publishedAt,
	}); err != nil {
		return err
	}
	if err := firstOrCreateLessonReport(ctx, db, models.LessonReport{
		LessonSessionID:        sessions[2].ID,
		TutorID:                mathTutor.ID,
		StudentID:              alya.ID,
		MaterialSummary:        "Sesi batal karena jadwal keluarga.",
		ProgressSummary:        stringPtr("Belum ada progres baru."),
		IssueNotes:             stringPtr("Perlu jadwal pengganti."),
		HomePracticeSuggestion: stringPtr("Kerjakan ulang catatan pekan lalu."),
	}); err != nil {
		return err
	}
	if err := firstOrCreateLessonReport(ctx, db, models.LessonReport{
		LessonSessionID:        sessions[4].ID,
		TutorID:                englishTutor.ID,
		StudentID:              bima.ID,
		MaterialSummary:        "Reading comprehension dan vocabulary.",
		ProgressSummary:        stringPtr("Bima memahami ide utama, tetapi perlu latihan kosakata."),
		Homework:               stringPtr("Baca satu artikel pendek."),
		IssueNotes:             stringPtr("Perlu tindak lanjut pronunciation."),
		HomePracticeSuggestion: stringPtr("Latihan membaca nyaring 10 menit."),
		PublishedAt:            &publishedAt,
	}); err != nil {
		return err
	}
	if err := deleteLessonReport(ctx, db, sessions[6].ID); err != nil {
		return err
	}
	return nil
}

func lessonSession(schedule *models.Schedule, tutorID uuid.UUID, studentID uuid.UUID, subject string, startAt time.Time, endAt time.Time, status string, location string) models.LessonSession {
	return models.LessonSession{
		ScheduleID:       &schedule.ID,
		TutorID:          tutorID,
		StudentID:        studentID,
		Subject:          stringPtr(subject),
		ScheduledStartAt: startAt,
		ScheduledEndAt:   endAt,
		Status:           status,
		Location:         stringPtr(location),
	}
}

func firstOrCreateLessonSession(ctx context.Context, db *gorm.DB, session *models.LessonSession) (*models.LessonSession, error) {
	var existing models.LessonSession
	err := db.WithContext(ctx).
		Where("tutor_id = ? AND student_id = ? AND scheduled_start_at = ?", session.TutorID, session.StudentID, session.ScheduledStartAt).
		First(&existing).Error
	if err == nil {
		updates := map[string]any{
			"schedule_id":      session.ScheduleID,
			"subject":          session.Subject,
			"scheduled_end_at": session.ScheduledEndAt,
			"status":           session.Status,
			"location":         session.Location,
		}
		if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return session, db.WithContext(ctx).Create(session).Error
}

func firstOrCreateLessonReport(ctx context.Context, db *gorm.DB, report models.LessonReport) error {
	var existing models.LessonReport
	err := db.WithContext(ctx).Where("lesson_session_id = ?", report.LessonSessionID).First(&existing).Error
	if err == nil {
		return db.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"tutor_id":                 report.TutorID,
			"student_id":               report.StudentID,
			"material_summary":         report.MaterialSummary,
			"progress_summary":         report.ProgressSummary,
			"homework":                 report.Homework,
			"issue_notes":              report.IssueNotes,
			"home_practice_suggestion": report.HomePracticeSuggestion,
			"published_at":             report.PublishedAt,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return db.WithContext(ctx).Create(&report).Error
}

func deleteLessonReport(ctx context.Context, db *gorm.DB, lessonSessionID uuid.UUID) error {
	return db.WithContext(ctx).Where("lesson_session_id = ?", lessonSessionID).Delete(&models.LessonReport{}).Error
}

func stringPtr(value string) *string {
	return &value
}
