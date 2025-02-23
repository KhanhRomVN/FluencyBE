package reading

import (
	"context"
	"fluencybe/internal/app/model/reading"

	"github.com/google/uuid"
)

type FillInBlankQuestionService interface {
	GetQuestionsByReadingQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingFillInTheBlankQuestion, error)
}

type FillInBlankAnswerService interface {
	GetAnswersByReadingFillInTheBlankQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingFillInTheBlankAnswer, error)
}

type ChoiceOneQuestionService interface {
	GetQuestionsByReadingQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingChoiceOneQuestion, error)
}

type ChoiceOneOptionService interface {
	GetOptionsByQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingChoiceOneOption, error)
}

type ChoiceMultiQuestionService interface {
	GetQuestionsByReadingQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingChoiceMultiQuestion, error)
}

type ChoiceMultiOptionService interface {
	GetOptionsByQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingChoiceMultiOption, error)
}

type TrueFalseService interface {
	GetByReadingQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingTrueFalse, error)
}

type MatchingService interface {
	GetByReadingQuestionID(ctx context.Context, id uuid.UUID) ([]*reading.ReadingMatching, error)
}
