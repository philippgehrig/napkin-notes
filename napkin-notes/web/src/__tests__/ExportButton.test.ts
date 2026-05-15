import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ExportButton from '../components/ExportButton.vue'

// Mock the api client
vi.mock('../api/client', () => ({
  default: {
    get: vi.fn(),
  },
}))

import api from '../api/client'

describe('ExportButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with Export text', () => {
    const wrapper = mount(ExportButton, {
      props: { noteId: 'note-1' },
    })
    expect(wrapper.text()).toBe('Export')
  })

  it('shows loading state while exporting', async () => {
    // Make the API call hang
    const mockGet = vi.mocked(api.get)
    mockGet.mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(ExportButton, {
      props: { noteId: 'note-1' },
    })

    await wrapper.find('button').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toBe('Exporting...')
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })

  it('calls API with correct URL and responseType', async () => {
    const mockBlob = new Blob(['fake-png'], { type: 'image/png' })
    const mockGet = vi.mocked(api.get)
    mockGet.mockResolvedValue({ data: mockBlob })

    // Mock URL.createObjectURL and revokeObjectURL
    const mockUrl = 'blob:http://localhost/fake-url'
    globalThis.URL.createObjectURL = vi.fn(() => mockUrl)
    globalThis.URL.revokeObjectURL = vi.fn()

    const wrapper = mount(ExportButton, {
      props: { noteId: 'note-123' },
    })

    await wrapper.find('button').trigger('click')
    // Wait for async operations
    await vi.waitFor(() => {
      expect(mockGet).toHaveBeenCalledWith('/notes/note-123/export?format=png', {
        responseType: 'blob',
      })
    })
  })

  it('resets loading state after export completes', async () => {
    const mockBlob = new Blob(['fake-png'], { type: 'image/png' })
    const mockGet = vi.mocked(api.get)
    mockGet.mockResolvedValue({ data: mockBlob })

    globalThis.URL.createObjectURL = vi.fn(() => 'blob:fake')
    globalThis.URL.revokeObjectURL = vi.fn()

    const wrapper = mount(ExportButton, {
      props: { noteId: 'note-1' },
    })

    await wrapper.find('button').trigger('click')
    await vi.waitFor(() => {
      expect(wrapper.text()).toBe('Export')
    })
  })

  it('resets loading state on error', async () => {
    const mockGet = vi.mocked(api.get)
    mockGet.mockRejectedValue(new Error('Network error'))

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const wrapper = mount(ExportButton, {
      props: { noteId: 'note-1' },
    })

    await wrapper.find('button').trigger('click')
    await vi.waitFor(() => {
      expect(wrapper.text()).toBe('Export')
    })

    consoleSpy.mockRestore()
  })
})
