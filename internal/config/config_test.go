package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/config"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_URL", "http://localhost:9090")
	t.Setenv("APP_SECRET", "test-secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "kb")
	t.Setenv("DB_PASS", "password")
	t.Setenv("DB_NAME", "kalabelajar_test")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("GOOGLE_CLIENT_ID", "google-client")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:9090/auth/google/callback")
	t.Setenv("SESSION_TTL_HOURS", "48")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "test", cfg.App.Env)
	assert.Equal(t, "9090", cfg.App.Port)
	assert.Equal(t, "http://localhost:9090", cfg.App.URL)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "kb", cfg.Database.User)
	assert.Equal(t, "kalabelajar_test", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "google-client", cfg.Google.ClientID)
	assert.Equal(t, 48*time.Hour, cfg.Session.TTL)
}

func TestLoadDefaultsOptionalValues(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_URL", "http://localhost:8080")
	t.Setenv("APP_SECRET", "test-secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "kb")
	t.Setenv("DB_PASS", "password")
	t.Setenv("DB_NAME", "kalabelajar")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.App.Port)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 24*time.Hour, cfg.Session.TTL)
}

func TestLoadReturnsValidationErrorForMissingRequiredValues(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	cfg, err := config.Load()

	assert.Empty(t, cfg.App.Secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_SECRET")
	assert.Contains(t, err.Error(), "DB_HOST")
}
