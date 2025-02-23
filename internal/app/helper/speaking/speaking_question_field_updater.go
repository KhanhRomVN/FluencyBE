package speaking

import (
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/speaking"
	speakingValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"
	"fmt"
)

type SpeakingQuestionFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewSpeakingQuestionFieldUpdater(logger *logger.PrettyLogger) *SpeakingQuestionFieldUpdater {
	return &SpeakingQuestionFieldUpdater{
		logger: logger,
	}
}

func (u *SpeakingQuestionFieldUpdater) UpdateField(question *speaking.SpeakingQuestion, update speakingDTO.UpdateSpeakingQuestionFieldRequest) error {
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

func (u *SpeakingQuestionFieldUpdater) updateTopic(question *speaking.SpeakingQuestion, value interface{}) error {
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

	topics, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid topic format: expected string array")
	}
	if err := speakingValidator.ValidateSpeakingQuestionTopic(topics); err != nil {
		return err
	}
	question.Topic = topics
	return nil
}

func (u *SpeakingQuestionFieldUpdater) updateInstruction(question *speaking.SpeakingQuestion, value interface{}) error {
	instruction, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid instruction format: expected string")
	}
	if err := speakingValidator.ValidateSpeakingQuestionInstruction(instruction); err != nil {
		return err
	}
	question.Instruction = instruction
	return nil
}

func (u *SpeakingQuestionFieldUpdater) updateImageURLs(question *speaking.SpeakingQuestion, value interface{}) error {
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

	urls, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid image URLs format: expected string array")
	}
	if err := speakingValidator.ValidateSpeakingQuestionImageURLs(urls); err != nil {
		return err
	}
	question.ImageURLs = urls
	return nil
}

func (u *SpeakingQuestionFieldUpdater) updateMaxTime(question *speaking.SpeakingQuestion, value interface{}) error {
	var maxTime int

	switch v := value.(type) {
	case float64:
		maxTime = int(v)
	case int:
		maxTime = v
	default:
		return fmt.Errorf("invalid max time format: expected number")
	}

	if err := speakingValidator.ValidateSpeakingQuestionMaxTime(maxTime); err != nil {
		return err
	}
	question.MaxTime = maxTime
	return nil
}
