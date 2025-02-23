package account

import (
	"context"
	"encoding/json"
	accountDTO "fluencybe/internal/app/dto"
	accountModel "fluencybe/internal/app/model/account"
	accountService "fluencybe/internal/app/service/account"
	"fluencybe/internal/core/constants"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DeveloperHandler struct {
	service *accountService.DeveloperService
}

func NewDeveloperHandler(service *accountService.DeveloperService) *DeveloperHandler {
	return &DeveloperHandler{service: service}
}

func (h *DeveloperHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req accountDTO.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dev := &accountModel.Developer{
		ID:       uuid.New(),
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.service.Register(ctx, dev); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        dev.ID,
		Email:     dev.Email,
		Username:  dev.Username,
		CreatedAt: dev.CreatedAt,
	})
}

func (h *DeveloperHandler) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req accountDTO.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dev, token, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.LoginResponse{
		ID:    dev.ID,
		Token: token,
	})
}

func (h *DeveloperHandler) GetDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	devID := vars["id"]

	dev, err := h.service.GetDeveloper(ctx, devID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        dev.ID,
		Email:     dev.Email,
		Username:  dev.Username,
		CreatedAt: dev.CreatedAt,
	})
}

func (h *DeveloperHandler) UpdateDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	devID := vars["id"]

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	if err := h.service.UpdateDeveloper(ctx, devID, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DeveloperHandler) DeleteDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	devID := vars["id"]

	if err := h.service.DeleteDeveloper(ctx, devID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeveloperHandler) GetMyDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get developer ID from gin context
	devID, ok := ginCtx.Get("user_id") // Note: middleware sets both user and developer IDs as "user_id"
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert devID to string
	devIDStr, ok := devID.(string)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dev, err := h.service.GetDeveloper(ctx, devIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        dev.ID,
		Email:     dev.Email,
		Username:  dev.Username,
		CreatedAt: dev.CreatedAt,
	})
}

func (h *DeveloperHandler) UpdateMyDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	devID, ok := ginCtx.Get("user_id")
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	devIDStr, ok := devID.(string)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	if err := h.service.UpdateDeveloper(ctx, devIDStr, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DeveloperHandler) DeleteMyDeveloper(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get developer ID from gin context
	devID, ok := ginCtx.Get("user_id") // Note: middleware sets both user and developer IDs as "user_id"
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert devID to string
	devIDStr, ok := devID.(string)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.service.DeleteDeveloper(ctx, devIDStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeveloperHandler) GetListDeveloperWithPagination(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters from query string
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	developers, total, err := h.service.GetDeveloperList(ctx, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert developers to response DTOs
	var developerResponses []accountDTO.UserResponse
	for _, dev := range developers {
		developerResponses = append(developerResponses, accountDTO.UserResponse{
			ID:        dev.ID,
			Email:     dev.Email,
			Username:  dev.Username,
			CreatedAt: dev.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"developers": developerResponses,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}
