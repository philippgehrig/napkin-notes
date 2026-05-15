# Phase 8: Gallery & Editor Views

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Build the napkin gallery (scattered cards) and the note editor with font rendering.

**Branch:** `feat/phase-08-gallery-editor`

---

## File Structure

```
web/src/
├── stores/
│   └── notesStore.ts
├── views/
│   ├── GalleryView.vue (replace placeholder)
│   └── NoteEditorView.vue (replace placeholder)
├── components/
│   ├── NapkinCard.vue
│   └── NapkinTexture.vue
├── assets/
│   ├── textures/
│   │   └── napkin.png       (placeholder CC0 texture)
│   └── fonts/
│       └── Caveat.woff2     (placeholder handwriting font)
└── __tests__/
    ├── notesStore.test.ts
    └── NapkinCard.test.ts
```

---

### Task 1: Notes store

**Files:**
- Create: `web/src/stores/notesStore.ts`
- Create: `web/src/__tests__/notesStore.test.ts`

- [ ] **Step 1: Write notes store test**

Create `web/src/__tests__/notesStore.test.ts`:
```ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNotesStore } from '../stores/notesStore'

vi.mock('../api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('notesStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with empty notes', () => {
    const store = useNotesStore()
    expect(store.notes).toEqual([])
  })

  it('fetches notes from API', async () => {
    const { default: api } = await import('../api/client')
    vi.mocked(api.get).mockResolvedValueOnce({
      data: [
        { id: '1', content: 'Hello', user_id: 'u1', created_at: '2026-01-01' },
      ],
    })

    const store = useNotesStore()
    await store.fetchNotes()

    expect(store.notes).toHaveLength(1)
    expect(store.notes[0].content).toBe('Hello')
  })

  it('creates a note', async () => {
    const { default: api } = await import('../api/client')
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { id: '2', content: 'New note', user_id: 'u1', created_at: '2026-01-01' },
    })

    const store = useNotesStore()
    await store.createNote('New note')

    expect(store.notes).toHaveLength(1)
    expect(store.notes[0].content).toBe('New note')
  })

  it('deletes a note (soft)', async () => {
    const { default: api } = await import('../api/client')
    vi.mocked(api.delete).mockResolvedValueOnce({})

    const store = useNotesStore()
    store.notes = [{ id: '1', content: 'To delete', user_id: 'u1', created_at: '2026-01-01', updated_at: '2026-01-01' }]

    await store.deleteNote('1')

    expect(store.notes).toHaveLength(0)
  })
})
```

- [ ] **Step 2: Implement notes store**

Create `web/src/stores/notesStore.ts`:
```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api/client'

export interface Note {
  id: string
  user_id: string
  content: string
  font_id?: string
  deleted_at?: string
  created_at: string
  updated_at: string
}

export const useNotesStore = defineStore('notes', () => {
  const notes = ref<Note[]>([])
  const loading = ref(false)

  async function fetchNotes() {
    loading.value = true
    try {
      const { data } = await api.get('/notes')
      notes.value = data || []
    } finally {
      loading.value = false
    }
  }

  async function createNote(content: string, fontId?: string) {
    const { data } = await api.post('/notes', { content, font_id: fontId })
    notes.value.unshift(data)
    return data
  }

  async function updateNote(id: string, content: string, fontId?: string) {
    const { data } = await api.put(`/notes/${id}`, { content, font_id: fontId })
    const idx = notes.value.findIndex((n) => n.id === id)
    if (idx !== -1) notes.value[idx] = data
    return data
  }

  async function deleteNote(id: string) {
    await api.delete(`/notes/${id}`)
    notes.value = notes.value.filter((n) => n.id !== id)
  }

  return { notes, loading, fetchNotes, createNote, updateNote, deleteNote }
})
```

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/stores/notesStore.ts web/src/__tests__/notesStore.test.ts
git commit -m "feat: add Pinia notes store with CRUD operations"
```

---

### Task 2: NapkinTexture component

**Files:**
- Create: `web/src/components/NapkinTexture.vue`

- [ ] **Step 1: Create NapkinTexture**

Create `web/src/components/NapkinTexture.vue`:
```vue
<script setup lang="ts">
defineProps<{
  width?: string
  height?: string
}>()
</script>

<template>
  <div
    class="napkin-texture"
    :style="{ width: width || '100%', height: height || '100%' }"
  >
    <slot />
  </div>
</template>

<style scoped>
.napkin-texture {
  background-color: #FFF8E7;
  background-image:
    radial-gradient(ellipse at 20% 50%, rgba(92, 61, 46, 0.03) 0%, transparent 50%),
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 28px,
      rgba(92, 61, 46, 0.04) 28px,
      rgba(92, 61, 46, 0.04) 29px
    );
  border-radius: 2px;
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.12),
    0 1px 2px rgba(0, 0, 0, 0.08);
  position: relative;
  overflow: hidden;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/NapkinTexture.vue
git commit -m "feat: add NapkinTexture component with CSS-based texture"
```

---

### Task 3: NapkinCard component

**Files:**
- Create: `web/src/components/NapkinCard.vue`
- Create: `web/src/__tests__/NapkinCard.test.ts`

- [ ] **Step 1: Write NapkinCard test**

Create `web/src/__tests__/NapkinCard.test.ts`:
```ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NapkinCard from '../components/NapkinCard.vue'

describe('NapkinCard', () => {
  it('renders note content', () => {
    const wrapper = mount(NapkinCard, {
      props: {
        note: {
          id: '1',
          content: 'Hello world',
          user_id: 'u1',
          created_at: '2026-01-01',
          updated_at: '2026-01-01',
        },
      },
    })
    expect(wrapper.text()).toContain('Hello world')
  })

  it('applies random rotation', () => {
    const wrapper = mount(NapkinCard, {
      props: {
        note: {
          id: '1',
          content: 'Test',
          user_id: 'u1',
          created_at: '2026-01-01',
          updated_at: '2026-01-01',
        },
      },
    })
    const el = wrapper.find('.napkin-card')
    const transform = el.attributes('style')
    expect(transform).toContain('rotate')
  })

  it('emits click event', async () => {
    const wrapper = mount(NapkinCard, {
      props: {
        note: {
          id: '1',
          content: 'Test',
          user_id: 'u1',
          created_at: '2026-01-01',
          updated_at: '2026-01-01',
        },
      },
    })
    await wrapper.trigger('click')
    expect(wrapper.emitted('open')).toHaveLength(1)
  })
})
```

- [ ] **Step 2: Implement NapkinCard**

Create `web/src/components/NapkinCard.vue`:
```vue
<script setup lang="ts">
import { computed } from 'vue'
import NapkinTexture from './NapkinTexture.vue'
import type { Note } from '../stores/notesStore'

const props = defineProps<{
  note: Note
}>()

const emit = defineEmits<{
  open: [id: string]
}>()

const rotation = computed(() => {
  const hash = props.note.id.split('').reduce((a, c) => a + c.charCodeAt(0), 0)
  return ((hash % 7) - 3)
})

const preview = computed(() => {
  const text = props.note.content
  return text.length > 80 ? text.slice(0, 80) + '...' : text
})
</script>

<template>
  <div
    class="napkin-card"
    :style="{ transform: `rotate(${rotation}deg)` }"
    @click="emit('open', note.id)"
  >
    <NapkinTexture width="220px" height="200px">
      <p class="napkin-content">{{ preview }}</p>
    </NapkinTexture>
  </div>
</template>

<style scoped>
.napkin-card {
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.napkin-card:hover {
  transform: rotate(0deg) scale(1.05) !important;
  z-index: 10;
}
.napkin-content {
  padding: 1.5rem;
  font-family: 'Caveat', cursive;
  font-size: 1.1rem;
  line-height: 1.4;
  color: #2D2D2D;
  word-break: break-word;
}
</style>
```

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/components/NapkinCard.vue web/src/__tests__/NapkinCard.test.ts
git commit -m "feat: add NapkinCard component with rotation and preview"
```

---

### Task 4: Gallery view

**Files:**
- Modify: `web/src/views/GalleryView.vue`

- [ ] **Step 1: Implement GalleryView**

Replace `web/src/views/GalleryView.vue`:
```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinCard from '../components/NapkinCard.vue'

const store = useNotesStore()
const router = useRouter()

onMounted(() => {
  store.fetchNotes()
})

function openNote(id: string) {
  router.push({ name: 'editor', params: { id } })
}

function createNewNote() {
  router.push({ name: 'editor' })
}
</script>

<template>
  <div class="gallery-container">
    <div class="gallery-header">
      <h1>Your Napkins</h1>
      <button class="new-note-btn" @click="createNewNote">+ New Napkin</button>
    </div>

    <div v-if="store.loading" class="loading">Loading...</div>

    <div v-else-if="store.notes.length === 0" class="empty-state">
      <p>No napkins yet. Write your first note!</p>
      <button @click="createNewNote">Create Napkin</button>
    </div>

    <div v-else class="gallery-grid">
      <NapkinCard
        v-for="note in store.notes"
        :key="note.id"
        :note="note"
        @open="openNote"
      />
    </div>
  </div>
</template>

<style scoped>
.gallery-container {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
.gallery-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}
.new-note-btn {
  padding: 0.75rem 1.5rem;
  background: #5C3D2E;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}
.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 2rem;
  padding: 1rem;
}
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}
.empty-state button {
  margin-top: 1rem;
  padding: 0.75rem 1.5rem;
  background: #5C3D2E;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.loading {
  text-align: center;
  padding: 4rem;
  opacity: 0.6;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/views/GalleryView.vue
git commit -m "feat: implement Gallery view with napkin grid"
```

---

### Task 5: Note editor view

**Files:**
- Modify: `web/src/views/NoteEditorView.vue`

- [ ] **Step 1: Implement NoteEditorView**

Replace `web/src/views/NoteEditorView.vue`:
```vue
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'

const route = useRoute()
const router = useRouter()
const store = useNotesStore()

const content = ref('')
const saving = ref(false)
const noteId = computed(() => route.params.id as string | undefined)
const isNew = computed(() => !noteId.value)

onMounted(async () => {
  if (noteId.value) {
    const existingNote = store.notes.find((n) => n.id === noteId.value)
    if (existingNote) {
      content.value = existingNote.content
    } else {
      await store.fetchNotes()
      const note = store.notes.find((n) => n.id === noteId.value)
      if (note) content.value = note.content
    }
  }
})

async function save() {
  if (!content.value.trim()) return
  saving.value = true
  try {
    if (isNew.value) {
      const note = await store.createNote(content.value)
      router.replace({ name: 'editor', params: { id: note.id } })
    } else {
      await store.updateNote(noteId.value!, content.value)
    }
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.push({ name: 'gallery' })
}
</script>

<template>
  <div class="editor-container">
    <div class="editor-toolbar">
      <button @click="goBack" class="back-btn">← Back</button>
      <button @click="save" :disabled="saving" class="save-btn">
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
    </div>

    <div class="editor-napkin">
      <NapkinTexture width="100%" height="100%">
        <textarea
          v-model="content"
          class="napkin-textarea"
          placeholder="Write your note..."
          autofocus
        />
      </NapkinTexture>
    </div>
  </div>
</template>

<style scoped>
.editor-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  padding: 1rem;
}
.editor-toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
}
.back-btn, .save-btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}
.back-btn {
  background: transparent;
  border: 1px solid #5C3D2E;
  color: #5C3D2E;
}
.save-btn {
  background: #5C3D2E;
  color: white;
}
.save-btn:disabled {
  opacity: 0.6;
}
.editor-napkin {
  flex: 1;
  max-width: 600px;
  margin: 0 auto;
  width: 100%;
}
.napkin-textarea {
  width: 100%;
  height: 100%;
  min-height: 400px;
  border: none;
  background: transparent;
  resize: none;
  padding: 2rem;
  font-family: 'Caveat', cursive;
  font-size: 1.4rem;
  line-height: 1.6;
  color: #2D2D2D;
  outline: none;
}
.napkin-textarea::placeholder {
  color: #999;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/views/NoteEditorView.vue
git commit -m "feat: implement Note editor with napkin textarea"
```

---

### Task 6: Add placeholder font CSS

**Files:**
- Modify: `web/index.html` or create `web/src/assets/fonts.css`

- [ ] **Step 1: Add font import to index.html**

Add to `<head>` in `web/index.html`:
```html
<link href="https://fonts.googleapis.com/css2?family=Caveat:wght@400;700&display=swap" rel="stylesheet">
```

- [ ] **Step 2: Commit**

```bash
git add web/index.html
git commit -m "feat: add Caveat placeholder handwriting font"
```

---

### Task 7: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-08-gallery-editor
gh pr create --title "feat: gallery view and note editor with napkin styling" --body "## Summary
- Add Pinia notes store with CRUD operations
- Add NapkinTexture and NapkinCard components
- Implement Gallery view with scattered napkin grid
- Implement Note editor with napkin-textured textarea
- Add Caveat placeholder handwriting font

## Test plan
- [ ] \`cd web && yarn test\` passes
- [ ] Gallery displays napkin cards from store
- [ ] Cards rotate slightly for scattered aesthetic
- [ ] Clicking card navigates to editor
- [ ] Editor saves new and existing notes

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
