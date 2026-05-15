# Phase 9: Rip-to-Delete Animation

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Implement the drag-to-rip gesture that tears a napkin in half to delete it.

**Branch:** `feat/phase-09-rip-animation`

---

## File Structure

```
web/src/
├── components/
│   ├── RipAnimation.vue
│   └── NapkinCard.vue (modified — add rip gesture)
├── composables/
│   └── useRipGesture.ts
├── views/
│   └── TrashView.vue (replace placeholder)
└── __tests__/
    └── useRipGesture.test.ts
```

---

### Task 1: Rip gesture composable

**Files:**
- Create: `web/src/composables/useRipGesture.ts`
- Create: `web/src/__tests__/useRipGesture.test.ts`

- [ ] **Step 1: Write rip gesture test**

Create `web/src/__tests__/useRipGesture.test.ts`:
```ts
import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useRipGesture } from '../composables/useRipGesture'

describe('useRipGesture', () => {
  it('starts inactive', () => {
    const el = ref(null)
    const { isRipping, progress } = useRipGesture(el, { onRip: () => {} })
    expect(isRipping.value).toBe(false)
    expect(progress.value).toBe(0)
  })

  it('calculates progress from drag distance', () => {
    const el = ref(null)
    const { progress, simulateDrag } = useRipGesture(el, {
      onRip: () => {},
      threshold: 0.4,
    })

    simulateDrag(0.5)
    expect(progress.value).toBe(0.5)
  })

  it('triggers onRip when past threshold', () => {
    const el = ref(null)
    let ripped = false
    const { simulateDrag, simulateRelease } = useRipGesture(el, {
      onRip: () => { ripped = true },
      threshold: 0.4,
    })

    simulateDrag(0.5)
    simulateRelease()
    expect(ripped).toBe(true)
  })

  it('does not rip when below threshold', () => {
    const el = ref(null)
    let ripped = false
    const { simulateDrag, simulateRelease } = useRipGesture(el, {
      onRip: () => { ripped = true },
      threshold: 0.4,
    })

    simulateDrag(0.2)
    simulateRelease()
    expect(ripped).toBe(false)
  })
})
```

- [ ] **Step 2: Implement rip gesture composable**

Create `web/src/composables/useRipGesture.ts`:
```ts
import { ref, onMounted, onUnmounted, type Ref } from 'vue'

interface RipGestureOptions {
  onRip: () => void
  threshold?: number
}

export function useRipGesture(
  elementRef: Ref<HTMLElement | null>,
  options: RipGestureOptions
) {
  const isRipping = ref(false)
  const progress = ref(0)
  const threshold = options.threshold ?? 0.4

  let startX = 0
  let elementWidth = 0
  let isDragging = false

  function onPointerDown(e: PointerEvent) {
    if (!elementRef.value) return
    isDragging = true
    isRipping.value = true
    startX = e.clientX
    elementWidth = elementRef.value.offsetWidth
    elementRef.value.setPointerCapture(e.pointerId)
  }

  function onPointerMove(e: PointerEvent) {
    if (!isDragging || !elementWidth) return
    const dx = Math.abs(e.clientX - startX)
    progress.value = Math.min(dx / elementWidth, 1)
  }

  function onPointerUp() {
    if (!isDragging) return
    isDragging = false

    if (progress.value >= threshold) {
      options.onRip()
    }

    isRipping.value = false
    progress.value = 0
  }

  function simulateDrag(pct: number) {
    isRipping.value = true
    progress.value = pct
  }

  function simulateRelease() {
    if (progress.value >= threshold) {
      options.onRip()
    }
    isRipping.value = false
    progress.value = 0
  }

  onMounted(() => {
    const el = elementRef.value
    if (!el) return
    el.addEventListener('pointerdown', onPointerDown)
    el.addEventListener('pointermove', onPointerMove)
    el.addEventListener('pointerup', onPointerUp)
    el.addEventListener('pointercancel', onPointerUp)
  })

  onUnmounted(() => {
    const el = elementRef.value
    if (!el) return
    el.removeEventListener('pointerdown', onPointerDown)
    el.removeEventListener('pointermove', onPointerMove)
    el.removeEventListener('pointerup', onPointerUp)
    el.removeEventListener('pointercancel', onPointerUp)
  })

  return { isRipping, progress, simulateDrag, simulateRelease }
}
```

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/composables/useRipGesture.ts web/src/__tests__/useRipGesture.test.ts
git commit -m "feat: add rip gesture composable with threshold detection"
```

---

### Task 2: RipAnimation component

**Files:**
- Create: `web/src/components/RipAnimation.vue`

- [ ] **Step 1: Create RipAnimation component**

Create `web/src/components/RipAnimation.vue`:
```vue
<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  progress: number
  isActive: boolean
}>()

const emit = defineEmits<{
  complete: []
}>()

const tearPath = computed(() => {
  const p = props.progress
  const points: string[] = []
  const segments = 20
  for (let i = 0; i <= segments; i++) {
    const y = (i / segments) * 100
    const jitter = Math.sin(i * 2.7) * 3 + Math.cos(i * 4.1) * 2
    const x = 50 + jitter
    points.push(`${x}% ${y}%`)
  }
  return points.join(', ')
})

const leftClip = computed(() => {
  if (!props.isActive) return 'none'
  const points = tearPath.value.split(', ').map((p) => {
    const [x, y] = p.split(' ')
    return `${x} ${y}`
  })
  return `polygon(0% 0%, ${points.join(', ')}, 0% 100%)`
})

const rightClip = computed(() => {
  if (!props.isActive) return 'none'
  const points = tearPath.value.split(', ').map((p) => {
    const [x, y] = p.split(' ')
    return `${x} ${y}`
  })
  return `polygon(${points.join(', ')}, 100% 100%, 100% 0%)`
})

const leftTransform = computed(() => {
  const offset = props.progress * -30
  const rotate = props.progress * -2
  return `translateX(${offset}px) rotate(${rotate}deg)`
})

const rightTransform = computed(() => {
  const offset = props.progress * 30
  const rotate = props.progress * 2
  return `translateX(${offset}px) rotate(${rotate}deg)`
})

const opacity = computed(() => {
  if (props.progress > 0.8) return 1 - (props.progress - 0.8) * 5
  return 1
})
</script>

<template>
  <div class="rip-container" :class="{ active: isActive }">
    <div
      v-if="isActive"
      class="rip-half left"
      :style="{
        clipPath: leftClip,
        transform: leftTransform,
        opacity: opacity,
      }"
    >
      <slot />
    </div>
    <div
      v-if="isActive"
      class="rip-half right"
      :style="{
        clipPath: rightClip,
        transform: rightTransform,
        opacity: opacity,
      }"
    >
      <slot />
    </div>
    <div v-if="!isActive">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.rip-container {
  position: relative;
}
.rip-half {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  transition: transform 0.1s ease-out, opacity 0.2s ease;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/RipAnimation.vue
git commit -m "feat: add RipAnimation component with procedural tear"
```

---

### Task 3: Integrate rip gesture into NapkinCard

**Files:**
- Modify: `web/src/components/NapkinCard.vue`

- [ ] **Step 1: Update NapkinCard with rip gesture**

Replace `web/src/components/NapkinCard.vue`:
```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import NapkinTexture from './NapkinTexture.vue'
import RipAnimation from './RipAnimation.vue'
import { useRipGesture } from '../composables/useRipGesture'
import type { Note } from '../stores/notesStore'

const props = defineProps<{
  note: Note
}>()

const emit = defineEmits<{
  open: [id: string]
  rip: [id: string]
}>()

const cardRef = ref<HTMLElement | null>(null)

const { isRipping, progress } = useRipGesture(cardRef, {
  onRip: () => emit('rip', props.note.id),
  threshold: 0.4,
})

const rotation = computed(() => {
  const hash = props.note.id.split('').reduce((a, c) => a + c.charCodeAt(0), 0)
  return ((hash % 7) - 3)
})

const preview = computed(() => {
  const text = props.note.content
  return text.length > 80 ? text.slice(0, 80) + '...' : text
})

function handleClick() {
  if (!isRipping.value) {
    emit('open', props.note.id)
  }
}
</script>

<template>
  <div
    ref="cardRef"
    class="napkin-card"
    :style="{ transform: `rotate(${rotation}deg)` }"
    @click="handleClick"
  >
    <RipAnimation :progress="progress" :is-active="isRipping">
      <NapkinTexture width="220px" height="200px">
        <p class="napkin-content">{{ preview }}</p>
      </NapkinTexture>
    </RipAnimation>
  </div>
</template>

<style scoped>
.napkin-card {
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  touch-action: none;
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

- [ ] **Step 2: Update GalleryView to handle rip event**

Add to GalleryView.vue template, updating NapkinCard:
```vue
<NapkinCard
  v-for="note in store.notes"
  :key="note.id"
  :note="note"
  @open="openNote"
  @rip="handleRip"
/>
```

Add to script:
```ts
async function handleRip(id: string) {
  await store.deleteNote(id)
}
```

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/components/NapkinCard.vue web/src/views/GalleryView.vue
git commit -m "feat: integrate rip-to-delete gesture into napkin cards"
```

---

### Task 4: Trash view

**Files:**
- Modify: `web/src/views/TrashView.vue`
- Modify: `web/src/stores/notesStore.ts` (add trash methods)

- [ ] **Step 1: Add trash methods to notes store**

Add to `web/src/stores/notesStore.ts`:
```ts
  const trashedNotes = ref<Note[]>([])

  async function fetchTrashed() {
    const { data } = await api.get('/notes/trash')
    trashedNotes.value = data || []
  }

  async function restoreNote(id: string) {
    await api.post(`/notes/${id}/restore`)
    trashedNotes.value = trashedNotes.value.filter((n) => n.id !== id)
  }
```

Return them from the store function.

- [ ] **Step 2: Implement TrashView**

Replace `web/src/views/TrashView.vue`:
```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'

const store = useNotesStore()

onMounted(() => {
  store.fetchTrashed()
})

async function restore(id: string) {
  await store.restoreNote(id)
}
</script>

<template>
  <div class="trash-container">
    <h1>Trash</h1>
    <p class="subtitle">Ripped napkins — tape them back together to restore.</p>

    <div v-if="store.trashedNotes.length === 0" class="empty-state">
      <p>No ripped napkins.</p>
    </div>

    <div v-else class="trash-grid">
      <div v-for="note in store.trashedNotes" :key="note.id" class="trashed-card">
        <NapkinTexture width="220px" height="200px">
          <div class="tape-overlay"></div>
          <p class="napkin-content">{{ note.content }}</p>
        </NapkinTexture>
        <button class="restore-btn" @click="restore(note.id)">
          Restore
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.trash-container {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
.subtitle {
  opacity: 0.6;
  margin-bottom: 2rem;
}
.trash-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 2rem;
}
.trashed-card {
  position: relative;
  opacity: 0.7;
}
.tape-overlay {
  position: absolute;
  top: 50%;
  left: 10%;
  right: 10%;
  height: 20px;
  background: rgba(200, 180, 100, 0.4);
  transform: rotate(-2deg) translateY(-50%);
  border-radius: 2px;
}
.napkin-content {
  padding: 1.5rem;
  font-family: 'Caveat', cursive;
  font-size: 1rem;
  color: #2D2D2D;
}
.restore-btn {
  display: block;
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.5rem;
  background: transparent;
  border: 1px solid #5C3D2E;
  color: #5C3D2E;
  border-radius: 4px;
  cursor: pointer;
}
.empty-state {
  text-align: center;
  padding: 4rem;
  opacity: 0.6;
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
git add web/src/views/TrashView.vue web/src/stores/notesStore.ts
git commit -m "feat: add Trash view with restore and taped-together style"
```

---

### Task 5: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-09-rip-animation
gh pr create --title "feat: rip-to-delete animation with drag gesture" --body "## Summary
- Add useRipGesture composable with threshold-based tear detection
- Add RipAnimation component with procedural jagged tear path
- Integrate rip gesture into NapkinCard (drag to rip, click to open)
- Implement Trash view showing taped-together napkins with restore

## Test plan
- [ ] \`cd web && yarn test\` passes
- [ ] Dragging napkin past threshold triggers rip animation
- [ ] Releasing before threshold cancels gesture
- [ ] Ripped notes appear in Trash view
- [ ] Restore moves note back to gallery

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
