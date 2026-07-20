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

// NewRouter wires the REST API and, if staticDir is non-empty, serves the
// built frontend from that directory for everything else — so in
// production a single binary + a copy of web/dist is the whole deployable
// unit, no separate frontend host needed. Pass "" to skip static serving
// entirely (e.g. in tests, where there's nothing to serve).
func NewRouter(query *app.QueryService, staticDir string) http.Handler {
	h := &Handler{query: query}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.HandleFunc("GET /api/quotes", h.handleListQuotes)
	mux.HandleFunc("GET /api/quotes/{symbol}", h.handleGetQuote)
	mux.HandleFunc("GET /api/history/{symbol}", h.handleGetHistory)
	mux.HandleFunc("GET /api/symbols", h.handleListSymbols)

	if staticDir != "" {
		mux.Handle("/", spaFileServer(staticDir))
	}
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
