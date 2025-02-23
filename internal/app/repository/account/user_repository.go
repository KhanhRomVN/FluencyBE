package account

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fluencybe/internal/app/model/account"
	"fluencybe/internal/core/status"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SafeUser là struct để lưu cache, không chứa password
type SafeUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toSafeUser(user *account.User) *SafeUser {
	return &SafeUser{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Type:      user.Type,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

const (
	userCachePrefix   = "user:"
	userCacheDuration = 15 * time.Minute
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInvalidInput   = errors.New("invalid input data")
	ErrUserDuplicateEmail = errors.New("email already exists")
)

type UserRepository struct {
	db     *sql.DB
	cache  cache.Cache
	logger *logger.PrettyLogger
}

func NewUserRepository(db *sql.DB, cache cache.Cache, logger *logger.PrettyLogger) *UserRepository {
	return &UserRepository{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (r *UserRepository) getCacheKey(id interface{}) string {
	return fmt.Sprintf("%s%v", userCachePrefix, id)
}

// tryCache attempts to perform a cache operation only if Redis is healthy
func (r *UserRepository) tryCache(operation func() error) {
	if status.GetRedisStatus() {
		if err := operation(); err != nil {
			r.logger.Warning("CACHE_OPERATION_FAILED", map[string]interface{}{
				"error": err.Error(),
			}, "Cache operation failed, continuing without cache")
		}
	}
}

func (r *UserRepository) Create(ctx context.Context, user *account.User) error {
	r.logger.Info("create_user_start", map[string]interface{}{
		"email": user.Email,
	}, "Starting to create user in database")

	if user == nil {
		r.logger.Error("create_user_invalid_input", nil, "User object is nil")
		return ErrUserInvalidInput
	}

	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("create_user_tx_begin_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Prepare the insert statement
	query := `
		INSERT INTO users (id, email, username, password, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	r.logger.Debug("create_user_execute", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	}, "Executing create user query")

	// Execute the statement
	_, err = tx.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.Password,
		user.Type,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Rollback the transaction
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("create_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}

		// Check for duplicate email
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.logger.Error("create_user_duplicate_email", map[string]interface{}{
				"email": user.Email,
			}, "Duplicate email address")
			return ErrUserDuplicateEmail
		}

		r.logger.Error("create_user_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create user")
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error("create_user_commit_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("create_user_success", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
	}, "User created successfully")

	// Only try to cache if Redis is healthy
	r.tryCache(func() error {
		safeUser := toSafeUser(user)
		userJSON, err := json.Marshal(safeUser)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %v", err)
		}
		return r.cache.Set(ctx, r.getCacheKey(user.ID), string(userJSON), userCacheDuration)
	})

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*account.User, error) {
	r.logger.Info("get_user_by_id_start", map[string]interface{}{
		"user_id": id.String(),
	}, "Starting to get user by ID")

	// Try to get from cache first if Redis is healthy
	if status.GetRedisStatus() {
		cacheKey := r.getCacheKey(id)
		var safeUser SafeUser
		cachedData, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &safeUser); err == nil {
				r.logger.Debug("get_user_by_id_cache_hit", map[string]interface{}{
					"user_id": id.String(),
				}, "User found in cache")

				// Convert SafeUser to User
				return &account.User{
					ID:        safeUser.ID,
					Email:     safeUser.Email,
					Username:  safeUser.Username,
					Type:      safeUser.Type,
					CreatedAt: safeUser.CreatedAt,
					UpdatedAt: safeUser.UpdatedAt,
				}, nil
			}
		}
	}

	r.logger.Debug("get_user_by_id_cache_miss", map[string]interface{}{
		"user_id": id.String(),
	}, "User not found in cache, querying database")

	// Query the database
	query := `
		SELECT id, email, username, password, type, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &account.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Type,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("get_user_by_id_not_found", map[string]interface{}{
				"user_id": id.String(),
			}, "User not found")
			return nil, ErrUserNotFound
		}
		r.logger.Error("get_user_by_id_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		}, "Failed to get user")
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	r.logger.Info("get_user_by_id_success", map[string]interface{}{
		"user_id": id.String(),
	}, "User retrieved successfully")

	// Cache the result if Redis is healthy
	r.tryCache(func() error {
		safeUser := toSafeUser(user)
		userJSON, err := json.Marshal(safeUser)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %v", err)
		}
		return r.cache.Set(ctx, r.getCacheKey(id), string(userJSON), userCacheDuration)
	})

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*account.User, error) {
	r.logger.Info("get_user_by_email_start", map[string]interface{}{
		"email": email,
	}, "Starting to get user by email")

	query := `
		SELECT id, email, username, password, type, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &account.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Type,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("get_user_by_email_not_found", map[string]interface{}{
				"email": email,
			}, "User not found")
			return nil, ErrUserNotFound
		}
		r.logger.Error("get_user_by_email_failed", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		}, "Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	r.logger.Info("get_user_by_email_success", map[string]interface{}{
		"email":   email,
		"user_id": user.ID.String(),
	}, "User retrieved successfully")

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	r.logger.Info("update_user_start", map[string]interface{}{
		"user_id": id.String(),
	}, "Starting to update user")

	if len(updates) == 0 {
		r.logger.Warning("update_user_no_updates", map[string]interface{}{
			"user_id": id.String(),
		}, "No updates provided")
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("update_user_tx_begin_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Build the update query
	query := "UPDATE users SET updated_at = NOW()"
	values := []interface{}{id}
	paramCount := 2

	for field := range updates {
		query += fmt.Sprintf(", %s = $%d", field, paramCount)
		values = append(values, updates[field])
		paramCount++
	}
	query += " WHERE id = $1"

	r.logger.Debug("update_user_execute", map[string]interface{}{
		"user_id": id.String(),
		"updates": updates,
	}, "Executing update user query")

	// Execute the update
	result, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("update_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.logger.Error("update_user_duplicate_email", map[string]interface{}{
				"user_id": id.String(),
			}, "Duplicate email address")
			return ErrUserDuplicateEmail
		}

		r.logger.Error("update_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		}, "Failed to update user")
		return fmt.Errorf("failed to update user: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("update_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}
		r.logger.Error("update_user_rows_affected_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		}, "Failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("update_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}
		r.logger.Error("update_user_not_found", map[string]interface{}{
			"user_id": id.String(),
		}, "User not found")
		return ErrUserNotFound
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error("update_user_commit_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("update_user_success", map[string]interface{}{
		"user_id": id.String(),
	}, "User updated successfully")

	// Invalidate cache if Redis is healthy
	r.tryCache(func() error {
		return r.cache.Delete(ctx, r.getCacheKey(id))
	})

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("delete_user_start", map[string]interface{}{
		"user_id": id.String(),
	}, "Starting to delete user")

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("delete_user_tx_begin_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	query := "DELETE FROM users WHERE id = $1"

	r.logger.Debug("delete_user_execute", map[string]interface{}{
		"user_id": id.String(),
	}, "Executing delete user query")

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("delete_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}

		r.logger.Error("delete_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		}, "Failed to delete user")
		return fmt.Errorf("failed to delete user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("delete_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}
		r.logger.Error("delete_user_rows_affected_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id.String(),
		}, "Failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("delete_user_rollback_failed", map[string]interface{}{
				"error": rbErr.Error(),
			}, "Failed to rollback transaction")
		}
		r.logger.Error("delete_user_not_found", map[string]interface{}{
			"user_id": id.String(),
		}, "User not found")
		return ErrUserNotFound
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error("delete_user_commit_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	r.logger.Info("delete_user_success", map[string]interface{}{
		"user_id": id.String(),
	}, "User deleted successfully")

	// Invalidate cache if Redis is healthy
	r.tryCache(func() error {
		return r.cache.Delete(ctx, r.getCacheKey(id))
	})

	return nil
}

func (r *UserRepository) GetList(ctx context.Context, limit int, offset int) ([]*account.User, error) {
	r.logger.Info("get_user_list_start", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}, "Starting to get list of users")

	query := `
		SELECT id, email, username, password, type, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	r.logger.Debug("get_user_list_execute", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}, "Executing get user list query")

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("get_user_list_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get list of users")
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer rows.Close()

	var users []*account.User
	for rows.Next() {
		user := &account.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Type,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("get_user_list_scan_failed", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to scan user row")
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("get_user_list_rows_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Error occurred while iterating rows")
		return nil, fmt.Errorf("error occurred while iterating rows: %v", err)
	}

	r.logger.Info("get_user_list_success", map[string]interface{}{
		"count": len(users),
	}, "User list retrieved successfully")

	return users, nil
}

func (r *UserRepository) GetTotalCount(ctx context.Context) (int64, error) {
	r.logger.Info("get_total_count_start", nil, "Starting to get total user count")

	var count int64
	query := "SELECT COUNT(*) FROM users"

	r.logger.Debug("get_total_count_execute", nil, "Executing get total count query")

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		r.logger.Error("get_total_count_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get total user count")
		return 0, fmt.Errorf("failed to get total count: %v", err)
	}

	r.logger.Info("get_total_count_success", map[string]interface{}{
		"count": count,
	}, "Total user count retrieved successfully")

	return count, nil
}
