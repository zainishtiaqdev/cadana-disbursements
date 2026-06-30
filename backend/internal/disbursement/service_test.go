package disbursement_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"cadana/internal/disbursement"
	"cadana/internal/provider"
	"cadana/internal/store"

	"github.com/shopspring/decimal"
)

// spyProvider counts how many times each worker is paid and can be told to fail
// specific workers — enough to prove idempotency and partial-failure isolation
// without the randomness of the real mock.
type spyProvider struct {
	mu      sync.Mutex
	calls   map[string]int
	failFor map[string]bool
}

func newSpy() *spyProvider {
	return &spyProvider{calls: map[string]int{}, failFor: map[string]bool{}}
}

func (s *spyProvider) Pay(_ context.Context, req provider.PaymentRequest) (provider.PaymentResult, error) {
	s.mu.Lock()
	s.calls[req.WorkerID]++
	fail := s.failFor[req.WorkerID]
	s.mu.Unlock()
	if fail {
		return provider.PaymentResult{}, provider.ErrProviderTimeout
	}
	return provider.PaymentResult{ProviderTxnID: "ptx-" + req.WorkerID}, nil
}

func (s *spyProvider) callsFor(id string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls[id]
}

func testWorkers() []disbursement.Worker {
	return []disbursement.Worker{
		{ID: "w-1", Name: "A", Amount: decimal.RequireFromString("100.00"), Currency: "USD"},
		{ID: "w-2", Name: "B", Amount: decimal.RequireFromString("200.50"), Currency: "EUR"},
		{ID: "w-3", Name: "C", Amount: decimal.RequireFromString("300.25"), Currency: "USD"},
	}
}

// TestResubmitDoesNotDoublePay fires several concurrent submissions of the same
// new batch_id. Exactly one should win the create race and process; every
// worker must be paid exactly once.
func TestResubmitDoesNotDoublePay(t *testing.T) {
	spy := newSpy()
	svc := disbursement.NewService(store.NewMemory(), spy, testWorkers())
	ids := []string{"w-1", "w-2", "w-3"}

	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Submit(context.Background(), "batch-x", ids); err != nil {
				t.Errorf("submit: %v", err)
			}
		}()
	}
	wg.Wait()

	final := waitForBatch(t, svc, "batch-x")
	if len(final.Results) != len(ids) {
		t.Fatalf("got %d results, want %d", len(final.Results), len(ids))
	}
	for _, id := range ids {
		if got := spy.callsFor(id); got != 1 {
			t.Errorf("worker %s paid %d times, want 1", id, got)
		}
	}
}

// TestPartialFailureIsolatesWorkers proves a failed disbursement doesn't block
// the others: w-2 fails, w-1 and w-3 still succeed.
func TestPartialFailureIsolatesWorkers(t *testing.T) {
	spy := newSpy()
	spy.failFor["w-2"] = true
	svc := disbursement.NewService(store.NewMemory(), spy, testWorkers())

	if _, err := svc.Submit(context.Background(), "batch-y", []string{"w-1", "w-2", "w-3"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	batch := waitForBatch(t, svc, "batch-y")

	byID := make(map[string]disbursement.Disbursement, len(batch.Results))
	for _, d := range batch.Results {
		byID[d.WorkerID] = d
	}

	if got := byID["w-2"]; got.Status != disbursement.StatusFailed || got.Error != "provider_timeout" {
		t.Errorf("w-2 = %+v, want failed/provider_timeout", got)
	}
	for _, id := range []string{"w-1", "w-3"} {
		if got := byID[id]; got.Status != disbursement.StatusSuccess || got.ProviderTxnID == "" {
			t.Errorf("%s = %+v, want success with a txn id", id, got)
		}
	}
}

func TestSubmitRejectsUnknownWorker(t *testing.T) {
	svc := disbursement.NewService(store.NewMemory(), newSpy(), testWorkers())
	_, err := svc.Submit(context.Background(), "batch-z", []string{"w-1", "ghost"})
	var unknown disbursement.ErrUnknownWorker
	if !errors.As(err, &unknown) {
		t.Fatalf("err = %v, want ErrUnknownWorker", err)
	}
	if unknown.WorkerID != "ghost" {
		t.Errorf("unknown worker = %q, want ghost", unknown.WorkerID)
	}
}

// waitForBatch polls until the batch has no pending results. Processing is
// asynchronous, so tests observe it the same way the frontend does.
func waitForBatch(t *testing.T, svc *disbursement.Service, batchID string) disbursement.Batch {
	t.Helper()
	for range 200 {
		batch, found, err := svc.Get(context.Background(), batchID)
		if err != nil {
			t.Fatalf("get batch: %v", err)
		}
		if found && settled(batch) {
			return batch
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("batch %s did not settle in time", batchID)
	return disbursement.Batch{}
}

func settled(b disbursement.Batch) bool {
	if len(b.Results) == 0 {
		return false
	}
	for _, d := range b.Results {
		if d.Status == disbursement.StatusPending {
			return false
		}
	}
	return true
}
