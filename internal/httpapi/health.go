package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"smartseller-lite-starter/internal/db"
)

type HealthResponse struct {
	App  string `json:"app"`
	DB   string `json:"db"`
	Time string `json:"time"`
}

// HealthHandler reports simple application + database status.
func HealthHandler(appName string, store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "down"
		if store != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
			if err := store.DB().PingContext(ctx); err == nil {
				status = "ok"
			}
			cancel()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{
			App:  appName,
			DB:   status,
			Time: time.Now().Format(time.RFC3339),
		})
	}
}
