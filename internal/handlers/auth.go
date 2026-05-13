package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
)

const SessionCookieName = "kb_session"

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	router.GET("/login", h.ShowLogin)
	router.POST("/login", h.Login)
	router.GET("/register", h.ShowRegister)
	router.POST("/register", h.Register)
	router.POST("/logout", h.Logout)
	router.GET("/pending", h.Pending)
	router.GET("/dashboard", h.Dashboard)
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{"Title": "Masuk - Kala Belajar"})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.tmpl", gin.H{"Title": "Daftar - Kala Belajar"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	result, err := h.auth.Register(c.Request.Context(), services.RegisterInput{
		Name:                 c.PostForm("name"),
		Email:                c.PostForm("email"),
		Password:             c.PostForm("password"),
		PasswordConfirmation: c.PostForm("password_confirmation"),
		IPAddress:            c.ClientIP(),
		UserAgent:            c.Request.UserAgent(),
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "register.tmpl", gin.H{
			"Title": "Daftar - Kala Belajar",
			"Error": registerErrorMessage(err),
			"Name":  c.PostForm("name"),
			"Email": c.PostForm("email"),
		})
		return
	}

	setSessionCookie(c, result.Session.Token)
	redirectAfterAuth(c, result.User)
}

func (h *AuthHandler) Login(c *gin.Context) {
	result, err := h.auth.Login(c.Request.Context(), services.LoginInput{
		Email:     c.PostForm("email"),
		Password:  c.PostForm("password"),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.tmpl", gin.H{
			"Title": "Masuk - Kala Belajar",
			"Error": loginErrorMessage(err),
			"Email": c.PostForm("email"),
		})
		return
	}

	setSessionCookie(c, result.Session.Token)
	redirectAfterAuth(c, result.User)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) Pending(c *gin.Context) {
	session, ok := h.currentSession(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "pending.tmpl", gin.H{
		"Title": "Menunggu Role - Kala Belajar",
		"User":  session.User,
	})
}

func (h *AuthHandler) Dashboard(c *gin.Context) {
	session, ok := h.currentSession(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if session.User.RoleID == nil {
		c.Redirect(http.StatusFound, "/pending")
		return
	}

	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Title":     "Dashboard - Kala Belajar",
		"User":      session.User,
		"RoleLabel": roleLabel(session.User.RoleID),
	})
}

func (h *AuthHandler) currentSession(c *gin.Context) (*models.Session, bool) {
	token, err := c.Cookie(SessionCookieName)
	if err != nil {
		return nil, false
	}
	session, err := h.auth.FindSession(c.Request.Context(), token)
	if err != nil || !session.User.IsActive {
		return nil, false
	}
	return session, true
}

func setSessionCookie(c *gin.Context, token string) {
	c.SetCookie(SessionCookieName, token, 60*60*24, "/", "", false, true)
}

func redirectAfterAuth(c *gin.Context, user *models.User) {
	if user.RoleID == nil {
		c.Redirect(http.StatusFound, "/pending")
		return
	}
	c.Redirect(http.StatusFound, "/dashboard")
}

func roleLabel(roleID *uint) string {
	if roleID == nil {
		return "Menunggu Role"
	}
	switch *roleID {
	case models.RoleIDSuperAdmin:
		return "Super Admin"
	case models.RoleIDTutor:
		return "Tutor"
	case models.RoleIDParent:
		return "Parent"
	default:
		return "Role belum dikenal"
	}
}

func registerErrorMessage(err error) string {
	switch {
	case errors.Is(err, services.ErrEmailAlreadyUsed):
		return "Email sudah digunakan."
	case errors.Is(err, services.ErrPasswordConfirmationMismatch):
		return "Konfirmasi password belum cocok."
	case errors.Is(err, services.ErrInvalidRegisterInput):
		return "Ada data yang belum lengkap atau password kurang dari 8 karakter."
	default:
		return "Registrasi belum berhasil. Coba lagi sebentar."
	}
}

func loginErrorMessage(err error) string {
	switch {
	case errors.Is(err, services.ErrInactiveUser):
		return "Akun ini sedang nonaktif."
	case errors.Is(err, services.ErrPasswordLoginUnavailable):
		return "Akun ini memakai Google SSO. Masuk dengan Google ya."
	default:
		return "Email atau password belum cocok."
	}
}
