-- Kindergarten Behavioral Documentation System - SQLite Database
-- Complete DDL with tables, constraints, indexes, and sample data

-- =============================================================================
-- TABLE DEFINITIONS
-- =============================================================================

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) UNIQUE NOT NULL,
    username_hmac VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- e.g., 'teacher', 'admin'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_username_not_empty CHECK (LENGTH(TRIM(username)) > 0),
    CONSTRAINT chk_password_hash_not_empty CHECK (LENGTH(TRIM(password_hash)) > 0),
    CONSTRAINT chk_role_valid CHECK (role IN ('teacher', 'admin'))
);

-- Categories Table (Observation Categories)
CREATE TABLE IF NOT EXISTS categories (
    category_id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_name VARCHAR(200) UNIQUE NOT NULL,
    description TEXT,
    CONSTRAINT chk_category_name_not_empty CHECK (LENGTH(TRIM(category_name)) > 0)
);

-- Teachers Table
CREATE TABLE IF NOT EXISTS teachers (
    teacher_id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_teacher_first_name_not_empty CHECK (LENGTH(TRIM(first_name)) > 0),
    CONSTRAINT chk_teacher_last_name_not_empty CHECK (LENGTH(TRIM(last_name)) > 0)
);


-- Children Table
CREATE TABLE IF NOT EXISTS children (
    child_id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    birthdate VARCHAR(100) NOT NULL,
    admission_date DATE,
    expected_school_enrollment DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_child_first_name_not_empty CHECK (LENGTH(TRIM(first_name)) > 0),
    CONSTRAINT chk_child_last_name_not_empty CHECK (LENGTH(TRIM(last_name)) > 0)
);

-- Child-Teacher Assignments Table (Many-to-Many with Time Intervals)
CREATE TABLE IF NOT EXISTS child_teacher_assignments (
    assignment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    teacher_id INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(child_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(teacher_id) ON DELETE RESTRICT ON UPDATE CASCADE
);

-- Documentation Entries Table (Bildungsdokumentations)
CREATE TABLE IF NOT EXISTS documentation_entries (
    entry_id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id INTEGER NOT NULL,
    documenting_teacher_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    observation_description TEXT NOT NULL,
    observation_date DATE NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT 0,
    approved_by_teacher_id INTEGER,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(child_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (documenting_teacher_id) REFERENCES teachers(teacher_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (approved_by_teacher_id) REFERENCES teachers(teacher_id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT chk_observation_description_not_empty CHECK (LENGTH(TRIM(observation_description)) > 0)
);

-- Processes Table
CREATE TABLE IF NOT EXISTS processes (
    process_id INTEGER PRIMARY KEY AUTOINCREMENT,
    status VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Kita Masterdata table
CREATE TABLE IF NOT EXISTS kita_masterdata (
    name TEXT NOT NULL,
    street TEXT NOT NULL,
    house_number TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    city TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Indexes on frequently queried columns
CREATE INDEX IF NOT EXISTS idx_assignments_child ON child_teacher_assignments(child_id);
CREATE INDEX IF NOT EXISTS idx_documentation_child ON documentation_entries(child_id);
CREATE INDEX IF NOT EXISTS idx_documentation_date ON documentation_entries(observation_date);
CREATE INDEX IF NOT EXISTS idx_documentation_approved ON documentation_entries(approved);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username_hmac);

-- =============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =============================================================================

-- Trigger to update updated_at for teachers
CREATE TRIGGER IF NOT EXISTS trg_teachers_updated_at
    AFTER UPDATE ON teachers
    FOR EACH ROW
BEGIN
    UPDATE teachers SET updated_at = CURRENT_TIMESTAMP WHERE teacher_id = NEW.teacher_id;
END;

-- Trigger to update updated_at for children
CREATE TRIGGER IF NOT EXISTS trg_children_updated_at
    AFTER UPDATE ON children
    FOR EACH ROW
BEGIN
    UPDATE children SET updated_at = CURRENT_TIMESTAMP WHERE child_id = NEW.child_id;
END;

-- Trigger to update updated_at for documentation_entries
CREATE TRIGGER IF NOT EXISTS trg_documentation_updated_at
    AFTER UPDATE ON documentation_entries
    FOR EACH ROW
BEGIN
    UPDATE documentation_entries SET updated_at = CURRENT_TIMESTAMP WHERE entry_id = NEW.entry_id;
END;

-- Trigger to update updated_at for users
CREATE TRIGGER IF NOT EXISTS trg_users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id;
END;

-- Trigger to update updated_at for kita_masterdata
CREATE TRIGGER IF NOT EXISTS trg_kita_masterdata_updated_at
    AFTER UPDATE ON kita_masterdata
    FOR EACH ROW
BEGIN
    UPDATE kita_masterdata SET updated_at = CURRENT_TIMESTAMP WHERE name = NEW.name;
END;
