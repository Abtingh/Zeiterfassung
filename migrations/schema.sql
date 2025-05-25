-- =====================================================================
-- Zeiterfassung SQL Schema (PostgreSQL 15+)
-- ---------------------------------------------------------------------
-- نسخهٔ اصلاح‌شده بر اساس بازخورد:
--   • هر دانشجو باید **حتماً** یک سرپرست (supervisor) داشته باشد.
--   • جدول پروژه‌ها (projects) و وابستگی آن حذف شد.
-- تمام کامنت‌ها به‌فارسی توضیح می‌دهند چه تغییری داده‌ایم و چرا.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 0) افزونهٔ UUID
-- ---------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ---------------------------------------------------------------------
-- 1) انواع ENUM
-- ---------------------------------------------------------------------
CREATE TYPE role_enum AS ENUM ('student', 'supervisor', 'admin');
CREATE TYPE notification_type_enum AS ENUM ('EMAIL', 'IN_APP');

-- ---------------------------------------------------------------------
-- 2) جدول Users
-- ---------------------------------------------------------------------
-- این جدول واحد همهٔ انواع کاربر را نگه می‌دارد. برای اجرای الزام «هر
-- دانشجو باید سرپرست داشته باشد» دو کار می‌کنیم:
--   1. ستون supervisor_id را NOT NULL نمی‌کنیم (چون سرپرست/ادمین خودشان
--      نیازی به سرپرست ندارند)؛ اما…
--   2. یک CHECK CONSTRAINT اضافه می‌کنیم: اگر role = 'student' آنگاه
--      supervisor_id باید NOT NULL باشد.
--   3. برای اطمینان بیشتر، یک TRIGGER داریم که تأیید می‌کند supervisor_id
--      واقعاً به کاربری با role = 'supervisor' اشاره می‌کند.
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT,
    last_name     TEXT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          role_enum NOT NULL DEFAULT 'student',

    supervisor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    -- اگر کاربر دانشجو باشد باید مقداردهی شود (بعداً CHECK)

    start_date    DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- الزام شمارهٔ ۱ بالا
    CONSTRAINT chk_student_has_supervisor CHECK (
        role <> 'student' OR supervisor_id IS NOT NULL
    )
);

-- ایندکس ترکیبی برای جستجو
CREATE INDEX idx_users_name_email ON users (last_name, first_name, email);

-- Trigger برای اطمینان از این‌که supervisor_id → role='supervisor'
CREATE OR REPLACE FUNCTION enforce_supervisor_role()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.supervisor_id IS NOT NULL THEN
        PERFORM 1 FROM users u WHERE u.id = NEW.supervisor_id AND u.role = 'supervisor';
        IF NOT FOUND THEN
            RAISE EXCEPTION 'Supervisor (id=%) باید نقش supervisor داشته باشد', NEW.supervisor_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_supervisor_role
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE enforce_supervisor_role();

-- ---------------------------------------------------------------------
-- 3) جدول Time Entries (بدون project_id)
-- ---------------------------------------------------------------------
CREATE TABLE time_entries (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_date   DATE   NOT NULL,
    start_time   TIME   NOT NULL,
    end_time     TIME   NOT NULL,
    break_min    INTEGER DEFAULT 0,
    duration_h   NUMERIC(5,2) NOT NULL,      -- ساعت کار خالص
    note         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_time_valid CHECK (end_time > start_time)
);

CREATE INDEX idx_time_entries_user_date ON time_entries (user_id, entry_date);

-- ---------------------------------------------------------------------
-- 4) جدول Notification Logs
-- ---------------------------------------------------------------------
CREATE TABLE notification_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type_enum NOT NULL,
    payload     JSONB NOT NULL,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------
-- 5) تابع و تریگر set_updated_at (مشترک)
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_set_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TRIGGER trg_time_entries_set_updated
    BEFORE UPDATE ON time_entries
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- ---------------------------------------------------------------------
-- پایان اسکریپت
-- ---------------------------------------------------------------------
