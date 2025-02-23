package writing

import (
	"context"
	"fluencybe/internal/app/model/writing"

	"github.com/google/uuid"
)

type SentenceCompletionService interface {
	GetByWritingQuestionID(ctx context.Context, id uuid.UUID) ([]*writing.WritingSentenceCompletion, error)
}

type EssayService interface {
	GetByWritingQuestionID(ctx context.Context, id uuid.UUID) ([]*writing.WritingEssay, error)
}
