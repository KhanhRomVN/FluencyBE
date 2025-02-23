package redis

import (
	"context"
	"encoding/json"
	"fluencybe/internal/core/status"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	grammarDTO "fluencybe/internal/app/dto"

	"github.com/google/uuid"
)

type GrammarQuestionRedis struct {
	cache  cache.Cache
	logger *logger.PrettyLogger
}

func NewGrammarQuestionRedis(cache cache.Cache, logger *logger.PrettyLogger) *GrammarQuestionRedis {
	return &GrammarQuestionRedis{
		cache:  cache,
		logger: logger,
	}
}

func (r *GrammarQuestionRedis) GetCache() cache.Cache {
	return r.cache
}

func (r *GrammarQuestionRedis) SetCacheGrammarQuestionDetail(ctx context.Context, question *grammarDTO.GrammarQuestionDetail, isComplete bool) error {
	if !status.GetRedisStatus() {
		return nil
	}

	cacheKey := r.GenerateCacheKeyForGrammarQuestion(question.ID, question.Version, isComplete)
	questionJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, cacheKey, string(questionJSON), 24*time.Hour); err != nil {
		r.logger.Error("grammar_question_redis.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to cache question")
		return err
	}

	return nil
}

func (r *GrammarQuestionRedis) GenerateCacheKeyForGrammarQuestion(id uuid.UUID, version int, isComplete bool) string {
	status := "uncomplete"
	if isComplete {
		status = "complete"
	}
	return fmt.Sprintf("grammar_question:%s:%s:%d", id.String(), status, version)
}

func (r *GrammarQuestionRedis) RemoveGrammarQuestionCacheEntries(ctx context.Context, id uuid.UUID) error {
	if !status.GetRedisStatus() {
		return nil
	}

	pattern := fmt.Sprintf("grammar_question:%s:*", id)
	return r.cache.DeletePattern(ctx, pattern)
}

func (r *GrammarQuestionRedis) UpdateCachedGrammarQuestion(ctx context.Context, question *grammarDTO.GrammarQuestionDetail, isComplete bool) error {
	if !status.GetRedisStatus() {
		return nil
	}

	newCacheKey := r.GenerateCacheKeyForGrammarQuestion(question.ID, question.Version, isComplete)

	questionJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, newCacheKey, string(questionJSON), 24*time.Hour); err != nil {
		r.logger.Error("grammar_question_redis.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to set new cache entry")
		return err
	}

	oldPattern := fmt.Sprintf("grammar_question:%s:*", question.ID)
	keys, err := r.cache.Keys(ctx, oldPattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key != newCacheKey {
			if err := r.cache.Delete(ctx, key); err != nil {
				r.logger.Error("grammar_question_redis.update_cache.delete_old", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				}, "Failed to delete old cache entry")
			}
		}
	}

	return nil
}
