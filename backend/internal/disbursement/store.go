package disbursement

import "context"

// Store persists batches and their per-worker results. It is the port the
// service depends on; the adapters (in-memory, Postgres) live in internal/store.
//
// CreateBatch is the idempotency boundary: it must atomically create the batch
// with every result pending, and report created=false if the ID already exists.
type Store interface {
	CreateBatch(ctx context.Context, batchID string, workerIDs []string) (created bool, err error)
	UpdateResult(ctx context.Context, batchID string, d Disbursement) error
	GetBatch(ctx context.Context, batchID string) (batch Batch, found bool, err error)
}
