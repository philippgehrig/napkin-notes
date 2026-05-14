import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useRipGesture } from '../composables/useRipGesture'

describe('useRipGesture', () => {
  function createMockElement(width = 200) {
    const el = document.createElement('div')
    Object.defineProperty(el, 'offsetWidth', { value: width, configurable: true })
    return el
  }

  it('starts inactive with progress 0', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { isRipping, progress } = useRipGesture(elRef, { onRip })

    expect(isRipping.value).toBe(false)
    expect(progress.value).toBe(0)
  })

  it('calculates progress from simulateDrag', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { progress, isRipping, simulateDrag } = useRipGesture(elRef, { onRip })

    simulateDrag(0.5)
    expect(progress.value).toBe(0.5)
    expect(isRipping.value).toBe(true)
  })

  it('triggers onRip when released past threshold', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { simulateDrag, simulateRelease } = useRipGesture(elRef, { onRip })

    simulateDrag(0.5)
    simulateRelease()
    expect(onRip).toHaveBeenCalledOnce()
  })

  it('does not trigger onRip when released below threshold', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { simulateDrag, simulateRelease, progress, isRipping } = useRipGesture(elRef, { onRip })

    simulateDrag(0.3)
    simulateRelease()
    expect(onRip).not.toHaveBeenCalled()
    expect(progress.value).toBe(0)
    expect(isRipping.value).toBe(false)
  })

  it('respects custom threshold', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { simulateDrag, simulateRelease } = useRipGesture(elRef, { onRip, threshold: 0.8 })

    simulateDrag(0.5)
    simulateRelease()
    expect(onRip).not.toHaveBeenCalled()

    simulateDrag(0.9)
    simulateRelease()
    expect(onRip).toHaveBeenCalledOnce()
  })

  it('resets state after release below threshold', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { simulateDrag, simulateRelease, progress, isRipping } = useRipGesture(elRef, { onRip })

    simulateDrag(0.2)
    expect(isRipping.value).toBe(true)

    simulateRelease()
    expect(isRipping.value).toBe(false)
    expect(progress.value).toBe(0)
  })

  it('clamps progress between 0 and 1', () => {
    const elRef = ref(createMockElement())
    const onRip = vi.fn()
    const { simulateDrag, progress } = useRipGesture(elRef, { onRip })

    simulateDrag(1.5)
    expect(progress.value).toBe(1)

    simulateDrag(-0.3)
    expect(progress.value).toBe(0)
  })
})
