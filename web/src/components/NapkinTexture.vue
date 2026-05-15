<template>
  <div class="napkin-texture" :style="{ width: width, height: height, backgroundImage: `url(${texture})` }">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  width?: string
  height?: string
  variant?: number
}>(), {
  width: '100%',
  height: '100%',
  variant: 1,
})

const textures = ['/textures/napkin1.png', '/textures/napkin2.png', '/textures/napkin3.png']

const texture = computed(() => textures[(props.variant - 1) % textures.length])
</script>

<style scoped>
.napkin-texture {
  background-color: #f5f2ed;
  background-size: cover;
  background-position: center;
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.04);
  border-radius: 4px;
}
</style>
