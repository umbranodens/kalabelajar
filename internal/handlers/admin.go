package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/umbranodens/kalabelajar/internal/middleware"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db         *gorm.DB
	admin      *services.AdminService
	assignment *services.AssignmentService
}

func NewAdminHandler(db *gorm.DB, admin *services.AdminService, assignment *services.AssignmentService) *AdminHandler {
	return &AdminHandler{db: db, admin: admin, assignment: assignment}
}

func (h *AdminHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/dashboard", middleware.RequireAuth(), h.Dashboard)
	admin := router.Group("/admin", middleware.RequireRoles(models.RoleIDSuperAdmin))
	admin.GET("/users", h.Users)
	admin.POST("/users/:id/active", h.SetUserActive)
	admin.GET("/tutors", h.Tutors)
	admin.POST("/tutors/:id/verify", h.VerifyTutor)
	admin.GET("/students", h.Students)
	admin.POST("/students/:id/assign", h.AssignTutor)
	admin.GET("/activity", h.Activity)
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	if user.RoleID == nil || *user.RoleID != models.RoleIDSuperAdmin {
		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{"Title": "Dashboard - Kala Belajar", "User": user, "RoleLabel": roleLabel(user.RoleID)})
		return
	}
	var totalTutors, totalStudents, totalParents, pendingTutors, activities int64
	h.db.Model(&models.Tutor{}).Count(&totalTutors)
	h.db.Model(&models.Student{}).Count(&totalStudents)
	h.db.Model(&models.User{}).Where("role_id = ?", models.RoleIDParent).Count(&totalParents)
	h.db.Model(&models.Tutor{}).Where("is_verified = ?", false).Count(&pendingTutors)
	h.db.Model(&models.ActivityLog{}).Count(&activities)
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Title":         "Dashboard - Kala Belajar",
		"User":          user,
		"RoleLabel":     "Super Admin",
		"IsSuperAdmin":  true,
		"TotalTutors":   totalTutors,
		"TotalStudents": totalStudents,
		"TotalParents":  totalParents,
		"PendingTutors": pendingTutors,
		"ActivityCount": activities,
	})
}

func (h *AdminHandler) Users(c *gin.Context) {
	var users []models.User
	query := h.db.Preload("Role").Order("created_at desc")
	if role := c.Query("role"); role != "" {
		switch role {
		case "pending":
			query = query.Where("role_id IS NULL")
		case models.RoleNameSuperAdmin:
			query = query.Where("role_id = ?", models.RoleIDSuperAdmin)
		case models.RoleNameTutor:
			query = query.Where("role_id = ?", models.RoleIDTutor)
		case models.RoleNameParent:
			query = query.Where("role_id = ?", models.RoleIDParent)
		}
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("is_active = ?", status == "active")
	}
	if search := c.Query("q"); search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", "%"+search+"%", "%"+search+"%")
	}
	query.Find(&users)
	c.HTML(http.StatusOK, "admin_users.tmpl", gin.H{"Title": "Manajemen User", "Users": users, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) SetUserActive(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err == nil {
		isActive, _ := strconv.ParseBool(c.PostForm("is_active"))
		_ = h.admin.SetUserActive(c.Request.Context(), services.SetUserActiveInput{ActorUserID: user.ID, UserID: id, IsActive: isActive, IPAddress: c.ClientIP()})
	}
	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *AdminHandler) Tutors(c *gin.Context) {
	var tutors []models.Tutor
	query := h.db.Preload("User").Order("created_at desc")
	if status := c.Query("status"); status != "" {
		switch status {
		case "verified":
			query = query.Where("is_verified = ?", true)
		case "unverified":
			query = query.Where("is_verified = ?", false)
		case "inactive":
			query = query.Joins("JOIN users ON users.id = tutors.user_id").Where("users.is_active = ?", false)
		}
	}
	query.Find(&tutors)
	c.HTML(http.StatusOK, "admin_tutors.tmpl", gin.H{"Title": "Manajemen Tutor", "Tutors": tutors, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) VerifyTutor(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err == nil {
		_ = h.admin.VerifyTutor(c.Request.Context(), services.AdminActionInput{ActorUserID: user.ID, EntityID: id, IPAddress: c.ClientIP()})
	}
	c.Redirect(http.StatusFound, "/admin/tutors")
}

func (h *AdminHandler) Students(c *gin.Context) {
	var students []models.Student
	query := h.db.Preload("Parent").Preload("AssignedTutor").Preload("AssignedTutor.User").Order("created_at desc")
	if assignment := c.Query("assignment"); assignment == "unassigned" {
		query = query.Where("assigned_tutor_id IS NULL")
	} else if assignment == "assigned" {
		query = query.Where("assigned_tutor_id IS NOT NULL")
	}
	if search := c.Query("q"); search != "" {
		query = query.Where("LOWER(students.name) LIKE LOWER(?)", "%"+search+"%")
	}
	query.Find(&students)
	var tutors []models.Tutor
	h.db.Preload("User").
		Joins("JOIN users ON users.id = tutors.user_id").
		Where("tutors.is_verified = ? AND users.is_active = ?", true, true).
		Order("users.name asc").
		Find(&tutors)
	c.HTML(http.StatusOK, "admin_students.tmpl", gin.H{"Title": "Manajemen Murid", "Students": students, "Tutors": tutors, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) AssignTutor(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	studentID, studentErr := uuid.Parse(c.Param("id"))
	tutorID, tutorErr := uuid.Parse(c.PostForm("tutor_id"))
	if studentErr == nil && tutorErr == nil {
		_, _ = h.assignment.AssignTutorToStudent(c.Request.Context(), services.AssignTutorInput{ActorUserID: user.ID, StudentID: studentID, TutorID: tutorID, IPAddress: c.ClientIP()})
	}
	c.Redirect(http.StatusFound, "/admin/students")
}

func (h *AdminHandler) Activity(c *gin.Context) {
	var logs []models.ActivityLog
	query := h.db.Preload("User").Order("created_at desc").Limit(100)
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	query.Find(&logs)
	c.HTML(http.StatusOK, "admin_activity.tmpl", gin.H{"Title": "Activity Log", "Logs": logs, "Query": c.Request.URL.Query()})
}
