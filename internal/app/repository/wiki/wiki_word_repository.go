package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordRepository interface {
	Create(ctx context.Context, word *wikiModel.WikiWord) error
	Update(ctx context.Context, word *wikiModel.WikiWord) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWord, error)
	GetByWord(ctx context.Context, word string) (*wikiModel.WikiWord, error)
	List(ctx context.Context, page, pageSize int, query, sortBy, sortOrder string) ([]wikiModel.WikiWord, int64, error)
}

type wikiWordRepository struct {
	db *gorm.DB
}

func NewWikiWordRepository(db *gorm.DB) WikiWordRepository {
	return &wikiWordRepository{db: db}
}

func (r *wikiWordRepository) Create(ctx context.Context, word *wikiModel.WikiWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *wikiWordRepository) Update(ctx context.Context, word *wikiModel.WikiWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

func (r *wikiWordRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiWord{}, id).Error
}

func (r *wikiWordRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiWord, error) {
	var word wikiModel.WikiWord
	err := r.db.WithContext(ctx).First(&word, id).Error
	if err != nil {
		return nil, err
	}
	return &word, nil
}

func (r *wikiWordRepository) GetByWord(ctx context.Context, word string) (*wikiModel.WikiWord, error) {
	var wikiWord wikiModel.WikiWord
	err := r.db.WithContext(ctx).Where("word = ?", word).First(&wikiWord).Error
	if err != nil {
		return nil, err
	}
	return &wikiWord, nil
}

func (r *wikiWordRepository) List(ctx context.Context, page, pageSize int, query, sortBy, sortOrder string) ([]wikiModel.WikiWord, int64, error) {
	var words []wikiModel.WikiWord
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiWord{})

	if query != "" {
		db = db.Where("word ILIKE ?", "%"+query+"%")
	}

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
	err = db.Offset(offset).Limit(pageSize).Find(&words).Error
	if err != nil {
		return nil, 0, err
	}

	return words, total, nil
}
