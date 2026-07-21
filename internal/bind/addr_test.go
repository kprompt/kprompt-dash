package bind

import "testing"

func TestCheckLoopbackOK(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:7474", "localhost:7474", "[::1]:7474"} {
		if err := Check(addr, false); err != nil {
			t.Fatalf("%s: %v", addr, err)
		}
	}
}

func TestCheckBindAllRefused(t *testing.T) {
	if err := Check(":7474", false); err == nil {
		t.Fatal("expected refuse :7474")
	}
	if err := Check(":7474", true); err == nil {
		t.Fatal("expected refuse :7474 even with allow-remote")
	}
}

func TestCheckNonLoopbackRequiresFlag(t *testing.T) {
	if err := Check("0.0.0.0:7474", false); err == nil {
		t.Fatal("expected refuse without --allow-remote")
	}
	if err := Check("0.0.0.0:7474", true); err != nil {
		t.Fatalf("allow-remote should permit: %v", err)
	}
	if err := Check("192.168.1.10:7474", false); err == nil {
		t.Fatal("expected refuse LAN bind")
	}
}

func TestIsLoopback(t *testing.T) {
	if !IsLoopback("127.0.0.1:7474") {
		t.Fatal("127.0.0.1")
	}
	if IsLoopback("0.0.0.0:7474") {
		t.Fatal("0.0.0.0")
	}
}
