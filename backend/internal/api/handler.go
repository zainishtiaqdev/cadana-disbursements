package api

import (
	"encoding/json"
	"log"
	"net/http"

	"cadana/internal/disbursement"
)

// Handler adapts HTTP to the disbursement service. It owns the wire DTOs and
// keeps mapping logic out of the domain.
type Handler struct {
	svc *disbursement.Service
}

func NewHandler(svc *disbursement.Service) *Handler { return &Handler{svc: svc} }

// --- wire contract (mirrored in frontend/src/types/api.ts) ---

type workerDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type submitRequest struct {
	BatchID   string   `json:"batch_id"`
	WorkerIDs []string `json:"worker_ids"`
}

type resultDTO struct {
	WorkerID      string `json:"worker_id"`
	Status        string `json:"status"`
	ProviderTxnID string `json:"provider_txn_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type batchDTO struct {
	BatchID string      `json:"batch_id"`
	Results []resultDTO `json:"results"`
}

func (h *Handler) Workers(w http.ResponseWriter, _ *http.Request) {
	workers := h.svc.Workers()
	out := make([]workerDTO, len(workers))
	for i, wk := range workers {
		out[i] = workerDTO{
			ID:       wk.ID,
			Name:     wk.Name,
			Amount:   wk.Amount.StringFixed(2),
			Currency: wk.Currency,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	batch, err := h.svc.Submit(r.Context(), req.BatchID, req.WorkerIDs)
	if err != nil {
		if disbursement.IsValidation(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("submit batch=%s: %v", req.BatchID, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// 202: work has been accepted and is processing; poll GET for results.
	writeJSON(w, http.StatusAccepted, toBatchDTO(batch))
}

func (h *Handler) GetBatch(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batch_id")
	batch, found, err := h.svc.Get(r.Context(), batchID)
	if err != nil {
		log.Printf("get batch=%s: %v", batchID, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	writeJSON(w, http.StatusOK, toBatchDTO(batch))
}

func toBatchDTO(b disbursement.Batch) batchDTO {
	results := make([]resultDTO, len(b.Results))
	for i, d := range b.Results {
		results[i] = resultDTO{
			WorkerID:      d.WorkerID,
			Status:        string(d.Status),
			ProviderTxnID: d.ProviderTxnID,
			Error:         d.Error,
		}
	}
	return batchDTO{BatchID: b.ID, Results: results}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
