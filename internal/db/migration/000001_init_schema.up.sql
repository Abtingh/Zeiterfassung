-- =====================================================================
-- 000001_init_schema.up.sql
-- Komplettes Schema zur Erstellung der Tabellen users, time_entries und notification_logs
-- für das Projekt Zeiterfassung
-- =====================================================================

-- 1) Aktivieren der uuid-ossp-Erweiterung für native UUID-Erzeugung
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2) Definition der ENUM-Typen
CREATE TYPE role_enum AS ENUM ('student', 'supervisor', 'admin');
CREATE TYPE notification_type_enum AS ENUM ('EMAIL', 'IN_APP');

-- 3) Tabelle users
-- Diese Tabelle speichert alle Benutzer (Studenten, Supervisoren, Admins).
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT,                           -- Vorname des Benutzers
    last_name     TEXT,                           -- Nachname des Benutzers
    email         TEXT NOT NULL UNIQUE,           -- Eindeutige E-Mail-Adresse für Login
    password_hash TEXT NOT NULL,                  -- Gehashter Passwortwert
    role          role_enum NOT NULL DEFAULT 'student', -- Rolle: 'student', 'supervisor' oder 'admin'

    -- Wenn Rolle = 'student', muss supervisor_id nicht NULL sein
    supervisor_id BIGINT REFERENCES users(id) ON DELETE SET NULL, -- Verweis auf den Supervisor

    start_date    DATE,                            -- Datum des Arbeitsbeginns
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Erstellungszeitpunkt
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Letzte Aktualisierung

    CONSTRAINT chk_student_has_supervisor
        CHECK (role <> 'student' OR supervisor_id IS NOT NULL)  -- Prüft, dass Studenten einen Supervisor haben
);

-- 4) Tabelle time_entries
-- Speichert Arbeitszeiteinträge für jeden Benutzer.
CREATE TABLE IF NOT EXISTS time_entries (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Eindeutige UUID als Primärschlüssel
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Verweis auf Benutzer
    entry_date   DATE   NOT NULL,             -- Datum des Eintrags
    start_time   TIME   NOT NULL,             -- Arbeitsbeginn
    end_time     TIME   NOT NULL,             -- Arbeitsende
    break_min    INTEGER DEFAULT 0,            -- Pausenlänge in Minuten
    duration_h   NUMERIC(5,2) NOT NULL,        -- Nettoarbeitszeit in Stunden (z.B. 7.50)
    note         TEXT,                         -- Freitext-Notiz
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Erstellungszeitpunkt
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Letzte Aktualisierung

    CONSTRAINT chk_time_valid
        CHECK (end_time > start_time)        -- Prüft, dass end_time nach start_time liegt
);

-- 5) Tabelle notification_logs
-- Speichert Protokolle aller gesendeten Benachrichtigungen (E-Mail oder In-App).
CREATE TABLE IF NOT EXISTS notification_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Verweis auf Benutzer
    type        notification_type_enum NOT NULL,   -- Art der Benachrichtigung: 'EMAIL' oder 'IN_APP'
    payload     JSONB NOT NULL,                    -- Rohinhalt der Benachrichtigung (Audit-Zwecke)
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Zeitpunkt des Versands
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()  -- Erstellungszeitpunkt des Logs
);

-- 6) Funktion und Trigger zur automatischen Aktualisierung von updated_at
-- Diese Funktion setzt updated_at bei jedem UPDATE auf den aktuellen Zeitstempel.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger für Tabelle users
CREATE TRIGGER trg_users_set_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- Trigger für Tabelle time_entries
CREATE TRIGGER trg_time_entries_set_updated
    BEFORE UPDATE ON time_entries
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- =====================================================================
-- Ende der Migration für create (up)
-- =====================================================================
