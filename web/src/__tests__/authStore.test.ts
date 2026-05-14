import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../stores/authStore'

vi.mock('../api/client', () => ({
  default: {
    post: vi.fn(),
  },
}))

import api from '../api/client'

describe('authStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts logged out', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('sets user and token on login', async () => {
    const mockResponse = {
      data: {
        access_token: 'test-access-token',
        refresh_token: 'test-refresh-token',
        user: { id: '1', email: 'test@example.com', display_name: 'Test' },
      },
    }
    vi.mocked(api.post).mockResolvedValue(mockResponse)

    const auth = useAuthStore()
    await auth.login('test@example.com', 'password')

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user).toEqual(mockResponse.data.user)
    expect(auth.accessToken).toBe('test-access-token')
    expect(localStorage.getItem('access_token')).toBe('test-access-token')
    expect(localStorage.getItem('refresh_token')).toBe('test-refresh-token')
  })

  it('sets user and token on register', async () => {
    const mockResponse = {
      data: {
        access_token: 'reg-access-token',
        refresh_token: 'reg-refresh-token',
        user: { id: '2', email: 'new@example.com', display_name: 'New User' },
      },
    }
    vi.mocked(api.post).mockResolvedValue(mockResponse)

    const auth = useAuthStore()
    await auth.register('new@example.com', 'password', 'New User')

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user).toEqual(mockResponse.data.user)
    expect(localStorage.getItem('access_token')).toBe('reg-access-token')
  })

  it('clears user and token on logout', async () => {
    const mockResponse = {
      data: {
        access_token: 'token',
        refresh_token: 'refresh',
        user: { id: '1', email: 'test@example.com', display_name: 'Test' },
      },
    }
    vi.mocked(api.post).mockResolvedValue(mockResponse)

    const auth = useAuthStore()
    await auth.login('test@example.com', 'password')
    expect(auth.isAuthenticated).toBe(true)

    auth.logout()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
    expect(auth.accessToken).toBeNull()
    expect(localStorage.getItem('access_token')).toBeNull()
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })
})
