package account

import (
	"context"
	"encoding/json"
	accountDTO "fluencybe/internal/app/dto"
	accountModel "fluencybe/internal/app/model/account"
	accountService "fluencybe/internal/app/service/account"
	"fluencybe/internal/core/constants"
	"fluencybe/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	service *accountService.UserService
	logger  *logger.PrettyLogger
}

func NewUserHandler(service *accountService.UserService) *UserHandler {
	enabledLevels := make(map[logger.LogLevel]bool)
	enabledLevels[logger.LevelDebug] = true
	enabledLevels[logger.LevelInfo] = true
	enabledLevels[logger.LevelWarning] = true
	enabledLevels[logger.LevelError] = true

	return &UserHandler{
		service: service,
		logger:  logger.GetGlobalLogger(),
	}
}

func (h *UserHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("register_start", nil, "Starting user registration process")

	var req accountDTO.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("register_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := &accountModel.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Type:     req.Type,
	}

	h.logger.Debug("register_attempt", map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
		"type":     req.Type,
	}, "Attempting to register new user")

	if err := h.service.Register(ctx, user); err != nil {
		h.logger.Error("register_failed", map[string]interface{}{
			"error":    err.Error(),
			"email":    req.Email,
			"username": req.Username,
		}, "User registration failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("register_success", map[string]interface{}{
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"username": user.Username,
	}, "User registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Type:      user.Type,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UserHandler) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("login_start", nil, "Starting user login process")

	var req accountDTO.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("login_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("login_attempt", map[string]interface{}{
		"email": req.Email,
	}, "Attempting user login")

	user, token, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("login_failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		}, "Login failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(accountDTO.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("login_success", map[string]interface{}{
		"email": req.Email,
	}, "User logged in successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.LoginResponse{
		ID:    user.ID,
		Token: token,
	})
}

func (h *UserHandler) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_user_start", nil, "Starting to get user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("get_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from URL parameter
	userID := ginCtx.Param("id")
	if userID == "" {
		h.logger.Error("get_user_invalid_id", nil, "User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("get_user_attempt", map[string]interface{}{
		"user_id": userID,
	}, "Attempting to get user")

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		h.logger.Error("get_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to get user")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Info("get_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UserHandler) UpdateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update_user_start", nil, "Starting to update user")

	vars := mux.Vars(r)
	userID := vars["id"]

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("update_user_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("update_user_attempt", map[string]interface{}{
		"user_id": userID,
		"field":   req.Field,
		"value":   req.Value,
	}, "Attempting to update user")

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	if err := h.service.UpdateUser(ctx, userID, updates); err != nil {
		h.logger.Error("update_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to update user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("update_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User updated successfully")

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) DeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete_user_start", nil, "Starting to delete user")

	vars := mux.Vars(r)
	userID := vars["id"]

	h.logger.Debug("delete_user_attempt", map[string]interface{}{
		"user_id": userID,
	}, "Attempting to delete user")

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		h.logger.Error("delete_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("delete_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_my_user_start", nil, "Starting to get my user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("get_my_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("get_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("get_my_user_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("get_my_user_attempt", map[string]interface{}{
		"user_id": userIDStr,
	}, "Attempting to get my user")

	user, err := h.service.GetUser(ctx, userIDStr)
	if err != nil {
		h.logger.Error("get_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to get my user")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Info("get_my_user_success", map[string]interface{}{
		"user_id": userIDStr,
	}, "My user retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UserHandler) UpdateMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update_my_user_start", nil, "Starting to update my user")

	// Get gin context from standard context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("update_my_user_invalid_context", nil, "Failed to get gin context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, exists := ginCtx.Get("user_id")
	if !exists {
		h.logger.Error("update_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("update_my_user_invalid_id_type", nil, "Invalid user ID type")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("update_my_user_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("update_my_user_attempt", map[string]interface{}{
		"user_id": userIDStr,
		"field":   req.Field,
		"value":   req.Value,
	}, "Attempting to update my user")

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	if err := h.service.UpdateUser(ctx, userIDStr, updates); err != nil {
		h.logger.Error("update_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to update my user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("update_my_user_success", map[string]interface{}{
		"user_id": userIDStr,
	}, "My user updated successfully")

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) DeleteMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete_my_user_start", nil, "Starting to delete my user")

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error("delete_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Debug("delete_my_user_attempt", map[string]interface{}{
		"user_id": userID,
	}, "Attempting to delete my user")

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		h.logger.Error("delete_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete my user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("delete_my_user_success", map[string]interface{}{
		"user_id": userID,
	}, "My user deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetListUserWithPagination(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_list_user_start", nil, "Starting to get list of users")

	// Get pagination parameters from query string
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	h.logger.Debug("get_list_user_attempt", map[string]interface{}{
		"page":  page,
		"limit": limit,
	}, "Attempting to get list of users")

	users, total, err := h.service.GetUserList(ctx, page, limit)
	if err != nil {
		h.logger.Error("get_list_user_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get list of users")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("get_list_user_success", map[string]interface{}{
		"total": total,
	}, "List of users retrieved successfully")

	// Convert users to response DTOs
	var userResponses []accountDTO.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, accountDTO.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userResponses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
