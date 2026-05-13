package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/database"
	"github.com/umbranodens/kalabelajar/internal/handlers"
	"github.com/umbranodens/kalabelajar/internal/middleware"
	"github.com/umbranodens/kalabelajar/internal/repositories"
	"github.com/umbranodens/kalabelajar/internal/seed"
	"github.com/umbranodens/kalabelajar/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := database.Ping(context.Background(), db); err != nil {
		log.Fatalf("database health check failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	seedService := services.NewSeedService(
		repositories.NewRoleRepository(db),
		repositories.NewUserRepository(db),
	)
	if _, err := seedService.SeedRolesAndSuperAdmin(context.Background()); err != nil {
		log.Fatalf("seed roles and super admin: %v", err)
	}
	if cfg.App.Env != "production" {
		if err := seed.LocalDevelopmentData(context.Background(), db); err != nil {
			log.Fatalf("seed local development data: %v", err)
		}
	}

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*.html")
	router.Static("/static", "./web/static")
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	authService := services.NewAuthService(
		repositories.NewUserRepository(db),
		repositories.NewSessionRepository(db),
		cfg.Session.TTL,
	)
	router.Use(middleware.LoadSession(repositories.NewSessionRepository(db)))
	handlers.NewAuthHandler(authService, cfg.Google, cfg.App.URL).RegisterRoutes(router)
	adminService := services.NewAdminService(
		repositories.NewTutorRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewActivityLogRepository(db),
	)
	assignmentService := services.NewAssignmentService(
		repositories.NewStudentRepository(db),
		repositories.NewTutorRepository(db),
		repositories.NewActivityLogRepository(db),
	)
	scheduleService := services.NewScheduleService(
		repositories.NewScheduleRepository(db),
		repositories.NewActivityLogRepository(db),
	)
	reportService := services.NewLessonReportService(
		repositories.NewLessonReportRepository(db),
		repositories.NewActivityLogRepository(db),
	)
	handlers.NewAdminHandler(db, adminService, assignmentService, scheduleService, reportService).RegisterRoutes(router)

	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
