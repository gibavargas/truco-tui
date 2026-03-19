package appcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"truco-tui/internal/netp2p"
	"truco-tui/internal/truco"
	"truco-tui/internal/ui"
)

type Runtime struct {
	mu sync.Mutex

	mode       string
	host       *netp2p.HostSession
	client     *netp2p.ClientSession
	game       *truco.Game
	localSeat  int
	inviteKey  string
	seedLo     uint64
	seedHi     uint64
	useSeed    bool
	closed     bool
	sessionGen uint64

	lastError *AppError
	nextSeq   int64
	events    []AppEvent
	eventLog  []string

	match *truco.Snapshot
	lobby *LobbySnapshot
}

func NewRuntime() *Runtime {
	return &Runtime{
		mode:      ModeIdle,
		localSeat: -1,
	}
}

func (r *Runtime) Versions() CoreVersions {
	return CoreVersions{
		CoreAPIVersion:  CoreAPIVersion,
		ProtocolVersion: netp2p.ProtocolVersion,
		SnapshotSchema:  SnapshotSchemaMajor,
	}
}

func (r *Runtime) DispatchIntent(intent AppIntent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("runtime encerrado")
	}
	r.recordIntentLocked(intent)

	switch intent.Kind {
	case IntentSetLocale:
		var payload SetLocalePayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		if !ui.SetLocale(payload.Locale) {
			return r.failLocked("invalid_locale", fmt.Errorf("locale inválido: %s", payload.Locale))
		}
		r.queueEventLocked(EventLocaleChanged, map[string]any{"locale": ui.LocaleCode()})
		return nil

	case IntentNewOfflineGame:
		var payload NewOfflineGamePayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.startOfflineLocked(payload)

	case IntentCreateHostSession:
		var payload CreateHostPayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.createHostLocked(payload)

	case IntentStartHostedMatch:
		return r.startHostedMatchLocked()

	case IntentJoinSession:
		var payload JoinSessionPayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.joinSessionLocked(payload)

	case IntentGameAction:
		var payload GameActionPayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.applyGameActionLocked(payload)

	case IntentSendChat:
		var payload SendChatPayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.sendChatLocked(payload.Text)

	case IntentVoteHost:
		var payload HostVotePayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.voteHostLocked(payload.CandidateSeat)

	case IntentRequestReplacementInvite:
		var payload ReplacementInvitePayload
		if err := decodeIntentPayload(intent.Payload, &payload); err != nil {
			return r.failLocked("invalid_intent", err)
		}
		return r.requestReplacementInviteLocked(payload.TargetSeat)

	case IntentCloseSession:
		r.teardownSessionLocked()
		r.mode = ModeIdle
		r.match = nil
		r.lobby = nil
		r.queueEventLocked(EventSessionClosed, nil)
		return nil
	}

	return r.failLocked("unknown_intent", fmt.Errorf("intent desconhecido: %s", intent.Kind))
}

func (r *Runtime) PollEvent() (AppEvent, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.events) == 0 {
		return AppEvent{}, false
	}
	ev := r.events[0]
	r.events = r.events[1:]
	return ev, true
}

func (r *Runtime) SnapshotBundle() SnapshotBundle {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshotBundleLocked()
}

func (r *Runtime) SnapshotJSON() (string, error) {
	b, err := json.Marshal(r.SnapshotBundle())
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	r.teardownSessionLocked()
	return nil
}

func (r *Runtime) snapshotBundleLocked() SnapshotBundle {
	var match *truco.Snapshot
	if r.match != nil {
		s := cloneMatchSnapshot(*r.match)
		match = &s
	}
	var lobby *LobbySnapshot
	if r.lobby != nil {
		c := cloneLobbySnapshot(*r.lobby)
		lobby = &c
	}
	return SnapshotBundle{
		Versions: r.Versions(),
		Mode:     r.mode,
		Locale:   ui.LocaleCode(),
		Match:    match,
		Lobby:    lobby,
		UI:       r.uiStateLocked(match, lobby),
		Connection: ConnectionSnapshot{
			Status:       r.mode,
			IsOnline:     strings.Contains(r.mode, "host") || strings.Contains(r.mode, "client"),
			IsHost:       strings.Contains(r.mode, "host"),
			LastError:    cloneError(r.lastError),
			LastEventSeq: r.nextSeq,
		},
		Diagnostics: DiagnosticsSnapshot{
			EventBacklog: len(r.events),
			ReplaySeedLo: r.seedLo,
			ReplaySeedHi: r.seedHi,
			EventLog:     append([]string(nil), r.eventLog...),
		},
	}
}

func (r *Runtime) recordIntentLocked(intent AppIntent) {
	raw := intent.Kind
	if len(intent.Payload) > 0 {
		raw += ":" + string(intent.Payload)
	}
	r.eventLog = append(r.eventLog, raw)
	if len(r.eventLog) > 256 {
		r.eventLog = r.eventLog[len(r.eventLog)-256:]
	}
}

func (r *Runtime) queueEventLocked(kind string, payload any) {
	r.nextSeq++
	r.events = append(r.events, AppEvent{
		Kind:      kind,
		Sequence:  r.nextSeq,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:   payload,
	})
	if len(r.events) > 256 {
		r.events = r.events[len(r.events)-256:]
	}
}

func (r *Runtime) failLocked(code string, err error) error {
	r.lastError = &AppError{Code: code, Message: err.Error()}
	r.queueEventLocked(EventError, r.lastError)
	return err
}

func (r *Runtime) clearErrorLocked() {
	r.lastError = nil
}

func (r *Runtime) teardownSessionLocked() {
	r.sessionGen++
	if r.host != nil {
		_ = r.host.Close()
	}
	if r.client != nil {
		_ = r.client.Close()
	}
	r.host = nil
	r.client = nil
	r.game = nil
	r.localSeat = -1
	r.inviteKey = ""
}

func (r *Runtime) startOfflineLocked(payload NewOfflineGamePayload) error {
	r.teardownSessionLocked()
	r.clearErrorLocked()

	var (
		game *truco.Game
		err  error
	)
	if payload.SeedLo != 0 || payload.SeedHi != 0 {
		r.seedLo, r.seedHi, r.useSeed = payload.SeedLo, payload.SeedHi, true
		game, err = truco.NewGameWithSeed(payload.PlayerNames, payload.CPUFlags, payload.SeedLo, payload.SeedHi)
	} else {
		r.seedLo, r.seedHi, r.useSeed = 0, 0, false
		game, err = truco.NewGame(payload.PlayerNames, payload.CPUFlags)
	}
	if err != nil {
		return r.failLocked("create_offline_failed", err)
	}
	r.game = game
	r.mode = ModeOfflineMatch
	r.localSeat = 0
	snap := game.Snapshot(0)
	r.setMatchLocked(snap)
	r.lobby = nil
	gen := r.sessionGen
	go r.runGameTicker(gen)
	r.queueEventLocked(EventSessionReady, map[string]any{"mode": r.mode})
	return nil
}

func (r *Runtime) createHostLocked(payload CreateHostPayload) error {
	r.teardownSessionLocked()
	r.clearErrorLocked()
	if payload.BindAddr == "" {
		payload.BindAddr = "0.0.0.0:0"
	}
	host, key, err := netp2p.NewHostSessionWithConfig(payload.BindAddr, payload.HostName, payload.NumPlayers, netp2p.HostConfig{
		RelayURL:      payload.RelayURL,
		TransportMode: payload.TransportMode,
	})
	if err != nil {
		return r.failLocked("create_host_failed", err)
	}
	r.host = host
	r.mode = ModeHostLobby
	r.localSeat = 0
	r.inviteKey = key
	r.updateHostLobbyLocked()
	r.queueEventLocked(EventHostCreated, map[string]any{"invite_key": key})
	gen := r.sessionGen
	go r.runHostEventLoop(gen, host)
	return nil
}

func (r *Runtime) startHostedMatchLocked() error {
	if r.host == nil {
		return r.failLocked("invalid_state", errors.New("host ausente"))
	}
	if err := r.host.StartGame(); err != nil {
		return r.failLocked("start_hosted_match_failed", err)
	}
	game, err := truco.NewGame(r.host.Slots(), make([]bool, len(r.host.Slots())))
	if err != nil {
		return r.failLocked("create_host_game_failed", err)
	}
	r.game = game
	r.mode = ModeHostMatch
	snap := game.Snapshot(0)
	r.setMatchLocked(snap)
	pushSnapshotsToClients(r.host, game)
	r.updateHostLobbyLocked()
	gen := r.sessionGen
	go r.runHostActionLoop(gen, r.host)
	go r.runGameTicker(gen)
	r.queueEventLocked(EventMatchStarted, map[string]any{"mode": r.mode})
	return nil
}

func (r *Runtime) joinSessionLocked(payload JoinSessionPayload) error {
	r.teardownSessionLocked()
	r.clearErrorLocked()
	client, err := netp2p.JoinSession(payload.Key, payload.PlayerName, payload.DesiredRole)
	if err != nil {
		return r.failLocked("join_session_failed", err)
	}
	r.client = client
	r.mode = ModeClientLobby
	r.localSeat = client.AssignedSeat()
	r.lobby = &LobbySnapshot{
		Slots:        client.Slots(),
		AssignedSeat: client.AssignedSeat(),
		NumPlayers:   len(client.Slots()),
		Started:      client.GameStarted(),
		Role:         payload.DesiredRole,
	}
	gen := r.sessionGen
	go r.runClientLoop(gen, client)
	r.queueEventLocked(EventClientJoined, map[string]any{"seat": r.localSeat})
	return nil
}

func (r *Runtime) applyGameActionLocked(payload GameActionPayload) error {
	r.clearErrorLocked()
	switch r.mode {
	case ModeOfflineMatch:
		if err := applyGameActionHost(r.game, r.localSeat, payload); err != nil {
			return r.failLocked("game_action_failed", err)
		}
		r.setMatchLocked(r.game.Snapshot(r.localSeat))
		return nil
	case ModeHostMatch:
		if err := applyGameActionHost(r.game, r.localSeat, payload); err != nil {
			return r.failLocked("game_action_failed", err)
		}
		pushSnapshotsToClients(r.host, r.game)
		r.setMatchLocked(r.game.Snapshot(r.localSeat))
		return nil
	case ModeClientMatch:
		if r.client == nil {
			return r.failLocked("invalid_state", errors.New("cliente ausente"))
		}
		if err := r.client.SendGameAction(payload.Action, payload.CardIndex); err != nil {
			return r.failLocked("game_action_failed", err)
		}
		return nil
	default:
		return r.failLocked("invalid_state", fmt.Errorf("ação de jogo indisponível no modo %s", r.mode))
	}
}

func (r *Runtime) sendChatLocked(text string) error {
	r.clearErrorLocked()
	switch r.mode {
	case ModeOfflineMatch:
		r.queueEventLocked(EventChat, map[string]any{"author": "local", "text": text})
		return nil
	case ModeHostLobby, ModeHostMatch:
		r.host.SendHostChat(text)
		return nil
	case ModeClientLobby, ModeClientMatch:
		if err := r.client.SendChat(text); err != nil {
			return r.failLocked("send_chat_failed", err)
		}
		return nil
	default:
		return r.failLocked("invalid_state", fmt.Errorf("chat indisponível no modo %s", r.mode))
	}
}

func (r *Runtime) voteHostLocked(candidateSeat int) error {
	switch r.mode {
	case ModeHostLobby, ModeHostMatch:
		if err := r.host.CastHostVote(0, candidateSeat); err != nil {
			return r.failLocked("vote_host_failed", err)
		}
		r.updateHostLobbyLocked()
		return nil
	case ModeClientLobby, ModeClientMatch:
		if err := r.client.SendHostVote(candidateSeat); err != nil {
			return r.failLocked("vote_host_failed", err)
		}
		return nil
	default:
		return r.failLocked("invalid_state", fmt.Errorf("voto de host indisponível no modo %s", r.mode))
	}
}

func (r *Runtime) requestReplacementInviteLocked(targetSeat int) error {
	switch r.mode {
	case ModeHostLobby, ModeHostMatch:
		key, err := r.host.RequestReplacementInvite(0, targetSeat)
		if err != nil {
			return r.failLocked("replacement_invite_failed", err)
		}
		r.queueEventLocked(EventReplacementInvite, map[string]any{"target_seat": targetSeat, "invite_key": key})
		return nil
	case ModeClientLobby, ModeClientMatch:
		if err := r.client.RequestReplacementInvite(targetSeat); err != nil {
			return r.failLocked("replacement_invite_failed", err)
		}
		return nil
	default:
		return r.failLocked("invalid_state", fmt.Errorf("convite de substituição indisponível no modo %s", r.mode))
	}
}

func (r *Runtime) setMatchLocked(s truco.Snapshot) {
	r.match = &s
	r.queueEventLocked(EventMatchUpdated, s)
}

func (r *Runtime) updateHostLobbyLocked() {
	if r.host == nil {
		r.lobby = nil
		return
	}
	r.lobby = &LobbySnapshot{
		InviteKey:      r.inviteKey,
		Slots:          r.host.Slots(),
		AssignedSeat:   0,
		NumPlayers:     len(r.host.Slots()),
		Started:        r.host.GameStarted(),
		HostSeat:       r.host.CurrentHostSeat(),
		ConnectedSeats: r.host.ConnectedSeats(),
	}
	r.queueEventLocked(EventLobbyUpdated, r.lobby)
}

func (r *Runtime) updateClientLobbyLocked() {
	if r.client == nil {
		r.lobby = nil
		return
	}
	role := ""
	if r.lobby != nil {
		role = r.lobby.Role
	}
	r.lobby = &LobbySnapshot{
		Slots:          r.client.Slots(),
		AssignedSeat:   r.client.AssignedSeat(),
		NumPlayers:     len(r.client.Slots()),
		Started:        r.client.GameStarted(),
		HostSeat:       r.client.CurrentHostSeat(),
		ConnectedSeats: r.client.ConnectedSeats(),
		Role:           role,
	}
	r.queueEventLocked(EventLobbyUpdated, r.lobby)
}

func (r *Runtime) runHostEventLoop(gen uint64, host *netp2p.HostSession) {
	for ev := range host.Events() {
		r.mu.Lock()
		if r.closed || gen != r.sessionGen || r.host != host {
			r.mu.Unlock()
			return
		}
		r.updateHostLobbyLocked()
		r.queueEventLocked(EventSystem, map[string]any{"text": ev})
		r.mu.Unlock()
	}
}

func (r *Runtime) runHostActionLoop(gen uint64, host *netp2p.HostSession) {
	for action := range host.Actions() {
		r.mu.Lock()
		if r.closed || gen != r.sessionGen || r.host != host || r.game == nil {
			r.mu.Unlock()
			return
		}
		if err := applyRemoteAction(r.game, action); err != nil {
			r.lastError = &AppError{Code: "remote_action_invalid", Message: err.Error()}
			r.queueEventLocked(EventError, r.lastError)
			host.SendSystemToSeat(action.Seat, err.Error())
			r.mu.Unlock()
			continue
		}
		pushSnapshotsToClients(r.host, r.game)
		r.setMatchLocked(r.game.Snapshot(r.localSeat))
		r.mu.Unlock()
	}
}

func (r *Runtime) runClientLoop(gen uint64, client *netp2p.ClientSession) {
	go func() {
		for state := range client.StateUpdates() {
			r.mu.Lock()
			if r.closed || gen != r.sessionGen || r.client != client {
				r.mu.Unlock()
				return
			}
			r.mode = ModeClientMatch
			r.localSeat = client.AssignedSeat()
			r.setMatchLocked(state)
			r.updateClientLobbyLocked()
			r.mu.Unlock()
		}
	}()

	for ev := range client.Events() {
		if ev == netp2p.ClientEventHostLostFailover {
			r.handleClientFailover(gen, client)
			continue
		}
		r.mu.Lock()
		if r.closed || gen != r.sessionGen || r.client != client {
			r.mu.Unlock()
			return
		}
		r.updateClientLobbyLocked()
		r.queueEventLocked(EventSystem, map[string]any{"text": ev})
		r.mu.Unlock()
	}
}

func (r *Runtime) handleClientFailover(gen uint64, client *netp2p.ClientSession) {
	fs := client.FailoverState()
	if !fs.Ready {
		r.mu.Lock()
		if !r.closed && gen == r.sessionGen && r.client == client {
			r.failLocked("failover_unavailable", errors.New("estado insuficiente para failover automático"))
		}
		r.mu.Unlock()
		return
	}
	targetSeat := selectFailoverSeat(fs)
	if targetSeat < 0 {
		r.mu.Lock()
		if !r.closed && gen == r.sessionGen && r.client == client {
			r.failLocked("failover_unavailable", errors.New("não foi possível eleger novo host"))
		}
		r.mu.Unlock()
		return
	}
	inv := fs.Invite
	hostAddr := strings.TrimSpace(fs.PeerHosts[targetSeat])
	if inv.Transport != "relay_quic_v2" {
		if hostAddr == "" {
			r.mu.Lock()
			if !r.closed && gen == r.sessionGen && r.client == client {
				r.failLocked("failover_unavailable", errors.New("endereço do novo host indisponível"))
			}
			r.mu.Unlock()
			return
		}
		addr := net.JoinHostPort(hostAddr, strconv.Itoa(fs.HandoffPort))
		inv.Addr = addr
	} else {
		inv.RelayAuthorityPeer = fs.RouteHint
	}

	if fs.AssignedSeat == targetSeat {
		rotatedSnap, err := netp2p.RotateFailoverSnapshot(*fs.FullState, targetSeat)
		if err != nil {
			r.mu.Lock()
			if !r.closed && gen == r.sessionGen && r.client == client {
				r.failLocked("failover_rotate_failed", err)
			}
			r.mu.Unlock()
			return
		}
		rotatedSlots := netp2p.RotateSeatSlice(fs.Slots, targetSeat)
		rotatedPeers := netp2p.RotateSeatMapString(fs.PeerHosts, targetSeat, fs.NumPlayers)
		rotatedSeatIDs := netp2p.RotateSeatMapString(fs.SeatSessionIDs, targetSeat, fs.NumPlayers)
		game, err := truco.NewGameFromSnapshot(rotatedSnap)
		if err != nil {
			r.mu.Lock()
			if !r.closed && gen == r.sessionGen && r.client == client {
				r.failLocked("failover_game_failed", err)
			}
			r.mu.Unlock()
			return
		}
		host, _, err := netp2p.NewRecoveredHostSession(
			net.JoinHostPort("0.0.0.0", strconv.Itoa(fs.HandoffPort)),
			rotatedSlots[0],
			fs.NumPlayers,
			netp2p.RecoveredHostState{
				Token:               inv.Token,
				TLSSeed:             fs.TLSSeed,
				RelayHostAdminToken: fs.RelayHostAdminToken,
				RelayHostPeerID:     fmt.Sprintf("seat-%d", targetSeat),
				RelayEpoch:          fs.Epoch + 1,
				Slots:               rotatedSlots,
				SeatSessionIDs:      rotatedSeatIDs,
				PeerHosts:           rotatedPeers,
				TableHostSeat:       0,
			},
			netp2p.HostConfig{
				HandoffPort:         fs.HandoffPort,
				AdvertiseHost:       hostAddr,
				RelayURL:            inv.RelayURL,
				RelaySPKIPin:        inv.RelaySPKIPin,
				TransportMode:       inv.Transport,
				RelaySessionID:      inv.RelaySessionID,
				RelayHostAdminToken: fs.RelayHostAdminToken,
			},
		)
		if err != nil {
			r.mu.Lock()
			if !r.closed && gen == r.sessionGen && r.client == client {
				r.failLocked("failover_host_failed", err)
			}
			r.mu.Unlock()
			return
		}

		r.mu.Lock()
		defer r.mu.Unlock()
		if r.closed || gen != r.sessionGen || r.client != client {
			_ = host.Close()
			return
		}
		_ = r.client.Close()
		r.client = nil
		r.host = host
		r.game = game
		r.mode = ModeHostMatch
		r.localSeat = 0
		r.inviteKey = ""
		r.setMatchLocked(game.Snapshot(0))
		r.updateHostLobbyLocked()
		pushSnapshotsToClients(host, game)
		r.queueEventLocked(EventFailoverPromoted, map[string]any{"host_seat": 0})
		newGen := r.sessionGen
		go r.runHostEventLoop(newGen, host)
		go r.runHostActionLoop(newGen, host)
		go r.runGameTicker(newGen)
		return
	}

	var lastErr error
	for attempt := 1; attempt <= 16; attempt++ {
		newClient, err := netp2p.RejoinSession(inv, fs.Name, fs.DesiredRole, fs.SessionID, 1)
		if err == nil {
			r.mu.Lock()
			if r.closed || gen != r.sessionGen || r.client != client {
				r.mu.Unlock()
				_ = newClient.Close()
				return
			}
			_ = r.client.Close()
			r.client = newClient
			r.mode = ModeClientLobby
			r.localSeat = newClient.AssignedSeat()
			r.updateClientLobbyLocked()
			r.queueEventLocked(EventFailoverRejoined, map[string]any{"seat": r.localSeat})
			r.mu.Unlock()
			go r.runClientLoop(gen, newClient)
			return
		}
		lastErr = err
		time.Sleep(time.Duration(minInt(attempt, 6)) * 300 * time.Millisecond)
	}

	r.mu.Lock()
	if !r.closed && gen == r.sessionGen && r.client == client {
		r.failLocked("failover_rejoin_failed", lastErr)
	}
	r.mu.Unlock()
}

func (r *Runtime) runGameTicker(gen uint64) {
	tk := time.NewTicker(850 * time.Millisecond)
	defer tk.Stop()
	for range tk.C {
		r.mu.Lock()
		if r.closed || gen != r.sessionGen || r.game == nil {
			r.mu.Unlock()
			return
		}
		if r.mode == ModeHostMatch {
			r.syncProvisionalCPUsLocked()
		}
		changed, err := r.stepCPUIfNeededLocked()
		if err != nil {
			r.failLocked("cpu_step_failed", err)
			r.mu.Unlock()
			continue
		}
		if changed && r.mode == ModeHostMatch && r.host != nil {
			pushSnapshotsToClients(r.host, r.game)
		}
		r.mu.Unlock()
	}
}

func (r *Runtime) syncProvisionalCPUsLocked() bool {
	if r.host == nil || r.game == nil {
		return false
	}
	connected := r.host.ConnectedSeats()
	changed := false
	snap := r.game.Snapshot(r.localSeat)
	for i := range snap.Players {
		if i == 0 {
			continue
		}
		playerID := snap.Players[i].ID
		_, provisional := r.game.PlayerCPU(playerID)
		seatOnline := connected[i]
		if !seatOnline && !provisional {
			if r.game.SetPlayerCPU(playerID, true, true) {
				changed = true
				r.queueEventLocked(EventSystem, map[string]any{"text": fmt.Sprintf(ui.Translate("online_provisional_cpu_on"), snap.Players[i].Name)})
			}
			continue
		}
		if seatOnline && provisional {
			if r.game.SetPlayerCPU(playerID, false, false) {
				changed = true
				r.queueEventLocked(EventSystem, map[string]any{"text": fmt.Sprintf(ui.Translate("online_provisional_cpu_off"), snap.Players[i].Name)})
			}
		}
	}
	if changed {
		r.setMatchLocked(r.game.Snapshot(r.localSeat))
	}
	return changed
}

func (r *Runtime) stepCPUIfNeededLocked() (bool, error) {
	if r.game == nil {
		return false, nil
	}
	snap := r.game.Snapshot(r.localSeat)
	if snap.MatchFinished {
		return false, nil
	}
	isCPU, pid := r.game.IsCPUTurn()
	if !isCPU {
		return false, nil
	}
	if r.mode == ModeHostMatch && pid == 0 {
		return false, nil
	}
	act := truco.DecideCPUAction(r.game, pid)
	if err := applyCPUActionToGame(r.game, pid, act); err != nil {
		return false, err
	}
	r.setMatchLocked(r.game.Snapshot(r.localSeat))
	return true, nil
}

func (r *Runtime) uiStateLocked(match *truco.Snapshot, lobby *LobbySnapshot) UIStateSnapshot {
	state := UIStateSnapshot{
		Actions: ActionSnapshot{
			LocalPlayerID:   -1,
			LocalTeam:       -1,
			CanCloseSession: r.mode != ModeIdle,
		},
	}
	if match != nil {
		localID := match.CurrentPlayerIdx
		localTeam := -1
		for _, p := range match.Players {
			if p.ID == localID {
				localTeam = p.Team
				break
			}
		}
		mustRespond := !match.MatchFinished && match.PendingRaiseFor == localTeam && localTeam >= 0
		state.Actions = ActionSnapshot{
			LocalPlayerID:   localID,
			LocalTeam:       localTeam,
			CanPlayCard:     !match.MatchFinished && match.TurnPlayer == localID && match.PendingRaiseFor == -1,
			CanAskOrRaise:   !match.MatchFinished && ((match.TurnPlayer == localID && match.PendingRaiseFor == -1 && match.CanAskTruco) || mustRespond),
			MustRespond:     mustRespond,
			CanAccept:       mustRespond,
			CanRefuse:       mustRespond,
			CanCloseSession: r.mode != ModeIdle,
		}
	}
	if lobby == nil {
		return state
	}
	state.LobbySlots = make([]LobbySlotState, 0, len(lobby.Slots))
	localIsTableHost := lobby.AssignedSeat == lobby.HostSeat
	for seat, name := range lobby.Slots {
		slot := LobbySlotState{
			Seat:        seat,
			Name:        name,
			IsEmpty:     strings.TrimSpace(name) == "",
			IsLocal:     lobby.AssignedSeat == seat,
			IsHost:      lobby.HostSeat == seat,
			IsConnected: lobby.ConnectedSeats[seat],
		}
		slot.IsOccupied = !slot.IsEmpty
		slot.IsProvisionalCPU = r.isSeatProvisionalCPULocked(seat)
		slot.CanVoteHost = slot.IsOccupied && !slot.IsLocal
		slot.CanRequestReplacement = (r.mode == ModeHostLobby || r.mode == ModeHostMatch || ((r.mode == ModeClientLobby || r.mode == ModeClientMatch) && localIsTableHost)) &&
			lobby.Started && slot.IsOccupied && !slot.IsLocal && !slot.IsConnected
		switch {
		case slot.IsEmpty:
			slot.Status = "empty"
		case slot.IsProvisionalCPU:
			slot.Status = "provisional_cpu"
		case slot.IsConnected:
			slot.Status = "occupied_online"
		default:
			slot.Status = "occupied_offline"
		}
		state.LobbySlots = append(state.LobbySlots, slot)
	}
	return state
}

func (r *Runtime) isSeatProvisionalCPULocked(seat int) bool {
	if r.game == nil || seat < 0 {
		return false
	}
	snap := r.game.Snapshot(r.localSeat)
	if seat >= len(snap.Players) {
		return false
	}
	_, provisional := r.game.PlayerCPU(snap.Players[seat].ID)
	return provisional
}

func decodeIntentPayload[T any](raw json.RawMessage, dest *T) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func applyGameActionHost(game *truco.Game, playerID int, payload GameActionPayload) error {
	if game == nil {
		return errors.New("partida ausente")
	}
	switch payload.Action {
	case "play":
		return game.PlayCard(playerID, payload.CardIndex)
	case "truco":
		return requestOrRaiseTruco(game, playerID)
	case "accept":
		return game.RespondTruco(playerID, true)
	case "refuse":
		return game.RespondTruco(playerID, false)
	default:
		return fmt.Errorf("ação desconhecida: %s", payload.Action)
	}
}

func requestOrRaiseTruco(g *truco.Game, playerID int) error {
	pendingTeam := g.PendingTeamToRespond()
	if pendingTeam != -1 && g.TeamOfPlayer(playerID) == pendingTeam {
		return g.RaiseTruco(playerID)
	}
	return g.AskTruco(playerID)
}

func applyCPUActionToGame(g *truco.Game, pid int, a truco.CPUAction) error {
	switch a.Kind {
	case "ask_truco":
		return g.AskTruco(pid)
	case "raise":
		return g.RaiseTruco(pid)
	case "accept":
		return g.RespondTruco(pid, true)
	case "refuse":
		return g.RespondTruco(pid, false)
	case "play":
		return g.PlayCard(pid, a.CardIndex)
	}
	return nil
}

func applyRemoteAction(game *truco.Game, a netp2p.ClientAction) error {
	switch a.Action {
	case "play":
		return game.PlayCard(a.Seat, a.CardIndex)
	case "truco":
		return requestOrRaiseTruco(game, a.Seat)
	case "accept":
		return game.RespondTruco(a.Seat, true)
	case "refuse":
		return game.RespondTruco(a.Seat, false)
	default:
		return fmt.Errorf("ação desconhecida: %s", a.Action)
	}
}

func pushSnapshotsToClients(host *netp2p.HostSession, game *truco.Game) {
	slots := host.Slots()
	full := game.Snapshot(0)
	full.Logs = nil
	full.CurrentPlayerIdx = -1
	for seat := 1; seat < len(slots); seat++ {
		s := maskedSnapshotForSeat(game.Snapshot(seat), seat)
		host.SendGameStateToSeat(seat, netp2p.Message{Type: "game_state", State: &s, FullState: &full})
	}
}

func maskedSnapshotForSeat(s truco.Snapshot, seat int) truco.Snapshot {
	out := s
	out.Players = append([]truco.Player(nil), s.Players...)
	for i := range out.Players {
		if i != seat {
			out.Players[i].Hand = nil
			continue
		}
		out.Players[i].Hand = append([]truco.Card(nil), s.Players[i].Hand...)
	}
	return out
}

func selectFailoverSeat(fs netp2p.ClientFailoverState) int {
	if fs.HostSeat > 0 &&
		fs.HostSeat < fs.NumPlayers &&
		fs.HostSeat < len(fs.Slots) &&
		strings.TrimSpace(fs.Slots[fs.HostSeat]) != "" &&
		strings.TrimSpace(fs.PeerHosts[fs.HostSeat]) != "" {
		return fs.HostSeat
	}
	for seat := 1; seat < fs.NumPlayers; seat++ {
		if seat >= len(fs.Slots) || strings.TrimSpace(fs.Slots[seat]) == "" {
			continue
		}
		if strings.TrimSpace(fs.PeerHosts[seat]) != "" {
			return seat
		}
	}
	return -1
}

func cloneMatchSnapshot(in truco.Snapshot) truco.Snapshot {
	out := in
	out.Players = append([]truco.Player(nil), in.Players...)
	for i := range out.Players {
		out.Players[i].Hand = append([]truco.Card(nil), in.Players[i].Hand...)
	}
	out.CurrentHand.RoundCards = append([]truco.PlayedCard(nil), in.CurrentHand.RoundCards...)
	out.CurrentHand.TrickResults = append([]int(nil), in.CurrentHand.TrickResults...)
	out.LastTrickCards = append([]truco.PlayedCard(nil), in.LastTrickCards...)
	out.TrickPiles = make([]truco.TrickPile, len(in.TrickPiles))
	for i, pile := range in.TrickPiles {
		out.TrickPiles[i] = pile
		out.TrickPiles[i].Cards = append([]truco.PlayedCard(nil), pile.Cards...)
	}
	if in.CurrentHand.TrickWins != nil {
		out.CurrentHand.TrickWins = make(map[int]int, len(in.CurrentHand.TrickWins))
		for k, v := range in.CurrentHand.TrickWins {
			out.CurrentHand.TrickWins[k] = v
		}
	}
	if in.MatchPoints != nil {
		out.MatchPoints = make(map[int]int, len(in.MatchPoints))
		for k, v := range in.MatchPoints {
			out.MatchPoints[k] = v
		}
	}
	out.Logs = append([]string(nil), in.Logs...)
	return out
}

func cloneLobbySnapshot(in LobbySnapshot) LobbySnapshot {
	out := in
	out.Slots = append([]string(nil), in.Slots...)
	if in.ConnectedSeats != nil {
		out.ConnectedSeats = make(map[int]bool, len(in.ConnectedSeats))
		for k, v := range in.ConnectedSeats {
			out.ConnectedSeats[k] = v
		}
	}
	if in.Metadata != nil {
		out.Metadata = make(map[string]any, len(in.Metadata))
		for k, v := range in.Metadata {
			out.Metadata[k] = v
		}
	}
	return out
}

func cloneError(in *AppError) *AppError {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
