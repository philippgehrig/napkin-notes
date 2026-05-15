# Phase 7: Vue SPA Scaffold

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Set up Vue Router, Pinia stores, auth views, and API client.

**Branch:** `feat/phase-07-vue-scaffold`

---

## File Structure

```
web/src/
├── api/
│   └── client.ts          (axios instance with interceptors)
├── router/
│   └── index.ts           (Vue Router with auth guards)
├── stores/
│   ├── authStore.ts       (Pinia auth state)
│   └── notesStore.ts      (Pinia notes state)
├── views/
│   ├── LoginView.vue
│   └── RegisterView.vue
├── composables/
│   └── useAuth.ts
├── components/
│   └── AppNav.vue
├── App.vue (updated)
└── main.ts (updated)
```

---

### Task 1: Add dependencies

- [ ] **Step 1: Install packages**

```bash
cd web && yarn add vue-router@4 pinia axios && yarn add -D @types/node
```

- [ ] **Step 2: Commit**

```bash
git add web/package.json web/yarn.lock
git commit -m "feat: add vue-router, pinia, axios dependencies"
```

---

### Task 2: API client

**Files:**
- Create: `web/src/api/client.ts`

- [ ] **Step 1: Create API client**

Create `web/src/api/client.ts`:
```ts
import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const { data } = await axios.post('/api/auth/refresh', {
            refresh_token: refreshToken,
          })
          localStorage.setItem('access_token', data.access_token)
          localStorage.setItem('refresh_token', data.refresh_token)
          originalRequest.headers.Authorization = `Bearer ${data.access_token}`
          return api(originalRequest)
        } catch {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.href = '/login'
        }
      }
    }
    return Promise.reject(error)
  }
)

export default api
```

- [ ] **Step 2: Commit**

```bash
git add web/src/api/
git commit -m "feat: add axios API client with token refresh"
```

---

### Task 3: Auth store

**Files:**
- Create: `web/src/stores/authStore.ts`
- Create: `web/src/__tests__/authStore.test.ts`

- [ ] **Step 1: Write auth store test**

Create `web/src/__tests__/authStore.test.ts`:
```ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../stores/authStore'

vi.mock('../api/client', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}))

describe('authStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('starts logged out', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
  })

  it('sets user on successful login', async () => {
    const { default: api } = await import('../api/client')
    const mockApi = vi.mocked(api)
    mockApi.post.mockResolvedValueOnce({
      data: {
        user: { id: '1', email: 'test@test.com', display_name: 'Test' },
        access_token: 'access-123',
        refresh_token: 'refresh-123',
      },
    })

    const store = useAuthStore()
    await store.login('test@test.com', 'password')

    expect(store.isAuthenticated).toBe(true)
    expect(store.user?.email).toBe('test@test.com')
  })

  it('clears state on logout', async () => {
    const store = useAuthStore()
    store.user = { id: '1', email: 'test@test.com', display_name: 'Test' }
    store.accessToken = 'token'

    store.logout()

    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
  })
})
```

- [ ] **Step 2: Implement auth store**

Create `web/src/stores/authStore.ts`:
```ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api/client'

interface User {
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
    setTokens(data.access_token, data.refresh_token)
    user.value = data.user
  }

  async function register(email: string, password: string, displayName: string) {
    const { data } = await api.post('/auth/register', {
      email,
      password,
      display_name: displayName,
    })
    setTokens(data.access_token, data.refresh_token)
    user.value = data.user
  }

  function logout() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    accessToken.value = null
    user.value = null
  }

  function setTokens(access: string, refresh: string) {
    localStorage.setItem('access_token', access)
    localStorage.setItem('refresh_token', refresh)
    accessToken.value = access
  }

  return { user, accessToken, isAuthenticated, login, register, logout }
})
```

- [ ] **Step 3: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/stores/ web/src/__tests__/authStore.test.ts
git commit -m "feat: add Pinia auth store with login/register/logout"
```

---

### Task 4: Vue Router with auth guard

**Files:**
- Create: `web/src/router/index.ts`

- [ ] **Step 1: Create router**

Create `web/src/router/index.ts`:
```ts
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guest: true },
  },
  {
    path: '/',
    name: 'gallery',
    component: () => import('../views/GalleryView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/note/:id?',
    name: 'editor',
    component: () => import('../views/NoteEditorView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/trash',
    name: 'trash',
    component: () => import('../views/TrashView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/fonts',
    name: 'fonts',
    component: () => import('../views/FontsView.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('access_token')
  if (to.meta.requiresAuth && !token) {
    next({ name: 'login' })
  } else if (to.meta.guest && token) {
    next({ name: 'gallery' })
  } else {
    next()
  }
})

export default router
```

- [ ] **Step 2: Commit**

```bash
git add web/src/router/
git commit -m "feat: add Vue Router with auth guards"
```

---

### Task 5: Auth views (Login + Register)

**Files:**
- Create: `web/src/views/LoginView.vue`
- Create: `web/src/views/RegisterView.vue`
- Create: `web/src/views/GalleryView.vue` (placeholder)
- Create: `web/src/views/NoteEditorView.vue` (placeholder)
- Create: `web/src/views/TrashView.vue` (placeholder)
- Create: `web/src/views/FontsView.vue` (placeholder)

- [ ] **Step 1: Create LoginView**

Create `web/src/views/LoginView.vue`:
```vue
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push({ name: 'gallery' })
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <h1>Napkin Notes</h1>
    <form @submit.prevent="handleLogin" class="auth-form">
      <div v-if="error" class="error">{{ error }}</div>
      <input
        v-model="email"
        type="email"
        placeholder="Email"
        required
        autocomplete="email"
      />
      <input
        v-model="password"
        type="password"
        placeholder="Password"
        required
        autocomplete="current-password"
      />
      <button type="submit" :disabled="loading">
        {{ loading ? 'Signing in...' : 'Sign In' }}
      </button>
      <p>
        Don't have an account?
        <router-link to="/register">Sign up</router-link>
      </p>
    </form>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
}
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 320px;
}
.auth-form input {
  padding: 0.75rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}
.auth-form button {
  padding: 0.75rem;
  background: #5C3D2E;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}
.auth-form button:disabled {
  opacity: 0.6;
}
.error {
  color: #c0392b;
  font-size: 0.875rem;
}
</style>
```

- [ ] **Step 2: Create RegisterView**

Create `web/src/views/RegisterView.vue`:
```vue
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const displayName = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  error.value = ''
  loading.value = true
  try {
    await auth.register(email.value, password.value, displayName.value)
    router.push({ name: 'gallery' })
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <h1>Create Account</h1>
    <form @submit.prevent="handleRegister" class="auth-form">
      <div v-if="error" class="error">{{ error }}</div>
      <input
        v-model="displayName"
        type="text"
        placeholder="Display name"
        required
      />
      <input
        v-model="email"
        type="email"
        placeholder="Email"
        required
        autocomplete="email"
      />
      <input
        v-model="password"
        type="password"
        placeholder="Password (min 8 chars)"
        required
        minlength="8"
        autocomplete="new-password"
      />
      <button type="submit" :disabled="loading">
        {{ loading ? 'Creating...' : 'Create Account' }}
      </button>
      <p>
        Already have an account?
        <router-link to="/login">Sign in</router-link>
      </p>
    </form>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
}
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 320px;
}
.auth-form input {
  padding: 0.75rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}
.auth-form button {
  padding: 0.75rem;
  background: #5C3D2E;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}
.auth-form button:disabled {
  opacity: 0.6;
}
.error {
  color: #c0392b;
  font-size: 0.875rem;
}
</style>
```

- [ ] **Step 3: Create placeholder views**

Create `web/src/views/GalleryView.vue`:
```vue
<script setup lang="ts">
</script>

<template>
  <div class="gallery">
    <h1>Gallery</h1>
    <p>Your napkins will appear here.</p>
  </div>
</template>
```

Create `web/src/views/NoteEditorView.vue`:
```vue
<script setup lang="ts">
</script>

<template>
  <div class="editor">
    <h1>Editor</h1>
    <p>Note editor coming soon.</p>
  </div>
</template>
```

Create `web/src/views/TrashView.vue`:
```vue
<script setup lang="ts">
</script>

<template>
  <div class="trash">
    <h1>Trash</h1>
    <p>Ripped napkins will appear here.</p>
  </div>
</template>
```

Create `web/src/views/FontsView.vue`:
```vue
<script setup lang="ts">
</script>

<template>
  <div class="fonts">
    <h1>Fonts</h1>
    <p>Font management coming soon.</p>
  </div>
</template>
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/
git commit -m "feat: add auth views and placeholder views"
```

---

### Task 6: Update App.vue and main.ts

**Files:**
- Modify: `web/src/App.vue`
- Modify: `web/src/main.ts`

- [ ] **Step 1: Update main.ts**

Replace `web/src/main.ts`:
```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

- [ ] **Step 2: Update App.vue**

Replace `web/src/App.vue`:
```vue
<script setup lang="ts">
import AppNav from './components/AppNav.vue'
import { useAuthStore } from './stores/authStore'

const auth = useAuthStore()
</script>

<template>
  <AppNav v-if="auth.isAuthenticated" />
  <router-view />
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background: #FFF8E7;
  color: #2D2D2D;
}
</style>
```

- [ ] **Step 3: Create AppNav component**

Create `web/src/components/AppNav.vue`:
```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <nav class="app-nav">
    <router-link to="/" class="nav-brand">Napkin Notes</router-link>
    <div class="nav-links">
      <router-link to="/">Gallery</router-link>
      <router-link to="/trash">Trash</router-link>
      <router-link to="/fonts">Fonts</router-link>
      <button @click="handleLogout">Logout</button>
    </div>
  </nav>
</template>

<style scoped>
.app-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: #5C3D2E;
  color: white;
}
.nav-brand {
  font-size: 1.25rem;
  font-weight: bold;
  color: white;
  text-decoration: none;
}
.nav-links {
  display: flex;
  gap: 1rem;
  align-items: center;
}
.nav-links a {
  color: white;
  text-decoration: none;
  opacity: 0.8;
}
.nav-links a.router-link-active {
  opacity: 1;
  text-decoration: underline;
}
.nav-links button {
  background: transparent;
  border: 1px solid white;
  color: white;
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  cursor: pointer;
}
</style>
```

- [ ] **Step 4: Run tests**

```bash
cd web && yarn test
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/
git commit -m "feat: wire up App with router, pinia, and nav"
```

---

### Task 7: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-07-vue-scaffold
gh pr create --title "feat: Vue SPA scaffold with auth, router, and stores" --body "## Summary
- Add Vue Router with auth guards
- Add Pinia auth store with login/register/logout
- Add axios API client with token refresh interceptor
- Add Login and Register views
- Add placeholder views for Gallery, Editor, Trash, Fonts
- Add AppNav component

## Test plan
- [ ] \`cd web && yarn test\` passes
- [ ] Login view renders and submits form
- [ ] Auth guard redirects unauthenticated users to login
- [ ] Nav appears only when authenticated

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
