package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
)

var (
	ErrInvalidAssignmentInput = errors.New("assignment input is incomplete")
	ErrTutorNotAssignable     = errors.New("tutor is not assignable")
)

type StudentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Student, error)
	UpdateAssignedTutor(ctx context.Context, studentID uuid.UUID, tutorID uuid.UUID) (*models.Student, error)
}

type TutorRepository interface {
	FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Tutor, error)
}

type ActivityLogger interface {
	CreateActivityLog(ctx context.Context, entry *models.ActivityLog) error
}

type AssignmentService struct {
	students StudentRepository
	tutors   TutorRepository
	logs     ActivityLogger
}

type AssignTutorInput struct {
	ActorUserID uuid.UUID
	StudentID   uuid.UUID
	TutorID     uuid.UUID
	IPAddress   string
}

func NewAssignmentService(students StudentRepository, tutors TutorRepository, logs ActivityLogger) *AssignmentService {
	return &AssignmentService{students: students, tutors: tutors, logs: logs}
}

func (s *AssignmentService) AssignTutorToStudent(ctx context.Context, input AssignTutorInput) (*models.Student, error) {
	if input.ActorUserID == uuid.Nil || input.StudentID == uuid.Nil || input.TutorID == uuid.Nil {
		return nil, ErrInvalidAssignmentInput
	}

	student, err := s.students.FindByID(ctx, input.StudentID)
	if err != nil {
		return nil, fmt.Errorf("find student: %w", err)
	}

	tutor, err := s.tutors.FindByIDWithUser(ctx, input.TutorID)
	if err != nil {
		return nil, fmt.Errorf("find tutor: %w", err)
	}

	if err := ensureTutorAssignable(tutor); err != nil {
		return nil, err
	}

	updatedStudent, err := s.students.UpdateAssignedTutor(ctx, student.ID, tutor.ID)
	if err != nil {
		return nil, fmt.Errorf("assign tutor to student: %w", err)
	}

	entityID := updatedStudent.ID
	if err := s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     models.ActivityAssignTutor,
		EntityType: models.EntityStudent,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"student_id":        updatedStudent.ID.String(),
			"assigned_tutor_id": tutor.ID.String(),
		},
	}); err != nil {
		return nil, fmt.Errorf("log assign tutor activity: %w", err)
	}

	return updatedStudent, nil
}

func ensureTutorAssignable(tutor *models.Tutor) error {
	if tutor == nil {
		return fmt.Errorf("%w: tutor not found", ErrTutorNotAssignable)
	}
	if tutor.User.RoleID == nil || *tutor.User.RoleID != models.RoleIDTutor {
		return fmt.Errorf("%w: user must have tutor role", ErrTutorNotAssignable)
	}
	if !tutor.User.IsActive {
		return fmt.Errorf("%w: tutor user must be active", ErrTutorNotAssignable)
	}
	if !tutor.IsVerified {
		return fmt.Errorf("%w: tutor must be verified", ErrTutorNotAssignable)
	}
	return nil
}
