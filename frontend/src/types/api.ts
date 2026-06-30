// Wire contract — mirrors the Go DTOs in backend/internal/api/handler.go.
// Hand-written rather than generated: the surface is tiny and stable, so one
// source of truth here is simpler than an OpenAPI codegen step.

export interface Worker {
  id: string
  name: string
  amount: string // decimal as a string; never parsed into a float
  currency: string
}

export type DisbursementStatus = 'pending' | 'success' | 'failed'

export interface DisbursementResult {
  worker_id: string
  status: DisbursementStatus
  provider_txn_id?: string
  error?: string
}

export interface Batch {
  batch_id: string
  results: DisbursementResult[]
}

export interface SubmitRequest {
  batch_id: string
  worker_ids: string[]
}
