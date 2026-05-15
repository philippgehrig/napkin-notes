<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const router = useRouter()
const auth = useAuthStore()

const displayName = ref('')
const email = ref('')
const password = ref('')
const error = ref('')

async function handleSubmit() {
  error.value = ''
  try {
    await auth.register(email.value, password.value, displayName.value)
    router.push({ name: 'napkin' })
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Registration failed'
  }
}
</script>

<template>
  <div class="register-view">
    <h1>Register</h1>
    <form @submit.prevent="handleSubmit">
      <div v-if="error" class="error">{{ error }}</div>
      <div class="field">
        <label for="displayName">Display Name</label>
        <input id="displayName" v-model="displayName" type="text" required />
      </div>
      <div class="field">
        <label for="email">Email</label>
        <input id="email" v-model="email" type="email" required />
      </div>
      <div class="field">
        <label for="password">Password</label>
        <input id="password" v-model="password" type="password" required />
      </div>
      <button type="submit">Register</button>
    </form>
    <p>
      Already have an account?
      <router-link to="/login">Login</router-link>
    </p>
  </div>
</template>
