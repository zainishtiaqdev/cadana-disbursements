package api

import "net/http"

// NewRouter wires routes (stdlib method-aware patterns, Go 1.22+) behind a CORS
// middleware so the deployed frontend can call the API cross-origin.
func NewRouter(h *Handler, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /workers", h.Workers)
	mux.HandleFunc("POST /disbursements", h.Submit)
	mux.HandleFunc("GET /disbursements/{batch_id}", h.GetBatch)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return cors(allowedOrigin)(mux)
}

func cors(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
