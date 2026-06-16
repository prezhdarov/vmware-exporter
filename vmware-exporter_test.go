package main

import "testing"

func TestWebConfigUsesListenAddress(t *testing.T) {
	addr := ":9999"

	cfg := webConfig(&addr)

	if cfg == nil {
		t.Fatal("webConfig() returned nil")
	}

	if cfg.WebListenAddresses == nil {
		t.Fatal("WebListenAddresses is nil")
	}

	if len(*cfg.WebListenAddresses) != 1 {
		t.Fatalf("expected exactly 1 listen address, got %d", len(*cfg.WebListenAddresses))
	}

	if got := (*cfg.WebListenAddresses)[0]; got != addr {
		t.Fatalf("listen address = %q, want %q", got, addr)
	}

	if cfg.WebSystemdSocket == nil {
		t.Fatal("WebSystemdSocket is nil")
	}

	if *cfg.WebSystemdSocket {
		t.Fatal("WebSystemdSocket = true, want false")
	}

	if cfg.WebConfigFile == nil {
		t.Fatal("WebConfigFile is nil")
	}

	if *cfg.WebConfigFile != "" {
		t.Fatalf("WebConfigFile = %q, want empty string", *cfg.WebConfigFile)
	}
}