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

type WikiPhraseDefinitionService struct {
	repository wikiRepo.WikiPhraseDefinitionRepository
	phraseRepo wikiRepo.WikiPhraseRepository
	redis      *redisClient.WikiRedis
	updator    *wiki.WikiUpdator
	logger     *logger.PrettyLogger
}

func NewWikiPhraseDefinitionService(
	repository wikiRepo.WikiPhraseDefinitionRepository,
	phraseRepo wikiRepo.WikiPhraseRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiPhraseDefinitionService {
	return &WikiPhraseDefinitionService{
		repository: repository,
		phraseRepo: phraseRepo,
		redis:      redisClient.NewWikiRedis(cache, logger),
		updator:    updator,
		logger:     logger,
	}
}

func (s *WikiPhraseDefinitionService) Create(ctx context.Context, definition *wikiModel.WikiPhraseDefinition) error {
	// Check if phrase exists
	if _, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID); err != nil {
		return fmt.Errorf("phrase not found: %w", err)
	}

	// If this is a main definition, unset any existing main definition
	if definition.IsMainDefinition {
		if err := s.unsetMainDefinition(ctx, definition.WikiPhraseID); err != nil {
			return fmt.Errorf("failed to unset main definition: %w", err)
		}
	}

	if err := s.repository.Create(ctx, definition); err != nil {
		return fmt.Errorf("failed to create definition: %w", err)
	}

	// Update phrase cache since definitions are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_service.create.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_service.create.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}

func (s *WikiPhraseDefinitionService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinition, error) {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}
	return definition, nil
}

func (s *WikiPhraseDefinitionService) GetByPhraseID(ctx context.Context, phraseID uuid.UUID) ([]wikiModel.WikiPhraseDefinition, error) {
	definitions, err := s.repository.GetByPhraseID(ctx, phraseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get definitions: %w", err)
	}
	return definitions, nil
}

func (s *WikiPhraseDefinitionService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiPhraseDefinitionRequest) error {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get definition: %w", err)
	}

	if req.Mean != nil {
		definition.Mean = *req.Mean
	}

	if req.IsMainDefinition != nil && *req.IsMainDefinition != definition.IsMainDefinition {
		if *req.IsMainDefinition {
			// Unset any existing main definition before setting this one
			if err := s.unsetMainDefinition(ctx, definition.WikiPhraseID); err != nil {
				return fmt.Errorf("failed to unset main definition: %w", err)
			}
		}
		definition.IsMainDefinition = *req.IsMainDefinition
	}

	if err := s.repository.Update(ctx, definition); err != nil {
		return fmt.Errorf("failed to update definition: %w", err)
	}

	// Update phrase cache since definitions are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_service.update.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_service.update.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}

func (s *WikiPhraseDefinitionService) Delete(ctx context.Context, id uuid.UUID) error {
	definition, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get definition: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete definition: %w", err)
	}

	// Update phrase cache since definitions are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_service.delete.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_service.delete.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}

func (s *WikiPhraseDefinitionService) unsetMainDefinition(ctx context.Context, phraseID uuid.UUID) error {
	mainDefinition, err := s.repository.GetMainDefinitionByPhraseID(ctx, phraseID)
	if err != nil {
		return nil // No main definition exists
	}

	mainDefinition.IsMainDefinition = false
	if err := s.repository.Update(ctx, mainDefinition); err != nil {
		return fmt.Errorf("failed to unset main definition: %w", err)
	}

	return nil
}
