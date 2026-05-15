<template>
  <div class="napkin-texture" :style="{ width: width, height: height }">
    <img :src="texture" class="napkin-texture__img" />
    <div class="napkin-texture__content">
      <slot />
    </div>
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
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.napkin-texture__img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  pointer-events: none;
}

.napkin-texture__content {
  position: absolute;
  top: 12%;
  left: 25%;
  width: 50%;
  height: 76%;
  pointer-events: none;
}
</style>
