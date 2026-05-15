<template>
  <div class="trash">
    <header class="trash__header">
      <h1 class="trash__title">Ripped Napkins</h1>
    </header>

    <div v-if="store.loading" class="trash__loading">
      Loading trashed napkins...
    </div>

    <div v-else-if="store.trashedNotes.length === 0" class="trash__empty">
      <p>No ripped napkins.</p>
    </div>

    <div v-else class="trash__grid">
      <div
        v-for="note in store.trashedNotes"
        :key="note.id"
        class="trash__card"
      >
        <NapkinTexture width="100%" height="100%" :variant="note.texture_variant || 1">
          <div class="trash__card-content">
            {{ truncate(note.content) }}
          </div>
          <div class="trash__tape" />
        </NapkinTexture>
        <div class="trash__actions">
          <button class="trash__restore-btn" @click="handleRestore(note.id)">
            Restore
          </button>
          <button class="trash__delete-btn" @click="handlePermanentDelete(note.id)">
            Delete Forever
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useNotesStore } from '../stores/notesStore'
import NapkinTexture from '../components/NapkinTexture.vue'

const store = useNotesStore()

onMounted(() => {
  store.fetchTrashed()
})

function truncate(content: string): string {
  if (content.length > 80) {
    return content.slice(0, 80) + '...'
  }
  return content
}

function handleRestore(id: string) {
  store.restoreNote(id)
}

function handlePermanentDelete(id: string) {
  store.permanentlyDelete(id)
}
</script>

<style scoped>
.trash {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.trash__header {
  margin-bottom: 2rem;
}

.trash__title {
  font-family: var(--handwriting);
  font-size: 2.5rem;
  color: #FFF8E7;
  margin: 0;
}

.trash__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
  gap: 2rem;
}

.trash__card {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 2;
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.04);
  border-radius: 4px;
  overflow: hidden;
}

.trash__card-content {
  padding: 1.2rem;
  font-family: var(--handwriting);
  font-size: 1.2rem;
  color: #2D2D2D;
  word-break: break-word;
  overflow: hidden;
  height: 100%;
  box-sizing: border-box;
  opacity: 0.6;
}

.trash__tape {
  position: absolute;
  top: 50%;
  left: 10%;
  right: 10%;
  height: 20px;
  transform: translateY(-50%) rotate(-2deg);
  background-color: rgba(222, 198, 160, 0.6);
  border-top: 1px solid rgba(180, 160, 130, 0.4);
  border-bottom: 1px solid rgba(180, 160, 130, 0.4);
  pointer-events: none;
}

.trash__actions {
  position: absolute;
  bottom: 8px;
  right: 8px;
  display: flex;
  gap: 0.4rem;
}

.trash__restore-btn {
  background-color: #5C3D2E;
  color: #FFF8E7;
  border: none;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-family: var(--handwriting);
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.trash__restore-btn:hover {
  background-color: #3d2820;
}

.trash__delete-btn {
  background-color: #8B2020;
  color: #FFF8E7;
  border: none;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-family: var(--handwriting);
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.trash__delete-btn:hover {
  background-color: #5c1515;
}

.trash__loading,
.trash__empty {
  text-align: center;
  padding: 4rem 2rem;
  color: #FFF8E7;
  font-family: var(--handwriting);
  font-size: 1.4rem;
}
</style>
