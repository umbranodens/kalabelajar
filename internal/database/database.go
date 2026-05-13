package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BuildDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)
}

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(BuildDSN(cfg)), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	return db, nil
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database handle: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func MigrationModels() []any {
	return []any{
		&models.Role{},
		&models.User{},
		&models.Tutor{},
		&models.Student{},
		&models.Schedule{},
		&models.Session{},
		&models.ActivityLog{},
	}
}

func MigrationModelNames() []string {
	migrationModels := MigrationModels()
	names := make([]string, 0, len(migrationModels))
	for _, model := range migrationModels {
		modelType := reflect.Indirect(reflect.ValueOf(model)).Type()
		names = append(names, modelType.Name())
	}
	return names
}

func Migrate(db *gorm.DB) error {
	if db.Dialector.Name() == "postgres" {
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
			return fmt.Errorf("ensure pgcrypto extension: %w", err)
		}
	}

	if err := db.AutoMigrate(MigrationModels()...); err != nil {
		return fmt.Errorf("auto migrate core data models: %w", err)
	}

	return nil
}
