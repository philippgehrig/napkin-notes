import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNotesStore } from '../stores/notesStore'

vi.mock('../api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import api from '../api/client'

const mockNote = {
  id: 'note-1',
  user_id: 'user-1',
  content: 'Hello world',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('notesStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts with empty notes', () => {
    const store = useNotesStore()
    expect(store.notes).toEqual([])
    expect(store.trashedNotes).toEqual([])
    expect(store.loading).toBe(false)
  })

  it('fetches notes', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: [mockNote] })

    const store = useNotesStore()
    await store.fetchNotes()

    expect(api.get).toHaveBeenCalledWith('/notes')
    expect(store.notes).toEqual([mockNote])
    expect(store.loading).toBe(false)
  })

  it('creates a note', async () => {
    vi.mocked(api.post).mockResolvedValue({ data: mockNote })

    const store = useNotesStore()
    const result = await store.createNote('Hello world')

    expect(api.post).toHaveBeenCalledWith('/notes', { content: 'Hello world', texture_variant: 1 })
    expect(store.notes).toContainEqual(mockNote)
    expect(result).toEqual(mockNote)
  })

  it('creates a note with texture variant', async () => {
    vi.mocked(api.post).mockResolvedValue({ data: { ...mockNote, texture_variant: 2 } })

    const store = useNotesStore()
    await store.createNote('Hello world', 2)

    expect(api.post).toHaveBeenCalledWith('/notes', { content: 'Hello world', texture_variant: 2 })
  })

  it('updates a note', async () => {
    const updatedNote = { ...mockNote, content: 'Updated content' }
    vi.mocked(api.get).mockResolvedValue({ data: [mockNote] })
    vi.mocked(api.put).mockResolvedValue({ data: updatedNote })

    const store = useNotesStore()
    await store.fetchNotes()
    const result = await store.updateNote('note-1', 'Updated content')

    expect(api.put).toHaveBeenCalledWith('/notes/note-1', { content: 'Updated content', font_id: undefined })
    expect(store.notes[0].content).toBe('Updated content')
    expect(result).toEqual(updatedNote)
  })

  it('deletes a note', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: [mockNote] })
    vi.mocked(api.delete).mockResolvedValue({})

    const store = useNotesStore()
    await store.fetchNotes()
    expect(store.notes).toHaveLength(1)

    await store.deleteNote('note-1')

    expect(api.delete).toHaveBeenCalledWith('/notes/note-1')
    expect(store.notes).toHaveLength(0)
  })

  it('fetches trashed notes', async () => {
    const trashedNote = { ...mockNote, deleted_at: '2024-01-02T00:00:00Z' }
    vi.mocked(api.get).mockResolvedValue({ data: [trashedNote] })

    const store = useNotesStore()
    await store.fetchTrashed()

    expect(api.get).toHaveBeenCalledWith('/notes/trash')
    expect(store.trashedNotes).toEqual([trashedNote])
  })

  it('restores a note', async () => {
    const trashedNote = { ...mockNote, deleted_at: '2024-01-02T00:00:00Z' }
    const restoredNote = { ...mockNote, deleted_at: undefined }

    vi.mocked(api.get).mockResolvedValue({ data: [trashedNote] })
    vi.mocked(api.post).mockResolvedValue({ data: restoredNote })

    const store = useNotesStore()
    await store.fetchTrashed()
    expect(store.trashedNotes).toHaveLength(1)

    await store.restoreNote('note-1')

    expect(api.post).toHaveBeenCalledWith('/notes/note-1/restore')
    expect(store.trashedNotes).toHaveLength(0)
    expect(store.notes).toContainEqual(restoredNote)
  })

  it('permanently deletes a note', async () => {
    const trashedNote = { ...mockNote, deleted_at: '2024-01-02T00:00:00Z' }

    vi.mocked(api.get).mockResolvedValue({ data: [trashedNote] })
    vi.mocked(api.delete).mockResolvedValue({})

    const store = useNotesStore()
    await store.fetchTrashed()
    expect(store.trashedNotes).toHaveLength(1)

    await store.permanentlyDelete('note-1')

    expect(api.delete).toHaveBeenCalledWith('/notes/note-1/permanent')
    expect(store.trashedNotes).toHaveLength(0)
  })

  it('sets loading during fetchNotes', async () => {
    let resolvePromise: (value: unknown) => void
    const pending = new Promise((resolve) => { resolvePromise = resolve })
    vi.mocked(api.get).mockReturnValue(pending as never)

    const store = useNotesStore()
    const fetchPromise = store.fetchNotes()

    expect(store.loading).toBe(true)

    resolvePromise!({ data: [] })
    await fetchPromise

    expect(store.loading).toBe(false)
  })
})
