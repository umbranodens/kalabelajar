package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleIDSuperAdmin uint = 1
	RoleIDTutor      uint = 2
	RoleIDParent     uint = 3
)

const (
	RoleNameSuperAdmin = "superadmin"
	RoleNameTutor      = "tutor"
	RoleNameParent     = "parent"
)

const (
	ProviderLocal  = "local"
	ProviderGoogle = "google"
)

const (
	ActivityLogin                = "login"
	ActivityAssignRole           = "assign_role"
	ActivityVerifyTutor          = "verify_tutor"
	ActivityAssignTutor          = "assign_tutor"
	ActivityCreateSchedule       = "create_schedule"
	ActivityUpdateScheduleStatus = "update_schedule_status"
	ActivitySaveLessonReport     = "save_lesson_report"
)

const (
	EntityStudent       = "student"
	EntityTutor         = "tutor"
	EntityUser          = "user"
	EntitySchedule      = "schedule"
	EntityLessonSession = "lesson_session"
	EntityLessonReport  = "lesson_report"
)

const (
	LessonSessionScheduled        = "scheduled"
	LessonSessionCompleted        = "completed"
	LessonSessionCanceled         = "canceled"
	LessonSessionStudentAbsent    = "student_absent"
	LessonSessionFollowUpRequired = "follow_up_required"
	LessonSessionRescheduled      = "rescheduled"
)

type Role struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(50);unique;not null"`
	DisplayName string    `gorm:"type:varchar(100);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	Users       []User
}

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string         `gorm:"type:varchar(255);unique;not null"`
	Name         string         `gorm:"type:varchar(255);not null"`
	AvatarURL    *string        `gorm:"type:varchar(512)"`
	Provider     string         `gorm:"type:varchar(50);default:local;not null"`
	ProviderID   *string        `gorm:"type:varchar(255)"`
	PasswordHash *string        `gorm:"type:text"`
	RoleID       *uint          `gorm:"index"`
	Role         *Role          `gorm:"foreignKey:RoleID"`
	IsActive     bool           `gorm:"default:true;not null"`
	LastLoginAt  *time.Time     `gorm:"type:timestamptz"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Tutor        *Tutor
	Students     []Student `gorm:"foreignKey:ParentID"`
	Sessions     []Session
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type Tutor struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;unique;not null"`
	User       User
	Phone      *string  `gorm:"type:varchar(20)"`
	Subjects   []string `gorm:"type:jsonb;serializer:json"`
	Bio        *string  `gorm:"type:text"`
	IsVerified bool     `gorm:"default:false;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Students   []Student `gorm:"foreignKey:AssignedTutorID"`
	Schedules  []Schedule
}

func (t *Tutor) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type Student struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ParentID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	Parent          User       `gorm:"foreignKey:ParentID"`
	Name            string     `gorm:"type:varchar(255);not null"`
	Grade           *string    `gorm:"type:varchar(50)"`
	School          *string    `gorm:"type:varchar(255)"`
	AssignedTutorID *uuid.UUID `gorm:"type:uuid;index"`
	AssignedTutor   *Tutor     `gorm:"foreignKey:AssignedTutorID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Schedules       []Schedule
}

func (s *Student) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type Schedule struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TutorID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Tutor     Tutor     `gorm:"foreignKey:TutorID"`
	StudentID uuid.UUID `gorm:"type:uuid;not null;index"`
	Student   Student   `gorm:"foreignKey:StudentID"`
	Subject   *string   `gorm:"type:varchar(100)"`
	DayOfWeek int       `gorm:"not null"`
	StartTime string    `gorm:"type:varchar(5);not null"`
	EndTime   string    `gorm:"type:varchar(5);not null"`
	Location  *string   `gorm:"type:varchar(255)"`
	IsActive  bool      `gorm:"default:true;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Sessions  []LessonSession
}

func (s *Schedule) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s Schedule) DayLabel() string {
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	if s.DayOfWeek < 0 || s.DayOfWeek >= len(days) {
		return "-"
	}
	return days[s.DayOfWeek]
}

func (s Schedule) SubjectLabel() string {
	if s.Subject == nil || *s.Subject == "" {
		return "-"
	}
	return *s.Subject
}

func (s Schedule) LocationLabel() string {
	if s.Location == nil || *s.Location == "" {
		return "-"
	}
	return *s.Location
}

func (s Schedule) StatusLabel() string {
	if s.IsActive {
		return "Aktif"
	}
	return "Nonaktif"
}

type LessonSession struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ScheduleID       *uuid.UUID `gorm:"type:uuid;index"`
	Schedule         *Schedule  `gorm:"foreignKey:ScheduleID"`
	TutorID          uuid.UUID  `gorm:"type:uuid;not null;index;index:idx_lesson_sessions_tutor_start"`
	Tutor            Tutor      `gorm:"foreignKey:TutorID"`
	StudentID        uuid.UUID  `gorm:"type:uuid;not null;index;index:idx_lesson_sessions_student_start"`
	Student          Student    `gorm:"foreignKey:StudentID"`
	Subject          *string    `gorm:"type:varchar(100)"`
	ScheduledStartAt time.Time  `gorm:"type:timestamptz;not null;index:idx_lesson_sessions_tutor_start;index:idx_lesson_sessions_student_start;index:idx_lesson_sessions_status_start"`
	ScheduledEndAt   time.Time  `gorm:"type:timestamptz;not null"`
	ActualStartAt    *time.Time `gorm:"type:timestamptz"`
	ActualEndAt      *time.Time `gorm:"type:timestamptz"`
	Status           string     `gorm:"type:varchar(30);default:scheduled;not null;index:idx_lesson_sessions_status_start"`
	Location         *string    `gorm:"type:varchar(255)"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Report           *LessonReport
}

func (s *LessonSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = LessonSessionScheduled
	}
	return nil
}

func (s LessonSession) SubjectLabel() string {
	if s.Subject == nil || *s.Subject == "" {
		return "-"
	}
	return *s.Subject
}

func (s LessonSession) LocationLabel() string {
	if s.Location == nil || *s.Location == "" {
		return "-"
	}
	return *s.Location
}

func (s LessonSession) StatusLabel() string {
	switch s.Status {
	case LessonSessionScheduled:
		return "Terjadwal"
	case LessonSessionCompleted:
		return "Selesai"
	case LessonSessionCanceled:
		return "Dibatalkan"
	case LessonSessionStudentAbsent:
		return "Murid Tidak Hadir"
	case LessonSessionFollowUpRequired:
		return "Perlu Tindak Lanjut"
	case LessonSessionRescheduled:
		return "Dijadwalkan Ulang"
	default:
		return s.Status
	}
}

func (s LessonSession) DateLabel() string {
	return s.ScheduledStartAt.Format("02 Jan 2006")
}

func (s LessonSession) TimeLabel() string {
	return s.ScheduledStartAt.Format("15:04") + " - " + s.ScheduledEndAt.Format("15:04")
}

func (s LessonSession) NeedsReportLabel() string {
	if s.Report == nil {
		return "Laporan belum tersedia"
	}
	return s.Report.StatusLabel()
}

type LessonReport struct {
	ID                     uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	LessonSessionID        uuid.UUID     `gorm:"type:uuid;unique;not null"`
	LessonSession          LessonSession `gorm:"foreignKey:LessonSessionID"`
	TutorID                uuid.UUID     `gorm:"type:uuid;not null;index"`
	Tutor                  Tutor         `gorm:"foreignKey:TutorID"`
	StudentID              uuid.UUID     `gorm:"type:uuid;not null;index"`
	Student                Student       `gorm:"foreignKey:StudentID"`
	MaterialSummary        string        `gorm:"type:text;not null"`
	ProgressSummary        *string       `gorm:"type:text"`
	Homework               *string       `gorm:"type:text"`
	IssueNotes             *string       `gorm:"type:text"`
	HomePracticeSuggestion *string       `gorm:"type:text"`
	PublishedAt            *time.Time    `gorm:"type:timestamptz"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (r *LessonReport) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (r LessonReport) StatusLabel() string {
	if r.PublishedAt == nil {
		return "Draft"
	}
	return "Terbit"
}

func (r LessonReport) OptionalLabel(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}

type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	User         User      `gorm:"foreignKey:UserID"`
	Token        string    `gorm:"type:varchar(512);unique;not null"`
	RefreshToken *string   `gorm:"type:text"`
	ExpiresAt    time.Time `gorm:"not null"`
	IPAddress    string    `gorm:"type:varchar(45)"`
	UserAgent    string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (s *Session) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type ActivityLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	User       User           `gorm:"foreignKey:UserID"`
	Action     string         `gorm:"type:varchar(100);not null"`
	EntityType string         `gorm:"type:varchar(50)"`
	EntityID   *uuid.UUID     `gorm:"type:uuid"`
	Metadata   map[string]any `gorm:"type:jsonb;serializer:json"`
	IPAddress  string         `gorm:"type:varchar(45)"`
	CreatedAt  time.Time      `gorm:"not null;index"`
}

func (l *ActivityLog) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func DefaultRoles() []Role {
	return []Role{
		{ID: RoleIDSuperAdmin, Name: RoleNameSuperAdmin, DisplayName: "Super Admin"},
		{ID: RoleIDTutor, Name: RoleNameTutor, DisplayName: "Tutor"},
		{ID: RoleIDParent, Name: RoleNameParent, DisplayName: "Parent"},
	}
}
