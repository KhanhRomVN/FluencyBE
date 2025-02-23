package grammar

import (
	"context"
	"encoding/json"
	"errors"
	grammarHelper "fluencybe/internal/app/helper/grammar"
	"fluencybe/internal/app/model/grammar"
	searchClient "fluencybe/internal/app/opensearch"
	GrammarRepository "fluencybe/internal/app/repository/grammar"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDuplicateQuestion = errors.New("duplicate fill in the blank question")
)

type GrammarFillInTheBlankQuestionService struct {
	repo                *GrammarRepository.GrammarFillInTheBlankQuestionRepository
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository
	logger              *logger.PrettyLogger
	cache               cache.Cache
	search              *searchClient.GrammarQuestionSearch
	questionUpdator     *grammarHelper.GrammarQuestionUpdator
}

func NewGrammarFillInTheBlankQuestionService(
	repo *GrammarRepository.GrammarFillInTheBlankQuestionRepository,
	grammarQuestionRepo *GrammarRepository.GrammarQuestionRepository,
	logger *logger.PrettyLogger,
	cache cache.Cache,
	search *searchClient.GrammarQuestionSearch,
	questionUpdator *grammarHelper.GrammarQuestionUpdator,
) *GrammarFillInTheBlankQuestionService {
	return &GrammarFillInTheBlankQuestionService{
		repo:                repo,
		grammarQuestionRepo: grammarQuestionRepo,
		logger:              logger,
		cache:               cache,
		search:              search,
		questionUpdator:     questionUpdator,
	}
}

func (s *GrammarFillInTheBlankQuestionService) SetQuestionUpdator(updator *grammarHelper.GrammarQuestionUpdator) {
	s.questionUpdator = updator
}

func (s *GrammarFillInTheBlankQuestionService) validateQuestion(question *grammar.GrammarFillInTheBlankQuestion) error {
	if question == nil {
		return ErrInvalidInput
	}
	if question.GrammarQuestionID == uuid.Nil {
		return errors.New("grammar question ID is required")
	}
	if question.Question == "" {
		return errors.New("question text is required")
	}
	return nil
}

func (s *GrammarFillInTheBlankQuestionService) CreateQuestion(ctx context.Context, question *grammar.GrammarFillInTheBlankQuestion) error {
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
		s.logger.Error("grammar_fill_in_the_blank_question_service.create", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to create question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankQuestionService) GetQuestion(ctx context.Context, id uuid.UUID) (*grammar.GrammarFillInTheBlankQuestion, error) {
	// Get directly from database without caching
	question, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("grammar_fill_in_the_blank_question_service.get", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to get question")
		return nil, err
	}

	return question, nil
}

func (s *GrammarFillInTheBlankQuestionService) UpdateQuestion(ctx context.Context, question *grammar.GrammarFillInTheBlankQuestion) error {
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
		s.logger.Error("grammar_fill_in_the_blank_question_service.update", map[string]interface{}{
			"error": err.Error(),
			"id":    question.ID,
		}, "Failed to update question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankQuestionService) getGrammarQuestionsCacheKey(grammarQuestionID uuid.UUID) string {
	return fmt.Sprintf("grammar_question:%s:fill_in_the_blank_questions", grammarQuestionID.String())
}

func (s *GrammarFillInTheBlankQuestionService) GetQuestionsByGrammarQuestionID(ctx context.Context, grammarQuestionID uuid.UUID) ([]*grammar.GrammarFillInTheBlankQuestion, error) {
	// Try to get from cache first
	cacheKey := s.getGrammarQuestionsCacheKey(grammarQuestionID)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var questions []*grammar.GrammarFillInTheBlankQuestion
		if err := json.Unmarshal([]byte(cachedData), &questions); err == nil {
			s.logger.Info("get_by_grammar_question.cache", map[string]interface{}{
				"id":           grammarQuestionID,
				"cache_status": "hit",
			}, "Cache hit for fill in blank questions")
			return questions, nil
		}
	}

	// If not in cache or error, get from DB
	questions, err := s.repo.GetByGrammarQuestionID(ctx, grammarQuestionID)
	if err != nil {
		s.logger.Error("get_by_grammar_question", map[string]interface{}{
			"error": err.Error(),
			"id":    grammarQuestionID,
		}, "Failed to get questions")
		return nil, err
	}

	s.logger.Debug("get_by_grammar_question.db", map[string]interface{}{
		"id":    grammarQuestionID,
		"count": len(questions),
	}, "Retrieved questions from database")

	return questions, nil
}

func (s *GrammarFillInTheBlankQuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
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
		s.logger.Error("grammar_fill_in_the_blank_question_service.delete", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		}, "Failed to delete question")
		return err
	}

	// Get parent grammar question for cache/search update
	parentQuestion, err := s.grammarQuestionRepo.GetGrammarQuestionByID(ctx, question.GrammarQuestionID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.get_parent", map[string]interface{}{
			"error": err.Error(),
			"id":    question.GrammarQuestionID,
		}, "Failed to get parent question")
		return err
	}

	// Update cache and search
	if err := s.questionUpdator.UpdateCacheAndSearch(ctx, parentQuestion); err != nil {
		tx.Rollback()
		s.logger.Error("grammar_fill_in_the_blank_question_service.update_cache", map[string]interface{}{
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

func (s *GrammarFillInTheBlankQuestionService) GetDB() *gorm.DB {
	return s.repo.GetDB()
}
