package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type ActorHandler struct {
	actorService *services.ActorService
}

// NewActorHandler creates a new ActorHandler with dependency injection
func NewActorHandler(actorService *services.ActorService) *ActorHandler {
	return &ActorHandler{
		actorService: actorService,
	}
}

// CreateActor handles POST /api/actors
func (h *ActorHandler) CreateActor(c echo.Context) error {
	var req struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Metadata    string `json:"metadata"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and type are required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	createdBy, _ := sess.Values["actor_id"].(int64)

	if err := h.actorService.CreateActor(c.Request().Context(), req.Type, req.Name, req.DisplayName, req.AvatarURL, req.Metadata, req.IsActive, createdBy, createdBy); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create actor"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Actor created successfully"})
}

// GetActor handles GET /api/actors/{id}
func (h *ActorHandler) GetActor(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Actor ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid actor ID"})
	}

	actor, err := h.actorService.GetActorByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Actor not found"})
	}

	return c.JSON(http.StatusOK, actor)
}

// GetActorByName handles GET /api/actors/by-name?name={name}&type={type}
func (h *ActorHandler) GetActorByName(c echo.Context) error {
	name := c.QueryParam("name")
	actorType := c.QueryParam("type")
	if name == "" || actorType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and type are required"})
	}

	actor, err := h.actorService.GetActorByName(c.Request().Context(), name, actorType)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Actor not found"})
	}

	return c.JSON(http.StatusOK, actor)
}

// GetActorsByType handles GET /api/actors/by-type?type={type}
func (h *ActorHandler) GetActorsByType(c echo.Context) error {
	actorType := c.QueryParam("type")
	if actorType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Type is required"})
	}

	actors, err := h.actorService.GetActorsByType(c.Request().Context(), actorType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch actors"})
	}

	return c.JSON(http.StatusOK, actors)
}

// UpdateActor handles PUT /api/actors/{id}
func (h *ActorHandler) UpdateActor(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Actor ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid actor ID"})
	}

	var req struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Metadata    string `json:"metadata"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	updatedBy, _ := sess.Values["actor_id"].(int64)

	if err := h.actorService.UpdateActor(c.Request().Context(), id, req.Name, req.Type, req.DisplayName, req.AvatarURL, req.Metadata, req.IsActive, updatedBy); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update actor"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Actor updated successfully"})
}

// DeleteActor handles DELETE /api/actors/{id}
func (h *ActorHandler) DeleteActor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Actor ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		return
	}

	if err := h.actorService.DeleteActor(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete actor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
