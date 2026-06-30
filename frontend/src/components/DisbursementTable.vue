<script setup lang="ts">
import type { DisbursementResult } from '../types/api'
import StatusBadge from './StatusBadge.vue'

defineProps<{ results: DisbursementResult[]; nameById: Record<string, string> }>()
</script>

<template>
  <table class="table">
    <thead>
      <tr>
        <th>Worker</th>
        <th>Status</th>
        <th>Detail</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="r in results" :key="r.worker_id">
        <td>{{ nameById[r.worker_id] ?? r.worker_id }}</td>
        <td><StatusBadge :status="r.status" /></td>
        <td class="detail">
          <code v-if="r.provider_txn_id">{{ r.provider_txn_id }}</code>
          <span v-else-if="r.error" class="err">{{ r.error }}</span>
          <span v-else class="muted">—</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<style scoped>
.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
th {
  text-align: left;
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #94a3b8;
  padding: 0.4rem 0.5rem;
  border-bottom: 1px solid #e2e8f0;
}
td {
  padding: 0.6rem 0.5rem;
  border-bottom: 1px solid #f1f5f9;
}
.detail code {
  font-size: 0.8rem;
  color: #475569;
  background: #f1f5f9;
  padding: 0.1rem 0.4rem;
  border-radius: 5px;
}
.err {
  color: #b91c1c;
  font-size: 0.85rem;
}
.muted {
  color: #cbd5e1;
}
</style>
