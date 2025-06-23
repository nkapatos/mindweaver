package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nkapatos/mindweaver/internal/services"
)

type PromptHandler struct {
	promptService *services.PromptService
}

// NewPromptHandler creates a new PromptHandler with dependency injection
func NewPromptHandler(promptService *services.PromptService) *PromptHandler {
	return &PromptHandler{
		promptService: promptService,
	}
}

// CreatePrompt handles POST /api/prompts
func (h *PromptHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   *int64 `json:"user_id,omitempty"`
		Title    string `json:"title"`
		Content  string `json:"content"`
		IsSystem bool   `json:"is_system"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	if err := h.promptService.CreatePrompt(r.Context(), req.UserID, req.Title, req.Content, req.IsSystem); err != nil {
		http.Error(w, "Failed to create prompt", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetPrompt handles GET /api/prompts/{id}
func (h *PromptHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Prompt ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	prompt, err := h.promptService.GetPromptByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Prompt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

// GetAllPrompts handles GET /api/prompts
func (h *PromptHandler) GetAllPrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompts, err := h.promptService.GetAllPrompts(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch prompts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// GetPromptsByUserID handles GET /api/prompts/by-user?user_id={id}
func (h *PromptHandler) GetPromptsByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	prompts, err := h.promptService.GetPromptsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch prompts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// GetSystemPrompts handles GET /api/prompts/system
func (h *PromptHandler) GetSystemPrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompts, err := h.promptService.GetSystemPrompts(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch system prompts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// UpdatePrompt handles PUT /api/prompts/{id}
func (h *PromptHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Prompt ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		IsSystem bool   `json:"is_system"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	if err := h.promptService.UpdatePrompt(r.Context(), id, req.Title, req.Content, req.IsSystem); err != nil {
		http.Error(w, "Failed to update prompt", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePrompt handles DELETE /api/prompts/{id}
func (h *PromptHandler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Prompt ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	if err := h.promptService.DeletePrompt(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete prompt", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
