--! =================================================================
--! COMMON FUNCTIONS
--! =================================================================
-- Function kiểm tra email hợp lệ
CREATE OR REPLACE FUNCTION is_valid_email(email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$';
END;
$$ LANGUAGE plpgsql;

-- Function kiểm tra username hợp lệ
CREATE OR REPLACE FUNCTION is_valid_username(username TEXT) 
RETURNS BOOLEAN AS $$
BEGIN
    RETURN username ~* '^[a-zA-Z0-9_]{3,30}$';
END;
$$ LANGUAGE plpgsql;

-- Function cập nhật updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--! =================================================================
--! USERS TABLE
--! =================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL CHECK (is_valid_email(email)),
    username TEXT NOT NULL CHECK (is_valid_username(username)),
    password TEXT NOT NULL CHECK (length(password) >= 60),
    type VARCHAR(25) NOT NULL CHECK (type IN ('basic', 'google', 'facebook', 'other')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_unique UNIQUE (username)
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_type ON users(type);

-- Trigger for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--! =================================================================
--! DEVELOPERS TABLE
--! =================================================================
CREATE TABLE IF NOT EXISTS developers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL CHECK (is_valid_email(email)),
    username TEXT NOT NULL CHECK (is_valid_username(username)),
    password TEXT NOT NULL CHECK (length(password) >= 60),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT developers_email_unique UNIQUE (email),
    CONSTRAINT developers_username_unique UNIQUE (username)
);

-- Indexes
CREATE INDEX idx_developers_email ON developers(email);
CREATE INDEX idx_developers_username ON developers(username);

-- Trigger for updated_at
CREATE TRIGGER update_developers_updated_at
    BEFORE UPDATE ON developers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--! =================================================================
--! ADDITIONAL CONSTRAINTS
--! =================================================================
-- Prevent duplicate emails across both tables
CREATE OR REPLACE FUNCTION check_email_unique_across_tables()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM users WHERE email = NEW.email
        UNION
        SELECT 1 FROM developers WHERE email = NEW.email AND id != NEW.id
    ) THEN
        RAISE EXCEPTION 'Email already exists in users or developers table';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply cross-table email check triggers
CREATE TRIGGER check_users_email
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION check_email_unique_across_tables();

CREATE TRIGGER check_developers_email
    BEFORE INSERT OR UPDATE ON developers
    FOR EACH ROW
    EXECUTE FUNCTION check_email_unique_across_tables();
