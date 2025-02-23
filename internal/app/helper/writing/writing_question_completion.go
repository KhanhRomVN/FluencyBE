package writing

import (
	writingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
)

type WritingQuestionCompletionHelper struct {
	logger *logger.PrettyLogger
}

func NewWritingQuestionCompletionHelper(logger *logger.PrettyLogger) *WritingQuestionCompletionHelper {
	return &WritingQuestionCompletionHelper{
		logger: logger,
	}
}

func (h *WritingQuestionCompletionHelper) IsQuestionComplete(question *writingDTO.WritingQuestionDetail) bool {
	switch question.Type {
	case "SENTENCE_COMPLETION":
		return h.isSentenceCompletionComplete(question)
	case "ESSAY":
		return h.isEssayComplete(question)
	default:
		return false
	}
}

func (h *WritingQuestionCompletionHelper) isSentenceCompletionComplete(question *writingDTO.WritingQuestionDetail) bool {
	return len(question.SentenceCompletion) >= 1
}

func (h *WritingQuestionCompletionHelper) isEssayComplete(question *writingDTO.WritingQuestionDetail) bool {
	return len(question.Essay) >= 1
}
