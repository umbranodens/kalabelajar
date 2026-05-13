package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound                 = errors.New("user not found")
	ErrEmailAlreadyUsed             = errors.New("email already used")
	ErrInvalidRegisterInput         = errors.New("register input is invalid")
	ErrPasswordConfirmationMismatch = errors.New("password confirmation does not match")
	ErrInvalidCredentials           = errors.New("email or password is invalid")
	ErrInactiveUser                 = errors.New("user is inactive")
	ErrPasswordLoginUnavailable     = errors.New("password login is unavailable for this account")
)

type AuthUserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdateOAuthProfile(ctx context.Context, user *models.User) error
	TouchLastLogin(ctx context.Context, userID uuid.UUID, when time.Time) error
}

type AuthSessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	FindValidByToken(ctx context.Context, token string, now time.Time) (*models.Session, error)
}

type AuthService struct {
	users      AuthUserRepository
	sessions   AuthSessionRepository
	sessionTTL time.Duration
}

type RegisterInput struct {
	Name                 string
	Email                string
	Password             string
	PasswordConfirmation string
	IPAddress            string
	UserAgent            string
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type OAuthLoginInput struct {
	IPAddress string
	UserAgent string
}

type GoogleProfile struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

type AuthResult struct {
	User    *models.User
	Session *models.Session
}

func NewAuthService(users AuthUserRepository, sessions AuthSessionRepository, sessionTTL time.Duration) *AuthService {
	return &AuthService{users: users, sessions: sessions, sessionTTL: sessionTTL}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)
	if name == "" || !validEmail(email) || len(input.Password) < 8 {
		return nil, ErrInvalidRegisterInput
	}
	if input.Password != input.PasswordConfirmation {
		return nil, ErrPasswordConfirmationMismatch
	}

	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil && !isUserNotFound(err) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyUsed
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:        email,
		Name:         name,
		Provider:     models.ProviderLocal,
		PasswordHash: &hash,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	session, err := s.createSession(ctx, user.ID, input.IPAddress, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.users.FindByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}
	if user.PasswordHash == nil || strings.TrimSpace(*user.PasswordHash) == "" {
		return nil, ErrPasswordLoginUnavailable
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := s.users.TouchLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}
	session, err := s.createSession(ctx, user.ID, input.IPAddress, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

func (s *AuthService) LoginWithGoogleProfile(ctx context.Context, profile GoogleProfile, input OAuthLoginInput) (*AuthResult, error) {
	email := normalizeEmail(profile.Email)
	name := strings.TrimSpace(profile.Name)
	providerID := strings.TrimSpace(profile.ProviderID)
	if !validEmail(email) || name == "" || providerID == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil && !isUserNotFound(err) {
		return nil, fmt.Errorf("find google user: %w", err)
	}
	if user != nil {
		if !user.IsActive {
			return nil, ErrInactiveUser
		}
		user.Name = name
		user.Provider = models.ProviderGoogle
		user.ProviderID = &providerID
		user.AvatarURL = optionalString(profile.AvatarURL)
		if err := s.users.UpdateOAuthProfile(ctx, user); err != nil {
			return nil, fmt.Errorf("update google profile: %w", err)
		}
	} else {
		user = &models.User{
			Email:      email,
			Name:       name,
			Provider:   models.ProviderGoogle,
			ProviderID: &providerID,
			AvatarURL:  optionalString(profile.AvatarURL),
			IsActive:   true,
		}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create google user: %w", err)
		}
	}

	now := time.Now().UTC()
	if err := s.users.TouchLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}
	session, err := s.createSession(ctx, user.ID, input.IPAddress, input.UserAgent)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Session: session}, nil
}

func (s *AuthService) FindSession(ctx context.Context, token string) (*models.Session, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrInvalidCredentials
	}
	return s.sessions.FindValidByToken(ctx, token, time.Now().UTC())
}

func (s *AuthService) createSession(ctx context.Context, userID uuid.UUID, ipAddress string, userAgent string) (*models.Session, error) {
	token, err := NewSessionToken()
	if err != nil {
		return nil, err
	}
	session := &models.Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(s.sessionTTL),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create auth session: %w", err)
	}
	return session, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func NewSessionToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
