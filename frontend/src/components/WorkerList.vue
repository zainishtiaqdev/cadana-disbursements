<script setup lang="ts">
import { computed } from 'vue'
import type { Worker } from '../types/api'
import WorkerRow from './WorkerRow.vue'

const props = defineProps<{
  workers: Worker[]
  loading: boolean
  error: string | null
  selected: Set<string>
  submitting: boolean
}>()

const emit = defineEmits<{
  toggle: [id: string]
  toggleAll: []
  reload: []
  disburse: []
}>()

const allSelected = computed(
  () => props.workers.length > 0 && props.selected.size === props.workers.length,
)
const canDisburse = computed(() => props.selected.size > 0 && !props.submitting)
</script>

<template>
  <section class="panel">
    <header class="panel-head">
      <h2>Pending payouts</h2>
      <button v-if="!loading && !error" class="link" @click="emit('toggleAll')">
        {{ allSelected ? 'Clear' : 'Select all' }}
      </button>
    </header>

    <div v-if="loading" class="skeleton-list">
      <div v-for="n in 6" :key="n" class="skeleton-row" />
    </div>

    <div v-else-if="error" class="error-card">
      <p>{{ error }}</p>
      <button class="ghost" @click="emit('reload')">Retry</button>
    </div>

    <div v-else class="rows">
      <WorkerRow
        v-for="w in workers"
        :key="w.id"
        :worker="w"
        :selected="selected.has(w.id)"
        :disabled="submitting"
        @toggle="emit('toggle', $event)"
      />
    </div>

    <footer v-if="!loading && !error" class="panel-foot">
      <button class="primary" :disabled="!canDisburse" @click="emit('disburse')">
        {{ submitting ? 'Submitting…' : `Disburse${selected.size ? ` ${selected.size}` : ''}` }}
      </button>
    </footer>
  </section>
</template>
