package reading

import (
	readingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
)

type ReadingQuestionCompletionHelper struct {
	logger *logger.PrettyLogger
}

func NewReadingQuestionCompletionHelper(logger *logger.PrettyLogger) *ReadingQuestionCompletionHelper {
	return &ReadingQuestionCompletionHelper{
		logger: logger,
	}
}

func (h *ReadingQuestionCompletionHelper) IsQuestionComplete(question *readingDTO.ReadingQuestionDetail) bool {
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		return h.isFillInTheBlankComplete(question)
	case "CHOICE_ONE":
		return h.isChoiceOneComplete(question)
	case "CHOICE_MULTI":
		return h.isChoiceMultiComplete(question)
	case "MATCHING":
		return h.isMatchingComplete(question)
	case "TRUE_FALSE":
		return h.isTrueFalseComplete(question)
	default:
		return false
	}
}

func (h *ReadingQuestionCompletionHelper) isFillInTheBlankComplete(question *readingDTO.ReadingQuestionDetail) bool {
	if question.FillInTheBlankQuestion == nil {
		return false
	}
	return len(question.FillInTheBlankAnswers) >= 2
}

func (h *ReadingQuestionCompletionHelper) isChoiceOneComplete(question *readingDTO.ReadingQuestionDetail) bool {
	if question.ChoiceOneQuestion == nil {
		return false
	}
	if len(question.ChoiceOneOptions) < 2 {
		return false
	}
	hasTrueOption := false
	hasFalseOption := false
	for _, opt := range question.ChoiceOneOptions {
		if opt.IsCorrect {
			hasTrueOption = true
		} else {
			hasFalseOption = true
		}
	}
	return hasTrueOption && hasFalseOption
}

func (h *ReadingQuestionCompletionHelper) isChoiceMultiComplete(question *readingDTO.ReadingQuestionDetail) bool {
	if question.ChoiceMultiQuestion == nil {
		return false
	}
	if len(question.ChoiceMultiOptions) < 3 {
		return false
	}
	trueCount := 0
	falseCount := 0
	for _, opt := range question.ChoiceMultiOptions {
		if opt.IsCorrect {
			trueCount++
		} else {
			falseCount++
		}
	}
	return trueCount >= 2 && falseCount >= 1
}

func (h *ReadingQuestionCompletionHelper) isMatchingComplete(question *readingDTO.ReadingQuestionDetail) bool {
	return question.Matching != nil
}

func (h *ReadingQuestionCompletionHelper) isTrueFalseComplete(question *readingDTO.ReadingQuestionDetail) bool {
	return len(question.TrueFalse) >= 2
}
