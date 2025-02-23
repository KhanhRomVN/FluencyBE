package validator

import (
	"fluencybe/internal/app/model/reading"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrReadingQuestionInvalidInput = fmt.Errorf("invalid input")

func ValidateReadingQuestion(question *reading.ReadingQuestion) error {
	if question == nil {
		return ErrReadingQuestionInvalidInput
	}

	if err := ValidateReadingQuestionType(question.Type); err != nil {
		return err
	}

	if err := ValidateReadingQuestionTopic(question.Topic); err != nil {
		return err
	}

	if err := ValidateReadingQuestionInstruction(question.Instruction); err != nil {
		return err
	}

	if err := ValidateReadingQuestionTitle(question.Title); err != nil {
		return err
	}

	if err := ValidateReadingQuestionPassages(question.Passages); err != nil {
		return err
	}

	if err := ValidateReadingQuestionImageURLs(question.ImageURLs); err != nil {
		return err
	}

	if err := ValidateReadingQuestionMaxTime(question.MaxTime); err != nil {
		return err
	}

	return nil
}

func ValidateReadingQuestionType(questionType string) error {
	validTypes := map[string]bool{
		"FILL_IN_THE_BLANK": true,
		"CHOICE_ONE":        true,
		"CHOICE_MULTI":      true,
		"MATCHING":          true,
		"TRUE_FALSE":        true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingQuestionTopic(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: at least one topic is required", ErrReadingQuestionInvalidInput)
	}
	for _, topic := range topics {
		if len(topic) > constants.MaxTopicLength {
			return fmt.Errorf("%w: topic length exceeds maximum", ErrReadingQuestionInvalidInput)
		}
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("%w: empty topic after trimming", ErrReadingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateReadingQuestionInstruction(instruction string) error {
	if len(instruction) > constants.MaxInstructionLength {
		return fmt.Errorf("%w: instruction length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("%w: instruction is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingQuestionTitle(title string) error {
	if len(title) > constants.MaxTitleLength {
		return fmt.Errorf("%w: title length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingQuestionPassages(passages []string) error {
	if len(passages) == 0 {
		return fmt.Errorf("%w: at least one passage is required", ErrReadingQuestionInvalidInput)
	}
	if len(passages) > constants.MaxPassages {
		return fmt.Errorf("%w: too many passages", ErrReadingQuestionInvalidInput)
	}
	for _, passage := range passages {
		if len(passage) > constants.MaxPassageLength {
			return fmt.Errorf("%w: passage length exceeds maximum", ErrReadingQuestionInvalidInput)
		}
		if strings.TrimSpace(passage) == "" {
			return fmt.Errorf("%w: empty passage after trimming", ErrReadingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateReadingQuestionImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrReadingQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsReadingQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrReadingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateReadingQuestionMaxTime(maxTime int) error {
	if maxTime < constants.MinMaxTime || maxTime > constants.MaxMaxTime {
		return fmt.Errorf("%w: max time must be between %d and %d seconds",
			ErrReadingQuestionInvalidInput,
			constants.MinMaxTime,
			constants.MaxMaxTime)
	}
	return nil
}

func IsReadingQuestionValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// Validation functions for sub-types

func ValidateReadingTrueFalseAnswer(answer string) error {
	validAnswers := map[string]bool{
		"TRUE":      true,
		"FALSE":     true,
		"NOT GIVEN": true,
	}
	if !validAnswers[answer] {
		return fmt.Errorf("%w: invalid true/false answer, must be TRUE, FALSE, or NOT GIVEN",
			ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingExplanation(explain string) error {
	if len(explain) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: explanation length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(explain) == "" {
		return fmt.Errorf("%w: explanation is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingQuestionText(question string) error {
	if len(question) > constants.MaxQuestionLength {
		return fmt.Errorf("%w: question length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(question) == "" {
		return fmt.Errorf("%w: question text is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingAnswerText(answer string) error {
	if len(answer) > constants.MaxAnswerLength {
		return fmt.Errorf("%w: answer length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(answer) == "" {
		return fmt.Errorf("%w: answer text is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingOptionText(option string) error {
	if len(option) > constants.MaxOptionsLength {
		return fmt.Errorf("%w: option length exceeds maximum", ErrReadingQuestionInvalidInput)
	}
	if strings.TrimSpace(option) == "" {
		return fmt.Errorf("%w: option text is required", ErrReadingQuestionInvalidInput)
	}
	return nil
}

func ValidateReadingMatchingPair(question, answer string) error {
	if err := ValidateReadingQuestionText(question); err != nil {
		return err
	}
	if err := ValidateReadingAnswerText(answer); err != nil {
		return err
	}
	return nil
}
