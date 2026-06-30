<script setup lang="ts">
import type { DisbursementStatus } from '../types/api'

defineProps<{ status: DisbursementStatus }>()

const LABELS: Record<DisbursementStatus, string> = {
  pending: 'Pending',
  success: 'Success',
  failed: 'Failed',
}
</script>

<template>
  <span class="badge" :class="status">
    <span class="dot" />
    {{ LABELS[status] }}
  </span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 600;
  white-space: nowrap;
}
.dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
}
.pending {
  background: #fef3c7;
  color: #92400e;
}
.pending .dot {
  background: #d97706;
  animation: pulse 1.2s ease-in-out infinite;
}
.success {
  background: #dcfce7;
  color: #166534;
}
.success .dot {
  background: #16a34a;
}
.failed {
  background: #fee2e2;
  color: #991b1b;
}
.failed .dot {
  background: #dc2626;
}
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.3;
  }
}
</style>
