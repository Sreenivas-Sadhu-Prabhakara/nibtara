package backend

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type memStore struct{ items []Record }

func (m *memStore) Save(r Record) (Record, error) {
	r.ID = int64(len(m.items) + 1)
	m.items = append([]Record{r}, m.items...)
	return r, nil
}
func (m *memStore) List(limit int) ([]Record, error) { return m.items, nil }

func TestBucket(t *testing.T) {
	today := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	if b := (Due{SinceDate: "2026-08-10"}).Bucket(today); b != "current" {
		t.Fatalf("bucket=%s want current", b)
	}
	if b := (Due{SinceDate: "2026-06-10"}).Bucket(today); b != "aging" { // ~69 days
		t.Fatalf("bucket=%s want aging", b)
	}
	if b := (Due{SinceDate: "2026-01-10"}).Bucket(today); b != "stale" {
		t.Fatalf("bucket=%s want stale", b)
	}
}

func TestSettlementOptions(t *testing.T) {
	opts := SettlementOptions(1000)
	if len(opts) != 3 {
		t.Fatalf("want 3 options, got %d", len(opts))
	}
	if math.Abs(opts[1].Amount-900) > 1e-9 {
		t.Fatalf("waiver option=%v want 900", opts[1].Amount)
	}
	if math.Abs(opts[2].Amount-1000.0/3.0) > 1e-2 {
		t.Fatalf("instalment=%v", opts[2].Amount)
	}
}

func TestLogEndpoint(t *testing.T) {
	srv := NewServer(&memStore{})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/log",
		strings.NewReader(`{"customer":"Kumar","amount":1000,"sinceDate":"2026-05-01"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("log %d", rec.Code)
	}
}
