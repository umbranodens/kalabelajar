package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
)

const ActivitySetUserActive = "set_user_active"

type AdminTutorRepository interface {
	VerifyTutor(ctx context.Context, tutorID uuid.UUID) error
}

type AdminUserRepository interface {
	SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error
}

type AdminService struct {
	tutors AdminTutorRepository
	users  AdminUserRepository
	logs   ActivityLogger
}

type AdminActionInput struct {
	ActorUserID uuid.UUID
	EntityID    uuid.UUID
	IPAddress   string
}

type SetUserActiveInput struct {
	ActorUserID uuid.UUID
	UserID      uuid.UUID
	IsActive    bool
	IPAddress   string
}

func NewAdminService(tutors AdminTutorRepository, users AdminUserRepository, logs ActivityLogger) *AdminService {
	return &AdminService{tutors: tutors, users: users, logs: logs}
}

func (s *AdminService) VerifyTutor(ctx context.Context, input AdminActionInput) error {
	if input.ActorUserID == uuid.Nil || input.EntityID == uuid.Nil {
		return ErrInvalidAssignmentInput
	}
	if err := s.tutors.VerifyTutor(ctx, input.EntityID); err != nil {
		return fmt.Errorf("verify tutor: %w", err)
	}
	entityID := input.EntityID
	return s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     models.ActivityVerifyTutor,
		EntityType: models.EntityTutor,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"tutor_id": input.EntityID.String(),
		},
	})
}

func (s *AdminService) SetUserActive(ctx context.Context, input SetUserActiveInput) error {
	if input.ActorUserID == uuid.Nil || input.UserID == uuid.Nil {
		return ErrInvalidAssignmentInput
	}
	if err := s.users.SetUserActive(ctx, input.UserID, input.IsActive); err != nil {
		return fmt.Errorf("set user active: %w", err)
	}
	entityID := input.UserID
	return s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     ActivitySetUserActive,
		EntityType: models.EntityUser,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"is_active": input.IsActive,
		},
	})
}
