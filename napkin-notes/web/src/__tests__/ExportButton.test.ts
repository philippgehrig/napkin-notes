import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ExportButton from '../components/ExportButton.vue'

vi.mock('html-to-image', () => ({
  toPng: vi.fn(),
}))

import { toPng } from 'html-to-image'

describe('ExportButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with Export text', () => {
    const wrapper = mount(ExportButton, {
      props: { targetSelector: '.napkin' },
    })
    expect(wrapper.text()).toBe('Export')
  })

  it('shows loading state while exporting', async () => {
    vi.mocked(toPng).mockImplementation(() => new Promise(() => {}))

    const el = document.createElement('div')
    el.className = 'napkin'
    document.body.appendChild(el)

    const wrapper = mount(ExportButton, {
      props: { targetSelector: '.napkin' },
    })

    await wrapper.find('button').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toBe('Exporting...')
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()

    document.body.removeChild(el)
  })

  it('resets loading state after export completes', async () => {
    vi.mocked(toPng).mockResolvedValue('data:image/png;base64,fake')

    const el = document.createElement('div')
    el.className = 'napkin'
    document.body.appendChild(el)

    const wrapper = mount(ExportButton, {
      props: { targetSelector: '.napkin' },
    })

    await wrapper.find('button').trigger('click')
    await vi.waitFor(() => {
      expect(wrapper.text()).toBe('Export')
    })

    document.body.removeChild(el)
  })

  it('resets loading state on error', async () => {
    vi.mocked(toPng).mockRejectedValue(new Error('render failed'))
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const el = document.createElement('div')
    el.className = 'napkin'
    document.body.appendChild(el)

    const wrapper = mount(ExportButton, {
      props: { targetSelector: '.napkin' },
    })

    await wrapper.find('button').trigger('click')
    await vi.waitFor(() => {
      expect(wrapper.text()).toBe('Export')
    })

    document.body.removeChild(el)
    consoleSpy.mockRestore()
  })

  it('does nothing if target element not found', async () => {
    const wrapper = mount(ExportButton, {
      props: { targetSelector: '.nonexistent' },
    })

    await wrapper.find('button').trigger('click')
    await wrapper.vm.$nextTick()

    expect(toPng).not.toHaveBeenCalled()
    expect(wrapper.text()).toBe('Export')
  })
})
