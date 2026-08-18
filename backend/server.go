package backend

import (
	"encoding/json"
	"io"
	"net/http"
)

// NewServer wires the journal routes: POST /log, GET /records, GET /summary.
// store may be nil (no DB configured), in which case writes/reads report 503.
func NewServer(store Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/log", logHandler(store))
	mux.HandleFunc("/records", recordsHandler(store))
	mux.HandleFunc("/summary", summaryHandler(store))
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func logHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeErr(w, http.StatusBadRequest, "could not read body")
			return
		}
		headline, label, err := parseEntry(raw)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if store == nil {
			writeErr(w, http.StatusServiceUnavailable, "no database configured")
			return
		}
		rec, err := store.Save(Record{Input: raw, Headline: headline, Label: label})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not save")
			return
		}
		writeJSON(w, http.StatusCreated, rec)
	}
}

func recordsHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeErr(w, http.StatusServiceUnavailable, "no database configured")
			return
		}
		items, err := store.List(200)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not read")
			return
		}
		writeJSON(w, http.StatusOK, items)
	}
}

func summaryHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeErr(w, http.StatusServiceUnavailable, "no database configured")
			return
		}
		items, err := store.List(1000)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not read")
			return
		}
		writeJSON(w, http.StatusOK, Summarize(items))
	}
}
