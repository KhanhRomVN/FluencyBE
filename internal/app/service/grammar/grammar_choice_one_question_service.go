package grammar

import (
	"context"
	"database/sql"
	"errors"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	"fluencybe/internal/app/model/grammar"
	GrammarRepository "fluencybe/internal/app/repository/grammar"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GrammarChoiceOneQuestionService struct {
	repo                *GrammarRepository.GrammarChoiceOneQuestionRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarChoiceOneQuestionService(
	repo *GrammarRepository.GrammarChoiceOneQuestionRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarChoiceOneQuestionService {
	return &GrammarChoiceOneQuestionService{
		repo:                repo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarChoiceOneQuestionService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarChoiceOneQuestionService) validateQuestion(question *grammar.GrammarChoiceOneQuestion) error {
	if question == nil {
		return errors.New("invalid input")
	}
	if question.GrammarQuestionID == uuid.Nil {
		return errors.New("grammar question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	if question.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *GrammarChoiceOneQuestionService) CreateQuestion(ctx context.Context, question *grammar.GrammarChoiceOneQuestion) error {
	if err := s.validateQuestion(question); err != nil {
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

	if err := s.repo.Create(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarChoiceOneQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*grammar.GrammarChoiceOneQuestion, error) {
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("grammar_choice_one_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}
	return question, nil
}

func (s *GrammarChoiceOneQuestionService) UpdateQuestion(ctx context.Context, question *grammar.GrammarChoiceOneQuestion) error {
	if err := s.validateQuestion(question); err != nil {
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

	if err := s.repo.Update(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarChoiceOneQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// Get question before deletion to get parent ID
	question, err := s.GetQuestion(ctx, id)
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
		s.logger.Error("grammar_choice_one_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarChoiceOneQuestionService) GetQuestionsByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarChoiceOneQuestion, error) {
	questions, err := s.repo.GetByGrammarQuestionID(ctx, grammarQuestionID)
	if err != nil {
		s.logger.Error("grammar_choice_one_question_service.get_by_grammar_id", map[string]interface{}{
			"error":             err.Error(),
			"grammarQuestionID": grammarQuestionID,
		}, "Failed to get questions by grammar question ID")
		return nil, err
	}
	return questions, nil
}
