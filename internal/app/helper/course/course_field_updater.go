package course

import (
	courseDTO "fluencybe/internal/app/dto"
	"fluencybe/internal/app/model/course"
	courseValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"
	"fmt"

	"github.com/lib/pq"
)

type CourseFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewCourseFieldUpdater(logger *logger.PrettyLogger) *CourseFieldUpdater {
	return &CourseFieldUpdater{
		logger: logger,
	}
}

func (u *CourseFieldUpdater) UpdateField(course *course.Course, update courseDTO.UpdateCourseFieldRequest) error {
	switch update.Field {
	case "type":
		return u.updateType(course, update.Value)
	case "title":
		return u.updateTitle(course, update.Value)
	case "overview":
		return u.updateOverview(course, update.Value)
	case "skills":
		return u.updateSkills(course, update.Value)
	case "band":
		return u.updateBand(course, update.Value)
	case "image_urls":
		return u.updateImageURLs(course, update.Value)
	default:
		return fmt.Errorf("invalid field: %s", update.Field)
	}
}

func (u *CourseFieldUpdater) updateType(course *course.Course, value interface{}) error {
	courseType, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type format: expected string")
	}
	if err := courseValidator.ValidateCourseType(courseType); err != nil {
		return err
	}
	course.Type = courseType
	return nil
}

func (u *CourseFieldUpdater) updateTitle(course *course.Course, value interface{}) error {
	title, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid title format: expected string")
	}
	if err := courseValidator.ValidateCourseTitle(title); err != nil {
		return err
	}
	course.Title = title
	return nil
}

func (u *CourseFieldUpdater) updateOverview(course *course.Course, value interface{}) error {
	overview, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid overview format: expected string")
	}
	if err := courseValidator.ValidateCourseOverview(overview); err != nil {
		return err
	}
	course.Overview = overview
	return nil
}

func (u *CourseFieldUpdater) updateSkills(course *course.Course, value interface{}) error {
	skills, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid skills format: expected string array")
	}
	if err := courseValidator.ValidateCourseSkills(skills); err != nil {
		return err
	}
	course.Skills = pq.StringArray(skills)
	return nil
}

func (u *CourseFieldUpdater) updateBand(course *course.Course, value interface{}) error {
	band, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid band format: expected string")
	}
	if err := courseValidator.ValidateCourseBand(band); err != nil {
		return err
	}
	course.Band = band
	return nil
}

func (u *CourseFieldUpdater) updateImageURLs(course *course.Course, value interface{}) error {
	urls, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid image URLs format: expected string array")
	}
	if err := courseValidator.ValidateCourseImageURLs(urls); err != nil {
		return err
	}
	course.ImageURLs = pq.StringArray(urls)
	return nil
}
