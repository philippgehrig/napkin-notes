import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NapkinCard from '../components/NapkinCard.vue'
import type { Note } from '../stores/notesStore'

const baseNote: Note = {
  id: 'note-abc',
  user_id: 'user-1',
  content: 'Hello napkin world',
  texture_variant: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('NapkinCard', () => {
  it('renders note content', () => {
    const wrapper = mount(NapkinCard, {
      props: { note: baseNote },
    })
    expect(wrapper.text()).toContain('Hello napkin world')
  })

  it('truncates content longer than 80 characters', () => {
    const longContent = 'A'.repeat(100)
    const wrapper = mount(NapkinCard, {
      props: { note: { ...baseNote, content: longContent } },
    })
    expect(wrapper.text()).toContain('A'.repeat(80) + '...')
    expect(wrapper.text()).not.toContain('A'.repeat(81))
  })

  it('does not truncate content of exactly 80 characters', () => {
    const exactContent = 'B'.repeat(80)
    const wrapper = mount(NapkinCard, {
      props: { note: { ...baseNote, content: exactContent } },
    })
    expect(wrapper.text()).toContain(exactContent)
    expect(wrapper.text()).not.toContain('...')
  })

  it('applies rotation style based on note id', () => {
    const wrapper = mount(NapkinCard, {
      props: { note: baseNote },
    })
    const style = wrapper.attributes('style')
    expect(style).toContain('rotate(')
    // Rotation should be within -3 to +3 degrees
    const match = style?.match(/rotate\((-?\d+(?:\.\d+)?)deg\)/)
    expect(match).not.toBeNull()
    const deg = parseFloat(match![1])
    expect(deg).toBeGreaterThanOrEqual(-3)
    expect(deg).toBeLessThanOrEqual(3)
  })

  it('emits open event with note id on click', async () => {
    const wrapper = mount(NapkinCard, {
      props: { note: baseNote },
    })
    await wrapper.trigger('click')
    expect(wrapper.emitted('open')).toBeTruthy()
    expect(wrapper.emitted('open')![0]).toEqual(['note-abc'])
  })

  it('uses NapkinTexture as wrapper', () => {
    const wrapper = mount(NapkinCard, {
      props: { note: baseNote },
    })
    expect(wrapper.find('.napkin-texture').exists()).toBe(true)
  })
})
