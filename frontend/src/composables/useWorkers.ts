import { ref } from 'vue'
import type { Worker } from '../types/api'
import { getWorkers } from '../api/client'

// Owns the roster fetch and its loading/error state. Extracted as a composable
// so the view stays declarative and the async concern is testable in isolation.
export function useWorkers() {
  const workers = ref<Worker[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      workers.value = await getWorkers()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load workers'
    } finally {
      loading.value = false
    }
  }

  return { workers, loading, error, load }
}
