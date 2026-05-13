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
	ActivityLogin       = "login"
	ActivityAssignRole  = "assign_role"
	ActivityVerifyTutor = "verify_tutor"
	ActivityAssignTutor = "assign_tutor"
)

const (
	EntityStudent = "student"
	EntityTutor   = "tutor"
	EntityUser    = "user"
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
}

func (s *Student) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
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
