package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

// ErrProviderTimeout is the failure surfaced by the mock's simulated flakiness.
var ErrProviderTimeout = errors.New("provider_timeout")

// Mock is a flaky payment rail: random latency plus a failure rate. rand/v2's
// top-level functions are safe for concurrent use, so a batch can call Pay from
// many goroutines at once.
type Mock struct {
	failureRate float64
	minLatency  time.Duration
	maxLatency  time.Duration
}

func NewMock() *Mock {
	return &Mock{
		failureRate: 0.30,
		minLatency:  50 * time.Millisecond,
		maxLatency:  200 * time.Millisecond,
	}
}

func (m *Mock) Pay(ctx context.Context, req PaymentRequest) (PaymentResult, error) {
	latency := m.minLatency + time.Duration(rand.Int64N(int64(m.maxLatency-m.minLatency)))
	select {
	case <-ctx.Done():
		return PaymentResult{}, ctx.Err()
	case <-time.After(latency):
	}

	if rand.Float64() < m.failureRate {
		return PaymentResult{}, ErrProviderTimeout
	}
	return PaymentResult{ProviderTxnID: fmt.Sprintf("ptx-%08x", rand.Uint32())}, nil
}
