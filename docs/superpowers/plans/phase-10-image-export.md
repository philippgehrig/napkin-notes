# Phase 10: Image Export

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Server-side PNG rendering of napkins for download/sharing.

**Branch:** `feat/phase-10-image-export`

---

## File Structure

```
services/api/internal/
├── export/
│   ├── renderer.go
│   ├── renderer_test.go
│   ├── handler.go
│   └── handler_test.go
└── (existing)
web/src/
├── components/
│   └── ExportButton.vue
└── views/
    └── NoteEditorView.vue (modified)
```

---

### Task 1: Image renderer

**Files:**
- Create: `services/api/internal/export/renderer.go`
- Create: `services/api/internal/export/renderer_test.go`

- [ ] **Step 1: Add gg dependency**

```bash
cd services/api && go get github.com/fogleman/gg golang.org/x/image/font golang.org/x/image/font/opentype
```

- [ ] **Step 2: Write renderer test**

Create `services/api/internal/export/renderer_test.go`:
```go
package export

import (
	"image/png"
	"os"
	"testing"
)

func TestRenderNapkin_CreatesPNG(t *testing.T) {
	opts := RenderOptions{
		Content:  "Hello, World!",
		Width:    800,
		Height:   600,
		FontSize: 32,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 600 {
		t.Errorf("expected 800x600, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRenderNapkin_SavesPNG(t *testing.T) {
	opts := RenderOptions{
		Content:  "Test napkin content\nLine two",
		Width:    800,
		Height:   600,
		FontSize: 28,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	f, err := os.CreateTemp("", "napkin-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("PNG encode failed: %v", err)
	}

	info, _ := f.Stat()
	if info.Size() < 100 {
		t.Error("PNG file suspiciously small")
	}
}

func TestRenderNapkin_EmptyContent(t *testing.T) {
	opts := RenderOptions{
		Content:  "",
		Width:    800,
		Height:   600,
		FontSize: 28,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}
```

- [ ] **Step 3: Implement renderer**

Create `services/api/internal/export/renderer.go`:
```go
package export

import (
	"image"
	"image/color"
	"strings"

	"github.com/fogleman/gg"
)

type RenderOptions struct {
	Content  string
	Width    int
	Height   int
	FontSize float64
	FontPath string
}

func RenderNapkin(opts RenderOptions) (image.Image, error) {
	dc := gg.NewContext(opts.Width, opts.Height)

	// Napkin background
	dc.SetColor(color.RGBA{255, 248, 231, 255}) // #FFF8E7
	dc.Clear()

	// Subtle texture lines
	dc.SetColor(color.RGBA{92, 61, 46, 10})
	for y := 30; y < opts.Height; y += 30 {
		dc.DrawLine(0, float64(y), float64(opts.Width), float64(y))
		dc.SetLineWidth(0.5)
		dc.Stroke()
	}

	// Shadow/border
	dc.SetColor(color.RGBA{0, 0, 0, 20})
	dc.DrawRectangle(0, 0, float64(opts.Width), float64(opts.Height))
	dc.SetLineWidth(2)
	dc.Stroke()

	// Text
	fontSize := opts.FontSize
	if fontSize == 0 {
		fontSize = 28
	}

	if opts.FontPath != "" {
		if err := dc.LoadFontFace(opts.FontPath, fontSize); err != nil {
			// Fall back to default
			dc.LoadFontFace("", fontSize)
		}
	}

	dc.SetColor(color.RGBA{45, 45, 45, 255}) // #2D2D2D

	padding := 40.0
	maxWidth := float64(opts.Width) - padding*2

	lines := strings.Split(opts.Content, "\n")
	y := padding + fontSize
	lineHeight := fontSize * 1.6

	for _, line := range lines {
		wrappedLines := dc.WordWrap(line, maxWidth)
		for _, wl := range wrappedLines {
			if y > float64(opts.Height)-padding {
				break
			}
			dc.DrawString(wl, padding, y)
			y += lineHeight
		}
	}

	return dc.Image(), nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd services/api && go test ./internal/export/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/export/ services/api/go.mod services/api/go.sum
git commit -m "feat: add napkin image renderer with gg"
```

---

### Task 2: Export HTTP handler

**Files:**
- Create: `services/api/internal/export/handler.go`
- Create: `services/api/internal/export/handler_test.go`

- [ ] **Step 1: Write handler test**

Create `services/api/internal/export/handler_test.go`:
```go
package export

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/notes"
)

type mockNoteGetter struct{}

func (m *mockNoteGetter) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	if id == "note-1" {
		return &models.Note{
			ID:      "note-1",
			UserID:  "user-1",
			Content: "Hello export!",
		}, nil
	}
	return nil, notes.ErrNoteNotFound
}

func TestExportHandler_Success(t *testing.T) {
	h := NewHandler(&mockNoteGetter{})

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, "user-1")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/api/notes/{id}/export", h.Export)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/note-1/export?format=png", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() < 100 {
		t.Error("response body too small for a PNG")
	}
}

func TestExportHandler_NotFound(t *testing.T) {
	h := NewHandler(&mockNoteGetter{})

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, "user-1")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/api/notes/{id}/export", h.Export)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/nonexistent/export?format=png", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Implement handler**

Create `services/api/internal/export/handler.go`:
```go
package export

import (
	"context"
	"encoding/json"
	"image/png"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type NoteGetter interface {
	GetByID(ctx context.Context, id, userID string) (*models.Note, error)
}

type Handler struct {
	notes NoteGetter
}

func NewHandler(notes NoteGetter) *Handler {
	return &Handler{notes: notes}
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.notes.GetByID(r.Context(), noteID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "note not found"})
		return
	}

	img, err := RenderNapkin(RenderOptions{
		Content:  note.Content,
		Width:    800,
		Height:   600,
		FontSize: 28,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "render failed"})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=napkin.png")
	w.WriteHeader(http.StatusOK)
	png.Encode(w, img)
}
```

- [ ] **Step 3: Run tests**

```bash
cd services/api && go test ./internal/export/ -v
```
Expected: PASS

- [ ] **Step 4: Wire export route in main.go**

Add inside the `/api/notes` route group:
```go
r.Get("/{id}/export", exportHandler.Export)
```

Where `exportHandler := export.NewHandler(noteSvc)`.

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/export/ services/api/main.go
git commit -m "feat: add image export endpoint for napkin PNG rendering"
```

---

### Task 3: Frontend export button

**Files:**
- Create: `web/src/components/ExportButton.vue`
- Modify: `web/src/views/NoteEditorView.vue`

- [ ] **Step 1: Create ExportButton component**

Create `web/src/components/ExportButton.vue`:
```vue
<script setup lang="ts">
import api from '../api/client'

const props = defineProps<{
  noteId: string
}>()

const loading = ref(false)

import { ref } from 'vue'

async function exportPng() {
  loading.value = true
  try {
    const response = await api.get(`/notes/${props.noteId}/export?format=png`, {
      responseType: 'blob',
    })
    const url = URL.createObjectURL(response.data)
    const a = document.createElement('a')
    a.href = url
    a.download = 'napkin.png'
    a.click()
    URL.revokeObjectURL(url)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <button @click="exportPng" :disabled="loading" class="export-btn">
    {{ loading ? 'Exporting...' : 'Export PNG' }}
  </button>
</template>

<style scoped>
.export-btn {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid #5C3D2E;
  color: #5C3D2E;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}
.export-btn:disabled {
  opacity: 0.6;
}
</style>
```

- [ ] **Step 2: Add to NoteEditorView toolbar**

In `NoteEditorView.vue`, add after the Save button:
```vue
<ExportButton v-if="!isNew" :note-id="noteId!" />
```

Import ExportButton at the top of script.

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/ExportButton.vue web/src/views/NoteEditorView.vue
git commit -m "feat: add PNG export button to note editor"
```

---

### Task 4: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-10-image-export
gh pr create --title "feat: server-side napkin PNG export" --body "## Summary
- Add Go image renderer using fogleman/gg with napkin texture
- Add export HTTP handler at GET /api/notes/:id/export?format=png
- Add ExportButton component in Vue editor view
- Export triggers file download as napkin.png

## Test plan
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] GET /api/notes/:id/export returns valid PNG
- [ ] PNG has napkin background with text rendered
- [ ] Frontend button triggers download

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
