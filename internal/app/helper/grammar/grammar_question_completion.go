package grammar

import (
	grammarDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
)

type GrammarQuestionCompletionHelper struct {
	logger *logger.PrettyLogger
}

func NewGrammarQuestionCompletionHelper(logger *logger.PrettyLogger) *GrammarQuestionCompletionHelper {
	return &GrammarQuestionCompletionHelper{
		logger: logger,
	}
}

func (h *GrammarQuestionCompletionHelper) IsQuestionComplete(question *grammarDTO.GrammarQuestionDetail) bool {
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		return h.isFillInTheBlankComplete(question)
	case "CHOICE_ONE":
		return h.isChoiceOneComplete(question)
	case "ERROR_IDENTIFICATION":
		return h.isErrorIdentificationComplete(question)
	case "SENTENCE_TRANSFORMATION":
		return h.isSentenceTransformationComplete(question)
	default:
		return false
	}
}

func (h *GrammarQuestionCompletionHelper) isFillInTheBlankComplete(question *grammarDTO.GrammarQuestionDetail) bool {
	if question.FillInTheBlankQuestion == nil {
		return false
	}
	return len(question.FillInTheBlankAnswers) >= 1
}

func (h *GrammarQuestionCompletionHelper) isChoiceOneComplete(question *grammarDTO.GrammarQuestionDetail) bool {
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

func (h *GrammarQuestionCompletionHelper) isErrorIdentificationComplete(question *grammarDTO.GrammarQuestionDetail) bool {
	return question.ErrorIdentification != nil
}

func (h *GrammarQuestionCompletionHelper) isSentenceTransformationComplete(question *grammarDTO.GrammarQuestionDetail) bool {
	return question.SentenceTransformation != nil
}
