import { computed, onScopeDispose, ref } from 'vue'
import type { Batch, DisbursementResult } from '../types/api'
import { getBatch, submitDisbursements } from '../api/client'

const POLL_INTERVAL_MS = 1200

// Owns a single batch's lifecycle: submit, then poll until every row settles.
// State that is raw server data lives in `ref`s; everything derived (the
// summary, the failed set) is a `computed`, so it can never drift out of sync.
export function useDisbursementBatch() {
  const batchId = ref<string | null>(null)
  const results = ref<DisbursementResult[]>([])
  const submitting = ref(false)
  const polling = ref(false)
  const error = ref<string | null>(null)

  let timer: ReturnType<typeof setInterval> | null = null

  const summary = computed(() => {
    const counts = { total: results.value.length, success: 0, failed: 0, pending: 0 }
    for (const r of results.value) counts[r.status]++
    return counts
  })

  const failedWorkerIds = computed(() =>
    results.value.filter((r) => r.status === 'failed').map((r) => r.worker_id),
  )

  const hasPending = computed(() => results.value.some((r) => r.status === 'pending'))

  function stopPolling() {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
    polling.value = false
  }

  function startPolling() {
    if (timer !== null || batchId.value === null) return
    polling.value = true
    timer = setInterval(async () => {
      if (batchId.value === null) return
      try {
        const batch = await getBatch(batchId.value)
        results.value = batch.results
        if (!hasPending.value) stopPolling()
      } catch (e) {
        error.value = e instanceof Error ? e.message : 'Lost connection while polling'
      }
    }, POLL_INTERVAL_MS)
  }

  async function submit(workerIds: string[]) {
    if (workerIds.length === 0) return
    submitting.value = true
    error.value = null
    stopPolling()
    // Client-supplied batch id doubles as the idempotency key.
    const id = `batch-${crypto.randomUUID()}`
    try {
      const batch: Batch = await submitDisbursements({ batch_id: id, worker_ids: workerIds })
      batchId.value = batch.batch_id
      results.value = batch.results
      if (hasPending.value) startPolling()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to submit disbursements'
    } finally {
      submitting.value = false
    }
  }

  // Retrying re-batches only the failed workers under a fresh id. Already-paid
  // workers are never resubmitted, so there's nothing to double-pay.
  function retryFailed() {
    return submit(failedWorkerIds.value)
  }

  onScopeDispose(stopPolling)

  return {
    batchId,
    results,
    submitting,
    polling,
    error,
    summary,
    failedWorkerIds,
    submit,
    retryFailed,
  }
}
