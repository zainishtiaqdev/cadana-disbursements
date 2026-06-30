<script setup lang="ts">
defineProps<{
  summary: { total: number; success: number; failed: number; pending: number }
  polling: boolean
  canRetry: boolean
}>()

const emit = defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="summary">
    <div class="stats">
      <span class="stat success">{{ summary.success }} succeeded</span>
      <span class="stat failed">{{ summary.failed }} failed</span>
      <span v-if="summary.pending" class="stat pending">{{ summary.pending }} pending</span>
    </div>
    <div class="actions">
      <span v-if="polling" class="polling"><span class="spinner" />updating…</span>
      <button v-if="canRetry" class="ghost danger" @click="emit('retry')">Retry failed</button>
    </div>
  </div>
</template>

<style scoped>
.summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.stats {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.stat {
  font-size: 0.85rem;
  font-weight: 600;
  padding: 0.25rem 0.65rem;
  border-radius: 7px;
}
.stat.success {
  background: #dcfce7;
  color: #166534;
}
.stat.failed {
  background: #fee2e2;
  color: #991b1b;
}
.stat.pending {
  background: #fef3c7;
  color: #92400e;
}
.actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.polling {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.82rem;
  color: #64748b;
}
.spinner {
  width: 0.8rem;
  height: 0.8rem;
  border: 2px solid #cbd5e1;
  border-top-color: #2563eb;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
