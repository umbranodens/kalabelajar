package seed

import (
	"context"
	"fmt"

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

	if err := firstOrCreateSchedule(ctx, db, models.Schedule{
		TutorID:   mathTutor.ID,
		StudentID: students[0].ID,
		Subject:   stringPtr("Matematika"),
		DayOfWeek: 1,
		StartTime: "15:00",
		EndTime:   "16:30",
		Location:  stringPtr("Rumah Alya"),
		IsActive:  true,
	}); err != nil {
		return err
	}
	if err := firstOrCreateSchedule(ctx, db, models.Schedule{
		TutorID:   englishTutor.ID,
		StudentID: students[1].ID,
		Subject:   stringPtr("Bahasa Inggris"),
		DayOfWeek: 3,
		StartTime: "18:00",
		EndTime:   "19:30",
		Location:  stringPtr("Online"),
		IsActive:  true,
	}); err != nil {
		return err
	}
	if err := firstOrCreateSchedule(ctx, db, models.Schedule{
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

func firstOrCreateSchedule(ctx context.Context, db *gorm.DB, schedule models.Schedule) error {
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
		return db.WithContext(ctx).Model(&existing).Updates(updates).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return db.WithContext(ctx).Create(&schedule).Error
}

func stringPtr(value string) *string {
	return &value
}
