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

-- Function quản lý version cho speaking questions
CREATE OR REPLACE FUNCTION speaking_question_version_update()
RETURNS TRIGGER AS $$   
BEGIN
IF TG_OP IN ('INSERT', 'UPDATE', 'DELETE') THEN
        IF TG_TABLE_NAME = 'speaking_questions' THEN
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
                    IF TG_TABLE_NAME IN ('speaking_word_repetitions',
                                       'speaking_phrase_repetitions', 
                                       'speaking_paragraph_repetitions',
                                       'speaking_open_paragraphs') THEN
                        parent_id := OLD.speaking_question_id;
                        UPDATE speaking_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = parent_id;
                        RETURN OLD;
                    END IF;

                    -- For nested child tables
                    IF TG_TABLE_NAME = 'speaking_conversational_repetition_qas' THEN
                        UPDATE speaking_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id IN (
                            SELECT q.speaking_question_id 
                            FROM speaking_conversational_repetitions q
                            WHERE q.id = OLD.speaking_conversational_repetition_id
                        );
                        RETURN OLD;
                    END IF;
                ELSE
                    record_id := NEW.id;
                    -- Handle direct child tables
                    IF TG_TABLE_NAME IN ('speaking_word_repetitions',
                                       'speaking_phrase_repetitions',
                                       'speaking_paragraph_repetitions',
                                       'speaking_open_paragraphs') THEN
                        parent_id := NEW.speaking_question_id;
                        UPDATE speaking_questions
                        SET version = version + 1,
                            updated_at = CURRENT_TIMESTAMP
                        WHERE id = parent_id;
                        RETURN NEW;
                    END IF;
                END IF;

                -- Handle other tables through joins
                UPDATE speaking_questions
                SET version = version + 1,
                    updated_at = CURRENT_TIMESTAMP
                WHERE id IN (
                    SELECT q.speaking_question_id 
                    FROM speaking_conversational_repetitions q
                    INNER JOIN speaking_conversational_repetition_qas a 
                    ON a.speaking_conversational_repetition_id = q.id
                    WHERE a.id = record_id
                );
            END;
        END IF;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

--! =================================================================
--! SPEAKING QUESTIONS - Bảng câu hỏi gốc
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_questions (
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
CREATE INDEX IF NOT EXISTS idx_speaking_questions_type 
ON speaking_questions(type);

-- Index tối ưu tìm kiếm theo chủ đề
CREATE INDEX IF NOT EXISTS idx_speaking_questions_topic 
ON speaking_questions USING GIN(topic);

--! =================================================================
--! WORD REPETITION - Lặp lại từ
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_word_repetitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    word TEXT NOT NULL CHECK (length(trim(word)) > 0),
    mean TEXT NOT NULL CHECK (length(trim(mean)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_word_per_question UNIQUE (speaking_question_id, word)
);

-- Index tối ưu cho Word Repetition
CREATE INDEX IF NOT EXISTS idx_speaking_word_repetition_question_id 
ON speaking_word_repetitions(speaking_question_id);

CREATE INDEX IF NOT EXISTS idx_speaking_word_repetition_word 
ON speaking_word_repetitions USING gin(to_tsvector('english', word));

--! =================================================================
--! PHRASE REPETITION - Lặp lại cụm từ
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_phrase_repetitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    phrase TEXT NOT NULL CHECK (length(trim(phrase)) > 0),
    mean TEXT NOT NULL CHECK (length(trim(mean)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_phrase_per_question UNIQUE (speaking_question_id, phrase)
);

-- Index tối ưu cho Phrase Repetition
CREATE INDEX IF NOT EXISTS idx_speaking_phrase_repetition_question_id 
ON speaking_phrase_repetitions(speaking_question_id);

CREATE INDEX IF NOT EXISTS idx_speaking_phrase_repetition_phrase 
ON speaking_phrase_repetitions USING gin(to_tsvector('english', phrase));

--! =================================================================
--! PARAGRAPH REPETITION - Lặp lại đoạn văn
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_paragraph_repetitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    paragraph TEXT NOT NULL CHECK (length(trim(paragraph)) > 0),
    mean TEXT NOT NULL CHECK (length(trim(mean)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_paragraph_per_question UNIQUE (speaking_question_id)
);

-- Index tối ưu cho Paragraph Repetition
CREATE INDEX IF NOT EXISTS idx_speaking_paragraph_repetition_question_id 
ON speaking_paragraph_repetitions(speaking_question_id);

CREATE INDEX IF NOT EXISTS idx_speaking_paragraph_repetition_text 
ON speaking_paragraph_repetitions USING gin(to_tsvector('english', paragraph));

--! =================================================================
--! OPEN PARAGRAPH - Đoạn văn tự do
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_open_paragraphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    example_passage TEXT NOT NULL CHECK (length(trim(example_passage)) > 0),
    mean_of_example_passage TEXT NOT NULL CHECK (length(trim(mean_of_example_passage)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_open_paragraph_per_question UNIQUE (speaking_question_id)
);

-- Index tối ưu cho Open Paragraph
CREATE INDEX IF NOT EXISTS idx_speaking_open_paragraph_question_id 
ON speaking_open_paragraphs(speaking_question_id);

CREATE INDEX IF NOT EXISTS idx_speaking_open_paragraph_text 
ON speaking_open_paragraphs USING gin(to_tsvector('english', question));

--! =================================================================
--! CONVERSATIONAL REPETITION - Hội thoại có sẵn
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_conversational_repetitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    overview TEXT NOT NULL CHECK (length(trim(overview)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_conversational_repetition_per_question UNIQUE (speaking_question_id)
);

CREATE TABLE IF NOT EXISTS speaking_conversational_repetition_qas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_conversational_repetition_id UUID NOT NULL 
        REFERENCES speaking_conversational_repetitions(id) ON DELETE CASCADE,
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    answer TEXT NOT NULL CHECK (length(trim(answer)) > 0),
    mean_of_question TEXT NOT NULL CHECK (length(trim(mean_of_question)) > 0),
    mean_of_answer TEXT NOT NULL CHECK (length(trim(mean_of_answer)) > 0),
    explain TEXT NOT NULL CHECK (length(trim(explain)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index tối ưu cho Conversational Repetition
CREATE INDEX IF NOT EXISTS idx_speaking_conv_repetition_question_id 
ON speaking_conversational_repetitions(speaking_question_id);

CREATE INDEX IF NOT EXISTS idx_speaking_conv_repetition_qas 
ON speaking_conversational_repetition_qas(speaking_conversational_repetition_id);

CREATE INDEX IF NOT EXISTS idx_speaking_conv_repetition_qa_text 
ON speaking_conversational_repetition_qas USING gin(to_tsvector('english', question));

--! =================================================================
--! CONVERSATIONAL OPEN - Hội thoại tự do
--! =================================================================
CREATE TABLE IF NOT EXISTS speaking_conversational_opens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    speaking_question_id UUID NOT NULL REFERENCES speaking_questions(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    overview TEXT NOT NULL CHECK (length(trim(overview)) > 0),
    example_conversation TEXT NOT NULL CHECK (length(trim(example_conversation)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_conversational_open_per_question UNIQUE (speaking_question_id)
);

-- Index tối ưu cho Conversational Open
CREATE INDEX IF NOT EXISTS idx_speaking_conv_open_question_id 
ON speaking_conversational_opens(speaking_question_id);

--! =================================================================
--! TRIGGERS - Create triggers for version tracking
--! =================================================================

-- Triggers for speaking_questions
CREATE TRIGGER trigger_speaking_questions_version_insert
AFTER INSERT ON speaking_questions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_questions_version_update
AFTER UPDATE ON speaking_questions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_questions_version_delete  
AFTER DELETE ON speaking_questions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for word repetitions
CREATE TRIGGER trigger_speaking_word_repetitions_version_insert
AFTER INSERT ON speaking_word_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_word_repetitions_version_update
AFTER UPDATE ON speaking_word_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_word_repetitions_version_delete
AFTER DELETE ON speaking_word_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for phrase repetitions
CREATE TRIGGER trigger_speaking_phrase_repetitions_version_insert
AFTER INSERT ON speaking_phrase_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_phrase_repetitions_version_update
AFTER UPDATE ON speaking_phrase_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_phrase_repetitions_version_delete
AFTER DELETE ON speaking_phrase_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for paragraph repetitions
CREATE TRIGGER trigger_speaking_paragraph_repetitions_version_insert
AFTER INSERT ON speaking_paragraph_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_paragraph_repetitions_version_update
AFTER UPDATE ON speaking_paragraph_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_paragraph_repetitions_version_delete
AFTER DELETE ON speaking_paragraph_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for open paragraphs
CREATE TRIGGER trigger_speaking_open_paragraphs_version_insert
AFTER INSERT ON speaking_open_paragraphs
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_open_paragraphs_version_update
AFTER UPDATE ON speaking_open_paragraphs
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_open_paragraphs_version_delete
AFTER DELETE ON speaking_open_paragraphs
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for conversational repetitions
CREATE TRIGGER trigger_speaking_conv_repetitions_version_insert
AFTER INSERT ON speaking_conversational_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_repetitions_version_update
AFTER UPDATE ON speaking_conversational_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_repetitions_version_delete
AFTER DELETE ON speaking_conversational_repetitions
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for conversational repetition QAs
CREATE TRIGGER trigger_speaking_conv_repetition_qas_version_insert
AFTER INSERT ON speaking_conversational_repetition_qas
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_repetition_qas_version_update
AFTER UPDATE ON speaking_conversational_repetition_qas
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_repetition_qas_version_delete
AFTER DELETE ON speaking_conversational_repetition_qas
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for conversational opens
CREATE TRIGGER trigger_speaking_conv_opens_version_insert
AFTER INSERT ON speaking_conversational_opens
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_opens_version_update
AFTER UPDATE ON speaking_conversational_opens
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

CREATE TRIGGER trigger_speaking_conv_opens_version_delete
AFTER DELETE ON speaking_conversational_opens
FOR EACH ROW
EXECUTE FUNCTION speaking_question_version_update();

-- Triggers for updated_at timestamp
CREATE TRIGGER update_speaking_questions_updated_at
    BEFORE UPDATE ON speaking_questions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_word_repetitions_updated_at
    BEFORE UPDATE ON speaking_word_repetitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_phrase_repetitions_updated_at
    BEFORE UPDATE ON speaking_phrase_repetitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_paragraph_repetitions_updated_at
    BEFORE UPDATE ON speaking_paragraph_repetitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_open_paragraphs_updated_at
    BEFORE UPDATE ON speaking_open_paragraphs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_conv_repetitions_updated_at
    BEFORE UPDATE ON speaking_conversational_repetitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_conv_repetition_qas_updated_at
    BEFORE UPDATE ON speaking_conversational_repetition_qas
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_speaking_conv_opens_updated_at
    BEFORE UPDATE ON speaking_conversational_opens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--! =================================================================
--! COMMENTS - Giải thích các bảng
--! =================================================================
COMMENT ON TABLE speaking_questions 
IS 'Bảng gốc chứa thông tin chung của tất cả loại câu hỏi speaking';

COMMENT ON TABLE speaking_word_repetitions
IS 'Bảng chứa các câu hỏi dạng lặp lại từ';

COMMENT ON TABLE speaking_phrase_repetitions
IS 'Bảng chứa các câu hỏi dạng lặp lại cụm từ';

COMMENT ON TABLE speaking_paragraph_repetitions
IS 'Bảng chứa các câu hỏi dạng lặp lại đoạn văn';

COMMENT ON TABLE speaking_open_paragraphs
IS 'Bảng chứa các câu hỏi dạng đoạn văn tự do';

COMMENT ON TABLE speaking_conversational_repetitions
IS 'Bảng chứa các câu hỏi dạng hội thoại có sẵn';

COMMENT ON TABLE speaking_conversational_opens
IS 'Bảng chứa các câu hỏi dạng hội thoại tự do';

COMMENT ON COLUMN speaking_questions.version 
IS 'Version number that auto-increments when the question or its related data is modified';

COMMENT ON COLUMN speaking_questions.id
IS 'New primary key using UUID with auto-generation';

COMMENT ON COLUMN speaking_questions.type 
IS 'Loại câu hỏi speaking';

COMMENT ON COLUMN speaking_questions.topic 
IS 'Các chủ đề của câu hỏi';

COMMENT ON COLUMN speaking_questions.instruction 
IS 'Hướng dẫn làm bài';

COMMENT ON COLUMN speaking_questions.max_time 
IS 'Thời gian tối đa cho phép (tính bằng giây)';