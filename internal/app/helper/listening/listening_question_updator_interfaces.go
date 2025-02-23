package listening

import (
	"context"
	"fluencybe/internal/app/model/listening"

	"github.com/google/uuid"
)

type FillInBlankQuestionService interface {
	GetQuestionsByListeningQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningFillInTheBlankQuestion, error)
}

type FillInBlankAnswerService interface {
	GetAnswersByListeningFillInTheBlankQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningFillInTheBlankAnswer, error)
}

type ChoiceOneQuestionService interface {
	GetQuestionsByListeningQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningChoiceOneQuestion, error)
}

type ChoiceOneOptionService interface {
	GetOptionsByQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningChoiceOneOption, error)
}

type ChoiceMultiQuestionService interface {
	GetQuestionsByListeningQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningChoiceMultiQuestion, error)
}

type ChoiceMultiOptionService interface {
	GetOptionsByQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningChoiceMultiOption, error)
}

type MapLabellingService interface {
	GetByListeningQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningMapLabelling, error)
}

type MatchingService interface {
	GetByListeningQuestionID(ctx context.Context, id uuid.UUID) ([]*listening.ListeningMatching, error)
}
