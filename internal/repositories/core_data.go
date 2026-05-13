package repositories

import (
	"context"
	"fmt"

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
