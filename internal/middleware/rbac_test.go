package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbranodens/kalabelajar/internal/middleware"
	"github.com/umbranodens/kalabelajar/internal/models"
)

func TestRequireRolesAllowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	roleID := models.RoleIDSuperAdmin
	source := fakeSessionSource{session: &models.Session{
		Token: "valid",
		User:  models.User{ID: uuid.New(), RoleID: &roleID, IsActive: true},
	}}
	router := gin.New()
	router.Use(middleware.LoadSession(source))
	router.GET("/admin", middleware.RequireRoles(models.RoleIDSuperAdmin), func(c *gin.Context) {
		user, ok := middleware.CurrentUser(c)
		require.True(t, ok)
		assert.Equal(t, roleID, *user.RoleID)
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "valid"})
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "ok", recorder.Body.String())
}

func TestRequireRolesRejectsWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	roleID := models.RoleIDParent
	source := fakeSessionSource{session: &models.Session{
		Token: "valid",
		User:  models.User{ID: uuid.New(), RoleID: &roleID, IsActive: true},
	}}
	router := gin.New()
	router.Use(middleware.LoadSession(source))
	router.GET("/admin", middleware.RequireRoles(models.RoleIDSuperAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "valid"})
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestRequireAuthRedirectsMissingSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.LoadSession(fakeSessionSource{}))
	router.GET("/dashboard", middleware.RequireAuth(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, "/login", recorder.Header().Get("Location"))
}

type fakeSessionSource struct {
	session *models.Session
}

func (f fakeSessionSource) FindValidByToken(ctx context.Context, token string, now time.Time) (*models.Session, error) {
	if f.session == nil || f.session.Token != token {
		return nil, assert.AnError
	}
	return f.session, nil
}
