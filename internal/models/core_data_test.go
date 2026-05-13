package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/models"
)

func TestCoreDataModelsAssignUUIDsBeforeCreate(t *testing.T) {
	user := models.User{
		Email:    "parent@example.com",
		Name:     "Parent Kala",
		Provider: models.ProviderLocal,
		IsActive: true,
	}
	tutor := models.Tutor{UserID: uuid.New()}
	student := models.Student{ParentID: user.ID, Name: "Murid Kala"}
	log := models.ActivityLog{UserID: user.ID, Action: models.ActivityLogin}

	require.NoError(t, user.BeforeCreate(nil))
	require.NoError(t, tutor.BeforeCreate(nil))
	require.NoError(t, student.BeforeCreate(nil))
	require.NoError(t, log.BeforeCreate(nil))

	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotEqual(t, uuid.Nil, tutor.ID)
	assert.NotEqual(t, uuid.Nil, student.ID)
	assert.NotEqual(t, uuid.Nil, log.ID)
}

func TestRoleConstantsMatchTechSpecIDs(t *testing.T) {
	assert.Equal(t, uint(1), models.RoleIDSuperAdmin)
	assert.Equal(t, uint(2), models.RoleIDTutor)
	assert.Equal(t, uint(3), models.RoleIDParent)
}
