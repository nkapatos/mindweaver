package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/store"
)

// AuthService handles authentication and session management
type AuthService struct {
	actorService *ActorService
	logger       *slog.Logger
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	ActorID   int64     `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ActorAuthMetadata represents authentication information stored in actor metadata
type ActorAuthMetadata struct {
	AuthStrategy string            `json:"auth_strategy"`
	Credentials  map[string]string `json:"credentials"`
	LastLogin    *string           `json:"last_login,omitempty"`
	IsActive     bool              `json:"is_active"`
}

// NewAuthService creates a new AuthService
func NewAuthService(actorService *ActorService) *AuthService {
	return &AuthService{
		actorService: actorService,
		logger:       slog.Default(),
	}
}

// CreateSession creates a new session for an actor
func (s *AuthService) CreateSession(ctx context.Context, actorID int64) (*Session, error) {
	s.logger.Info("Creating session", "actor_id", actorID)

	// Generate session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Create session with 24 hour expiration
	expiresAt := time.Now().Add(24 * time.Hour)

	session := &Session{
		ID:        sessionID,
		ActorID:   actorID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	s.logger.Info("Session created successfully", "session_id", sessionID, "actor_id", actorID)
	return session, nil
}

// GetSessionFromContext extracts session from Echo context
func (s *AuthService) GetSessionFromContext(c echo.Context) (*Session, error) {
	session := c.Get("session")
	if session == nil {
		return nil, fmt.Errorf("no session found in context")
	}

	// For now, we'll use a simple session storage
	// In production, this should use Redis or database
	sessionID := "mock-session"

	// TODO: Implement proper session storage and validation
	// For now, return a mock session with actor ID 1
	sess := &Session{
		ID:        sessionID,
		ActorID:   1, // Default to actor ID 1 for now
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return sess, nil
}

// GetSessionFromRequest extracts session from HTTP request
func (s *AuthService) GetSessionFromRequest(r *http.Request) (*Session, error) {
	// This will be handled by Echo's session middleware
	// For now, we'll use a simple session storage
	// In production, this should use Redis or database

	// TODO: Implement proper session storage and validation
	// For now, return a mock session with actor ID 1
	session := &Session{
		ID:        "mock-session",
		ActorID:   1, // Default to actor ID 1 for now
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return session, nil
}

// GetActorIDFromContext gets actor ID from Echo context
func (s *AuthService) GetActorIDFromContext(c echo.Context) (int64, error) {
	// Try to get actor ID from context first
	if actorID, ok := c.Get("actor_id").(int64); ok && actorID > 0 {
		return actorID, nil
	}

	// Try to get from session
	session := c.Get("session")
	if session == nil {
		return 0, fmt.Errorf("no session found")
	}

	// For now, return a mock actor ID (1) for testing
	// TODO: Implement proper session storage and retrieval
	return 1, nil
}

// GetActorIDFromRequest extracts actor ID from HTTP request
func (s *AuthService) GetActorIDFromRequest(r *http.Request) (int64, error) {
	session, err := s.GetSessionFromRequest(r)
	if err != nil {
		return 0, err
	}
	return session.ActorID, nil
}

// AuthenticateActor authenticates an actor by username and password
func (s *AuthService) AuthenticateActor(ctx context.Context, username, password string) (*store.Actor, error) {
	s.logger.Info("Authenticating actor", "username", username)

	// Get actor by username
	actor, err := s.actorService.GetActorByName(ctx, username, "user")
	if err != nil {
		s.logger.Error("Failed to get actor by name", "username", username, "error", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if actor has metadata for auth
	if !actor.Metadata.Valid {
		s.logger.Error("Actor has no auth metadata", "actor_id", actor.ID, "username", username)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Parse auth metadata
	var authMetadata ActorAuthMetadata
	if err := json.Unmarshal([]byte(actor.Metadata.String), &authMetadata); err != nil {
		s.logger.Error("Failed to parse auth metadata", "actor_id", actor.ID, "error", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if auth strategy is password-based
	if authMetadata.AuthStrategy != "password" {
		s.logger.Error("Unsupported auth strategy", "actor_id", actor.ID, "strategy", authMetadata.AuthStrategy)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if actor is active
	if !authMetadata.IsActive {
		s.logger.Error("Actor is not active", "actor_id", actor.ID, "username", username)
		return nil, fmt.Errorf("account is disabled")
	}

	// TODO: Implement proper password hashing and verification
	// For now, check against stored credentials (in production, this should be hashed)
	storedPassword, exists := authMetadata.Credentials["password"]
	if !exists {
		s.logger.Error("No password found in credentials", "actor_id", actor.ID, "username", username)
		return nil, fmt.Errorf("invalid credentials")
	}

	if storedPassword != password {
		s.logger.Error("Password mismatch", "actor_id", actor.ID, "username", username)
		return nil, fmt.Errorf("invalid credentials")
	}

	// TODO: Update last login timestamp in metadata
	// TODO: Implement proper session management with database storage

	s.logger.Info("Actor authenticated successfully", "actor_id", actor.ID, "username", username)
	return &actor, nil
}

// generateSessionID generates a random session ID
func (s *AuthService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SetSessionCookie sets the session cookie on the response
func (s *AuthService) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})
}

// ClearSessionCookie clears the session cookie
func (s *AuthService) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
