// Package http translates HTTP requests into calls on the app layer's use
// cases and their results back into JSON — the domain and app packages
// know nothing about net/http.
package http

import (
	"net/http"

	"investment-map/internal/app"
)

type Handler struct {
	query *app.QueryService
}

func NewRouter(query *app.QueryService) http.Handler {
	h := &Handler{query: query}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.HandleFunc("GET /api/quotes", h.handleListQuotes)
	mux.HandleFunc("GET /api/quotes/{symbol}", h.handleGetQuote)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
