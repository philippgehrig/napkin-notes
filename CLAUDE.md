# Napkin Notes

A notes app that renders user text in their own handwriting font on cocktail napkin textures.

## Architecture

Monorepo with three services:
- **services/api/** — Go REST API (net/http)
- **services/fontforge/** — Python font generation worker (fonttools, Pillow, Redis)
- **web/** — Vue 3 + TypeScript SPA (Vite)

## Test Commands

```bash
# Run all tests
make test

# Individual services
make test-api        # Go tests
make test-web        # Vitest (Vue)
make test-fontforge  # pytest (Python)
```

## Development

```bash
# Start each service individually:
cd services/api && go run main.go       # API on :8080
cd web && npm run dev                   # Vue dev server
cd services/fontforge && python3 worker.py
```

## Conventions

- Go: standard library preferred; tests in `_test.go` files alongside source
- Python: pytest for testing; pyproject.toml for config; type hints encouraged
- TypeScript: strict mode; vitest for unit tests; happy-dom as test environment
- Commits: atomic, descriptive messages
- CI: GitHub Actions runs all three test suites in parallel on push/PR to main
