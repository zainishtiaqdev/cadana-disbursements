# Demo walkthrough

A 5-minute path through the app, with the talking points that map each step to what the exercise is
grading. Run `docs/architecture.html` alongside this for the diagrams.

## Setup (two terminals)

```bash
# terminal 1 — backend
cd backend && make run

# terminal 2 — frontend
cd frontend && npm install && npm run dev
```

Open <http://localhost:5173>.

---

## The click-through

1. **Workers load.** Nine workers across USD / EUR / GBP, amounts formatted per currency.
   - *Say:* amounts are `decimal` on the server and strings on the wire — no floats touch money.
     The list is a `ref` filled by the `useWorkers` composable; the loading state is a skeleton, not a spinner-in-the-void.

2. **Select a few and hit Disburse.** Rows immediately show **Pending**, a subtle “updating…” spinner appears.
   - *Say:* `POST` returns `202` instantly with everything pending and processes in the background. The UI
     polls until nothing is pending — that's why the provider's latency and failures are *visible*.

3. **Rows settle** into **Success** (with a `ptx-…` txn id) and **Failed** (`provider_timeout`). The summary
   bar updates live (X succeeded, Y failed).
   - *Say:* one goroutine per worker, bounded by a semaphore, joined by a `WaitGroup`. A failure on one
     worker never blocks the others — partial failure is isolated. Everything runs clean under `-race`.

4. **Click “Retry failed.”** Only the failed workers go again, under a brand-new batch id.
   - *Say:* already-paid workers are never resubmitted, so there's nothing to double-pay. Retry is just a new
     idempotent batch.

5. **Idempotency, on the API.** In a terminal: `./demo.sh` — it submits a batch, then resubmits the same
   `batch_id` and proves the results (and txn ids) are identical: no double payment.
   - *Say:* `batch_id` is a client-supplied idempotency key. Create-is-atomic; resubmit returns the existing
     batch and never calls the provider again.

6. **(Optional) Durability.** Set `DATABASE_URL`, restart the server mid-demo, and the batch is still there —
   idempotency is enforced by the `batches(id)` primary key, so it survives restarts and multiple instances.

7. **CI.** Open the **Actions** tab on GitHub: `go test -race` runs on every push, with the pass/fail report
   in the run summary.

---

## Rubric → where to point

| What they're grading | Show this |
|---|---|
| Code organization | `backend/internal/{api,disbursement,store,provider}`; FE `composables/` vs `components/` |
| Type safety API ↔ FE | `frontend/src/types/api.ts` mirrors the Go DTOs in `api/handler.go` |
| Reactivity choices | `useDisbursementBatch` — `ref` for server data, `computed` summary, polling cleanup |
| Async / failure UX | skeleton → pending → live poll → success/failed → retry |
| Concurrency in Go | `service.go` `process()` — goroutines + semaphore + `WaitGroup`; `-race` |
| Idempotency | `batch_id` create-if-absent (mutex in memory, `ON CONFLICT` in Postgres) |
| Decimal money | `shopspring/decimal`, string on the wire, `Intl.NumberFormat` on the FE |
| README / trade-off | polling vs SSE, in-memory-default vs durable Postgres |
