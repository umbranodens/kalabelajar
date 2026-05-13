package handlers

import (
	"net/http"
	"strconv"
	"time"

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
	reports    *services.LessonReportService
}

func NewAdminHandler(db *gorm.DB, admin *services.AdminService, assignment *services.AssignmentService, schedules *services.ScheduleService, reports *services.LessonReportService) *AdminHandler {
	return &AdminHandler{db: db, admin: admin, assignment: assignment, schedules: schedules, reports: reports}
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
	admin.GET("/lesson-sessions", h.AdminLessonSessions)
	admin.GET("/lesson-sessions/:id", h.AdminLessonSessionDetail)
	admin.GET("/activity", h.Activity)
	router.GET("/tutor/schedules", middleware.RequireRoles(models.RoleIDTutor), h.TutorSchedules)
	router.GET("/tutor/sessions", middleware.RequireRoles(models.RoleIDTutor), h.TutorSessions)
	router.GET("/tutor/sessions/:id/report", middleware.RequireRoles(models.RoleIDTutor), h.TutorReportForm)
	router.POST("/tutor/sessions/:id/report", middleware.RequireRoles(models.RoleIDTutor), h.SaveTutorReport)
	router.GET("/tutor/report/create", middleware.RequireRoles(models.RoleIDTutor), h.TutorCreateReportForm)
	router.POST("/tutor/report/create", middleware.RequireRoles(models.RoleIDTutor), h.TutorCreateReport)
	router.GET("/parent/schedules", middleware.RequireRoles(models.RoleIDParent), h.ParentSchedules)
	router.GET("/parent/reports", middleware.RequireRoles(models.RoleIDParent), h.ParentReports)
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
		c.HTML(http.StatusOK, "dashboard.html", data)
		return
	}
	var totalTutors, totalStudents, totalParents, pendingTutors, activeSchedules, activities int64
	h.db.Model(&models.Tutor{}).Count(&totalTutors)
	h.db.Model(&models.Student{}).Count(&totalStudents)
	h.db.Model(&models.User{}).Where("role_id = ?", models.RoleIDParent).Count(&totalParents)
	h.db.Model(&models.Tutor{}).Where("is_verified = ?", false).Count(&pendingTutors)
	h.db.Model(&models.Schedule{}).Where("is_active = ?", true).Count(&activeSchedules)
	h.db.Model(&models.ActivityLog{}).Count(&activities)
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
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
	c.HTML(http.StatusOK, "admin_users.html", gin.H{"Title": "Manajemen User", "Users": users, "Query": c.Request.URL.Query()})
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
	c.HTML(http.StatusOK, "admin_tutors.html", gin.H{"Title": "Manajemen Tutor", "Tutors": tutors, "Query": c.Request.URL.Query()})
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
	c.HTML(http.StatusOK, "admin_students.html", gin.H{"Title": "Manajemen Murid", "Students": students, "Tutors": tutors, "Query": c.Request.URL.Query()})
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
		c.HTML(http.StatusOK, "role_schedules.html", gin.H{"Title": "Jadwal Saya", "User": user, "RoleLabel": "Tutor", "Schedules": []models.Schedule{}})
		return
	}
	var schedules []models.Schedule
	h.db.Preload("Tutor.User").
		Preload("Student.Parent").
		Where("tutor_id = ? AND is_active = ?", tutor.ID, true).
		Order("day_of_week asc, start_time asc").
		Find(&schedules)
	c.HTML(http.StatusOK, "role_schedules.html", gin.H{"Title": "Jadwal Saya", "User": user, "RoleLabel": "Tutor", "Schedules": schedules, "ScheduleTitle": "Jadwal Saya", "EmptyText": "Belum ada jadwal hari ini."})
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
	c.HTML(http.StatusOK, "role_schedules.html", gin.H{"Title": "Jadwal Les Anak", "User": user, "RoleLabel": "Parent", "Schedules": schedules, "ScheduleTitle": "Jadwal Les Anak", "EmptyText": "Belum ada jadwal anak yang aktif."})
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

	c.HTML(status, "admin_schedules.html", gin.H{
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
	c.HTML(http.StatusOK, "admin_activity.html", gin.H{"Title": "Activity Log", "Logs": logs, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) AdminLessonSessions(c *gin.Context) {
	var sessions []models.LessonSession
	query := h.db.Preload("Tutor.User").Preload("Student.Parent").Preload("Report").Order("scheduled_start_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&sessions)
	c.HTML(http.StatusOK, "admin_lesson_sessions.html", gin.H{"Title": "Laporan Sesi / Kehadiran", "Sessions": sessions, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) AdminLessonSessionDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/lesson-sessions")
		return
	}
	var session models.LessonSession
	if err := h.db.Preload("Tutor.User").Preload("Student.Parent").Preload("Report").First(&session, "id = ?", id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/lesson-sessions")
		return
	}
	c.HTML(http.StatusOK, "admin_lesson_session_detail.html", gin.H{"Title": "Detail Laporan Sesi", "Session": session})
}

func (h *AdminHandler) TutorSessions(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	tutor, err := h.currentTutor(user.ID)
	if err != nil {
		c.HTML(http.StatusOK, "tutor_sessions.html", gin.H{"Title": "Riwayat Mengajar", "User": user, "Sessions": []models.LessonSession{}})
		return
	}
	var sessions []models.LessonSession
	h.db.Preload("Student.Parent").
		Preload("Report").
		Where("tutor_id = ?", tutor.ID).
		Order("scheduled_start_at desc").
		Find(&sessions)
	c.HTML(http.StatusOK, "tutor_sessions.html", gin.H{"Title": "Riwayat Mengajar", "User": user, "Sessions": sessions})
}

func (h *AdminHandler) TutorReportForm(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	h.renderTutorReportForm(c, user, http.StatusOK, "")
}

func (h *AdminHandler) SaveTutorReport(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	tutor, err := h.currentTutor(user.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/tutor/sessions")
		return
	}
	publish := c.PostForm("publish") == "on"
	_, err = h.reports.SaveReport(c.Request.Context(), services.SaveLessonReportInput{
		ActorUserID:            user.ID,
		LessonSessionID:        sessionID,
		TutorID:                tutor.ID,
		Status:                 c.PostForm("status"),
		MaterialSummary:        c.PostForm("material_summary"),
		ProgressSummary:        c.PostForm("progress_summary"),
		Homework:               c.PostForm("homework"),
		IssueNotes:             c.PostForm("issue_notes"),
		HomePracticeSuggestion: c.PostForm("home_practice_suggestion"),
		Publish:                publish,
		IPAddress:              c.ClientIP(),
	})
	if err != nil {
		h.renderTutorReportForm(c, user, http.StatusBadRequest, "Laporan belum bisa disimpan. Untuk publish, materi wajib diisi dan sesi harus milik Tutor yang sedang login.")
		return
	}
	c.Redirect(http.StatusFound, "/tutor/sessions")
}

func (h *AdminHandler) ParentReports(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var reports []models.LessonReport
	query := h.db.Preload("LessonSession").Preload("Tutor.User").Preload("Student.Parent").
		Joins("JOIN students ON students.id = lesson_reports.student_id").
		Where("students.parent_id = ? AND lesson_reports.published_at IS NOT NULL", user.ID).
		Order("lesson_reports.published_at desc")
	if studentID := c.Query("student_id"); studentID != "" {
		query = query.Where("lesson_reports.student_id = ?", studentID)
	}
	query.Find(&reports)
	var students []models.Student
	h.db.Where("parent_id = ?", user.ID).Order("name asc").Find(&students)
	c.HTML(http.StatusOK, "parent_reports.html", gin.H{"Title": "Laporan Progres Anak", "User": user, "Reports": reports, "Students": students, "Query": c.Request.URL.Query()})
}

func (h *AdminHandler) renderTutorReportForm(c *gin.Context, user *models.User, status int, errorMessage string) {
	tutor, err := h.currentTutor(user.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/tutor/sessions")
		return
	}
	var session models.LessonSession
	if err := h.db.Preload("Student.Parent").Preload("Report").First(&session, "id = ? AND tutor_id = ?", sessionID, tutor.ID).Error; err != nil {
		c.Redirect(http.StatusFound, "/tutor/sessions")
		return
	}
	c.HTML(status, "tutor_report_form.html", gin.H{"Title": "Isi Laporan", "User": user, "Session": session, "Error": errorMessage})
}

func (h *AdminHandler) TutorCreateReportForm(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	tutor, err := h.currentTutor(user.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	var students []models.Student
	h.db.Where("assigned_tutor_id = ?", tutor.ID).Order("name asc").Find(&students)
	today := time.Now().Format("2006-01-02")
	c.HTML(http.StatusOK, "tutor_create_report.html", gin.H{
		"Title":    "Buat Laporan",
		"User":     user,
		"Students": students,
		"Today":    today,
	})
}

func (h *AdminHandler) TutorCreateReport(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	tutor, err := h.currentTutor(user.ID)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	studentID, err := uuid.Parse(c.PostForm("student_id"))
	if err != nil {
		h.renderTutorCreateReportForm(c, user, tutor, http.StatusBadRequest, "Murid wajib dipilih.")
		return
	}
	publish := c.PostForm("publish") == "on"
	_, err = h.reports.CreateReportWithSession(c.Request.Context(), services.CreateReportWithSessionInput{
		ActorUserID:            user.ID,
		TutorID:                tutor.ID,
		StudentID:              studentID,
		Subject:                c.PostForm("subject"),
		Date:                   c.PostForm("date"),
		StartTime:              c.PostForm("start_time"),
		EndTime:                c.PostForm("end_time"),
		Location:               c.PostForm("location"),
		Status:                 c.PostForm("status"),
		MaterialSummary:        c.PostForm("material_summary"),
		ProgressSummary:        c.PostForm("progress_summary"),
		Homework:               c.PostForm("homework"),
		IssueNotes:             c.PostForm("issue_notes"),
		HomePracticeSuggestion: c.PostForm("home_practice_suggestion"),
		Publish:                publish,
		IPAddress:              c.ClientIP(),
	})
	if err != nil {
		h.renderTutorCreateReportForm(c, user, tutor, http.StatusBadRequest, "Laporan belum bisa disimpan. Pastikan murid terdaftar, tanggal valid, dan jam selesai setelah jam mulai. Untuk publish, materi wajib diisi.")
		return
	}
	c.Redirect(http.StatusFound, "/tutor/sessions")
}

func (h *AdminHandler) renderTutorCreateReportForm(c *gin.Context, user *models.User, tutor *models.Tutor, status int, errorMessage string) {
	var students []models.Student
	h.db.Where("assigned_tutor_id = ?", tutor.ID).Order("name asc").Find(&students)
	today := time.Now().Format("2006-01-02")
	c.HTML(status, "tutor_create_report.html", gin.H{
		"Title":    "Buat Laporan",
		"User":     user,
		"Students": students,
		"Today":    today,
		"Error":    errorMessage,
	})
}

func (h *AdminHandler) currentTutor(userID uuid.UUID) (*models.Tutor, error) {
	var tutor models.Tutor
	if err := h.db.Preload("User").First(&tutor, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &tutor, nil
}
