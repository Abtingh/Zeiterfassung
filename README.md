# Zeiterfassung

Ein webbasiertes Zeiterfassungssystem für Studierende, Teamleiter, Administratoren und die Buchhaltung.

## 📋 Überblick

Diese Anwendung ermöglicht die vollständige digitale Erfassung, Genehmigung und Verwaltung von Arbeitszeiten. Das System unterstützt verschiedene Benutzerrollen mit spezifischen Berechtigungen und Workflows.

### Benutzerrollen

| Rolle | Beschreibung |
|-------|-------------|
| **Student** | Erfasst eigene Arbeitszeiten und reicht diese wöchentlich ein |
| **Supervisor (Teamleiter)** | Genehmigt oder korrigiert eingereichte Zeiten der Team-Mitglieder |
| **Admin** | Verwaltet Benutzer und Teams |
| **Accounting (Buchhaltung)** | Bearbeitet genehmigte Zeiteinträge final |

### Workflow

```
Student erfasst Zeit → Einreichung → Teamleiter genehmigt → Buchhaltung bearbeitet
                              ↓
                      Korrektur (bei Bedarf)
```

## 🛠️ Technologie-Stack

- **Backend:** Go 1.24 mit Standard-HTTP-Bibliothek
- **Datenbank:** PostgreSQL 16
- **ORM/Query:** SQLC (typsichere SQL-Queries)
- **Frontend:** HTML, CSS, JavaScript (Vanilla)
- **Containerisierung:** Docker & Docker Compose
- **KI-Integration:** n8n Webhook für automatische Zeiterfassung

## 📁 Projektstruktur

```
zeiterfassung/
├── main.go                    # Einstiegspunkt der Anwendung
├── docker-compose.yml         # Docker-Konfiguration
├── Dockerfile                 # Container-Build-Anweisungen
├── sqlc.yaml                  # SQLC-Konfiguration
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP-Handler für alle Endpunkte
│   │   │   ├── admin.go       # Admin-Funktionen
│   │   │   ├── ai.go          # KI-gestützte Zeiterfassung
│   │   │   ├── auth.go        # Authentifizierung
│   │   │   ├── buchhaltung.go # Buchhaltungs-Funktionen
│   │   │   ├── student.go     # Studenten-Funktionen
│   │   │   └── supervisor.go  # Teamleiter-Funktionen
│   │   └── middleware/        # HTTP-Middleware
│   ├── db/
│   │   ├── migration/         # Datenbankmigrationen
│   │   ├── query/             # SQL-Abfragen für SQLC
│   │   └── sqlc/              # Generierter Code
│   ├── routes/                # Routing-Konfiguration
│   ├── server/                # Server-Konfiguration
│   └── util/                  # Hilfsfunktionen
├── public/
│   ├── public-static/         # HTML-Seiten
│   └── src/
│       ├── services/          # JavaScript-Services
│       └── styles/            # CSS-Stylesheets
└── data/                      # PostgreSQL-Daten (Volume)
```

## 🚀 Installation & Start

### Voraussetzungen

- Docker & Docker Compose
- Git

### Schnellstart

1. **Repository klonen:**
   ```bash
   git clone https://github.com/Abtingh/Zeiterfassung.git
   cd zeiterfassung
   ```

2. **Anwendung starten:**
   ```bash
   docker compose up -d
   ```

3. **Anwendung aufrufen:**
   - **Web-App:** http://localhost:8080
   - **pgAdmin:** http://localhost:5050
     - E-Mail: `admin@localhost.com`
     - Passwort: `admin`

### Dienste stoppen

```bash
docker compose down
```

### Datenbank zurücksetzen

```bash
docker compose down
rm -rf ./data
docker compose up -d
```

## ⚙️ Konfiguration

Die Anwendung wird über Umgebungsvariablen konfiguriert. Diese sind in der `docker-compose.yml` definiert:

| Variable | Beschreibung | Standardwert |
|----------|-------------|--------------|
| `APP_PORT` | Port der Anwendung | `8080` |
| `APP_DEBUG` | Debug-Modus | `false` |
| `DB_HOST` | Datenbank-Host | `db` |
| `DB_PORT` | Datenbank-Port | `5432` |
| `DB_USERNAME` | Datenbank-Benutzer | `postgres` |
| `DB_PASSWORD` | Datenbank-Passwort | `postgres` |
| `DB_DATABASE` | Datenbankname | `zeiterfassung` |
| `JWT_SECRET` | Geheimer Schlüssel für JWT | - |
| `N8N_WEBHOOK_URL` | URL für KI-Zeiteintrag | - |

## 📡 API-Endpunkte

### Authentifizierung
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| POST | `/login` | Benutzer-Anmeldung |
| POST | `/logout` | Benutzer-Abmeldung |
| GET | `/me` | Aktuelle Benutzerinfos |
| GET | `/reset-passwort/{token}` | Passwort zurücksetzen |

### Student
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| POST | `/api/time-entries/submit` | Woche einreichen |
| GET | `/api/time-entries/week` | Wocheneinträge abrufen |

### Teamleiter (Supervisor)
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/api/supervisor/pending-entries` | Ausstehende Einreichungen |
| GET | `/api/supervisor/students` | Team-Mitglieder |
| POST | `/api/supervisor/approve` | Einreichung genehmigen |
| POST | `/api/supervisor/reject` | Einreichung ablehnen |

### Admin
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| POST | `/admin/users/create` | Benutzer anlegen |
| GET | `/admin/users/list` | Benutzer auflisten |
| PUT | `/admin/users/update` | Benutzer aktualisieren |
| PUT | `/admin/users/deactivate` | Benutzer deaktivieren |
| POST | `/admin/teams/create` | Team erstellen |
| GET | `/admin/teams/list` | Teams auflisten |
| DELETE | `/admin/teams/delete` | Team löschen |

### Buchhaltung
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/api/accounting/approved-entries` | Genehmigte Einreichungen |
| GET | `/api/accounting/students` | Alle Studenten |
| POST | `/api/accounting/mark-processed` | Als bearbeitet markieren |

### KI-Integration
| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| POST | `/api/ai/parse-time-entry` | Automatische Zeiterkennung |

## 🗄️ Datenbankschema

### Haupttabellen

- **users** - Benutzerkonten mit Rollen
- **teams** - Teams mit zugewiesenem Supervisor
- **weekly_submissions** - Wöchentliche Zeiterfassungs-Einreichungen
- **time_entries** - Einzelne Tages-Zeiteinträge
- **notifications** - Benachrichtigungs-Warteschlange

### Status einer Einreichung

| Status | Beschreibung |
|--------|--------------|
| `offen` | In Bearbeitung durch den Studenten |
| `gesendet` | Zur Genehmigung übermittelt |
| `korrektur` | Zur Korrektur zurückgesendet |
| `bestaetigt` | Vom Teamleiter genehmigt |
| `erledigt` | Von der Buchhaltung bearbeitet |

## 🔧 Entwicklung

### SQLC-Queries neu generieren

```bash
sqlc generate
```

### Lokale Entwicklung ohne Docker

1. PostgreSQL installieren und starten
2. Datenbank erstellen:
   ```sql
   CREATE DATABASE zeiterfassung;
   ```
3. Migration ausführen (SQL in `internal/db/migration/`)
4. Umgebungsvariablen setzen
5. Anwendung starten:
   ```bash
   go run main.go
   ```

## 📄 Lizenz

Dieses Projekt ist für interne Nutzung bestimmt.

## 👥 Mitwirkende

- Abtin Ghaffari ([@Abtingh](https://github.com/Abtingh))
