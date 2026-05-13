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
	schedules  *services.ScheduleService
}

func NewAdminHandler(db *gorm.DB, admin *services.AdminService, assignment *services.AssignmentService, schedules *services.ScheduleService) *AdminHandler {
	return &AdminHandler{db: db, admin: admin, assignment: assignment, schedules: schedules}
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
	admin.GET("/schedules", h.AdminSchedules)
	admin.POST("/schedules", h.CreateSchedule)
	admin.POST("/schedules/:id/active", h.SetScheduleActive)
	admin.GET("/activity", h.Activity)
	router.GET("/tutor/schedules", middleware.RequireRoles(models.RoleIDTutor), h.TutorSchedules)
	router.GET("/parent/schedules", middleware.RequireRoles(models.RoleIDParent), h.ParentSchedules)
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	if user.RoleID == nil || *user.RoleID != models.RoleIDSuperAdmin {
		data := gin.H{"Title": "Dashboard - Kala Belajar", "User": user, "RoleLabel": roleLabel(user.RoleID)}
		if user.RoleID != nil && *user.RoleID == models.RoleIDTutor {
			var tutor models.Tutor
			if err := h.db.Where("user_id = ?", user.ID).First(&tutor).Error; err == nil {
				var scheduleCount, studentCount int64
				h.db.Model(&models.Schedule{}).Where("tutor_id = ? AND is_active = ?", tutor.ID, true).Count(&scheduleCount)
				h.db.Model(&models.Student{}).Where("assigned_tutor_id = ?", tutor.ID).Count(&studentCount)
				data["IsTutor"] = true
				data["ScheduleCount"] = scheduleCount
				data["StudentCount"] = studentCount
			}
		}
		if user.RoleID != nil && *user.RoleID == models.RoleIDParent {
			var childCount, scheduleCount int64
			h.db.Model(&models.Student{}).Where("parent_id = ?", user.ID).Count(&childCount)
			h.db.Model(&models.Schedule{}).
				Joins("JOIN students ON students.id = schedules.student_id").
				Where("students.parent_id = ? AND schedules.is_active = ?", user.ID, true).
				Count(&scheduleCount)
			data["IsParent"] = true
			data["ChildCount"] = childCount
			data["ScheduleCount"] = scheduleCount
		}
		c.HTML(http.StatusOK, "dashboard.tmpl", data)
		return
	}
	var totalTutors, totalStudents, totalParents, pendingTutors, activeSchedules, activities int64
	h.db.Model(&models.Tutor{}).Count(&totalTutors)
	h.db.Model(&models.Student{}).Count(&totalStudents)
	h.db.Model(&models.User{}).Where("role_id = ?", models.RoleIDParent).Count(&totalParents)
	h.db.Model(&models.Tutor{}).Where("is_verified = ?", false).Count(&pendingTutors)
	h.db.Model(&models.Schedule{}).Where("is_active = ?", true).Count(&activeSchedules)
	h.db.Model(&models.ActivityLog{}).Count(&activities)
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Title":           "Dashboard - Kala Belajar",
		"User":            user,
		"RoleLabel":       "Super Admin",
		"IsSuperAdmin":    true,
		"TotalTutors":     totalTutors,
		"TotalStudents":   totalStudents,
		"TotalParents":    totalParents,
		"PendingTutors":   pendingTutors,
		"ActiveSchedules": activeSchedules,
		"ActivityCount":   activities,
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

func (h *AdminHandler) AdminSchedules(c *gin.Context) {
	h.renderAdminSchedules(c, http.StatusOK, "")
}

func (h *AdminHandler) CreateSchedule(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	tutorID, tutorErr := uuid.Parse(c.PostForm("tutor_id"))
	studentID, studentErr := uuid.Parse(c.PostForm("student_id"))
	dayOfWeek, dayErr := strconv.Atoi(c.PostForm("day_of_week"))
	if tutorErr != nil || studentErr != nil || dayErr != nil {
		h.renderAdminSchedules(c, http.StatusBadRequest, "Tutor, murid, dan hari wajib dipilih dengan benar.")
		return
	}
	_, err := h.schedules.CreateSchedule(c.Request.Context(), services.CreateScheduleInput{
		ActorUserID: user.ID,
		TutorID:     tutorID,
		StudentID:   studentID,
		Subject:     c.PostForm("subject"),
		DayOfWeek:   dayOfWeek,
		StartTime:   c.PostForm("start_time"),
		EndTime:     c.PostForm("end_time"),
		Location:    c.PostForm("location"),
		IPAddress:   c.ClientIP(),
	})
	if err != nil {
		h.renderAdminSchedules(c, http.StatusBadRequest, "Jadwal belum bisa disimpan. Pastikan tutor aktif dan terverifikasi, murid sudah di-assign ke tutor itu, hari valid, dan jam selesai setelah jam mulai.")
		return
	}
	c.Redirect(http.StatusFound, "/admin/schedules")
}

func (h *AdminHandler) SetScheduleActive(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err == nil {
		isActive, _ := strconv.ParseBool(c.PostForm("is_active"))
		_ = h.schedules.SetScheduleActive(c.Request.Context(), services.SetScheduleActiveInput{ActorUserID: user.ID, ScheduleID: id, IsActive: isActive, IPAddress: c.ClientIP()})
	}
	c.Redirect(http.StatusFound, "/admin/schedules")
}

func (h *AdminHandler) TutorSchedules(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var tutor models.Tutor
	if err := h.db.Where("user_id = ?", user.ID).First(&tutor).Error; err != nil {
		c.HTML(http.StatusOK, "role_schedules.tmpl", gin.H{"Title": "Jadwal Saya", "User": user, "RoleLabel": "Tutor", "Schedules": []models.Schedule{}})
		return
	}
	var schedules []models.Schedule
	h.db.Preload("Tutor.User").
		Preload("Student.Parent").
		Where("tutor_id = ? AND is_active = ?", tutor.ID, true).
		Order("day_of_week asc, start_time asc").
		Find(&schedules)
	c.HTML(http.StatusOK, "role_schedules.tmpl", gin.H{"Title": "Jadwal Saya", "User": user, "RoleLabel": "Tutor", "Schedules": schedules, "ScheduleTitle": "Jadwal Saya", "EmptyText": "Belum ada jadwal hari ini."})
}

func (h *AdminHandler) ParentSchedules(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var schedules []models.Schedule
	h.db.Preload("Tutor.User").
		Preload("Student.Parent").
		Joins("JOIN students ON students.id = schedules.student_id").
		Where("students.parent_id = ? AND schedules.is_active = ?", user.ID, true).
		Order("schedules.day_of_week asc, schedules.start_time asc").
		Find(&schedules)
	c.HTML(http.StatusOK, "role_schedules.tmpl", gin.H{"Title": "Jadwal Les Anak", "User": user, "RoleLabel": "Parent", "Schedules": schedules, "ScheduleTitle": "Jadwal Les Anak", "EmptyText": "Belum ada jadwal anak yang aktif."})
}

func (h *AdminHandler) renderAdminSchedules(c *gin.Context, status int, errorMessage string) {
	var schedules []models.Schedule
	query := h.db.Preload("Tutor.User").Preload("Student.Parent").Order("day_of_week asc, start_time asc")
	if state := c.Query("status"); state == "active" {
		query = query.Where("is_active = ?", true)
	} else if state == "inactive" {
		query = query.Where("is_active = ?", false)
	}
	query.Find(&schedules)

	var tutors []models.Tutor
	h.db.Preload("User").
		Joins("JOIN users ON users.id = tutors.user_id").
		Where("tutors.is_verified = ? AND users.is_active = ?", true, true).
		Order("users.name asc").
		Find(&tutors)

	var students []models.Student
	h.db.Preload("Parent").Preload("AssignedTutor").Preload("AssignedTutor.User").
		Where("assigned_tutor_id IS NOT NULL").
		Order("name asc").
		Find(&students)

	c.HTML(status, "admin_schedules.tmpl", gin.H{
		"Title":     "Manajemen Jadwal",
		"Schedules": schedules,
		"Tutors":    tutors,
		"Students":  students,
		"Error":     errorMessage,
		"Query":     c.Request.URL.Query(),
	})
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
