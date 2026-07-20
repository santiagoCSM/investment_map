package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"investment-map/internal/app"
	"investment-map/internal/domain"
)

func TestSpaFileServer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "<html>app shell</html>")
	writeFile(t, filepath.Join(dir, "app.js"), "console.log('hi')")

	handler := spaFileServer(dir)

	tests := []struct {
		name       string
		path       string
		wantBody   string
		wantStatus int
	}{
		{"root serves index", "/", "<html>app shell</html>", http.StatusOK},
		{"real file served directly", "/app.js", "console.log('hi')", http.StatusOK},
		{"unknown path falls back to index (SPA routing)", "/some/client/route", "<html>app shell</html>", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestNewRouter_WithoutStaticDir_DoesNotServeRoot(t *testing.T) {
	cache := &fakeLiveCache{quotes: map[string]domain.Quote{}}
	store := &fakeHistoryStore{points: map[string][]domain.HistoryPoint{}}
	query := app.NewQueryService(cache, store, nil)
	router := NewRouter(query, "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (no static handler should be mounted)", rec.Code, http.StatusNotFound)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
