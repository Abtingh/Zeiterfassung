-- =====================================================================
-- Definitives Datenbankschema für das Projekt "Zeiterfassung"
-- Version: 4.0
-- Beschreibung: Finale Version mit bereinigten Kommentaren.
-- =====================================================================


-- 1. Erweiterungen und benutzerdefinierte Typen (ENUMs)
-- ---------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Definiert die möglichen Benutzerrollen im System.
CREATE TYPE role_enum AS ENUM (
    'student',
    'supervisor',
    'admin',
    'accounting'
);

-- Definiert die möglichen Versandarten für Benachrichtigungen.
CREATE TYPE notification_type_enum AS ENUM (
    'EMAIL',    -- Benachrichtigung per E-Mail
    'IN_APP'    -- Benachrichtigung innerhalb der Anwendung
);

-- Definiert den Lebenszyklus-Status einer wöchentlichen Zeiterfassung.
CREATE TYPE submission_status_enum AS ENUM (
    'offen',        -- In Bearbeitung durch den Studenten
    'gesendet',     -- Zur Genehmigung an den Vorgesetzten übermittelt
    'korrektur',    -- Vom Vorgesetzten zur Korrektur zurückgesendet
    'bestaetigt',   -- Vom Vorgesetzten genehmigt
    'erledigt'      -- Von der Buchhaltung final bearbeitet
);


-- 2. Haupttabellen
-- ---------------------------------------------------------------------

-- Tabelle zur Verwaltung der Teams.
CREATE TABLE IF NOT EXISTS teams (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,                               -- Eindeutiger Name des Teams
    supervisor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,     -- Der dem Team zugeordnete Vorgesetzte
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabelle zur Speicherung aller Benutzerkonten.
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT,
    last_name     TEXT,
    email         TEXT NOT NULL UNIQUE,                               -- Eindeutige E-Mail für den Login
    password_hash TEXT NOT NULL,
    role          role_enum NOT NULL DEFAULT 'student',
    team_id       BIGINT REFERENCES teams(id) ON DELETE SET NULL,     -- Verknüpfung zu einem Team
    start_date    DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Kern-Tabelle: Speichert den Status einer kompletten Arbeitswoche pro Benutzer.
CREATE TABLE IF NOT EXISTS weekly_submissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_number     INTEGER NOT NULL,                                   -- Kalenderwoche (z.B. 42)
    year            INTEGER NOT NULL,                                   -- Jahr (z.B. 2025)
    status          submission_status_enum NOT NULL DEFAULT 'offen',    -- Aktueller Status der Einreichung
    submitted_at    TIMESTAMPTZ,                                        -- Zeitpunkt der Einreichung
    approved_at     TIMESTAMPTZ,                                        -- Zeitpunkt der Genehmigung
    processed_by    BIGINT REFERENCES users(id) ON DELETE SET NULL,     -- Welcher Buchhalter hat es bearbeitet
    processed_at    TIMESTAMPTZ,                                        -- Zeitpunkt des Abschlusses durch die Buchhaltung
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Eindeutiger Index, um doppelte Wochen pro Benutzer zu verhindern.
    UNIQUE(user_id, week_number, year)
);

-- Speichert die einzelnen täglichen Zeiteinträge für eine Arbeitswoche.
CREATE TABLE IF NOT EXISTS time_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id   UUID NOT NULL REFERENCES weekly_submissions(id) ON DELETE CASCADE, -- Verknüpfung zur Arbeitswoche
    entry_date      DATE NOT NULL,                                      -- Datum des Eintrags
    start_time      TIME NOT NULL,                                      -- Arbeitsbeginn
    end_time        TIME NOT NULL,                                      -- Arbeitsende
    break_min       INTEGER DEFAULT 0,                                  -- Pause in Minuten
    duration_h      NUMERIC(5, 2) NOT NULL,                             -- Netto-Arbeitsstunden
    note            TEXT,                                               -- Optionale Anmerkungen
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Stellt die logische Korrektheit der Zeitangaben sicher.
    CONSTRAINT chk_time_valid CHECK (end_time > start_time)
);

-- Dient als Warteschlange für alle zu versendenden Benachrichtigungen.
CREATE TABLE IF NOT EXISTS notifications (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Empfänger der Benachrichtigung
    message     TEXT NOT NULL,                                      -- Inhalt der Nachricht
    type        notification_type_enum NOT NULL,                    -- Versandart: 'EMAIL' oder 'IN_APP'
    status      TEXT NOT NULL DEFAULT 'pending',                    -- Verarbeitungsstatus (pending, sent, failed)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at     TIMESTAMPTZ                                         -- Zeitpunkt des tatsächlichen Versands
);


-- 3. Trigger-Funktionen (Beispielhaft, Implementierung wie zuvor)
-- ---------------------------------------------------------------------
-- Die Trigger zur automatischen Aktualisierung der 'updated_at'-Spalte
-- sollten hier wie im ursprünglichen Skript definiert werden.


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

-- Trigger für Tabelle weekly_submissions
CREATE TRIGGER trg_weekly_submissions_set_updated
    BEFORE UPDATE ON weekly_submissions
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- =====================================================================
-- Ende der Migration für create (up)
-- =====================================================================