package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
)

var (
	ErrInvalidScheduleInput         = errors.New("schedule input is incomplete")
	ErrInvalidScheduleDay           = errors.New("schedule day must be 0 through 6")
	ErrInvalidScheduleTime          = errors.New("schedule time is invalid")
	ErrScheduleStudentTutorMismatch = errors.New("student is not assigned to tutor")
)

type ScheduleRepository interface {
	FindTutorForSchedule(ctx context.Context, id uuid.UUID) (*models.Tutor, error)
	FindStudentForSchedule(ctx context.Context, id uuid.UUID) (*models.Student, error)
	CreateSchedule(ctx context.Context, schedule *models.Schedule) (*models.Schedule, error)
	SetScheduleActive(ctx context.Context, scheduleID uuid.UUID, isActive bool) (*models.Schedule, error)
}

type ScheduleService struct {
	schedules ScheduleRepository
	logs      ActivityLogger
}

type CreateScheduleInput struct {
	ActorUserID uuid.UUID
	TutorID     uuid.UUID
	StudentID   uuid.UUID
	Subject     string
	DayOfWeek   int
	StartTime   string
	EndTime     string
	Location    string
	IPAddress   string
}

type SetScheduleActiveInput struct {
	ActorUserID uuid.UUID
	ScheduleID  uuid.UUID
	IsActive    bool
	IPAddress   string
}

func NewScheduleService(schedules ScheduleRepository, logs ActivityLogger) *ScheduleService {
	return &ScheduleService{schedules: schedules, logs: logs}
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*models.Schedule, error) {
	if input.ActorUserID == uuid.Nil || input.TutorID == uuid.Nil || input.StudentID == uuid.Nil {
		return nil, ErrInvalidScheduleInput
	}
	if input.DayOfWeek < 0 || input.DayOfWeek > 6 {
		return nil, ErrInvalidScheduleDay
	}
	if err := validateScheduleTime(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	tutor, err := s.schedules.FindTutorForSchedule(ctx, input.TutorID)
	if err != nil {
		return nil, fmt.Errorf("find tutor: %w", err)
	}
	if err := ensureTutorAssignable(tutor); err != nil {
		return nil, err
	}

	student, err := s.schedules.FindStudentForSchedule(ctx, input.StudentID)
	if err != nil {
		return nil, fmt.Errorf("find student: %w", err)
	}
	if student.AssignedTutorID == nil || *student.AssignedTutorID != tutor.ID {
		return nil, ErrScheduleStudentTutorMismatch
	}

	schedule := &models.Schedule{
		TutorID:   tutor.ID,
		StudentID: student.ID,
		Subject:   optionalScheduleString(input.Subject),
		DayOfWeek: input.DayOfWeek,
		StartTime: strings.TrimSpace(input.StartTime),
		EndTime:   strings.TrimSpace(input.EndTime),
		Location:  optionalScheduleString(input.Location),
		IsActive:  true,
	}
	created, err := s.schedules.CreateSchedule(ctx, schedule)
	if err != nil {
		return nil, err
	}

	entityID := created.ID
	if err := s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     models.ActivityCreateSchedule,
		EntityType: models.EntitySchedule,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"tutor_id":    tutor.ID.String(),
			"student_id":  student.ID.String(),
			"day_of_week": input.DayOfWeek,
			"start_time":  created.StartTime,
			"end_time":    created.EndTime,
		},
	}); err != nil {
		return nil, fmt.Errorf("log create schedule activity: %w", err)
	}

	return created, nil
}

func (s *ScheduleService) SetScheduleActive(ctx context.Context, input SetScheduleActiveInput) error {
	if input.ActorUserID == uuid.Nil || input.ScheduleID == uuid.Nil {
		return ErrInvalidScheduleInput
	}
	schedule, err := s.schedules.SetScheduleActive(ctx, input.ScheduleID, input.IsActive)
	if err != nil {
		return err
	}
	entityID := schedule.ID
	if err := s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     models.ActivityUpdateScheduleStatus,
		EntityType: models.EntitySchedule,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"is_active": input.IsActive,
		},
	}); err != nil {
		return fmt.Errorf("log schedule status activity: %w", err)
	}
	return nil
}

func validateScheduleTime(startTime string, endTime string) error {
	start, err := time.Parse("15:04", strings.TrimSpace(startTime))
	if err != nil {
		return fmt.Errorf("%w: start time must use HH:MM", ErrInvalidScheduleTime)
	}
	end, err := time.Parse("15:04", strings.TrimSpace(endTime))
	if err != nil {
		return fmt.Errorf("%w: end time must use HH:MM", ErrInvalidScheduleTime)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: end time must be after start time", ErrInvalidScheduleTime)
	}
	return nil
}

func optionalScheduleString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
