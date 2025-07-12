package web

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SignInPage renders the sign in page
func (h *AuthHandler) SignInPage(c echo.Context) error {
	return views.SignIn().Render(c.Request().Context(), c.Response().Writer)
}

// SignIn handles POST /auth/signin
func (h *AuthHandler) SignIn(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Validate required fields
	if username == "" || password == "" {
		return c.Redirect(http.StatusSeeOther, "/auth/signin?error=Username and password are required")
	}

	// Authenticate actor
	actor, err := h.authService.AuthenticateActor(c.Request().Context(), username, password)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/signin?error=Invalid credentials")
	}

	// Store actor ID in session using gorilla/sessions
	sess, _ := session.Get("session", c)
	sess.Values["actor_id"] = actor.ID
	sess.Save(c.Request(), c.Response())

	// Redirect to home page
	return c.Redirect(http.StatusSeeOther, "/")
}

// SignUpPage renders the sign up page
func (h *AuthHandler) SignUpPage(c echo.Context) error {
	// TODO: Implement sign up page
	return c.String(http.StatusNotImplemented, "Sign up not implemented yet")
}

// SignUp handles POST /auth/signup
func (h *AuthHandler) SignUp(c echo.Context) error {
	// TODO: Implement sign up
	return c.String(http.StatusNotImplemented, "Sign up not implemented yet")
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear session using gorilla/sessions
	sess, _ := session.Get("session", c)
	delete(sess.Values, "actor_id")
	sess.Save(c.Request(), c.Response())

	// Redirect to sign in page
	return c.Redirect(http.StatusSeeOther, "/auth/signin")
}

// RequireAuth middleware that requires authentication
func (h *AuthHandler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Try to get actor ID from session
		sess, _ := session.Get("session", c)
		actorID, ok := sess.Values["actor_id"].(int64)
		if !ok || actorID == 0 {
			// Redirect to sign in page if not authenticated
			return c.Redirect(http.StatusSeeOther, "/auth/signin")
		}

		// Store actor ID in context for handlers to use
		c.Set("actor_id", actorID)

		return next(c)
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
