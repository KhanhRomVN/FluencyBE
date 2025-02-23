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

type GrammarErrorIdentificationService struct {
	repo                *GrammarRepository.GrammarErrorIdentificationRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarErrorIdentificationService(
	repo *GrammarRepository.GrammarErrorIdentificationRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarErrorIdentificationService {
	return &GrammarErrorIdentificationService{
		repo:                repo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarErrorIdentificationService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarErrorIdentificationService) validateErrorIdentification(identification *grammar.GrammarErrorIdentification) error {
	if identification == nil {
		return errors.New("invalid input")
	}
	if identification.GrammarQuestionID == uuid.Nil {
		return errors.New("grammar question ID is required")
	}
	if identification.ErrorSentence == "" {
		return errors.New("error sentence is required")
	}
	if identification.ErrorWord == "" {
		return errors.New("error word is required")
	}
	if identification.CorrectWord == "" {
		return errors.New("correct word is required")
	}
	if identification.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *GrammarErrorIdentificationService) Create(ctx context.Context, identification *grammar.GrammarErrorIdentification) error {
	if err := s.validateErrorIdentification(identification); err != nil {
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

	if err := s.repo.Create(ctx, identification); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    identification.ID,
		}, "Failed to create error identification")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, identification.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    identification.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.update_cache", map[string]interface{}{
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

func (s *GrammarErrorIdentificationService) GetByID(ctx context.Context, id uuid.UUID) (*grammar.GrammarErrorIdentification, error) {
	identification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("grammar_error_identification_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get error identification")
		return nil, err
	}
	return identification, nil
}

func (s *GrammarErrorIdentificationService) Update(ctx context.Context, identification *grammar.GrammarErrorIdentification) error {
	if err := s.validateErrorIdentification(identification); err != nil {
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

	if err := s.repo.Update(ctx, identification); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    identification.ID,
		}, "Failed to update error identification")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, identification.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    identification.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.update_cache", map[string]interface{}{
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

func (s *GrammarErrorIdentificationService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get identification first to get parent ID
	identification, err := s.GetByID(ctx, id)
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
		s.logger.Error("grammar_error_identification_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete error identification")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, identification.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    identification.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_error_identification_service.update_cache", map[string]interface{}{
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

func (s *GrammarErrorIdentificationService) GetByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarErrorIdentification, error) {
	identifications, err := s.repo.GetByGrammarQuestionID(ctx, grammarQuestionID)
	if err != nil {
		s.logger.Error("grammar_error_identification_service.get_by_grammar_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get error identifications")
		return nil, err
	}
	return identifications, nil
}
