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

type ListeningMapLabellingService struct {
	repo                  *ListeningRepository.ListeningMapLabellingRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningMapLabellingService(
	repo *ListeningRepository.ListeningMapLabellingRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningMapLabellingService {
	return &ListeningMapLabellingService{
		repo:                  repo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningMapLabellingService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningMapLabellingService) validateMapLabelling(qa *listening.ListeningMapLabelling) error {
	if qa == nil {
		return errors.New("invalid input")
	}
	if qa.ListeningQuestionID == uuid.Nil {
		return errors.New("listening question ID is required")
	}
	if qa.Question == "" {
		return errors.New("question text is required")
	}
	if qa.Answer == "" {
		return errors.New("answer text is required")
	}
	if qa.Explain == "" {
		return errors.New("explanation is required")
	}
	return nil
}

func (s *ListeningMapLabellingService) GetByID(ctx context.Context, id uuid.UUID) (*listening.ListeningMapLabelling, error) {
	qa, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listening_map_labelling_service.get_by_id", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get map labelling by ID")
		return nil, err
	}
	return qa, nil
}

func (s *ListeningMapLabellingService) GetByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningMapLabelling, error) {
	qas, err := s.repo.GetByListeningQuestionID(ctx, listeningQuestionID)
	if err != nil {
		s.logger.Error("listening_map_labelling_service.get_by_listening_question_id", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get map labellings")
		return nil, err
	}
	return qas, nil
}

func (s *ListeningMapLabellingService) Create(ctx context.Context, qa *listening.ListeningMapLabelling) error {
	if err := s.validateMapLabelling(qa); err != nil {
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

	// Create QA
	if err := s.repo.Create(ctx, qa); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    qa.ID,
		}, "Failed to create map labelling")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, qa.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.update_cache", map[string]interface{}{
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

func (s *ListeningMapLabellingService) Update(ctx context.Context, qa *listening.ListeningMapLabelling) error {
	if err := s.validateMapLabelling(qa); err != nil {
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

	// Update QA
	if err := s.repo.Update(ctx, qa); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    qa.ID,
		}, "Failed to update map labelling")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, qa.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.update_cache", map[string]interface{}{
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

func (s *ListeningMapLabellingService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get QA first to get parent ID
	qa, err := s.GetByID(ctx, id)
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

	// Delete QA
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete map labelling")
		return err
	}

	// Get parent listening question for cache/search update
	listeningQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, qa.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get listening question: %w", err)
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, listeningQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_map_labelling_service.update_cache", map[string]interface{}{
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
