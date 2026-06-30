package disbursement

import "github.com/shopspring/decimal"

// SeedWorkers returns the roster of workers with pending disbursements:
// decimal amounts across three currencies, in stable display order.
func SeedWorkers() []Worker {
	rows := []struct{ id, name, amount, currency string }{
		{"w-001", "Ada Lovelace", "1500.50", "USD"},
		{"w-002", "Linus Torvalds", "2300.00", "EUR"},
		{"w-003", "Grace Hopper", "4200.75", "USD"},
		{"w-004", "Alan Turing", "3100.00", "GBP"},
		{"w-005", "Margaret Hamilton", "2750.20", "USD"},
		{"w-006", "Dennis Ritchie", "1899.99", "EUR"},
		{"w-007", "Katherine Johnson", "3650.00", "USD"},
		{"w-008", "Edsger Dijkstra", "2100.40", "EUR"},
		{"w-009", "Barbara Liskov", "5400.10", "GBP"},
	}
	workers := make([]Worker, len(rows))
	for i, r := range rows {
		workers[i] = Worker{
			ID:       r.id,
			Name:     r.name,
			Amount:   decimal.RequireFromString(r.amount),
			Currency: r.currency,
		}
	}
	return workers
}
