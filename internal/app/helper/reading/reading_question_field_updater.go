package reading

import (
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/reading"
	readingValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"
	"fmt"
)

type ReadingQuestionFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewReadingQuestionFieldUpdater(logger *logger.PrettyLogger) *ReadingQuestionFieldUpdater {
	return &ReadingQuestionFieldUpdater{
		logger: logger,
	}
}

func (u *ReadingQuestionFieldUpdater) UpdateField(question *reading.ReadingQuestion, update readingDTO.UpdateReadingQuestionFieldRequest) error {
	switch update.Field {
	case "topic":
		return u.updateTopic(question, update.Value)
	case "instruction":
		return u.updateInstruction(question, update.Value)
	case "title":
		return u.updateTitle(question, update.Value)
	case "passages":
		return u.updatePassages(question, update.Value)
	case "image_urls":
		return u.updateImageURLs(question, update.Value)
	case "max_time":
		return u.updateMaxTime(question, update.Value)
	default:
		return fmt.Errorf("invalid field: %s", update.Field)
	}
}

func (u *ReadingQuestionFieldUpdater) updateTopic(question *reading.ReadingQuestion, value interface{}) error {
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
	if err := readingValidator.ValidateReadingQuestionTopic(topics); err != nil {
		return err
	}
	question.Topic = topics
	return nil
}

func (u *ReadingQuestionFieldUpdater) updateInstruction(question *reading.ReadingQuestion, value interface{}) error {
	instruction, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid instruction format: expected string")
	}
	if err := readingValidator.ValidateReadingQuestionInstruction(instruction); err != nil {
		return err
	}
	question.Instruction = instruction
	return nil
}

func (u *ReadingQuestionFieldUpdater) updateTitle(question *reading.ReadingQuestion, value interface{}) error {
	title, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid title format: expected string")
	}
	if err := readingValidator.ValidateReadingQuestionTitle(title); err != nil {
		return err
	}
	question.Title = title
	return nil
}

func (u *ReadingQuestionFieldUpdater) updatePassages(question *reading.ReadingQuestion, value interface{}) error {
	// Handle the case where value is a []interface{} from JSON decoding
	if passages, ok := value.([]interface{}); ok {
		strPassages := make([]string, len(passages))
		for i, passage := range passages {
			strPassage, ok := passage.(string)
			if !ok {
				return fmt.Errorf("invalid passage at index %d: expected string", i)
			}
			strPassages[i] = strPassage
		}
		value = strPassages
	}

	passages, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid passages format: expected string array")
	}
	if err := readingValidator.ValidateReadingQuestionPassages(passages); err != nil {
		return err
	}
	question.Passages = passages
	return nil
}

func (u *ReadingQuestionFieldUpdater) updateImageURLs(question *reading.ReadingQuestion, value interface{}) error {
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
	if err := readingValidator.ValidateReadingQuestionImageURLs(urls); err != nil {
		return err
	}
	question.ImageURLs = urls
	return nil
}

func (u *ReadingQuestionFieldUpdater) updateMaxTime(question *reading.ReadingQuestion, value interface{}) error {
	var maxTime int

	switch v := value.(type) {
	case float64:
		maxTime = int(v)
	case int:
		maxTime = v
	default:
		return fmt.Errorf("invalid max time format: expected number")
	}

	if err := readingValidator.ValidateReadingQuestionMaxTime(maxTime); err != nil {
		return err
	}
	question.MaxTime = maxTime
	return nil
}
