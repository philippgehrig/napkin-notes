<template>
  <div class="napkin-page">
    <div class="napkin-page__container" @click="focusEditor">
      <NapkinTexture width="100%" height="100%" class="napkin-page__napkin">
        <textarea
          ref="textareaRef"
          v-model="content"
          class="napkin-page__input"
          placeholder="Write on your napkin..."
          :maxlength="CHAR_LIMIT"
          @input="onInput"
        />
      </NapkinTexture>
    </div>

    <div class="napkin-page__footer">
      <span class="napkin-page__char-count">{{ content.length }} / {{ CHAR_LIMIT }}</span>
      <button class="napkin-page__new-btn" @click="newNapkin">New Napkin</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'

const CHAR_LIMIT = 500

const route = useRoute()
const router = useRouter()
const store = useNotesStore()

const content = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const currentNoteId = ref<string | null>(null)

onMounted(async () => {
  const id = route.params.id as string | undefined
  if (id) {
    currentNoteId.value = id
    const existing = store.notes.find((n) => n.id === id)
    if (existing) {
      content.value = existing.content
    } else {
      await store.fetchNotes()
      const note = store.notes.find((n) => n.id === id)
      if (note) {
        content.value = note.content
      }
    }
  }
  textareaRef.value?.focus()
})

function focusEditor() {
  textareaRef.value?.focus()
}

function onInput() {
  if (content.value.length > CHAR_LIMIT) {
    content.value = content.value.slice(0, CHAR_LIMIT)
  }
}

async function saveCurrentNapkin() {
  if (!content.value.trim()) return

  if (currentNoteId.value) {
    await store.updateNote(currentNoteId.value, content.value)
  } else {
    const note = await store.createNote(content.value)
    currentNoteId.value = note.id
  }
}

async function newNapkin() {
  await saveCurrentNapkin()
  content.value = ''
  currentNoteId.value = null
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
  height: 80vh;
  max-height: 700px;
  cursor: text;
}

.napkin-page__napkin {
  height: 100%;
  display: flex;
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
  padding: 2rem;
  box-sizing: border-box;
  line-height: 1.8;
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

.napkin-page__char-count {
  font-family: var(--handwriting);
  font-size: 1rem;
  color: rgba(255, 248, 231, 0.6);
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
</style>
