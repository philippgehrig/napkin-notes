import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api from '../api/client'

export interface User {
  id: string
  email: string
  display_name: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem('access_token'))

  const isAuthenticated = computed(() => !!accessToken.value)

  async function login(email: string, password: string) {
    const { data } = await api.post('/auth/login', { email, password })
    accessToken.value = data.access_token
    user.value = data.user
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
  }

  async function register(email: string, password: string, displayName: string) {
    const { data } = await api.post('/auth/register', {
      email,
      password,
      display_name: displayName,
    })
    accessToken.value = data.access_token
    user.value = data.user
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
  }

  function logout() {
    user.value = null
    accessToken.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return { user, accessToken, isAuthenticated, login, register, logout }
})
