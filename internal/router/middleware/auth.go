package middleware

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: Implement proper session-based authentication
			// For now, redirect to signin page if not authenticated
			// In a real implementation, you'd check session or JWT tokens

			// Check if user is authenticated by looking for actor_id in session
			sess, _ := session.Get("session", c)
			actorID, ok := sess.Values["actor_id"].(int64)

			if !ok || actorID == 0 {
				// Not authenticated, redirect to signin
				return c.Redirect(http.StatusSeeOther, "/auth/signin")
			}

			// Store actor ID in context for handlers to use
			c.Set("actor_id", actorID)

			return next(c)
		}
	}
}

// GetActorIDFromContext gets actor ID from echo context
func GetActorIDFromContext(c echo.Context) int64 {
	actorID, ok := c.Get("actor_id").(int64)
	if !ok {
		return 0 // Return 0 to indicate no actor ID
	}
	return actorID
}
