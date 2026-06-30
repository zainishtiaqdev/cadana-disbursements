package store

import (
	"context"
	_ "embed"
	"fmt"

	"cadana/internal/disbursement"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

// Postgres is a durable Store (e.g. Supabase). Idempotency is enforced by the
// primary key on batches(id): CreateBatch inserts ON CONFLICT DO NOTHING, so a
// resubmitted batch_id never re-creates rows or re-triggers payment.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) CreateBatch(ctx context.Context, batchID string, workerIDs []string) (bool, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `INSERT INTO batches (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, batchID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil // batch already exists -> idempotent no-op
	}

	rows := make([][]any, len(workerIDs))
	for i, id := range workerIDs {
		rows[i] = []any{batchID, id, i, string(disbursement.StatusPending)}
	}
	if _, err := tx.CopyFrom(ctx,
		pgx.Identifier{"disbursements"},
		[]string{"batch_id", "worker_id", "ord", "status"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (p *Postgres) UpdateResult(ctx context.Context, batchID string, d disbursement.Disbursement) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE disbursements SET status = $1, provider_txn_id = $2, error = $3
		 WHERE batch_id = $4 AND worker_id = $5`,
		string(d.Status), nullify(d.ProviderTxnID), nullify(d.Error), batchID, d.WorkerID)
	return err
}

func (p *Postgres) GetBatch(ctx context.Context, batchID string) (disbursement.Batch, bool, error) {
	var exists bool
	if err := p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM batches WHERE id = $1)`, batchID).Scan(&exists); err != nil {
		return disbursement.Batch{}, false, err
	}
	if !exists {
		return disbursement.Batch{}, false, nil
	}

	rows, err := p.pool.Query(ctx,
		`SELECT worker_id, status, COALESCE(provider_txn_id, ''), COALESCE(error, '')
		 FROM disbursements WHERE batch_id = $1 ORDER BY ord`, batchID)
	if err != nil {
		return disbursement.Batch{}, false, err
	}
	defer rows.Close()

	var results []disbursement.Disbursement
	for rows.Next() {
		var d disbursement.Disbursement
		var status string
		if err := rows.Scan(&d.WorkerID, &status, &d.ProviderTxnID, &d.Error); err != nil {
			return disbursement.Batch{}, false, err
		}
		d.Status = disbursement.Status(status)
		results = append(results, d)
	}
	return disbursement.Batch{ID: batchID, Results: results}, true, rows.Err()
}

// nullify maps an empty string to SQL NULL so optional columns stay clean.
func nullify(s string) any {
	if s == "" {
		return nil
	}
	return s
}
