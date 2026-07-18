package http

import "net/http"

func (h *Handler) handleListQuotes(w http.ResponseWriter, r *http.Request) {
	quotes, err := h.query.ListQuotes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, quotes)
}

func (h *Handler) handleGetQuote(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")

	quote, err := h.query.GetQuote(r.Context(), symbol)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, quote)
}
