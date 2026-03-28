package appcore

import (
	"reflect"
	"testing"
)

func TestContractEnumeratesStableRuntimeSurface(t *testing.T) {
	wantIntents := []string{
		"set_locale",
		"new_offline_game",
		"create_host_session",
		"join_session",
		"start_hosted_match",
		"game_action",
		"send_chat",
		"vote_host",
		"request_replacement_invite",
		"close_session",
	}
	wantEvents := []string{
		"chat",
		"client_joined",
		"error",
		"failover_promoted",
		"failover_rejoined",
		"host_created",
		"lobby_updated",
		"locale_changed",
		"match_started",
		"match_updated",
		"replacement_invite",
		"session_closed",
		"session_ready",
		"system",
	}
	wantModes := []string{
		"idle",
		"host_lobby",
		"client_lobby",
		"offline_match",
		"host_match",
		"client_match",
	}
	wantLocales := []string{"pt-BR", "en-US"}
	wantRoles := []string{"auto", "partner", "opponent"}

	if got := SupportedIntentKinds(); !reflect.DeepEqual(got, wantIntents) {
		t.Fatalf("SupportedIntentKinds = %v, want %v", got, wantIntents)
	}
	if got := SupportedEventKinds(); !reflect.DeepEqual(got, wantEvents) {
		t.Fatalf("SupportedEventKinds = %v, want %v", got, wantEvents)
	}
	if got := SupportedModes(); !reflect.DeepEqual(got, wantModes) {
		t.Fatalf("SupportedModes = %v, want %v", got, wantModes)
	}
	if got := SupportedLocales(); !reflect.DeepEqual(got, wantLocales) {
		t.Fatalf("SupportedLocales = %v, want %v", got, wantLocales)
	}
	if got := SupportedDesiredRoles(); !reflect.DeepEqual(got, wantRoles) {
		t.Fatalf("SupportedDesiredRoles = %v, want %v", got, wantRoles)
	}
}

func TestSupportedContractSlicesAreDefensiveCopies(t *testing.T) {
	intents := SupportedIntentKinds()
	intents[0] = "mutated"
	if SupportedIntentKinds()[0] != "set_locale" {
		t.Fatal("SupportedIntentKinds should return a defensive copy")
	}
}
