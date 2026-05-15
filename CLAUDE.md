# Projects Monorepo

Multi-project repository. Each project lives in its own top-level directory.

## Projects

### napkin-notes/
A notes app that renders user text in their own handwriting font on cocktail napkin textures.

**Architecture:** Go REST API + Python font worker + Vue 3 SPA, deployed via Docker Compose with Traefik.

- `napkin-notes/services/api/` — Go REST API (net/http)
- `napkin-notes/services/fontforge/` — Python font generation worker (fonttools, Pillow, Redis)
- `napkin-notes/web/` — Vue 3 + TypeScript SPA (Vite)
- `napkin-notes/docker/` — Docker Compose configs (prod + dev)
- `napkin-notes/e2e/` — Playwright E2E tests

## Test Commands

```bash
# Run all tests
make test

# Individual services
make test-api        # Go tests
make test-web        # Vitest (Vue)
make test-fontforge  # pytest (Python)
make test-e2e        # Playwright
```

## Development

```bash
# Full stack via Docker:
make dev

# Individual services:
cd napkin-notes/services/api && go run main.go       # API on :8080
cd napkin-notes/web && npm run dev                   # Vue dev server
cd napkin-notes/services/fontforge && python3 worker.py
```

## Conventions

- Go: standard library preferred; tests in `_test.go` files alongside source
- Python: pytest for testing; pyproject.toml for config; type hints encouraged
- TypeScript: strict mode; vitest for unit tests; happy-dom as test environment
- Commits: atomic, descriptive messages
- CI: GitHub Actions with path filters — each project only triggers its own jobs
- Deploy: automatic on push to main after all tests pass
