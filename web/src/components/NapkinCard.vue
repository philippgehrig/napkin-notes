<template>
  <div
    class="napkin-card"
    :style="{ transform: `rotate(${rotation}deg)` }"
    @click="$emit('open', note.id)"
  >
    <NapkinTexture width="100%" height="100%">
      <div class="napkin-card__content">
        {{ preview }}
      </div>
    </NapkinTexture>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import NapkinTexture from './NapkinTexture.vue'
import type { Note } from '../stores/notesStore'

const props = defineProps<{
  note: Note
}>()

defineEmits<{
  open: [id: string]
}>()

const rotation = computed(() => {
  let hash = 0
  for (let i = 0; i < props.note.id.length; i++) {
    hash = ((hash << 5) - hash) + props.note.id.charCodeAt(i)
    hash |= 0
  }
  // Map to range -3 to +3
  return ((hash % 7) - 3) * (6 / 6)
})

const preview = computed(() => {
  const content = props.note.content
  if (content.length > 80) {
    return content.slice(0, 80) + '...'
  }
  return content
})
</script>

<style scoped>
.napkin-card {
  width: 220px;
  height: 200px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.napkin-card:hover {
  transform: rotate(0deg) scale(1.05) !important;
  z-index: 1;
}

.napkin-card__content {
  padding: 1.2rem;
  font-family: 'Caveat', cursive;
  font-size: 1.2rem;
  color: #2D2D2D;
  word-break: break-word;
  overflow: hidden;
  height: 100%;
  box-sizing: border-box;
}
</style>
