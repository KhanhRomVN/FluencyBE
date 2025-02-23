package validator

import (
	"fluencybe/internal/app/model/listening"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrListeningQuestionInvalidInput = fmt.Errorf("invalid input")

func ValidateListeningQuestion(question *listening.ListeningQuestion) error {
	if question == nil {
		return ErrListeningQuestionInvalidInput
	}

	if err := ValidateListeningQuestionType(question.Type); err != nil {
		return err
	}

	if err := ValidateListeningQuestionTopic(question.Topic); err != nil {
		return err
	}

	if err := ValidateListeningQuestionInstruction(question.Instruction); err != nil {
		return err
	}

	if err := ValidateListeningQuestionAudioURLs(question.AudioURLs); err != nil {
		return err
	}

	if err := ValidateListeningQuestionImageURLs(question.ImageURLs); err != nil {
		return err
	}

	if err := ValidateListeningQuestionTranscript(question.Transcript); err != nil {
		return err
	}

	if err := ValidateListeningQuestionMaxTime(question.MaxTime); err != nil {
		return err
	}

	return nil
}

func ValidateListeningQuestionType(questionType string) error {
	validTypes := map[string]bool{
		"FILL_IN_THE_BLANK": true,
		"CHOICE_ONE":        true,
		"CHOICE_MULTI":      true,
		"MAP_LABELLING":     true,
		"MATCHING":          true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrListeningQuestionInvalidInput)
	}
	return nil
}

func ValidateListeningQuestionTopic(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: at least one topic is required", ErrListeningQuestionInvalidInput)
	}
	for _, topic := range topics {
		if len(topic) > constants.MaxTopicLength {
			return fmt.Errorf("%w: topic length exceeds maximum", ErrListeningQuestionInvalidInput)
		}
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("%w: empty topic after trimming", ErrListeningQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateListeningQuestionInstruction(instruction string) error {
	if len(instruction) > constants.MaxInstructionLength {
		return fmt.Errorf("%w: instruction length exceeds maximum", ErrListeningQuestionInvalidInput)
	}
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("%w: instruction is required", ErrListeningQuestionInvalidInput)
	}
	return nil
}

func ValidateListeningQuestionAudioURLs(urls []string) error {
	if len(urls) == 0 {
		return fmt.Errorf("%w: at least one audio URL is required", ErrListeningQuestionInvalidInput)
	}
	if len(urls) > constants.MaxAudioURLs {
		return fmt.Errorf("%w: too many audio URLs", ErrListeningQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsListeningQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid audio URL format", ErrListeningQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateListeningQuestionImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrListeningQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsListeningQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrListeningQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateListeningQuestionTranscript(transcript string) error {
	if len(transcript) > constants.MaxTranscriptLength {
		return fmt.Errorf("%w: transcript length exceeds maximum", ErrListeningQuestionInvalidInput)
	}
	if strings.TrimSpace(transcript) == "" {
		return fmt.Errorf("%w: transcript is required", ErrListeningQuestionInvalidInput)
	}
	return nil
}

func ValidateListeningQuestionMaxTime(maxTime int) error {
	if maxTime < constants.MinMaxTime || maxTime > constants.MaxMaxTime {
		return fmt.Errorf("%w: max time must be between %d and %d seconds", ErrListeningQuestionInvalidInput, constants.MinMaxTime, constants.MaxMaxTime)
	}
	return nil
}

func IsListeningQuestionValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
