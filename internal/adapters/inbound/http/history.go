package http

import (
	"fmt"
	"net/http"

	"investment-map/internal/domain"
)

// defaultHistoryRange matches the scrubber's "last 30-90 days" spec.
const defaultHistoryRange = domain.HistoryRange3Month

func (h *Handler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")

	rng, err := parseHistoryRange(r.URL.Query().Get("range"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	points, err := h.query.GetHistory(r.Context(), symbol, rng)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func parseHistoryRange(raw string) (domain.HistoryRange, error) {
	if raw == "" {
		return defaultHistoryRange, nil
	}
	switch domain.HistoryRange(raw) {
	case domain.HistoryRange1Month, domain.HistoryRange3Month, domain.HistoryRange1Year, domain.HistoryRangeMax:
		return domain.HistoryRange(raw), nil
	default:
		return "", fmt.Errorf("invalid range %q (expected one of: 1mo, 3mo, 1y, max)", raw)
	}
}
