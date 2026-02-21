package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"truco-tui/internal/netp2p"
	"truco-tui/internal/truco"
)

// TUI concentra toda a interface textual em ASCII.
//
// Observação importante de frontend:
// Este arquivo é intencionalmente comentado em maior detalhe porque os fluxos de
// tela + input em terminal costumam ser a área com mais manutenção e regressão.
type TUI struct {
	in  *bufio.Reader
	out *os.File
}

func New() *TUI {
	return &TUI{
		in:  bufio.NewReader(os.Stdin),
		out: os.Stdout,
	}
}

func (t *TUI) Run() error {
	for {
		menu := newLobbyMenuModel()
		p := tea.NewProgram(menu)
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		choice := finalModel.(lobbyMenuModel).selected

		switch choice {
		case tr("lobby_choose_offline"):
			if err := t.runOffline(); err != nil {
				t.waitErr(err)
			}
		case tr("lobby_choose_online"):
			if err := t.runOnlineLobby(); err != nil {
				t.waitErr(err)
			}
		case tr("lobby_choose_language"):
			t.switchLocaleFlow()
		case tr("lobby_choose_exit"):
			return nil
		default:
			t.wait(tr("menu_invalid_option"))
		}
	}
}

func (t *TUI) switchLocaleFlow() {
	choice := strings.TrimSpace(t.ask(tr("language_prompt")))
	if !setLocale(choice) {
		t.wait(tr("language_invalid"))
		return
	}
	t.wait(fmt.Sprintf(tr("language_changed"), localeCode()))
}

func (t *TUI) runOffline() error {
	numPlayers, names, cpus, err := t.collectOfflineSetup()
	if err != nil {
		return err
	}
	_ = numPlayers

	game, err := truco.NewGame(names, cpus)
	if err != nil {
		return err
	}

	// Initialize Bubble Tea Model
	model := InitialModel(game)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf(tr("bubble_tea_run_error"), err)
	}
	return nil
}

func (t *TUI) collectOfflineSetup() (int, []string, []bool, error) {
	playersRaw := t.ask(tr("prompt_players_count"))
	numPlayers, err := strconv.Atoi(playersRaw)
	if err != nil || (numPlayers != 2 && numPlayers != 4) {
		return 0, nil, nil, errors.New(tr("error_invalid_player_count"))
	}

	localName := strings.TrimSpace(t.ask(tr("prompt_your_name")))
	if localName == "" {
		localName = tr("default_you")
	}

	names := make([]string, numPlayers)
	cpus := make([]bool, numPlayers)
	names[0] = localName
	cpus[0] = false

	for i := 1; i < numPlayers; i++ {
		ans := strings.ToLower(strings.TrimSpace(t.ask(fmt.Sprintf(tr("prompt_player_cpu"), i+1))))
		if ans == "s" || ans == "sim" || ans == "y" || ans == "yes" || ans == "" {
			names[i] = fmt.Sprintf("CPU-%d", i+1)
			cpus[i] = true
			continue
		}
		names[i] = strings.TrimSpace(t.ask(fmt.Sprintf(tr("prompt_player_name"), i+1)))
		if names[i] == "" {
			names[i] = fmt.Sprintf(tr("default_player_name"), i+1)
		}
	}
	return numPlayers, names, cpus, nil
}

func (t *TUI) runOnlineLobby() error {
	t.clear()
	fmt.Fprintln(t.out, tr("online_title"))
	fmt.Fprintln(t.out, "1) "+tr("online_create_host"))
	fmt.Fprintln(t.out, "2) "+tr("online_join"))
	fmt.Fprintln(t.out, "0) "+tr("online_back"))
	choice := t.ask(tr("online_choice"))

	switch choice {
	case "1":
		return t.hostLobbyFlow()
	case "2":
		return t.joinLobbyFlow()
	default:
		return nil
	}
}

func (t *TUI) hostLobbyFlow() error {
	hostName := strings.TrimSpace(t.ask(tr("host_name_prompt")))
	if hostName == "" {
		hostName = tr("host_default_name")
	}
	nRaw := strings.TrimSpace(t.ask(tr("host_players_prompt")))
	n, err := strconv.Atoi(nRaw)
	if err != nil {
		return err
	}

	host, key, err := netp2p.NewHostSession("0.0.0.0:0", hostName, n)
	if err != nil {
		return err
	}
	defer host.Close()

	model := newOnlineLobbyHostModel(host, key)
	p := tea.NewProgram(model, tea.WithAltScreen())
	fm, err := p.Run()
	if err != nil {
		return err
	}
	final := fm.(onlineLobbyModel)
	if final.startMatch {
		return t.runHostMatch(host)
	}
	return nil
}

func (t *TUI) joinLobbyFlow() error {
	name := strings.TrimSpace(t.ask(tr("join_name_prompt")))
	if name == "" {
		name = tr("join_default_name")
	}
	key := strings.TrimSpace(t.ask(tr("join_key_prompt")))
	role := strings.ToLower(strings.TrimSpace(t.ask(tr("join_role_prompt"))))
	if role == "" {
		role = "auto"
	}

	cli, err := netp2p.JoinSession(key, name, role)
	if err != nil {
		return err
	}
	defer cli.Close()

	model := newOnlineLobbyClientModel(cli)
	p := tea.NewProgram(model, tea.WithAltScreen())
	fm, err := p.Run()
	if err != nil {
		return err
	}
	final := fm.(onlineLobbyModel)
	if final.startMatch {
		return t.runClientMatch(cli)
	}
	return nil
}

func (t *TUI) runHostMatch(host *netp2p.HostSession) error {
	slots := host.Slots()
	cpu := make([]bool, len(slots))
	game, err := truco.NewGame(slots, cpu)
	if err != nil {
		return err
	}
	model := newOnlineHostModel(host, game)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
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
		return fmt.Errorf(tr("online_unknown_action_format"), a.Action)
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

func (t *TUI) runClientMatch(cli *netp2p.ClientSession) error {
	var initial *truco.Snapshot
	for initial == nil {
		select {
		case st := <-cli.StateUpdates():
			s := st
			initial = &s
		case <-time.After(2 * time.Second):
			return fmt.Errorf("%s", tr("timeout_initial_state"))
		}
	}
	model := newOnlineClientModel(cli, *initial)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func (t *TUI) clear() {
	fmt.Fprint(t.out, "\033[2J\033[H")
}

func (t *TUI) ask(label string) string {
	fmt.Fprintf(t.out, "%s: ", label)
	line, _ := t.in.ReadString('\n')
	return strings.TrimSpace(line)
}

func (t *TUI) wait(msg string) {
	fmt.Fprintln(t.out, msg)
	fmt.Fprintln(t.out, tr("press_enter_continue"))
	_, _ = t.in.ReadString('\n')
}

func (t *TUI) waitErr(err error) {
	t.wait(tr("error_prefix") + err.Error())
}
