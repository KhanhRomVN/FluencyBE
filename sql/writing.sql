-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

--! =================================================================
--! FUNCTIONS
--! =================================================================
-- Function to manage version for writing questions
CREATE OR REPLACE FUNCTION writing_question_version_update()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP IN ('INSERT', 'UPDATE', 'DELETE') THEN
        IF TG_TABLE_NAME = 'writing_questions' THEN
            IF TG_OP = 'UPDATE' THEN
                IF NEW IS DISTINCT FROM OLD THEN
                    NEW.version := OLD.version + 1;
                END IF;
            END IF;
            RETURN NEW;
        ELSE
            DECLARE
                record_id UUID;
                parent_id UUID;
            BEGIN
                IF TG_OP = 'DELETE' THEN
                    record_id := OLD.id;
                    IF TG_TABLE_NAME IN ('writing_sentence_completions', 'writing_essays') THEN
                        parent_id := OLD.writing_question_id;
                        UPDATE writing_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = parent_id;
                        RETURN OLD;
                    END IF;
                ELSE
                    record_id := NEW.id;
                    IF TG_TABLE_NAME IN ('writing_sentence_completions', 'writing_essays') THEN
                        parent_id := NEW.writing_question_id;
                        UPDATE writing_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = parent_id;
                        RETURN NEW;
                    END IF;
                END IF;
            END;
        END IF;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

--! =================================================================
--! WRITING QUESTIONS - Base table
--! =================================================================
CREATE TABLE IF NOT EXISTS writing_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    topic VARCHAR(100)[] NOT NULL CHECK (array_length(topic, 1) > 0),
    instruction TEXT NOT NULL CHECK (length(trim(instruction)) > 0),
    image_urls TEXT[],
    max_time INT NOT NULL CHECK (max_time > 0),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Optimize search by type
CREATE INDEX IF NOT EXISTS idx_writing_questions_type 
ON writing_questions(type);

-- Optimize search by topic
CREATE INDEX IF NOT EXISTS idx_writing_questions_topic 
ON writing_questions USING GIN(topic);

--! =================================================================
--! SENTENCE COMPLETION
--! =================================================================
CREATE TABLE IF NOT EXISTS writing_sentence_completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    writing_question_id UUID NOT NULL REFERENCES writing_questions(id) ON DELETE CASCADE,
    example_sentence TEXT NOT NULL CHECK (length(trim(example_sentence)) > 0),
    given_part_sentence TEXT NOT NULL CHECK (length(trim(given_part_sentence)) > 0),
    position VARCHAR(10) NOT NULL CHECK (position IN ('start', 'end')),
    required_words TEXT[] NOT NULL CHECK (array_length(required_words, 1) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    min_words INT NOT NULL CHECK (min_words > 0),
    max_words INT NOT NULL CHECK (max_words >= min_words),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_sentence_completion_per_question UNIQUE (writing_question_id)
);

-- Optimize search for sentence completions
CREATE INDEX IF NOT EXISTS idx_writing_sentence_completion_question_id 
ON writing_sentence_completions(writing_question_id);

CREATE INDEX IF NOT EXISTS idx_writing_sentence_completion_text 
ON writing_sentence_completions USING gin(to_tsvector('english', example_sentence));

--! =================================================================
--! ESSAYS
--! =================================================================
CREATE TABLE IF NOT EXISTS writing_essays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    writing_question_id UUID NOT NULL REFERENCES writing_questions(id) ON DELETE CASCADE,
    essay_type VARCHAR(50) NOT NULL CHECK (length(trim(essay_type)) > 0),
    required_points TEXT[] NOT NULL CHECK (array_length(required_points, 1) > 0),
    min_words INT NOT NULL CHECK (min_words > 0),
    max_words INT NOT NULL CHECK (max_words >= min_words),
    sample_essay TEXT NOT NULL CHECK (length(trim(sample_essay)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_essay_per_question UNIQUE (writing_question_id)
);

-- Optimize search for essays
CREATE INDEX IF NOT EXISTS idx_writing_essays_question_id 
ON writing_essays(writing_question_id);

CREATE INDEX IF NOT EXISTS idx_writing_essays_type 
ON writing_essays(essay_type);

CREATE INDEX IF NOT EXISTS idx_writing_essays_text 
ON writing_essays USING gin(to_tsvector('english', sample_essay));

--! =================================================================
--! TRIGGERS
--! =================================================================

-- Version tracking triggers for writing_questions
CREATE TRIGGER trigger_writing_questions_version_insert
AFTER INSERT ON writing_questions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_questions_version_update
AFTER UPDATE ON writing_questions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_questions_version_delete
AFTER DELETE ON writing_questions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

-- Version tracking triggers for sentence completions
CREATE TRIGGER trigger_writing_sentence_completions_version_insert
AFTER INSERT ON writing_sentence_completions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_sentence_completions_version_update
AFTER UPDATE ON writing_sentence_completions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_sentence_completions_version_delete
AFTER DELETE ON writing_sentence_completions
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

-- Version tracking triggers for essays
CREATE TRIGGER trigger_writing_essays_version_insert
AFTER INSERT ON writing_essays
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_essays_version_update
AFTER UPDATE ON writing_essays
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

CREATE TRIGGER trigger_writing_essays_version_delete
AFTER DELETE ON writing_essays
FOR EACH ROW
EXECUTE FUNCTION writing_question_version_update();

-- Updated timestamp triggers
CREATE TRIGGER update_writing_questions_updated_at
BEFORE UPDATE ON writing_questions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_writing_sentence_completions_updated_at
BEFORE UPDATE ON writing_sentence_completions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_writing_essays_updated_at
BEFORE UPDATE ON writing_essays
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

--! =================================================================
--! COMMENTS
--! =================================================================
COMMENT ON TABLE writing_questions 
IS 'Base table containing common information for all writing question types';

COMMENT ON TABLE writing_sentence_completions
IS 'Table containing sentence completion questions';

COMMENT ON TABLE writing_essays
IS 'Table containing essay writing questions';

COMMENT ON COLUMN writing_questions.version 
IS 'Version number that auto-increments when the question or its related data is modified';

COMMENT ON COLUMN writing_questions.id
IS 'New primary key using UUID with auto-generation';

COMMENT ON COLUMN writing_questions.type 
IS 'Type of writing question';

COMMENT ON COLUMN writing_questions.topic 
IS 'Topics covered by the question';

COMMENT ON COLUMN writing_questions.instruction 
IS 'Instructions for completing the question';

COMMENT ON COLUMN writing_questions.max_time 
IS 'Maximum allowed time in seconds';