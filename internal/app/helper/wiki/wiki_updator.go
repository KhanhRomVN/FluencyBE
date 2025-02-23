package wiki

import (
	"context"
	"fluencybe/internal/app/dto"
	wikiModel "fluencybe/internal/app/model/wiki"
	redisClient "fluencybe/internal/app/redis"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type WikiWordService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWord, error)
}

type WikiWordDefinitionService interface {
	GetByWordID(ctx context.Context, wordID uuid.UUID) ([]wikiModel.WikiWordDefinition, error)
}

type WikiWordDefinitionSampleService interface {
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordDefinitionSample, error)
}

type WikiWordSynonymService interface {
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordSynonym, error)
}

type WikiWordAntonymService interface {
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordAntonym, error)
}

type WikiPhraseService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhrase, error)
}

type WikiPhraseDefinitionService interface {
	GetByPhraseID(ctx context.Context, phraseID uuid.UUID) ([]wikiModel.WikiPhraseDefinition, error)
}

type WikiPhraseDefinitionSampleService interface {
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiPhraseDefinitionSample, error)
}

type WikiUpdator struct {
	logger                        *logger.PrettyLogger
	redis                         *redisClient.WikiRedis
	wordService                   WikiWordService
	wordDefinitionService         WikiWordDefinitionService
	wordDefinitionSampleService   WikiWordDefinitionSampleService
	wordSynonymService            WikiWordSynonymService
	wordAntonymService            WikiWordAntonymService
	phraseService                 WikiPhraseService
	phraseDefinitionService       WikiPhraseDefinitionService
	phraseDefinitionSampleService WikiPhraseDefinitionSampleService
}

func NewWikiUpdator(
	log *logger.PrettyLogger,
	cache cache.Cache,
	wordService WikiWordService,
	wordDefinitionService WikiWordDefinitionService,
	wordDefinitionSampleService WikiWordDefinitionSampleService,
	wordSynonymService WikiWordSynonymService,
	wordAntonymService WikiWordAntonymService,
	phraseService WikiPhraseService,
	phraseDefinitionService WikiPhraseDefinitionService,
	phraseDefinitionSampleService WikiPhraseDefinitionSampleService,
) *WikiUpdator {
	return &WikiUpdator{
		logger:                        log,
		redis:                         redisClient.NewWikiRedis(cache, log),
		wordService:                   wordService,
		wordDefinitionService:         wordDefinitionService,
		wordDefinitionSampleService:   wordDefinitionSampleService,
		wordSynonymService:            wordSynonymService,
		wordAntonymService:            wordAntonymService,
		phraseService:                 phraseService,
		phraseDefinitionService:       phraseDefinitionService,
		phraseDefinitionSampleService: phraseDefinitionSampleService,
	}
}

// Word update methods
func (u *WikiUpdator) UpdateWordCache(ctx context.Context, word *wikiModel.WikiWord) error {
	wordResponse, err := u.buildWordResponse(ctx, word)
	if err != nil {
		return fmt.Errorf("failed to build word response: %w", err)
	}

	if err := u.redis.UpdateCachedWord(ctx, wordResponse); err != nil {
		u.logger.Error("wiki_updator.update_word_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    word.ID,
		}, "Failed to update word cache")
		return err
	}

	return nil
}

func (u *WikiUpdator) buildWordResponse(ctx context.Context, word *wikiModel.WikiWord) (*dto.WikiWordResponse, error) {
	response := &dto.WikiWordResponse{
		ID:            word.ID,
		Word:          word.Word,
		Pronunciation: word.Pronunciation,
		CreatedAt:     word.CreatedAt,
		UpdatedAt:     word.UpdatedAt,
	}

	return response, nil
}

// Phrase update methods
func (u *WikiUpdator) UpdatePhraseCache(ctx context.Context, phrase *wikiModel.WikiPhrase) error {
	phraseResponse, err := u.buildPhraseResponse(ctx, phrase)
	if err != nil {
		return fmt.Errorf("failed to build phrase response: %w", err)
	}

	if err := u.redis.UpdateCachedPhrase(ctx, phraseResponse); err != nil {
		u.logger.Error("wiki_updator.update_phrase_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    phrase.ID,
		}, "Failed to update phrase cache")
		return err
	}

	return nil
}

func (u *WikiUpdator) buildPhraseResponse(ctx context.Context, phrase *wikiModel.WikiPhrase) (*dto.WikiPhraseResponse, error) {
	response := &dto.WikiPhraseResponse{
		ID:              phrase.ID,
		Phrase:          phrase.Phrase,
		Type:            phrase.Type,
		DifficultyLevel: phrase.DifficultyLevel,
		CreatedAt:       phrase.CreatedAt,
		UpdatedAt:       phrase.UpdatedAt,
	}

	return response, nil
}
