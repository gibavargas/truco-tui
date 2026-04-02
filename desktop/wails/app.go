//go:build wails

package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"truco-tui/internal/appcore"
)

const (
	runtimeUpdateEvent  = "truco:runtime:update"
	defaultPumpInterval = 120 * time.Millisecond
)

type RuntimeUpdate struct {
	Bundle appcore.SnapshotBundle `json:"bundle"`
	Events []appcore.AppEvent     `json:"events,omitempty"`
}

type App struct {
	ctx          context.Context
	runtime      *appcore.Runtime
	updateMu     sync.Mutex
	emit         func(eventName string, payload any)
	pumpInterval time.Duration
	stopPump     chan struct{}
	pumpDone     chan struct{}
}

func NewApp() *App {
	return &App{
		runtime:      appcore.NewRuntime(),
		pumpInterval: defaultPumpInterval,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.emit == nil {
		a.emit = func(eventName string, payload any) {
			wailsruntime.EventsEmit(ctx, eventName, payload)
		}
	}
	a.startPump()
}

func (a *App) shutdown(context.Context) {
	a.stopPumpLoop()
	if a.runtime != nil {
		_ = a.runtime.Close()
	}
}

func (a *App) Snapshot() appcore.SnapshotBundle {
	return a.runtime.SnapshotBundle()
}

func (a *App) PollEvents() []appcore.AppEvent {
	return a.drainEvents()
}

func (a *App) RuntimeUpdateEventName() string {
	return runtimeUpdateEvent
}

func (a *App) DispatchIntentJSON(intentJSON string) *appcore.AppError {
	var intent appcore.AppIntent
	if err := json.Unmarshal([]byte(intentJSON), &intent); err != nil {
		return &appcore.AppError{Code: "invalid_json", Message: err.Error()}
	}
	if err := a.runtime.DispatchIntent(intent); err != nil {
		return &appcore.AppError{Code: "dispatch_failed", Message: err.Error()}
	}
	a.emitRuntimeUpdate(true)
	return nil
}

func (a *App) StartOfflineGame(playerName string, numPlayers int) *appcore.AppError {
	numPlayers = normalizeNumPlayers(numPlayers)
	names := make([]string, numPlayers)
	cpus := make([]bool, numPlayers)
	names[0] = normalizeDesktopName(playerName)
	for i := 1; i < numPlayers; i++ {
		names[i] = "CPU-" + strconv.Itoa(i+1)
		cpus[i] = true
	}
	return a.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerNames: names,
		CPUFlags:    cpus,
	})
}

func (a *App) Tick(maxSteps int) *appcore.AppError {
	return a.dispatch(appcore.IntentTick, appcore.TickPayload{MaxSteps: maxSteps})
}

func (a *App) PlayCard(cardIndex int) *appcore.AppError {
	return a.dispatch(appcore.IntentGameAction, appcore.GameActionPayload{Action: "play", CardIndex: cardIndex})
}

func (a *App) PlayFaceDownCard(cardIndex int) *appcore.AppError {
	return a.dispatch(appcore.IntentGameAction, appcore.GameActionPayload{
		Action:    "play",
		CardIndex: cardIndex,
		FaceDown:  true,
	})
}

func (a *App) RequestTruco() *appcore.AppError {
	return a.dispatch(appcore.IntentGameAction, appcore.GameActionPayload{Action: "truco"})
}

func (a *App) AcceptTruco() *appcore.AppError {
	return a.dispatch(appcore.IntentGameAction, appcore.GameActionPayload{Action: "accept"})
}

func (a *App) RefuseTruco() *appcore.AppError {
	return a.dispatch(appcore.IntentGameAction, appcore.GameActionPayload{Action: "refuse"})
}

func (a *App) NewHand() *appcore.AppError {
	return a.dispatch(appcore.IntentNewHand, nil)
}

func (a *App) SetLocale(locale string) *appcore.AppError {
	return a.dispatch(appcore.IntentSetLocale, appcore.SetLocalePayload{Locale: strings.TrimSpace(locale)})
}

func (a *App) CreateHostSession(hostName string, numPlayers int, bindAddr string, relayURL string, transportMode string) *appcore.AppError {
	return a.dispatch(appcore.IntentCreateHostSession, appcore.CreateHostPayload{
		BindAddr:      strings.TrimSpace(bindAddr),
		HostName:      normalizeDesktopName(hostName),
		NumPlayers:    normalizeNumPlayers(numPlayers),
		RelayURL:      strings.TrimSpace(relayURL),
		TransportMode: normalizeTransportMode(transportMode),
	})
}

func (a *App) JoinSession(key string, playerName string, desiredRole string) *appcore.AppError {
	return a.dispatch(appcore.IntentJoinSession, appcore.JoinSessionPayload{
		Key:         strings.TrimSpace(key),
		PlayerName:  normalizeDesktopName(playerName),
		DesiredRole: normalizeDesiredRole(desiredRole),
	})
}

func (a *App) StartHostedMatch() *appcore.AppError {
	return a.dispatch(appcore.IntentStartHostedMatch, nil)
}

func (a *App) SendChat(text string) *appcore.AppError {
	return a.dispatch(appcore.IntentSendChat, appcore.SendChatPayload{Text: strings.TrimSpace(text)})
}

func (a *App) VoteHost(candidateSeat int) *appcore.AppError {
	return a.dispatch(appcore.IntentVoteHost, appcore.HostVotePayload{CandidateSeat: candidateSeat})
}

func (a *App) RequestReplacementInvite(targetSeat int) *appcore.AppError {
	return a.dispatch(appcore.IntentRequestReplacementInvite, appcore.ReplacementInvitePayload{TargetSeat: targetSeat})
}

func (a *App) Reset() *appcore.AppError {
	return a.CloseSession()
}

func (a *App) CloseSession() *appcore.AppError {
	return a.dispatch(appcore.IntentCloseSession, nil)
}

func (a *App) startPump() {
	if a.stopPump != nil {
		return
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	a.stopPump = stop
	a.pumpDone = done
	interval := a.pumpInterval
	if interval <= 0 {
		interval = defaultPumpInterval
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(done)

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				a.emitRuntimeUpdate(false)
			}
		}
	}()
}

func (a *App) stopPumpLoop() {
	if a.stopPump == nil {
		return
	}
	close(a.stopPump)
	<-a.pumpDone
	a.stopPump = nil
	a.pumpDone = nil
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
	a.emitRuntimeUpdate(true)
	return nil
}

func (a *App) emitRuntimeUpdate(force bool) bool {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()

	if a.emit == nil {
		return false
	}

	update := RuntimeUpdate{
		Bundle: a.runtime.SnapshotBundle(),
		Events: a.drainEvents(),
	}
	if !force && len(update.Events) == 0 {
		return false
	}
	a.emit(runtimeUpdateEvent, update)
	return true
}

func (a *App) drainEvents() []appcore.AppEvent {
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

func normalizeDesktopName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Jogador"
	}
	return value
}

func normalizeDesiredRole(value string) string {
	switch strings.TrimSpace(value) {
	case appcore.DesiredRolePartner:
		return appcore.DesiredRolePartner
	case appcore.DesiredRoleOpponent:
		return appcore.DesiredRoleOpponent
	default:
		return appcore.DesiredRoleAuto
	}
}

func normalizeTransportMode(value string) string {
	switch strings.TrimSpace(value) {
	case "tcp_tls":
		return "tcp_tls"
	case "relay_quic_v2":
		return "relay_quic_v2"
	default:
		return ""
	}
}

func normalizeNumPlayers(numPlayers int) int {
	if numPlayers == 4 {
		return 4
	}
	return 2
}
