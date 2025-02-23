package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordSynonymRepository interface {
	Create(ctx context.Context, synonym *wikiModel.WikiWordSynonym) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordSynonym, error)
	GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordSynonym, error)
	GetBySynonymID(ctx context.Context, synonymID uuid.UUID) ([]wikiModel.WikiWordSynonym, error)
	List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordSynonym, int64, error)
	DeleteByDefinitionIDAndSynonymID(ctx context.Context, definitionID, synonymID uuid.UUID) error
}

type wikiWordSynonymRepository struct {
	db *gorm.DB
}

func NewWikiWordSynonymRepository(db *gorm.DB) WikiWordSynonymRepository {
	return &wikiWordSynonymRepository{db: db}
}

func (r *wikiWordSynonymRepository) Create(ctx context.Context, synonym *wikiModel.WikiWordSynonym) error {
	return r.db.WithContext(ctx).Create(synonym).Error
}

func (r *wikiWordSynonymRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiWordSynonym{}, id).Error
}

func (r *wikiWordSynonymRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWordSynonym, error) {
	var synonym wikiModel.WikiWordSynonym
	err := r.db.WithContext(ctx).First(&synonym, id).Error
	if err != nil {
		return nil, err
	}
	return &synonym, nil
}

func (r *wikiWordSynonymRepository) GetByDefinitionID(ctx context.Context, definitionID uuid.UUID) ([]wikiModel.WikiWordSynonym, error) {
	var synonyms []wikiModel.WikiWordSynonym
	err := r.db.WithContext(ctx).Where("wiki_word_definition_id = ?", definitionID).Find(&synonyms).Error
	if err != nil {
		return nil, err
	}
	return synonyms, nil
}

func (r *wikiWordSynonymRepository) GetBySynonymID(ctx context.Context, synonymID uuid.UUID) ([]wikiModel.WikiWordSynonym, error) {
	var synonyms []wikiModel.WikiWordSynonym
	err := r.db.WithContext(ctx).Where("wiki_synonym_id = ?", synonymID).Find(&synonyms).Error
	if err != nil {
		return nil, err
	}
	return synonyms, nil
}

func (r *wikiWordSynonymRepository) List(ctx context.Context, page, pageSize int, sortBy, sortOrder string) ([]wikiModel.WikiWordSynonym, int64, error) {
	var synonyms []wikiModel.WikiWordSynonym
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiWordSynonym{})

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
	err = db.Offset(offset).Limit(pageSize).Find(&synonyms).Error
	if err != nil {
		return nil, 0, err
	}

	return synonyms, total, nil
}

func (r *wikiWordSynonymRepository) DeleteByDefinitionIDAndSynonymID(ctx context.Context, definitionID, synonymID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("wiki_word_definition_id = ? AND wiki_synonym_id = ?", definitionID, synonymID).
		Delete(&wikiModel.WikiWordSynonym{}).Error
}
