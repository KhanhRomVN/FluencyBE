package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiPhraseDefinitionRepository interface {
	Create(ctx context.Context, definition *wikiModel.WikiPhraseDefinition) error
	Update(ctx context.Context, definition *wikiModel.WikiPhraseDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinition, error)
	GetByPhraseID(ctx context.Context, phraseID uuid.UUID) ([]wikiModel.WikiPhraseDefinition, error)
	GetMainDefinitionByPhraseID(ctx context.Context, phraseID uuid.UUID) (*wikiModel.WikiPhraseDefinition, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiPhraseDefinition, int64, error)
}

type wikiPhraseDefinitionRepository struct {
	db *gorm.DB
}

func NewWikiPhraseDefinitionRepository(db *gorm.DB) WikiPhraseDefinitionRepository {
	return &wikiPhraseDefinitionRepository{db: db}
}

func (r *wikiPhraseDefinitionRepository) Create(ctx context.Context, definition *wikiModel.WikiPhraseDefinition) error {
	return r.db.WithContext(ctx).Create(definition).Error
}

func (r *wikiPhraseDefinitionRepository) Update(ctx context.Context, definition *wikiModel.WikiPhraseDefinition) error {
	return r.db.WithContext(ctx).Save(definition).Error
}

func (r *wikiPhraseDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiPhraseDefinition{}, id).Error
}

func (r *wikiPhraseDefinitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhraseDefinition, error) {
	var definition wikiModel.WikiPhraseDefinition
	err := r.db.WithContext(ctx).First(&definition, id).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

func (r *wikiPhraseDefinitionRepository) GetByPhraseID(ctx context.Context, phraseID uuid.UUID) ([]wikiModel.WikiPhraseDefinition, error) {
	var definitions []wikiModel.WikiPhraseDefinition
	err := r.db.WithContext(ctx).Where("wiki_phrase_id = ?", phraseID).Find(&definitions).Error
	if err != nil {
		return nil, err
	}
	return definitions, nil
}

func (r *wikiPhraseDefinitionRepository) GetMainDefinitionByPhraseID(ctx context.Context, phraseID uuid.UUID) (*wikiModel.WikiPhraseDefinition, error) {
	var definition wikiModel.WikiPhraseDefinition
	err := r.db.WithContext(ctx).
		Where("wiki_phrase_id = ? AND is_main_definition = true", phraseID).
		First(&definition).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

func (r *wikiPhraseDefinitionRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiPhraseDefinition, int64, error) {
	var definitions []wikiModel.WikiPhraseDefinition
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiPhraseDefinition{})

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
	err = db.Offset(offset).Limit(pageSize).Find(&definitions).Error
	if err != nil {
		return nil, 0, err
	}

	return definitions, total, nil
}
