package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiPhraseDefinitionSampleRepository interface {
	Create(ctx context.Context, sample *wikiModel.WikiPhraseDefinitionSample) error
	Update(ctx context.Context, sample *wikiModel.WikiPhraseDefinitionSample) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinitionSample, error)
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiPhraseDefinitionSample, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiPhraseDefinitionSample, int64, error)
}

type wikiPhraseDefinitionSampleRepository struct {
	db *gorm.DB
}

func NewWikiPhraseDefinitionSampleRepository(db *gorm.DB) WikiPhraseDefinitionSampleRepository {
	return &wikiPhraseDefinitionSampleRepository{db: db}
}

func (r *wikiPhraseDefinitionSampleRepository) Create(ctx context.Context, sample *wikiModel.WikiPhraseDefinitionSample) error {
	return r.db.WithContext(ctx).Create(sample).Error
}

func (r *wikiPhraseDefinitionSampleRepository) Update(ctx context.Context, sample *wikiModel.WikiPhraseDefinitionSample) error {
	return r.db.WithContext(ctx).Save(sample).Error
}

func (r *wikiPhraseDefinitionSampleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiPhraseDefinitionSample{}, id).Error
}

func (r *wikiPhraseDefinitionSampleRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinitionSample, error) {
	var sample wikiModel.WikiPhraseDefinitionSample
	err := r.db.WithContext(ctx).First(&sample, id).Error
	if err != nil {
		return nil, err
	}
	return &sample, nil
}

func (r *wikiPhraseDefinitionSampleRepository) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiPhraseDefinitionSample, error) {
	var samples []wikiModel.WikiPhraseDefinitionSample
	err := r.db.WithContext(ctx).Where("wiki_phrase_definition_id = ?", definitionID).Find(&samples).Error
	if err != nil {
		return nil, err
	}
	return samples, nil
}

func (r *wikiPhraseDefinitionSampleRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiPhraseDefinitionSample, int64, error) {
	var samples []wikiModel.WikiPhraseDefinitionSample
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiPhraseDefinitionSample{})

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
