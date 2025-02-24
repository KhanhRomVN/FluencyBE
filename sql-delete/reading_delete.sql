-- Drop all triggers
DROP TRIGGER IF EXISTS trigger_reading_questions_version_insert ON reading_questions;
DROP TRIGGER IF EXISTS trigger_reading_questions_version_update ON reading_questions;
DROP TRIGGER IF EXISTS trigger_reading_questions_version_delete ON reading_questions;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_questions_version_insert ON reading_fill_in_the_blank_questions;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_questions_version_update ON reading_fill_in_the_blank_questions;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_questions_version_delete ON reading_fill_in_the_blank_questions;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_answers_version_insert ON reading_fill_in_the_blank_answers;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_answers_version_update ON reading_fill_in_the_blank_answers;
DROP TRIGGER IF EXISTS trigger_reading_fill_in_the_blank_answers_version_delete ON reading_fill_in_the_blank_answers;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_questions_version_insert ON reading_choice_one_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_questions_version_update ON reading_choice_one_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_questions_version_delete ON reading_choice_one_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_options_version_insert ON reading_choice_one_options;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_options_version_update ON reading_choice_one_options;
DROP TRIGGER IF EXISTS trigger_reading_choice_one_options_version_delete ON reading_choice_one_options;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_questions_version_insert ON reading_choice_multi_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_questions_version_update ON reading_choice_multi_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_questions_version_delete ON reading_choice_multi_questions;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_options_version_insert ON reading_choice_multi_options;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_options_version_update ON reading_choice_multi_options;
DROP TRIGGER IF EXISTS trigger_reading_choice_multi_options_version_delete ON reading_choice_multi_options;
DROP TRIGGER IF EXISTS trigger_reading_true_falses_version_insert ON reading_true_falses;
DROP TRIGGER IF EXISTS trigger_reading_true_falses_version_update ON reading_true_falses;
DROP TRIGGER IF EXISTS trigger_reading_true_falses_version_delete ON reading_true_falses;
DROP TRIGGER IF EXISTS trigger_reading_matchings_version_insert ON reading_matchings;
DROP TRIGGER IF EXISTS trigger_reading_matchings_version_update ON reading_matchings;
DROP TRIGGER IF EXISTS trigger_reading_matchings_version_delete ON reading_matchings;

-- Drop all indexes
DROP INDEX IF EXISTS idx_reading_questions_type;
DROP INDEX IF EXISTS idx_reading_questions_topic;
DROP INDEX IF EXISTS idx_reading_fill_blank_question_id;
DROP INDEX IF EXISTS idx_reading_fill_blank_answers_question_id;
DROP INDEX IF EXISTS idx_reading_one_correct_option_per_question;
DROP INDEX IF EXISTS idx_reading_choice_one_options;
DROP INDEX IF EXISTS idx_reading_choice_one_correct_options;
DROP INDEX IF EXISTS idx_reading_choice_multi_options;
DROP INDEX IF EXISTS idx_reading_choice_multi_correct_options;
DROP INDEX IF EXISTS idx_reading_true_false_question_id;
DROP INDEX IF EXISTS idx_reading_matching_question_id;
DROP INDEX IF EXISTS idx_reading_matching_question;
DROP INDEX IF EXISTS idx_reading_matching_answer;

-- Drop constraints
ALTER TABLE IF EXISTS reading_matchings
DROP CONSTRAINT IF EXISTS unique_reading_question_per_matching,
DROP CONSTRAINT IF EXISTS unique_reading_answer_per_matching;

ALTER TABLE IF EXISTS reading_true_falses
DROP CONSTRAINT IF EXISTS unique_question_per_true_false;

ALTER TABLE IF EXISTS reading_choice_multi_options
DROP CONSTRAINT IF EXISTS unique_multi_option_per_question;

ALTER TABLE IF EXISTS reading_choice_one_options
DROP CONSTRAINT IF EXISTS unique_option_per_question;

ALTER TABLE IF EXISTS reading_fill_in_the_blank_questions
DROP CONSTRAINT IF EXISTS unique_reading_fill_in_the_blank_question;

-- Drop all tables (with CASCADE)
DROP TABLE IF EXISTS reading_fill_in_the_blank_answers CASCADE;
DROP TABLE IF EXISTS reading_fill_in_the_blank_questions CASCADE;
DROP TABLE IF EXISTS reading_choice_one_options CASCADE;
DROP TABLE IF EXISTS reading_choice_one_questions CASCADE;
DROP TABLE IF EXISTS reading_choice_multi_options CASCADE;
DROP TABLE IF EXISTS reading_choice_multi_questions CASCADE;
DROP TABLE IF EXISTS reading_true_falses CASCADE;
DROP TABLE IF EXISTS reading_matchings CASCADE;
DROP TABLE IF EXISTS reading_questions CASCADE;

-- Drop all functions
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS reading_question_version_update();
