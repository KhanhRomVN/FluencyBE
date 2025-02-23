package listening

import (
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/listening"
	listeningValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"
	"fmt"
)

type ListeningQuestionFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewListeningQuestionFieldUpdater(logger *logger.PrettyLogger) *ListeningQuestionFieldUpdater {
	return &ListeningQuestionFieldUpdater{
		logger: logger,
	}
}

func (u *ListeningQuestionFieldUpdater) UpdateField(question *listening.ListeningQuestion, update listeningDTO.UpdateListeningQuestionFieldRequest) error {
	switch update.Field {
	case "topic":
		return u.updateTopic(question, update.Value)
	case "instruction":
		return u.updateInstruction(question, update.Value)
	case "audio_urls":
		return u.updateAudioURLs(question, update.Value)
	case "image_urls":
		return u.updateImageURLs(question, update.Value)
	case "transcript":
		return u.updateTranscript(question, update.Value)
	case "max_time":
		return u.updateMaxTime(question, update.Value)
	default:
		return fmt.Errorf("invalid field: %s", update.Field)
	}
}

func (u *ListeningQuestionFieldUpdater) updateTopic(question *listening.ListeningQuestion, value interface{}) error {
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
	if err := listeningValidator.ValidateListeningQuestionTopic(topics); err != nil {
		return err
	}
	question.Topic = topics
	return nil
}

func (u *ListeningQuestionFieldUpdater) updateInstruction(question *listening.ListeningQuestion, value interface{}) error {
	instruction, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid instruction format: expected string")
	}
	if err := listeningValidator.ValidateListeningQuestionInstruction(instruction); err != nil {
		return err
	}
	question.Instruction = instruction
	return nil
}

func (u *ListeningQuestionFieldUpdater) updateAudioURLs(question *listening.ListeningQuestion, value interface{}) error {
	// Handle the case where value is a []interface{} from JSON decoding
	if urls, ok := value.([]interface{}); ok {
		strUrls := make([]string, len(urls))
		for i, url := range urls {
			strUrl, ok := url.(string)
			if !ok {
				return fmt.Errorf("invalid audio URL at index %d: expected string", i)
			}
			strUrls[i] = strUrl
		}
		value = strUrls
	}

	urls, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid audio URLs format: expected string array")
	}
	if err := listeningValidator.ValidateListeningQuestionAudioURLs(urls); err != nil {
		return err
	}
	question.AudioURLs = urls
	return nil
}

func (u *ListeningQuestionFieldUpdater) updateImageURLs(question *listening.ListeningQuestion, value interface{}) error {
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
	if err := listeningValidator.ValidateListeningQuestionImageURLs(urls); err != nil {
		return err
	}
	question.ImageURLs = urls
	return nil
}

func (u *ListeningQuestionFieldUpdater) updateTranscript(question *listening.ListeningQuestion, value interface{}) error {
	transcript, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid transcript format: expected string")
	}
	if err := listeningValidator.ValidateListeningQuestionTranscript(transcript); err != nil {
		return err
	}
	question.Transcript = transcript
	return nil
}

func (u *ListeningQuestionFieldUpdater) updateMaxTime(question *listening.ListeningQuestion, value interface{}) error {
	var maxTime int

	switch v := value.(type) {
	case float64:
		maxTime = int(v)
	case int:
		maxTime = v
	default:
		return fmt.Errorf("invalid max time format: expected number")
	}

	if err := listeningValidator.ValidateListeningQuestionMaxTime(maxTime); err != nil {
		return err
	}
	question.MaxTime = maxTime
	return nil
}
