package course

import (
	"context"
	"fluencybe/internal/app/model/course"

	"github.com/google/uuid"
)

type LessonService interface {
	GetByCourseID(ctx context.Context, id uuid.UUID) ([]*course.Lesson, error)
}

type LessonQuestionService interface {
	GetByLessonID(ctx context.Context, id uuid.UUID) ([]*course.LessonQuestion, error)
}

type CourseBookService interface {
	GetByCourseID(ctx context.Context, id uuid.UUID) (*course.CourseBook, error)
}

type CourseOtherService interface {
	GetByCourseID(ctx context.Context, id uuid.UUID) (*course.CourseOther, error)
}
