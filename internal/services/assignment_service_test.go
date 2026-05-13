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

func TestAssignTutorToStudentUpdatesStudentAndLogsActivity(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	studentID := uuid.New()
	tutorID := uuid.New()
	tutorRole := models.RoleIDTutor

	students := &fakeStudentRepository{
		student: models.Student{ID: studentID, ParentID: uuid.New(), Name: "Alya"},
	}
	tutors := &fakeTutorRepository{
		tutor: models.Tutor{
			ID:         tutorID,
			UserID:     uuid.New(),
			IsVerified: true,
			User:       models.User{ID: uuid.New(), RoleID: &tutorRole, IsActive: true},
		},
	}
	logs := &fakeActivityLogger{}
	service := services.NewAssignmentService(students, tutors, logs)

	student, err := service.AssignTutorToStudent(ctx, services.AssignTutorInput{
		ActorUserID: actorID,
		StudentID:   studentID,
		TutorID:     tutorID,
		IPAddress:   "127.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, student.AssignedTutorID)
	assert.Equal(t, tutorID, *student.AssignedTutorID)
	assert.Equal(t, tutorID, students.assignedTutorID)
	require.Len(t, logs.entries, 1)
	assert.Equal(t, actorID, logs.entries[0].UserID)
	assert.Equal(t, models.ActivityAssignTutor, logs.entries[0].Action)
	assert.Equal(t, models.EntityStudent, logs.entries[0].EntityType)
	require.NotNil(t, logs.entries[0].EntityID)
	assert.Equal(t, studentID, *logs.entries[0].EntityID)
}

func TestAssignTutorToStudentRejectsUnverifiedTutor(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	tutorID := uuid.New()
	tutorRole := models.RoleIDTutor
	students := &fakeStudentRepository{
		student: models.Student{ID: studentID, ParentID: uuid.New(), Name: "Alya"},
	}
	tutors := &fakeTutorRepository{
		tutor: models.Tutor{
			ID:         tutorID,
			UserID:     uuid.New(),
			IsVerified: false,
			User:       models.User{ID: uuid.New(), RoleID: &tutorRole, IsActive: true},
		},
	}
	logs := &fakeActivityLogger{}
	service := services.NewAssignmentService(students, tutors, logs)

	student, err := service.AssignTutorToStudent(ctx, services.AssignTutorInput{
		ActorUserID: uuid.New(),
		StudentID:   studentID,
		TutorID:     tutorID,
	})

	require.Error(t, err)
	assert.Nil(t, student)
	assert.Contains(t, err.Error(), "verified")
	assert.False(t, students.updated)
	assert.Empty(t, logs.entries)
}

func TestAssignTutorToStudentRejectsInactiveTutorUser(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	tutorID := uuid.New()
	tutorRole := models.RoleIDTutor
	service := services.NewAssignmentService(
		&fakeStudentRepository{student: models.Student{ID: studentID, ParentID: uuid.New(), Name: "Alya"}},
		&fakeTutorRepository{tutor: models.Tutor{
			ID:         tutorID,
			UserID:     uuid.New(),
			IsVerified: true,
			User:       models.User{ID: uuid.New(), RoleID: &tutorRole, IsActive: false},
		}},
		&fakeActivityLogger{},
	)

	student, err := service.AssignTutorToStudent(ctx, services.AssignTutorInput{
		ActorUserID: uuid.New(),
		StudentID:   studentID,
		TutorID:     tutorID,
	})

	require.Error(t, err)
	assert.Nil(t, student)
	assert.Contains(t, err.Error(), "active")
}

type fakeStudentRepository struct {
	student         models.Student
	assignedTutorID uuid.UUID
	updated         bool
	err             error
}

func (r *fakeStudentRepository) FindByID(_ context.Context, id uuid.UUID) (*models.Student, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.student.ID != id {
		return nil, errors.New("student not found")
	}
	return &r.student, nil
}

func (r *fakeStudentRepository) UpdateAssignedTutor(_ context.Context, studentID uuid.UUID, tutorID uuid.UUID) (*models.Student, error) {
	r.updated = true
	r.assignedTutorID = tutorID
	if r.student.ID != studentID {
		return nil, errors.New("student not found")
	}
	r.student.AssignedTutorID = &tutorID
	return &r.student, nil
}

type fakeTutorRepository struct {
	tutor models.Tutor
	err   error
}

func (r *fakeTutorRepository) FindByIDWithUser(_ context.Context, id uuid.UUID) (*models.Tutor, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.tutor.ID != id {
		return nil, errors.New("tutor not found")
	}
	return &r.tutor, nil
}

type fakeActivityLogger struct {
	entries []models.ActivityLog
	err     error
}

func (l *fakeActivityLogger) CreateActivityLog(_ context.Context, entry *models.ActivityLog) error {
	if l.err != nil {
		return l.err
	}
	l.entries = append(l.entries, *entry)
	return nil
}
