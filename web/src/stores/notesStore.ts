import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '../api/client'

export interface Note {
  id: string
  user_id: string
  content: string
  font_id?: string
  deleted_at?: string
  created_at: string
  updated_at: string
}

export const useNotesStore = defineStore('notes', () => {
  const notes = ref<Note[]>([])
  const trashedNotes = ref<Note[]>([])
  const loading = ref(false)

  async function fetchNotes() {
    loading.value = true
    try {
      const { data } = await api.get('/notes')
      notes.value = data
    } finally {
      loading.value = false
    }
  }

  async function createNote(content: string, fontId?: string) {
    const { data } = await api.post('/notes', { content, font_id: fontId })
    notes.value.push(data)
    return data
  }

  async function updateNote(id: string, content: string, fontId?: string) {
    const { data } = await api.put(`/notes/${id}`, { content, font_id: fontId })
    const index = notes.value.findIndex((n) => n.id === id)
    if (index !== -1) {
      notes.value[index] = data
    }
    return data
  }

  async function deleteNote(id: string) {
    await api.delete(`/notes/${id}`)
    notes.value = notes.value.filter((n) => n.id !== id)
  }

  async function fetchTrashed() {
    loading.value = true
    try {
      const { data } = await api.get('/notes/trash')
      trashedNotes.value = data
    } finally {
      loading.value = false
    }
  }

  async function restoreNote(id: string) {
    const { data } = await api.post(`/notes/${id}/restore`)
    trashedNotes.value = trashedNotes.value.filter((n) => n.id !== id)
    notes.value.push(data)
    return data
  }

  return {
    notes,
    trashedNotes,
    loading,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    fetchTrashed,
    restoreNote,
  }
})
