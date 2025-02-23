--! =================================================================
--! CLEANUP SCRIPT FOR COURSE MODULE
--! =================================================================

--! =================================================================
--! Drop tables
--! =================================================================
-- Drop in correct order to respect foreign key constraints
DROP TABLE IF EXISTS lesson_questions CASCADE;
DROP TABLE IF EXISTS lessons CASCADE;
DROP TABLE IF EXISTS course_books CASCADE;
DROP TABLE IF EXISTS course_others CASCADE;
DROP TABLE IF EXISTS courses CASCADE;

--! =================================================================
--! Drop functions
--! =================================================================
-- Drop timestamp function
DROP FUNCTION IF EXISTS update_updated_at() CASCADE;

-- Drop sequence management functions
DROP FUNCTION IF EXISTS get_next_lesson_sequence(UUID) CASCADE;
DROP FUNCTION IF EXISTS get_next_lesson_question_sequence(UUID) CASCADE;
DROP FUNCTION IF EXISTS resequence_lessons() CASCADE;
DROP FUNCTION IF EXISTS resequence_lesson_questions() CASCADE;
DROP FUNCTION IF EXISTS swap_lesson_sequence(UUID, UUID) CASCADE;
DROP FUNCTION IF EXISTS swap_lesson_question_sequence(UUID, UUID) CASCADE;

--! =================================================================
--! Drop indexes
--! =================================================================
-- Drop course indexes
DROP INDEX IF EXISTS idx_courses_skills CASCADE;
DROP INDEX IF EXISTS idx_courses_band CASCADE;
DROP INDEX IF EXISTS idx_courses_title_search CASCADE;

-- Drop course book indexes
DROP INDEX IF EXISTS idx_course_books_course_id CASCADE;
DROP INDEX IF EXISTS idx_course_books_publishers CASCADE;
DROP INDEX IF EXISTS idx_course_books_authors CASCADE;
DROP INDEX IF EXISTS idx_course_books_year CASCADE;

-- Drop course other indexes
DROP INDEX IF EXISTS idx_course_others_course_id CASCADE;

-- Drop lesson indexes
DROP INDEX IF EXISTS idx_lessons_course_id CASCADE;
DROP INDEX IF EXISTS idx_lessons_sequence CASCADE;
DROP INDEX IF EXISTS idx_lessons_title_search CASCADE;

-- Drop lesson question indexes
DROP INDEX IF EXISTS idx_lesson_questions_lesson_id CASCADE;
DROP INDEX IF EXISTS idx_lesson_questions_question_id CASCADE;
DROP INDEX IF EXISTS idx_lesson_questions_sequence CASCADE;

--! =================================================================
--! Drop triggers
--! =================================================================
-- Drop sequence management triggers
DROP TRIGGER IF EXISTS trigger_lessons_resequence ON lessons CASCADE;
DROP TRIGGER IF EXISTS trigger_lesson_questions_resequence ON lesson_questions CASCADE;

-- Drop timestamp triggers
DROP TRIGGER IF EXISTS trigger_courses_updated_at ON courses CASCADE;
DROP TRIGGER IF EXISTS trigger_course_books_updated_at ON course_books CASCADE;
DROP TRIGGER IF EXISTS trigger_course_others_updated_at ON course_others CASCADE;
DROP TRIGGER IF EXISTS trigger_lessons_updated_at ON lessons CASCADE;
DROP TRIGGER IF EXISTS trigger_lesson_questions_updated_at ON lesson_questions CASCADE;

--! =================================================================
--! Drop comments
--! =================================================================
-- Drop table comments
COMMENT ON TABLE courses IS NULL;
COMMENT ON TABLE course_books IS NULL;
COMMENT ON TABLE course_others IS NULL;
COMMENT ON TABLE lessons IS NULL;
COMMENT ON TABLE lesson_questions IS NULL;

-- Drop column comments
COMMENT ON COLUMN courses.type IS NULL;
COMMENT ON COLUMN courses.skills IS NULL;
COMMENT ON COLUMN courses.band IS NULL;

-- Drop function comments
COMMENT ON FUNCTION get_next_lesson_sequence(UUID) IS NULL;
COMMENT ON FUNCTION get_next_lesson_question_sequence(UUID) IS NULL;
COMMENT ON FUNCTION resequence_lessons() IS NULL;
COMMENT ON FUNCTION resequence_lesson_questions() IS NULL;
COMMENT ON FUNCTION swap_lesson_sequence(UUID, UUID) IS NULL;
COMMENT ON FUNCTION swap_lesson_question_sequence(UUID, UUID) IS NULL;