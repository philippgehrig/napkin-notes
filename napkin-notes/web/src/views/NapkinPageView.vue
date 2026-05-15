<template>
  <div class="napkin-page">
    <div class="napkin-page__container" @click="focusEditor">
      <NapkinTexture width="100%" height="100%" :variant="napkinVariant" class="napkin-page__napkin">
        <textarea
          ref="textareaRef"
          v-model="content"
          class="napkin-page__input"
          placeholder="Write on your napkin..."
          @input="onInput"
        />
      </NapkinTexture>
    </div>

    <div class="napkin-page__footer">
      <button class="napkin-page__new-btn" @click="newNapkin">New Napkin</button>
      <button class="napkin-page__discard-btn" @click="discardNapkin">Discard</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'

const route = useRoute()
const router = useRouter()
const store = useNotesStore()

const content = ref('')
const lastValidContent = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const currentNoteId = ref<string | null>(null)
const napkinVariant = ref(Math.floor(Math.random() * 3) + 1)

let autoSaveTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  const id = route.params.id as string | undefined
  if (id) {
    currentNoteId.value = id
    const existing = store.notes.find((n) => n.id === id)
    if (existing) {
      content.value = existing.content
      napkinVariant.value = existing.texture_variant || 1
    } else {
      await store.fetchNotes()
      const note = store.notes.find((n) => n.id === id)
      if (note) {
        content.value = note.content
        napkinVariant.value = note.texture_variant || 1
      }
    }
  }
  lastValidContent.value = content.value
  textareaRef.value?.focus()
})

onUnmounted(() => {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
})

watch(content, () => {
  scheduleAutoSave()
})

function scheduleAutoSave() {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    autoSave()
  }, 1000)
}

async function autoSave() {
  if (!content.value.trim()) return

  if (currentNoteId.value) {
    await store.updateNote(currentNoteId.value, content.value, napkinVariant.value)
  } else {
    const note = await store.createNote(content.value, napkinVariant.value)
    currentNoteId.value = note.id
    router.replace({ name: 'napkin-edit', params: { id: note.id } })
  }
}

function focusEditor() {
  textareaRef.value?.focus()
}

function onInput() {
  const el = textareaRef.value
  if (!el) return
  if (el.scrollHeight > el.clientHeight) {
    content.value = lastValidContent.value
  } else {
    lastValidContent.value = content.value
  }
}

function newNapkin() {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  content.value = ''
  lastValidContent.value = ''
  currentNoteId.value = null
  napkinVariant.value = Math.floor(Math.random() * 3) + 1
  router.replace({ name: 'napkin' })
  textareaRef.value?.focus()
}

async function discardNapkin() {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  if (currentNoteId.value) {
    await store.deleteNote(currentNoteId.value)
  }
  content.value = ''
  lastValidContent.value = ''
  currentNoteId.value = null
  napkinVariant.value = Math.floor(Math.random() * 3) + 1
  router.replace({ name: 'napkin' })
  textareaRef.value?.focus()
}
</script>

<style scoped>
.napkin-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 50px);
  padding: 2rem;
  box-sizing: border-box;
}

.napkin-page__container {
  width: 100%;
  max-width: 700px;
  cursor: text;
}

.napkin-page__napkin {
  width: 100%;
}

.napkin-page__input {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  font-family: var(--handwriting);
  font-size: 1.5rem;
  color: #2D2D2D;
  padding: 1rem;
  box-sizing: border-box;
  line-height: 1.8;
  overflow: hidden;
  pointer-events: auto;
}

.napkin-page__input::placeholder {
  color: rgba(92, 61, 46, 0.4);
}

.napkin-page__footer {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-top: 1.5rem;
}

.napkin-page__new-btn {
  background-color: #5C3D2E;
  color: #FFF8E7;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-family: var(--handwriting);
  font-size: 1.2rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.napkin-page__new-btn:hover {
  background-color: #3d2820;
}

.napkin-page__discard-btn {
  background-color: #8B2020;
  color: #FFF8E7;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-family: var(--handwriting);
  font-size: 1.2rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.napkin-page__discard-btn:hover {
  background-color: #5c1515;
}

@media (max-width: 600px) and (orientation: portrait) {
  .napkin-page {
    padding: 1rem;
    height: calc(100vh - 50px);
    justify-content: flex-start;
    padding-top: 15vh;
  }

  .napkin-page__container {
    max-width: 100%;
    width: 100%;
  }

  .napkin-page__input {
    font-size: 1.2rem;
    padding: 0.75rem;
  }

  .napkin-page__footer {
    margin-top: 1rem;
    gap: 1rem;
  }

  .napkin-page__new-btn,
  .napkin-page__discard-btn {
    padding: 0.6rem 1.2rem;
    font-size: 1rem;
  }
}

@media (max-width: 600px) and (orientation: landscape) {
  .napkin-page {
    padding: 1rem;
  }

  .napkin-page__container {
    max-width: 100%;
  }

  .napkin-page__input {
    font-size: 1.2rem;
    padding: 0.75rem;
  }
}
</style>
