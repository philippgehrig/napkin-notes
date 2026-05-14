import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import App from '../App.vue'

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<h1>Your Napkins</h1>' } },
      { path: '/login', component: { template: '<h1>Login</h1>' } },
    ],
  })
}

describe('App', () => {

  it('mounts without errors', async () => {
    const router = createTestRouter()
    router.push('/login')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [createPinia(), router],
      },
    })
    expect(wrapper.exists()).toBe(true)
  })

  it('does not show nav when not authenticated', async () => {
    const router = createTestRouter()
    router.push('/login')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [createPinia(), router],
      },
    })
    expect(wrapper.find('.app-nav').exists()).toBe(false)
  })

  it('shows nav when authenticated', async () => {
    localStorage.setItem('access_token', 'fake-token')

    const router = createTestRouter()
    router.push('/')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [createPinia(), router],
      },
    })
    expect(wrapper.find('.app-nav').exists()).toBe(true)
  })
})
