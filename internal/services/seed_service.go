package services

import (
	"context"
	"fmt"

	"github.com/umbranodens/kalabelajar/internal/models"
)

const (
	DefaultSuperAdminEmail    = "admin@kalabelajar.com"
	DefaultSuperAdminPassword = "admin123"
)

type RoleSeeder interface {
	UpsertRoles(ctx context.Context, roles []models.Role) error
}

type SeedService struct {
	roles RoleSeeder
	users AuthUserRepository
}

func NewSeedService(roles RoleSeeder, users AuthUserRepository) *SeedService {
	return &SeedService{roles: roles, users: users}
}

func (s *SeedService) SeedRolesAndSuperAdmin(ctx context.Context) (*models.User, error) {
	if err := s.roles.UpsertRoles(ctx, models.DefaultRoles()); err != nil {
		return nil, fmt.Errorf("seed roles: %w", err)
	}

	existing, err := s.users.FindByEmail(ctx, DefaultSuperAdminEmail)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !isUserNotFound(err) {
		return nil, fmt.Errorf("check super admin: %w", err)
	}

	hash, err := HashPassword(DefaultSuperAdminPassword)
	if err != nil {
		return nil, err
	}
	roleID := models.RoleIDSuperAdmin
	admin := &models.User{
		Email:        DefaultSuperAdminEmail,
		Name:         "Super Admin Kala Belajar",
		Provider:     models.ProviderLocal,
		PasswordHash: &hash,
		RoleID:       &roleID,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("create super admin: %w", err)
	}
	return admin, nil
}
