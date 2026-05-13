package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
)

func TestLessonReportServicePublishesReportAndUpdatesSessionStatus(t *testing.T) {
	tutorID := uuid.New()
	studentID := uuid.New()
	sessionID := uuid.New()
	repo := &fakeLessonReportRepository{
		session: models.LessonSession{ID: sessionID, TutorID: tutorID, StudentID: studentID, Status: models.LessonSessionScheduled},
	}
	logs := &fakeScheduleActivityLogger{}
	service := services.NewLessonReportService(repo, logs)

	report, err := service.SaveReport(context.Background(), services.SaveLessonReportInput{
		ActorUserID:            uuid.New(),
		LessonSessionID:        sessionID,
		TutorID:                tutorID,
		Status:                 models.LessonSessionCompleted,
		MaterialSummary:        "Pecahan campuran",
		ProgressSummary:        "Alya mulai stabil.",
		Homework:               "Latihan 5 soal.",
		IssueNotes:             "Butuh ulang pembagian.",
		HomePracticeSuggestion: "Review 15 menit.",
		Publish:                true,
	})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.NotNil(t, report.PublishedAt)
	assert.Equal(t, models.LessonSessionCompleted, repo.session.Status)
	assert.Equal(t, models.ActivitySaveLessonReport, logs.entries[0].Action)
}

func TestLessonReportServiceRejectsPublishedReportWithoutMaterial(t *testing.T) {
	service := services.NewLessonReportService(&fakeLessonReportRepository{}, &fakeScheduleActivityLogger{})

	report, err := service.SaveReport(context.Background(), services.SaveLessonReportInput{
		ActorUserID:     uuid.New(),
		LessonSessionID: uuid.New(),
		TutorID:         uuid.New(),
		Status:          models.LessonSessionCompleted,
		Publish:         true,
	})

	require.ErrorIs(t, err, services.ErrLessonReportMaterialRequired)
	assert.Nil(t, report)
}

func TestLessonReportServiceRejectsSessionOwnedByAnotherTutor(t *testing.T) {
	tutorID := uuid.New()
	otherTutorID := uuid.New()
	sessionID := uuid.New()
	service := services.NewLessonReportService(&fakeLessonReportRepository{
		session: models.LessonSession{ID: sessionID, TutorID: otherTutorID, StudentID: uuid.New(), Status: models.LessonSessionScheduled},
	}, &fakeScheduleActivityLogger{})

	report, err := service.SaveReport(context.Background(), services.SaveLessonReportInput{
		ActorUserID:     uuid.New(),
		LessonSessionID: sessionID,
		TutorID:         tutorID,
		Status:          models.LessonSessionCompleted,
		MaterialSummary: "Materi",
		Publish:         true,
	})

	require.ErrorIs(t, err, services.ErrLessonSessionOwnership)
	assert.Nil(t, report)
}

type fakeLessonReportRepository struct {
	session models.LessonSession
	report  models.LessonReport
}

func (r *fakeLessonReportRepository) FindLessonSessionForReport(ctx context.Context, id uuid.UUID) (*models.LessonSession, error) {
	if r.session.ID == uuid.Nil {
		return nil, errors.New("not found")
	}
	return &r.session, nil
}

func (r *fakeLessonReportRepository) SaveLessonReport(ctx context.Context, report *models.LessonReport, status string) (*models.LessonReport, error) {
	r.session.Status = status
	report.ID = uuid.New()
	report.CreatedAt = time.Now()
	r.report = *report
	return report, nil
}

func (r *fakeLessonReportRepository) FindStudentByIDForTutor(ctx context.Context, studentID uuid.UUID, tutorID uuid.UUID) (*models.Student, error) {
	return &models.Student{ID: studentID, AssignedTutorID: &tutorID}, nil
}

func (r *fakeLessonReportRepository) CreateLessonSessionAndReport(ctx context.Context, session *models.LessonSession, report *models.LessonReport) (*models.LessonReport, error) {
	session.ID = uuid.New()
	report.ID = uuid.New()
	report.LessonSessionID = session.ID
	report.CreatedAt = time.Now()
	r.report = *report
	return report, nil
}
