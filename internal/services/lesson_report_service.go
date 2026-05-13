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
	ErrStudentNotAssignedToTutor    = errors.New("student is not assigned to this tutor")
)

type LessonReportRepository interface {
	FindLessonSessionForReport(ctx context.Context, id uuid.UUID) (*models.LessonSession, error)
	SaveLessonReport(ctx context.Context, report *models.LessonReport, status string) (*models.LessonReport, error)
	FindStudentByIDForTutor(ctx context.Context, studentID uuid.UUID, tutorID uuid.UUID) (*models.Student, error)
	CreateLessonSessionAndReport(ctx context.Context, session *models.LessonSession, report *models.LessonReport) (*models.LessonReport, error)
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

type CreateReportWithSessionInput struct {
	ActorUserID            uuid.UUID
	TutorID                uuid.UUID
	StudentID              uuid.UUID
	Subject                string
	Date                   string // "2006-01-02"
	StartTime              string // "15:04"
	EndTime                string // "16:30"
	Location               string
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

func (s *LessonReportService) CreateReportWithSession(ctx context.Context, input CreateReportWithSessionInput) (*models.LessonReport, error) {
	if input.ActorUserID == uuid.Nil || input.TutorID == uuid.Nil || input.StudentID == uuid.Nil {
		return nil, ErrLessonReportInputRequired
	}
	if !isValidLessonSessionStatus(input.Status) {
		return nil, ErrLessonSessionStatusInvalid
	}
	if input.Publish && strings.TrimSpace(input.MaterialSummary) == "" {
		return nil, ErrLessonReportMaterialRequired
	}
	if err := validateScheduleTime(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	dateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(input.Date))
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	startParsed, _ := time.Parse("15:04", strings.TrimSpace(input.StartTime))
	endParsed, _ := time.Parse("15:04", strings.TrimSpace(input.EndTime))
	scheduledStart := time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), startParsed.Hour(), startParsed.Minute(), 0, 0, time.UTC)
	scheduledEnd := time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), endParsed.Hour(), endParsed.Minute(), 0, 0, time.UTC)

	student, err := s.reports.FindStudentByIDForTutor(ctx, input.StudentID, input.TutorID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStudentNotAssignedToTutor, err)
	}

	var publishedAt *time.Time
	if input.Publish {
		now := time.Now().UTC()
		publishedAt = &now
	}

	session := &models.LessonSession{
		TutorID:          input.TutorID,
		StudentID:        student.ID,
		Subject:          optionalScheduleString(input.Subject),
		ScheduledStartAt: scheduledStart,
		ScheduledEndAt:   scheduledEnd,
		Status:           input.Status,
		Location:         optionalScheduleString(input.Location),
	}
	report := &models.LessonReport{
		TutorID:                input.TutorID,
		StudentID:              student.ID,
		MaterialSummary:        strings.TrimSpace(input.MaterialSummary),
		ProgressSummary:        optionalScheduleString(input.ProgressSummary),
		Homework:               optionalScheduleString(input.Homework),
		IssueNotes:             optionalScheduleString(input.IssueNotes),
		HomePracticeSuggestion: optionalScheduleString(input.HomePracticeSuggestion),
		PublishedAt:            publishedAt,
	}

	saved, err := s.reports.CreateLessonSessionAndReport(ctx, session, report)
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
			"student_id":        student.ID.String(),
			"status":            input.Status,
			"published":         input.Publish,
			"self_service":      true,
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
