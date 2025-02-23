--! =================================================================
--! WORDS - Từ vựng
--! =================================================================
CREATE TABLE IF NOT EXISTS wiki_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word TEXT NOT NULL CHECK (length(trim(word)) BETWEEN 1 AND 100),
    pronunciation TEXT NOT NULL CHECK (length(trim(pronunciation)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_word UNIQUE (LOWER(word))
);

-- Optimize word search
CREATE INDEX IF NOT EXISTS idx_wiki_words_word ON wiki_words(LOWER(word));

--! =================================================================
--! WORD DEFINITIONS - Định nghĩa từ
--! =================================================================
CREATE TABLE IF NOT EXISTS wiki_word_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_word_id UUID NOT NULL REFERENCES wiki_words(id) ON DELETE CASCADE,
    means TEXT[] NOT NULL CHECK (array_length(means, 1) > 0),
    is_main_definition BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT one_main_definition_per_word UNIQUE (wiki_word_id, is_main_definition) 
    WHERE is_main_definition = TRUE
);

-- Optimize definition lookups
CREATE INDEX IF NOT EXISTS idx_wiki_word_definitions_word_id 
ON wiki_word_definitions(wiki_word_id);

--! =================================================================
--! DEFINITION SAMPLES - Mẫu câu của định nghĩa từ
--! =================================================================
CREATE TABLE IF NOT EXISTS wiki_word_definition_samples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_word_definition_id UUID NOT NULL REFERENCES wiki_word_definitions(id) ON DELETE CASCADE,
    sample_sentence TEXT NOT NULL CHECK (length(trim(sample_sentence)) > 0),
    sample_sentence_mean TEXT NOT NULL CHECK (length(trim(sample_sentence_mean)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Optimize sample lookups
CREATE INDEX IF NOT EXISTS idx_wiki_word_definition_samples_def_id 
ON wiki_word_definition_samples(wiki_word_definition_id);

--! =================================================================
--! SYNONYMS & ANTONYMS - Từ đồng nghĩa & trái nghĩa của *mỗi* định nghĩa từ
--! =================================================================
CREATE TABLE IF NOT EXISTS wiki_word_synonyms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_word_definition_id UUID NOT NULL REFERENCES wiki_word_definitions(id) ON DELETE CASCADE,
    wiki_synonym_id UUID NOT NULL REFERENCES wiki_words(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_synonym_pair UNIQUE (wiki_word_definition_id, wiki_synonym_id)
);

CREATE TABLE IF NOT EXISTS wiki_word_antonyms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_word_definition_id UUID NOT NULL REFERENCES wiki_word_definitions(id) ON DELETE CASCADE,
    wiki_antonym_id UUID NOT NULL REFERENCES wiki_words(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_antonym_pair UNIQUE (wiki_word_definition_id, wiki_antonym_id)
);

-- Optimize synonym/antonym lookups
CREATE INDEX IF NOT EXISTS idx_wiki_word_synonyms_def_id 
ON wiki_word_synonyms(wiki_word_definition_id);

CREATE INDEX IF NOT EXISTS idx_wiki_word_antonyms_def_id 
ON wiki_word_antonyms(wiki_word_definition_id);

--! =================================================================
--! PHRASES - Cụm từ
--! =================================================================
CREATE TABLE IF NOT EXISTS wiki_phrases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase TEXT NOT NULL CHECK (length(trim(phrase)) BETWEEN 2 AND 255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_phrase UNIQUE (LOWER(phrase))
);

CREATE TABLE IF NOT EXISTS wiki_phrase_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_phrase_id UUID NOT NULL REFERENCES wiki_phrases(id) ON DELETE CASCADE,
    means TEXT[] NOT NULL CHECK (array_length(means, 1) > 0),    is_main_definition BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT one_main_definition_per_phrase UNIQUE (wiki_phrase_id, is_main_definition) 
    WHERE is_main_definition = TRUE
);

CREATE TABLE IF NOT EXISTS wiki_phrase_definition_samples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wiki_phrase_definition_id UUID NOT NULL REFERENCES wiki_phrase_definitions(id) ON DELETE CASCADE,
    sample_sentence TEXT NOT NULL CHECK (length(trim(sample_sentence)) > 0),
    sample_sentence_mean TEXT NOT NULL CHECK (length(trim(sample_sentence_mean)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Optimize phrase lookups
CREATE INDEX IF NOT EXISTS idx_wiki_phrase_definition_samples_def_id 
ON wiki_phrase_definition_samples(wiki_phrase_definition_id);

--! =================================================================
--! TRIGGERS - For timestamp updates
--! =================================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ 
BEGIN
    FOR tbl IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public'
        AND tablename LIKE 'wiki_%'
    LOOP
        EXECUTE format('
            CREATE TRIGGER trigger_%I_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at();',
            tbl.tablename, tbl.tablename);
    END LOOP;
END $$;