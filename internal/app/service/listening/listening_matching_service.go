package listening

import (
	"context"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
)

type ListeningMatchingService struct {
	repo                  *ListeningRepository.ListeningMatchingRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningMatchingService(
	repo *ListeningRepository.ListeningMatchingRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningMatchingService {
	return &ListeningMatchingService{
		repo:                  repo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningMatchingService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningMatchingService) validateMatching(matching *listening.ListeningMatching) error {
	if matching == nil {
		return errors.New("invalid input")
	}
	if matching.ListeningQuestionID == uuid.Nil {
		return errors.New("listening question ID is required")
	}
	if matching.Question == "" {
		return errors.New("question text is required")
	}
	if matching.Answer == "" {
		return errors.New("answer text is required")
	}
	if matching.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ListeningMatchingService) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningMatching, error) {
	matching, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listening_matching_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get matching by ID")
		return nil, err
	}
	return matching, nil
}

func (s *ListeningMatchingService) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningMatching, error) {
	matchings, err := s.repo.GetByListeningQuestionID(ctx, listeningQuestionID)
	if err != nil {
		s.logger.Error("listening_matching_service.get_by_listening_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get matchings")
		return nil, err
	}
	return matchings, nil
}

func (s *ListeningMatchingService) Create(ctx context.Context, matching *listening.ListeningMatching) error {
	if err := s.validateMatching(matching); err != nil {
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

	// Create matching
	if err := s.repo.Create(ctx, matching); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ID,
		}, "Failed to create matching")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, matching.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.update_cache", map[string]interface{}{
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

func (s *ListeningMatchingService) Update(ctx context.Context, matching *listening.ListeningMatching) error {
	if err := s.validateMatching(matching); err != nil {
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

	// Update matching
	if err := s.repo.Update(ctx, matching); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    matching.ID,
		}, "Failed to update matching")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, matching.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.update_cache", map[string]interface{}{
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

func (s *ListeningMatchingService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get matching first to get parent ID
	matching, err := s.GetByID(ctx, id)
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

	// Delete matching
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete matching")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, matching.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_matching_service.update_cache", map[string]interface{}{
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
