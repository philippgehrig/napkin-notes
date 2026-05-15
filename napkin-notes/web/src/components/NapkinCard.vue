<template>
  <div
    ref="cardRef"
    class="napkin-card"
    :style="{ transform: `rotate(${rotation}deg)` }"
    @click="handleClick"
  >
    <RipAnimation :progress="progress" :is-active="isRipping">
      <NapkinTexture width="100%" height="100%" :variant="napkinVariant">
        <div class="napkin-card__content">
          {{ preview }}
        </div>
      </NapkinTexture>
    </RipAnimation>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
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
})

function handleClick() {
  emit('open', props.note.id)
}

const napkinVariant = computed(() => {
  return props.note.texture_variant || 1
})

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
  width: 100%;
  aspect-ratio: 3 / 2;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  touch-action: pan-y;
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.04);
  border-radius: 4px;
  overflow: hidden;
}

.napkin-card:hover {
  transform: rotate(0deg) scale(1.05) !important;
  z-index: 1;
}

.napkin-card__content {
  padding: 1.2rem;
  font-family: var(--handwriting);
  font-size: 1.2rem;
  color: #2D2D2D;
  word-break: break-word;
  overflow: hidden;
  height: 100%;
  box-sizing: border-box;
}
</style>
