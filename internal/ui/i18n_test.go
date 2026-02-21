package ui

import "testing"

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
