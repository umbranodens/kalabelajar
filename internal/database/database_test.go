package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
