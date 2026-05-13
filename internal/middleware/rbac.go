package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umbranodens/kalabelajar/internal/models"
)

const (
	SessionCookieName = "kb_session"
	currentUserKey    = "current_user"
)

type SessionSource interface {
	FindValidByToken(ctx context.Context, token string, now time.Time) (*models.Session, error)
}

func LoadSession(source SessionSource) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err == nil && token != "" {
			session, sessionErr := source.FindValidByToken(c.Request.Context(), token, time.Now().UTC())
			if sessionErr == nil && session.User.IsActive {
				c.Set(currentUserKey, &session.User)
			}
		}
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUser(c); !ok {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireRoles(allowedRoleIDs ...uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		if user.RoleID == nil {
			c.Redirect(http.StatusFound, "/pending")
			c.Abort()
			return
		}
		for _, roleID := range allowedRoleIDs {
			if *user.RoleID == roleID {
				c.Next()
				return
			}
		}
		c.String(http.StatusForbidden, "Akses ditolak.")
		c.Abort()
	}
}

func CurrentUser(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return nil, false
	}
	user, ok := value.(*models.User)
	return user, ok && user != nil
}
