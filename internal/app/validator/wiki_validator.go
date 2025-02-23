package validator

import (
	"fmt"
	"strings"
)

// Word validation
func ValidateWordText(word string) error {
	word = strings.TrimSpace(word)
	if word == "" {
		return fmt.Errorf("word cannot be empty")
	}
	if len(word) > 100 {
		return fmt.Errorf("word length cannot exceed 100 characters")
	}
	return nil
}

func ValidatePronunciation(pronunciation string) error {
	pronunciation = strings.TrimSpace(pronunciation)
	if pronunciation == "" {
		return fmt.Errorf("pronunciation cannot be empty")
	}
	if len(pronunciation) > 100 {
		return fmt.Errorf("pronunciation length cannot exceed 100 characters")
	}
	return nil
}

func ValidateWordMeanings(meanings []string) error {
	if len(meanings) == 0 {
		return fmt.Errorf("at least one meaning is required")
	}
	for _, meaning := range meanings {
		meaning = strings.TrimSpace(meaning)
		if meaning == "" {
			return fmt.Errorf("meaning cannot be empty")
		}
		if len(meaning) > 500 {
			return fmt.Errorf("meaning length cannot exceed 500 characters")
		}
	}
	return nil
}

// Sample sentence validation
func ValidateSampleSentence(sentence string) error {
	sentence = strings.TrimSpace(sentence)
	if sentence == "" {
		return fmt.Errorf("sample sentence cannot be empty")
	}
	if len(sentence) > 1000 {
		return fmt.Errorf("sample sentence length cannot exceed 1000 characters")
	}
	return nil
}

func ValidateSampleSentenceMean(meaning string) error {
	meaning = strings.TrimSpace(meaning)
	if meaning == "" {
		return fmt.Errorf("sample sentence meaning cannot be empty")
	}
	if len(meaning) > 1000 {
		return fmt.Errorf("sample sentence meaning length cannot exceed 1000 characters")
	}
	return nil
}

// Phrase validation
func ValidatePhraseText(phrase string) error {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return fmt.Errorf("phrase cannot be empty")
	}
	if len(phrase) > 255 {
		return fmt.Errorf("phrase length cannot exceed 255 characters")
	}
	return nil
}

func ValidatePhraseType(phraseType string) error {
	phraseType = strings.TrimSpace(phraseType)
	if phraseType == "" {
		return fmt.Errorf("phrase type cannot be empty")
	}
	if len(phraseType) > 25 {
		return fmt.Errorf("phrase type length cannot exceed 25 characters")
	}

	validTypes := []string{"idiom", "collocation", "phrasal_verb", "common_phrase"}
	for _, validType := range validTypes {
		if phraseType == validType {
			return nil
		}
	}
	return fmt.Errorf("invalid phrase type: must be one of %v", validTypes)
}

func ValidateDifficultyLevel(level int) error {
	if level < 1 || level > 5 {
		return fmt.Errorf("difficulty level must be between 1 and 5")
	}
	return nil
}

func ValidatePhraseMeaning(meaning string) error {
	meaning = strings.TrimSpace(meaning)
	if meaning == "" {
		return fmt.Errorf("phrase meaning cannot be empty")
	}
	if len(meaning) > 500 {
		return fmt.Errorf("phrase meaning length cannot exceed 500 characters")
	}
	return nil
}
