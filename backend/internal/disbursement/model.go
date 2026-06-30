package disbursement

import "github.com/shopspring/decimal"

// Worker is a payee with a pending disbursement.
type Worker struct {
	ID       string
	Name     string
	Amount   decimal.Decimal
	Currency string
}

// Status is the lifecycle state of a single disbursement.
type Status string

const (
	StatusPending Status = "pending"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

// Disbursement is the outcome of paying one worker within a batch.
type Disbursement struct {
	WorkerID      string
	Status        Status
	ProviderTxnID string
	Error         string
}

// Batch is a set of disbursements submitted together under a client-supplied ID.
type Batch struct {
	ID      string
	Results []Disbursement
}
