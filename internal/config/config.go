package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Google   GoogleConfig
	Session  SessionConfig
}

type AppConfig struct {
	Env    string
	Port   string
	URL    string
	Secret string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type SessionConfig struct {
	TTL time.Duration
}

func Load() (Config, error) {
	if os.Getenv("APP_ENV") != "test" {
		_ = godotenv.Load()
	}

	ttlHours, err := intEnv("SESSION_TTL_HOURS", 24)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		App: AppConfig{
			Env:    stringEnv("APP_ENV", "development"),
			Port:   stringEnv("APP_PORT", "8080"),
			URL:    os.Getenv("APP_URL"),
			Secret: os.Getenv("APP_SECRET"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     stringEnv("DB_PORT", "5432"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  stringEnv("DB_SSLMODE", "disable"),
		},
		Google: GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		Session: SessionConfig{
			TTL: time.Duration(ttlHours) * time.Hour,
		},
	}

	return cfg, cfg.Validate()
}

func (cfg Config) Validate() error {
	required := map[string]string{
		"APP_URL":    cfg.App.URL,
		"APP_SECRET": cfg.App.Secret,
		"DB_HOST":    cfg.Database.Host,
		"DB_USER":    cfg.Database.User,
		"DB_PASS":    cfg.Database.Password,
		"DB_NAME":    cfg.Database.Name,
	}

	var missing []string
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment values: %s", strings.Join(missing, ", "))
	}
	if cfg.Session.TTL <= 0 {
		return errors.New("SESSION_TTL_HOURS must be greater than 0")
	}

	return nil
}

func stringEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return parsed, nil
}
