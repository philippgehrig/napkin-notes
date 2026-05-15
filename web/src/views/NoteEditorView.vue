<template>
  <div class="editor">
    <header class="editor__header">
      <button class="editor__back-btn" @click="goBack">← Back</button>
      <div class="editor__actions">
        <ExportButton v-if="noteId" :noteId="noteId" />
        <button class="editor__save-btn" @click="save">Save</button>
      </div>
    </header>

    <div class="editor__wrapper">
      <NapkinTexture width="100%" height="100%">
        <textarea
          ref="textareaRef"
          v-model="content"
          class="editor__textarea"
          placeholder="Write on your napkin..."
        />
      </NapkinTexture>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'
import ExportButton from '../components/ExportButton.vue'

const route = useRoute()
const router = useRouter()
const store = useNotesStore()

const content = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const noteId = route.params.id as string | undefined

onMounted(async () => {
  if (noteId) {
    // Load existing note content from store or fetch
    const existing = store.notes.find((n) => n.id === noteId)
    if (existing) {
      content.value = existing.content
    } else {
      await store.fetchNotes()
      const note = store.notes.find((n) => n.id === noteId)
      if (note) {
        content.value = note.content
      }
    }
  }
  textareaRef.value?.focus()
})

function goBack() {
  router.push('/')
}

async function save() {
  if (!content.value.trim()) return

  if (noteId) {
    await store.updateNote(noteId, content.value)
  } else {
    await store.createNote(content.value)
  }
  router.push('/')
}
</script>

<style scoped>
.editor {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  height: calc(100vh - 4rem);
  display: flex;
  flex-direction: column;
}

.editor__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.editor__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.editor__back-btn,
.editor__save-btn {
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-family: var(--handwriting);
  font-size: 1.1rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.editor__back-btn {
  background-color: transparent;
  color: #5C3D2E;
  border: 1px solid #5C3D2E;
}

.editor__back-btn:hover {
  background-color: rgba(92, 61, 46, 0.08);
}

.editor__save-btn {
  background-color: #5C3D2E;
  color: #FFF8E7;
}

.editor__save-btn:hover {
  background-color: #3d2820;
}

.editor__wrapper {
  flex: 1;
  min-height: 0;
}

.editor__textarea {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  font-family: var(--handwriting);
  font-size: 1.4rem;
  color: #2D2D2D;
  padding: 1.5rem;
  box-sizing: border-box;
  line-height: 1.6;
}

.editor__textarea::placeholder {
  color: rgba(92, 61, 46, 0.4);
}
</style>
