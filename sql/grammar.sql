-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

--! =================================================================
--! FUNCTIONS
--! =================================================================
-- Function tự động cập nhật timestamp 
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function quản lý version cho grammar questions
CREATE OR REPLACE FUNCTION grammar_question_version_update()
RETURNS TRIGGER AS $$
BEGIN
IF TG_OP IN ('INSERT', 'UPDATE', 'DELETE') THEN
        IF TG_TABLE_NAME = 'grammar_questions' THEN
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
                    IF TG_TABLE_NAME IN ('grammar_fill_in_the_blank_questions',
                                       'grammar_choice_one_questions',
                                       'grammar_error_identifications',
                                       'grammar_sentence_transformations') THEN
                        parent_id := OLD.grammar_question_id;
                        UPDATE grammar_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = parent_id;
                        RETURN OLD;
                    END IF;

                    -- For nested child tables
                    IF TG_TABLE_NAME = 'grammar_fill_in_the_blank_answers' THEN
                        UPDATE grammar_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id IN (
                            SELECT q.grammar_question_id 
                            FROM grammar_fill_in_the_blank_questions q
                            WHERE q.id = OLD.grammar_fill_in_the_blank_question_id
                        );
                        RETURN OLD;
                    ELSIF TG_TABLE_NAME = 'grammar_choice_one_options' THEN
                        UPDATE grammar_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id IN (
                            SELECT q.grammar_question_id
                            FROM grammar_choice_one_questions q
                            WHERE q.id = OLD.grammar_choice_one_question_id
                        );
                        RETURN OLD;
                    END IF;
                ELSE
                    record_id := NEW.id;
                END IF;

                -- Handle other tables through joins
                UPDATE grammar_questions
                SET version = version + 1,
                    updated_at = CURRENT_TIMESTAMP
                WHERE id IN (
                    SELECT q.grammar_question_id 
                    FROM grammar_fill_in_the_blank_questions q
                    INNER JOIN grammar_fill_in_the_blank_answers a 
                    ON a.grammar_fill_in_the_blank_question_id = q.id
                    WHERE a.id = record_id
                    
                    UNION
                    SELECT q.grammar_question_id
                    FROM grammar_choice_one_questions q
                    INNER JOIN grammar_choice_one_options o
                    ON o.grammar_choice_one_question_id = q.id
                    WHERE o.id = record_id

                    UNION
                    SELECT grammar_question_id FROM grammar_fill_in_the_blank_questions WHERE id = record_id
                    UNION  
                    SELECT grammar_question_id FROM grammar_choice_one_questions WHERE id = record_id
                    UNION  
                    SELECT grammar_question_id FROM grammar_error_identifications WHERE id = record_id
                    UNION
                    SELECT grammar_question_id FROM grammar_sentence_transformations WHERE id = record_id
                );
            END;
        END IF;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

--! =================================================================
--! GRAMMAR QUESTIONS - Base table
--! =================================================================
CREATE TABLE IF NOT EXISTS grammar_questions (
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

-- Index tối ưu tìm kiếm theo loại câu hỏi
CREATE INDEX IF NOT EXISTS idx_grammar_questions_type 
ON grammar_questions(type);

-- Index tối ưu tìm kiếm theo chủ đề
CREATE INDEX IF NOT EXISTS idx_grammar_questions_topic 
ON grammar_questions USING GIN(topic);

--! =================================================================
--! FILL IN THE BLANK
--! =================================================================
CREATE TABLE IF NOT EXISTS grammar_fill_in_the_blank_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_question_id UUID NOT NULL REFERENCES grammar_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_grammar_fill_in_the_blank_question UNIQUE (grammar_question_id)
);

CREATE TABLE IF NOT EXISTS grammar_fill_in_the_blank_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_fill_in_the_blank_question_id UUID NOT NULL 
        REFERENCES grammar_fill_in_the_blank_questions(id) ON DELETE CASCADE,
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu tìm kiếm câu hỏi theo grammar_question_id
CREATE INDEX IF NOT EXISTS idx_grammar_fill_blank_question_id 
ON grammar_fill_in_the_blank_questions(grammar_question_id);

-- Index tối ưu tìm kiếm đáp án theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_grammar_fill_blank_answers_question_id 
ON grammar_fill_in_the_blank_answers(grammar_fill_in_the_blank_question_id);

--! =================================================================
--! CHOICE ONE
--! =================================================================
CREATE TABLE IF NOT EXISTS grammar_choice_one_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_question_id UUID NOT NULL REFERENCES grammar_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_grammar_choice_one_question UNIQUE (grammar_question_id)
);

CREATE TABLE IF NOT EXISTS grammar_choice_one_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_choice_one_question_id UUID NOT NULL 
        REFERENCES grammar_choice_one_questions(id) ON DELETE CASCADE,
    options TEXT NOT NULL CHECK (length(trim(options)) > 0),
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_choice_one_option_per_question_idx 
        UNIQUE (grammar_choice_one_question_id, options)
);

-- Create a partial unique index for correct options
CREATE UNIQUE INDEX IF NOT EXISTS idx_grammar_one_correct_option_per_question 
ON grammar_choice_one_options (grammar_choice_one_question_id) 
WHERE is_correct = TRUE;

-- Index tối ưu tìm kiếm các lựa chọn theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_grammar_choice_one_options 
ON grammar_choice_one_options(grammar_choice_one_question_id, options);

-- Index tối ưu tìm kiếm đáp án đúng
CREATE INDEX IF NOT EXISTS idx_grammar_choice_one_correct_options 
ON grammar_choice_one_options(grammar_choice_one_question_id) 
WHERE is_correct = TRUE;

--! =================================================================
--! ERROR IDENTIFICATION
--! =================================================================
CREATE TABLE IF NOT EXISTS grammar_error_identifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_question_id UUID NOT NULL REFERENCES grammar_questions(id) ON DELETE CASCADE,
    error_sentence TEXT NOT NULL CHECK (length(trim(error_sentence)) > 0),
    error_word TEXT NOT NULL CHECK (length(trim(error_word)) > 0),
    correct_word TEXT NOT NULL CHECK (length(trim(correct_word)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_error_identification_per_question UNIQUE (grammar_question_id)
);

-- Index tối ưu cho Error Identification
CREATE INDEX IF NOT EXISTS idx_grammar_error_identification_question_id
ON grammar_error_identifications(grammar_question_id);

CREATE INDEX IF NOT EXISTS idx_grammar_error_identification_sentence
ON grammar_error_identifications USING gin(to_tsvector('english', error_sentence));

--! =================================================================
--! SENTENCE TRANSFORMATION
--! =================================================================
CREATE TABLE IF NOT EXISTS grammar_sentence_transformations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grammar_question_id UUID NOT NULL REFERENCES grammar_questions(id) ON DELETE CASCADE,
    original_sentence TEXT NOT NULL CHECK (length(trim(original_sentence)) > 0),
    beginning_word TEXT,
    example_correct_sentence TEXT NOT NULL CHECK (length(trim(example_correct_sentence)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_sentence_transformation_per_question UNIQUE (grammar_question_id)
);

-- Index tối ưu cho Sentence Transformation
CREATE INDEX IF NOT EXISTS idx_grammar_sentence_transformation_question_id
ON grammar_sentence_transformations(grammar_question_id);

CREATE INDEX IF NOT EXISTS idx_grammar_sentence_transformation_original
ON grammar_sentence_transformations USING gin(to_tsvector('english', original_sentence));

--! =================================================================
--! TRIGGERS - Create triggers for version tracking
--! =================================================================

-- Triggers for grammar_questions
CREATE TRIGGER trigger_grammar_questions_version_insert
AFTER INSERT ON grammar_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_questions_version_update
AFTER UPDATE ON grammar_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_questions_version_delete  
AFTER DELETE ON grammar_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for fill in blank questions
CREATE TRIGGER trigger_grammar_fill_in_the_blank_questions_version_insert
AFTER INSERT ON grammar_fill_in_the_blank_questions
FOR EACH ROW 
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_fill_in_the_blank_questions_version_update
AFTER UPDATE ON grammar_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_fill_in_the_blank_questions_version_delete
AFTER DELETE ON grammar_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for fill in blank answers
CREATE TRIGGER trigger_grammar_fill_in_the_blank_answers_version_insert
AFTER INSERT ON grammar_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_fill_in_the_blank_answers_version_update
AFTER UPDATE ON grammar_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_fill_in_the_blank_answers_version_delete
AFTER DELETE ON grammar_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for choice one questions
CREATE TRIGGER trigger_grammar_choice_one_questions_version_insert
AFTER INSERT ON grammar_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_choice_one_questions_version_update
AFTER UPDATE ON grammar_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_choice_one_questions_version_delete
AFTER DELETE ON grammar_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for choice one options
CREATE TRIGGER trigger_grammar_choice_one_options_version_insert
AFTER INSERT ON grammar_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_choice_one_options_version_update
AFTER UPDATE ON grammar_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_choice_one_options_version_delete
AFTER DELETE ON grammar_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for error identifications
CREATE TRIGGER trigger_grammar_error_identifications_version_insert
AFTER INSERT ON grammar_error_identifications
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_error_identifications_version_update
AFTER UPDATE ON grammar_error_identifications
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_error_identifications_version_delete
AFTER DELETE ON grammar_error_identifications
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for sentence transformations
CREATE TRIGGER trigger_grammar_sentence_transformations_version_insert
AFTER INSERT ON grammar_sentence_transformations
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_sentence_transformations_version_update
AFTER UPDATE ON grammar_sentence_transformations
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

CREATE TRIGGER trigger_grammar_sentence_transformations_version_delete
AFTER DELETE ON grammar_sentence_transformations
FOR EACH ROW
EXECUTE FUNCTION grammar_question_version_update();

-- Triggers for updated_at timestamp
CREATE TRIGGER update_grammar_questions_updated_at
    BEFORE UPDATE ON grammar_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_fill_in_the_blank_questions_updated_at
    BEFORE UPDATE ON grammar_fill_in_the_blank_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_fill_in_the_blank_answers_updated_at
    BEFORE UPDATE ON grammar_fill_in_the_blank_answers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_choice_one_questions_updated_at
    BEFORE UPDATE ON grammar_choice_one_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_choice_one_options_updated_at
    BEFORE UPDATE ON grammar_choice_one_options
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_error_identifications_updated_at
    BEFORE UPDATE ON grammar_error_identifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_grammar_sentence_transformations_updated_at
    BEFORE UPDATE ON grammar_sentence_transformations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--! =================================================================
--! COMMENTS - Giải thích các bảng
--! =================================================================
COMMENT ON TABLE grammar_questions 
IS 'Bảng gốc chứa thông tin chung của tất cả loại câu hỏi grammar';

COMMENT ON TABLE grammar_fill_in_the_blank_questions 
IS 'Bảng chứa các câu hỏi dạng điền vào chỗ trống';

COMMENT ON TABLE grammar_choice_one_questions 
IS 'Bảng chứa các câu hỏi dạng chọn một đáp án';

COMMENT ON TABLE grammar_error_identifications
IS 'Bảng chứa các câu hỏi dạng nhận diện lỗi sai';

COMMENT ON TABLE grammar_sentence_transformations
IS 'Bảng chứa các câu hỏi dạng biến đổi câu';

COMMENT ON COLUMN grammar_questions.version 
IS 'Version number that auto-increments when the question or its related data is modified';

COMMENT ON COLUMN grammar_questions.id
IS 'Primary key using UUID with auto-generation';