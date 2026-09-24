package clientparity

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEbitengineSourceExposesPlayableParityControls(t *testing.T) {
	source := readEbitengineSource(t)
	controls := []string{
		"Partida Offline",
		"Criar Host",
		"Entrar Online",
		"Idioma: PT-BR",
		"Transport: Auto",
		"Transport: TCP + TLS",
		"Transport: Tailnet",
		"Transport: Relay QUIC v2",
		"TRUCO",
		"SEIS",
		"NOVE",
		"DOZE",
		"Votar Host",
		"Convite",
		"new-hand",
		"send",
	}

	for _, control := range controls {
		if !strings.Contains(source, control) {
			t.Fatalf("desktop/ebitengine/main.go is missing UI control %q", control)
		}
	}
}

func TestEbitengineSourceDispatchesSharedRuntimeIntents(t *testing.T) {
	source := readEbitengineSource(t)
	intents := []string{
		"IntentNewOfflineGame",
		"IntentCreateHostSession",
		"IntentJoinSession",
		"IntentStartHostedMatch",
		"IntentVoteHost",
		"IntentRequestReplacementInvite",
		"IntentGameAction",
		"IntentNewHand",
		"IntentSendChat",
		"IntentCloseSession",
	}

	for _, intent := range intents {
		if !strings.Contains(source, intent) {
			t.Fatalf("desktop/ebitengine/main.go is missing runtime intent %q", intent)
		}
	}
	for _, action := range []string{`Action:    "play"`, `Action: "truco"`, `Action: "accept"`, `Action: "refuse"`} {
		if !strings.Contains(source, action) {
			t.Fatalf("desktop/ebitengine/main.go is missing game action dispatch %s", action)
		}
	}
	if !strings.Contains(source, "TransportMode: g.setupTransport") {
		t.Fatal("host setup must pass selected transport mode into the shared runtime")
	}
}

func readEbitengineSource(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "desktop", "ebitengine", "main.go")
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, path, nil, 0); err != nil {
		t.Fatalf("desktop/ebitengine/main.go must remain parseable: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read desktop/ebitengine/main.go: %v", err)
	}
	return string(b)
}
