<template>
  <div class="gallery">
    <header class="gallery__header">
      <h1 class="gallery__title">Your Napkins</h1>
      <button class="gallery__new-btn" @click="createNew">+ New Napkin</button>
    </header>

    <div v-if="store.loading" class="gallery__loading">
      Loading your napkins...
    </div>

    <div v-else-if="store.notes.length === 0" class="gallery__empty">
      <p>No napkins yet. Create your first one!</p>
    </div>

    <div v-else class="gallery__grid">
      <NapkinCard
        v-for="note in store.notes"
        :key="note.id"
        :note="note"
        @open="openNote"
        @rip="handleRip"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotesStore } from '../stores/notesStore'
import NapkinCard from '../components/NapkinCard.vue'

const router = useRouter()
const store = useNotesStore()

onMounted(() => {
  store.fetchNotes()
})

function openNote(id: string) {
  router.push(`/note/${id}`)
}

function handleRip(id: string) {
  store.deleteNote(id)
}

function createNew() {
  router.push('/note')
}
</script>

<style scoped>
.gallery {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.gallery__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
}

.gallery__title {
  font-family: 'Caveat', cursive;
  font-size: 2.5rem;
  color: #2D2D2D;
  margin: 0;
}

.gallery__new-btn {
  background-color: #5C3D2E;
  color: #FFF8E7;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-family: 'Caveat', cursive;
  font-size: 1.2rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.gallery__new-btn:hover {
  background-color: #3d2820;
}

.gallery__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 2rem;
}

.gallery__loading,
.gallery__empty {
  text-align: center;
  padding: 4rem 2rem;
  color: #5C3D2E;
  font-family: 'Caveat', cursive;
  font-size: 1.4rem;
}
</style>
