package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// IDParam reads a path value (Go 1.22 ServeMux) as int64.
func IDParam(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(r.PathValue(key), 10, 64)
}

func QueryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// QueryFloat: second bool is true when the param is present & valid.
func QueryFloat(r *http.Request, key string, def float64) (float64, bool) {
	if v := r.URL.Query().Get(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return def, false
}

// QueryStrings supports both ?tags=a&tags=b and ?tags=a,b
func QueryStrings(r *http.Request, key string) []string {
	raw := r.URL.Query()[key]
	out := []string{}
	for _, v := range raw {
		for _, part := range strings.Split(v, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}
