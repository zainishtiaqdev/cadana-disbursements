package disbursement

import (
	"context"
	"errors"
	"log"
	"sync"

	"cadana/internal/provider"
)

const defaultConcurrency = 5

// Validation errors are surfaced to the client as 400s.
var (
	ErrBatchIDRequired = errors.New("batch_id is required")
	ErrNoWorkers       = errors.New("no valid workers in submission")
)

// ErrUnknownWorker is returned when a submitted worker ID isn't on the roster.
type ErrUnknownWorker struct{ WorkerID string }

func (e ErrUnknownWorker) Error() string { return "unknown worker: " + e.WorkerID }

// IsValidation reports whether err is a client-input error (vs. an internal one).
func IsValidation(err error) bool {
	var unknown ErrUnknownWorker
	return errors.Is(err, ErrBatchIDRequired) ||
		errors.Is(err, ErrNoWorkers) ||
		errors.As(err, &unknown)
}

// Service holds the business logic: it validates submissions, enforces
// idempotency through the store, and pays a batch concurrently.
type Service struct {
	store       Store
	provider    provider.PaymentProvider
	roster      []Worker
	byID        map[string]Worker
	concurrency int
}

func NewService(store Store, prov provider.PaymentProvider, workers []Worker) *Service {
	byID := make(map[string]Worker, len(workers))
	for _, w := range workers {
		byID[w.ID] = w
	}
	return &Service{
		store:       store,
		provider:    prov,
		roster:      workers,
		byID:        byID,
		concurrency: defaultConcurrency,
	}
}

// Workers returns the roster in stable display order.
func (s *Service) Workers() []Worker { return s.roster }

// Submit creates a batch and pays it concurrently. Resubmitting an existing
// batch_id is a no-op that returns the current state — the idempotency
// guarantee that prevents double payment.
func (s *Service) Submit(ctx context.Context, batchID string, workerIDs []string) (Batch, error) {
	if batchID == "" {
		return Batch{}, ErrBatchIDRequired
	}
	workers, err := s.resolve(workerIDs)
	if err != nil {
		return Batch{}, err
	}

	created, err := s.store.CreateBatch(ctx, batchID, idsOf(workers))
	if err != nil {
		return Batch{}, err
	}
	if created {
		// Detach from the request context so processing isn't cancelled when
		// the HTTP handler returns; keep request values for tracing.
		go s.process(context.WithoutCancel(ctx), batchID, workers)
	}

	batch, _, err := s.store.GetBatch(ctx, batchID)
	return batch, err
}

// Get returns the current state of a batch.
func (s *Service) Get(ctx context.Context, batchID string) (Batch, bool, error) {
	return s.store.GetBatch(ctx, batchID)
}

// resolve validates worker IDs against the roster and de-duplicates them while
// preserving submission order.
func (s *Service) resolve(workerIDs []string) ([]Worker, error) {
	seen := make(map[string]bool, len(workerIDs))
	workers := make([]Worker, 0, len(workerIDs))
	for _, id := range workerIDs {
		if seen[id] {
			continue
		}
		w, ok := s.byID[id]
		if !ok {
			return nil, ErrUnknownWorker{WorkerID: id}
		}
		seen[id] = true
		workers = append(workers, w)
	}
	if len(workers) == 0 {
		return nil, ErrNoWorkers
	}
	return workers, nil
}

// process pays every worker in the batch concurrently, bounded by a semaphore
// so we don't overwhelm the provider. A failure on one worker never blocks the
// others.
func (s *Service) process(ctx context.Context, batchID string, workers []Worker) {
	sem := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		sem <- struct{}{}
		go func(w Worker) {
			defer wg.Done()
			defer func() { <-sem }()
			s.pay(ctx, batchID, w)
		}(w)
	}
	wg.Wait()
}

func (s *Service) pay(ctx context.Context, batchID string, w Worker) {
	result := Disbursement{WorkerID: w.ID}
	res, err := s.provider.Pay(ctx, provider.PaymentRequest{
		WorkerID: w.ID,
		Amount:   w.Amount,
		Currency: w.Currency,
	})
	if err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
	} else {
		result.Status = StatusSuccess
		result.ProviderTxnID = res.ProviderTxnID
	}
	if err := s.store.UpdateResult(ctx, batchID, result); err != nil {
		log.Printf("update result batch=%s worker=%s: %v", batchID, w.ID, err)
	}
}

func idsOf(workers []Worker) []string {
	ids := make([]string, len(workers))
	for i, w := range workers {
		ids[i] = w.ID
	}
	return ids
}
