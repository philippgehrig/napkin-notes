<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const router = useRouter()
const auth = useAuthStore()

function handleLogout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <nav class="app-nav">
    <router-link to="/">Gallery</router-link>
    <router-link to="/trash">Trash</router-link>
    <router-link to="/fonts">Fonts</router-link>
    <button @click="handleLogout">Logout</button>
  </nav>
</template>

<style scoped>
.app-nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1.5rem;
  background-color: #5C3D2E;
  color: #fff;
}

.app-nav a {
  color: #fff;
  text-decoration: none;
}

.app-nav a:hover {
  text-decoration: underline;
}

.app-nav button {
  margin-left: auto;
  background: none;
  border: 1px solid #fff;
  color: #fff;
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  cursor: pointer;
}

.app-nav button:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
</style>
