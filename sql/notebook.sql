-- Tạo ENUM cho notebook type
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notebook_type') THEN
        CREATE TYPE notebook_type AS ENUM ('word', 'phrase');
    END IF;
END $$;

-- Tạo function cập nhật updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Tạo bảng notebooks
CREATE TABLE IF NOT EXISTS notebooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(50) NOT NULL,
    description VARCHAR(500) NOT NULL,
    type notebook_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notebooks_user_id ON notebooks(user_id);
CREATE INDEX IF NOT EXISTS idx_notebooks_title ON notebooks(title);

-- Trigger cập nhật updated_at cho notebooks
CREATE TRIGGER trigger_notebooks_updated_at
BEFORE UPDATE ON notebooks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- Tạo bảng notebook_words
CREATE TABLE IF NOT EXISTS notebook_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notebook_id UUID NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE,
    sequence INT NOT NULL,
    wiki_word_id UUID NOT NULL REFERENCES wiki_words(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(notebook_id, sequence) -- Đảm bảo thứ tự trong mỗi notebook là duy nhất
);

CREATE INDEX IF NOT EXISTS idx_notebook_words_notebook_id ON notebook_words(notebook_id);
CREATE INDEX IF NOT EXISTS idx_notebook_words_wiki_word_id ON notebook_words(wiki_word_id);

-- Trigger cập nhật updated_at cho notebook_words
CREATE TRIGGER trigger_notebook_words_updated_at
BEFORE UPDATE ON notebook_words
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- Tạo bảng notebook_phrases
CREATE TABLE IF NOT EXISTS notebook_phrases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notebook_id UUID NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE,
    sequence INT NOT NULL,
    wiki_phrase_id UUID NOT NULL REFERENCES wiki_phrases(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(notebook_id, sequence) -- Đảm bảo thứ tự trong mỗi notebook là duy nhất
);

CREATE INDEX IF NOT EXISTS idx_notebook_phrases_notebook_id ON notebook_phrases(notebook_id);
CREATE INDEX IF NOT EXISTS idx_notebook_phrases_wiki_phrase_id ON notebook_phrases(wiki_phrase_id);
