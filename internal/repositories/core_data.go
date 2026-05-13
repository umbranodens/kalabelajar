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
