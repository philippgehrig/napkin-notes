# Napkin Notes — Design Specification

## Overview

A notes application inspired by Don Draper's napkin note-taking in Mad Men. Users type notes that render in their own handwriting font on a realistic cocktail napkin texture. Notes are browsed in a gallery view (scattered napkins on a surface) and deleted via an interactive drag-to-rip animation.

## Architecture

**Pattern:** Monorepo with separate deployable services.

```
napkin-notes/
├── services/
│   ├── api/              (Go REST API)
│   └── fontforge/        (Font generation worker)
├── web/                  (Vue 3 SPA)
├── docker/               (Dockerfiles, compose, Traefik config)
└── docs/                 (DNS setup, Figma deliverables)
```

### Services

| Service | Tech | Responsibility |
|---------|------|----------------|
| `api` | Go (Chi or Gin) | REST API — notes CRUD, auth, font management, image export |
| `fontforge` | Python (fonttools + Pillow for glyph extraction) | Async font generation from scanned template sheets |
| `web` | Vue 3 + TypeScript + Yarn | SPA — editor, gallery, font upload, rip-to-delete |
| `db` | PostgreSQL 16 | Persistent storage |
| `redis` | Redis 7 | Job queue for font generation |
| `traefik` | Traefik v3 | Reverse proxy, TLS termination, routing |

### Routing (Traefik)

- `yourdomain.com/` → Vue SPA (nginx container)
- `yourdomain.com/api/` → Go API
- TLS via Let's Encrypt (HTTP or DNS challenge)
- Dev mode: plain HTTP on localhost

### Communication

- Web/Mobile → API: REST over HTTPS (JSON)
- API → fontforge worker: Redis job queue
- API → DB: direct PostgreSQL connection

### Storage

- PostgreSQL: users, notes, font metadata
- Docker volume (local) / S3-compatible (production): font files, template scans, exported images

## Data Model

### users

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key |
| email | VARCHAR | Unique |
| password_hash | VARCHAR | bcrypt |
| display_name | VARCHAR | |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### notes

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key |
| user_id | UUID | FK → users |
| content | TEXT | Note text |
| font_id | UUID | FK → fonts, nullable (falls back to default) |
| deleted_at | TIMESTAMP | Nullable — soft delete (ripped) |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### fonts

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key |
| user_id | UUID | FK → users |
| name | VARCHAR | Display name |
| file_path | VARCHAR | Path to .woff2/.ttf in object storage |
| status | ENUM | pending / processing / ready / failed |
| template_scan_path | VARCHAR | Uploaded scan image path |
| is_default | BOOLEAN | System placeholder font flag |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

## API Endpoints

### Authentication

- `POST /api/auth/register` — create account (email + password)
- `POST /api/auth/login` — returns JWT access + refresh token
- `POST /api/auth/refresh` — refresh access token
- `POST /api/auth/logout` — invalidate refresh token

JWT strategy: access token (15min, Bearer header), refresh token (7 days, httpOnly cookie).

### Notes

- `GET /api/notes` — list user's notes (paginated, excludes soft-deleted)
  - Query params: `sort`, `order`, `limit`, `offset`
- `POST /api/notes` — create napkin
- `GET /api/notes/:id` — get single note
- `PUT /api/notes/:id` — update content/font
- `DELETE /api/notes/:id` — soft delete ("rip")
- `GET /api/notes/trash` — list ripped napkins
- `POST /api/notes/:id/restore` — un-rip (restore from trash)
- `GET /api/notes/:id/export?format=png` — render napkin as PNG image

### Fonts

- `GET /api/fonts` — list user's fonts + system default
- `POST /api/fonts` — upload template scan (triggers async generation)
- `GET /api/fonts/:id` — font metadata + status
- `GET /api/fonts/:id/file` — serve font file (with CORS + cache headers)
- `DELETE /api/fonts/:id`

### Users

- `GET /api/users/me` — current user profile
- `PUT /api/users/me` — update profile

## Image Export

Server-side rendering of napkins as PNG images for saving to phone camera roll.

**Implementation:**
- Go graphics library (e.g., `fogleman/gg`) composites napkin texture + renders text in user's font
- Returns PNG with appropriate dimensions for phone wallpaper/photo saving
- Web: triggers file download
- Mobile (future): same endpoint, piped through native share sheet → save to Photos

## Frontend Structure

```
web/src/
├── views/
│   ├── LoginView.vue
│   ├── RegisterView.vue
│   ├── GalleryView.vue
│   ├── NoteEditorView.vue
│   ├── TrashView.vue
│   └── FontsView.vue
├── components/
│   ├── NapkinCard.vue
│   ├── NapkinTexture.vue
│   ├── RipAnimation.vue
│   ├── FontPreview.vue
│   └── AppNav.vue
├── composables/
│   ├── useAuth.ts
│   ├── useNotes.ts
│   └── useFonts.ts
├── stores/ (Pinia)
│   ├── authStore.ts
│   ├── notesStore.ts
│   └── fontsStore.ts
├── assets/
│   ├── textures/
│   └── fonts/
└── router/
    └── index.ts
```

## Key Interactions

### Gallery View

- Grid of napkin cards showing text preview in note's font on napkin texture
- Cards slightly rotated at random angles (scattered on bar/desk aesthetic)
- Click napkin → opens editor
- Drag gesture on napkin → initiates rip-to-delete

### Note Editor

- Full-screen napkin texture background
- User types, text renders in selected custom font
- Minimal UI — the napkin IS the interface
- Font selector (if user has multiple fonts)
- Export button (downloads PNG)

### Rip-to-Delete (Drag-to-Rip)

1. User grabs one edge of the napkin (touch or mouse)
2. As they drag, a jagged tear progressively reveals along an SVG-templated or procedurally-generated path
3. Once past threshold (~40% across), tear completes automatically
4. Two halves separate with spring physics (via `motion` library), rotate slightly, fade out
5. If released before threshold, napkin snaps back (cancel gesture)
6. Note is soft-deleted, appears in Trash view
7. Trash view shows "taped back together" napkins that can be restored

### Font Upload Flow

1. User navigates to Fonts view
2. Downloads printable template sheet (PDF with character boxes)
3. Fills it in by hand, scans or photographs it
4. Uploads scan image
5. Status shows "processing..." while fontforge worker generates font
6. Once ready (status: `ready`), font becomes available in editor

## Docker & Infrastructure

### docker-compose.yml (production)

Services: traefik, api, fontforge, web, db, redis

**Build strategy:**
- `api`: Multi-stage Go build → scratch/alpine runtime
- `fontforge`: Python image with fonttools + Pillow for template scan processing and glyph extraction
- `web`: Multi-stage Node/Yarn build → nginx:alpine serving `/dist`
- Images tagged via git SHA

### docker-compose.dev.yml (development override)

- Vue dev server with hot reload (replaces nginx)
- Go API with air for hot reload
- No TLS (plain HTTP on localhost)
- Traefik routes to dev ports

### Traefik Configuration

- Entrypoints: 80 (redirect → 443), 443 (TLS)
- Let's Encrypt automatic certificate provisioning
- Labels-based routing per service in docker-compose
- Optional dashboard on separate subdomain (auth-protected)

## DNS Setup Documentation

The docs will guide users through:

1. Pointing an A record to server IP (e.g., `napkin.yourdomain.com`)
2. Optional wildcard or separate record for Traefik dashboard
3. Environment variables for domain config (no hardcoded domains)
4. Traefik auto-provisions TLS via Let's Encrypt
5. Examples for: Cloudflare, Namecheap, generic registrars
6. Firewall: ports 80 + 443 open

## Figma Deliverables (Friend's Scope)

### What he provides:

| Deliverable | Format | Purpose |
|-------------|--------|---------|
| Napkin texture(s) | PNG/SVG, tileable or sized | Note background — real cocktail napkin feel (cream, fiber grain) |
| Color palette | Hex values | App chrome, text, accents — 1960s bar/office mood |
| Handwriting font | .ttf + .woff2 | Default "house font" for all users |
| Visual style guide | Figma page or PDF | Spacing, corners, shadows, gallery napkin appearance |
| Coffee ring / stain overlays | PNG with transparency | Optional decorative napkin elements |
| Rip texture/mask | SVG path | Torn edge shape for rip-to-delete animation |
| App icon / logo | SVG + PNG @2x | Branding |

### What he does NOT need to provide:

- Layout/wireframes (handled in code)
- Component design (buttons, inputs, nav)
- Responsive behavior
- Interaction timing/motion design

### Placeholder Strategy

Until final assets are delivered:
- Font: Google Fonts "Caveat" or "Kalam"
- Texture: CC0 paper/napkin stock texture
- Palette: Cream (#FFF8E7), warm brown (#5C3D2E), charcoal (#2D2D2D)
- Torn edge: Procedurally generated jagged line (Perlin noise zigzag)

All placeholders in `web/src/assets/`, referenced via theme constants — swapping is a one-line change per asset.

## Technology Stack Summary

| Layer | Technology |
|-------|-----------|
| Frontend | Vue 3, TypeScript, Yarn, Pinia, Vue Router |
| Animation | Motion (spring physics), CSS clip-path |
| Backend | Go, Chi/Gin, JWT |
| Database | PostgreSQL 16 |
| Queue | Redis 7 |
| Font generation | fonttools / fontforge |
| Image export | fogleman/gg (Go) |
| Infrastructure | Docker, Traefik v3, Let's Encrypt |
| Dev tooling | air (Go hot reload), Vite (Vue HMR) |
