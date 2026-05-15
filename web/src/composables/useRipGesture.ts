import { ref, watch, type Ref } from 'vue'

export interface RipGestureOptions {
  onRip: () => void
  threshold?: number
}

export function useRipGesture(
  elementRef: Ref<HTMLElement | null | undefined>,
  options: RipGestureOptions,
) {
  const threshold = options.threshold ?? 0.4
  const isRipping = ref(false)
  const progress = ref(0)
  const didRip = ref(false)

  let startX = 0
  let tracking = false

  function clamp(value: number, min: number, max: number) {
    return Math.min(Math.max(value, min), max)
  }

  function onPointerDown(e: PointerEvent) {
    startX = e.clientX
    tracking = true
    isRipping.value = true
    didRip.value = false
    ;(e.currentTarget as HTMLElement)?.setPointerCapture(e.pointerId)
  }

  function onPointerMove(e: PointerEvent) {
    if (!tracking) return
    const el = elementRef.value
    if (!el) return
    const dx = Math.abs(e.clientX - startX)
    const width = el.offsetWidth
    progress.value = clamp(dx / width, 0, 1)
    if (progress.value > 0.05) {
      didRip.value = true
    }
  }

  function onPointerUp() {
    if (!tracking) return
    tracking = false
    if (progress.value >= threshold) {
      options.onRip()
    }
    // Reset state
    progress.value = 0
    isRipping.value = false

    // If a drag occurred, suppress the upcoming click event
    if (didRip.value) {
      const el = elementRef.value
      if (el) {
        const suppressClick = (e: Event) => {
          e.stopPropagation()
          e.preventDefault()
          el.removeEventListener('click', suppressClick, true)
        }
        el.addEventListener('click', suppressClick, true)
      }
    }
  }

  // Attach/detach event listeners when the element ref changes
  watch(
    elementRef,
    (el, oldEl) => {
      if (oldEl) {
        oldEl.removeEventListener('pointerdown', onPointerDown as EventListener)
        oldEl.removeEventListener('pointermove', onPointerMove as EventListener)
        oldEl.removeEventListener('pointerup', onPointerUp as EventListener)
        oldEl.removeEventListener('pointercancel', onPointerUp as EventListener)
      }
      if (el) {
        el.addEventListener('pointerdown', onPointerDown as EventListener)
        el.addEventListener('pointermove', onPointerMove as EventListener)
        el.addEventListener('pointerup', onPointerUp as EventListener)
        el.addEventListener('pointercancel', onPointerUp as EventListener)
      }
    },
    { immediate: true },
  )

  // Test helpers
  function simulateDrag(pct: number) {
    isRipping.value = true
    tracking = true
    progress.value = clamp(pct, 0, 1)
  }

  function simulateRelease() {
    if (progress.value >= threshold) {
      options.onRip()
    }
    progress.value = 0
    isRipping.value = false
    tracking = false
  }

  return {
    isRipping,
    progress,
    didRip,
    simulateDrag,
    simulateRelease,
  }
}
