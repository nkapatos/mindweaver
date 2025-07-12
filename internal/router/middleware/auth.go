package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: Implement proper session-based authentication
			// For now, we'll use a simple approach with actor ID 1 (test user)
			// In a real implementation, you'd check session or JWT tokens

			// Check if user is authenticated (this is a placeholder)
			// You can implement your own authentication logic here
			actorID := int64(2) // Default to test user actor for now (ID 2)

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
