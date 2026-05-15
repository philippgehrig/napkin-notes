<script setup lang="ts">
import { useAuthStore } from './stores/authStore'
import AppNav from './components/AppNav.vue'

const auth = useAuthStore()
</script>

<template>
  <AppNav v-if="auth.isAuthenticated" />
  <router-view />
</template>
