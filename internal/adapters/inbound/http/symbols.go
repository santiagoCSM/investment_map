package http

import "net/http"

func (h *Handler) handleListSymbols(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.query.ListSymbols())
}
