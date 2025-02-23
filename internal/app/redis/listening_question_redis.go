package redis

import (
	"context"
	"encoding/json"
	"fluencybe/internal/core/status"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	listeningDTO "fluencybe/internal/app/dto"

	"github.com/google/uuid"
)

type ListeningQuestionRedis struct {
	cache  cache.Cache
	logger *logger.PrettyLogger
}

func NewListeningQuestionRedis(cache cache.Cache, logger *logger.PrettyLogger) *ListeningQuestionRedis {
	return &ListeningQuestionRedis{
		cache:  cache,
		logger: logger,
	}
}

func (r *ListeningQuestionRedis) GetCache() cache.Cache {
	return r.cache
}

func (r *ListeningQuestionRedis) SetCacheListeningQuestionDetail(ctx context.Context, question *listeningDTO.ListeningQuestionDetail, isComplete bool) error {
	if !status.GetRedisStatus() {
		return nil
	}

	cacheKey := r.GenerateCacheKeyForListeningQuestion(question.ID, question.Version, isComplete)
	questionJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, cacheKey, string(questionJSON), 24*time.Hour); err != nil {
		r.logger.Error("listening_question_redis.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to cache question")
		return err
	}

	return nil
}

func (r *ListeningQuestionRedis) GenerateCacheKeyForListeningQuestion(id uuid.UUID, version int, isComplete bool) string {
	status := "uncomplete"
	if isComplete {
		status = "complete"
	}
	return fmt.Sprintf("listening_question:%s:%s:%d", id.String(), status, version)
}

func (r *ListeningQuestionRedis) RemoveListeningQuestionCacheEntries(ctx context.Context, id uuid.UUID) error {
	if !status.GetRedisStatus() {
		return nil
	}

	pattern := fmt.Sprintf("listening_question:%s:*", id)
	return r.cache.DeletePattern(ctx, pattern)
}

func (r *ListeningQuestionRedis) UpdateCachedListeningQuestion(ctx context.Context, question *listeningDTO.ListeningQuestionDetail, isComplete bool) error {
	if !status.GetRedisStatus() {
		return nil
	}

	newCacheKey := r.GenerateCacheKeyForListeningQuestion(question.ID, question.Version, isComplete)

	questionJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, newCacheKey, string(questionJSON), 24*time.Hour); err != nil {
		r.logger.Error("listening_question_redis.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to set new cache entry")
		return err
	}

	oldPattern := fmt.Sprintf("listening_question:%s:*", question.ID)
	keys, err := r.cache.Keys(ctx, oldPattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key != newCacheKey {
			if err := r.cache.Delete(ctx, key); err != nil {
				r.logger.Error("listening_question_redis.update_cache.delete_old", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				}, "Failed to delete old cache entry")
			}
		}
	}

	return nil
}
