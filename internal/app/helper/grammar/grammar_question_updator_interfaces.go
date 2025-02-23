package grammar

import (
	"context"
	"fluencybe/internal/app/model/grammar"

	"github.com/google/uuid"
)

type FillInBlankQuestionService interface {
	GetQuestionsByGrammarQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarFillInTheBlankQuestion, error)
}

type FillInBlankAnswerService interface {
	GetAnswersByGrammarFillInTheBlankQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarFillInTheBlankAnswer, error)
}

type ChoiceOneQuestionService interface {
	GetQuestionsByGrammarQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarChoiceOneQuestion, error)
}

type ChoiceOneOptionService interface {
	GetOptionsByQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarChoiceOneOption, error)
}

type ErrorIdentificationService interface {
	GetByGrammarQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarErrorIdentification, error)
}

type SentenceTransformationService interface {
	GetByGrammarQuestionID(ctx context.Context, id uuid.UUID) ([]*grammar.GrammarSentenceTransformation, error)
}
