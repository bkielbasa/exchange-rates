package rates_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

func TestNew_DoesNotPanicWithNilDB(t *testing.T) {
	r := rates.New((*sql.DB)(nil), nil)
	if r == nil {
		t.Fatalf("New returned nil")
	}
}

func TestRegister_MuxHas404ForUnknownPath(t *testing.T) {
	mux := http.NewServeMux()
	r := rates.New((*sql.DB)(nil), nil)
	r.Register(mux)
	req := httptest.NewRequest(http.MethodGet, "/no/such/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}
