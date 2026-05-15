# Phase 1: Project Scaffold

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Set up the monorepo structure with all service directories, basic tooling, and CI.

**Branch:** `feat/phase-01-scaffold`

---

## File Structure

```
napkin-notes/
├── services/
│   ├── api/
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── go.sum
│   └── fontforge/
│       ├── worker.py
│       ├── requirements.txt
│       └── pyproject.toml
├── web/
│   └── (Vite scaffold — next step)
├── docker/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── .env.example
├── docs/
├── .github/
│   └── workflows/
│       └── ci.yml
├── CLAUDE.md
├── .gitignore
└── Makefile
```

---

### Task 1: Initialize Go API module

**Files:**
- Create: `services/api/go.mod`
- Create: `services/api/main.go`
- Create: `services/api/main_test.go`

- [ ] **Step 1: Create go.mod**

```bash
mkdir -p services/api && cd services/api && go mod init github.com/philippgehrig/napkin-notes/services/api
```

- [ ] **Step 2: Write a minimal health endpoint test**

Create `services/api/main_test.go`:
```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
```

- [ ] **Step 3: Run test — expect fail**

```bash
cd services/api && go test ./... -v
```
Expected: FAIL — `healthHandler` undefined

- [ ] **Step 4: Implement minimal main.go**

Create `services/api/main.go`:
```go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("API listening on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
```

- [ ] **Step 5: Run test — expect pass**

```bash
cd services/api && go test ./... -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add services/api/
git commit -m "feat: initialize Go API with health endpoint"
```

---

### Task 2: Initialize Python fontforge worker

**Files:**
- Create: `services/fontforge/pyproject.toml`
- Create: `services/fontforge/requirements.txt`
- Create: `services/fontforge/worker.py`
- Create: `services/fontforge/tests/__init__.py`
- Create: `services/fontforge/tests/test_worker.py`

- [ ] **Step 1: Create pyproject.toml**

```toml
[project]
name = "napkin-fontforge"
version = "0.1.0"
requires-python = ">=3.11"

[tool.pytest.ini_options]
testpaths = ["tests"]
```

- [ ] **Step 2: Create requirements.txt**

```
fonttools>=4.47.0
Pillow>=10.2.0
redis>=5.0.0
pytest>=8.0.0
```

- [ ] **Step 3: Write a placeholder test**

Create `services/fontforge/tests/__init__.py` (empty).

Create `services/fontforge/tests/test_worker.py`:
```python
from worker import ping


def test_ping():
    assert ping() == "pong"
```

- [ ] **Step 4: Write minimal worker.py**

Create `services/fontforge/worker.py`:
```python
def ping() -> str:
    return "pong"


if __name__ == "__main__":
    print("fontforge worker starting...")
```

- [ ] **Step 5: Run test**

```bash
cd services/fontforge && pip install -r requirements.txt && python -m pytest tests/ -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add services/fontforge/
git commit -m "feat: initialize fontforge worker with placeholder"
```

---

### Task 3: Initialize Vue SPA with Vite

**Files:**
- Create: `web/` (via Vite scaffold)
- Modify: `web/package.json` (add vitest)
- Create: `web/src/__tests__/App.test.ts`

- [ ] **Step 1: Scaffold Vue project**

```bash
cd web && yarn create vite . --template vue-ts
yarn install
```

Note: If directory exists, say yes to overwrite.

- [ ] **Step 2: Add vitest and testing-library**

```bash
cd web && yarn add -D vitest @vue/test-utils happy-dom
```

Add to `web/vite.config.ts`:
```ts
/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'happy-dom',
  },
})
```

- [ ] **Step 3: Write a smoke test**

Create `web/src/__tests__/App.test.ts`:
```ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '../App.vue'

describe('App', () => {
  it('mounts without error', () => {
    const wrapper = mount(App)
    expect(wrapper.exists()).toBe(true)
  })
})
```

- [ ] **Step 4: Add test script to package.json**

In `web/package.json`, add under `"scripts"`:
```json
"test": "vitest run",
"test:watch": "vitest"
```

- [ ] **Step 5: Run test**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add web/
git commit -m "feat: initialize Vue 3 SPA with Vite and vitest"
```

---

### Task 4: Add Makefile and root .gitignore

**Files:**
- Create: `Makefile`
- Modify: `.gitignore`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: test test-api test-web test-fontforge dev

test: test-api test-web test-fontforge

test-api:
	cd services/api && go test ./... -v

test-web:
	cd web && yarn test

test-fontforge:
	cd services/fontforge && python -m pytest tests/ -v

dev:
	docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up --build
```

- [ ] **Step 2: Update .gitignore**

```gitignore
# Dependencies
node_modules/
web/node_modules/

# Build output
web/dist/
services/api/api

# Environment
.env
docker/.env

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Python
__pycache__/
*.pyc
.venv/
venv/

# Test
coverage/
.coverage

# Superpowers
.superpowers/
```

- [ ] **Step 3: Commit**

```bash
git add Makefile .gitignore
git commit -m "feat: add Makefile and update .gitignore"
```

---

### Task 5: Add GitHub Actions CI

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create CI workflow**

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test-api:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: cd services/api && go test ./... -v

  test-web:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd web && yarn install --frozen-lockfile
      - run: cd web && yarn test

  test-fontforge:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: cd services/fontforge && pip install -r requirements.txt
      - run: cd services/fontforge && python -m pytest tests/ -v
```

- [ ] **Step 2: Commit**

```bash
git add .github/
git commit -m "ci: add GitHub Actions workflow for all services"
```

---

### Task 6: Create CLAUDE.md

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: Create CLAUDE.md**

```markdown
# Napkin Notes

## Quick reference

- **Test all:** `make test`
- **Test API:** `cd services/api && go test ./... -v`
- **Test web:** `cd web && yarn test`
- **Test fontforge:** `cd services/fontforge && python -m pytest tests/ -v`
- **Dev environment:** `make dev` (requires Docker)

## Structure

- `services/api/` — Go REST API (Chi router)
- `services/fontforge/` — Python font generation worker
- `web/` — Vue 3 + TypeScript SPA (Vite, Pinia, Vue Router)
- `docker/` — Docker Compose files and Traefik config

## Conventions

- Go: standard library style, `testify` for assertions
- Vue: Composition API with `<script setup>`, Pinia stores
- Python: type hints, pytest
- Commits: conventional commits (feat/fix/ci/docs)
- Branches: `feat/phase-NN-description`
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add CLAUDE.md with project conventions"
```

---

### Task 7: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-01-scaffold
gh pr create --title "feat: project scaffold with Go API, Vue SPA, and Python worker" --body "## Summary
- Initialize Go API service with health endpoint and tests
- Initialize Vue 3 SPA with Vite, TypeScript, and vitest
- Initialize Python fontforge worker with pytest
- Add Makefile for unified test commands
- Add GitHub Actions CI for all three services
- Add CLAUDE.md with project conventions

## Test plan
- [ ] \`make test\` passes all three test suites
- [ ] CI workflow runs on PR

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
