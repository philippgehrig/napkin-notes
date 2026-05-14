<template>
  <button
    class="export-btn"
    :disabled="loading"
    @click="exportNote"
  >
    {{ loading ? 'Exporting...' : 'Export' }}
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import api from '../api/client'

const props = defineProps<{
  noteId: string
}>()

const loading = ref(false)

async function exportNote() {
  loading.value = true
  try {
    const response = await api.get(`/notes/${props.noteId}/export?format=png`, {
      responseType: 'blob',
    })

    const url = URL.createObjectURL(response.data)
    const link = document.createElement('a')
    link.href = url
    link.download = 'napkin.png'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Export failed:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.export-btn {
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-family: 'Caveat', cursive;
  font-size: 1.1rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
  background-color: #5C3D2E;
  color: #FFF8E7;
}

.export-btn:hover:not(:disabled) {
  background-color: #3d2820;
}

.export-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
