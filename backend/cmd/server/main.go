package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cadana/internal/api"
	"cadana/internal/disbursement"
	"cadana/internal/provider"
	"cadana/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st, closeStore, err := buildStore(ctx)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer closeStore()

	svc := disbursement.NewService(st, provider.NewMock(), disbursement.SeedWorkers())
	router := api.NewRouter(api.NewHandler(svc), env("ALLOWED_ORIGIN", "*"))

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// buildStore selects Postgres when DATABASE_URL is set, otherwise the in-memory
// store (the zero-setup default).
func buildStore(ctx context.Context) (disbursement.Store, func(), error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("store: in-memory (set DATABASE_URL for Postgres)")
		return store.NewMemory(), func() {}, nil
	}
	pg, err := store.NewPostgres(ctx, dsn)
	if err != nil {
		return nil, nil, err
	}
	log.Println("store: postgres")
	return pg, pg.Close, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
