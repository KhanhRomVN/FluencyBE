--! =================================================================
--! COURSES - Bảng khóa học chính
--! =================================================================
CREATE TABLE IF NOT EXISTS courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    overview TEXT NOT NULL CHECK (length(trim(overview)) > 0),
    skills TEXT[] NOT NULL CHECK (array_length(skills, 1) > 0),
    band TEXT NOT NULL CHECK (length(trim(band)) > 0),
    image_urls TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_title UNIQUE (title)
);

-- Tối ưu tìm kiếm
CREATE INDEX IF NOT EXISTS idx_courses_skills ON courses USING GIN(skills);
CREATE INDEX IF NOT EXISTS idx_courses_band ON courses(band);
CREATE INDEX IF NOT EXISTS idx_courses_title_search 
ON courses USING gin(to_tsvector('english', title || ' ' || overview));

--! =================================================================
--! COURSE BOOKS -  Khóa học dạng sách
--! =================================================================
CREATE TABLE IF NOT EXISTS course_books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    publishers TEXT[] NOT NULL CHECK (array_length(publishers, 1) > 0),
    authors TEXT[] NOT NULL CHECK (array_length(authors, 1) > 0),
    publication_year INT NOT NULL CHECK (
        publication_year >= 1900 AND 
        publication_year <= EXTRACT(YEAR FROM CURRENT_DATE)
    ),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_book UNIQUE (course_id)
);

-- Tối ưu tìm kiếm
CREATE INDEX IF NOT EXISTS idx_course_books_course_id ON course_books(course_id);
CREATE INDEX IF NOT EXISTS idx_course_books_publishers ON course_books USING GIN(publishers);
CREATE INDEX IF NOT EXISTS idx_course_books_authors ON course_books USING GIN(authors);
CREATE INDEX IF NOT EXISTS idx_course_books_year ON course_books(publication_year);

--! =================================================================
--! COURSE OTHERS - Các khóa học khác
--! =================================================================
CREATE TABLE IF NOT EXISTS course_others (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_other UNIQUE (course_id)
);

CREATE INDEX IF NOT EXISTS idx_course_others_course_id ON course_others(course_id);

--! =================================================================
--! LESSONS - Các bài học
--! =================================================================
CREATE TABLE IF NOT EXISTS lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    sequence INT NOT NULL CHECK (sequence > 0),
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    overview TEXT NOT NULL CHECK (length(trim(overview)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_lesson_sequence UNIQUE(course_id, sequence),
    CONSTRAINT unique_lesson_title_per_course UNIQUE(course_id, title)
);

-- Tối ưu tìm kiếm
CREATE INDEX IF NOT EXISTS idx_lessons_course_id ON lessons(course_id);
CREATE INDEX IF NOT EXISTS idx_lessons_sequence ON lessons(sequence);
CREATE INDEX IF NOT EXISTS idx_lessons_title_search 
ON lessons USING gin(to_tsvector('english', title || ' ' || overview));

--! =================================================================
--! LESSON QUESTIONS - Câu hỏi trong bài học
--! =================================================================
CREATE TABLE IF NOT EXISTS lesson_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    sequence INT NOT NULL CHECK (sequence > 0),
    question_id UUID NOT NULL,
    question_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_question_sequence UNIQUE(lesson_id, sequence),
    CONSTRAINT unique_question_per_lesson UNIQUE(lesson_id, question_id)
);

-- Tối ưu tìm kiếm
CREATE INDEX IF NOT EXISTS idx_lesson_questions_lesson_id ON lesson_questions(lesson_id);
CREATE INDEX IF NOT EXISTS idx_lesson_questions_question_id ON lesson_questions(question_id);
CREATE INDEX IF NOT EXISTS idx_lesson_questions_sequence ON lesson_questions(sequence);

--! =================================================================
--! FUNCTIONS - Quản lý sequence và timestamp
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

-- Function để lấy sequence tiếp theo cho lesson
CREATE OR REPLACE FUNCTION get_next_lesson_sequence(course_uuid UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN COALESCE(
        (SELECT MAX(sequence) + 1
         FROM lessons
         WHERE course_id = course_uuid),
        1
    );
END;
$$ LANGUAGE plpgsql;

-- Function để lấy sequence tiếp theo cho lesson question
CREATE OR REPLACE FUNCTION get_next_lesson_question_sequence(lesson_uuid UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN COALESCE(
        (SELECT MAX(sequence) + 1
         FROM lesson_questions
         WHERE lesson_id = lesson_uuid),
        1
    );
END;
$$ LANGUAGE plpgsql;

-- Function để resequence lessons sau khi xóa
CREATE OR REPLACE FUNCTION resequence_lessons()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE lessons
    SET sequence = sq.new_sequence
    FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY course_id 
            ORDER BY sequence
        ) as new_sequence
        FROM lessons
        WHERE course_id = OLD.course_id
          AND sequence > OLD.sequence
    ) sq
    WHERE lessons.id = sq.id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Function để resequence lesson questions sau khi xóa
CREATE OR REPLACE FUNCTION resequence_lesson_questions()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE lesson_questions
    SET sequence = sq.new_sequence
    FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY lesson_id 
            ORDER BY sequence
        ) as new_sequence
        FROM lesson_questions
        WHERE lesson_id = OLD.lesson_id
          AND sequence > OLD.sequence
    ) sq
    WHERE lesson_questions.id = sq.id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Function để swap sequence của lessons
CREATE OR REPLACE FUNCTION swap_lesson_sequence(
    lesson1_uuid UUID,
    lesson2_uuid UUID
)
RETURNS VOID AS $$
DECLARE
    seq1 INTEGER;
    seq2 INTEGER;
    course_id1 UUID;
    course_id2 UUID;
    temp_seq INTEGER;
BEGIN
    -- Kiểm tra xem lessons có cùng course không
    SELECT course_id, sequence INTO course_id1, seq1 
    FROM lessons WHERE id = lesson1_uuid;
    
    SELECT course_id, sequence INTO course_id2, seq2 
    FROM lessons WHERE id = lesson2_uuid;
    
    IF course_id1 != course_id2 THEN
        RAISE EXCEPTION 'Lessons must belong to the same course';
    END IF;

    -- Lấy một sequence tạm thời lớn hơn max hiện tại
    SELECT COALESCE(MAX(sequence), 0) + 1000 INTO temp_seq 
    FROM lessons 
    WHERE course_id = course_id1;

    -- Thực hiện swap trong 3 bước để tránh vi phạm constraint
    -- Bước 1: Đổi lesson1 sang sequence tạm
    UPDATE lessons 
    SET sequence = temp_seq,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = lesson1_uuid;

    -- Bước 2: Đổi lesson2 sang sequence của lesson1
    UPDATE lessons 
    SET sequence = seq1,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = lesson2_uuid;

    -- Bước 3: Đổi lesson1 từ sequence tạm sang sequence của lesson2
    UPDATE lessons 
    SET sequence = seq2,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = lesson1_uuid;

END;
$$ LANGUAGE plpgsql;

-- Function để swap sequence của lesson questions (tương tự)
CREATE OR REPLACE FUNCTION swap_lesson_question_sequence(
    question1_uuid UUID,
    question2_uuid UUID
)
RETURNS VOID AS $$
DECLARE
    seq1 INTEGER;
    seq2 INTEGER;
    lesson_id1 UUID;
    lesson_id2 UUID;
    temp_seq INTEGER;
BEGIN
    -- Kiểm tra xem questions có cùng lesson không
    SELECT lesson_id, sequence INTO lesson_id1, seq1 
    FROM lesson_questions WHERE id = question1_uuid;
    
    SELECT lesson_id, sequence INTO lesson_id2, seq2 
    FROM lesson_questions WHERE id = question2_uuid;
    
    IF lesson_id1 != lesson_id2 THEN
        RAISE EXCEPTION 'Questions must belong to the same lesson';
    END IF;

    -- Lấy một sequence tạm thời lớn hơn max hiện tại
    SELECT COALESCE(MAX(sequence), 0) + 1000 INTO temp_seq 
    FROM lesson_questions 
    WHERE lesson_id = lesson_id1;

    -- Thực hiện swap trong 3 bước để tránh vi phạm constraint
    -- Bước 1: Đổi question1 sang sequence tạm
    UPDATE lesson_questions 
    SET sequence = temp_seq,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = question1_uuid;

    -- Bước 2: Đổi question2 sang sequence của question1
    UPDATE lesson_questions 
    SET sequence = seq1,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = question2_uuid;

    -- Bước 3: Đổi question1 từ sequence tạm sang sequence của question2
    UPDATE lesson_questions 
    SET sequence = seq2,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = question1_uuid;

END;
$$ LANGUAGE plpgsql;

--! =================================================================
--! TRIGGERS
--! =================================================================
-- Triggers cho timestamp
DO $$ 
DECLARE 
    tbl RECORD;
BEGIN
    FOR tbl IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public'
        AND tablename IN (
            'courses', 'course_books', 'course_others',
            'lessons', 'lesson_questions'
        )
    LOOP
        EXECUTE format('
            CREATE TRIGGER trigger_%I_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at();',
            tbl.tablename, tbl.tablename);
    END LOOP;
END $$;

-- Trigger để tự động resequence lessons sau khi xóa
DROP TRIGGER IF EXISTS trigger_lessons_resequence ON lessons;
CREATE TRIGGER trigger_lessons_resequence
    AFTER DELETE ON lessons
    FOR EACH ROW
    EXECUTE FUNCTION resequence_lessons();

-- Trigger để tự động resequence lesson questions sau khi xóa
DROP TRIGGER IF EXISTS trigger_lesson_questions_resequence ON lesson_questions;
CREATE TRIGGER trigger_lesson_questions_resequence
    AFTER DELETE ON lesson_questions
    FOR EACH ROW
    EXECUTE FUNCTION resequence_lesson_questions();

--! =================================================================
--! COMMENTS - Giải thích các bảng
--! =================================================================
COMMENT ON TABLE courses 
IS 'Bảng chứa thông tin chung về các khóa học';

COMMENT ON TABLE course_books 
IS 'Bảng chứa thông tin chi tiết về các khóa học dạng sách giáo trình';

COMMENT ON TABLE course_others 
IS 'Bảng chứa thông tin chi tiết về các khóa học không phải dạng sách';

COMMENT ON TABLE lessons 
IS 'Bảng chứa thông tin về các bài học trong khóa học';

COMMENT ON TABLE lesson_questions 
IS 'Bảng liên kết giữa bài học và câu hỏi';

COMMENT ON COLUMN courses.type 
IS 'Loại khóa học: book (sách giáo trình) hoặc other (khác)';

COMMENT ON COLUMN courses.skills 
IS 'Các kỹ năng được đào tạo trong khóa học';

COMMENT ON COLUMN courses.band 
IS 'Cấp độ của khóa học';

-- Comments cho các functions
COMMENT ON FUNCTION get_next_lesson_sequence 
IS 'Function tự động lấy sequence tiếp theo cho lesson mới';

COMMENT ON FUNCTION get_next_lesson_question_sequence 
IS 'Function tự động lấy sequence tiếp theo cho question mới';

COMMENT ON FUNCTION resequence_lessons 
IS 'Function tự động cập nhật lại sequence của lessons sau khi xóa';

COMMENT ON FUNCTION resequence_lesson_questions 
IS 'Function tự động cập nhật lại sequence của questions sau khi xóa';

COMMENT ON FUNCTION swap_lesson_sequence 
IS 'Function để swap sequence của hai lessons';

COMMENT ON FUNCTION swap_lesson_question_sequence 
IS 'Function để swap sequence của hai questions';