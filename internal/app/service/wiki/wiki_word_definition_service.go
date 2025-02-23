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

type WikiWordDefinitionService struct {
	repository wikiRepo.WikiWordDefinitionRepository
	wordRepo   wikiRepo.WikiWordRepository
	redis      *redisClient.WikiRedis
	updator    *wiki.WikiUpdator
	logger     *logger.PrettyLogger
}

func NewWikiWordDefinitionService(
	repository wikiRepo.WikiWordDefinitionRepository,
	wordRepo wikiRepo.WikiWordRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiWordDefinitionService {
	return &WikiWordDefinitionService{
		repository: repository,
		wordRepo:   wordRepo,
		redis:      redisClient.NewWikiRedis(cache, logger),
		updator:    updator,
		logger:     logger,
	}
}

func (s *WikiWordDefinitionService) Create(ctx context.Context, definition *wikiModel.WikiWordDefinition) error {
	// Check if word exists
	if _, err := s.wordRepo.GetByID(ctx, definition.WikiWordID); err != nil {
		return fmt.Errorf("word not found: %w", err)
	}

	// If this is a main definition, unset any existing main definition
	if definition.IsMainDefinition {
		if err := s.unsetMainDefinition(ctx, definition.WikiWordID); err != nil {
			return fmt.Errorf("failed to unset main definition: %w", err)
		}
	}

	if err := s.repository.Create(ctx, definition); err != nil {
		return fmt.Errorf("failed to create definition: %w", err)
	}

	// Update word cache since definitions are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_service.create.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_service.create.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordDefinitionService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinition, error) {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}
	return definition, nil
}

func (s *WikiWordDefinitionService) GetByWordID(ctx context.Context, wordID uuid.UUID) ([]wikiModel.WikiWordDefinition, error) {
	definitions, err := s.repository.GetByWordID(ctx, wordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get definitions: %w", err)
	}
	return definitions, nil
}

func (s *WikiWordDefinitionService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiWordDefinitionRequest) error {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get definition: %w", err)
	}

	if req.Means != nil {
		definition.Means = req.Means
	}

	if req.IsMainDefinition != nil && *req.IsMainDefinition != definition.IsMainDefinition {
		if *req.IsMainDefinition {
			// Unset any existing main definition before setting this one
			if err := s.unsetMainDefinition(ctx, definition.WikiWordID); err != nil {
				return fmt.Errorf("failed to unset main definition: %w", err)
			}
		}
		definition.IsMainDefinition = *req.IsMainDefinition
	}

	if err := s.repository.Update(ctx, definition); err != nil {
		return fmt.Errorf("failed to update definition: %w", err)
	}

	// Update word cache since definitions are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_service.update.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_service.update.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordDefinitionService) Delete(ctx context.Context, id uuid.UUID) error {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get definition: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete definition: %w", err)
	}

	// Update word cache since definitions are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_service.delete.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_service.delete.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordDefinitionService) unsetMainDefinition(ctx context.Context, wordID uuid.UUID) error {
	mainDefinition, err := s.repository.GetMainDefinitionByWordID(ctx, wordID)
	if err != nil {
		return nil // No main definition exists
	}

	mainDefinition.IsMainDefinition = false
	if err := s.repository.Update(ctx, mainDefinition); err != nil {
		return fmt.Errorf("failed to unset main definition: %w", err)
	}

	return nil
}
