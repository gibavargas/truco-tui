package appcore

import (
	"reflect"
	"testing"
)

func TestContractEnumeratesStableRuntimeSurface(t *testing.T) {
	wantIntents := []string{
		IntentSetLocale,
		IntentNewOfflineGame,
		IntentCreateHostSession,
		IntentJoinSession,
		IntentStartHostedMatch,
		IntentGameAction,
		IntentSendChat,
		IntentVoteHost,
		IntentRequestReplacementInvite,
		IntentCloseSession,
	}
	wantEvents := []string{
		EventChat,
		EventClientJoined,
		EventError,
		EventFailoverPromoted,
		EventFailoverRejoined,
		EventHostCreated,
		EventLobbyUpdated,
		EventLocaleChanged,
		EventMatchStarted,
		EventMatchUpdated,
		EventReplacementInvite,
		EventSessionClosed,
		EventSessionReady,
		EventSystem,
	}
	wantModes := []string{
		ModeIdle,
		ModeHostLobby,
		ModeClientLobby,
		ModeOfflineMatch,
		ModeHostMatch,
		ModeClientMatch,
	}
	wantLocales := []string{LocalePTBR, LocaleENUS}
	wantRoles := []string{DesiredRoleAuto, DesiredRolePartner, DesiredRoleOpponent}

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
	if SupportedIntentKinds()[0] != IntentSetLocale {
		t.Fatal("SupportedIntentKinds should return a defensive copy")
	}
}
