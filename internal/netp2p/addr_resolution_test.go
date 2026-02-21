package netp2p

import (
	"net"
	"testing"
)

func TestBuildInviteAddrAvoidsWildcardHost(t *testing.T) {
	got := buildInviteAddr("[::]:62225", "")
	host, port, err := net.SplitHostPort(got)
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", got, err)
	}
	if port != "62225" {
		t.Fatalf("port=%q, want 62225", port)
	}
	if host == "" || host == "::" || host == "0.0.0.0" {
		t.Fatalf("host=%q should not be wildcard", host)
	}
}

func TestBuildInviteAddrHonorsAdvertiseHost(t *testing.T) {
	got := buildInviteAddr("0.0.0.0:19001", "192.168.0.42")
	if got != "192.168.0.42:19001" {
		t.Fatalf("invite addr=%q, want %q", got, "192.168.0.42:19001")
	}
}

func TestInviteDialAddrsExpandWildcard(t *testing.T) {
	got := inviteDialAddrs("[::]:62225")
	need := map[string]bool{
		"[::]:62225":      true,
		"127.0.0.1:62225": true,
		"[::1]:62225":     true,
		"localhost:62225": true,
	}
	for _, addr := range got {
		delete(need, addr)
	}
	if len(need) > 0 {
		t.Fatalf("missing fallback addresses: %+v (got=%v)", need, got)
	}
}
