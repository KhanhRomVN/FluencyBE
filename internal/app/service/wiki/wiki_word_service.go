package wiki

import (
	"context"
	"fluencybe/internal/app/dto"
	"fluencybe/internal/app/helper/wiki"
	wikiModel "fluencybe/internal/app/model/wiki"
	redisClient "fluencybe/internal/app/redis"
	wikiRepo "fluencybe/internal/app/repository/wiki"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type WikiWordService struct {
	repository wikiRepo.WikiWordRepository
	redis      *redisClient.WikiRedis
	updator    *wiki.WikiUpdator
	logger     *logger.PrettyLogger
}

func NewWikiWordService(
	repository wikiRepo.WikiWordRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiWordService {
	return &WikiWordService{
		repository: repository,
		redis:      redisClient.NewWikiRedis(cache, logger),
		updator:    updator,
		logger:     logger,
	}
}

func (s *WikiWordService) Create(ctx context.Context, word *wikiModel.WikiWord) error {
	if err := s.repository.Create(ctx, word); err != nil {
		return fmt.Errorf("failed to create word: %w", err)
	}

	if err := s.updator.UpdateWordCache(ctx, word); err != nil {
		s.logger.Error("wiki_word_service.create.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    word.ID,
		}, "Failed to update word cache")
	}

	return nil
}

func (s *WikiWordService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWord, error) {
	// Try to get from cache first
	cachedWord, err := s.redis.GetCacheWord(ctx, id)
	if err == nil {
		// Convert DTO to model
		return &wikiModel.WikiWord{
			ID:            cachedWord.ID,
			Word:          cachedWord.Word,
			Pronunciation: cachedWord.Pronunciation,
			CreatedAt:     cachedWord.CreatedAt,
			UpdatedAt:     cachedWord.UpdatedAt,
		}, nil
	}

	// If not in cache, get from database
	word, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	// Update cache
	if err := s.updator.UpdateWordCache(ctx, word); err != nil {
		s.logger.Error("wiki_word_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update word cache")
	}

	return word, nil
}

func (s *WikiWordService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiWordRequest) error {
	word, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get word: %w", err)
	}

	if req.Word != nil {
		word.Word = *req.Word
	}
	if req.Pronunciation != nil {
		word.Pronunciation = *req.Pronunciation
	}

	if err := s.repository.Update(ctx, word); err != nil {
		return fmt.Errorf("failed to update word: %w", err)
	}

	if err := s.updator.UpdateWordCache(ctx, word); err != nil {
		s.logger.Error("wiki_word_service.update.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update word cache")
	}

	return nil
}

func (s *WikiWordService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete word: %w", err)
	}

	if err := s.redis.RemoveWordCacheEntries(ctx, id); err != nil {
		s.logger.Error("wiki_word_service.delete.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to remove word cache entries")
	}

	return nil
}

func (s *WikiWordService) Search(ctx context.Context, params dto.WikiSearchParams) ([]dto.WikiWordResponse, int64, error) {
	words, total, err := s.repository.List(ctx, params.Page, params.PageSize, params.Query, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search words: %w", err)
	}

	// Convert models to DTOs
	responses := make([]dto.WikiWordResponse, len(words))
	for i, word := range words {
		responses[i] = dto.WikiWordResponse{
			ID:            word.ID,
			Word:          word.Word,
			Pronunciation: word.Pronunciation,
			CreatedAt:     word.CreatedAt,
			UpdatedAt:     word.UpdatedAt,
		}
	}

	return responses, total, nil
}
