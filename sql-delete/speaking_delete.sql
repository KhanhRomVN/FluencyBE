-- Drop all triggers
DROP TRIGGER IF EXISTS trigger_speaking_questions_version_insert ON speaking_questions;
DROP TRIGGER IF EXISTS trigger_speaking_questions_version_update ON speaking_questions;
DROP TRIGGER IF EXISTS trigger_speaking_questions_version_delete ON speaking_questions;

DROP TRIGGER IF EXISTS trigger_speaking_word_repetitions_version_insert ON speaking_word_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_word_repetitions_version_update ON speaking_word_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_word_repetitions_version_delete ON speaking_word_repetitions;

DROP TRIGGER IF EXISTS trigger_speaking_phrase_repetitions_version_insert ON speaking_phrase_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_phrase_repetitions_version_update ON speaking_phrase_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_phrase_repetitions_version_delete ON speaking_phrase_repetitions;

DROP TRIGGER IF EXISTS trigger_speaking_paragraph_repetitions_version_insert ON speaking_paragraph_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_paragraph_repetitions_version_update ON speaking_paragraph_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_paragraph_repetitions_version_delete ON speaking_paragraph_repetitions;

DROP TRIGGER IF EXISTS trigger_speaking_open_paragraphs_version_insert ON speaking_open_paragraphs;
DROP TRIGGER IF EXISTS trigger_speaking_open_paragraphs_version_update ON speaking_open_paragraphs;
DROP TRIGGER IF EXISTS trigger_speaking_open_paragraphs_version_delete ON speaking_open_paragraphs;

DROP TRIGGER IF EXISTS trigger_speaking_picture_descriptions_version_insert ON speaking_picture_descriptions;
DROP TRIGGER IF EXISTS trigger_speaking_picture_descriptions_version_update ON speaking_picture_descriptions;
DROP TRIGGER IF EXISTS trigger_speaking_picture_descriptions_version_delete ON speaking_picture_descriptions;

DROP TRIGGER IF EXISTS trigger_speaking_conv_repetitions_version_insert ON speaking_conversational_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_conv_repetitions_version_update ON speaking_conversational_repetitions;
DROP TRIGGER IF EXISTS trigger_speaking_conv_repetitions_version_delete ON speaking_conversational_repetitions;

DROP TRIGGER IF EXISTS trigger_speaking_conv_repetition_qas_version_insert ON speaking_conversational_repetition_qas;
DROP TRIGGER IF EXISTS trigger_speaking_conv_repetition_qas_version_update ON speaking_conversational_repetition_qas;
DROP TRIGGER IF EXISTS trigger_speaking_conv_repetition_qas_version_delete ON speaking_conversational_repetition_qas;

DROP TRIGGER IF EXISTS trigger_speaking_conv_opens_version_insert ON speaking_conversational_opens;
DROP TRIGGER IF EXISTS trigger_speaking_conv_opens_version_update ON speaking_conversational_opens;
DROP TRIGGER IF EXISTS trigger_speaking_conv_opens_version_delete ON speaking_conversational_opens;

DROP TRIGGER IF EXISTS trigger_speaking_conv_open_qas_version_insert ON speaking_conversational_open_qas;
DROP TRIGGER IF EXISTS trigger_speaking_conv_open_qas_version_update ON speaking_conversational_open_qas;
DROP TRIGGER IF EXISTS trigger_speaking_conv_open_qas_version_delete ON speaking_conversational_open_qas;

-- Drop updated_at triggers
DROP TRIGGER IF EXISTS update_speaking_questions_updated_at ON speaking_questions;
DROP TRIGGER IF EXISTS update_speaking_word_repetitions_updated_at ON speaking_word_repetitions;
DROP TRIGGER IF EXISTS update_speaking_phrase_repetitions_updated_at ON speaking_phrase_repetitions;
DROP TRIGGER IF EXISTS update_speaking_paragraph_repetitions_updated_at ON speaking_paragraph_repetitions;
DROP TRIGGER IF EXISTS update_speaking_open_paragraphs_updated_at ON speaking_open_paragraphs;
DROP TRIGGER IF EXISTS update_speaking_picture_descriptions_updated_at ON speaking_picture_descriptions;
DROP TRIGGER IF EXISTS update_speaking_conv_repetitions_updated_at ON speaking_conversational_repetitions;
DROP TRIGGER IF EXISTS update_speaking_conv_repetition_qas_updated_at ON speaking_conversational_repetition_qas;
DROP TRIGGER IF EXISTS update_speaking_conv_opens_updated_at ON speaking_conversational_opens;
DROP TRIGGER IF EXISTS update_speaking_conv_open_qas_updated_at ON speaking_conversational_open_qas;

-- Drop all indexes
DROP INDEX IF EXISTS idx_speaking_questions_type;
DROP INDEX IF EXISTS idx_speaking_questions_topic;

DROP INDEX IF EXISTS idx_speaking_word_repetition_question_id;
DROP INDEX IF EXISTS idx_speaking_word_repetition_word;

DROP INDEX IF EXISTS idx_speaking_phrase_repetition_question_id;
DROP INDEX IF EXISTS idx_speaking_phrase_repetition_phrase;

DROP INDEX IF EXISTS idx_speaking_paragraph_repetition_question_id;
DROP INDEX IF EXISTS idx_speaking_paragraph_repetition_text;

DROP INDEX IF EXISTS idx_speaking_open_paragraph_question_id;
DROP INDEX IF EXISTS idx_speaking_open_paragraph_text;

DROP INDEX IF EXISTS idx_speaking_picture_description_question_id;
DROP INDEX IF EXISTS idx_speaking_picture_description_text;

DROP INDEX IF EXISTS idx_speaking_conv_repetition_question_id;
DROP INDEX IF EXISTS idx_speaking_conv_repetition_qas;
DROP INDEX IF EXISTS idx_speaking_conv_repetition_qa_text;

DROP INDEX IF EXISTS idx_speaking_conv_open_question_id;
DROP INDEX IF EXISTS idx_speaking_conv_open_qas;
DROP INDEX IF EXISTS idx_speaking_conv_open_qa_text;

-- Drop constraints
ALTER TABLE IF EXISTS speaking_word_repetitions
DROP CONSTRAINT IF EXISTS unique_word_per_question;

ALTER TABLE IF EXISTS speaking_phrase_repetitions
DROP CONSTRAINT IF EXISTS unique_phrase_per_question;

ALTER TABLE IF EXISTS speaking_paragraph_repetitions
DROP CONSTRAINT IF EXISTS unique_paragraph_per_question;

ALTER TABLE IF EXISTS speaking_open_paragraphs
DROP CONSTRAINT IF EXISTS unique_open_paragraph_per_question;

ALTER TABLE IF EXISTS speaking_picture_descriptions
DROP CONSTRAINT IF EXISTS unique_picture_description_per_question;

ALTER TABLE IF EXISTS speaking_conversational_repetitions
DROP CONSTRAINT IF EXISTS unique_conversational_repetition_per_question;

ALTER TABLE IF EXISTS speaking_conversational_opens
DROP CONSTRAINT IF EXISTS unique_conversational_open_per_question;

-- Drop all tables (in correct order due to dependencies)
DROP TABLE IF EXISTS speaking_conversational_open_qas CASCADE;
DROP TABLE IF EXISTS speaking_conversational_opens CASCADE;
DROP TABLE IF EXISTS speaking_conversational_repetition_qas CASCADE;
DROP TABLE IF EXISTS speaking_conversational_repetitions CASCADE;
DROP TABLE IF EXISTS speaking_picture_descriptions CASCADE;
DROP TABLE IF EXISTS speaking_open_paragraphs CASCADE;
DROP TABLE IF EXISTS speaking_paragraph_repetitions CASCADE;
DROP TABLE IF EXISTS speaking_phrase_repetitions CASCADE;
DROP TABLE IF EXISTS speaking_word_repetitions CASCADE;
DROP TABLE IF EXISTS speaking_questions CASCADE;

-- Drop all functions
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS speaking_question_version_update();

-- Drop ENUM type
DROP TYPE IF EXISTS speaking_question_type;

-- Clear all comments
COMMENT ON TABLE speaking_questions IS NULL;
COMMENT ON TABLE speaking_word_repetitions IS NULL;
COMMENT ON TABLE speaking_phrase_repetitions IS NULL;
COMMENT ON TABLE speaking_paragraph_repetitions IS NULL;
COMMENT ON TABLE speaking_open_paragraphs IS NULL;
COMMENT ON TABLE speaking_picture_descriptions IS NULL;
COMMENT ON TABLE speaking_conversational_repetitions IS NULL;
COMMENT ON TABLE speaking_conversational_opens IS NULL;

COMMENT ON COLUMN speaking_questions.version IS NULL;
COMMENT ON COLUMN speaking_questions.id IS NULL;
COMMENT ON COLUMN speaking_questions.type IS NULL;
COMMENT ON COLUMN speaking_questions.topic IS NULL;
COMMENT ON COLUMN speaking_questions.instruction IS NULL;
COMMENT ON COLUMN speaking_questions.max_time IS NULL;