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

type WikiWordDefinitionSampleService struct {
	repository     wikiRepo.WikiWordDefinitionSampleRepository
	definitionRepo wikiRepo.WikiWordDefinitionRepository
	wordRepo       wikiRepo.WikiWordRepository
	redis          *redisClient.WikiRedis
	updator        *wiki.WikiUpdator
	logger         *logger.PrettyLogger
}

func NewWikiWordDefinitionSampleService(
	repository wikiRepo.WikiWordDefinitionSampleRepository,
	definitionRepo wikiRepo.WikiWordDefinitionRepository,
	wordRepo wikiRepo.WikiWordRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiWordDefinitionSampleService {
	return &WikiWordDefinitionSampleService{
		repository:     repository,
		definitionRepo: definitionRepo,
		wordRepo:       wordRepo,
		redis:          redisClient.NewWikiRedis(cache, logger),
		updator:        updator,
		logger:         logger,
	}
}

func (s *WikiWordDefinitionSampleService) Create(ctx context.Context, sample *wikiModel.WikiWordDefinitionSample) error {
	// Check if definition exists
	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Create(ctx, sample); err != nil {
		return fmt.Errorf("failed to create sample: %w", err)
	}

	// Update word cache since samples are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_sample_service.create.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_sample_service.create.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordDefinitionSampleService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinitionSample, error) {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sample: %w", err)
	}
	return sample, nil
}

func (s *WikiWordDefinitionSampleService) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordDefinitionSample, error) {
	samples, err := s.repository.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get samples: %w", err)
	}
	return samples, nil
}

func (s *WikiWordDefinitionSampleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiWordDefinitionSampleRequest) error {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sample: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if req.SampleSentence != nil {
		sample.SampleSentence = *req.SampleSentence
	}
	if req.SampleSentenceMean != nil {
		sample.SampleSentenceMean = *req.SampleSentenceMean
	}

	if err := s.repository.Update(ctx, sample); err != nil {
		return fmt.Errorf("failed to update sample: %w", err)
	}

	// Update word cache since samples are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_sample_service.update.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_sample_service.update.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}

func (s *WikiWordDefinitionSampleService) Delete(ctx context.Context, id uuid.UUID) error {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sample: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiWordDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sample: %w", err)
	}

	// Update word cache since samples are part of word details
	word, err := s.wordRepo.GetByID(ctx, definition.WikiWordID)
	if err != nil {
		s.logger.Error("wiki_word_definition_sample_service.delete.get_word", map[string]interface{}{
			"error":   err.Error(),
			"word_id": definition.WikiWordID,
		}, "Failed to get word for cache update")
	} else {
		if err := s.updator.UpdateWordCache(ctx, word); err != nil {
			s.logger.Error("wiki_word_definition_sample_service.delete.cache", map[string]interface{}{
				"error":   err.Error(),
				"word_id": definition.WikiWordID,
			}, "Failed to update word cache")
		}
	}

	return nil
}
