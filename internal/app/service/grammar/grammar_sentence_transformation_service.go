package grammar

import (
	"context"
	"errors"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	"fluencybe/internal/app/model/grammar"
	GrammarRepository "fluencybe/internal/app/repository/grammar"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type GrammarSentenceTransformationService struct {
	repo                *GrammarRepository.GrammarSentenceTransformationRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarSentenceTransformationService(
	repo *GrammarRepository.GrammarSentenceTransformationRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarSentenceTransformationService {
	return &GrammarSentenceTransformationService{
		repo:                repo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarSentenceTransformationService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarSentenceTransformationService) validateTransformation(transformation *grammar.GrammarSentenceTransformation) error {
	if transformation == nil {
		return errors.New("invalid input")
	}
	if transformation.GrammarQuestionID == uuid.Nil {
		return errors.New("grammar question ID is required")
	}
	if transformation.OriginalSentence == "" {
		return errors.New("original sentence is required")
	}
	if transformation.ExampleCorrectSentence == "" {
		return errors.New("example correct sentence is required")
	}
	if transformation.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *GrammarSentenceTransformationService) Create(ctx context.Context, transformation *grammar.GrammarSentenceTransformation) error {
	if err := s.validateTransformation(transformation); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Create(ctx, transformation); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    transformation.ID,
		}, "Failed to create sentence transformation")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, transformation.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    transformation.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarSentenceTransformationService) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarSentenceTransformation, error) {
	transformation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("grammar_sentence_transformation_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get sentence transformation")
		return nil, err
	}
	return transformation, nil
}

func (s *GrammarSentenceTransformationService) Update(ctx context.Context, transformation *grammar.GrammarSentenceTransformation) error {
	if err := s.validateTransformation(transformation); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Update(ctx, transformation); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    transformation.ID,
		}, "Failed to update sentence transformation")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, transformation.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    transformation.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarSentenceTransformationService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get transformation first to get parent ID
	transformation, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete sentence transformation")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, transformation.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    transformation.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_sentence_transformation_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarSentenceTransformationService) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarSentenceTransformation, error) {
	transformations, err := s.repo.GetByGrammarQuestionID(ctx, grammarQuestionID)
	if err != nil {
		s.logger.Error("grammar_sentence_transformation_service.get_by_grammar_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get sentence transformations")
		return nil, err
	}
	return transformations, nil
}
