package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

// SetupData represents the data needed for the setup form
type SetupData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ActorAuthMetadata represents authentication information stored in actor metadata
type ActorAuthMetadata struct {
	AuthStrategy string            `json:"auth_strategy"`
	Credentials  map[string]string `json:"credentials"`
	LastLogin    *string           `json:"last_login,omitempty"`
	IsActive     bool              `json:"is_active"`
}

type SetupHandler struct {
	actorService *services.ActorService
}

func NewSetupHandler(actorService *services.ActorService) *SetupHandler {
	return &SetupHandler{
		actorService: actorService,
	}
}

// SetupPage handles GET /setup - shows the initial setup wizard
func (h *SetupHandler) SetupPage(c echo.Context) error {
	// Check if setup is needed (database exists and has actors)
	needsSetup, err := h.checkIfSetupNeeded()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check setup status")
	}

	if !needsSetup {
		// Setup already completed, redirect to home
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return views.SetupPage().Render(c.Request().Context(), c.Response().Writer)
}

// SetupApplication handles POST /setup - processes the setup form
func (h *SetupHandler) SetupApplication(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Validate required fields
	if username == "" || password == "" {
		return c.Redirect(http.StatusSeeOther, "/setup?error=Username and password are required")
	}

	// Create auth metadata for the user
	authMetadata := ActorAuthMetadata{
		AuthStrategy: "password",
		Credentials: map[string]string{
			"username": username,
			"password": password, // In production, this would be hashed
		},
		IsActive: true,
	}

	// Serialize auth metadata to JSON
	metadataJSON, err := json.Marshal(authMetadata)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user metadata")
	}

	// Create the user actor
	err = h.actorService.CreateActor(
		c.Request().Context(),
		"user",
		username,
		username, // Display name same as username for now
		"",
		string(metadataJSON), // Store auth info in metadata
		true,
		1, // created_by - system actor
		1, // updated_by - system actor
	)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/setup?error=Failed to create user: "+err.Error())
	}

	// Create a setup completion marker file
	setupMarker := "setup_completed"
	if err := os.WriteFile(setupMarker, []byte("setup completed"), 0644); err != nil {
		// Log but don't fail the setup
		c.Logger().Error("Failed to create setup marker file", "error", err)
	}

	return c.Redirect(http.StatusSeeOther, "/auth/signin?success=Setup completed successfully. Please sign in with your credentials.")
}

// checkIfSetupNeeded checks if the application needs initial setup
func (h *SetupHandler) checkIfSetupNeeded() (bool, error) {
	// Check if setup marker file exists
	if _, err := os.Stat("setup_completed"); err == nil {
		// Setup marker exists, setup is not needed
		return false, nil
	}

	// Check if any user actors exist
	actors, err := h.actorService.GetActorsByType(context.Background(), "user")
	if err != nil {
		// If we can't get actors, assume setup is needed
		return true, nil
	}

	// If no user actors exist, setup is needed
	return len(actors) == 0, nil
}
