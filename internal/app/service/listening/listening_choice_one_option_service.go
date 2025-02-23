package listening

import (
	"context"
	"database/sql"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListeningChoiceOneOptionService struct {
	repo                  *ListeningRepository.ListeningChoiceOneOptionRepository
	questionRepo          *ListeningRepository.ListeningChoiceOneQuestionRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningChoiceOneOptionService(
	repo *ListeningRepository.ListeningChoiceOneOptionRepository,
	questionRepo *ListeningRepository.ListeningChoiceOneQuestionRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningChoiceOneOptionService {
	return &ListeningChoiceOneOptionService{
		repo:                  repo,
		questionRepo:          questionRepo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningChoiceOneOptionService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningChoiceOneOptionService) CreateOption(ctx context.Context, option *listening.ListeningChoiceOneOption) error {
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
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.ListeningChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("listening_choice_one_option_service.create", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("listening_choice_one_option_service.create", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Create(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_one_option_service.create", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceOneOptionService) GetOption(ctx context.Context, id uuid.UUID) (*listening.ListeningChoiceOneOption, error) {
	option, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		s.logger.Error("listening_choice_one_option_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get option")
		return nil, err
	}

	return option, nil
}

func (s *ListeningChoiceOneOptionService) UpdateOption(ctx context.Context, option *listening.ListeningChoiceOneOption) error {
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
		existingCorrectOption, err := s.repo.GetCorrectOption(ctx, option.ListeningChoiceOneQuestionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			s.logger.Error("listening_choice_one_option_service.update", map[string]interface{}{
				"error": err.Error(),
			}, "Failed to check existing correct option")
			return err
		}

		if existingCorrectOption != nil && existingCorrectOption.ID != option.ID {
			// Found an existing correct option, update it to be incorrect
			existingCorrectOption.IsCorrect = false
			if err := s.repo.Update(ctx, existingCorrectOption); err != nil {
				tx.Rollback()
				s.logger.Error("listening_choice_one_option_service.update", map[string]interface{}{
					"error": err.Error(),
				}, "Failed to update existing correct option")
				return err
			}
		}
	}

	if err := s.repo.Update(ctx, option); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_one_option_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    option.ID,
		}, "Failed to update option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceOneOptionService) DeleteOption(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("listening_choice_one_option_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete option")
		return err
	}

	// Get parent question
	parentQuestion, err := s.questionRepo.GetByID(ctx, option.ListeningChoiceOneQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get parent question: %w", err)
	}

	// Get listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, parentQuestion.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_choice_one_option_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningChoiceOneOptionService) GetOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*listening.ListeningChoiceOneOption, error) {
	// Get directly from DB
	options, err := s.repo.GetByQuestionID(ctx, questionID)
	if err != nil {
		s.logger.Error("listening_choice_one_option_service.get_by_question_id", map[string]interface{}{
			"error":       err.Error(),
			"question_id": questionID,
		}, "Failed to get options by question ID")
		return nil, err
	}
	return options, nil
}
