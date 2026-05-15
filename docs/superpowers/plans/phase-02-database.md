# Phase 2: Database & Migrations

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Set up PostgreSQL schema with migration tooling (golang-migrate) and seed data.

**Branch:** `feat/phase-02-database`

---

## File Structure

```
services/api/
├── internal/
│   └── database/
│       ├── database.go        (connection pool)
│       └── database_test.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_fonts.up.sql
│   ├── 000002_create_fonts.down.sql
│   ├── 000003_create_notes.up.sql
│   └── 000003_create_notes.down.sql
└── go.mod (updated with deps)
```

---

### Task 1: Add database dependencies

**Files:**
- Modify: `services/api/go.mod`

- [ ] **Step 1: Add dependencies**

```bash
cd services/api && go get github.com/lib/pq github.com/golang-migrate/migrate/v4 github.com/golang-migrate/migrate/v4/database/postgres github.com/golang-migrate/migrate/v4/source/file
```

- [ ] **Step 2: Commit**

```bash
git add services/api/go.mod services/api/go.sum
git commit -m "feat: add database dependencies (lib/pq, golang-migrate)"
```

---

### Task 2: Create migration files

**Files:**
- Create: `services/api/migrations/000001_create_users.up.sql`
- Create: `services/api/migrations/000001_create_users.down.sql`
- Create: `services/api/migrations/000002_create_fonts.up.sql`
- Create: `services/api/migrations/000002_create_fonts.down.sql`
- Create: `services/api/migrations/000003_create_notes.up.sql`
- Create: `services/api/migrations/000003_create_notes.down.sql`

- [ ] **Step 1: Create users migration**

`services/api/migrations/000001_create_users.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

`services/api/migrations/000001_create_users.down.sql`:
```sql
DROP TABLE IF EXISTS users;
```

- [ ] **Step 2: Create fonts migration**

`services/api/migrations/000002_create_fonts.up.sql`:
```sql
CREATE TYPE font_status AS ENUM ('pending', 'processing', 'ready', 'failed');

CREATE TABLE fonts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    file_path VARCHAR(500) NOT NULL DEFAULT '',
    status font_status NOT NULL DEFAULT 'pending',
    template_scan_path VARCHAR(500) NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fonts_user_id ON fonts(user_id);
```

`services/api/migrations/000002_create_fonts.down.sql`:
```sql
DROP TABLE IF EXISTS fonts;
DROP TYPE IF EXISTS font_status;
```

- [ ] **Step 3: Create notes migration**

`services/api/migrations/000003_create_notes.up.sql`:
```sql
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    font_id UUID REFERENCES fonts(id) ON DELETE SET NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_deleted_at ON notes(deleted_at);
```

`services/api/migrations/000003_create_notes.down.sql`:
```sql
DROP TABLE IF EXISTS notes;
```

- [ ] **Step 4: Commit**

```bash
git add services/api/migrations/
git commit -m "feat: add database migration files for users, fonts, notes"
```

---

### Task 3: Database connection package

**Files:**
- Create: `services/api/internal/database/database.go`
- Create: `services/api/internal/database/database_test.go`

- [ ] **Step 1: Write database connection test**

Create `services/api/internal/database/database_test.go`:
```go
package database

import (
	"testing"
)

func TestParseDSN(t *testing.T) {
	dsn := BuildDSN("localhost", "5432", "napkin", "user", "pass", "disable")
	expected := "host=localhost port=5432 dbname=napkin user=user password=pass sslmode=disable"
	if dsn != expected {
		t.Errorf("expected %q, got %q", expected, dsn)
	}
}

func TestBuildDSNFromEnv_Defaults(t *testing.T) {
	dsn := BuildDSN("", "", "", "", "", "")
	expected := "host=localhost port=5432 dbname=napkin_notes user=postgres password= sslmode=disable"
	if dsn != expected {
		t.Errorf("expected %q, got %q", expected, dsn)
	}
}
```

- [ ] **Step 2: Run test — expect fail**

```bash
cd services/api && go test ./internal/database/ -v
```
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement database.go**

Create `services/api/internal/database/database.go`:
```go
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func BuildDSN(host, port, dbname, user, password, sslmode string) string {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if dbname == "" {
		dbname = "napkin_notes"
	}
	if user == "" {
		user = "postgres"
	}
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, password, sslmode)
}

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}
```

- [ ] **Step 4: Run test — expect pass**

```bash
cd services/api && go test ./internal/database/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/
git commit -m "feat: add database connection package with DSN builder"
```

---

### Task 4: Add migrate command to main

**Files:**
- Modify: `services/api/main.go`

- [ ] **Step 1: Add migration runner to main.go**

Add to `services/api/main.go` before `main()`:
```go
import (
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/philippgehrig/napkin-notes/services/api/internal/database"
)

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		ex, _ := os.Executable()
		migrationsPath = filepath.Join(filepath.Dir(ex), "migrations")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}
```

Update `main()` to optionally run migrations:
```go
func main() {
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		dsn := database.BuildDSN(
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_SSLMODE"),
		)
		db, err := database.Connect(dsn)
		if err != nil {
			log.Fatalf("database connection failed: %v", err)
		}
		if err := runMigrations(db); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
		log.Println("migrations completed successfully")
	}

	http.HandleFunc("/health", healthHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("API listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd services/api && go build ./...
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add services/api/
git commit -m "feat: add migration runner to API startup"
```

---

### Task 5: Add docker-compose with PostgreSQL for dev

**Files:**
- Create: `docker/docker-compose.yml`
- Create: `docker/.env.example`

- [ ] **Step 1: Create docker-compose.yml**

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${DB_NAME:-napkin_notes}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

- [ ] **Step 2: Create .env.example**

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=napkin_notes
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
REDIS_URL=redis://localhost:6379
PORT=8080
JWT_SECRET=change-me-in-production
```

- [ ] **Step 3: Commit**

```bash
git add docker/
git commit -m "feat: add docker-compose with PostgreSQL and Redis"
```

---

### Task 6: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-02-database
gh pr create --title "feat: database schema and migrations" --body "## Summary
- Add golang-migrate for schema management
- Create migrations for users, fonts, and notes tables
- Add database connection package with DSN builder
- Add migration runner to API startup
- Add docker-compose with PostgreSQL and Redis

## Test plan
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] \`docker compose -f docker/docker-compose.yml up db\` starts PostgreSQL
- [ ] Migrations run successfully with RUN_MIGRATIONS=true

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
