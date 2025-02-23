package wiki

import (
	"fluencybe/internal/app/dto"
	wikiModel "fluencybe/internal/app/model/wiki"
)

type WikiWordUpdator interface {
	UpdateField(word *wikiModel.WikiWord, update dto.UpdateWikiWordRequest) error
}

type WikiWordDefinitionUpdator interface {
	UpdateField(definition *wikiModel.WikiWordDefinition, update dto.UpdateWikiWordDefinitionRequest) error
}

type WikiWordDefinitionSampleUpdator interface {
	UpdateField(sample *wikiModel.WikiWordDefinitionSample, update dto.UpdateWikiWordDefinitionSampleRequest) error
}

type WikiPhraseUpdator interface {
	UpdateField(phrase *wikiModel.WikiPhrase, update dto.UpdateWikiPhraseRequest) error
}

type WikiPhraseDefinitionUpdator interface {
	UpdateField(definition *wikiModel.WikiPhraseDefinition, update dto.UpdateWikiPhraseDefinitionRequest) error
}

type WikiPhraseDefinitionSampleUpdator interface {
	UpdateField(sample *wikiModel.WikiPhraseDefinitionSample, update dto.UpdateWikiPhraseDefinitionSampleRequest) error
}
