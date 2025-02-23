package grammar

import (
	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/grammar"
	grammarValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"
	"fmt"
)

type GrammarQuestionFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewGrammarQuestionFieldUpdater(logger *logger.PrettyLogger) *GrammarQuestionFieldUpdater {
	return &GrammarQuestionFieldUpdater{
		logger: logger,
	}
}

func (u *GrammarQuestionFieldUpdater) UpdateField(question *grammar.GrammarQuestion, update grammarDTO.UpdateGrammarQuestionFieldRequest) error {
	switch update.Field {
	case "topic":
		return u.updateTopic(question, update.Value)
	case "instruction":
		return u.updateInstruction(question, update.Value)
	case "image_urls":
		return u.updateImageURLs(question, update.Value)
	case "max_time":
		return u.updateMaxTime(question, update.Value)
	default:
		return fmt.Errorf("invalid field: %s", update.Field)
	}
}

func (u *GrammarQuestionFieldUpdater) updateTopic(question *grammar.GrammarQuestion, value interface{}) error {
	// Handle the case where value is a []interface{} from JSON decoding
	if topics, ok := value.([]interface{}); ok {
		strTopics := make([]string, len(topics))
		for i, topic := range topics {
			strTopic, ok := topic.(string)
			if !ok {
				return fmt.Errorf("invalid topic at index %d: expected string", i)
			}
			strTopics[i] = strTopic
		}
		value = strTopics
	}

	// Now try to convert to []string
	topics, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid topic format: expected string array")
	}

	if err := grammarValidator.ValidateGrammarQuestionTopic(topics); err != nil {
		return err
	}
	question.Topic = topics
	return nil
}

func (u *GrammarQuestionFieldUpdater) updateInstruction(question *grammar.GrammarQuestion, value interface{}) error {
	instruction, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid instruction format: expected string")
	}
	if err := grammarValidator.ValidateGrammarQuestionInstruction(instruction); err != nil {
		return err
	}
	question.Instruction = instruction
	return nil
}

func (u *GrammarQuestionFieldUpdater) updateImageURLs(question *grammar.GrammarQuestion, value interface{}) error {
	// Handle the case where value is a []interface{} from JSON decoding
	if urls, ok := value.([]interface{}); ok {
		strUrls := make([]string, len(urls))
		for i, url := range urls {
			strUrl, ok := url.(string)
			if !ok {
				return fmt.Errorf("invalid image URL at index %d: expected string", i)
			}
			strUrls[i] = strUrl
		}
		value = strUrls
	}

	// Now try to convert to []string
	urls, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid image URLs format: expected string array")
	}

	if err := grammarValidator.ValidateGrammarQuestionImageURLs(urls); err != nil {
		return err
	}
	question.ImageURLs = urls
	return nil
}

func (u *GrammarQuestionFieldUpdater) updateMaxTime(question *grammar.GrammarQuestion, value interface{}) error {
	var maxTime int

	switch v := value.(type) {
	case float64:
		maxTime = int(v)
	case int:
		maxTime = v
	default:
		return fmt.Errorf("invalid max time format: expected number")
	}

	if err := grammarValidator.ValidateGrammarQuestionMaxTime(maxTime); err != nil {
		return err
	}
	question.MaxTime = maxTime
	return nil
}
