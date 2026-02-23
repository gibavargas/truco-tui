package ui

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"truco-tui/internal/netp2p"
	"truco-tui/internal/truco"
)

type onlineMode int

const (
	onlineModeHost onlineMode = iota
	onlineModeClient
)

type onlineMatchModel struct {
	mode onlineMode

	host *netp2p.HostSession
	cli  *netp2p.ClientSession

	failoverRunning bool

	UIModel
}

type hostActionMsg struct {
	action netp2p.ClientAction
	ok     bool
}

type hostEventMsg struct {
	text string
	ok   bool
}

type clientStateMsg struct {
	snapshot truco.Snapshot
	ok       bool
}

type clientEventMsg struct {
	text string
	ok   bool
}

type hostCPUStepMsg struct {
	snapshot truco.Snapshot
	changed  bool
	err      error
}

type clientFailoverMsg struct {
	promoted bool
	host     *netp2p.HostSession
	game     *truco.Game
	cli      *netp2p.ClientSession
	snapshot truco.Snapshot
	note     string
	err      error
}

func newOnlineHostModel(host *netp2p.HostSession, game *truco.Game) onlineMatchModel {
	s := game.Snapshot(0)
	base := UIModel{
		game:           game,
		snapshot:       s,
		activeTab:      "mesa",
		chatLog:        []string{tr("chat_online_started")},
		localPlayerIdx: 0,
		isOnline:       true,
		isHost:         true,
		visualState:    newVisualState(s),
	}
	return onlineMatchModel{
		mode:    onlineModeHost,
		host:    host,
		UIModel: base,
	}
}

func newOnlineClientModel(cli *netp2p.ClientSession, initial truco.Snapshot) onlineMatchModel {
	seat := cli.AssignedSeat()
	if initial.CurrentPlayerIdx >= 0 {
		seat = initial.CurrentPlayerIdx
	}
	base := UIModel{
		snapshot:       initial,
		activeTab:      "mesa",
		chatLog:        []string{tr("chat_online_started")},
		localPlayerIdx: seat,
		isOnline:       true,
		isHost:         false,
		visualState:    newVisualState(initial),
	}
	return onlineMatchModel{
		mode:    onlineModeClient,
		cli:     cli,
		UIModel: base,
	}
}

func (m onlineMatchModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
		uiTickCmd(),
	}
	if m.mode == onlineModeHost {
		pushSnapshotsToClients(m.host, m.game)
		cmds = append(cmds, waitHostActionCmd(m.host), waitHostEventCmd(m.host), hostCPUStepCmd(m.game, m.host, m.localPlayerIdx))
	}
	if m.mode == onlineModeClient {
		cmds = append(cmds, waitClientStateCmd(m.cli), waitClientEventCmd(m.cli))
	}
	return tea.Batch(cmds...)
}

func (m onlineMatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case clearTrickOverlayMsg:
		return m.handleClearTrickOverlay(msg)

	case clearErrorMsg:
		return m.handleClearError(msg)

	case uiTickMsg:
		return m.handleUITick()

	case playAnimTickMsg:
		return m.handlePlayAnimTick(msg)

	case hostActionMsg:
		return m.handleHostAction(msg)

	case hostEventMsg:
		return m.handleHostEvent(msg)

	case hostCPUStepMsg:
		return m.handleHostCPUStep(msg)

	case clientStateMsg:
		return m.handleClientState(msg)

	case clientEventMsg:
		return m.handleClientEvent(msg)

	case clientFailoverMsg:
		return m.handleClientFailover(msg)

	case tea.KeyMsg:
		return m.handleOnlineKey(msg)
	}
	return m, nil
}

func (m onlineMatchModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

func (m onlineMatchModel) handleClearTrickOverlay(msg clearTrickOverlayMsg) (tea.Model, tea.Cmd) {
	m.visualState.onClearTrickOverlay(msg.id)
	return m, nil
}

func (m onlineMatchModel) handleClearError(msg clearErrorMsg) (tea.Model, tea.Cmd) {
	m.visualState.onClearError(msg.id, &m.err)
	return m, nil
}

func (m onlineMatchModel) handleUITick() (tea.Model, tea.Cmd) {
	m.visualState.onUITick()
	if m.mode == onlineModeHost && m.syncProvisionalCPUs() {
		return m, tea.Batch(
			m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
			m.visualState.updatePlayAnimation(m.snapshot),
		)
	}
	return m, uiTickCmd()
}

func (m onlineMatchModel) handlePlayAnimTick(msg playAnimTickMsg) (tea.Model, tea.Cmd) {
	if cmd := m.visualState.onPlayAnimTick(msg.id); cmd != nil {
		return m, cmd
	}
	return m, nil
}

func (m onlineMatchModel) handleHostAction(msg hostActionMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		return m, nil
	}
	prev := m.snapshot
	if err := applyRemoteAction(m.game, msg.action); err != nil {
		m.host.SendSystemToSeat(msg.action.Seat, tr("online_invalid_action_prefix")+err.Error())
	} else {
		pushSnapshotsToClients(m.host, m.game)
	}
	m.snapshot = m.game.Snapshot(m.localPlayerIdx)
	m.visualState.applySnapshotVisualTransitions(prev, m.snapshot)
	return m, tea.Batch(
		waitHostActionCmd(m.host),
		m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
		m.visualState.updatePlayAnimation(m.snapshot),
	)
}

func (m onlineMatchModel) handleHostEvent(msg hostEventMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		return m, nil
	}
	if strings.TrimSpace(msg.text) != "" {
		m.chatLog = append(m.chatLog, msg.text)
	}
	return m, waitHostEventCmd(m.host)
}

func (m onlineMatchModel) handleHostCPUStep(msg hostCPUStepMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m, tea.Batch(m.setTransientError(msg.err), hostCPUStepCmd(m.game, m.host, m.localPlayerIdx))
	}
	if msg.changed {
		prev := m.snapshot
		m.snapshot = msg.snapshot
		m.visualState.applySnapshotVisualTransitions(prev, m.snapshot)
		return m, tea.Batch(
			hostCPUStepCmd(m.game, m.host, m.localPlayerIdx),
			m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
			m.visualState.updatePlayAnimation(m.snapshot),
		)
	}
	return m, hostCPUStepCmd(m.game, m.host, m.localPlayerIdx)
}

func (m onlineMatchModel) handleClientState(msg clientStateMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		return m, nil
	}
	prev := m.snapshot
	m.snapshot = msg.snapshot
	if msg.snapshot.CurrentPlayerIdx >= 0 {
		m.localPlayerIdx = msg.snapshot.CurrentPlayerIdx
	}
	m.visualState.applySnapshotVisualTransitions(prev, m.snapshot)
	return m, tea.Batch(
		waitClientStateCmd(m.cli),
		m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
		m.visualState.updatePlayAnimation(m.snapshot),
	)
}

func (m onlineMatchModel) handleClientEvent(msg clientEventMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		return m, nil
	}
	if msg.text == netp2p.ClientEventHostLostFailover {
		if m.failoverRunning || m.cli == nil {
			return m, nil
		}
		m.failoverRunning = true
		return m, attemptClientFailoverCmd(m.cli)
	}
	if strings.TrimSpace(msg.text) != "" {
		m.chatLog = append(m.chatLog, msg.text)
	}
	return m, waitClientEventCmd(m.cli)
}

func (m onlineMatchModel) handleClientFailover(msg clientFailoverMsg) (tea.Model, tea.Cmd) {
	m.failoverRunning = false
	if msg.err != nil {
		return m, m.setTransientError(msg.err)
	}
	if strings.TrimSpace(msg.note) != "" {
		m.chatLog = append(m.chatLog, msg.note)
	}
	if msg.promoted {
		if m.cli != nil {
			_ = m.cli.Close()
		}
		m.mode = onlineModeHost
		m.host = msg.host
		m.cli = nil
		m.game = msg.game
		prev := m.snapshot
		m.snapshot = msg.snapshot
		m.localPlayerIdx = 0
		m.isOnline = true
		m.isHost = true
		m.visualState.applySnapshotVisualTransitions(prev, m.snapshot)
		return m, tea.Batch(
			waitHostActionCmd(m.host),
			waitHostEventCmd(m.host),
			hostCPUStepCmd(m.game, m.host, m.localPlayerIdx),
			m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
			m.visualState.updatePlayAnimation(m.snapshot),
		)
	}
	if msg.cli != nil {
		if m.cli != nil {
			_ = m.cli.Close()
		}
		m.mode = onlineModeClient
		m.cli = msg.cli
		m.isOnline = true
		m.isHost = false
		return m, tea.Batch(waitClientStateCmd(m.cli), waitClientEventCmd(m.cli))
	}
	return m, waitClientEventCmd(m.cli)
}

func (m onlineMatchModel) handleOnlineKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if quit, cmd := (&m).handleKey(msg); quit {
		return m, cmd
	}
	return m, nil
}

func (m *onlineMatchModel) handleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.mode == onlineModeHost && m.host != nil {
			m.host.SendSystem(tr("online_host_closed_match"))
		}
		return true, tea.Quit
	case "tab":
		if m.activeTab == "mesa" {
			m.activeTab = "chat"
		} else if m.activeTab == "chat" {
			m.activeTab = "log"
		} else {
			m.activeTab = "mesa"
		}
		if m.activeTab == "chat" {
			m.chatCursor = len([]rune(m.chatInput))
		}
		return false, nil
	}

	if submitted, handled := m.handleChatInputKey(msg); handled {
		if submitted != "" {
			if consumed, err := m.handleOnlineChatCommand(submitted); consumed {
				if err != nil {
					return false, m.setTransientError(err)
				}
				return false, nil
			}
			if m.mode == onlineModeHost && m.host != nil {
				m.host.SendHostChat(submitted)
			}
			if m.mode == onlineModeClient && m.cli != nil {
				_ = m.cli.SendChat(submitted)
			}
		}
		return false, nil
	}

	switch msg.String() {
	case "1", "2", "3", "t", "a", "r":
		if m.trickOverlayMsg != "" {
			return false, nil
		}
		if err := m.applyKeyAction(msg.String()); err != nil {
			return false, tea.Batch(
				m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
				m.visualState.updatePlayAnimation(m.snapshot),
				m.setTransientError(err),
			)
		}
		if msg.String() == "1" || msg.String() == "2" || msg.String() == "3" {
			idx := int(msg.String()[0] - '1')
			m.visualState.onCardAccepted(idx)
		}
		return false, tea.Batch(
			m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
			m.visualState.updatePlayAnimation(m.snapshot),
		)
	default:
		return false, nil
	}
}

func (m *onlineMatchModel) handleOnlineChatCommand(input string) (bool, error) {
	line := strings.TrimSpace(input)
	if !strings.HasPrefix(line, "/") {
		return false, nil
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return true, nil
	}
	switch strings.ToLower(parts[0]) {
	case "/host":
		if len(parts) != 2 {
			return true, errors.New(tr("online_cmd_use_host"))
		}
		slot, err := strconv.Atoi(parts[1])
		if err != nil || slot < 1 || slot > len(m.snapshot.Players) {
			return true, errors.New(tr("online_cmd_invalid_slot"))
		}
		candidate := slot - 1
		if m.mode == onlineModeHost {
			return true, m.host.CastHostVote(0, candidate)
		}
		return true, m.cli.SendHostVote(candidate)
	case "/invite":
		if len(parts) != 2 {
			return true, errors.New(tr("online_cmd_use_invite"))
		}
		slot, err := strconv.Atoi(parts[1])
		if err != nil || slot < 2 || slot > len(m.snapshot.Players) {
			return true, errors.New(tr("online_cmd_invalid_slot"))
		}
		target := slot - 1
		if m.mode == onlineModeHost {
			key, err := m.host.RequestReplacementInvite(0, target)
			if err != nil {
				return true, err
			}
			m.chatLog = append(m.chatLog, fmt.Sprintf("[host] convite de reposição slot %d: %s", slot, key))
			return true, nil
		}
		return true, m.cli.RequestReplacementInvite(target)
	default:
		return false, nil
	}
}

func (m *onlineMatchModel) syncProvisionalCPUs() bool {
	if m.mode != onlineModeHost || m.host == nil || m.game == nil {
		return false
	}
	connected := m.host.ConnectedSeats()
	changed := false
	for i := range m.snapshot.Players {
		if i == 0 {
			continue
		}
		playerID := m.snapshot.Players[i].ID
		_, provisional := m.game.PlayerCPU(playerID)
		seatOnline := connected[i]
		if !seatOnline && !provisional {
			if m.game.SetPlayerCPU(playerID, true, true) {
				changed = true
				m.chatLog = append(m.chatLog, fmt.Sprintf(tr("online_provisional_cpu_on"), m.snapshot.Players[i].Name))
			}
			continue
		}
		if seatOnline && provisional {
			if m.game.SetPlayerCPU(playerID, false, false) {
				changed = true
				m.chatLog = append(m.chatLog, fmt.Sprintf(tr("online_provisional_cpu_off"), m.snapshot.Players[i].Name))
			}
		}
	}
	if changed {
		m.snapshot = m.game.Snapshot(m.localPlayerIdx)
		pushSnapshotsToClients(m.host, m.game)
	}
	return changed
}

func hostCPUStepCmd(game *truco.Game, host *netp2p.HostSession, localSeat int) tea.Cmd {
	return tea.Tick(850*time.Millisecond, func(time.Time) tea.Msg {
		if game == nil || host == nil {
			return hostCPUStepMsg{}
		}
		snap := game.Snapshot(localSeat)
		if snap.MatchFinished {
			return hostCPUStepMsg{}
		}
		isCPU, pid := game.IsCPUTurn()
		if !isCPU || pid == 0 {
			return hostCPUStepMsg{}
		}
		act := truco.DecideCPUAction(game, pid)
		if err := applyCPUActionToGame(game, pid, act); err != nil {
			return hostCPUStepMsg{err: err}
		}
		pushSnapshotsToClients(host, game)
		return hostCPUStepMsg{
			snapshot: game.Snapshot(localSeat),
			changed:  true,
		}
	})
}

func (m *onlineMatchModel) applyKeyAction(key string) error {
	if m.snapshot.MatchFinished {
		return nil
	}

	if m.mode == onlineModeHost {
		switch key {
		case "1", "2", "3":
			idx := int(key[0] - '1')
			if err := m.game.PlayCard(m.localPlayerIdx, idx); err != nil {
				return err
			}
		case "t":
			if err := requestOrRaiseTruco(m.game, m.localPlayerIdx); err != nil {
				return err
			}
		case "a":
			if err := m.game.RespondTruco(m.localPlayerIdx, true); err != nil {
				return err
			}
		case "r":
			if err := m.game.RespondTruco(m.localPlayerIdx, false); err != nil {
				return err
			}
		}
		pushSnapshotsToClients(m.host, m.game)
		m.snapshot = m.game.Snapshot(m.localPlayerIdx)
		return nil
	}

	if m.mode == onlineModeClient {
		switch key {
		case "1", "2", "3":
			idx := int(key[0] - '1')
			return m.cli.SendGameAction("play", idx)
		case "t":
			return m.cli.SendGameAction("truco", 0)
		case "a":
			return m.cli.SendGameAction("accept", 0)
		case "r":
			return m.cli.SendGameAction("refuse", 0)
		}
	}
	return nil
}

func (m onlineMatchModel) View() string {
	return m.UIModel.View()
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

func attemptClientFailoverCmd(cli *netp2p.ClientSession) tea.Cmd {
	return func() tea.Msg {
		if cli == nil {
			return clientFailoverMsg{err: fmt.Errorf("sessão cliente ausente para failover")}
		}
		fs := cli.FailoverState()
		if !fs.Ready {
			return clientFailoverMsg{err: fmt.Errorf("estado insuficiente para failover automático")}
		}
		targetSeat := selectFailoverSeat(fs)
		if targetSeat < 0 {
			return clientFailoverMsg{err: fmt.Errorf("não foi possível eleger novo host")}
		}
		hostAddr := strings.TrimSpace(fs.PeerHosts[targetSeat])
		if hostAddr == "" {
			return clientFailoverMsg{err: fmt.Errorf("endereço do host eleito indisponível")}
		}
		addr := net.JoinHostPort(hostAddr, strconv.Itoa(fs.HandoffPort))
		inv := fs.Invite
		inv.Addr = addr

		if fs.AssignedSeat == targetSeat {
			rotatedSnap, err := netp2p.RotateFailoverSnapshot(*fs.FullState, targetSeat)
			if err != nil {
				return clientFailoverMsg{err: err}
			}
			rotatedSlots := netp2p.RotateSeatSlice(fs.Slots, targetSeat)
			rotatedPeers := netp2p.RotateSeatMapString(fs.PeerHosts, targetSeat, fs.NumPlayers)
			rotatedSeatIDs := netp2p.RotateSeatMapString(fs.SeatSessionIDs, targetSeat, fs.NumPlayers)

			game, err := truco.NewGameFromSnapshot(rotatedSnap)
			if err != nil {
				return clientFailoverMsg{err: err}
			}
			host, _, err := netp2p.NewRecoveredHostSession(
				net.JoinHostPort("0.0.0.0", strconv.Itoa(fs.HandoffPort)),
				rotatedSlots[0],
				fs.NumPlayers,
				netp2p.RecoveredHostState{
					Token:          inv.Token,
					Slots:          rotatedSlots,
					SeatSessionIDs: rotatedSeatIDs,
					PeerHosts:      rotatedPeers,
					TableHostSeat:  0,
				},
				netp2p.HostConfig{
					HandoffPort:   fs.HandoffPort,
					AdvertiseHost: hostAddr,
				},
			)
			if err != nil {
				return clientFailoverMsg{err: err}
			}
			pushSnapshotsToClients(host, game)
			return clientFailoverMsg{
				promoted: true,
				host:     host,
				game:     game,
				snapshot: game.Snapshot(0),
				note:     tr("online_failover_promoted"),
			}
		}

		var lastErr error
		for attempt := 1; attempt <= 16; attempt++ {
			newCli, err := netp2p.RejoinSession(inv, fs.Name, fs.DesiredRole, fs.SessionID, 1)
			if err == nil {
				return clientFailoverMsg{
					cli:  newCli,
					note: tr("online_failover_rejoined"),
				}
			}
			lastErr = err
			time.Sleep(time.Duration(minInt(attempt, 6)) * 300 * time.Millisecond)
		}
		return clientFailoverMsg{err: fmt.Errorf("failover não concluiu reconexão: %w", lastErr)}
	}
}

func waitHostActionCmd(host *netp2p.HostSession) tea.Cmd {
	return func() tea.Msg {
		a, ok := <-host.Actions()
		return hostActionMsg{action: a, ok: ok}
	}
}

func waitHostEventCmd(host *netp2p.HostSession) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-host.Events()
		return hostEventMsg{text: ev, ok: ok}
	}
}

func waitClientStateCmd(cli *netp2p.ClientSession) tea.Cmd {
	return func() tea.Msg {
		st, ok := <-cli.StateUpdates()
		return clientStateMsg{snapshot: st, ok: ok}
	}
}

func waitClientEventCmd(cli *netp2p.ClientSession) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-cli.Events()
		return clientEventMsg{text: ev, ok: ok}
	}
}
