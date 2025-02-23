package account

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fluencybe/internal/app/model/account"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	developerCachePrefix   = "dev:"
	developerCacheDuration = 15 * time.Minute
)

var (
	ErrDeveloperNotFound       = errors.New("developer not found")
	ErrDeveloperInvalidInput   = errors.New("invalid input data")
	ErrDeveloperDuplicateEmail = errors.New("email already exists")
)

type DeveloperRepository struct {
	db     *sql.DB
	logger *logger.PrettyLogger
	cache  cache.Cache
}

func NewDeveloperRepository(db *sql.DB, cache cache.Cache) *DeveloperRepository {
	return &DeveloperRepository{
		db:     db,
		logger: logger.GetGlobalLogger(),
		cache:  cache,
	}
}

func (r *DeveloperRepository) getCacheKey(id interface{}) string {
	return fmt.Sprintf("%s%v", developerCachePrefix, id)
}

func (r *DeveloperRepository) Create(ctx context.Context, dev *account.Developer) error {
	if dev == nil || dev.Email == "" || dev.Password == "" {
		return ErrDeveloperInvalidInput
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.logger.Error("developer_repository.create.begin_transaction", map[string]interface{}{"error": err.Error()}, "Failed to begin transaction")
		return fmt.Errorf("database error: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO developers 
        (id, email, username, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, query,
		dev.ID,
		dev.Email,
		dev.Username,
		dev.Password,
		now,
		now,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return ErrDeveloperDuplicateEmail
			default:
				r.logger.Error("developer_repository.create.exec", map[string]interface{}{
					"error": err.Error(),
					"code":  pqErr.Code,
				}, "Database error")
			}
		}
		return fmt.Errorf("database error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("developer_repository.create.commit", map[string]interface{}{"error": err.Error()}, "Failed to commit transaction")
		return fmt.Errorf("database error: %w", err)
	}

	// Cache the new developer
	if jsonData, err := json.Marshal(dev); err == nil {
		if err := r.cache.Set(ctx, r.getCacheKey(dev.ID), string(jsonData), developerCacheDuration); err != nil {
			r.logger.Warning("developer_repository.create.cache", map[string]interface{}{"error": err.Error()}, "Failed to cache developer")
		}
	}

	return nil
}

func (r *DeveloperRepository) GetByID(ctx context.Context, id uuid.UUID) (*account.Developer, error) {
	// Try to get from cache first
	var dev account.Developer
	cacheKey := r.getCacheKey(id)

	if cachedData, err := r.cache.Get(ctx, cacheKey); err == nil {
		if err := json.Unmarshal([]byte(cachedData), &dev); err == nil {
			return &dev, nil
		}
	}

	query := `SELECT id, email, username, password, created_at, updated_at
        FROM developers WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&dev.ID,
		&dev.Email,
		&dev.Username,
		&dev.Password,
		&dev.CreatedAt,
		&dev.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeveloperNotFound
		}
		r.logger.Error("developer_repository.get_by_id.scan", map[string]interface{}{"error": err.Error()}, "Failed to get developer")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Cache the developer
	if jsonData, err := json.Marshal(&dev); err == nil {
		if err := r.cache.Set(ctx, cacheKey, string(jsonData), developerCacheDuration); err != nil {
			r.logger.Warning("developer_repository.get_by_id.cache", map[string]interface{}{"error": err.Error()}, "Failed to cache developer")
		}
	}

	return &dev, nil
}

func (r *DeveloperRepository) GetByEmail(ctx context.Context, email string) (*account.Developer, error) {
	query := `SELECT id, email, username, password, created_at, updated_at
        FROM developers WHERE email = $1`

	row := r.db.QueryRowContext(ctx, query, email)

	var dev account.Developer
	err := row.Scan(
		&dev.ID,
		&dev.Email,
		&dev.Username,
		&dev.Password,
		&dev.CreatedAt,
		&dev.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeveloperNotFound
		}
		r.logger.Error("developer_repository.get_by_email.scan", map[string]interface{}{"error": err.Error()}, "Failed to get developer")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Cache the developer
	if jsonData, err := json.Marshal(&dev); err == nil {
		if err := r.cache.Set(ctx, r.getCacheKey(dev.ID), string(jsonData), developerCacheDuration); err != nil {
			r.logger.Warning("developer_repository.get_by_email.cache", map[string]interface{}{"error": err.Error()}, "Failed to cache developer")
		}
	}

	return &dev, nil
}

func (r *DeveloperRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	allowedFields := map[string]bool{
		"email":    true,
		"username": true,
		"password": true,
	}

	cleanUpdates := make(map[string]interface{})
	for k, v := range updates {
		if allowedFields[k] {
			cleanUpdates[k] = v
		}
	}

	if len(cleanUpdates) == 0 {
		return ErrDeveloperInvalidInput
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.logger.Error("developer_repository.update.begin_transaction", map[string]interface{}{"error": err.Error()}, "Failed to begin transaction")
		return fmt.Errorf("database error: %w", err)
	}
	defer tx.Rollback()

	query := "UPDATE developers SET "
	params := []interface{}{}
	i := 1

	for field, value := range cleanUpdates {
		query += fmt.Sprintf("%s = $%d, ", field, i)
		params = append(params, value)
		i++
	}

	query += "updated_at = $" + fmt.Sprint(i)
	params = append(params, time.Now().UTC())
	i++

	query += " WHERE id = $" + fmt.Sprint(i)
	params = append(params, id)

	result, err := tx.ExecContext(ctx, query, params...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDeveloperDuplicateEmail
		}
		r.logger.Error("developer_repository.update.exec", map[string]interface{}{"error": err.Error()}, "Failed to update developer")
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("developer_repository.update.rows_affected", map[string]interface{}{"error": err.Error()}, "Failed to get rows affected")
		return fmt.Errorf("database error: %w", err)
	}

	if rowsAffected == 0 {
		return ErrDeveloperNotFound
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("developer_repository.update.commit", map[string]interface{}{"error": err.Error()}, "Failed to commit transaction")
		return fmt.Errorf("database error: %w", err)
	}

	// Invalidate cache
	if err := r.cache.Delete(ctx, r.getCacheKey(id)); err != nil {
		r.logger.Warning("developer_repository.update.cache_invalidate", map[string]interface{}{"error": err.Error()}, "Failed to invalidate cache")
	}

	return nil
}

func (r *DeveloperRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.logger.Error("developer_repository.delete.begin_transaction", map[string]interface{}{"error": err.Error()}, "Failed to begin transaction")
		return fmt.Errorf("database error: %w", err)
	}
	defer tx.Rollback()

	query := "DELETE FROM developers WHERE id = $1"
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("developer_repository.delete.exec", map[string]interface{}{"error": err.Error()}, "Failed to delete developer")
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("developer_repository.delete.rows_affected", map[string]interface{}{"error": err.Error()}, "Failed to get rows affected")
		return fmt.Errorf("database error: %w", err)
	}

	if rowsAffected == 0 {
		return ErrDeveloperNotFound
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("developer_repository.delete.commit", map[string]interface{}{"error": err.Error()}, "Failed to commit transaction")
		return fmt.Errorf("database error: %w", err)
	}

	// Invalidate cache
	if err := r.cache.Delete(ctx, r.getCacheKey(id)); err != nil {
		r.logger.Warning("developer_repository.delete.cache_invalidate", map[string]interface{}{"error": err.Error()}, "Failed to invalidate cache")
	}

	return nil
}

func (r *DeveloperRepository) GetList(ctx context.Context, limit int, offset int) ([]*account.Developer, error) {
	query := `
		SELECT id, email, username, password, created_at, updated_at
		FROM developers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("developer_repository.get_list.query", map[string]interface{}{"error": err.Error()}, "Failed to get developer list")
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var developers []*account.Developer
	for rows.Next() {
		var dev account.Developer
		err := rows.Scan(
			&dev.ID,
			&dev.Email,
			&dev.Username,
			&dev.Password,
			&dev.CreatedAt,
			&dev.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("developer_repository.get_list.scan", map[string]interface{}{"error": err.Error()}, "Failed to scan developer")
			return nil, fmt.Errorf("database error: %w", err)
		}
		developers = append(developers, &dev)

		// Cache each developer
		if jsonData, err := json.Marshal(&dev); err == nil {
			if err := r.cache.Set(ctx, r.getCacheKey(dev.ID), string(jsonData), developerCacheDuration); err != nil {
				r.logger.Warning("developer_repository.get_list.cache", map[string]interface{}{"error": err.Error()}, "Failed to cache developer")
			}
		}
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("developer_repository.get_list.rows_err", map[string]interface{}{"error": err.Error()}, "Error iterating over rows")
		return nil, fmt.Errorf("database error: %w", err)
	}

	return developers, nil
}

func (r *DeveloperRepository) GetTotalCount(ctx context.Context) (int64, error) {
	// Try to get from cache first
	cacheKey := "developer_total_count"
	if cachedData, err := r.cache.Get(ctx, cacheKey); err == nil {
		var count int64
		if err := json.Unmarshal([]byte(cachedData), &count); err == nil {
			return count, nil
		}
	}

	query := "SELECT COUNT(*) FROM developers"

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		r.logger.Error("developer_repository.get_total_count.query", map[string]interface{}{"error": err.Error()}, "Failed to get total count")
		return 0, fmt.Errorf("database error: %w", err)
	}

	// Cache the count
	if jsonData, err := json.Marshal(count); err == nil {
		if err := r.cache.Set(ctx, cacheKey, string(jsonData), developerCacheDuration); err != nil {
			r.logger.Warning("developer_repository.get_total_count.cache", map[string]interface{}{"error": err.Error()}, "Failed to cache total count")
		}
	}

	return count, nil
}
