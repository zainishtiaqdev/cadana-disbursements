package store

import (
	"context"
	"fmt"
	"sync"

	"cadana/internal/disbursement"
)

// Memory is the zero-dependency default Store, guarded by a single RWMutex so a
// batch's concurrent result writes stay race-free.
type Memory struct {
	mu      sync.RWMutex
	batches map[string]*memBatch
}

type memBatch struct {
	order   []string // worker IDs in submission order, for stable output
	results map[string]disbursement.Disbursement
}

func NewMemory() *Memory {
	return &Memory{batches: make(map[string]*memBatch)}
}

func (m *Memory) CreateBatch(_ context.Context, batchID string, workerIDs []string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.batches[batchID]; exists {
		return false, nil
	}
	b := &memBatch{
		order:   append([]string(nil), workerIDs...),
		results: make(map[string]disbursement.Disbursement, len(workerIDs)),
	}
	for _, id := range workerIDs {
		b.results[id] = disbursement.Disbursement{WorkerID: id, Status: disbursement.StatusPending}
	}
	m.batches[batchID] = b
	return true, nil
}

func (m *Memory) UpdateResult(_ context.Context, batchID string, d disbursement.Disbursement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.batches[batchID]
	if !ok {
		return fmt.Errorf("batch not found: %s", batchID)
	}
	if _, ok := b.results[d.WorkerID]; !ok {
		return fmt.Errorf("worker %s not in batch %s", d.WorkerID, batchID)
	}
	b.results[d.WorkerID] = d
	return nil
}

func (m *Memory) GetBatch(_ context.Context, batchID string) (disbursement.Batch, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.batches[batchID]
	if !ok {
		return disbursement.Batch{}, false, nil
	}
	results := make([]disbursement.Disbursement, 0, len(b.order))
	for _, id := range b.order {
		results = append(results, b.results[id])
	}
	return disbursement.Batch{ID: batchID, Results: results}, true, nil
}
