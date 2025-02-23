package speaking

import (
	speakingDTO "fluencybe/internal/app/dto"
	"fluencybe/pkg/logger"
)

type SpeakingQuestionCompletionHelper struct {
	logger *logger.PrettyLogger
}

func NewSpeakingQuestionCompletionHelper(logger *logger.PrettyLogger) *SpeakingQuestionCompletionHelper {
	return &SpeakingQuestionCompletionHelper{
		logger: logger,
	}
}

func (h *SpeakingQuestionCompletionHelper) IsQuestionComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	switch question.Type {
	case "WORD_REPETITION":
		return h.isWordRepetitionComplete(question)
	case "PHRASE_REPETITION":
		return h.isPhraseRepetitionComplete(question)
	case "PARAGRAPH_REPETITION":
		return h.isParagraphRepetitionComplete(question)
	case "OPEN_PARAGRAPH":
		return h.isOpenParagraphComplete(question)
	case "CONVERSATIONAL_REPETITION":
		return h.isConversationalRepetitionComplete(question)
	case "CONVERSATIONAL_OPEN":
		return h.isConversationalOpenComplete(question)
	default:
		return false
	}
}

func (h *SpeakingQuestionCompletionHelper) isWordRepetitionComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	return len(question.WordRepetition) >= 1
}

func (h *SpeakingQuestionCompletionHelper) isPhraseRepetitionComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	return len(question.PhraseRepetition) >= 1
}

func (h *SpeakingQuestionCompletionHelper) isParagraphRepetitionComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	return len(question.ParagraphRepetition) >= 1
}

func (h *SpeakingQuestionCompletionHelper) isOpenParagraphComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	return len(question.OpenParagraph) >= 1
}

func (h *SpeakingQuestionCompletionHelper) isConversationalRepetitionComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	if question.ConversationalRepetition == nil {
		return false
	}
	return len(question.ConversationalRepetitionQAs) >= 2
}

func (h *SpeakingQuestionCompletionHelper) isConversationalOpenComplete(question *speakingDTO.SpeakingQuestionDetail) bool {
	return question.ConversationalOpen != nil
}
