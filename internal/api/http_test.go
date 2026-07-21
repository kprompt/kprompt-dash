package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"default", true},
		{"kube-system", true},
		{"ip-10-0-1-2.ec2.internal", true},
		{"", false},
		{"UPPER", false},
		{"bad_name", false},
		{"has space", false},
	}
	for _, tc := range cases {
		rr := httptest.NewRecorder()
		got := requireName(rr, "namespace", tc.name)
		if got != tc.ok {
			t.Fatalf("%q: got=%v want=%v body=%s", tc.name, got, tc.ok, rr.Body.String())
		}
		if !tc.ok && rr.Code != http.StatusBadRequest {
			t.Fatalf("%q status=%d", tc.name, rr.Code)
		}
	}
}

func TestListLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?limit=50", nil)
	if listLimit(req) != 50 {
		t.Fatal("50")
	}
	req = httptest.NewRequest(http.MethodGet, "/?limit=9999", nil)
	if listLimit(req) != maxListLimit {
		t.Fatal("cap")
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	if listLimit(req) != defaultListLimit {
		t.Fatal("default")
	}
}

func TestTruncated(t *testing.T) {
	if !truncated("token", 10, 200) {
		t.Fatal("continue")
	}
	if !truncated("", 200, 200) {
		t.Fatal("at limit")
	}
	if truncated("", 10, 200) {
		t.Fatal("short")
	}
}
