package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/models"
	"github.com/umbranodens/kalabelajar/internal/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const SessionCookieName = "kb_session"
const googleStateCookieName = "kb_google_state"

type AuthHandler struct {
	auth        *services.AuthService
	googleOAuth *oauth2.Config
	httpClient  *http.Client
}

func NewAuthHandler(auth *services.AuthService, googleCfg config.GoogleConfig, appURL string) *AuthHandler {
	redirectURL := googleCfg.RedirectURL
	if redirectURL == "" && appURL != "" {
		redirectURL = appURL + "/auth/google/callback"
	}
	return &AuthHandler{
		auth: auth,
		googleOAuth: &oauth2.Config{
			ClientID:     googleCfg.ClientID,
			ClientSecret: googleCfg.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		httpClient: http.DefaultClient,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	router.GET("/login", h.ShowLogin)
	router.POST("/login", h.Login)
	router.GET("/register", h.ShowRegister)
	router.POST("/register", h.Register)
	router.GET("/auth/google", h.GoogleStart)
	router.GET("/auth/google/callback", h.GoogleCallback)
	router.POST("/logout", h.Logout)
	router.GET("/pending", h.Pending)
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"Title":         "Masuk - Kala Belajar",
		"GoogleEnabled": h.googleConfigured(),
	})
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
			"Title":         "Masuk - Kala Belajar",
			"Error":         loginErrorMessage(err),
			"Email":         c.PostForm("email"),
			"GoogleEnabled": h.googleConfigured(),
		})
		return
	}

	setSessionCookie(c, result.Session.Token)
	redirectAfterAuth(c, result.User)
}

func (h *AuthHandler) GoogleStart(c *gin.Context) {
	if !h.googleConfigured() {
		c.HTML(http.StatusServiceUnavailable, "login.tmpl", gin.H{
			"Title": "Masuk - Kala Belajar",
			"Error": "Login Google belum dikonfigurasi.",
		})
		return
	}
	state, err := newOAuthState()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.tmpl", gin.H{
			"Title":         "Masuk - Kala Belajar",
			"Error":         "Login Google belum bisa dimulai.",
			"GoogleEnabled": true,
		})
		return
	}
	c.SetCookie(googleStateCookieName, state, 300, "/", "", false, true)
	c.Redirect(http.StatusFound, h.googleOAuth.AuthCodeURL(state, oauth2.AccessTypeOffline))
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if !h.googleConfigured() {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	expectedState, err := c.Cookie(googleStateCookieName)
	if err != nil || expectedState == "" || c.Query("state") != expectedState {
		c.HTML(http.StatusBadRequest, "login.tmpl", gin.H{
			"Title":         "Masuk - Kala Belajar",
			"Error":         "Sesi login Google tidak valid. Coba lagi ya.",
			"GoogleEnabled": true,
		})
		return
	}
	c.SetCookie(googleStateCookieName, "", -1, "/", "", false, true)

	token, err := h.googleOAuth.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.tmpl", gin.H{
			"Title":         "Masuk - Kala Belajar",
			"Error":         "Kode Google tidak valid.",
			"GoogleEnabled": true,
		})
		return
	}
	profile, err := h.fetchGoogleProfile(c.Request.Context(), token)
	if err != nil {
		c.HTML(http.StatusBadGateway, "login.tmpl", gin.H{
			"Title":         "Masuk - Kala Belajar",
			"Error":         "Profil Google belum bisa dibaca.",
			"GoogleEnabled": true,
		})
		return
	}
	result, err := h.auth.LoginWithGoogleProfile(c.Request.Context(), profile, services.OAuthLoginInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.tmpl", gin.H{
			"Title":         "Masuk - Kala Belajar",
			"Error":         loginErrorMessage(err),
			"GoogleEnabled": true,
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

func (h *AuthHandler) googleConfigured() bool {
	return h.googleOAuth != nil &&
		h.googleOAuth.ClientID != "" &&
		h.googleOAuth.ClientSecret != "" &&
		h.googleOAuth.RedirectURL != ""
}

func (h *AuthHandler) fetchGoogleProfile(ctx context.Context, token *oauth2.Token) (services.GoogleProfile, error) {
	client := h.googleOAuth.Client(ctx, token)
	if h.httpClient != nil {
		client = h.httpClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return services.GoogleProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return services.GoogleProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return services.GoogleProfile{}, fmt.Errorf("google userinfo returned %d", resp.StatusCode)
	}
	var payload struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return services.GoogleProfile{}, err
	}
	return services.GoogleProfile{
		ProviderID: payload.ID,
		Email:      payload.Email,
		Name:       payload.Name,
		AvatarURL:  payload.Picture,
	}, nil
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

func newOAuthState() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
