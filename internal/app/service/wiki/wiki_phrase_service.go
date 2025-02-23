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

type WikiPhraseService struct {
	repository wikiRepo.WikiPhraseRepository
	redis      *redisClient.WikiRedis
	updator    *wiki.WikiUpdator
	logger     *logger.PrettyLogger
}

func NewWikiPhraseService(
	repository wikiRepo.WikiPhraseRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiPhraseService {
	return &WikiPhraseService{
		repository: repository,
		redis:      redisClient.NewWikiRedis(cache, logger),
		updator:    updator,
		logger:     logger,
	}
}

func (s *WikiPhraseService) Create(ctx context.Context, phrase *wikiModel.WikiPhrase) error {
	if err := s.repository.Create(ctx, phrase); err != nil {
		return fmt.Errorf("failed to create phrase: %w", err)
	}

	if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
		s.logger.Error("wiki_phrase_service.create.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    phrase.ID,
		}, "Failed to update phrase cache")
	}

	return nil
}

func (s *WikiPhraseService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhrase, error) {
	// Try to get from cache first
	cachedPhrase, err := s.redis.GetCachePhrase(ctx, id)
	if err == nil {
		// Convert DTO to model
		return &wikiModel.WikiPhrase{
			ID:              cachedPhrase.ID,
			Phrase:          cachedPhrase.Phrase,
			Type:            cachedPhrase.Type,
			DifficultyLevel: cachedPhrase.DifficultyLevel,
			CreatedAt:       cachedPhrase.CreatedAt,
			UpdatedAt:       cachedPhrase.UpdatedAt,
		}, nil
	}

	// If not in cache, get from database
	phrase, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get phrase: %w", err)
	}

	// Update cache
	if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
		s.logger.Error("wiki_phrase_service.get.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update phrase cache")
	}

	return phrase, nil
}

func (s *WikiPhraseService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiPhraseRequest) error {
	phrase, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get phrase: %w", err)
	}

	if req.Phrase != nil {
		phrase.Phrase = *req.Phrase
	}
	if req.Type != nil {
		phrase.Type = *req.Type
	}
	if req.DifficultyLevel != nil {
		phrase.DifficultyLevel = *req.DifficultyLevel
	}

	if err := s.repository.Update(ctx, phrase); err != nil {
		return fmt.Errorf("failed to update phrase: %w", err)
	}

	if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
		s.logger.Error("wiki_phrase_service.update.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to update phrase cache")
	}

	return nil
}

func (s *WikiPhraseService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete phrase: %w", err)
	}

	if err := s.redis.RemovePhraseCacheEntries(ctx, id); err != nil {
		s.logger.Error("wiki_phrase_service.delete.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to remove phrase cache entries")
	}

	return nil
}

func (s *WikiPhraseService) Search(ctx context.Context, params dto.WikiSearchParams) ([]dto.WikiPhraseResponse, int64, error) {
	phrases, total, err := s.repository.List(ctx, params.Page, params.PageSize, params.Query, "", nil, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search phrases: %w", err)
	}

	// Convert models to DTOs
	responses := make([]dto.WikiPhraseResponse, len(phrases))
	for i, phrase := range phrases {
		responses[i] = dto.WikiPhraseResponse{
			ID:              phrase.ID,
			Phrase:          phrase.Phrase,
			Type:            phrase.Type,
			DifficultyLevel: phrase.DifficultyLevel,
			CreatedAt:       phrase.CreatedAt,
			UpdatedAt:       phrase.UpdatedAt,
		}
	}

	return responses, total, nil
}

func (s *WikiPhraseService) SearchWithFilter(ctx context.Context, params dto.WikiSearchParams, phraseType string, difficultyLevel *int) ([]dto.WikiPhraseResponse, int64, error) {
	phrases, total, err := s.repository.List(ctx, params.Page, params.PageSize, params.Query, phraseType, difficultyLevel, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search phrases: %w", err)
	}

	// Convert models to DTOs
	responses := make([]dto.WikiPhraseResponse, len(phrases))
	for i, phrase := range phrases {
		responses[i] = dto.WikiPhraseResponse{
			ID:              phrase.ID,
			Phrase:          phrase.Phrase,
			Type:            phrase.Type,
			DifficultyLevel: phrase.DifficultyLevel,
			CreatedAt:       phrase.CreatedAt,
			UpdatedAt:       phrase.UpdatedAt,
		}
	}

	return responses, total, nil
}
