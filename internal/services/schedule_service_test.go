package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
)

func TestScheduleServiceCreateValidatesAndLogsActivity(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	tutorID := uuid.New()
	studentID := uuid.New()
	tutorRole := models.RoleIDTutor
	repo := &fakeScheduleRepository{
		tutor: models.Tutor{
			ID:         tutorID,
			IsVerified: true,
			User:       models.User{ID: uuid.New(), RoleID: &tutorRole, IsActive: true},
		},
		student: models.Student{ID: studentID, ParentID: uuid.New(), Name: "Alya", AssignedTutorID: &tutorID},
	}
	logs := &fakeScheduleActivityLogger{}
	service := services.NewScheduleService(repo, logs)

	schedule, err := service.CreateSchedule(ctx, services.CreateScheduleInput{
		ActorUserID: actorID,
		TutorID:     tutorID,
		StudentID:   studentID,
		Subject:     "Matematika",
		DayOfWeek:   1,
		StartTime:   "15:00",
		EndTime:     "16:30",
		Location:    "Rumah Alya",
		IPAddress:   "127.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.True(t, schedule.IsActive)
	assert.Equal(t, models.ActivityCreateSchedule, logs.entries[0].Action)
	assert.Equal(t, models.EntitySchedule, logs.entries[0].EntityType)
}

func TestScheduleServiceRejectsInvalidTimeRange(t *testing.T) {
	service := services.NewScheduleService(&fakeScheduleRepository{}, &fakeScheduleActivityLogger{})

	schedule, err := service.CreateSchedule(context.Background(), services.CreateScheduleInput{
		ActorUserID: uuid.New(),
		TutorID:     uuid.New(),
		StudentID:   uuid.New(),
		DayOfWeek:   2,
		StartTime:   "17:00",
		EndTime:     "16:00",
	})

	require.ErrorIs(t, err, services.ErrInvalidScheduleTime)
	assert.Nil(t, schedule)
}

func TestScheduleServiceRejectsStudentAssignedToDifferentTutor(t *testing.T) {
	tutorID := uuid.New()
	otherTutorID := uuid.New()
	tutorRole := models.RoleIDTutor
	service := services.NewScheduleService(&fakeScheduleRepository{
		tutor: models.Tutor{
			ID:         tutorID,
			IsVerified: true,
			User:       models.User{ID: uuid.New(), RoleID: &tutorRole, IsActive: true},
		},
		student: models.Student{ID: uuid.New(), ParentID: uuid.New(), Name: "Bima", AssignedTutorID: &otherTutorID},
	}, &fakeScheduleActivityLogger{})

	schedule, err := service.CreateSchedule(context.Background(), services.CreateScheduleInput{
		ActorUserID: uuid.New(),
		TutorID:     tutorID,
		StudentID:   uuid.New(),
		DayOfWeek:   3,
		StartTime:   "15:00",
		EndTime:     "16:00",
	})

	require.ErrorIs(t, err, services.ErrScheduleStudentTutorMismatch)
	assert.Nil(t, schedule)
}

func TestScheduleServiceDeactivateLogsActivity(t *testing.T) {
	scheduleID := uuid.New()
	logs := &fakeScheduleActivityLogger{}
	repo := &fakeScheduleRepository{schedule: models.Schedule{ID: scheduleID, IsActive: true}}
	service := services.NewScheduleService(repo, logs)

	err := service.SetScheduleActive(context.Background(), services.SetScheduleActiveInput{
		ActorUserID: uuid.New(),
		ScheduleID:  scheduleID,
		IsActive:    false,
	})

	require.NoError(t, err)
	assert.False(t, repo.schedule.IsActive)
	assert.Equal(t, models.ActivityUpdateScheduleStatus, logs.entries[0].Action)
	assert.Equal(t, scheduleID, *logs.entries[0].EntityID)
}

type fakeScheduleRepository struct {
	tutor    models.Tutor
	student  models.Student
	schedule models.Schedule
}

func (r *fakeScheduleRepository) FindTutorForSchedule(ctx context.Context, id uuid.UUID) (*models.Tutor, error) {
	if r.tutor.ID == uuid.Nil {
		return nil, errors.New("not found")
	}
	return &r.tutor, nil
}

func (r *fakeScheduleRepository) FindStudentForSchedule(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	if r.student.ID == uuid.Nil {
		return nil, errors.New("not found")
	}
	return &r.student, nil
}

func (r *fakeScheduleRepository) CreateSchedule(ctx context.Context, schedule *models.Schedule) (*models.Schedule, error) {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	r.schedule = *schedule
	return schedule, nil
}

func (r *fakeScheduleRepository) SetScheduleActive(ctx context.Context, scheduleID uuid.UUID, isActive bool) (*models.Schedule, error) {
	if r.schedule.ID == uuid.Nil {
		r.schedule.ID = scheduleID
	}
	r.schedule.IsActive = isActive
	return &r.schedule, nil
}

type fakeScheduleActivityLogger struct {
	entries []models.ActivityLog
}

func (l *fakeScheduleActivityLogger) CreateActivityLog(ctx context.Context, entry *models.ActivityLog) error {
	l.entries = append(l.entries, *entry)
	return nil
}
