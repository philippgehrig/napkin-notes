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
    <router-link to="/" class="app-nav__brand">Napkin Notes</router-link>
    <div class="app-nav__links">
      <router-link to="/">Napkin</router-link>
      <router-link to="/gallery">Gallery</router-link>
      <router-link to="/trash">Trash</router-link>
      <button @click="handleLogout">Logout</button>
    </div>
  </nav>
</template>

<style scoped>
.app-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background: transparent;
  position: relative;
  z-index: 10;
}

.app-nav__brand {
  font-family: var(--handwriting);
  font-size: 1.8rem;
  color: #FFF8E7;
  text-decoration: none;
}

.app-nav__brand:hover {
  opacity: 0.8;
}

.app-nav__links {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.app-nav__links a {
  font-family: var(--handwriting);
  font-size: 1.2rem;
  color: #FFF8E7;
  text-decoration: none;
  transition: opacity 0.2s ease;
}

.app-nav__links a:hover {
  opacity: 0.7;
}

.app-nav__links a.router-link-active {
  border-bottom: 2px solid #FFF8E7;
  padding-bottom: 2px;
}

.app-nav__links button {
  font-family: var(--handwriting);
  font-size: 1.1rem;
  background: none;
  border: none;
  color: #FFF8E7;
  cursor: pointer;
  opacity: 0.8;
  transition: opacity 0.2s ease;
}

.app-nav__links button:hover {
  opacity: 0.6;
}
</style>
