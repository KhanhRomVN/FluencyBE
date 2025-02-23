package redis

import (
	"context"
	"encoding/json"
	"fluencybe/internal/app/dto"
	"fluencybe/internal/core/status"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CourseRedis struct {
	cache  cache.Cache
	logger *logger.PrettyLogger
}

func NewCourseRedis(cache cache.Cache, logger *logger.PrettyLogger) *CourseRedis {
	return &CourseRedis{
		cache:  cache,
		logger: logger,
	}
}

func (r *CourseRedis) GetCache() cache.Cache {
	return r.cache
}

func (r *CourseRedis) SetCacheCourseDetail(ctx context.Context, course *dto.CourseDetail) error {
	if !status.GetRedisStatus() {
		return nil
	}

	cacheKey := r.GenerateCacheKeyForCourse(course.ID)
	courseJSON, err := json.Marshal(course)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, cacheKey, string(courseJSON), 24*time.Hour); err != nil {
		r.logger.Error("course_redis.cache", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to cache course")
		return err
	}

	return nil
}

func (r *CourseRedis) GetCacheCourseDetail(ctx context.Context, id uuid.UUID) (*dto.CourseDetail, error) {
	if !status.GetRedisStatus() {
		return nil, fmt.Errorf("redis disabled")
	}

	cacheKey := r.GenerateCacheKeyForCourse(id)
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var course dto.CourseDetail
	if err := json.Unmarshal([]byte(data), &course); err != nil {
		return nil, err
	}

	return &course, nil
}

func (r *CourseRedis) GenerateCacheKeyForCourse(id uuid.UUID) string {
	return fmt.Sprintf("course:%s", id.String())
}

func (r *CourseRedis) RemoveCourseCacheEntries(ctx context.Context, id uuid.UUID) error {
	if !status.GetRedisStatus() {
		return nil
	}

	pattern := fmt.Sprintf("course:%s:*", id)
	return r.cache.DeletePattern(ctx, pattern)
}

func (r *CourseRedis) UpdateCachedCourse(ctx context.Context, course *dto.CourseDetail) error {
	if !status.GetRedisStatus() {
		return nil
	}

	newCacheKey := r.GenerateCacheKeyForCourse(course.ID)

	courseJSON, err := json.Marshal(course)
	if err != nil {
		return err
	}

	if err := r.cache.Set(ctx, newCacheKey, string(courseJSON), 24*time.Hour); err != nil {
		r.logger.Error("course_redis.update_cache", map[string]interface{}{
			"error": err.Error(),
			"id":    course.ID,
		}, "Failed to set new cache entry")
		return err
	}

	oldPattern := fmt.Sprintf("course:%s:*", course.ID)
	keys, err := r.cache.Keys(ctx, oldPattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key != newCacheKey {
			if err := r.cache.Delete(ctx, key); err != nil {
				r.logger.Error("course_redis.update_cache.delete_old", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				}, "Failed to delete old cache entry")
			}
		}
	}

	return nil
}
