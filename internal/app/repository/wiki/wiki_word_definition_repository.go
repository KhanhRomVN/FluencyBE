package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordDefinitionRepository interface {
	Create(ctx context.Context, definition *wikiModel.WikiWordDefinition) error
	Update(ctx context.Context, definition *wikiModel.WikiWordDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinition, error)
	GetByWordID(ctx context.Context, wordID uuid.UUID) ([]wikiModel.WikiWordDefinition, error)
	GetMainDefinitionByWordID(ctx context.Context, wordID uuid.UUID) (*wikiModel.WikiWordDefinition, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordDefinition, int64, error)
}

type wikiWordDefinitionRepository struct {
	db *gorm.DB
}

func NewWikiWordDefinitionRepository(db *gorm.DB) WikiWordDefinitionRepository {
	return &wikiWordDefinitionRepository{db: db}
}

func (r *wikiWordDefinitionRepository) Create(ctx context.Context, definition *wikiModel.WikiWordDefinition) error {
	return r.db.WithContext(ctx).Create(definition).Error
}

func (r *wikiWordDefinitionRepository) Update(ctx context.Context, definition *wikiModel.WikiWordDefinition) error {
	return r.db.WithContext(ctx).Save(definition).Error
}

func (r *wikiWordDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiWordDefinition{}, id).Error
}

func (r *wikiWordDefinitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordDefinition, error) {
	var definition wikiModel.WikiWordDefinition
	err := r.db.WithContext(ctx).First(&definition, id).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

func (r *wikiWordDefinitionRepository) GetByWordID(ctx context.Context, wordID uuid.UUID) ([]wikiModel.WikiWordDefinition, error) {
	var definitions []wikiModel.WikiWordDefinition
	err := r.db.WithContext(ctx).Where("wiki_word_id = ?", wordID).Find(&definitions).Error
	if err != nil {
		return nil, err
	}
	return definitions, nil
}

func (r *wikiWordDefinitionRepository) GetMainDefinitionByWordID(ctx context.Context, wordID uuid.UUID) (*wikiModel.WikiWordDefinition, error) {
	var definition wikiModel.WikiWordDefinition
	err := r.db.WithContext(ctx).
		Where("wiki_word_id = ? AND is_main_definition = true", wordID).
		First(&definition).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

func (r *wikiWordDefinitionRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordDefinition, int64, error) {
	var definitions []wikiModel.WikiWordDefinition
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiWordDefinition{})

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
