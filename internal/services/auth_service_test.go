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

func TestRegisterCreatesLocalUserWithPasswordHash(t *testing.T) {
	ctx := context.Background()
	users := newFakeAuthUsers()
	sessions := &fakeSessionRepository{}
	service := services.NewAuthService(users, sessions, time.Hour)

	result, err := service.Register(ctx, services.RegisterInput{
		Name:                 "Nia Parent",
		Email:                "NIA@Example.com",
		Password:             "password123",
		PasswordConfirmation: "password123",
		IPAddress:            "127.0.0.1",
		UserAgent:            "test",
	})

	require.NoError(t, err)
	assert.Equal(t, "nia@example.com", result.User.Email)
	assert.Equal(t, models.ProviderLocal, result.User.Provider)
	require.NotNil(t, result.User.PasswordHash)
	assert.NotEqual(t, "password123", *result.User.PasswordHash)
	assert.NotEmpty(t, result.Session.Token)
	require.Nil(t, result.User.RoleID)
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	users := newFakeAuthUsers()
	hash := "hashed"
	users.byEmail["nia@example.com"] = &models.User{ID: uuid.New(), Email: "nia@example.com", PasswordHash: &hash, Provider: models.ProviderLocal, IsActive: true}
	service := services.NewAuthService(users, &fakeSessionRepository{}, time.Hour)

	result, err := service.Register(ctx, services.RegisterInput{
		Name:                 "Nia Parent",
		Email:                "nia@example.com",
		Password:             "password123",
		PasswordConfirmation: "password123",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, services.ErrEmailAlreadyUsed)
}

func TestRegisterRejectsPasswordConfirmationMismatch(t *testing.T) {
	service := services.NewAuthService(newFakeAuthUsers(), &fakeSessionRepository{}, time.Hour)

	result, err := service.Register(context.Background(), services.RegisterInput{
		Name:                 "Nia Parent",
		Email:                "nia@example.com",
		Password:             "password123",
		PasswordConfirmation: "password456",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, services.ErrPasswordConfirmationMismatch)
}

func TestLoginCreatesSessionForActiveLocalUser(t *testing.T) {
	ctx := context.Background()
	users := newFakeAuthUsers()
	passwordHash, err := services.HashPassword("admin123")
	require.NoError(t, err)
	roleID := models.RoleIDSuperAdmin
	user := &models.User{ID: uuid.New(), Email: "admin@kalabelajar.com", Name: "Admin", Provider: models.ProviderLocal, PasswordHash: &passwordHash, RoleID: &roleID, IsActive: true}
	users.byEmail[user.Email] = user
	sessions := &fakeSessionRepository{}
	service := services.NewAuthService(users, sessions, time.Hour)

	result, err := service.Login(ctx, services.LoginInput{Email: "admin@kalabelajar.com", Password: "admin123", IPAddress: "127.0.0.1"})

	require.NoError(t, err)
	assert.Equal(t, user.ID, result.User.ID)
	assert.NotEmpty(t, result.Session.Token)
	assert.Equal(t, user.ID, sessions.created.UserID)
	assert.NotNil(t, users.lastLoginUserID)
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	users := newFakeAuthUsers()
	passwordHash, err := services.HashPassword("admin123")
	require.NoError(t, err)
	users.byEmail["admin@kalabelajar.com"] = &models.User{ID: uuid.New(), Email: "admin@kalabelajar.com", Provider: models.ProviderLocal, PasswordHash: &passwordHash, IsActive: true}
	service := services.NewAuthService(users, &fakeSessionRepository{}, time.Hour)

	result, err := service.Login(context.Background(), services.LoginInput{Email: "admin@kalabelajar.com", Password: "wrongpass"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestLoginRejectsInactiveUser(t *testing.T) {
	users := newFakeAuthUsers()
	passwordHash, err := services.HashPassword("admin123")
	require.NoError(t, err)
	users.byEmail["inactive@example.com"] = &models.User{ID: uuid.New(), Email: "inactive@example.com", Provider: models.ProviderLocal, PasswordHash: &passwordHash, IsActive: false}
	service := services.NewAuthService(users, &fakeSessionRepository{}, time.Hour)

	result, err := service.Login(context.Background(), services.LoginInput{Email: "inactive@example.com", Password: "admin123"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, services.ErrInactiveUser)
}

func TestLoginRejectsOAuthOnlyUser(t *testing.T) {
	users := newFakeAuthUsers()
	users.byEmail["google@example.com"] = &models.User{ID: uuid.New(), Email: "google@example.com", Provider: models.ProviderGoogle, PasswordHash: nil, IsActive: true}
	service := services.NewAuthService(users, &fakeSessionRepository{}, time.Hour)

	result, err := service.Login(context.Background(), services.LoginInput{Email: "google@example.com", Password: "admin123"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, services.ErrPasswordLoginUnavailable)
}

type fakeAuthUsers struct {
	byEmail         map[string]*models.User
	created         *models.User
	lastLoginUserID *uuid.UUID
}

func newFakeAuthUsers() *fakeAuthUsers {
	return &fakeAuthUsers{byEmail: map[string]*models.User{}}
}

func (r *fakeAuthUsers) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := r.byEmail[email]; ok {
		return user, nil
	}
	return nil, services.ErrUserNotFound
}

func (r *fakeAuthUsers) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	r.created = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *fakeAuthUsers) TouchLastLogin(ctx context.Context, userID uuid.UUID, when time.Time) error {
	r.lastLoginUserID = &userID
	return nil
}

type fakeSessionRepository struct {
	created models.Session
	err     error
}

func (r *fakeSessionRepository) Create(ctx context.Context, session *models.Session) error {
	if r.err != nil {
		return r.err
	}
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	r.created = *session
	return nil
}

func (r *fakeSessionRepository) FindValidByToken(ctx context.Context, token string, now time.Time) (*models.Session, error) {
	if r.created.Token != token || r.created.ExpiresAt.Before(now) {
		return nil, errors.New("session not found")
	}
	return &r.created, nil
}
