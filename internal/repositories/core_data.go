package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) TouchLastLogin(ctx context.Context, userID uuid.UUID, when time.Time) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", when).Error; err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateOAuthProfile(ctx context.Context, user *models.User) error {
	updates := map[string]any{
		"name":        user.Name,
		"provider":    user.Provider,
		"provider_id": user.ProviderID,
		"avatar_url":  user.AvatarURL,
	}
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update oauth profile: %w", err)
	}
	return nil
}

func (r *UserRepository) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_active", isActive).Error; err != nil {
		return fmt.Errorf("set user active: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Role").First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

type TutorRepository struct {
	db *gorm.DB
}

func NewTutorRepository(db *gorm.DB) *TutorRepository {
	return &TutorRepository{db: db}
}

func (r *TutorRepository) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Tutor, error) {
	var tutor models.Tutor
	if err := r.db.WithContext(ctx).Preload("User").First(&tutor, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find tutor by id: %w", err)
	}
	return &tutor, nil
}

func (r *TutorRepository) FindTutorForSchedule(ctx context.Context, id uuid.UUID) (*models.Tutor, error) {
	return r.FindByIDWithUser(ctx, id)
}

func (r *TutorRepository) VerifyTutor(ctx context.Context, tutorID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Tutor{}).Where("id = ?", tutorID).Update("is_verified", true).Error; err != nil {
		return fmt.Errorf("verify tutor: %w", err)
	}
	return nil
}

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	var student models.Student
	if err := r.db.WithContext(ctx).First(&student, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find student by id: %w", err)
	}
	return &student, nil
}

func (r *StudentRepository) FindStudentForSchedule(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	var student models.Student
	if err := r.db.WithContext(ctx).Preload("Parent").First(&student, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find student for schedule: %w", err)
	}
	return &student, nil
}

func (r *StudentRepository) UpdateAssignedTutor(ctx context.Context, studentID uuid.UUID, tutorID uuid.UUID) (*models.Student, error) {
	var student models.Student
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&student, "id = ?", studentID).Error; err != nil {
			return err
		}
		student.AssignedTutorID = &tutorID
		return tx.Save(&student).Error
	}); err != nil {
		return nil, fmt.Errorf("update student assigned tutor: %w", err)
	}
	return &student, nil
}

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) FindTutorForSchedule(ctx context.Context, id uuid.UUID) (*models.Tutor, error) {
	var tutor models.Tutor
	if err := r.db.WithContext(ctx).Preload("User").First(&tutor, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find tutor for schedule: %w", err)
	}
	return &tutor, nil
}

func (r *ScheduleRepository) FindStudentForSchedule(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	var student models.Student
	if err := r.db.WithContext(ctx).Preload("Parent").First(&student, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find student for schedule: %w", err)
	}
	return &student, nil
}

func (r *ScheduleRepository) CreateSchedule(ctx context.Context, schedule *models.Schedule) (*models.Schedule, error) {
	if err := r.db.WithContext(ctx).Create(schedule).Error; err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}
	if err := r.db.WithContext(ctx).
		Preload("Tutor.User").
		Preload("Student.Parent").
		First(schedule, "id = ?", schedule.ID).Error; err != nil {
		return nil, fmt.Errorf("reload schedule: %w", err)
	}
	return schedule, nil
}

func (r *ScheduleRepository) SetScheduleActive(ctx context.Context, scheduleID uuid.UUID, isActive bool) (*models.Schedule, error) {
	var schedule models.Schedule
	if err := r.db.WithContext(ctx).First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return nil, fmt.Errorf("find schedule by id: %w", err)
	}
	schedule.IsActive = isActive
	if err := r.db.WithContext(ctx).Save(&schedule).Error; err != nil {
		return nil, fmt.Errorf("set schedule active: %w", err)
	}
	return &schedule, nil
}

type LessonReportRepository struct {
	db *gorm.DB
}

func NewLessonReportRepository(db *gorm.DB) *LessonReportRepository {
	return &LessonReportRepository{db: db}
}

func (r *LessonReportRepository) FindLessonSessionForReport(ctx context.Context, id uuid.UUID) (*models.LessonSession, error) {
	var session models.LessonSession
	if err := r.db.WithContext(ctx).
		Preload("Tutor.User").
		Preload("Student.Parent").
		Preload("Report").
		First(&session, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find lesson session for report: %w", err)
	}
	return &session, nil
}

func (r *LessonReportRepository) SaveLessonReport(ctx context.Context, report *models.LessonReport, status string) (*models.LessonReport, error) {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.LessonSession{}).Where("id = ?", report.LessonSessionID).Update("status", status).Error; err != nil {
			return err
		}
		var existing models.LessonReport
		err := tx.Where("lesson_session_id = ?", report.LessonSessionID).First(&existing).Error
		if err == nil {
			report.ID = existing.ID
			report.CreatedAt = existing.CreatedAt
			return tx.Model(&existing).Updates(map[string]any{
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
		return tx.Create(report).Error
	}); err != nil {
		return nil, fmt.Errorf("save lesson report: %w", err)
	}
	if err := r.db.WithContext(ctx).
		Preload("LessonSession").
		Preload("Tutor.User").
		Preload("Student.Parent").
		First(report, "lesson_session_id = ?", report.LessonSessionID).Error; err != nil {
		return nil, fmt.Errorf("reload lesson report: %w", err)
	}
	return report, nil
}

type ActivityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) CreateActivityLog(ctx context.Context, entry *models.ActivityLog) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("create activity log: %w", err)
	}
	return nil
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) UpsertRoles(ctx context.Context, roles []models.Role) error {
	if err := r.db.WithContext(ctx).Save(&roles).Error; err != nil {
		return fmt.Errorf("upsert roles: %w", err)
	}
	return nil
}

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindValidByToken(ctx context.Context, token string, now time.Time) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Role").
		Where("token = ? AND expires_at > ?", token, now).
		First(&session).Error; err != nil {
		return nil, fmt.Errorf("find valid session: %w", err)
	}
	return &session, nil
}
