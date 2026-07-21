package api

import (
	"testing"
	"time"
)

func TestAge(t *testing.T) {
	if got := age(time.Time{}); got != "—" {
		t.Fatalf("zero=%q", got)
	}
	if got := age(time.Now().Add(-30 * time.Second)); got != "30s" {
		t.Fatalf("seconds=%q", got)
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hi", 10) != "hi" {
		t.Fatal("short")
	}
	got := truncate("abcdefghijklmnopqrstuvwxyz", 10)
	if got != "abcdefghij…" {
		t.Fatalf("got=%q", got)
	}
}
