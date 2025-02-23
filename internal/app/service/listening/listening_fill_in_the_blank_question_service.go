package listening

import (
	"context"
	"encoding/json"
	"errors"
	listeningHelper "fluencybe/internal/app/helper/listening"
	"fluencybe/internal/app/model/listening"
	searchClient "fluencybe/internal/app/opensearch"
	ListeningRepository "fluencybe/internal/app/repository/listening"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDuplicateQuestion = errors.New("duplicate fill in the blank question")
)

type ListeningFillInTheBlankQuestionService struct {
	repo                  *ListeningRepository.ListeningFillInTheBlankQuestionRepository
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository
	logger                *logger.PrettyLogger
	cache                 cache.Cache
	search                *searchClient.ListeningQuestionSearch
	questionUpdator       *listeningHelper.ListeningQuestionUpdator
}

func NewListeningFillInTheBlankQuestionService(
	repo *ListeningRepository.ListeningFillInTheBlankQuestionRepository,
	listeningQuestionRepo *ListeningRepository.ListeningQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.ListeningQuestionSearch,
	questionUpdator *listeningHelper.ListeningQuestionUpdator,
) *ListeningFillInTheBlankQuestionService {
	return &ListeningFillInTheBlankQuestionService{
		repo:                  repo,
		listeningQuestionRepo: listeningQuestionRepo,
		logger:                logger,
		cache:                 cache,
		search:                search,
		questionUpdator:       questionUpdator,
	}
}

func (s *ListeningFillInTheBlankQuestionService) SetQuestionUpdator(updator *listeningHelper.ListeningQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *ListeningFillInTheBlankQuestionService) validateQuestion(question *listening.ListeningFillInTheBlankQuestion) error {
	if question == nil {
		return ErrInvalidInput
	}
	if question.ListeningQuestionID == uuid.Nil {
		return errors.New("listening question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	return nil
}

func (s *ListeningFillInTheBlankQuestionService) CreateQuestion(ctx context.Context, question *listening.ListeningFillInTheBlankQuestion) error {
	if err := s.validateQuestion(question); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Start a transaction
	tx := s.repo.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create question in database
	if err := s.repo.Create(ctx, question); err != nil {
		tx.Rollback()
		if errors.Is(err, ErrDuplicateQuestion) {
			return ErrDuplicateQuestion
		}
		s.logger.Error("listening_fill_in_the_blank_question_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to create question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningFillInTheBlankQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*listening.ListeningFillInTheBlankQuestion, error) {
	// Get directly from database without caching
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listening_fill_in_the_blank_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}

	return question, nil
}

func (s *ListeningFillInTheBlankQuestionService) UpdateQuestion(ctx context.Context, question *listening.ListeningFillInTheBlankQuestion) error {
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

	// Update in database
	if err := s.repo.Update(ctx, question); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningFillInTheBlankQuestionService) getListeningQuestionsCacheKey(listeningQuestionID uuid.UUID) string {
	return fmt.Sprintf("listening_question:%s:fill_in_the_blank_questions", listeningQuestionID.String())
}

func (s *ListeningFillInTheBlankQuestionService) GetQuestionsByListeningQuestionID(ctx context.Context, listeningQuestionID uuid.UUID) ([]*listening.ListeningFillInTheBlankQuestion, error) {
	// Try to get from cache first
	cacheKey := s.getListeningQuestionsCacheKey(listeningQuestionID)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var questions []*listening.ListeningFillInTheBlankQuestion
		if err := json.Unmarshal([]byte(cachedData), &questions); err == nil {
			s.logger.Info("get_by_listening_question.cache", map[string]interface{}{
				"id":           listeningQuestionID,
				"cache_status": "hit",
			}, "Cache hit for fill in blank questions")
			return questions, nil
		}
	}

	// If not in cache or error, get from DB
	questions, err := s.repo.GetByListeningQuestionID(ctx, listeningQuestionID)
	if err != nil {
		s.logger.Error("get_by_listening_question", map[string]interface{}{
			"error": err.Error(),
			"id":    listeningQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	s.logger.Debug("get_by_listening_question.db", map[string]interface{}{
		"id":    listeningQuestionID,
		"count": len(questions),
	}, "Retrieved questions from database")

	return questions, nil
}

func (s *ListeningFillInTheBlankQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
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

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent listening question for cache/search update
	parentQuestion, err := s.listeningQuestionRepo.GetListeningQuestionByID(ctx, question.ListeningQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ListeningQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("listening_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    parentQuestion.ID,
		}, "Failed to update cache and search")
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ListeningFillInTheBlankQuestionService) GetDB() *gorm.DB {
	return s.repo.GetDB()
}
