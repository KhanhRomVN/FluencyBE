package speaking

import (
	"context"
	"fluencybe/internal/app/model/speaking"

	"github.com/google/uuid"
)

type WordRepetitionService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingWordRepetition, error)
}

type PhraseRepetitionService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingPhraseRepetition, error)
}

type ParagraphRepetitionService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingParagraphRepetition, error)
}

type OpenParagraphService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingOpenParagraph, error)
}

type ConversationalRepetitionService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingConversationalRepetition, error)
}

type ConversationalRepetitionQAService interface {
	GetBySpeakingConversationalRepetitionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingConversationalRepetitionQA, error)
}

type ConversationalOpenService interface {
	GetBySpeakingQuestionID(ctx context.Context, id uuid.UUID) ([]*speaking.SpeakingConversationalOpen, error)
}
