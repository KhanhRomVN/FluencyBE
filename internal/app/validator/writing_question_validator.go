package validator

import (
	"fluencybe/internal/app/model/writing"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrWritingQuestionInvalidInput = fmt.Errorf("invalid input")

func ValidateWritingQuestion(question *writing.WritingQuestion) error {
	if question == nil {
		return ErrWritingQuestionInvalidInput
	}

	if err := ValidateWritingQuestionType(question.Type); err != nil {
		return err
	}

	if err := ValidateWritingQuestionTopic(question.Topic); err != nil {
		return err
	}

	if err := ValidateWritingQuestionInstruction(question.Instruction); err != nil {
		return err
	}

	if err := ValidateWritingQuestionImageURLs(question.ImageURLs); err != nil {
		return err
	}

	if err := ValidateWritingQuestionMaxTime(question.MaxTime); err != nil {
		return err
	}

	return nil
}

func ValidateWritingQuestionType(questionType string) error {
	validTypes := map[string]bool{
		"SENTENCE_COMPLETION": true,
		"ESSAY":               true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingQuestionTopic(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: at least one topic is required", ErrWritingQuestionInvalidInput)
	}
	for _, topic := range topics {
		if len(topic) > constants.MaxTopicLength {
			return fmt.Errorf("%w: topic length exceeds maximum", ErrWritingQuestionInvalidInput)
		}
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("%w: empty topic after trimming", ErrWritingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateWritingQuestionInstruction(instruction string) error {
	if len(instruction) > constants.MaxInstructionLength {
		return fmt.Errorf("%w: instruction length exceeds maximum", ErrWritingQuestionInvalidInput)
	}
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("%w: instruction is required", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingQuestionImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrWritingQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsWritingQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrWritingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateWritingQuestionMaxTime(maxTime int) error {
	if maxTime < constants.MinMaxTime || maxTime > constants.MaxMaxTime {
		return fmt.Errorf("%w: max time must be between %d and %d seconds",
			ErrWritingQuestionInvalidInput,
			constants.MinMaxTime,
			constants.MaxMaxTime)
	}
	return nil
}

func IsWritingQuestionValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// Validation functions for sentence completion
func ValidateWritingSentenceCompletion(sentence string) error {
	if len(sentence) > constants.MaxAnswerLength {
		return fmt.Errorf("%w: sentence length exceeds maximum", ErrWritingQuestionInvalidInput)
	}
	if strings.TrimSpace(sentence) == "" {
		return fmt.Errorf("%w: sentence is required", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingPosition(position string) error {
	validPositions := map[string]bool{
		"start": true,
		"end":   true,
	}
	if !validPositions[position] {
		return fmt.Errorf("%w: invalid position", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingRequiredWords(words []string) error {
	if len(words) == 0 {
		return fmt.Errorf("%w: at least one required word is needed", ErrWritingQuestionInvalidInput)
	}
	for _, word := range words {
		if strings.TrimSpace(word) == "" {
			return fmt.Errorf("%w: empty required word", ErrWritingQuestionInvalidInput)
		}
	}
	return nil
}

// Validation functions for essay
func ValidateWritingEssayType(essayType string) error {
	if len(essayType) > constants.MaxTitleLength {
		return fmt.Errorf("%w: essay type length exceeds maximum", ErrWritingQuestionInvalidInput)
	}
	if strings.TrimSpace(essayType) == "" {
		return fmt.Errorf("%w: essay type is required", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingRequiredPoints(points []string) error {
	if len(points) == 0 {
		return fmt.Errorf("%w: at least one required point is needed", ErrWritingQuestionInvalidInput)
	}
	for _, point := range points {
		if strings.TrimSpace(point) == "" {
			return fmt.Errorf("%w: empty required point", ErrWritingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateWritingSampleEssay(essay string) error {
	if len(essay) > constants.MaxPassageLength {
		return fmt.Errorf("%w: sample essay length exceeds maximum", ErrWritingQuestionInvalidInput)
	}
	if strings.TrimSpace(essay) == "" {
		return fmt.Errorf("%w: sample essay is required", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingExplanation(explain string) error {
	if len(explain) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: explanation length exceeds maximum", ErrWritingQuestionInvalidInput)
	}
	if strings.TrimSpace(explain) == "" {
		return fmt.Errorf("%w: explanation is required", ErrWritingQuestionInvalidInput)
	}
	return nil
}

func ValidateWritingWordCount(minWords, maxWords int) error {
	if minWords < 1 {
		return fmt.Errorf("%w: minimum words must be at least 1", ErrWritingQuestionInvalidInput)
	}
	if maxWords < minWords {
		return fmt.Errorf("%w: maximum words must be greater than or equal to minimum words", ErrWritingQuestionInvalidInput)
	}
	return nil
}
