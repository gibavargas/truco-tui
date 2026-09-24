//go:build ebitengine_runtime_test

package main

import (
	"strings"
	"testing"

	"truco-tui/internal/appcore"
)

func newTestGame() *Game {
	g := &Game{
		runtime:          appcore.NewRuntime(),
		currentScreen:    "home",
		setupNumPlayers:  2,
		setupTransport:   "auto",
		setupDesiredRole: "auto",
		playedCardAnims:  make(map[string]*PlayedCardAnim),
	}
	g.inputName = TextBox{X: 250, Y: 200, W: 300, H: 30, Placeholder: "Nome do Jogador"}
	g.inputBindAddr = TextBox{X: 250, Y: 235, W: 300, H: 30, Placeholder: "0.0.0.0:0"}
	g.inputRelayURL = TextBox{X: 250, Y: 285, W: 300, H: 30, Placeholder: "Relay URL (Opcional)"}
	g.inputInviteKey = TextBox{X: 250, Y: 200, W: 300, H: 30, Placeholder: "Chave de Convite"}
	g.inputChat = TextBox{X: 530, Y: 330, W: 250, H: 30, Placeholder: "Mensagem chat..."}
	g.bundle = g.runtime.SnapshotBundle()
	return g
}

func TestEbitengineHomeExposesPrimaryFlows(t *testing.T) {
	g := newTestGame()
	defer func() { _ = g.runtime.Close() }()

	labels := buttonLabels(g.getButtons())
	for _, want := range []string{"Partida Offline", "Criar Host", "Entrar Online", "Idioma: PT-BR", "Sair"} {
		if !containsLabel(labels, want) {
			t.Fatalf("home labels = %v, missing %q", labels, want)
		}
	}
}

func TestEbitengineHostSetupCyclesAllTransportModes(t *testing.T) {
	g := newTestGame()
	defer func() { _ = g.runtime.Close() }()
	g.currentScreen = "setup_host"

	wantModes := []struct {
		label string
		mode  string
	}{
		{label: "Transport: Auto", mode: "auto"},
		{label: "Transport: TCP + TLS", mode: "tcp_tls"},
		{label: "Transport: Tailnet", mode: "tailnet_tsnet_v1"},
		{label: "Transport: Relay QUIC v2", mode: "relay_quic_v2"},
	}
	for _, want := range wantModes {
		btn := findButton(g.getButtons(), want.label)
		if btn == nil {
			t.Fatalf("transport button %q missing from labels %v", want.label, buttonLabels(g.getButtons()))
		}
		if g.setupTransport != want.mode {
			t.Fatalf("setupTransport = %q, want %q before clicking %q", g.setupTransport, want.mode, want.label)
		}
		btn.OnClick()
	}
	if g.setupTransport != "auto" {
		t.Fatalf("setupTransport after full cycle = %q, want auto", g.setupTransport)
	}
}

func TestEbitengineOfflineSetupStartsPlayableMatchSurface(t *testing.T) {
	g := newTestGame()
	defer func() { _ = g.runtime.Close() }()
	g.currentScreen = "setup_offline"
	g.inputName.Text = "Jogador"

	start := findButton(g.getButtons(), "Iniciar")
	if start == nil {
		t.Fatalf("offline setup labels = %v, missing Iniciar", buttonLabels(g.getButtons()))
	}
	start.OnClick()
	g.bundle = g.runtime.SnapshotBundle()
	g.autoNavigateScreen()

	if g.bundle.Mode != appcore.ModeOfflineMatch {
		t.Fatalf("mode = %q, want %q", g.bundle.Mode, appcore.ModeOfflineMatch)
	}
	if g.currentScreen != "match" {
		t.Fatalf("currentScreen = %q, want match", g.currentScreen)
	}
	if g.bundle.Match == nil {
		t.Fatal("match snapshot is nil")
	}
	if len(g.bundle.Match.Players) != 2 {
		t.Fatalf("players = %d, want 2", len(g.bundle.Match.Players))
	}

	labels := buttonLabels(g.getButtons())
	for _, want := range []string{"TRUCO", "Aceitar", "Recusar", "Nova mão", "Enviar"} {
		if !containsLabel(labels, want) {
			t.Fatalf("match labels = %v, missing %q", labels, want)
		}
	}
}

func TestEbitenginePlayableCardClickDispatchesGameAction(t *testing.T) {
	g := newTestGame()
	defer func() { _ = g.runtime.Close() }()
	g.currentScreen = "setup_offline"
	g.inputName.Text = "Jogador"

	start := findButton(g.getButtons(), "Iniciar")
	if start == nil {
		t.Fatalf("offline setup labels = %v, missing Iniciar", buttonLabels(g.getButtons()))
	}
	start.OnClick()
	g.bundle = g.runtime.SnapshotBundle()
	g.autoNavigateScreen()

	actions := g.bundle.UI.Actions
	if !actions.CanPlayCard {
		t.Skipf("seeded offline opening state is not immediately playable; actions = %+v", actions)
	}
	localSeat := actions.LocalPlayerID
	before := len(g.bundle.Match.Players[localSeat].Hand)
	layouts := g.matchCardLayout(before)
	if len(layouts) == 0 {
		t.Fatal("expected card hit targets for local hand")
	}
	target := layouts[0]
	g.updateMatchInteractions(target.cardX+target.cardW/2, target.cardY+target.cardH/2, true)
	g.bundle = g.runtime.SnapshotBundle()

	after := len(g.bundle.Match.Players[localSeat].Hand)
	if after >= before {
		t.Fatalf("local hand size after card click = %d, want less than %d", after, before)
	}
	if !logsContain(g.bundle.Match.Logs, "Jogador jogou") {
		t.Fatalf("match logs did not record local play: %v", g.bundle.Match.Logs)
	}
}

func buttonLabels(buttons []*Button) []string {
	labels := make([]string, 0, len(buttons))
	for _, btn := range buttons {
		labels = append(labels, btn.Label)
	}
	return labels
}

func findButton(buttons []*Button, label string) *Button {
	for _, btn := range buttons {
		if btn.Label == label {
			return btn
		}
	}
	return nil
}

func containsLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}

func logsContain(logs []string, want string) bool {
	for _, log := range logs {
		if strings.Contains(log, want) {
			return true
		}
	}
	return false
}
