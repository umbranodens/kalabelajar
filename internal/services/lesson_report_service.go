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
	ErrLessonReportInputRequired    = errors.New("lesson report input is incomplete")
	ErrLessonReportMaterialRequired = errors.New("material summary is required for published reports")
	ErrLessonSessionStatusInvalid   = errors.New("lesson session status is invalid")
	ErrLessonSessionOwnership       = errors.New("lesson session does not belong to tutor")
)

type LessonReportRepository interface {
	FindLessonSessionForReport(ctx context.Context, id uuid.UUID) (*models.LessonSession, error)
	SaveLessonReport(ctx context.Context, report *models.LessonReport, status string) (*models.LessonReport, error)
}

type LessonReportService struct {
	reports LessonReportRepository
	logs    ActivityLogger
}

type SaveLessonReportInput struct {
	ActorUserID            uuid.UUID
	LessonSessionID        uuid.UUID
	TutorID                uuid.UUID
	Status                 string
	MaterialSummary        string
	ProgressSummary        string
	Homework               string
	IssueNotes             string
	HomePracticeSuggestion string
	Publish                bool
	IPAddress              string
}

func NewLessonReportService(reports LessonReportRepository, logs ActivityLogger) *LessonReportService {
	return &LessonReportService{reports: reports, logs: logs}
}

func (s *LessonReportService) SaveReport(ctx context.Context, input SaveLessonReportInput) (*models.LessonReport, error) {
	if input.ActorUserID == uuid.Nil || input.LessonSessionID == uuid.Nil || input.TutorID == uuid.Nil {
		return nil, ErrLessonReportInputRequired
	}
	if !isValidLessonSessionStatus(input.Status) {
		return nil, ErrLessonSessionStatusInvalid
	}
	if input.Publish && strings.TrimSpace(input.MaterialSummary) == "" {
		return nil, ErrLessonReportMaterialRequired
	}

	session, err := s.reports.FindLessonSessionForReport(ctx, input.LessonSessionID)
	if err != nil {
		return nil, fmt.Errorf("find lesson session: %w", err)
	}
	if session.TutorID != input.TutorID {
		return nil, ErrLessonSessionOwnership
	}

	var publishedAt *time.Time
	if input.Publish {
		now := time.Now().UTC()
		publishedAt = &now
	}
	report := &models.LessonReport{
		LessonSessionID:        session.ID,
		TutorID:                session.TutorID,
		StudentID:              session.StudentID,
		MaterialSummary:        strings.TrimSpace(input.MaterialSummary),
		ProgressSummary:        optionalScheduleString(input.ProgressSummary),
		Homework:               optionalScheduleString(input.Homework),
		IssueNotes:             optionalScheduleString(input.IssueNotes),
		HomePracticeSuggestion: optionalScheduleString(input.HomePracticeSuggestion),
		PublishedAt:            publishedAt,
	}
	saved, err := s.reports.SaveLessonReport(ctx, report, input.Status)
	if err != nil {
		return nil, err
	}

	entityID := saved.ID
	if err := s.logs.CreateActivityLog(ctx, &models.ActivityLog{
		UserID:     input.ActorUserID,
		Action:     models.ActivitySaveLessonReport,
		EntityType: models.EntityLessonReport,
		EntityID:   &entityID,
		IPAddress:  input.IPAddress,
		Metadata: map[string]any{
			"lesson_session_id": session.ID.String(),
			"status":            input.Status,
			"published":         input.Publish,
		},
	}); err != nil {
		return nil, fmt.Errorf("log lesson report activity: %w", err)
	}

	return saved, nil
}

func isValidLessonSessionStatus(status string) bool {
	switch status {
	case models.LessonSessionScheduled,
		models.LessonSessionCompleted,
		models.LessonSessionCanceled,
		models.LessonSessionStudentAbsent,
		models.LessonSessionFollowUpRequired,
		models.LessonSessionRescheduled:
		return true
	default:
		return false
	}
}
