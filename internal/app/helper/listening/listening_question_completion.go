package listening

import (
	listeningDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
)

type ListeningQuestionCompletionHelper struct {
	logger *logger.PrettyLogger
}

// đưa data của 1 listening_question detail và kiểm tra tính complete/uncomplete
func NewListeningQuestionCompletionHelper(logger *logger.PrettyLogger) *ListeningQuestionCompletionHelper {
	return &ListeningQuestionCompletionHelper{
		logger: logger,
	}
}

func (h *ListeningQuestionCompletionHelper) IsQuestionComplete(question *listeningDTO.ListeningQuestionDetail) bool {
	switch question.Type {
	case "FILL_IN_THE_BLANK":
		return h.isFillInTheBlankComplete(question)
	case "CHOICE_ONE":
		return h.isChoiceOneComplete(question)
	case "CHOICE_MULTI":
		return h.isChoiceMultiComplete(question)
	case "MAP_LABELLING":
		return h.isMapLabellingComplete(question)
	case "MATCHING":
		return h.isMatchingComplete(question)
	default:
		return false
	}
}

func (h *ListeningQuestionCompletionHelper) isFillInTheBlankComplete(question *listeningDTO.ListeningQuestionDetail) bool {
	if question.FillInTheBlankQuestion == nil {
		return false
	}
	return len(question.FillInTheBlankAnswers) >= 2
}

func (h *ListeningQuestionCompletionHelper) isChoiceOneComplete(question *listeningDTO.ListeningQuestionDetail) bool {
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

func (h *ListeningQuestionCompletionHelper) isChoiceMultiComplete(question *listeningDTO.ListeningQuestionDetail) bool {
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

func (h *ListeningQuestionCompletionHelper) isMapLabellingComplete(question *listeningDTO.ListeningQuestionDetail) bool {
	if question.MapLabelling == nil {
		return false
	}
	return len(question.MapLabelling) >= 2
}

func (h *ListeningQuestionCompletionHelper) isMatchingComplete(question *listeningDTO.ListeningQuestionDetail) bool {
	if question.Matching == nil {
		return false
	}
	return len(question.Matching) >= 2
}
