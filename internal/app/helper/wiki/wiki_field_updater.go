package wiki

import (
	"fluencybe/internal/app/dto"
	wikiModel "fluencybe/internal/app/model/wiki"
	wikiValidator "fluencybe/internal/app/validator"
	"fluencybe/pkg/logger"

	"github.com/lib/pq"
)

type WikiFieldUpdater struct {
	logger *logger.PrettyLogger
}

func NewWikiFieldUpdater(logger *logger.PrettyLogger) *WikiFieldUpdater {
	return &WikiFieldUpdater{
		logger: logger,
	}
}

// Word field updates
func (u *WikiFieldUpdater) UpdateWordField(word *wikiModel.WikiWord, update dto.UpdateWikiWordRequest) error {
	if update.Word != nil {
		if err := wikiValidator.ValidateWordText(*update.Word); err != nil {
			return err
		}
		word.Word = *update.Word
	}

	if update.Pronunciation != nil {
		if err := wikiValidator.ValidatePronunciation(*update.Pronunciation); err != nil {
			return err
		}
		word.Pronunciation = *update.Pronunciation
	}

	return nil
}

// Word Definition field updates
func (u *WikiFieldUpdater) UpdateWordDefinitionField(definition *wikiModel.WikiWordDefinition, update dto.UpdateWikiWordDefinitionRequest) error {
	if update.Means != nil {
		if err := wikiValidator.ValidateWordMeanings(update.Means); err != nil {
			return err
		}
		definition.Means = pq.StringArray(update.Means)
	}

	if update.IsMainDefinition != nil {
		definition.IsMainDefinition = *update.IsMainDefinition
	}

	return nil
}

// Word Definition Sample field updates
func (u *WikiFieldUpdater) UpdateWordDefinitionSampleField(sample *wikiModel.WikiWordDefinitionSample, update dto.UpdateWikiWordDefinitionSampleRequest) error {
	if update.SampleSentence != nil {
		if err := wikiValidator.ValidateSampleSentence(*update.SampleSentence); err != nil {
			return err
		}
		sample.SampleSentence = *update.SampleSentence
	}

	if update.SampleSentenceMean != nil {
		if err := wikiValidator.ValidateSampleSentenceMean(*update.SampleSentenceMean); err != nil {
			return err
		}
		sample.SampleSentenceMean = *update.SampleSentenceMean
	}

	return nil
}

// Phrase field updates
func (u *WikiFieldUpdater) UpdatePhraseField(phrase *wikiModel.WikiPhrase, update dto.UpdateWikiPhraseRequest) error {
	if update.Phrase != nil {
		if err := wikiValidator.ValidatePhraseText(*update.Phrase); err != nil {
			return err
		}
		phrase.Phrase = *update.Phrase
	}

	if update.Type != nil {
		if err := wikiValidator.ValidatePhraseType(*update.Type); err != nil {
			return err
		}
		phrase.Type = *update.Type
	}

	if update.DifficultyLevel != nil {
		if err := wikiValidator.ValidateDifficultyLevel(*update.DifficultyLevel); err != nil {
			return err
		}
		phrase.DifficultyLevel = *update.DifficultyLevel
	}

	return nil
}

// Phrase Definition field updates
func (u *WikiFieldUpdater) UpdatePhraseDefinitionField(definition *wikiModel.WikiPhraseDefinition, update dto.UpdateWikiPhraseDefinitionRequest) error {
	if update.Mean != nil {
		if err := wikiValidator.ValidatePhraseMeaning(*update.Mean); err != nil {
			return err
		}
		definition.Mean = *update.Mean
	}

	if update.IsMainDefinition != nil {
		definition.IsMainDefinition = *update.IsMainDefinition
	}

	return nil
}

// Phrase Definition Sample field updates
func (u *WikiFieldUpdater) UpdatePhraseDefinitionSampleField(sample *wikiModel.WikiPhraseDefinitionSample, update dto.UpdateWikiPhraseDefinitionSampleRequest) error {
	if update.SampleSentence != nil {
		if err := wikiValidator.ValidateSampleSentence(*update.SampleSentence); err != nil {
			return err
		}
		sample.SampleSentence = *update.SampleSentence
	}

	if update.SampleSentenceMean != nil {
		if err := wikiValidator.ValidateSampleSentenceMean(*update.SampleSentenceMean); err != nil {
			return err
		}
		sample.SampleSentenceMean = *update.SampleSentenceMean
	}

	return nil
}
