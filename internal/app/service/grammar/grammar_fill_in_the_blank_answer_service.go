package grammar

import (
	"context"
	"database/sql"
	"errors"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	"fluencybe/internal/app/model/grammar"
	GrammarRepository "fluencybe/internal/app/repository/grammar"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GrammarFillInTheBlankAnswerService struct {
	repo                *GrammarRepository.GrammarFillInTheBlankAnswerRepository
	questionRepo        *GrammarRepository.GrammarFillInTheBlankQuestionRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarFillInTheBlankAnswerService(
	repo *GrammarRepository.GrammarFillInTheBlankAnswerRepository,
	questionRepo *GrammarRepository.GrammarFillInTheBlankQuestionRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarFillInTheBlankAnswerService {
	return &GrammarFillInTheBlankAnswerService{
		repo:                repo,
		questionRepo:        questionRepo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarFillInTheBlankAnswerService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarFillInTheBlankAnswerService) validateAnswer(answer *grammar.GrammarFillInTheBlankAnswer) error {
	if answer == nil {
		return errors.New("invalid input")
	}
	if answer.GrammarFillInTheBlankQuestionID == uuid.Nil {
		return errors.New("question ID is required")
	}
	if answer.Answer == "" {
		return errors.New("answer text is required")
	}
	if answer.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *GrammarFillInTheBlankAnswerService) CreateAnswer(ctx context.Context, answer *grammar.GrammarFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.GrammarFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
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

	// Create answer in database
	if err := s.repo.Create(ctx, answer); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create answer")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, fillInBlankQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.GrammarQuestionID,
		}, "Failed to get parent grammar question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankAnswerService) GetAnswer(ctx context.Context, id uuid.UUID) (*grammar.GrammarFillInTheBlankAnswer, error) {
	answer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("grammar_fill_in_the_blank_answer_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get answer")
		return nil, err
	}
	return answer, nil
}

func (s *GrammarFillInTheBlankAnswerService) UpdateAnswer(ctx context.Context, answer *grammar.GrammarFillInTheBlankAnswer) error {
	if err := s.validateAnswer(answer); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.GrammarFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
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

	// Update answer in database
	if err := s.repo.Update(ctx, answer); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    answer.ID,
		}, "Failed to update answer")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, fillInBlankQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.GrammarQuestionID,
		}, "Failed to get parent grammar question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankAnswerService) DeleteAnswer(ctx context.Context, id uuid.UUID) error {
	// Get answer before deletion to get parent question ID
	answer, err := s.GetAnswer(ctx, id)
	if err != nil {
		return err
	}

	// Get the parent fill in blank question
	fillInBlankQuestion, err := s.questionRepo.GetByID(ctx, answer.GrammarFillInTheBlankQuestionID)
	if err != nil {
		return fmt.Errorf("failed to get parent question: %w", err)
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

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete answer")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, fillInBlankQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    fillInBlankQuestion.GrammarQuestionID,
		}, "Failed to get parent grammar question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_answer_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankAnswerService) GetAnswersByGrammarFillInTheBlankQuestionID(ctx context.Context, questionID uuid.UUID) ([]*grammar.GrammarFillInTheBlankAnswer, error) {
	answers, err := s.repo.GetByGrammarFillInTheBlankQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("grammar_fill_in_the_blank_answer_service.get_by_question", map[string]interface{}{
			"error":      err.Error(),
			"questionID": questionID,
		}, "Failed to get answers")
		return nil, err
	}
	return answers, nil
}
