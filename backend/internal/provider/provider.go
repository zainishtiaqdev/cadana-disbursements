package provider

import (
	"context"

	"github.com/shopspring/decimal"
)

// PaymentRequest is a single disbursement handed to the downstream rail.
type PaymentRequest struct {
	WorkerID string
	Amount   decimal.Decimal
	Currency string
}

// PaymentResult is returned on a successful charge.
type PaymentResult struct {
	ProviderTxnID string
}

// PaymentProvider is the downstream payment rail. It is an interface so the
// service can run against a deterministic stub in tests.
type PaymentProvider interface {
	Pay(ctx context.Context, req PaymentRequest) (PaymentResult, error)
}
