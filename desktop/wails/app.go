//go:build wails

package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"truco-tui/internal/appcore"
)

type App struct {
	ctx     context.Context
	runtime *appcore.Runtime
}

func NewApp() *App {
	return &App{runtime: appcore.NewRuntime()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(context.Context) {
	if a.runtime != nil {
		_ = a.runtime.Close()
	}
}

func (a *App) Snapshot() appcore.SnapshotBundle {
	return a.runtime.SnapshotBundle()
}

func (a *App) PollEvents() []appcore.AppEvent {
	events := make([]appcore.AppEvent, 0, 16)
	for {
		ev, ok := a.runtime.PollEvent()
		if !ok {
			break
		}
		events = append(events, ev)
	}
	return events
}

func (a *App) DispatchIntentJSON(intentJSON string) *appcore.AppError {
	var intent appcore.AppIntent
	if err := json.Unmarshal([]byte(intentJSON), &intent); err != nil {
		return &appcore.AppError{Code: "invalid_json", Message: err.Error()}
	}
	if err := a.runtime.DispatchIntent(intent); err != nil {
		return &appcore.AppError{Code: "dispatch_failed", Message: err.Error()}
	}
	return nil
}

func (a *App) StartOfflineGame(playerName string, numPlayers int) *appcore.AppError {
	if numPlayers != 4 {
		numPlayers = 2
	}
	names := make([]string, numPlayers)
	cpus := make([]bool, numPlayers)
	names[0] = normalizeDesktopName(playerName)
	for i := 1; i < numPlayers; i++ {
		names[i] = "CPU-" + strconv.Itoa(i+1)
		cpus[i] = true
	}
	return a.dispatch("new_offline_game", appcore.NewOfflineGamePayload{
		PlayerNames: names,
		CPUFlags:    cpus,
	})
}

func (a *App) Tick(maxSteps int) *appcore.AppError {
	return a.dispatch("tick", appcore.TickPayload{MaxSteps: maxSteps})
}

func (a *App) PlayCard(cardIndex int) *appcore.AppError {
	return a.dispatch("game_action", appcore.GameActionPayload{Action: "play", CardIndex: cardIndex})
}

func (a *App) RequestTruco() *appcore.AppError {
	return a.dispatch("game_action", appcore.GameActionPayload{Action: "truco"})
}

func (a *App) AcceptTruco() *appcore.AppError {
	return a.dispatch("game_action", appcore.GameActionPayload{Action: "accept"})
}

func (a *App) RefuseTruco() *appcore.AppError {
	return a.dispatch("game_action", appcore.GameActionPayload{Action: "refuse"})
}

func (a *App) NewHand() *appcore.AppError {
	return a.dispatch("new_hand", nil)
}

func (a *App) SetLocale(locale string) *appcore.AppError {
	return a.dispatch("set_locale", appcore.SetLocalePayload{Locale: locale})
}

func (a *App) CreateHostSession(hostName string, numPlayers int, bindAddr string, relayURL string, transportMode string) *appcore.AppError {
	return a.dispatch("create_host_session", appcore.CreateHostPayload{
		BindAddr:      bindAddr,
		HostName:      normalizeDesktopName(hostName),
		NumPlayers:    numPlayers,
		RelayURL:      relayURL,
		TransportMode: transportMode,
	})
}

func (a *App) JoinSession(key string, playerName string, desiredRole string) *appcore.AppError {
	return a.dispatch("join_session", appcore.JoinSessionPayload{
		Key:         key,
		PlayerName:  normalizeDesktopName(playerName),
		DesiredRole: desiredRole,
	})
}

func (a *App) StartHostedMatch() *appcore.AppError {
	return a.dispatch("start_hosted_match", nil)
}

func (a *App) SendChat(text string) *appcore.AppError {
	return a.dispatch("send_chat", appcore.SendChatPayload{Text: text})
}

func (a *App) VoteHost(candidateSeat int) *appcore.AppError {
	return a.dispatch("vote_host", appcore.HostVotePayload{CandidateSeat: candidateSeat})
}

func (a *App) RequestReplacementInvite(targetSeat int) *appcore.AppError {
	return a.dispatch("request_replacement_invite", appcore.ReplacementInvitePayload{TargetSeat: targetSeat})
}

func (a *App) Reset() *appcore.AppError {
	return a.dispatch("reset", nil)
}

func (a *App) CloseSession() *appcore.AppError {
	return a.dispatch("close_session", nil)
}

func (a *App) dispatch(kind string, payload any) *appcore.AppError {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return &appcore.AppError{Code: "invalid_payload", Message: err.Error()}
		}
		raw = b
	}
	if err := a.runtime.DispatchIntent(appcore.AppIntent{Kind: kind, Payload: raw}); err != nil {
		return &appcore.AppError{Code: "dispatch_failed", Message: err.Error()}
	}
	return nil
}

func normalizeDesktopName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Jogador"
	}
	return value
}
