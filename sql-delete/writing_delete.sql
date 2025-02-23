-- Drop all triggers
DROP TRIGGER IF EXISTS trigger_writing_questions_version_insert ON writing_questions;
DROP TRIGGER IF EXISTS trigger_writing_questions_version_update ON writing_questions;
DROP TRIGGER IF EXISTS trigger_writing_questions_version_delete ON writing_questions;

DROP TRIGGER IF EXISTS trigger_writing_sentence_completions_version_insert ON writing_sentence_completions;
DROP TRIGGER IF EXISTS trigger_writing_sentence_completions_version_update ON writing_sentence_completions;
DROP TRIGGER IF EXISTS trigger_writing_sentence_completions_version_delete ON writing_sentence_completions;

DROP TRIGGER IF EXISTS trigger_writing_essays_version_insert ON writing_essays;
DROP TRIGGER IF EXISTS trigger_writing_essays_version_update ON writing_essays;
DROP TRIGGER IF EXISTS trigger_writing_essays_version_delete ON writing_essays;

DROP TRIGGER IF EXISTS update_writing_questions_updated_at ON writing_questions;
DROP TRIGGER IF EXISTS update_writing_sentence_completions_updated_at ON writing_sentence_completions;
DROP TRIGGER IF EXISTS update_writing_essays_updated_at ON writing_essays;

-- Drop all indexes
DROP INDEX IF EXISTS idx_writing_questions_type;
DROP INDEX IF EXISTS idx_writing_questions_topic;

DROP INDEX IF EXISTS idx_writing_sentence_completion_question_id;
DROP INDEX IF EXISTS idx_writing_sentence_completion_text;

DROP INDEX IF EXISTS idx_writing_essays_question_id;
DROP INDEX IF EXISTS idx_writing_essays_type;
DROP INDEX IF EXISTS idx_writing_essays_text;

-- Drop constraints
ALTER TABLE IF EXISTS writing_sentence_completions
DROP CONSTRAINT IF EXISTS unique_sentence_completion_per_question;

ALTER TABLE IF EXISTS writing_essays
DROP CONSTRAINT IF EXISTS unique_essay_per_question;

-- Drop all tables (in correct order due to dependencies)
DROP TABLE IF EXISTS writing_essays CASCADE;
DROP TABLE IF EXISTS writing_sentence_completions CASCADE;
DROP TABLE IF EXISTS writing_questions CASCADE;

-- Drop ENUM type
DROP TYPE IF EXISTS writing_question_type;

-- Drop functions
DROP FUNCTION IF EXISTS writing_question_version_update();

-- Clear all comments
COMMENT ON TABLE writing_questions IS NULL;
COMMENT ON TABLE writing_sentence_completions IS NULL;
COMMENT ON TABLE writing_essays IS NULL;
COMMENT ON COLUMN writing_questions.version IS NULL;
COMMENT ON COLUMN writing_questions.id IS NULL;
COMMENT ON COLUMN writing_questions.type IS NULL;
COMMENT ON COLUMN writing_questions.topic IS NULL;
COMMENT ON COLUMN writing_questions.instruction IS NULL;
COMMENT ON COLUMN writing_questions.max_time IS NULL;