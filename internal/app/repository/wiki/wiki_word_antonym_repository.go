package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordAntonymRepository interface {
	Create(ctx context.Context, antonym *wikiModel.WikiWordAntonym) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordAntonym, error)
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordAntonym, error)
	GetByAntonymID(ctx context.Context, antonymID uuid.UUID) ([]wikiModel.WikiWordAntonym, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordAntonym, int64, error)
	DeleteByDefinitionIDAndAntonymID(ctx context.Context, definitionID, antonymID uuid.UUID) error
}

type wikiWordAntonymRepository struct {
	db *gorm.DB
}

func NewWikiWordAntonymRepository(db *gorm.DB) WikiWordAntonymRepository {
	return &wikiWordAntonymRepository{db: db}
}

func (r *wikiWordAntonymRepository) Create(ctx context.Context, antonym *wikiModel.WikiWordAntonym) error {
	return r.db.WithContext(ctx).Create(antonym).Error
}

func (r *wikiWordAntonymRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiWordAntonym{}, id).Error
}

func (r *wikiWordAntonymRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordAntonym, error) {
	var antonym wikiModel.WikiWordAntonym
	err := r.db.WithContext(ctx).First(&antonym, id).Error
	if err != nil {
		return nil, err
	}
	return &antonym, nil
}

func (r *wikiWordAntonymRepository) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordAntonym, error) {
	var antonyms []wikiModel.WikiWordAntonym
	err := r.db.WithContext(ctx).Where("wiki_word_definition_id = ?", definitionID).Find(&antonyms).Error
	if err != nil {
		return nil, err
	}
	return antonyms, nil
}

func (r *wikiWordAntonymRepository) GetByAntonymID(ctx context.Context, antonymID uuid.UUID) ([]wikiModel.WikiWordAntonym, error) {
	var antonyms []wikiModel.WikiWordAntonym
	err := r.db.WithContext(ctx).Where("wiki_antonym_id = ?", antonymID).Find(&antonyms).Error
	if err != nil {
		return nil, err
	}
	return antonyms, nil
}

func (r *wikiWordAntonymRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordAntonym, int64, error) {
	var antonyms []wikiModel.WikiWordAntonym
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiWordAntonym{})

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
	err = db.Offset(offset).Limit(pageSize).Find(&antonyms).Error
	if err != nil {
		return nil, 0, err
	}

	return antonyms, total, nil
}

func (r *wikiWordAntonymRepository) DeleteByDefinitionIDAndAntonymID(ctx context.Context, definitionID, antonymID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("wiki_word_definition_id = ? AND wiki_antonym_id = ?", definitionID, antonymID).
		Delete(&wikiModel.WikiWordAntonym{}).Error
}
