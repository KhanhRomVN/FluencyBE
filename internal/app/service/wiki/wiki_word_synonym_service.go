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

type WikiWordSynonymService struct {
	repository     wikiRepo.WikiWordSynonymRepository
	definitionRepo wikiRepo.WikiWordDefinitionRepository
	wordRepo       wikiRepo.WikiWordRepository
	redis          *redisClient.WikiRedis
	updator        *wiki.WikiUpdator
	logger         *logger.PrettyLogger
}

func NewWikiWordSynonymService(
	repository wikiRepo.WikiWordSynonymRepository,
	definitionRepo wikiRepo.WikiWordDefinitionRepository,
	wordRepo wikiRepo.WikiWordRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiWordSynonymService {
	return &WikiWordSynonymService{
		repository:     repository,
		definitionRepo: definitionRepo,
		wordRepo:       wordRepo,
		redis:          redisClient.NewWikiRedis(cache, logger),
		updator:        updator,
		logger:         logger,
	}
}

func (s *WikiWordSynonymService) Create(ctx context.Context, synonym *wikiModel.WikiWordSynonym) error {
	// Check if definition exists
	definition, err := s.definitionRepo.GetByID(ctx, synonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	// Check if synonym word exists
	if _, err := s.wordRepo.GetByID(ctx, synonym.WikiSynonymID); err != nil {
		return fmt.Errorf("synonym word not found: %w", err)
	}

	// Check if synonym relationship already exists
	existingSynonyms, err := s.repository.GetByDefinitionID(ctx, synonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("failed to check existing synonyms: %w", err)
	}
	for _, existing := range existingSynonyms {
		if existing.WikiSynonymID == synonym.WikiSynonymID {
			return fmt.Errorf("synonym relationship already exists")
		}
	}

	if err := s.repository.Create(ctx, synonym); err != nil {
		return fmt.Errorf("failed to create synonym: %w", err)
	}

	// Update word cache since synonyms are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_synonym_service.create.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_synonym_service.create.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordSynonymService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordSynonym, error) {
	synonym, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get synonym: %w", err)
	}
	return synonym, nil
}

func (s *WikiWordSynonymService) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordSynonym, error) {
	synonyms, err := s.repository.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get synonyms: %w", err)
	}
	return synonyms, nil
}

func (s *WikiWordSynonymService) Delete(ctx context.Context, id uuid.UUID) error {
	synonym, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get synonym: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, synonym.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete synonym: %w", err)
	}

	// Update word cache since synonyms are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_synonym_service.delete.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_synonym_service.delete.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}
