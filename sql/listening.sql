-- Drop existing objects if needed
DO $$
BEGIN
    DROP FUNCTION IF EXISTS listening_question_version_update() CASCADE;
EXCEPTION
    WHEN OTHERS THEN NULL;
END $$;

-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

--! =================================================================
--! FUNCTIONS
--! =================================================================
-- Function tự động cập nhật timestamp - check if exists
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

-- Function quản lý version cho listening questions
DO $outer$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_proc WHERE proname = 'listening_question_version_update') THEN
        CREATE OR REPLACE FUNCTION listening_question_version_update()
        RETURNS TRIGGER AS $body$
        BEGIN
            IF TG_OP IN ('INSERT', 'UPDATE', 'DELETE') THEN
                IF TG_TABLE_NAME = 'listening_questions' THEN
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
                            IF TG_TABLE_NAME IN ('listening_fill_in_the_blank_questions',
                                               'listening_choice_one_questions',
                                               'listening_choice_multi_questions',
                                               'listening_map_labellings',
                                               'listening_matchings') THEN
                                parent_id := OLD.listening_question_id;
                                UPDATE listening_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id = parent_id;
                                RETURN OLD;
                            END IF;
                            -- For nested child tables
                            IF TG_TABLE_NAME = 'listening_fill_in_the_blank_answers' THEN
                                UPDATE listening_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.listening_question_id
                                    FROM listening_fill_in_the_blank_questions q
                                    WHERE q.id = OLD.listening_fill_in_the_blank_question_id
                                );
                                RETURN OLD;
                            ELSIF TG_TABLE_NAME = 'listening_choice_one_options' THEN
                                UPDATE listening_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.listening_question_id
                                    FROM listening_choice_one_questions q
                                    WHERE q.id = OLD.listening_choice_one_question_id
                                );
                                RETURN OLD;
                            ELSIF TG_TABLE_NAME = 'listening_choice_multi_options' THEN
                                UPDATE listening_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id IN (
                                    SELECT q.listening_question_id
                                    FROM listening_choice_multi_questions q
                                    WHERE q.id = OLD.listening_choice_multi_question_id
                                );
                                RETURN OLD;
                            END IF;
                        ELSE
                            record_id := NEW.id;
                            -- Handle direct child tables
                            IF TG_TABLE_NAME IN ('listening_map_labellings', 'listening_matchings') THEN
                                parent_id := NEW.listening_question_id;
                                UPDATE listening_questions
                                SET version = version + 1,
                                    updated_at = CURRENT_TIMESTAMP
                                WHERE id = parent_id;
                                RETURN NEW;
                            END IF;
                        END IF;
                        -- Handle other tables through joins
                        UPDATE listening_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id IN (
                            SELECT q.listening_question_id
                            FROM listening_fill_in_the_blank_questions q
                            INNER JOIN listening_fill_in_the_blank_answers a
                            ON a.listening_fill_in_the_blank_question_id = q.id
                            WHERE a.id = record_id
                            UNION
                            SELECT q.listening_question_id
                            FROM listening_choice_one_questions q
                            INNER JOIN listening_choice_one_options o
                            ON o.listening_choice_one_question_id = q.id
                            WHERE o.id = record_id
                            UNION
                            SELECT q.listening_question_id
                            FROM listening_choice_multi_questions q
                            INNER JOIN listening_choice_multi_options o
                            ON o.listening_choice_multi_question_id = q.id
                            WHERE o.id = record_id
                            UNION
                            SELECT listening_question_id FROM listening_fill_in_the_blank_questions WHERE id = record_id
                            UNION
                            SELECT listening_question_id FROM listening_choice_one_questions WHERE id = record_id
                            UNION
                            SELECT listening_question_id FROM listening_choice_multi_questions WHERE id = record_id
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
--! LISTENING QUESTIONS - Bảng câu hỏi gốc
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    topic VARCHAR(100)[] NOT NULL,
    instruction TEXT NOT NULL CHECK (length(trim(instruction)) > 0),
    audio_urls TEXT[] NOT NULL,
    image_urls TEXT[],
    transcript TEXT NOT NULL CHECK (length(trim(transcript)) > 0),
    max_time INT NOT NULL CHECK (max_time > 0),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu tìm kiếm theo loại câu hỏi
CREATE INDEX IF NOT EXISTS idx_listening_questions_type 
ON listening_questions(type);

-- Index tối ưu tìm kiếm theo chủ đề
CREATE INDEX IF NOT EXISTS idx_listening_questions_topic 
ON listening_questions USING GIN(topic);

--! =================================================================
--! FILL IN THE BLANK - Điền vào chỗ trống
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_fill_in_the_blank_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_question_id UUID NOT NULL REFERENCES listening_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_listening_fill_in_the_blank_question UNIQUE (listening_question_id)
);

CREATE TABLE IF NOT EXISTS listening_fill_in_the_blank_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_fill_in_the_blank_question_id UUID NOT NULL 
        REFERENCES listening_fill_in_the_blank_questions(id) ON DELETE CASCADE,
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu tìm kiếm câu hỏi theo listening_question_id
CREATE INDEX IF NOT EXISTS idx_listening_fill_blank_question_id 
ON listening_fill_in_the_blank_questions(listening_question_id);

-- Index tối ưu tìm kiếm đáp án theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_listening_fill_blank_answers_question_id 
ON listening_fill_in_the_blank_answers(listening_fill_in_the_blank_question_id);

--! =================================================================
--! CHOICE ONE - Chọn một đáp án    
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_choice_one_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_question_id UUID NOT NULL REFERENCES listening_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_listening_choice_one_question UNIQUE (listening_question_id)
);

CREATE TABLE IF NOT EXISTS listening_choice_one_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_choice_one_question_id UUID NOT NULL 
        REFERENCES listening_choice_one_questions(id) ON DELETE CASCADE,
    options TEXT NOT NULL CHECK (length(trim(options)) > 0),
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT unique_choice_one_option_per_question 
    UNIQUE (listening_choice_one_question_id, options)
);

-- Create a partial unique index instead of the WHERE constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_listening_one_correct_option_per_question 
ON listening_choice_one_options (listening_choice_one_question_id) 
WHERE is_correct = TRUE;

-- Index tối ưu tìm kiếm các lựa chọn theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_listening_choice_one_options 
ON listening_choice_one_options(listening_choice_one_question_id, options);

-- Index tối ưu tìm kiếm đáp án đúng
CREATE INDEX IF NOT EXISTS idx_listening_choice_one_correct_options 
ON listening_choice_one_options(listening_choice_one_question_id) 
WHERE is_correct = TRUE;

--! =================================================================
--! CHOICE MULTI - Chọn nhiều đáp án
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_choice_multi_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_question_id UUID NOT NULL REFERENCES listening_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS listening_choice_multi_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_choice_multi_question_id UUID NOT NULL 
        REFERENCES listening_choice_multi_questions(id) ON DELETE CASCADE,
    options TEXT NOT NULL CHECK (length(trim(options)) > 0),
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT unique_choice_multi_option_per_question
    UNIQUE (listening_choice_multi_question_id, options)
);

-- Index đảm bảo mỗi listening_question chỉ có 1 câu hỏi chọn nhiều
CREATE UNIQUE INDEX IF NOT EXISTS idx_listening_choice_multi_question 
ON listening_choice_multi_questions(listening_question_id);

-- Index tối ưu tìm kiếm các lựa chọn theo câu hỏi
CREATE INDEX IF NOT EXISTS idx_listening_choice_multi_options 
ON listening_choice_multi_options(listening_choice_multi_question_id, options);

-- Index tối ưu tìm kiếm các đáp án đúng
CREATE INDEX IF NOT EXISTS idx_listening_choice_multi_correct_options 
ON listening_choice_multi_options(listening_choice_multi_question_id) 
WHERE is_correct = TRUE;

--! =================================================================
--! MAP LABELLING - Bài tập gán nhãn trên bản đồ
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_map_labellings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_question_id UUID NOT NULL REFERENCES listening_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_question_per_labelling UNIQUE (listening_question_id, question),
    CONSTRAINT unique_answer_per_labelling UNIQUE (listening_question_id, answer)
);

-- Index tối ưu cho Map Labelling
CREATE INDEX IF NOT EXISTS idx_listening_map_labelling_question_id 
ON listening_map_labellings(listening_question_id);

CREATE INDEX IF NOT EXISTS idx_listening_map_labelling_question 
ON listening_map_labellings USING gin(to_tsvector('english', question));

CREATE INDEX IF NOT EXISTS idx_listening_map_labelling_answer 
ON listening_map_labellings USING gin(to_tsvector('english', answer));

--! =================================================================
--! MATCHING - Bài tập nối câu
--! =================================================================
CREATE TABLE IF NOT EXISTS listening_matchings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listening_question_id UUID NOT NULL REFERENCES listening_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_listening_question_per_matching UNIQUE (listening_question_id, question),
    CONSTRAINT unique_listening_answer_per_matching UNIQUE (listening_question_id, answer)
);

-- Index tối ưu cho Matching
CREATE INDEX IF NOT EXISTS idx_listening_matching_question_id 
ON listening_matchings(listening_question_id);

CREATE INDEX IF NOT EXISTS idx_listening_matching_question 
ON listening_matchings USING gin(to_tsvector('english', question));

CREATE INDEX IF NOT EXISTS idx_listening_matching_answer 
ON listening_matchings USING gin(to_tsvector('english', answer));

--! =================================================================
--! TRIGGERS - Create triggers for version tracking
--! =================================================================

-- Triggers for listening_questions
CREATE TRIGGER trigger_listening_questions_version_insert
AFTER INSERT ON listening_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_questions_version_update
AFTER UPDATE ON listening_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_questions_version_delete  
AFTER DELETE ON listening_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for fill in blank questions
CREATE TRIGGER trigger_listening_fill_in_the_blank_questions_version_insert
AFTER INSERT ON listening_fill_in_the_blank_questions
FOR EACH ROW 
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_fill_in_the_blank_questions_version_update
AFTER UPDATE ON listening_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_fill_in_the_blank_questions_version_delete
AFTER DELETE ON listening_fill_in_the_blank_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for fill in blank answers
CREATE TRIGGER trigger_listening_fill_in_the_blank_answers_version_insert
AFTER INSERT ON listening_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_fill_in_the_blank_answers_version_update
AFTER UPDATE ON listening_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_fill_in_the_blank_answers_version_delete
AFTER DELETE ON listening_fill_in_the_blank_answers
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for choice one questions
CREATE TRIGGER trigger_listening_choice_one_questions_version_insert
AFTER INSERT ON listening_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_one_questions_version_update
AFTER UPDATE ON listening_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_one_questions_version_delete
AFTER DELETE ON listening_choice_one_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for choice one options
CREATE TRIGGER trigger_listening_choice_one_options_version_insert
AFTER INSERT ON listening_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_one_options_version_update
AFTER UPDATE ON listening_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_one_options_version_delete
AFTER DELETE ON listening_choice_one_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for choice multi questions
CREATE TRIGGER trigger_listening_choice_multi_questions_version_insert
AFTER INSERT ON listening_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_multi_questions_version_update
AFTER UPDATE ON listening_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_multi_questions_version_delete
AFTER DELETE ON listening_choice_multi_questions
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for choice multi options
CREATE TRIGGER trigger_listening_choice_multi_options_version_insert
AFTER INSERT ON listening_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_multi_options_version_update
AFTER UPDATE ON listening_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_choice_multi_options_version_delete
AFTER DELETE ON listening_choice_multi_options
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for map labellings
CREATE TRIGGER trigger_listening_map_labellings_version_insert
AFTER INSERT ON listening_map_labellings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_map_labellings_version_update
AFTER UPDATE ON listening_map_labellings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_map_labellings_version_delete
AFTER DELETE ON listening_map_labellings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

-- Triggers for matchings
CREATE TRIGGER trigger_listening_matchings_version_insert
AFTER INSERT ON listening_matchings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_matchings_version_update
AFTER UPDATE ON listening_matchings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();

CREATE TRIGGER trigger_listening_matchings_version_delete
AFTER DELETE ON listening_matchings
FOR EACH ROW
EXECUTE FUNCTION listening_question_version_update();