package wiki

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiPhraseRepository interface {
	Create(ctx context.Context, phrase *wikiModel.WikiPhrase) error
	Update(ctx context.Context, phrase *wikiModel.WikiPhrase) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhrase, error)
	GetByPhrase(ctx context.Context, phrase string) (*wikiModel.WikiPhrase, error)
	List(ctx context.Context, page, pageSize int, query, phraseType string, difficultyLevel *int, sortBy, sortOrder string) ([]wikiModel.WikiPhrase, int64, error)
}

type wikiPhraseRepository struct {
	db *gorm.DB
}

func NewWikiPhraseRepository(db *gorm.DB) WikiPhraseRepository {
	return &wikiPhraseRepository{db: db}
}

func (r *wikiPhraseRepository) Create(ctx context.Context, phrase *wikiModel.WikiPhrase) error {
	return r.db.WithContext(ctx).Create(phrase).Error
}

func (r *wikiPhraseRepository) Update(ctx context.Context, phrase *wikiModel.WikiPhrase) error {
	return r.db.WithContext(ctx).Save(phrase).Error
}

func (r *wikiPhraseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&wikiModel.WikiPhrase{}, id).Error
}

func (r *wikiPhraseRepository) GetByID(ctx context.Context, id uuid.UUID) (*wikiModel.WikiPhrase, error) {
	var phrase wikiModel.WikiPhrase
	err := r.db.WithContext(ctx).First(&phrase, id).Error
	if err != nil {
		return nil, err
	}
	return &phrase, nil
}

func (r *wikiPhraseRepository) GetByPhrase(ctx context.Context, phrase string) (*wikiModel.WikiPhrase, error) {
	var wikiPhrase wikiModel.WikiPhrase
	err := r.db.WithContext(ctx).Where("phrase = ?", phrase).First(&wikiPhrase).Error
	if err != nil {
		return nil, err
	}
	return &wikiPhrase, nil
}

func (r *wikiPhraseRepository) List(ctx context.Context, page, pageSize int, query, phraseType string, difficultyLevel *int, sortBy, sortOrder string) ([]wikiModel.WikiPhrase, int64, error) {
	var phrases []wikiModel.WikiPhrase
	var total int64

	db := r.db.WithContext(ctx).Model(&wikiModel.WikiPhrase{})

	// Apply filters
	if query != "" {
		db = db.Where("phrase ILIKE ?", "%"+query+"%")
	}

	if phraseType != "" {
		db = db.Where("type = ?", phraseType)
	}

	if difficultyLevel != nil {
		db = db.Where("difficulty_level = ?", *difficultyLevel)
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
	err = db.Offset(offset).Limit(pageSize).Find(&phrases).Error
	if err != nil {
		return nil, 0, err
	}

	return phrases, total, nil
}
