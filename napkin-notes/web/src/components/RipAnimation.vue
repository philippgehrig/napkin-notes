<template>
  <div class="rip-animation" :class="{ 'rip-animation--active': isActive }">
    <template v-if="isActive && progress > 0">
      <div
        class="rip-animation__half rip-animation__left"
        :style="leftStyle"
      >
        <slot />
      </div>
      <div
        class="rip-animation__half rip-animation__right"
        :style="rightStyle"
      >
        <slot />
      </div>
    </template>
    <template v-else>
      <slot />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  progress: number
  isActive: boolean
}>(), {
  progress: 0,
  isActive: false,
})

/**
 * Generate a procedural jagged tear path down the middle.
 * Uses sin/cos to create organic jitter around the center line.
 */
function generateTearPath(side: 'left' | 'right'): string {
  const points: string[] = []
  const steps = 20
  const seed = 42

  for (let i = 0; i <= steps; i++) {
    const y = (i / steps) * 100
    // Create jagged jitter using sin/cos for organic feel
    const jitter = Math.sin(i * 1.7 + seed) * 3 + Math.cos(i * 2.3 + seed) * 2
    const x = 50 + jitter
    points.push(`${x}% ${y}%`)
  }

  if (side === 'left') {
    // Left half: from top-left, along tear path, to bottom-left
    return `polygon(0% 0%, ${points.join(', ')}, 0% 100%)`
  } else {
    // Right half: from tear path top, to top-right, bottom-right, back along tear bottom
    return `polygon(${points.join(', ')}, 100% 100%, 100% 0%)`
  }
}

const leftClipPath = generateTearPath('left')
const rightClipPath = generateTearPath('right')

const opacity = computed(() => {
  if (props.progress > 0.8) {
    // Fade from 1 to 0 over the last 20% of progress
    return 1 - (props.progress - 0.8) / 0.2
  }
  return 1
})

const leftStyle = computed(() => ({
  clipPath: leftClipPath,
  transform: `translateX(${-props.progress * 30}%) rotate(${-props.progress * 5}deg)`,
  opacity: opacity.value,
}))

const rightStyle = computed(() => ({
  clipPath: rightClipPath,
  transform: `translateX(${props.progress * 30}%) rotate(${props.progress * 5}deg)`,
  opacity: opacity.value,
}))
</script>

<style scoped>
.rip-animation {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
}

.rip-animation__half {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  transition: none;
  will-change: transform, opacity, clip-path;
}
</style>
