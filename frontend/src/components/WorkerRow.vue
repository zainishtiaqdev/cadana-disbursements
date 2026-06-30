<script setup lang="ts">
import { computed } from 'vue'
import type { Worker } from '../types/api'

const props = defineProps<{ worker: Worker; selected: boolean; disabled: boolean }>()
const emit = defineEmits<{ toggle: [id: string] }>()

const amount = computed(() =>
  new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: props.worker.currency,
  }).format(Number(props.worker.amount)),
)
</script>

<template>
  <label class="row" :class="{ selected, disabled }">
    <input
      type="checkbox"
      :checked="selected"
      :disabled="disabled"
      @change="emit('toggle', worker.id)"
    />
    <span class="name">{{ worker.name }}</span>
    <span class="amount">{{ amount }}</span>
    <span class="currency">{{ worker.currency }}</span>
  </label>
</template>

<style scoped>
.row {
  display: grid;
  grid-template-columns: auto 1fr auto auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.12s ease;
}
.row:hover {
  background: #f8fafc;
}
.row.selected {
  background: #eff6ff;
}
.row.disabled {
  cursor: default;
  opacity: 0.65;
}
.name {
  font-weight: 500;
}
.amount {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.currency {
  font-size: 0.75rem;
  color: #64748b;
  width: 2.5rem;
  text-align: right;
}
input {
  width: 1.05rem;
  height: 1.05rem;
  accent-color: #2563eb;
}
</style>
