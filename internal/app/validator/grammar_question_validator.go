package validator

import (
	"fluencybe/internal/app/model/grammar"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrGrammarQuestionInvalidInput = fmt.Errorf("invalid input")

func ValidateGrammarQuestion(question *grammar.GrammarQuestion) error {
	if question == nil {
		return ErrGrammarQuestionInvalidInput
	}

	if err := ValidateGrammarQuestionType(question.Type); err != nil {
		return err
	}

	if err := ValidateGrammarQuestionTopic(question.Topic); err != nil {
		return err
	}

	if err := ValidateGrammarQuestionInstruction(question.Instruction); err != nil {
		return err
	}

	if err := ValidateGrammarQuestionImageURLs(question.ImageURLs); err != nil {
		return err
	}

	if err := ValidateGrammarQuestionMaxTime(question.MaxTime); err != nil {
		return err
	}

	return nil
}

func ValidateGrammarQuestionType(questionType grammar.GrammarQuestionType) error {
	validTypes := map[grammar.GrammarQuestionType]bool{
		grammar.FillInTheBlank:         true,
		grammar.ChoiceOne:              true,
		grammar.ErrorIdentification:    true,
		grammar.SentenceTransformation: true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrGrammarQuestionInvalidInput)
	}
	return nil
}

func ValidateGrammarQuestionTopic(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: at least one topic is required", ErrGrammarQuestionInvalidInput)
	}
	for _, topic := range topics {
		if len(topic) > constants.MaxTopicLength {
			return fmt.Errorf("%w: topic length exceeds maximum", ErrGrammarQuestionInvalidInput)
		}
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("%w: empty topic after trimming", ErrGrammarQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateGrammarQuestionInstruction(instruction string) error {
	if len(instruction) > constants.MaxInstructionLength {
		return fmt.Errorf("%w: instruction length exceeds maximum", ErrGrammarQuestionInvalidInput)
	}
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("%w: instruction is required", ErrGrammarQuestionInvalidInput)
	}
	return nil
}

func ValidateGrammarQuestionImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrGrammarQuestionInvalidInput)
	}
	for _, u := range urls {
		if !IsGrammarQuestionValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrGrammarQuestionInvalidInput)
		}
	}
	return nil
}

func ValidateGrammarQuestionMaxTime(maxTime int) error {
	if maxTime < constants.MinMaxTime || maxTime > constants.MaxMaxTime {
		return fmt.Errorf("%w: max time must be between %d and %d seconds", ErrGrammarQuestionInvalidInput, constants.MinMaxTime, constants.MaxMaxTime)
	}
	return nil
}

func IsGrammarQuestionValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
