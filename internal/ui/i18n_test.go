package ui

import (
	"testing"

	"truco-tui/internal/netp2p"
)

func TestSetLocaleSwitchesTranslations(t *testing.T) {
	prev := localeCode()
	t.Cleanup(func() {
		_ = setLocale(prev)
	})

	if !setLocale("en-US") {
		t.Fatalf("expected setLocale to accept en-US")
	}
	if got := tr("menu_exit"); got != "Exit" {
		t.Fatalf("menu_exit in en-US = %q, want %q", got, "Exit")
	}

	if !setLocale("pt-BR") {
		t.Fatalf("expected setLocale to accept pt-BR")
	}
	if got := tr("menu_exit"); got != "Sair" {
		t.Fatalf("menu_exit in pt-BR = %q, want %q", got, "Sair")
	}
}

func TestSetLocaleRejectsUnknownCode(t *testing.T) {
	prev := localeCode()
	t.Cleanup(func() {
		_ = setLocale(prev)
	})
	if setLocale("xx-YY") {
		t.Fatalf("expected unknown locale to be rejected")
	}
}

func TestParseHostTransportChoice(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty defaults to auto", raw: "", want: netp2p.TransportAuto},
		{name: "direct option", raw: "1", want: netp2p.TransportTCPTLS},
		{name: "direct alias", raw: "direto", want: netp2p.TransportTCPTLS},
		{name: "tailnet option", raw: "2", want: netp2p.TransportTailnetTSNetV1},
		{name: "tailnet alias", raw: "tsnet", want: netp2p.TransportTailnetTSNetV1},
		{name: "relay option", raw: "3", want: netp2p.TransportRelayQUICV2},
		{name: "relay alias", raw: "relay_quic_v2", want: netp2p.TransportRelayQUICV2},
		{name: "unknown falls back to auto", raw: "banana", want: netp2p.TransportAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseHostTransportChoice(tt.raw); got != tt.want {
				t.Fatalf("parseHostTransportChoice(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
