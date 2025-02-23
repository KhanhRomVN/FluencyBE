package middleware

import (
	"context"
	"database/sql"
	"fluencybe/pkg/utils"
	"strings"
	"sync"

	constants "fluencybe/internal/core/constants"

	"github.com/gin-gonic/gin"
)

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func validateToken(c *gin.Context) (*utils.Claims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, constants.ErrAuthHeaderRequired
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, constants.ErrInvalidAuthFormat
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix
	return utils.ValidateJWT(tokenString)
}

var (
	userCache    = make(map[string]bool)
	userCacheMux sync.RWMutex
)

func UserAuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)

		claims, err := validateToken(c)
		if err != nil {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if claims.Role != constants.RoleUser {
			c.JSON(403, StandardResponse{
				Success: false,
				Error:   "Access denied: User role required",
			})
			c.Abort()
			return
		}

		userCacheMux.RLock()
		exists, found := userCache[claims.UserID]
		userCacheMux.RUnlock()

		if !found {
			err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", claims.UserID).Scan(&exists)
			if err != nil {
				c.JSON(401, StandardResponse{
					Success: false,
					Error:   "Error verifying user",
				})
				c.Abort()
				return
			}

			userCacheMux.Lock()
			userCache[claims.UserID] = exists
			userCacheMux.Unlock()
		}

		if !exists {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   "User not found",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func DeveloperAuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)

		claims, err := validateToken(c)
		if err != nil {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if claims.Role != constants.RoleDeveloper {
			c.JSON(403, StandardResponse{
				Success: false,
				Error:   "Access denied: Developer role required",
			})
			c.Abort()
			return
		}

		var exists bool
		err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM developers WHERE id = $1)", claims.UserID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   "Developer not found",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func UserOrDeveloperAuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), constants.GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)

		claims, err := validateToken(c)
		if err != nil {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if claims.Role != constants.RoleUser && claims.Role != constants.RoleDeveloper {
			c.JSON(403, StandardResponse{
				Success: false,
				Error:   "Access denied: User or Developer role required",
			})
			c.Abort()
			return
		}

		var exists bool
		var query string
		if claims.Role == constants.RoleUser {
			query = "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
		} else {
			query = "SELECT EXISTS(SELECT 1 FROM developers WHERE id = $1)"
		}

		err = db.QueryRowContext(ctx, query, claims.UserID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(401, StandardResponse{
				Success: false,
				Error:   "User/Developer not found",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
