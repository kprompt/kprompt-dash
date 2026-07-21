package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultListLimit = 200
	maxListLimit     = 500
)

// DNS-1123 subdomain (allows dots — node hostnames, some workload names).
var nameRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusBadGateway
	msg := err.Error()
	var statusErr *apierrors.StatusError
	if errors.As(err, &statusErr) {
		switch statusErr.Status().Code {
		case http.StatusNotFound:
			code = http.StatusNotFound
		case http.StatusForbidden:
			code = http.StatusForbidden
		case http.StatusUnauthorized:
			code = http.StatusUnauthorized
		case http.StatusTooManyRequests:
			code = http.StatusTooManyRequests
		case http.StatusBadRequest:
			code = http.StatusBadRequest
		}
		if statusErr.Status().Message != "" {
			msg = statusErr.Status().Message
		}
	}
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func requireName(w http.ResponseWriter, label, name string) bool {
	if name == "" || len(name) > 253 || !nameRE.MatchString(name) {
		writeBadRequest(w, "invalid "+label)
		return false
	}
	return true
}

func listLimit(r *http.Request) int64 {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return defaultListLimit
	}
	n, err := strconv.ParseInt(q, 10, 64)
	if err != nil || n < 1 {
		return defaultListLimit
	}
	if n > maxListLimit {
		return maxListLimit
	}
	return n
}

func listOpts(r *http.Request) metav1.ListOptions {
	return metav1.ListOptions{Limit: listLimit(r)}
}

func truncated(continueToken string, n int, limit int64) bool {
	return continueToken != "" || int64(n) >= limit
}
