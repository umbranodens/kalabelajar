package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
)

func TestSeedRolesAndSuperAdminCreatesDefaultLocalAdmin(t *testing.T) {
	ctx := context.Background()
	users := newFakeAuthUsers()
	roles := &fakeRoleSeeder{}
	seeder := services.NewSeedService(roles, users)

	admin, err := seeder.SeedRolesAndSuperAdmin(ctx)

	require.NoError(t, err)
	assert.Equal(t, []models.Role(models.DefaultRoles()), roles.upserted)
	require.NotNil(t, admin)
	assert.Equal(t, "admin@kalabelajar.com", admin.Email)
	assert.Equal(t, models.ProviderLocal, admin.Provider)
	require.NotNil(t, admin.RoleID)
	assert.Equal(t, models.RoleIDSuperAdmin, *admin.RoleID)
	require.NotNil(t, admin.PasswordHash)
	assert.NotEqual(t, "admin123", *admin.PasswordHash)
	assert.True(t, admin.IsActive)
	assert.Equal(t, admin, users.created)
}

func TestSeedRolesAndSuperAdminLeavesExistingAdmin(t *testing.T) {
	ctx := context.Background()
	users := newFakeAuthUsers()
	existingHash := "hash"
	existing := &models.User{ID: uuid.New(), Email: "admin@kalabelajar.com", PasswordHash: &existingHash, Provider: models.ProviderLocal, IsActive: true}
	users.byEmail[existing.Email] = existing
	seeder := services.NewSeedService(&fakeRoleSeeder{}, users)

	admin, err := seeder.SeedRolesAndSuperAdmin(ctx)

	require.NoError(t, err)
	assert.Equal(t, existing.ID, admin.ID)
	assert.Nil(t, users.created)
}

type fakeRoleSeeder struct {
	upserted []models.Role
}

func (r *fakeRoleSeeder) UpsertRoles(ctx context.Context, roles []models.Role) error {
	r.upserted = roles
	return nil
}
