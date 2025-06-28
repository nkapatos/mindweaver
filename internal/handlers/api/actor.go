package api

import (
	"encoding/json"
	"net/http"
	"strconv"

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
func (h *ActorHandler) CreateActor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Metadata    string `json:"metadata"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	if err := h.actorService.CreateActor(r.Context(), req.Type, req.Name, req.DisplayName, req.AvatarURL, req.Metadata, req.IsActive); err != nil {
		http.Error(w, "Failed to create actor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetActor handles GET /api/actors/{id}
func (h *ActorHandler) GetActor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	actor, err := h.actorService.GetActorByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Actor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actor)
}

// GetActorByName handles GET /api/actors/by-name?name={name}&type={type}
func (h *ActorHandler) GetActorByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	actorType := r.URL.Query().Get("type")
	if name == "" || actorType == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	actor, err := h.actorService.GetActorByName(r.Context(), name, actorType)
	if err != nil {
		http.Error(w, "Actor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actor)
}

// GetActorsByType handles GET /api/actors/by-type?type={type}
func (h *ActorHandler) GetActorsByType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actorType := r.URL.Query().Get("type")
	if actorType == "" {
		http.Error(w, "Type is required", http.StatusBadRequest)
		return
	}

	actors, err := h.actorService.GetActorsByType(r.Context(), actorType)
	if err != nil {
		http.Error(w, "Failed to fetch actors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors)
}

// UpdateActor handles PUT /api/actors/{id}
func (h *ActorHandler) UpdateActor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	var req struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Metadata    string `json:"metadata"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if err := h.actorService.UpdateActor(r.Context(), id, req.Name, req.Type, req.DisplayName, req.AvatarURL, req.Metadata, req.IsActive); err != nil {
		http.Error(w, "Failed to update actor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
