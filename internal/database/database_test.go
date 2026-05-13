package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/database"
)

func TestBuildDSNUsesPostgresConnectionFields(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "kb",
		Password: "secret",
		Name:     "kalabelajar",
		SSLMode:  "disable",
	}

	dsn := database.BuildDSN(cfg)

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=kb")
	assert.Contains(t, dsn, "password=secret")
	assert.Contains(t, dsn, "dbname=kalabelajar")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestMigrationModelsIncludesCoreDataTables(t *testing.T) {
	modelNames := database.MigrationModelNames()

	require.NotEmpty(t, modelNames)
	assert.Contains(t, modelNames, "Role")
	assert.Contains(t, modelNames, "User")
	assert.Contains(t, modelNames, "Tutor")
	assert.Contains(t, modelNames, "Student")
	assert.Contains(t, modelNames, "Schedule")
	assert.Contains(t, modelNames, "Session")
	assert.Contains(t, modelNames, "ActivityLog")
}
