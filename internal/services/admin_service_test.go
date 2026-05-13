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

func TestAdminServiceVerifyTutorLogsActivity(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	tutorID := uuid.New()
	tutors := &fakeAdminTutorAdminRepository{}
	logs := &fakeActivityLogger{}
	service := services.NewAdminService(tutors, &fakeAdminUserAdminRepository{}, logs)

	err := service.VerifyTutor(ctx, services.AdminActionInput{
		ActorUserID: actorID,
		EntityID:    tutorID,
		IPAddress:   "127.0.0.1",
	})

	require.NoError(t, err)
	assert.Equal(t, tutorID, tutors.verifiedTutorID)
	require.Len(t, logs.entries, 1)
	assert.Equal(t, models.ActivityVerifyTutor, logs.entries[0].Action)
	assert.Equal(t, models.EntityTutor, logs.entries[0].EntityType)
	require.NotNil(t, logs.entries[0].EntityID)
	assert.Equal(t, tutorID, *logs.entries[0].EntityID)
}

func TestAdminServiceSetUserActiveLogsActivity(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	users := &fakeAdminUserAdminRepository{}
	logs := &fakeActivityLogger{}
	service := services.NewAdminService(&fakeAdminTutorAdminRepository{}, users, logs)

	err := service.SetUserActive(ctx, services.SetUserActiveInput{
		ActorUserID: actorID,
		UserID:      userID,
		IsActive:    false,
		IPAddress:   "127.0.0.1",
	})

	require.NoError(t, err)
	assert.Equal(t, userID, users.activeUserID)
	assert.False(t, users.activeValue)
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "set_user_active", logs.entries[0].Action)
	assert.Equal(t, models.EntityUser, logs.entries[0].EntityType)
}

type fakeAdminTutorAdminRepository struct {
	verifiedTutorID uuid.UUID
}

func (r *fakeAdminTutorAdminRepository) VerifyTutor(ctx context.Context, tutorID uuid.UUID) error {
	r.verifiedTutorID = tutorID
	return nil
}

type fakeAdminUserAdminRepository struct {
	activeUserID uuid.UUID
	activeValue  bool
}

func (r *fakeAdminUserAdminRepository) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error {
	r.activeUserID = userID
	r.activeValue = isActive
	return nil
}
