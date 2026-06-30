<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useWorkers } from './composables/useWorkers'
import { useDisbursementBatch } from './composables/useDisbursementBatch'
import WorkerList from './components/WorkerList.vue'
import DisbursementTable from './components/DisbursementTable.vue'
import BatchSummary from './components/BatchSummary.vue'

const { workers, loading, error: workersError, load } = useWorkers()
const {
  batchId,
  results,
  submitting,
  polling,
  error: batchError,
  summary,
  failedWorkerIds,
  submit,
  retryFailed,
} = useDisbursementBatch()

const selected = ref<Set<string>>(new Set())

const nameById = computed(() => Object.fromEntries(workers.value.map((w) => [w.id, w.name])))
const hasBatch = computed(() => batchId.value !== null)
const canRetry = computed(
  () => failedWorkerIds.value.length > 0 && !submitting.value && summary.value.pending === 0,
)

function toggle(id: string) {
  const next = new Set(selected.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selected.value = next
}

function toggleAll() {
  selected.value =
    selected.value.size === workers.value.length
      ? new Set()
      : new Set(workers.value.map((w) => w.id))
}

function disburse() {
  void submit([...selected.value])
}

onMounted(load)
</script>

<template>
  <div class="app">
    <header class="topbar">
      <h1>Cadana · Disbursements</h1>
      <p class="sub">Review pending payouts, run a batch, and watch results land live.</p>
    </header>

    <main class="grid">
      <WorkerList
        :workers="workers"
        :loading="loading"
        :error="workersError"
        :selected="selected"
        :submitting="submitting"
        @toggle="toggle"
        @toggle-all="toggleAll"
        @reload="load"
        @disburse="disburse"
      />

      <section class="panel">
        <header class="panel-head"><h2>Batch status</h2></header>

        <div v-if="!hasBatch" class="empty">
          Select workers and hit <strong>Disburse</strong> to start a batch.
        </div>

        <template v-else>
          <p v-if="batchError" class="error-inline">{{ batchError }}</p>
          <BatchSummary
            :summary="summary"
            :polling="polling"
            :can-retry="canRetry"
            @retry="retryFailed"
          />
          <DisbursementTable :results="results" :name-by-id="nameById" />
        </template>
      </section>
    </main>
  </div>
</template>
