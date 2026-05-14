<template>
  <div class="napkin-texture" :style="{ width: width, height: height }">
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  width?: string
  height?: string
}>(), {
  width: '100%',
  height: '100%',
})
</script>

<style scoped>
.napkin-texture {
  background-color: #FFF8E7;
  background-image:
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 27px,
      rgba(92, 61, 46, 0.05) 27px,
      rgba(92, 61, 46, 0.05) 28px
    );
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.04);
  border-radius: 4px;
}
</style>
