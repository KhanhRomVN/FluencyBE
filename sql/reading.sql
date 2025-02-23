-- Drop existing objects if needed
DO $$ 
BEGIN
    DROP FUNCTION IF EXISTS update_updated_at() CASCADE;
    DROP FUNCTION IF EXISTS reading_question_version_update() CASCADE;
EXCEPTION 
    WHEN OTHERS THEN NULL;
END $$;

-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

--! =================================================================
--! FUNCTIONS
--! =================================================================
-- Function tự động cập nhật timestamp
DO $outer$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_proc WHERE proname = 'update_updated_at') THEN
        CREATE OR REPLACE FUNCTION update_updated_at()
        RETURNS TRIGGER AS $body$
        BEGIN
            IF NEW IS DISTINCT FROM OLD THEN
                NEW.updated_at = CURRENT_TIMESTAMP;
            END IF;
            RETURN NEW;
        END;
        $body$ LANGUAGE plpgsql;
    END IF;
END
$outer$;

-- Function quản lý version cho reading questions
DO $outer$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_proc WHERE proname = 'reading_question_version_update') THEN
        CREATE OR REPLACE FUNCTION reading_question_version_update()
        RETURNS TRIGGER AS $body$
        BEGIN
            IF TG_OP IN ('INSERT', 'UPDATE', 'DELETE') THEN
                IF TG_TABLE_NAME = 'reading_questions' THEN
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
                            -- Handle direct child tables
                            IF TG_TABLE_NAME IN ('reading_true_falses',
                                               'reading_fill_in_the_blank_questions',
                                               'reading_choice_one_questions',
                                               'reading_choice_multi_questions',
                                               'reading_matchings') THEN
                                parent_id := OLD.reading_question_id;
                                UPDATE reading_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id = parent_id;
                                RETURN OLD;
                            END IF;
                            -- For nested child tables
                            IF TG_TABLE_NAME = 'reading_fill_in_the_blank_answers' THEN
                                UPDATE reading_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.reading_question_id
                                    FROM reading_fill_in_the_blank_questions q
                                    WHERE q.id = OLD.reading_fill_in_the_blank_question_id
                                );
                                RETURN OLD;
                            ELSIF TG_TABLE_NAME = 'reading_choice_one_options' THEN
                                UPDATE reading_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.reading_question_id
                                    FROM reading_choice_one_questions q
                                    WHERE q.id = OLD.reading_choice_one_question_id
                                );
                                RETURN OLD;
                            ELSIF TG_TABLE_NAME = 'reading_choice_multi_options' THEN
                                UPDATE reading_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.reading_question_id
                                    FROM reading_choice_multi_questions q
                                    WHERE q.id = OLD.reading_choice_multi_question_id
                                );
                                RETURN OLD;
                            END IF;
                        ELSE
                            record_id := NEW.id;
                            -- Handle direct child tables
                            IF TG_TABLE_NAME IN ('reading_true_falses', 'reading_matchings') THEN
                                parent_id := NEW.reading_question_id;
                                UPDATE reading_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id = parent_id;
                                RETURN NEW;
                            END IF;
                        END IF;
                        -- Handle other tables through joins
                        UPDATE reading_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id IN (
                            SELECT q.reading_question_id
                            FROM reading_fill_in_the_blank_questions q
                            INNER JOIN reading_fill_in_the_blank_answers a
                            ON a.reading_fill_in_the_blank_question_id = q.id
                            WHERE a.id = record_id
                            UNION
                            SELECT q.reading_question_id
                            FROM reading_choice_one_questions q
                            INNER JOIN reading_choice_one_options o
                            ON o.reading_choice_one_question_id = q.id
                            WHERE o.id = record_id
                            UNION
                            SELECT q.reading_question_id
                            FROM reading_choice_multi_questions q
                            INNER JOIN reading_choice_multi_options o
                            ON o.reading_choice_multi_question_id = q.id
                            WHERE o.id = record_id
                            UNION
                            SELECT reading_question_id FROM reading_true_falses WHERE id = record_id
                            UNION
                            SELECT reading_question_id FROM reading_fill_in_the_blank_questions WHERE id = record_id
                            UNION
                            SELECT reading_question_id FROM reading_choice_one_questions WHERE id = record_id
                            UNION
                            SELECT reading_question_id FROM reading_choice_multi_questions WHERE id = record_id
                        );
                    END;
                END IF;
            END IF;
            RETURN COALESCE(NEW, OLD);
        END;
        $body$ LANGUAGE plpgsql;
    END IF;
END
$outer$;


--! =================================================================
--! READING QUESTIONS - Bảng câu hỏi gốc
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    topic VARCHAR(100)[] NOT NULL,
    instruction TEXT NOT NULL CHECK (length(trim(instruction)) > 0),
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    passages TEXT[] NOT NULL CHECK (array_length(passages, 1) > 0),
    image_urls TEXT[],
    max_time INT NOT NULL CHECK (max_time > 0),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu tìm kiếm theo loại câu hỏi
CREATE INDEX IF NOT EXISTS idx_reading_questions_type 
ON reading_questions(type);

-- Index tối ưu tìm kiếm theo chủ đề
CREATE INDEX IF NOT EXISTS idx_reading_questions_topic 
ON reading_questions USING GIN(topic);

--! =================================================================
--! TRUE/FALSE/NOT GIVEN
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_true_falses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_question_id UUID NOT NULL REFERENCES reading_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    answer TEXT NOT NULL CHECK (answer IN ('TRUE', 'FALSE', 'NOT GIVEN')),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_question_per_true_false UNIQUE (reading_question_id, question)
);

-- Index tối ưu cho True/False
CREATE INDEX IF NOT EXISTS idx_reading_true_false_question_id 
ON reading_true_falses(reading_question_id);

--! =================================================================
--! FILL IN THE BLANK - Điền vào chỗ trống  
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_fill_in_the_blank_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_question_id UUID NOT NULL REFERENCES reading_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_reading_fill_in_the_blank_question UNIQUE (reading_question_id)
);

CREATE TABLE IF NOT EXISTS reading_fill_in_the_blank_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_fill_in_the_blank_question_id UUID NOT NULL 
        REFERENCES reading_fill_in_the_blank_questions(id) ON DELETE CASCADE,
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu tìm kiếm câu hỏi theo reading_question_id
CREATE INDEX IF NOT EXISTS idx_reading_fill_blank_question_id 
ON reading_fill_in_the_blank_questions(reading_question_id);

-- Index tối ưu tìm kiếm đáp án theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_reading_fill_blank_answers_question_id 
ON reading_fill_in_the_blank_answers(reading_fill_in_the_blank_question_id);

--! =================================================================
--! MATCHING PARAGRAPHS - Nối đoạn văn
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_matchings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_question_id UUID NOT NULL REFERENCES reading_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_reading_question_per_matching UNIQUE (reading_question_id, question),
    CONSTRAINT unique_reading_answer_per_matching UNIQUE (reading_question_id, answer)
);

-- Index tối ưu cho Matching
CREATE INDEX IF NOT EXISTS idx_reading_matching_question_id 
ON reading_matchings(reading_question_id);

CREATE INDEX IF NOT EXISTS idx_reading_matching_question 
ON reading_matchings USING gin(to_tsvector('english', question));

CREATE INDEX IF NOT EXISTS idx_reading_matching_answer 
ON reading_matchings USING gin(to_tsvector('english', answer));

--! =================================================================
--! CHOICE ONE - Chọn một đáp án
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_choice_one_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_question_id UUID NOT NULL REFERENCES reading_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_reading_choice_one_question UNIQUE (reading_question_id)
);

CREATE TABLE IF NOT EXISTS reading_choice_one_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_choice_one_question_id UUID NOT NULL 
        REFERENCES reading_choice_one_questions(id) ON DELETE CASCADE,
    options TEXT NOT NULL CHECK (length(trim(options)) > 0),
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_option_per_question 
        UNIQUE (reading_choice_one_question_id, options)
);

-- Create a partial unique index instead of the WHERE constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_reading_one_correct_option_per_question 
ON reading_choice_one_options (reading_choice_one_question_id) 
WHERE is_correct = TRUE;

-- Index tối ưu tìm kiếm các lựa chọn theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_reading_choice_one_options 
ON reading_choice_one_options(reading_choice_one_question_id, options);

-- Index tối ưu tìm kiếm đáp án đúng
CREATE INDEX IF NOT EXISTS idx_reading_choice_one_correct_options 
ON reading_choice_one_options(reading_choice_one_question_id) 
WHERE is_correct = TRUE;

--! =================================================================
--! CHOICE MULTI - Chọn nhiều đáp án
--! =================================================================
CREATE TABLE IF NOT EXISTS reading_choice_multi_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_question_id UUID NOT NULL REFERENCES reading_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_reading_choice_multi_question UNIQUE (reading_question_id)
);

CREATE TABLE IF NOT EXISTS reading_choice_multi_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reading_choice_multi_question_id UUID NOT NULL 
        REFERENCES reading_choice_multi_questions(id) ON DELETE CASCADE,
    options TEXT NOT NULL CHECK (length(trim(options)) > 0),
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_multi_option_per_question 
        UNIQUE (reading_choice_multi_question_id, options)
);

-- Index tối ưu tìm kiếm các lựa chọn theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_reading_choice_multi_options 
ON reading_choice_multi_options(reading_choice_multi_question_id, options);

-- Index tối ưu tìm kiếm các đáp án đúng
CREATE INDEX IF NOT EXISTS idx_reading_choice_multi_correct_options 
ON reading_choice_multi_options(reading_choice_multi_question_id) 
WHERE is_correct = TRUE;

--! =================================================================
--! TRIGGERS - Create triggers for version tracking
--! =================================================================

-- Triggers for reading_questions
CREATE TRIGGER trigger_reading_questions_version_insert
AFTER INSERT ON reading_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_questions_version_update
AFTER UPDATE ON reading_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_questions_version_delete  
AFTER DELETE ON reading_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for true/false questions
CREATE TRIGGER trigger_reading_true_false_version_insert
AFTER INSERT ON reading_true_falses
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_true_false_version_update
AFTER UPDATE ON reading_true_falses
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_true_false_version_delete
AFTER DELETE ON reading_true_falses
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for fill in blank questions
CREATE TRIGGER trigger_reading_fill_in_the_blank_questions_version_insert
AFTER INSERT ON reading_fill_in_the_blank_questions
FOR EACH ROW 
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_fill_in_the_blank_questions_version_update
AFTER UPDATE ON reading_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_fill_in_the_blank_questions_version_delete
AFTER DELETE ON reading_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for fill in blank answers
CREATE TRIGGER trigger_reading_fill_in_the_blank_answers_version_insert
AFTER INSERT ON reading_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_fill_in_the_blank_answers_version_update
AFTER UPDATE ON reading_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_fill_in_the_blank_answers_version_delete
AFTER DELETE ON reading_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for choice one questions
CREATE TRIGGER trigger_reading_choice_one_questions_version_insert
AFTER INSERT ON reading_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_one_questions_version_update
AFTER UPDATE ON reading_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_one_questions_version_delete
AFTER DELETE ON reading_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for choice one options
CREATE TRIGGER trigger_reading_choice_one_options_version_insert
AFTER INSERT ON reading_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_one_options_version_update
AFTER UPDATE ON reading_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_one_options_version_delete
AFTER DELETE ON reading_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for choice multi questions
CREATE TRIGGER trigger_reading_choice_multi_questions_version_insert
AFTER INSERT ON reading_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_multi_questions_version_update
AFTER UPDATE ON reading_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_multi_questions_version_delete
AFTER DELETE ON reading_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for choice multi options
CREATE TRIGGER trigger_reading_choice_multi_options_version_insert
AFTER INSERT ON reading_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_multi_options_version_update
AFTER UPDATE ON reading_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_choice_multi_options_version_delete
AFTER DELETE ON reading_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for matchings
CREATE TRIGGER trigger_reading_matchings_version_insert
AFTER INSERT ON reading_matchings
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_matchings_version_update
AFTER UPDATE ON reading_matchings
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

CREATE TRIGGER trigger_reading_matchings_version_delete
AFTER DELETE ON reading_matchings
FOR EACH ROW
EXECUTE FUNCTION reading_question_version_update();

-- Triggers for updated_at timestamp
CREATE TRIGGER update_reading_questions_updated_at
    BEFORE UPDATE ON reading_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_true_falses_updated_at
    BEFORE UPDATE ON reading_true_falses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_fill_in_the_blank_questions_updated_at
    BEFORE UPDATE ON reading_fill_in_the_blank_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_fill_in_the_blank_answers_updated_at
    BEFORE UPDATE ON reading_fill_in_the_blank_answers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_choice_one_questions_updated_at
    BEFORE UPDATE ON reading_choice_one_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_choice_one_options_updated_at
    BEFORE UPDATE ON reading_choice_one_options
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_choice_multi_questions_updated_at
    BEFORE UPDATE ON reading_choice_multi_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_choice_multi_options_updated_at
    BEFORE UPDATE ON reading_choice_multi_options
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_reading_matchings_updated_at
    BEFORE UPDATE ON reading_matchings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();