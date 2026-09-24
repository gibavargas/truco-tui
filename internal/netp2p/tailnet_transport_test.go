package netp2p

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeTailnetCoordinator struct{}

func (fakeTailnetCoordinator) CreateTailnetSession(context.Context, TailnetCreateSessionRequest) (TailnetCreateSessionResponse, error) {
	return TailnetCreateSessionResponse{}, nil
}

func (fakeTailnetCoordinator) JoinTailnetSession(context.Context, TailnetJoinSessionRequest) (TailnetJoinSessionResponse, error) {
	return TailnetJoinSessionResponse{}, nil
}

func (fakeTailnetCoordinator) MintTailnetJoinTicket(context.Context, TailnetMintJoinTicketRequest) (TailnetMintJoinTicketResponse, error) {
	return TailnetMintJoinTicketResponse{}, nil
}

func (fakeTailnetCoordinator) PromoteTailnetAuthority(context.Context, TailnetPromoteAuthorityRequest) (TailnetPromoteAuthorityResponse, error) {
	return TailnetPromoteAuthorityResponse{}, nil
}

func TestHostConfigAutoSelectsTailnetWhenCoordinatorExists(t *testing.T) {
	cfg := HostConfig{
		TransportMode:      TransportAuto,
		TailnetCoordinator: fakeTailnetCoordinator{},
	}.normalized()

	if cfg.RequestedTransportMode != TransportAuto {
		t.Fatalf("RequestedTransportMode = %q, want %q", cfg.RequestedTransportMode, TransportAuto)
	}
	if cfg.TransportMode != TransportTailnetTSNetV1 {
		t.Fatalf("TransportMode = %q, want %q", cfg.TransportMode, TransportTailnetTSNetV1)
	}
}

func TestHostConfigAutoFallbackOrderWithoutTailnetCoordinator(t *testing.T) {
	withRelay := HostConfig{
		TransportMode: TransportAuto,
		RelayURL:      "https://relay.example",
	}.normalized()
	if withRelay.TransportMode != TransportRelayQUICV2 {
		t.Fatalf("relay TransportMode = %q, want %q", withRelay.TransportMode, TransportRelayQUICV2)
	}
	if withRelay.FallbackReason == "" {
		t.Fatal("relay fallback reason is empty")
	}

	direct := HostConfig{TransportMode: TransportAuto}.normalized()
	if direct.TransportMode != TransportTCPTLS {
		t.Fatalf("direct TransportMode = %q, want %q", direct.TransportMode, TransportTCPTLS)
	}
	if direct.FallbackReason == "" {
		t.Fatal("direct fallback reason is empty")
	}
}

func TestHTTPTailnetCoordinatorUsesExpectedRoutes(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/tailnet/sessions":
			writeJSON(t, w, TailnetCreateSessionResponse{
				SessionID:      "table-1",
				HostAdminToken: "admin",
				HostAuthKey:    "host-auth",
				HostNodeName:   "host-node",
				JoinTicket:     "join",
				ExpiresAt:      time.Now().UTC().Add(time.Minute),
			})
		case "/v1/tailnet/sessions/table-1/join":
			writeJSON(t, w, TailnetJoinSessionResponse{
				AuthKey:       "client-auth",
				NodeName:      "client-node",
				AuthorityNode: "host-node",
				ServicePort:   39001,
			})
		case "/v1/tailnet/sessions/table-1/join-tickets":
			writeJSON(t, w, TailnetMintJoinTicketResponse{JoinTicket: "replacement"})
		case "/v1/tailnet/sessions/table-1/authority":
			writeJSON(t, w, TailnetPromoteAuthorityResponse{
				HostAuthKey:   "promoted-auth",
				AuthorityNode: "client-node",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	coordinator := newHTTPTailnetCoordinator(srv.URL)
	ctx := context.Background()
	if _, err := coordinator.CreateTailnetSession(ctx, TailnetCreateSessionRequest{HostIdentity: "Host", NumPlayers: 2, ServicePort: 39001}); err != nil {
		t.Fatalf("CreateTailnetSession: %v", err)
	}
	if _, err := coordinator.JoinTailnetSession(ctx, TailnetJoinSessionRequest{SessionID: "table-1", JoinTicket: "join", PlayerName: "Ana"}); err != nil {
		t.Fatalf("JoinTailnetSession: %v", err)
	}
	if _, err := coordinator.MintTailnetJoinTicket(ctx, TailnetMintJoinTicketRequest{SessionID: "table-1", HostAdminToken: "admin", TargetSeat: 1}); err != nil {
		t.Fatalf("MintTailnetJoinTicket: %v", err)
	}
	if _, err := coordinator.PromoteTailnetAuthority(ctx, TailnetPromoteAuthorityRequest{SessionID: "table-1", HostAdminToken: "admin", HostIdentity: "Ana"}); err != nil {
		t.Fatalf("PromoteTailnetAuthority: %v", err)
	}

	want := []string{
		"POST /v1/tailnet/sessions",
		"POST /v1/tailnet/sessions/table-1/join",
		"POST /v1/tailnet/sessions/table-1/join-tickets",
		"POST /v1/tailnet/sessions/table-1/authority",
	}
	if len(seen) != len(want) {
		t.Fatalf("seen routes = %v, want %v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("route %d = %q, want %q", i, seen[i], want[i])
		}
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode response: %v", err)
	}
}
