package validator

import (
	"fluencybe/internal/app/model/speaking"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrSpeakingQuestionInvalidInput = fmt.Errorf("invalid input")

func ValidateSpeakingQuestion(question *speaking.SpeakingQuestion) error {
	if question == nil {
		return ErrSpeakingQuestionInvalidInput
	}

	if err := ValidateSpeakingQuestionType(question.Type); err != nil {
		return err
	}

	if err := ValidateSpeakingQuestionTopic(question.Topic); err != nil {
		return err
	}

	if err := ValidateSpeakingQuestionInstruction(question.Instruction); err != nil {
		return err
	}

	if err := ValidateSpeakingQuestionImageURLs(question.ImageURLs); err != nil {
		return err
	}

	if err := ValidateSpeakingQuestionMaxTime(question.MaxTime); err != nil {
		return err
	}

	return nil
}

func ValidateSpeakingQuestionType(questionType string) error {
	validTypes := map[string]bool{
		"WORD_REPETITION":           true,
		"PHRASE_REPETITION":         true,
		"PARAGRAPH_REPETITION":      true,
		"OPEN_PARAGRAPH":            true,
		"CONVERSATIONAL_REPETITION": true,
		"CONVERSATIONAL_OPEN":       true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingQuestionTopic(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: at least one topic is required", ErrSpeakingQuestionInvalidInput)
	}
	for _, topic := range topics {
		if len(topic) > constants.MaxTopicLength {
			return fmt.Errorf("%w: topic length exceeds maximum", ErrSpeakingQuestionInvalidInput)
		}
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("%w: empty topic after trimming", ErrSpeakingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateSpeakingQuestionInstruction(instruction string) error {
	if len(instruction) > constants.MaxInstructionLength {
		return fmt.Errorf("%w: instruction length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("%w: instruction is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingQuestionImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrSpeakingQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsSpeakingQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrSpeakingQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateSpeakingQuestionMaxTime(maxTime int) error {
	if maxTime < constants.MinMaxTime || maxTime > constants.MaxMaxTime {
		return fmt.Errorf("%w: max time must be between %d and %d seconds",
			ErrSpeakingQuestionInvalidInput,
			constants.MinMaxTime,
			constants.MaxMaxTime)
	}
	return nil
}

func IsSpeakingQuestionValidURL(rawURL string) bool {
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

func ValidateSpeakingWord(word string) error {
	if len(word) > constants.MaxAnswerLength {
		return fmt.Errorf("%w: word length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(word) == "" {
		return fmt.Errorf("%w: word is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingPhrase(phrase string) error {
	if len(phrase) > constants.MaxAnswerLength {
		return fmt.Errorf("%w: phrase length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(phrase) == "" {
		return fmt.Errorf("%w: phrase is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingParagraph(paragraph string) error {
	if len(paragraph) > constants.MaxPassageLength {
		return fmt.Errorf("%w: paragraph length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(paragraph) == "" {
		return fmt.Errorf("%w: paragraph is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingMeaning(mean string) error {
	if len(mean) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: meaning length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(mean) == "" {
		return fmt.Errorf("%w: meaning is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingTitle(title string) error {
	if len(title) > constants.MaxTitleLength {
		return fmt.Errorf("%w: title length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingOverview(overview string) error {
	if len(overview) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: overview length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(overview) == "" {
		return fmt.Errorf("%w: overview is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingExampleConversation(conversation string) error {
	if len(conversation) > constants.MaxPassageLength {
		return fmt.Errorf("%w: example conversation length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(conversation) == "" {
		return fmt.Errorf("%w: example conversation is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingAnswer(answer string) error {
	if len(answer) > constants.MaxAnswerLength {
		return fmt.Errorf("%w: answer length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(answer) == "" {
		return fmt.Errorf("%w: answer is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}

func ValidateSpeakingExplanation(explain string) error {
	if len(explain) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: explanation length exceeds maximum", ErrSpeakingQuestionInvalidInput)
	}
	if strings.TrimSpace(explain) == "" {
		return fmt.Errorf("%w: explanation is required", ErrSpeakingQuestionInvalidInput)
	}
	return nil
}
