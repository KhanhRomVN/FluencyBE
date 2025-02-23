package redis

import (
	"context"
	"encoding/json"
	"fluencybe/internal/app/dto"
	"fluencybe/internal/core/status"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WikiRedis struct {
	cache  cache.Cache
	logger *logger.PrettyLogger
}

func NewWikiRedis(cache cache.Cache, logger *logger.PrettyLogger) *WikiRedis {
	return &WikiRedis{
		cache:  cache,
		logger: logger,
	}
}

func (r *WikiRedis) GetCache() cache.Cache {
	return r.cache
}

// Word caching
func (r *WikiRedis) SetCacheWord(ctx context.Context, word *dto.WikiWordResponse) error {
	if !status.GetRedisStatus() {
		return nil
	}

	cacheKey := r.GenerateCacheKeyForWord(word.ID)
	wordJSON, err := json.Marshal(word)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, cacheKey, string(wordJSON), 24*time.Hour); err != nil {
		r.logger.Error("wiki_redis.cache_word", map[string]interface{}{
			"error": err.Error(),
			"id":    word.ID,
		}, "Failed to cache word")
		return err
	}

	return nil
}

func (r *WikiRedis) GetCacheWord(ctx context.Context, id uuid.UUID) (*dto.WikiWordResponse, error) {
	if !status.GetRedisStatus() {
		return nil, fmt.Errorf("redis disabled")
	}

	cacheKey := r.GenerateCacheKeyForWord(id)
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var word dto.WikiWordResponse
	if err := json.Unmarshal([]byte(data), &word); err != nil {
		return nil, err
	}

	return &word, nil
}

// Phrase caching
func (r *WikiRedis) SetCachePhrase(ctx context.Context, phrase *dto.WikiPhraseResponse) error {
	if !status.GetRedisStatus() {
		return nil
	}

	cacheKey := r.GenerateCacheKeyForPhrase(phrase.ID)
	phraseJSON, err := json.Marshal(phrase)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, cacheKey, string(phraseJSON), 24*time.Hour); err != nil {
		r.logger.Error("wiki_redis.cache_phrase", map[string]interface{}{
			"error": err.Error(),
			"id":    phrase.ID,
		}, "Failed to cache phrase")
		return err
	}

	return nil
}

func (r *WikiRedis) GetCachePhrase(ctx context.Context, id uuid.UUID) (*dto.WikiPhraseResponse, error) {
	if !status.GetRedisStatus() {
		return nil, fmt.Errorf("redis disabled")
	}

	cacheKey := r.GenerateCacheKeyForPhrase(id)
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var phrase dto.WikiPhraseResponse
	if err := json.Unmarshal([]byte(data), &phrase); err != nil {
		return nil, err
	}

	return &phrase, nil
}

// Cache key generators
func (r *WikiRedis) GenerateCacheKeyForWord(id uuid.UUID) string {
	return fmt.Sprintf("wiki:word:%s", id.String())
}

func (r *WikiRedis) GenerateCacheKeyForPhrase(id uuid.UUID) string {
	return fmt.Sprintf("wiki:phrase:%s", id.String())
}

// Cache removal
func (r *WikiRedis) RemoveWordCacheEntries(ctx context.Context, id uuid.UUID) error {
	if !status.GetRedisStatus() {
		return nil
	}

	pattern := fmt.Sprintf("wiki:word:%s:*", id)
	return r.cache.DeletePattern(ctx, pattern)
}

func (r *WikiRedis) RemovePhraseCacheEntries(ctx context.Context, id uuid.UUID) error {
	if !status.GetRedisStatus() {
		return nil
	}

	pattern := fmt.Sprintf("wiki:phrase:%s:*", id)
	return r.cache.DeletePattern(ctx, pattern)
}

// Cache updates
func (r *WikiRedis) UpdateCachedWord(ctx context.Context, word *dto.WikiWordResponse) error {
	if !status.GetRedisStatus() {
		return nil
	}

	newCacheKey := r.GenerateCacheKeyForWord(word.ID)

	wordJSON, err := json.Marshal(word)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, newCacheKey, string(wordJSON), 24*time.Hour); err != nil {
		r.logger.Error("wiki_redis.update_word_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    word.ID,
		}, "Failed to set new cache entry")
		return err
	}

	oldPattern := fmt.Sprintf("wiki:word:%s:*", word.ID)
	keys, err := r.cache.Keys(ctx, oldPattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key != newCacheKey {
			if err := r.cache.Delete(ctx, key); err != nil {
				r.logger.Error("wiki_redis.update_word_cache.delete_old", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				}, "Failed to delete old cache entry")
			}
		}
	}

	return nil
}

func (r *WikiRedis) UpdateCachedPhrase(ctx context.Context, phrase *dto.WikiPhraseResponse) error {
	if !status.GetRedisStatus() {
		return nil
	}

	newCacheKey := r.GenerateCacheKeyForPhrase(phrase.ID)

	phraseJSON, err := json.Marshal(phrase)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, newCacheKey, string(phraseJSON), 24*time.Hour); err != nil {
		r.logger.Error("wiki_redis.update_phrase_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    phrase.ID,
		}, "Failed to set new cache entry")
		return err
	}

	oldPattern := fmt.Sprintf("wiki:phrase:%s:*", phrase.ID)
	keys, err := r.cache.Keys(ctx, oldPattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key != newCacheKey {
			if err := r.cache.Delete(ctx, key); err != nil {
				r.logger.Error("wiki_redis.update_phrase_cache.delete_old", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				}, "Failed to delete old cache entry")
			}
		}
	}

	return nil
}
