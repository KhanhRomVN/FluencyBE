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

type WikiPhraseDefinitionSampleService struct {
	repository     wikiRepo.WikiPhraseDefinitionSampleRepository
	definitionRepo wikiRepo.WikiPhraseDefinitionRepository
	phraseRepo     wikiRepo.WikiPhraseRepository
	redis          *redisClient.WikiRedis
	updator        *wiki.WikiUpdator
	logger         *logger.PrettyLogger
}

func NewWikiPhraseDefinitionSampleService(
	repository wikiRepo.WikiPhraseDefinitionSampleRepository,
	definitionRepo wikiRepo.WikiPhraseDefinitionRepository,
	phraseRepo wikiRepo.WikiPhraseRepository,
	cache cache.Cache,
	updator *wiki.WikiUpdator,
	logger *logger.PrettyLogger,
) *WikiPhraseDefinitionSampleService {
	return &WikiPhraseDefinitionSampleService{
		repository:     repository,
		definitionRepo: definitionRepo,
		phraseRepo:     phraseRepo,
		redis:          redisClient.NewWikiRedis(cache, logger),
		updator:        updator,
		logger:         logger,
	}
}

func (s *WikiPhraseDefinitionSampleService) Create(ctx context.Context, sample *wikiModel.WikiPhraseDefinitionSample) error {
	// Check if definition exists
	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiPhraseDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Create(ctx, sample); err != nil {
		return fmt.Errorf("failed to create sample: %w", err)
	}

	// Update phrase cache since samples are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_sample_service.create.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_sample_service.create.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}

func (s *WikiPhraseDefinitionSampleService) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinitionSample, error) {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sample: %w", err)
	}
	return sample, nil
}

func (s *WikiPhraseDefinitionSampleService) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiPhraseDefinitionSample, error) {
	samples, err := s.repository.GetByDefinitionID(ctx, definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get samples: %w", err)
	}
	return samples, nil
}

func (s *WikiPhraseDefinitionSampleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateWikiPhraseDefinitionSampleRequest) error {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sample: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiPhraseDefinitionID)
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

	// Update phrase cache since samples are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_sample_service.update.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_sample_service.update.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}

func (s *WikiPhraseDefinitionSampleService) Delete(ctx context.Context, id uuid.UUID) error {
	sample, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sample: %w", err)
	}

	definition, err := s.definitionRepo.GetByID(ctx, sample.WikiPhraseDefinitionID)
	if err != nil {
		return fmt.Errorf("definition not found: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sample: %w", err)
	}

	// Update phrase cache since samples are part of phrase details
	phrase, err := s.phraseRepo.GetByID(ctx, definition.WikiPhraseID)
	if err != nil {
		s.logger.Error("wiki_phrase_definition_sample_service.delete.get_phrase", map[string]interface{}{
			"error":     err.Error(),
			"phrase_id": definition.WikiPhraseID,
		}, "Failed to get phrase for cache update")
	} else {
		if err := s.updator.UpdatePhraseCache(ctx, phrase); err != nil {
			s.logger.Error("wiki_phrase_definition_sample_service.delete.cache", map[string]interface{}{
				"error":     err.Error(),
				"phrase_id": definition.WikiPhraseID,
			}, "Failed to update phrase cache")
		}
	}

	return nil
}
