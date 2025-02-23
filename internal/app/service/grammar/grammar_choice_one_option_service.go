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

type GrammarChoiceOneOptionService struct {
	repo                *GrammarRepository.GrammarChoiceOneOptionRepository
	questionRepo        *GrammarRepository.GrammarChoiceOneQuestionRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarChoiceOneOptionService(
	repo *GrammarRepository.GrammarChoiceOneOptionRepository,
	questionRepo *GrammarRepository.GrammarChoiceOneQuestionRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarChoiceOneOptionService {
	return &GrammarChoiceOneOptionService{
		repo:                repo,
		questionRepo:        questionRepo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		cache:               cache,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarChoiceOneOptionService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarChoiceOneOptionService) validateOption(option *grammar.GrammarChoiceOneOption) error {
	if option == nil {
		return errors.New("invalid input")
	}
	if option.GrammarChoiceOneQuestionID == uuid.Nil {
		return errors.New("question ID is required")
	}
	if option.Options == "" {
		return errors.New("option text is required")
	}
	return nil
}

func (s *GrammarChoiceOneOptionService) CreateOption(ctx context.Context, option *grammar.GrammarChoiceOneOption) error {
	if err := s.validateOption(option); err != nil {
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

	if option.IsCorrect {
		// If this option is marked as correct, ensure no other option for this question is marked as correct
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.GrammarChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("grammar_choice_one_option_service.create", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("grammar_choice_one_option_service.create", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Create(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_option_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.GrammarChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get grammar question for cache/search update
	grammarQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, parentQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get grammar question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, grammarQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarChoiceOneOptionService) GetOption(ctx context.Context, id uuid.UUID) (*grammar.GrammarChoiceOneOption, error) {
	option, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("grammar_choice_one_option_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}

	return option, nil
}

func (s *GrammarChoiceOneOptionService) UpdateOption(ctx context.Context, option *grammar.GrammarChoiceOneOption) error {
	if err := s.validateOption(option); err != nil {
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

	if option.IsCorrect {
		// If this option is being marked as correct, ensure no other option is marked as correct
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.GrammarChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("grammar_choice_one_option_service.update", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil && existingCorrectOption.ID != option.ID {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("grammar_choice_one_option_service.update", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Update(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_option_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.GrammarChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get grammar question for cache/search update
	grammarQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, parentQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get grammar question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, grammarQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarChoiceOneOptionService) DeleteOption(ctx context.Context, id uuid.UUID) error {
	// Get the option first to get the question ID for cache invalidation
	option, err := s.GetOption(ctx, id)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sql.ErrNoRows
		}
		s.logger.Error("grammar_choice_one_option_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.GrammarChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get grammar question for cache/search update
	grammarQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, parentQuestion.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get grammar question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, grammarQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *GrammarChoiceOneOptionService) GetOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*grammar.GrammarChoiceOneOption, error) {
	// Get directly from DB
	options, err := s.repo.GetByQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("grammar_choice_one_option_service.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}
	return options, nil
}
