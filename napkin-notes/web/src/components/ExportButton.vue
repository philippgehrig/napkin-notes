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
import { toPng } from 'html-to-image'

const props = defineProps<{
  targetSelector: string
}>()

const loading = ref(false)

async function exportNote() {
  const node = document.querySelector(props.targetSelector) as HTMLElement
  if (!node) return

  loading.value = true
  try {
    const dataUrl = await toPng(node, {
      pixelRatio: 2,
      backgroundColor: 'transparent',
    })

    const link = document.createElement('a')
    link.href = dataUrl
    link.download = 'napkin.png'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
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
  font-family: var(--handwriting);
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
