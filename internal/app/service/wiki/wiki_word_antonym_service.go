package wiki

import (
	"context"
	"fluencybe/internal/app/helper/wiki"
	wikiModel "fluencybe/internal/app/model/wiki"
	redisClient "fluencybe/internal/app/redis"
	wikiRepo "fluencybe/internal/app/repository/wiki"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type WikiWordAntonymService struct {
	repository     wikiRepo.WikiWordAntonymRepository
	definitionRepo wikiRepo.WikiWordDefinitionRepository
	wordRepo       wikiRepo.WikiWordRepository
	redis          *redisClient.WikiRedis
	updator        *wiki.WikiUpdator
	logger         *logger.PrettyLogger
}

func NewWikiWordAntonymService(
	repository wikiRepo.WikiWordAntonymRepository,
	definitionRepo wikiRepo.WikiWordDefinitionRepository,
	wordRepo wikiRepo.WikiWordRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiWordAntonymService {
	return &WikiWordAntonymService{
		repository:     repository,
		definitionRepo: definitionRepo,
		wordRepo:       wordRepo,
		redis:          redisClient.NewWikiRedis(cache, logger),
		updator:        updator,
		logger:         logger,
	}
}

func (s *WikiWordAntonymService) Create(ctx context.Context, antonym *wikiModel.WikiWordAntonym) error {
	// Check if definition exists
	definition, err := s.definitionRepo.GetByID(ctx, antonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	// Check if antonym word exists
	if _, err := s.wordRepo.GetByID(ctx, antonym.WikiAntonymID); err != nil {
		return fmt.Errorf("antonym word not found: %w", err)
	}

	// Check if antonym relationship already exists
	existingAntonyms, err := s.repository.GetByDefinitionID(ctx, antonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("failed to check existing antonyms: %w", err)
	}
	for _, existing := range existingAntonyms {
		if existing.WikiAntonymID == antonym.WikiAntonymID {
			return fmt.Errorf("antonym relationship already exists")
		}
	}

	if err := s.repository.Create(ctx, antonym); err != nil {
		return fmt.Errorf("failed to create antonym: %w", err)
	}

	// Update word cache since antonyms are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_antonym_service.create.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_antonym_service.create.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordAntonymService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordAntonym, error) {
	antonym, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get antonym: %w", err)
	}
	return antonym, nil
}

func (s *WikiWordAntonymService) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordAntonym, error) {
	antonyms, err := s.repository.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get antonyms: %w", err)
	}
	return antonyms, nil
}

func (s *WikiWordAntonymService) Delete(ctx context.Context, id uuid.UUID) error {
	antonym, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get antonym: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, antonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete antonym: %w", err)
	}

	// Update word cache since antonyms are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_antonym_service.delete.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_antonym_service.delete.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}
