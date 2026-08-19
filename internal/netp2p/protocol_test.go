package netp2p

import "testing"

func TestDecodeInviteKeyAcceptsLegacyTCPTransportVersion(t *testing.T) {
	key, err := EncodeInviteKey(InviteKey{
		Addr:             "127.0.0.1:1234",
		Token:            "token",
		Fingerprint:      "abcd",
		Transport:        "tcp_tls",
		TransportVersion: 1,
	})
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	if inv.TransportVersion != 1 {
		t.Fatalf("TransportVersion = %d, want 1", inv.TransportVersion)
	}
	if inv.Transport != "tcp_tls" {
		t.Fatalf("Transport = %q, want tcp_tls", inv.Transport)
	}
}

func TestDecodeInviteKeyFallsBackToRelaySessionToken(t *testing.T) {
	key, err := EncodeInviteKey(InviteKey{
		Token:             "token",
		Fingerprint:       "abcd",
		Transport:         "relay_quic",
		TransportVersion:  1,
		RelayURL:          "https://relay.example",
		RelaySessionID:    "session",
		RelaySessionToken: "join-ticket",
	})
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	if inv.Transport != "relay_quic_v2" {
		t.Fatalf("Transport = %q, want relay_quic_v2", inv.Transport)
	}
	if inv.RelayJoinTicket != "join-ticket" {
		t.Fatalf("RelayJoinTicket = %q, want join-ticket", inv.RelayJoinTicket)
	}
}

func TestProtocolVersionCandidatesPreferNegotiatedVersion(t *testing.T) {
	got := protocolVersionCandidates(1)
	if len(got) < 2 {
		t.Fatalf("protocolVersionCandidates(1) = %v, want fallback candidates", got)
	}
	if got[0] != 1 || got[1] != 2 {
		t.Fatalf("protocolVersionCandidates(1) = %v, want [1 2]", got)
	}
}
