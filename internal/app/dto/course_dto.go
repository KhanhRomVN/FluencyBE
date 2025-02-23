package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Course =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

// ! ------------------------------------------------------------------------------
// ! Base Course Types
// ! ------------------------------------------------------------------------------
type CourseResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Overview  string    `json:"overview"`
	Skills    []string  `json:"skills"`
	Band      string    `json:"band"`
	ImageURLs []string  `json:"image_urls"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type CourseDetail struct {
	CourseResponse
	CourseBook  *CourseBookResponse  `json:"course_book,omitempty"`
	CourseOther *CourseOtherResponse `json:"course_other,omitempty"`
	Lessons     []LessonResponse     `json:"lessons,omitempty"`
}

type CreateCourseRequest struct {
	Type      string   `json:"type" validate:"required,oneof=BOOK OTHER"`
	Title     string   `json:"title" validate:"required"`
	Overview  string   `json:"overview" validate:"required"`
	Skills    []string `json:"skills" validate:"required,min=1"`
	Band      string   `json:"band" validate:"required"`
	ImageURLs []string `json:"image_urls"`
}

type UpdateCourseFieldRequest struct {
	Field string      `json:"field" validate:"required,oneof=type title overview skills band image_urls"`
	Value interface{} `json:"value" validate:"required"`
}

type ListCoursePagination struct {
	Courses  []CourseDetail `json:"courses"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type CourseSearchRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=BOOK OTHER"`
	Title    string `form:"title" binding:"omitempty"`
	Skills   string `form:"skills" binding:"omitempty"`
	Band     string `form:"band" binding:"omitempty"`
}

//! ------------------------------------------------------------------------------
//! Course Book Types
//! ------------------------------------------------------------------------------

type CourseBookResponse struct {
	ID              uuid.UUID `json:"id"`
	CourseID        uuid.UUID `json:"course_id"`
	Publishers      []string  `json:"publishers"`
	Authors         []string  `json:"authors"`
	PublicationYear int       `json:"publication_year"`
}

type CreateCourseBookRequest struct {
	CourseID        uuid.UUID `json:"course_id" validate:"required"`
	Publishers      []string  `json:"publishers" validate:"required,min=1"`
	Authors         []string  `json:"authors" validate:"required,min=1"`
	PublicationYear int       `json:"publication_year" validate:"required,min=1900"`
}

type UpdateCourseBookFieldRequest struct {
	CourseBookID uuid.UUID   `json:"course_book_id" validate:"required"`
	Field        string      `json:"field" validate:"required,oneof=publishers authors publication_year"`
	Value        interface{} `json:"value" validate:"required"`
}

//! ------------------------------------------------------------------------------
//! Course Other Types
//! ------------------------------------------------------------------------------

type CourseOtherResponse struct {
	ID       uuid.UUID `json:"id"`
	CourseID uuid.UUID `json:"course_id"`
}

type CreateCourseOtherRequest struct {
	CourseID uuid.UUID `json:"course_id" validate:"required"`
}

//! ------------------------------------------------------------------------------
//! Lesson Types
//! ------------------------------------------------------------------------------

type LessonResponse struct {
	ID        uuid.UUID                `json:"id"`
	CourseID  uuid.UUID                `json:"course_id"`
	Sequence  int                      `json:"sequence"`
	Title     string                   `json:"title"`
	Overview  string                   `json:"overview"`
	Questions []LessonQuestionResponse `json:"questions,omitempty"`
}

type CreateLessonRequest struct {
	CourseID uuid.UUID `json:"course_id" validate:"required"`
	Sequence int       `json:"sequence" validate:"required,min=1"`
	Title    string    `json:"title" validate:"required"`
	Overview string    `json:"overview" validate:"required"`
}

type UpdateLessonFieldRequest struct {
	LessonID uuid.UUID   `json:"lesson_id" validate:"required"`
	Field    string      `json:"field" validate:"required,oneof=sequence title overview"`
	Value    interface{} `json:"value" validate:"required"`
}

//! ------------------------------------------------------------------------------
//! Lesson Question Types
//! ------------------------------------------------------------------------------

type LessonQuestionResponse struct {
	ID           uuid.UUID `json:"id"`
	LessonID     uuid.UUID `json:"lesson_id"`
	Sequence     int       `json:"sequence"`
	QuestionID   uuid.UUID `json:"question_id"`
	QuestionType string    `json:"question_type"`
}

type CreateLessonQuestionRequest struct {
	LessonID     uuid.UUID `json:"lesson_id" validate:"required"`
	Sequence     int       `json:"sequence" validate:"required,min=1"`
	QuestionID   uuid.UUID `json:"question_id" validate:"required"`
	QuestionType string    `json:"question_type" validate:"required,oneof=GRAMMAR LISTENING READING SPEAKING WRITING"`
}

type UpdateLessonQuestionFieldRequest struct {
	LessonQuestionID uuid.UUID   `json:"lesson_question_id" validate:"required"`
	Field            string      `json:"field" validate:"required,oneof=sequence question_id question_type"`
	Value            interface{} `json:"value" validate:"required"`
}

type SwapSequenceRequest struct {
	ID1 uuid.UUID `json:"id1" validate:"required"`
	ID2 uuid.UUID `json:"id2" validate:"required"`
}
