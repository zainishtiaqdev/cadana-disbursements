import type { Batch, SubmitRequest, Worker } from '../types/api'

const BASE = import.meta.env.VITE_API_BASE ?? ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const message = await res
      .json()
      .then((body: { error?: string }) => body.error)
      .catch(() => undefined)
    throw new Error(message ?? `Request failed (${res.status})`)
  }
  return res.json() as Promise<T>
}

export function getWorkers(): Promise<Worker[]> {
  return request<Worker[]>('/workers')
}

export function submitDisbursements(body: SubmitRequest): Promise<Batch> {
  return request<Batch>('/disbursements', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function getBatch(batchId: string): Promise<Batch> {
  return request<Batch>(`/disbursements/${batchId}`)
}
