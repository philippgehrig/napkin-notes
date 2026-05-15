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
  <div class="auth">
    <div class="auth__card">
      <h1 class="auth__title">Grab a Napkin</h1>
      <p class="auth__subtitle">Create your account</p>

      <form class="auth__form" @submit.prevent="handleSubmit">
        <div v-if="error" class="auth__error">{{ error }}</div>

        <div class="auth__field">
          <input
            id="displayName"
            v-model="displayName"
            type="text"
            required
            placeholder="Display Name"
            autocomplete="name"
          />
        </div>

        <div class="auth__field">
          <input
            id="email"
            v-model="email"
            type="email"
            required
            placeholder="Email"
            autocomplete="email"
          />
        </div>

        <div class="auth__field">
          <input
            id="password"
            v-model="password"
            type="password"
            required
            placeholder="Password"
            autocomplete="new-password"
          />
        </div>

        <button type="submit" class="auth__submit">Create Account</button>
      </form>

      <p class="auth__switch">
        Already have an account?
        <router-link to="/login">Sign in</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1rem;
  box-sizing: border-box;
}

.auth__card {
  width: 100%;
  max-width: 380px;
  padding: 2.5rem 2rem;
  background: rgba(44, 30, 22, 0.85);
  backdrop-filter: blur(12px);
  border-radius: 16px;
  border: 1px solid rgba(255, 248, 231, 0.1);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
}

.auth__title {
  font-family: var(--handwriting);
  font-size: 2.2rem;
  color: #FFF8E7;
  margin: 0 0 0.3rem;
  text-align: center;
}

.auth__subtitle {
  font-family: var(--handwriting);
  font-size: 1.1rem;
  color: rgba(255, 248, 231, 0.6);
  margin: 0 0 2rem;
  text-align: center;
}

.auth__form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.auth__error {
  background: rgba(139, 32, 32, 0.3);
  border: 1px solid rgba(139, 32, 32, 0.6);
  color: #ff9b9b;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  text-align: center;
}

.auth__field input {
  width: 100%;
  padding: 0.9rem 1rem;
  border: 1px solid rgba(255, 248, 231, 0.15);
  border-radius: 10px;
  background: rgba(255, 248, 231, 0.05);
  color: #FFF8E7;
  font-family: var(--handwriting);
  font-size: 1.1rem;
  outline: none;
  transition: border-color 0.2s ease, background 0.2s ease;
  box-sizing: border-box;
}

.auth__field input::placeholder {
  color: rgba(255, 248, 231, 0.4);
}

.auth__field input:focus {
  border-color: rgba(255, 248, 231, 0.4);
  background: rgba(255, 248, 231, 0.08);
}

.auth__submit {
  width: 100%;
  padding: 0.9rem;
  margin-top: 0.5rem;
  border: none;
  border-radius: 10px;
  background: #5C3D2E;
  color: #FFF8E7;
  font-family: var(--handwriting);
  font-size: 1.2rem;
  cursor: pointer;
  transition: background-color 0.2s ease, transform 0.1s ease;
}

.auth__submit:hover {
  background: #3d2820;
}

.auth__submit:active {
  transform: scale(0.98);
}

.auth__switch {
  margin: 1.5rem 0 0;
  text-align: center;
  font-family: var(--handwriting);
  font-size: 1rem;
  color: rgba(255, 248, 231, 0.6);
}

.auth__switch a {
  color: #FFF8E7;
  text-decoration: none;
  border-bottom: 1px solid rgba(255, 248, 231, 0.3);
  transition: border-color 0.2s ease;
}

.auth__switch a:hover {
  border-color: #FFF8E7;
}

@media (max-width: 600px) {
  .auth__card {
    padding: 2rem 1.5rem;
  }

  .auth__title {
    font-size: 1.8rem;
  }
}
</style>
