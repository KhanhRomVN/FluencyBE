package validator

import (
	"fluencybe/internal/app/model/course"
	constants "fluencybe/internal/core/constants"
	"fmt"
	"net/url"
	"strings"
)

var ErrCourseInvalidInput = fmt.Errorf("invalid input")

// ValidateCourse validates a course model
func ValidateCourse(course *course.Course) error {
	if course == nil {
		return ErrCourseInvalidInput
	}

	if err := ValidateCourseType(course.Type); err != nil {
		return err
	}

	if err := ValidateCourseTitle(course.Title); err != nil {
		return err
	}

	if err := ValidateCourseOverview(course.Overview); err != nil {
		return err
	}

	if err := ValidateCourseSkills(course.Skills); err != nil {
		return err
	}

	if err := ValidateCourseBand(course.Band); err != nil {
		return err
	}

	if err := ValidateCourseImageURLs(course.ImageURLs); err != nil {
		return err
	}

	return nil
}

// ValidateCourseType validates course type
func ValidateCourseType(courseType string) error {
	validTypes := map[string]bool{
		"BOOK":  true,
		"OTHER": true,
	}

	if !validTypes[courseType] {
		return fmt.Errorf("%w: invalid course type", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateCourseTitle validates course title
func ValidateCourseTitle(title string) error {
	if len(title) > constants.MaxTitleLength {
		return fmt.Errorf("%w: title length exceeds maximum", ErrCourseInvalidInput)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateCourseOverview validates course overview
func ValidateCourseOverview(overview string) error {
	if len(overview) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: overview length exceeds maximum", ErrCourseInvalidInput)
	}
	if strings.TrimSpace(overview) == "" {
		return fmt.Errorf("%w: overview is required", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateCourseSkills validates course skills
func ValidateCourseSkills(skills []string) error {
	if len(skills) == 0 {
		return fmt.Errorf("%w: at least one skill is required", ErrCourseInvalidInput)
	}
	for _, skill := range skills {
		if strings.TrimSpace(skill) == "" {
			return fmt.Errorf("%w: empty skill after trimming", ErrCourseInvalidInput)
		}
	}
	return nil
}

// ValidateCourseBand validates course band
func ValidateCourseBand(band string) error {
	if strings.TrimSpace(band) == "" {
		return fmt.Errorf("%w: band is required", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateCourseImageURLs validates course image URLs
func ValidateCourseImageURLs(urls []string) error {
	if len(urls) > constants.MaxImageURLs {
		return fmt.Errorf("%w: too many image URLs", ErrCourseInvalidInput)
	}
	for _, u := range urls {
		if !IsValidURL(u) {
			return fmt.Errorf("%w: invalid image URL format", ErrCourseInvalidInput)
		}
	}
	return nil
}

// ValidateCourseBook validates a course book model
func ValidateCourseBook(book *course.CourseBook) error {
	if book == nil {
		return ErrCourseInvalidInput
	}

	if err := ValidateCourseBookPublishers(book.Publishers); err != nil {
		return err
	}

	if err := ValidateCourseBookAuthors(book.Authors); err != nil {
		return err
	}

	if err := ValidateCourseBookPublicationYear(book.PublicationYear); err != nil {
		return err
	}

	return nil
}

// ValidateCourseBookPublishers validates course book publishers
func ValidateCourseBookPublishers(publishers []string) error {
	if len(publishers) == 0 {
		return fmt.Errorf("%w: at least one publisher is required", ErrCourseInvalidInput)
	}
	for _, publisher := range publishers {
		if strings.TrimSpace(publisher) == "" {
			return fmt.Errorf("%w: empty publisher after trimming", ErrCourseInvalidInput)
		}
	}
	return nil
}

// ValidateCourseBookAuthors validates course book authors
func ValidateCourseBookAuthors(authors []string) error {
	if len(authors) == 0 {
		return fmt.Errorf("%w: at least one author is required", ErrCourseInvalidInput)
	}
	for _, author := range authors {
		if strings.TrimSpace(author) == "" {
			return fmt.Errorf("%w: empty author after trimming", ErrCourseInvalidInput)
		}
	}
	return nil
}

// ValidateCourseBookPublicationYear validates course book publication year
func ValidateCourseBookPublicationYear(year int) error {
	if year < 1900 {
		return fmt.Errorf("%w: publication year must be 1900 or later", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateLesson validates a lesson model
func ValidateLesson(lesson *course.Lesson) error {
	if lesson == nil {
		return ErrCourseInvalidInput
	}

	if err := ValidateLessonSequence(lesson.Sequence); err != nil {
		return err
	}

	if err := ValidateLessonTitle(lesson.Title); err != nil {
		return err
	}

	if err := ValidateLessonOverview(lesson.Overview); err != nil {
		return err
	}

	return nil
}

// ValidateLessonSequence validates lesson sequence
func ValidateLessonSequence(sequence int) error {
	if sequence < 1 {
		return fmt.Errorf("%w: sequence must be greater than 0", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateLessonTitle validates lesson title
func ValidateLessonTitle(title string) error {
	if len(title) > constants.MaxTitleLength {
		return fmt.Errorf("%w: title length exceeds maximum", ErrCourseInvalidInput)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateLessonOverview validates lesson overview
func ValidateLessonOverview(overview string) error {
	if len(overview) > constants.MaxExplanationLength {
		return fmt.Errorf("%w: overview length exceeds maximum", ErrCourseInvalidInput)
	}
	if strings.TrimSpace(overview) == "" {
		return fmt.Errorf("%w: overview is required", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateLessonQuestion validates a lesson question model
func ValidateLessonQuestion(question *course.LessonQuestion) error {
	if question == nil {
		return ErrCourseInvalidInput
	}

	if err := ValidateLessonQuestionSequence(question.Sequence); err != nil {
		return err
	}

	if err := ValidateLessonQuestionType(question.QuestionType); err != nil {
		return err
	}

	return nil
}

// ValidateLessonQuestionSequence validates lesson question sequence
func ValidateLessonQuestionSequence(sequence int) error {
	if sequence < 1 {
		return fmt.Errorf("%w: sequence must be greater than 0", ErrCourseInvalidInput)
	}
	return nil
}

// ValidateLessonQuestionType validates lesson question type
func ValidateLessonQuestionType(questionType string) error {
	validTypes := map[string]bool{
		"GRAMMAR":   true,
		"LISTENING": true,
		"READING":   true,
		"SPEAKING":  true,
		"WRITING":   true,
	}

	if !validTypes[questionType] {
		return fmt.Errorf("%w: invalid question type", ErrCourseInvalidInput)
	}
	return nil
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
