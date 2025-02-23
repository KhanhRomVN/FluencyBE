package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordDefinitionSampleRepository interface {
	Create(ctx context.Context, sample *wikiModel.WikiWordDefinitionSample) error
	Update(ctx context.Context, sample *wikiModel.WikiWordDefinitionSample) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinitionSample, error)
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordDefinitionSample, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordDefinitionSample, int64, error)
}

type wikiWordDefinitionSampleRepository struct {
	db *gorm.DB
}

func NewWikiWordDefinitionSampleRepository(db *gorm.DB) WikiWordDefinitionSampleRepository {
	return &wikiWordDefinitionSampleRepository{db: db}
}

func (r *wikiWordDefinitionSampleRepository) Create(ctx context.Context, sample *wikiModel.WikiWordDefinitionSample) error {
	return r.db.WithContext(ctx).Create(sample).Error
}

func (r *wikiWordDefinitionSampleRepository) Update(ctx context.Context, sample *wikiModel.WikiWordDefinitionSample) error {
	return r.db.WithContext(ctx).Save(sample).Error
}

func (r *wikiWordDefinitionSampleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiWordDefinitionSample{}, id).Error
}

func (r *wikiWordDefinitionSampleRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinitionSample, error) {
	var sample wikiModel.WikiWordDefinitionSample
	err := r.db.WithContext(ctx).First(&sample, id).Error
	if err != nil {
		return nil, err
	}
	return &sample, nil
}

func (r *wikiWordDefinitionSampleRepository) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordDefinitionSample, error) {
	var samples []wikiModel.WikiWordDefinitionSample
	err := r.db.WithContext(ctx).Where("wiki_word_definition_id = ?", definitionID).Find(&samples).Error
	if err != nil {
		return nil, err
	}
	return samples, nil
}

func (r *wikiWordDefinitionSampleRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordDefinitionSample, int64, error) {
	var samples []wikiModel.WikiWordDefinitionSample
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiWordDefinitionSample{})

	// Count total records
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "asc"
		}
		db = db.Order(sortBy + " " + sortOrder)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&samples).Error
	if err != nil {
		return nil, 0, err
	}

	return samples, total, nil
}
