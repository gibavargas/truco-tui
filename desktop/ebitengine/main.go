package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"truco-tui/internal/appcore"
	"truco-tui/internal/truco"
)

const (
	screenWidth      = 800
	screenHeight     = 600
	defaultWindowW   = 1280
	defaultWindowH   = 900
	minWindowW       = 960
	minWindowH       = 720
	cardWidth        = 90
	cardHeight       = 130
	cardSpacing      = 15
	textGlyphWidth   = 7
	textGlyphHeight  = 14
	statusBannerSecs = 7
)

var (
	woodDark     = color.RGBA{45, 26, 14, 255}
	woodMid      = color.RGBA{95, 55, 30, 255}
	feltDark     = color.RGBA{12, 55, 38, 255}
	feltMid      = color.RGBA{22, 112, 76, 255}
	gold         = color.RGBA{233, 197, 122, 255}
	ivory        = color.RGBA{255, 250, 239, 255}
	cardShadow   = color.RGBA{0, 0, 0, 92}
	panelBG      = color.RGBA{8, 13, 15, 220}
	accentRed    = color.RGBA{229, 57, 53, 255}
	accentGreen  = color.RGBA{67, 160, 71, 255}
	accentOrange = color.RGBA{245, 127, 23, 255}
)

type Button struct {
	X, Y, W, H int
	Label      string
	OnClick    func()
	Disabled   bool
}

func (b *Button) Update(mx, my int, clicked bool) {
	if b.Disabled {
		return
	}
	if clicked && mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H {
		if b.OnClick != nil {
			b.OnClick()
		}
	}
}

func (b *Button) Draw(screen *ebiten.Image, g *Game, mx, my int) {
	bg := panelBG
	if b.Disabled {
		bg = color.RGBA{40, 40, 40, 210}
	} else if mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H {
		bg = color.RGBA{44, 73, 85, 235}
	}

	vector.FillRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), bg, false)

	border := color.RGBA{255, 255, 255, 30}
	if !b.Disabled {
		border = gold
	}
	vector.StrokeRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), float32(max(1, g.ss(1))), border, false)

	textScale := g.bodyTextScale()
	textW, textH := g.measureText(b.Label, textScale)
	tx := b.X + (b.W-textW)/2
	ty := b.Y + (b.H-textH)/2
	textColor := ivory
	if b.Disabled {
		textColor = color.RGBA{185, 185, 185, 255}
	}
	g.drawText(screen, b.Label, tx, ty, textScale, textColor)
}

type TextBox struct {
	X, Y, W, H  int
	Text        string
	Placeholder string
	Focused     bool
}

func (t *TextBox) Update(mx, my int, clicked bool, typed []rune, backspace bool) {
	if clicked {
		t.Focused = mx >= t.X && mx <= t.X+t.W && my >= t.Y && my <= t.Y+t.H
	}
	if t.Focused {
		for _, r := range typed {
			if r >= 32 && r <= 126 {
				t.Text += string(r)
			}
		}
		if backspace && len(t.Text) > 0 {
			t.Text = t.Text[:len(t.Text)-1]
		}
	}
}

func (t *TextBox) Draw(screen *ebiten.Image, g *Game) {
	bg := color.RGBA{20, 20, 20, 220}
	vector.FillRect(screen, float32(t.X), float32(t.Y), float32(t.W), float32(t.H), bg, false)

	border := color.RGBA{255, 255, 255, 30}
	if t.Focused {
		border = gold
	}
	vector.StrokeRect(screen, float32(t.X), float32(t.Y), float32(t.W), float32(t.H), float32(max(1, g.ss(1))), border, false)

	val := t.Text
	textScale := g.bodyTextScale()
	textColor := ivory
	if val == "" && !t.Focused {
		val = t.Placeholder
		textColor = color.RGBA{180, 180, 180, 255}
	} else {
		if t.Focused && (time.Now().UnixNano()/500000000)%2 == 0 {
			val += "_"
		}
	}
	_, textH := g.measureText(val, textScale)
	g.drawText(screen, val, t.X+g.ss(8), t.Y+(t.H-textH)/2, textScale, textColor)
}

type PlayedCardAnim struct {
	StartX, StartY float64
	EndX, EndY     float64
	Progress       float64
}

type Game struct {
	runtime         *appcore.Runtime
	bundle          appcore.SnapshotBundle
	chatFeed        []string
	tickTimer       int
	trucoFlashTimer int
	diagnosticsOpen bool
	exitApp         bool

	inputName      TextBox
	inputInviteKey TextBox
	inputBindAddr  TextBox
	inputRelayURL  TextBox
	inputChat      TextBox

	setupNumPlayers  int
	setupTransport   string
	setupDesiredRole string
	currentScreen    string
	viewportW        int
	viewportH        int
	statusMessage    string
	statusTone       string
	statusUntil      time.Time
	textCache        map[string]*ebiten.Image

	cardHoverOffsets [3]float64
	playedCardAnims  map[string]*PlayedCardAnim
}


func clampFloat(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func (g *Game) viewSize() (int, int) {
	if g.viewportW <= 0 || g.viewportH <= 0 {
		return defaultWindowW, defaultWindowH
	}
	return g.viewportW, g.viewportH
}

func (g *Game) layoutScale() float64 {
	w, h := g.viewSize()
	return math.Min(float64(w)/screenWidth, float64(h)/screenHeight)
}

func (g *Game) layoutOffset() (int, int) {
	w, h := g.viewSize()
	scale := g.layoutScale()
	scaledW := int(math.Round(float64(screenWidth) * scale))
	scaledH := int(math.Round(float64(screenHeight) * scale))
	return (w - scaledW) / 2, (h - scaledH) / 2
}

func (g *Game) sx(v int) int {
	scale := g.layoutScale()
	ox, _ := g.layoutOffset()
	return ox + int(math.Round(float64(v)*scale))
}

func (g *Game) sy(v int) int {
	scale := g.layoutScale()
	_, oy := g.layoutOffset()
	return oy + int(math.Round(float64(v)*scale))
}

func (g *Game) ss(v int) int {
	scale := g.layoutScale()
	return max(1, int(math.Round(float64(v)*scale)))
}

func (g *Game) textScale(multiplier float64) float64 {
	return clampFloat(g.layoutScale()*multiplier, 1.0, 2.6)
}

func (g *Game) bodyTextScale() float64 {
	return g.textScale(0.9)
}

func (g *Game) titleTextScale() float64 {
	return g.textScale(1.3)
}

func (g *Game) textImage(text string) *ebiten.Image {
	if g.textCache == nil {
		g.textCache = make(map[string]*ebiten.Image)
	}
	if img, ok := g.textCache[text]; ok {
		return img
	}

	lines := strings.Split(text, "\n")
	maxWidth := 1
	for _, line := range lines {
		if w := len(line) * textGlyphWidth; w > maxWidth {
			maxWidth = w
		}
	}
	height := max(1, len(lines)*textGlyphHeight)
	img := ebiten.NewImage(maxWidth, height)
	for i, line := range lines {
		ebitenutil.DebugPrintAt(img, line, 0, i*textGlyphHeight)
	}
	g.textCache[text] = img
	return img
}

func colorScaleFor(clr color.Color) (float32, float32, float32, float32) {
	r, g, b, a := clr.RGBA()
	return float32(r) / 0xffff, float32(g) / 0xffff, float32(b) / 0xffff, float32(a) / 0xffff
}

func (g *Game) drawText(screen *ebiten.Image, text string, x, y int, scale float64, clr color.Color) {
	if text == "" {
		return
	}
	img := g.textImage(text)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	r, gg, b, a := colorScaleFor(clr)
	op.ColorScale.Scale(r, gg, b, a)
	screen.DrawImage(img, &op)
}

func (g *Game) measureText(text string, scale float64) (int, int) {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		if w := len(line) * textGlyphWidth; w > maxWidth {
			maxWidth = w
		}
	}
	return int(math.Round(float64(maxWidth) * scale)), int(math.Round(float64(len(lines)*textGlyphHeight) * scale))
}

func (g *Game) drawCenteredText(screen *ebiten.Image, text string, centerX, topY int, scale float64, clr color.Color) {
	w, _ := g.measureText(text, scale)
	g.drawText(screen, text, centerX-w/2, topY, scale, clr)
}

func (g *Game) wrapText(text string, maxChars int) []string {
	if maxChars <= 4 {
		return []string{text}
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := words[0]
		for _, word := range words[1:] {
			if len(current)+1+len(word) <= maxChars {
				current += " " + word
				continue
			}
			lines = append(lines, current)
			current = word
		}
		for len(current) > maxChars {
			lines = append(lines, current[:maxChars-1]+"-")
			current = current[maxChars-1:]
		}
		lines = append(lines, current)
	}
	return lines
}

func (g *Game) setStatus(message, tone string) {
	g.statusMessage = message
	g.statusTone = tone
	g.statusUntil = time.Now().Add(statusBannerSecs * time.Second)
}

func (g *Game) statusVisible() bool {
	return g.statusMessage != "" && time.Now().Before(g.statusUntil)
}

func (g *Game) scaledButton(x, y, w, h int, label string, disabled bool, onClick func()) *Button {
	return &Button{
		X:        g.sx(x),
		Y:        g.sy(y),
		W:        g.ss(w),
		H:        g.ss(h),
		Label:    label,
		Disabled: disabled,
		OnClick:  onClick,
	}
}

func (g *Game) syncTextBoxes() {
	g.inputName.X = g.sx(250)
	g.inputName.W = g.ss(300)
	g.inputName.H = g.ss(34)
	g.inputBindAddr.X = g.sx(250)
	g.inputBindAddr.W = g.ss(300)
	g.inputBindAddr.H = g.ss(34)
	g.inputRelayURL.X = g.sx(250)
	g.inputRelayURL.W = g.ss(300)
	g.inputRelayURL.H = g.ss(34)
	g.inputInviteKey.X = g.sx(250)
	g.inputInviteKey.W = g.ss(300)
	g.inputInviteKey.H = g.ss(34)

	switch g.currentScreen {
	case "setup_offline":
		g.inputName.Y = g.sy(200)
	case "setup_host":
		g.inputName.Y = g.sy(155)
		g.inputBindAddr.Y = g.sy(245)
		g.inputRelayURL.Y = g.sy(325)
	case "setup_join":
		g.inputInviteKey.Y = g.sy(190)
		g.inputName.Y = g.sy(285)
	case "lobby", "match":
		g.inputChat.X = g.sx(530)
		g.inputChat.Y = g.sy(328)
		g.inputChat.W = g.ss(188)
		g.inputChat.H = g.ss(34)
	}
}

func (g *Game) sendChat() {
	msg := strings.TrimSpace(g.inputChat.Text)
	if msg == "" {
		g.setStatus("Digite uma mensagem antes de enviar.", "warn")
		return
	}
	g.dispatchIntent(appcore.IntentSendChat, appcore.SendChatPayload{Text: msg})
	g.inputChat.Text = ""
	g.inputChat.Focused = false
	g.setStatus("Mensagem enviada.", "ok")
}

type handCardLayout struct {
	cardX int
	cardY int
	cardW int
	cardH int
	downY int
	downH int
}

func (g *Game) matchCardLayout(handSize int) []handCardLayout {
	layouts := make([]handCardLayout, 0, handSize)
	startX := (520 - (handSize*cardWidth + (handSize-1)*cardSpacing)) / 2
	startY := 380
	for i := 0; i < handSize; i++ {
		baseX := startX + i*(cardWidth+cardSpacing)
		layouts = append(layouts, handCardLayout{
			cardX: g.sx(baseX),
			cardY: g.sy(startY),
			cardW: g.ss(cardWidth),
			cardH: g.ss(cardHeight),
			downY: g.sy(startY + 135),
			downH: g.ss(24),
		})
	}
	return layouts
}

func (g *Game) Update() error {
	if g.exitApp {
		g.runtime.Close()
		return ebiten.Termination
	}

	g.drainEvents()

	g.tickTimer++
	if g.tickTimer >= 30 {
		g.tickTimer = 0
		g.dispatchIntent(appcore.IntentTick, appcore.TickPayload{MaxSteps: 5})
	}

	g.bundle = g.runtime.SnapshotBundle()
	g.autoNavigateScreen()
	g.syncTextBoxes()

	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	var typed []rune
	typed = ebiten.AppendInputChars(typed)
	backspace := inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyDelete)

	g.updateTextBoxes(mx, my, clicked, typed, backspace)
	g.updateButtons(mx, my, clicked)

	if g.currentScreen == "match" {
		g.updateMatchInteractions(mx, my, clicked)
	}

	g.updateMatchAnimations()

	g.updateCursorShape(mx, my)

	return nil
}

func (g *Game) autoNavigateScreen() {
	mode := g.bundle.Mode
	switch mode {
	case appcore.ModeIdle:
		if g.currentScreen == "lobby" || g.currentScreen == "match" {
			g.currentScreen = "home"
		}
	case appcore.ModeHostLobby, appcore.ModeClientLobby:
		g.currentScreen = "lobby"
	case appcore.ModeOfflineMatch, appcore.ModeHostMatch, appcore.ModeClientMatch:
		g.currentScreen = "match"
	}
}

func (g *Game) updateTextBoxes(mx, my int, clicked bool, typed []rune, backspace bool) {
	switch g.currentScreen {
	case "setup_offline":
		g.inputName.Update(mx, my, clicked, typed, backspace)
	case "setup_host":
		g.inputName.Update(mx, my, clicked, typed, backspace)
		g.inputBindAddr.Update(mx, my, clicked, typed, backspace)
		g.inputRelayURL.Update(mx, my, clicked, typed, backspace)
	case "setup_join":
		g.inputName.Update(mx, my, clicked, typed, backspace)
		g.inputInviteKey.Update(mx, my, clicked, typed, backspace)
	case "lobby", "match":
		g.inputChat.Update(mx, my, clicked, typed, backspace)
		if g.inputChat.Focused && inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.sendChat()
		}
	}
}

func (g *Game) updateButtons(mx, my int, clicked bool) {
	for _, btn := range g.getButtons() {
		btn.Update(mx, my, clicked)
	}
}

func (g *Game) updateMatchInteractions(mx, my int, clicked bool) {
	if g.bundle.Match == nil {
		return
	}

	actions := g.bundle.UI.Actions
	if !actions.CanPlayCard {
		return
	}

	localSeat := actions.LocalPlayerID
	if localSeat < 0 || localSeat >= len(g.bundle.Match.Players) {
		return
	}

	me := g.bundle.Match.Players[localSeat]
	hand := me.Hand
	layouts := g.matchCardLayout(len(hand))

	for i := range hand {
		cardLayout := layouts[i]
		if clicked && mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.cardY && my <= cardLayout.cardY+cardLayout.cardH {
			g.dispatchIntent(appcore.IntentGameAction, appcore.GameActionPayload{
				Action:    "play",
				CardIndex: i,
				FaceDown:  false,
			})
			break
		}

		canFaceDown := g.bundle.Match.CurrentHand.Round >= 2
		if canFaceDown && clicked && mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.downY && my <= cardLayout.downY+cardLayout.downH {
			g.dispatchIntent(appcore.IntentGameAction, appcore.GameActionPayload{
				Action:    "play",
				CardIndex: i,
				FaceDown:  true,
			})
			break
		}
	}
}

func (g *Game) updateCursorShape(mx, my int) {
	anyHovered := false
	for _, btn := range g.getButtons() {
		if !btn.Disabled && mx >= btn.X && mx <= btn.X+btn.W && my >= btn.Y && my <= btn.Y+btn.H {
			anyHovered = true
			break
		}
	}

	if !anyHovered && g.currentScreen == "match" && g.bundle.Match != nil {
		actions := g.bundle.UI.Actions
		localSeat := actions.LocalPlayerID
		if actions.CanPlayCard && localSeat >= 0 && localSeat < len(g.bundle.Match.Players) {
			me := g.bundle.Match.Players[localSeat]
			for _, cardLayout := range g.matchCardLayout(len(me.Hand)) {
				if mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.cardY && my <= cardLayout.cardY+cardLayout.cardH {
					anyHovered = true
				}
				canFaceDown := g.bundle.Match.CurrentHand.Round >= 2
				if canFaceDown && mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.downY && my <= cardLayout.downY+cardLayout.downH {
					anyHovered = true
				}
			}
		}
	}

	if anyHovered {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}

func (g *Game) dispatchIntent(kind string, payload any) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Erro serializando payload de intent %s: %v", kind, err)
			return
		}
		raw = b
	}

	intent := appcore.AppIntent{
		Kind:    kind,
		Payload: raw,
	}

	if err := g.runtime.DispatchIntent(intent); err != nil {
		log.Printf("Falha no dispatch da intent %s: %v", kind, err)
		g.chatFeed = append(g.chatFeed, fmt.Sprintf("[Sistema] Erro: %v", err))
		g.setStatus(fmt.Sprintf("Falha ao executar acao: %v", err), "error")
	}
}

func (g *Game) drainEvents() {
	for {
		ev, ok := g.runtime.PollEvent()
		if !ok {
			break
		}
		switch ev.Kind {
		case appcore.EventChat:
			if m, ok := ev.Payload.(map[string]any); ok {
				author := m["author"]
				text := m["text"]
				g.chatFeed = append(g.chatFeed, fmt.Sprintf("%v: %v", author, text))
			}
		case appcore.EventSystem:
			if m, ok := ev.Payload.(map[string]any); ok {
				text := m["text"]
				g.chatFeed = append(g.chatFeed, fmt.Sprintf("[Sistema] %v", text))
				g.setStatus(fmt.Sprintf("%v", text), "info")
			}
		case appcore.EventError:
			var errMsg string
			if m, ok := ev.Payload.(map[string]any); ok {
				if msg, exists := m["message"]; exists {
					errMsg = fmt.Sprintf("%v", msg)
				}
			}
			if errMsg == "" {
				errMsg = fmt.Sprintf("%v", ev.Payload)
			}
			g.chatFeed = append(g.chatFeed, fmt.Sprintf("[Erro] %s", errMsg))
			g.setStatus(errMsg, "error")
		case appcore.EventClientJoined:
			if m, ok := ev.Payload.(map[string]any); ok {
				name := m["name"]
				g.chatFeed = append(g.chatFeed, fmt.Sprintf("[Sistema] %v conectou-se.", name))
				g.setStatus(fmt.Sprintf("%v entrou na sala.", name), "ok")
			}
		case appcore.EventSessionClosed:
			g.chatFeed = append(g.chatFeed, "[Sistema] Conexão encerrada.")
			g.setStatus("Conexao encerrada.", "warn")
		case appcore.EventMatchStarted:
			g.chatFeed = append(g.chatFeed, "[Sistema] Partida iniciada!")
			g.setStatus("Partida iniciada.", "ok")
		}
	}
}

func (g *Game) getButtons() []*Button {
	var list []*Button
	locale := g.bundle.Locale
	if locale == "" {
		locale = "pt-BR"
	}
	t := func(key string) string {
		if translations[locale] != nil {
			if val, ok := translations[locale][key]; ok {
				return val
			}
		}
		return key
	}

	switch g.currentScreen {
	case "home":
		list = append(list, g.scaledButton(250, 180, 300, 40, t("play-offline"), false, func() {
			g.currentScreen = "setup_offline"
		}))
		list = append(list, g.scaledButton(250, 240, 300, 40, t("create-room"), false, func() {
			g.currentScreen = "setup_host"
		}))
		list = append(list, g.scaledButton(250, 300, 300, 40, t("join-room"), false, func() {
			g.currentScreen = "setup_join"
		}))
		langLabel := "Idioma: PT-BR"
		if locale == "en-US" {
			langLabel = "Language: EN-US"
		}
		list = append(list, g.scaledButton(250, 360, 300, 40, langLabel, false, func() {
			nextLocale := "pt-BR"
			if locale == "pt-BR" {
				nextLocale = "en-US"
			}
			g.dispatchIntent(appcore.IntentSetLocale, appcore.SetLocalePayload{Locale: nextLocale})
		}))
		list = append(list, g.scaledButton(250, 420, 300, 40, t("exit"), false, func() {
			g.exitApp = true
		}))

	case "setup_offline":
		list = append(list, g.scaledButton(250, 360, 140, 40, t("start"), false, func() {
			name := strings.TrimSpace(g.inputName.Text)
			if name == "" {
				name = "Jogador"
			}
			names := make([]string, g.setupNumPlayers)
			cpus := make([]bool, g.setupNumPlayers)
			names[0] = name
			for i := 1; i < g.setupNumPlayers; i++ {
				names[i] = fmt.Sprintf("CPU-%d", i+1)
				cpus[i] = true
			}
			g.dispatchIntent(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
				PlayerNames: names,
				CPUFlags:    cpus,
			})
		}))
		list = append(list, g.scaledButton(410, 360, 140, 40, t("back"), false, func() {
			g.currentScreen = "home"
		}))
		playersLabel := fmt.Sprintf("%s: %d", t("players"), g.setupNumPlayers)
		list = append(list, g.scaledButton(250, 280, 300, 40, playersLabel, false, func() {
			if g.setupNumPlayers == 2 {
				g.setupNumPlayers = 4
			} else {
				g.setupNumPlayers = 2
			}
		}))

	case "setup_host":
		list = append(list, g.scaledButton(250, 420, 140, 40, t("start"), false, func() {
			name := strings.TrimSpace(g.inputName.Text)
			if name == "" {
				name = "Jogador"
			}
			g.dispatchIntent(appcore.IntentCreateHostSession, appcore.CreateHostPayload{
				BindAddr:      g.inputBindAddr.Text,
				HostName:      name,
				NumPlayers:    g.setupNumPlayers,
				RelayURL:      g.inputRelayURL.Text,
				TransportMode: g.setupTransport,
			})
		}))
		list = append(list, g.scaledButton(410, 420, 140, 40, t("back"), false, func() {
			g.currentScreen = "home"
		}))
		playersLabel := fmt.Sprintf("%s: %d", t("players"), g.setupNumPlayers)
		list = append(list, g.scaledButton(250, 300, 300, 30, playersLabel, false, func() {
			if g.setupNumPlayers == 2 {
				g.setupNumPlayers = 4
			} else {
				g.setupNumPlayers = 2
			}
		}))
		transLabel := "Transport: Auto"
		if g.setupTransport == "tcp_tls" {
			transLabel = "Transport: TCP + TLS"
		}
		if g.setupTransport == "relay_quic_v2" {
			transLabel = "Transport: Relay QUIC v2"
		}
		if g.setupTransport == "tailnet_tsnet_v1" {
			transLabel = "Transport: Tailnet"
		}
		list = append(list, g.scaledButton(250, 340, 300, 30, transLabel, false, func() {
			if g.setupTransport == "auto" {
				g.setupTransport = "tcp_tls"
			} else if g.setupTransport == "tcp_tls" {
				g.setupTransport = "tailnet_tsnet_v1"
			} else if g.setupTransport == "tailnet_tsnet_v1" {
				g.setupTransport = "relay_quic_v2"
			} else {
				g.setupTransport = "auto"
			}
		}))

	case "setup_join":
		list = append(list, g.scaledButton(250, 380, 140, 40, t("connect"), false, func() {
			name := strings.TrimSpace(g.inputName.Text)
			if name == "" {
				name = "Jogador"
			}
			g.dispatchIntent(appcore.IntentJoinSession, appcore.JoinSessionPayload{
				Key:         g.inputInviteKey.Text,
				PlayerName:  name,
				DesiredRole: g.setupDesiredRole,
			})
		}))
		list = append(list, g.scaledButton(410, 380, 140, 40, t("back"), false, func() {
			g.currentScreen = "home"
		}))
		roleLabel := "Papel: Auto"
		if g.setupDesiredRole == "partner" {
			roleLabel = "Papel: Parceiro (Partner)"
		} else if g.setupDesiredRole == "opponent" {
			roleLabel = "Papel: Oponente (Opponent)"
		}
		list = append(list, g.scaledButton(250, 320, 300, 30, roleLabel, false, func() {
			if g.setupDesiredRole == "auto" {
				g.setupDesiredRole = "partner"
			} else if g.setupDesiredRole == "partner" {
				g.setupDesiredRole = "opponent"
			} else {
				g.setupDesiredRole = "auto"
			}
		}))

	case "lobby":
		isHost := g.bundle.Connection.IsHost
		canStart := isHost
		if isHost && g.bundle.Lobby != nil {
			filledCount := 0
			for _, name := range g.bundle.Lobby.Slots {
				if name != "" {
					filledCount++
				}
			}
			if filledCount < g.bundle.Lobby.NumPlayers {
				canStart = false
			}
		} else {
			canStart = false
		}

		list = append(list, g.scaledButton(40, 530, 160, 40, t("leave"), false, func() {
			g.dispatchIntent(appcore.IntentCloseSession, nil)
		}))
		diagLabel := t("diagnostics")
		if g.diagnosticsOpen {
			diagLabel = t("close-diag")
		}
		list = append(list, g.scaledButton(300, 530, 180, 40, diagLabel, false, func() {
			g.diagnosticsOpen = !g.diagnosticsOpen
		}))

		if isHost {
			list = append(list, g.scaledButton(40, 480, 180, 40, t("start-match"), !canStart, func() {
				g.dispatchIntent(appcore.IntentStartHostedMatch, nil)
			}))
		}

		if g.bundle.Lobby != nil && g.bundle.Lobby.InviteKey != "" {
			list = append(list, g.scaledButton(300, 110, 90, 30, t("copy"), false, func() {
				g.chatFeed = append(g.chatFeed, fmt.Sprintf("[Sistema] Chave da sala: %s", g.bundle.Lobby.InviteKey))
				g.setStatus("Chave exibida no painel de chat.", "info")
			}))
		}

		if g.bundle.UI.LobbySlots != nil {
			yOffset := 180
			for _, slot := range g.bundle.UI.LobbySlots {
				if slot.IsLocal {
					yOffset += 45
					continue
				}
				sSeat := slot.Seat
				if slot.CanVoteHost {
					list = append(list, g.scaledButton(300, yOffset+5, 90, 26, "Votar Host", false, func() {
						g.dispatchIntent(appcore.IntentVoteHost, appcore.HostVotePayload{CandidateSeat: sSeat})
					}))
				}
				if slot.CanRequestReplacement {
					list = append(list, g.scaledButton(400, yOffset+5, 80, 26, "Convite", false, func() {
						g.dispatchIntent(appcore.IntentRequestReplacementInvite, appcore.ReplacementInvitePayload{TargetSeat: sSeat})
					}))
				}
				yOffset += 45
			}
		}
		list = append(list, g.scaledButton(730, 328, 50, 34, t("send"), strings.TrimSpace(g.inputChat.Text) == "", func() {
			g.sendChat()
		}))

	case "match":
		list = append(list, g.scaledButton(720, 10, 70, 30, t("leave"), false, func() {
			g.dispatchIntent(appcore.IntentCloseSession, nil)
		}))

		diagLabel := t("diagnostics")
		if g.diagnosticsOpen {
			diagLabel = t("close-diag")
		}
		list = append(list, g.scaledButton(640, 10, 70, 30, diagLabel, false, func() {
			g.diagnosticsOpen = !g.diagnosticsOpen
		}))

		actions := g.bundle.UI.Actions

		trucoLabel := "TRUCO"
		if g.bundle.Match != nil {
			switch g.bundle.Match.CurrentHand.Stake {
			case 1, 2:
				trucoLabel = "TRUCO"
			case 3:
				trucoLabel = "SEIS"
			case 6:
				trucoLabel = "NOVE"
			case 9:
				trucoLabel = "DOZE"
			default:
				trucoLabel = "AUMENTAR"
			}
		}

		list = append(list, g.scaledButton(40, 535, 110, 40, trucoLabel, !actions.CanAskOrRaise, func() {
			g.dispatchIntent(appcore.IntentGameAction, appcore.GameActionPayload{Action: "truco"})
		}))

		list = append(list, g.scaledButton(160, 535, 110, 40, t("accept"), !actions.CanAccept, func() {
			g.dispatchIntent(appcore.IntentGameAction, appcore.GameActionPayload{Action: "accept"})
		}))

		list = append(list, g.scaledButton(280, 535, 110, 40, t("refuse"), !actions.CanRefuse, func() {
			g.dispatchIntent(appcore.IntentGameAction, appcore.GameActionPayload{Action: "refuse"})
		}))

		canNewHand := false
		if g.bundle.Match != nil && (g.bundle.Match.CurrentHand.Finished || g.bundle.Match.MatchFinished) {
			if actions.LocalPlayerID == 0 || g.bundle.Mode == appcore.ModeOfflineMatch {
				canNewHand = !g.bundle.Match.MatchFinished
			}
		}
		list = append(list, g.scaledButton(400, 535, 110, 40, t("new-hand"), !canNewHand, func() {
			g.dispatchIntent(appcore.IntentNewHand, nil)
		}))
		list = append(list, g.scaledButton(730, 328, 50, 34, t("send"), strings.TrimSpace(g.inputChat.Text) == "", func() {
			g.sendChat()
		}))
	}

	return list
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawWoodBackground(screen)
	g.drawTopBanner(screen)

	switch g.currentScreen {
	case "home":
		g.drawHomeScreen(screen)
	case "setup_offline":
		g.drawSetupOfflineScreen(screen)
	case "setup_host":
		g.drawSetupHostScreen(screen)
	case "setup_join":
		g.drawSetupJoinScreen(screen)
	case "lobby":
		g.drawLobbyScreen(screen)
	case "match":
		g.drawMatchScreen(screen)
	}
}

func (g *Game) drawWoodBackground(screen *ebiten.Image) {
	screen.Fill(woodDark)
	for i := 0; i < 8; i++ {
		x := float32(g.sx(i * (screenWidth / 8)))
		shade := woodMid
		if i%2 == 0 {
			shade = color.RGBA{75, 42, 22, 255}
		}
		vector.FillRect(screen, x, 0, float32(g.ss(screenWidth/8)), float32(g.viewSizeH()), shade, false)
		vector.FillRect(screen, x+float32(g.ss(screenWidth/8))-float32(g.ss(2)), 0, float32(g.ss(2)), float32(g.viewSizeH()), color.RGBA{0, 0, 0, 52}, false)
	}
}

func (g *Game) viewSizeH() int {
	_, h := g.viewSize()
	return h
}

func (g *Game) drawTopBanner(screen *ebiten.Image) {
	x := g.sx(18)
	y := g.sy(12)
	w := g.ss(500)
	h := g.ss(34)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{8, 14, 18, 210}, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 28}, false)
	title := "TRUCO PAULISTA · EBITENGINE"
	if g.bundle.Connection.IsOnline {
		title += " · ONLINE"
	} else {
		title += " · LOCAL"
	}
	g.drawText(screen, title, x+g.ss(10), y+g.ss(8), g.bodyTextScale(), ivory)

	if g.statusVisible() {
		statusX := g.sx(530)
		statusW := g.ss(252)
		statusColor := color.RGBA{43, 83, 63, 235}
		switch g.statusTone {
		case "warn":
			statusColor = color.RGBA{110, 70, 20, 235}
		case "error":
			statusColor = color.RGBA{118, 36, 36, 235}
		case "info":
			statusColor = color.RGBA{34, 60, 94, 235}
		}
		vector.FillRect(screen, float32(statusX), float32(y), float32(statusW), float32(h), statusColor, false)
		vector.StrokeRect(screen, float32(statusX), float32(y), float32(statusW), float32(h), float32(max(1, g.ss(1))), gold, false)
		g.drawText(screen, g.statusMessage, statusX+g.ss(10), y+g.ss(8), g.bodyTextScale(), ivory)
	}
}

func (g *Game) drawHomeScreen(screen *ebiten.Image) {
	x, y, w, h := g.sx(200), g.sy(100), g.ss(400), g.ss(400)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	g.drawCenteredText(screen, "TRUCO PAULISTA", x+w/2, y+g.ss(26), g.titleTextScale(), ivory)
	g.drawCenteredText(screen, "Partida local, lobby online e mesa 2D com mouse.", x+w/2, y+g.ss(68), g.bodyTextScale(), color.RGBA{220, 220, 220, 255})

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawSetupOfflineScreen(screen *ebiten.Image) {
	x, y, w, h := g.sx(200), g.sy(100), g.ss(400), g.ss(450)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	locale := g.bundle.Locale
	t := func(key string) string {
		if translations[locale] != nil {
			if val, ok := translations[locale][key]; ok {
				return val
			}
		}
		return key
	}

	g.drawCenteredText(screen, strings.ToUpper(t("play-offline")), x+w/2, y+g.ss(24), g.titleTextScale(), ivory)
	g.drawText(screen, "Nome do Jogador", g.sx(250), g.sy(175), g.bodyTextScale(), ivory)
	g.drawText(screen, "Troque entre 2 ou 4 jogadores antes de iniciar.", g.sx(250), g.sy(245), g.bodyTextScale(), color.RGBA{198, 198, 198, 255})
	g.inputName.Draw(screen, g)

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawSetupHostScreen(screen *ebiten.Image) {
	x, y, w, h := g.sx(200), g.sy(80), g.ss(400), g.ss(480)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	g.drawCenteredText(screen, "CRIAR SALA ONLINE", x+w/2, y+g.ss(20), g.titleTextScale(), ivory)
	g.drawText(screen, "Nome do host", g.sx(250), g.sy(130), g.bodyTextScale(), ivory)
	g.inputName.Draw(screen, g)

	g.drawText(screen, "Endereco de rede (bind)", g.sx(250), g.sy(220), g.bodyTextScale(), ivory)
	g.inputBindAddr.Draw(screen, g)

	g.drawText(screen, "Relay URL (opcional)", g.sx(250), g.sy(300), g.bodyTextScale(), ivory)
	g.inputRelayURL.Draw(screen, g)
	g.drawText(screen, "Os controles abaixo alternam jogadores e transporte.", g.sx(250), g.sy(390), g.bodyTextScale(), color.RGBA{200, 200, 200, 255})

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawSetupJoinScreen(screen *ebiten.Image) {
	x, y, w, h := g.sx(200), g.sy(100), g.ss(400), g.ss(450)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	g.drawCenteredText(screen, "ENTRAR EM SALA ONLINE", x+w/2, y+g.ss(24), g.titleTextScale(), ivory)

	g.drawText(screen, "Chave de convite", g.sx(250), g.sy(170), g.bodyTextScale(), ivory)
	g.inputInviteKey.Draw(screen, g)

	g.drawText(screen, "Seu nome", g.sx(250), g.sy(265), g.bodyTextScale(), ivory)
	g.inputName.Draw(screen, g)
	g.drawText(screen, "O papel desejado fica no seletor abaixo.", g.sx(250), g.sy(355), g.bodyTextScale(), color.RGBA{200, 200, 200, 255})

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawLobbyScreen(screen *ebiten.Image) {
	x, y, w, h := g.sx(30), g.sy(60), g.ss(460), g.ss(450)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	g.drawCenteredText(screen, "LOBBY DA PARTIDA", x+w/2, y+g.ss(16), g.titleTextScale(), ivory)

	inviteKey := "Aguardando..."
	if g.bundle.Lobby != nil && g.bundle.Lobby.InviteKey != "" {
		inviteKey = g.bundle.Lobby.InviteKey
	}
	g.drawText(screen, "Chave de convite", g.sx(60), g.sy(115), g.bodyTextScale(), ivory)
	g.drawText(screen, inviteKey, g.sx(60), g.sy(140), g.textScale(1.1), gold)
	g.drawText(screen, "Passe o mouse nos assentos para ver acoes e use o chat para coordenar a partida.", g.sx(60), g.sy(165), g.bodyTextScale(), color.RGBA{205, 205, 205, 255})

	slots := g.bundle.UI.LobbySlots
	yOffset := 180
	for _, slot := range slots {
		sBG := color.RGBA{20, 30, 40, 150}
		if slot.IsLocal {
			sBG = color.RGBA{30, 50, 40, 200}
		}
		vector.FillRect(screen, float32(g.sx(50)), float32(g.sy(yOffset)), float32(g.ss(420)), float32(g.ss(36)), sBG, false)
		borderClr := color.RGBA{100, 100, 100, 100}
		if slot.IsLocal {
			borderClr = gold
		}
		vector.StrokeRect(screen, float32(g.sx(50)), float32(g.sy(yOffset)), float32(g.ss(420)), float32(g.ss(36)), float32(max(1, g.ss(1))), borderClr, false)

		statusStr := "Vazio"
		if slot.IsOccupied {
			statusStr = "Conectado"
			if slot.IsProvisionalCPU {
				statusStr = "CPU Provisória"
			}
		}

		nameText := slot.Name
		if nameText == "" {
			nameText = "(Sem jogador)"
		}
		if slot.IsLocal {
			nameText += " (Você)"
		}
		if slot.IsHost {
			nameText += " [HOST]"
		}

		slotStr := fmt.Sprintf("Assento %d · %s · %s", slot.Seat+1, nameText, statusStr)
		g.drawText(screen, slotStr, g.sx(65), g.sy(yOffset+10), g.bodyTextScale(), ivory)

		yOffset += 45
	}

	g.drawSidebar(screen)

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawSidebar(screen *ebiten.Image) {
	x, y, w, h := g.sx(520), g.sy(50), g.ss(270), g.ss(530)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), color.RGBA{255, 255, 255, 32}, false)

	g.drawCenteredText(screen, "BATE-PAPO / LOGS", x+w/2, y+g.ss(12), g.titleTextScale(), ivory)

	g.drawChatFeed(screen, 530, 90, 250, 230)
	g.inputChat.Draw(screen, g)
	g.drawText(screen, "Enter envia · botao para mouse", g.sx(530), g.sy(370), g.bodyTextScale(), color.RGBA{190, 190, 190, 255})

	if g.diagnosticsOpen {
		g.drawDiagnostics(screen, 530, 380, 250, 190)
	} else {
		g.drawGameLogs(screen, 530, 380, 250, 190)
	}
}

func (g *Game) drawChatFeed(screen *ebiten.Image, x, y, w, h int) {
	vector.FillRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), color.RGBA{15, 22, 28, 240}, false)
	vector.StrokeRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 16}, false)

	scale := g.bodyTextScale()
	lineHeight := max(g.ss(16), 16)
	maxLines := max(1, (g.ss(h)-g.ss(12))/lineHeight)
	maxChars := max(12, (g.ss(w)-g.ss(18))/max(1, int(math.Round(float64(textGlyphWidth)*scale))))

	var wrapped []string
	for _, msg := range g.chatFeed {
		wrapped = append(wrapped, g.wrapText(msg, maxChars)...)
	}
	startIdx := max(0, len(wrapped)-maxLines)
	yOffset := g.sy(y) + g.ss(6)
	for i := startIdx; i < len(wrapped); i++ {
		g.drawText(screen, wrapped[i], g.sx(x)+g.ss(8), yOffset, scale, ivory)
		yOffset += lineHeight
	}
}

func (g *Game) drawGameLogs(screen *ebiten.Image, x, y, w, h int) {
	vector.FillRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), color.RGBA{12, 12, 12, 240}, false)
	vector.StrokeRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 16}, false)

	g.drawText(screen, "Logs de jogo", g.sx(x+8), g.sy(y+6), g.bodyTextScale(), ivory)

	logs := []string{"Aguardando início..."}
	if g.bundle.Match != nil {
		logs = g.bundle.Match.Logs
	}

	scale := g.bodyTextScale()
	lineHeight := max(g.ss(16), 16)
	maxLines := max(1, (g.ss(h)-g.ss(28))/lineHeight)
	maxChars := max(12, (g.ss(w)-g.ss(18))/max(1, int(math.Round(float64(textGlyphWidth)*scale))))
	var wrapped []string
	for _, msg := range logs {
		wrapped = append(wrapped, g.wrapText(msg, maxChars)...)
	}
	startIdx := max(0, len(wrapped)-maxLines)
	yOffset := g.sy(y + 26)
	for i := startIdx; i < len(wrapped); i++ {
		g.drawText(screen, wrapped[i], g.sx(x+8), yOffset, scale, ivory)
		yOffset += lineHeight
	}
}

func (g *Game) drawDiagnostics(screen *ebiten.Image, x, y, w, h int) {
	vector.FillRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), color.RGBA{24, 15, 15, 240}, false)
	vector.StrokeRect(screen, float32(g.sx(x)), float32(g.sy(y)), float32(g.ss(w)), float32(g.ss(h)), float32(max(1, g.ss(1))), gold, false)

	g.drawText(screen, "DIAGNOSTICOS", g.sx(x+8), g.sy(y+6), g.bodyTextScale(), ivory)

	diag := g.bundle.Diagnostics
	transportStr := "local"
	protoVer := 0
	if g.bundle.Connection.Network != nil {
		transportStr = g.bundle.Connection.Network.Transport
		protoVer = g.bundle.Connection.Network.NegotiatedProtocolVersion
	}
	if transportStr == "" {
		transportStr = "local"
	}

	lines := []string{
		fmt.Sprintf("Fila de Eventos: %d", diag.EventBacklog),
		fmt.Sprintf("Seed PCG Lo: %d", diag.ReplaySeedLo),
		fmt.Sprintf("Seed PCG Hi: %d", diag.ReplaySeedHi),
		fmt.Sprintf("Transporte: %s", transportStr),
		fmt.Sprintf("Online: %v", g.bundle.Connection.IsOnline),
		fmt.Sprintf("Protocolo Negoc.: v%d", protoVer),
	}

	yOffset := g.sy(y + 26)
	for _, line := range lines {
		g.drawText(screen, line, g.sx(x+8), yOffset, g.bodyTextScale(), ivory)
		yOffset += max(g.ss(16), 16)
	}
}

func (g *Game) drawMatchScreen(screen *ebiten.Image) {
	xc, yc := float32(g.sx(260)), float32(g.sy(280))
	vector.FillCircle(screen, xc, yc, float32(g.ss(210)), color.RGBA{65, 38, 20, 255}, false)
	vector.StrokeCircle(screen, xc, yc, float32(g.ss(210)), float32(g.ss(2)), gold, false)
	vector.FillCircle(screen, xc, yc, float32(g.ss(200)), feltDark, false)
	vector.FillCircle(screen, xc, yc, float32(g.ss(185)), feltMid, false)
	vector.StrokeCircle(screen, xc, yc, float32(g.ss(185)), float32(max(1, g.ss(1))), color.RGBA{255, 245, 214, 24}, false)
	g.drawCenteredText(screen, "Clique numa carta para jogar · use VIRADA quando permitido", g.sx(260), g.sy(472), g.bodyTextScale(), color.RGBA{236, 236, 236, 255})

	if g.bundle.Match == nil {
		return
	}

	localSeat := g.bundle.UI.Actions.LocalPlayerID
	numPlayers := g.bundle.Match.NumPlayers

	g.drawCard(screen, 160, 215, g.bundle.Match.CurrentHand.Vira, "VIRA", 0.0)
	g.drawManilhaIndicator(screen, 270, 215, string(g.bundle.Match.CurrentHand.Manilha))

	for i, player := range g.bundle.Match.Players {
		relSeat := (i - localSeat + numPlayers) % numPlayers
		if numPlayers == 2 && relSeat == 1 {
			relSeat = 2
		}
		g.drawSeat(screen, relSeat, player)
	}

	played := g.bundle.Match.CurrentHand.RoundCards
	for _, pc := range played {
		relSeat := (pc.PlayerID - localSeat + numPlayers) % numPlayers
		if numPlayers == 2 && relSeat == 1 {
			relSeat = 2
		}
		g.drawPlayedCardSpatially(screen, relSeat, pc)
	}

	g.drawScoreboardHUD(screen)
	g.drawLocalPlayerHand(screen, localSeat)

	if g.bundle.Match.CurrentHand.PendingRaiseFor != -1 {
		g.drawRaiseFlashOverlay(screen)
	}

	if g.bundle.Match.MatchFinished {
		g.drawGameOverOverlay(screen)
	}

	g.drawSidebar(screen)

	mx, my := ebiten.CursorPosition()
	for _, btn := range g.getButtons() {
		btn.Draw(screen, g, mx, my)
	}
}

func (g *Game) drawManilhaIndicator(screen *ebiten.Image, x, y int, manilha string) {
	fx, fy := float32(g.sx(x)), float32(g.sy(y))
	cw, ch := float32(g.ss(cardWidth)), float32(g.ss(cardHeight))
	vector.FillRect(screen, fx, fy, cw, ch, color.RGBA{40, 50, 45, 230}, false)
	vector.StrokeRect(screen, fx, fy, cw, ch, float32(g.ss(2)), gold, false)
	g.drawCenteredText(screen, "MANILHA", g.sx(x+cardWidth/2), g.sy(y+12), g.bodyTextScale(), ivory)
	g.drawCenteredText(screen, manilha, g.sx(x+cardWidth/2), g.sy(y+56), g.textScale(1.5), ivory)
}

func getRoleBadges(playerID, dealerID, numPlayers int) []string {
	var badges []string
	if playerID == dealerID {
		badges = append(badges, "Pé 🦶")
		badges = append(badges, "Distr 🃏")
	} else if playerID == (dealerID+1)%numPlayers {
		badges = append(badges, "Mão ✋")
	}
	return badges
}

func (g *Game) drawSeat(screen *ebiten.Image, relSeat int, player truco.Player) {
	var x, y int
	switch relSeat {
	case 0:
		x, y = 260, 360
	case 1:
		x, y = 430, 260
	case 2:
		x, y = 260, 95
	case 3:
		x, y = 90, 260
	}

	w, h := g.ss(130), g.ss(44)
	sx, sy := g.sx(x), g.sy(y)
	bx, by := float32(sx-w/2), float32(sy-h/2)

	isTurn := g.bundle.Match.TurnPlayer == player.ID

	bgClr := panelBG
	borderClr := color.RGBA{120, 120, 120, 120}
	if isTurn {
		bgClr = color.RGBA{30, 60, 45, 230}
		borderClr = gold
	}
	vector.FillRect(screen, bx, by, float32(w), float32(h), bgClr, false)
	vector.StrokeRect(screen, bx, by, float32(w), float32(h), float32(max(1, g.ss(1))), borderClr, false)

	nameStr := player.Name
	if len(nameStr) > 10 {
		nameStr = nameStr[:8] + ".."
	}
	lbl := fmt.Sprintf("%s T%d", nameStr, player.Team+1)
	g.drawText(screen, lbl, sx-w/2+g.ss(8), sy-h/2+g.ss(5), g.bodyTextScale(), ivory)

	badges := getRoleBadges(player.ID, g.bundle.Match.CurrentHand.Dealer, g.bundle.Match.NumPlayers)
	badgeStr := ""
	for _, b := range badges {
		badgeStr += b + " "
	}
	g.drawText(screen, strings.TrimSpace(badgeStr), sx-w/2+g.ss(8), sy-h/2+g.ss(22), g.bodyTextScale(), color.RGBA{224, 214, 176, 255})

	if relSeat != 0 {
		handSize := len(player.Hand)
		var cx, cy int
		switch relSeat {
		case 1:
			cx, cy = sx+w/2+g.ss(8), sy-g.ss(13)
			for i := 0; i < handSize; i++ {
				g.drawMiniCardBack(screen, cx+i*g.ss(16), cy)
			}
		case 2:
			cx, cy = sx-(handSize*g.ss(20))/2, sy-h/2-g.ss(28)
			for i := 0; i < handSize; i++ {
				g.drawMiniCardBack(screen, cx+i*g.ss(20), cy)
			}
		case 3:
			cx, cy = sx-w/2-g.ss(8)-handSize*g.ss(16), sy-g.ss(13)
			for i := 0; i < handSize; i++ {
				g.drawMiniCardBack(screen, cx+i*g.ss(16), cy)
			}
		}
	}
}

func (g *Game) drawMiniCardBack(screen *ebiten.Image, x, y int) {
	fx, fy := float32(x), float32(y)
	w, h := float32(g.ss(18)), float32(g.ss(26))
	vector.FillRect(screen, fx, fy, w, h, color.RGBA{18, 48, 62, 255}, false)
	vector.StrokeRect(screen, fx, fy, w, h, float32(max(1, g.ss(1))), gold, false)
	vector.StrokeLine(screen, fx, fy, fx+w, fy+h, float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 30}, false)
	vector.StrokeLine(screen, fx+w, fy, fx, fy+h, float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 30}, false)
}

func (g *Game) drawPlayedCardSpatially(screen *ebiten.Image, relSeat int, pc truco.PlayedCard) {
	var targetX, targetY int
	switch relSeat {
	case 0:
		targetX, targetY = 260-45, 315
	case 1:
		targetX, targetY = 370, 215
	case 2:
		targetX, targetY = 260-45, 75
	case 3:
		targetX, targetY = 60, 215
	}

	x, y := float64(g.sx(targetX)), float64(g.sy(targetY))
	key := fmt.Sprintf("%d-%s-%v", pc.PlayerID, pc.Card.String(), pc.FaceDown)
	if anim, ok := g.playedCardAnims[key]; ok {
		t := anim.Progress
		factor := 1.0 - (1.0-t)*(1.0-t)*(1.0-t)
		x = anim.StartX + (anim.EndX-anim.StartX)*factor
		y = anim.StartY + (anim.EndY-anim.StartY)*factor
	}

	lbl := fmt.Sprintf("P%d", pc.PlayerID+1)
	if pc.FaceDown {
		g.drawCardBackLarge(screen, int(x), int(y), lbl)
	} else {
		g.drawCard(screen, int(x), int(y), pc.Card, lbl, 0.0)
	}
}

func (g *Game) drawCardBackLarge(screen *ebiten.Image, x, y int, label string) {
	fx, fy := float32(x), float32(y)
	cw, ch := float32(g.ss(cardWidth)), float32(g.ss(cardHeight))
	vector.FillRect(screen, fx, fy, cw, ch, color.RGBA{18, 48, 62, 255}, false)
	vector.StrokeRect(screen, fx, fy, cw, ch, float32(g.ss(2)), gold, false)
	vector.StrokeLine(screen, fx, fy, fx+cw, fy+ch, float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 30}, false)
	vector.StrokeLine(screen, fx+cw, fy, fx, fy+ch, float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 30}, false)
	g.drawText(screen, label, x+g.ss(8), y+g.ss(8), g.bodyTextScale(), ivory)
	g.drawCenteredText(screen, "VIRADA", x+g.ss(cardWidth/2), y+g.ss(56), g.bodyTextScale(), ivory)
}

func (g *Game) drawScoreboardHUD(screen *ebiten.Image) {
	score0 := g.bundle.Match.MatchPoints[0]
	score1 := g.bundle.Match.MatchPoints[1]

	locale := g.bundle.Locale
	t := func(key string) string {
		if translations[locale] != nil {
			if val, ok := translations[locale][key]; ok {
				return val
			}
		}
		return key
	}

	g.drawPanel(screen, 28, 18, 126, 44, t("nos"), fmt.Sprintf("%d", score0), false)
	g.drawTricksLEDs(screen, 32, 68, 0)

	stakeVal := fmt.Sprintf("%d", g.bundle.Match.CurrentHand.Stake)
	if g.bundle.Match.CurrentHand.PendingRaiseFor != -1 {
		next := nextStake(g.bundle.Match.CurrentHand.Stake)
		stakeVal = fmt.Sprintf("%d->%d", g.bundle.Match.CurrentHand.Stake, next)
	}
	g.drawPanel(screen, 260-58, 18, 116, 50, t("vale"), stakeVal, true)

	g.drawPanel(screen, 520-154, 18, 126, 44, t("eles"), fmt.Sprintf("%d", score1), false)
	g.drawTricksLEDs(screen, 520-150, 68, 1)
}

func (g *Game) drawTricksLEDs(screen *ebiten.Image, x, y, team int) {
	results := g.bundle.Match.CurrentHand.TrickResults
	for i := 0; i < 3; i++ {
		cx := float32(g.sx(x + i*16))
		cy := float32(g.sy(y))

		clr := color.RGBA{80, 80, 80, 255}
		if i < len(results) {
			res := results[i]
			if res == -1 {
				clr = accentOrange
			} else if res == team {
				clr = accentGreen
			} else {
				clr = accentRed
			}
		}
		vector.FillCircle(screen, cx, cy, float32(g.ss(5)), clr, true)
		vector.StrokeCircle(screen, cx, cy, float32(g.ss(5)), float32(max(1, g.ss(1))), color.RGBA{255, 255, 255, 30}, true)
	}
}

func (g *Game) drawLocalPlayerHand(screen *ebiten.Image, localSeat int) {
	me := g.bundle.Match.Players[localSeat]
	hand := me.Hand

	mx, my := ebiten.CursorPosition()

	locale := g.bundle.Locale
	t := func(key string) string {
		if translations[locale] != nil {
			if val, ok := translations[locale][key]; ok {
				return val
			}
		}
		return key
	}

	layouts := g.matchCardLayout(len(hand))
	for i, card := range hand {
		cardLayout := layouts[i]

		g.drawCard(screen, cardLayout.cardX, cardLayout.cardY, card, fmt.Sprintf("%d", i+1), g.cardHoverOffsets[i])

		canFaceDown := g.bundle.Match.CurrentHand.Round >= 2
		if canFaceDown {
			btnBG := color.RGBA{20, 20, 20, 220}
			btnBorder := color.RGBA{150, 150, 150, 120}
			if mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.downY && my <= cardLayout.downY+cardLayout.downH {
				btnBG = color.RGBA{60, 40, 40, 220}
				btnBorder = gold
			}
			vector.FillRect(screen, float32(cardLayout.cardX), float32(cardLayout.downY), float32(cardLayout.cardW), float32(cardLayout.downH), btnBG, false)
			vector.StrokeRect(screen, float32(cardLayout.cardX), float32(cardLayout.downY), float32(cardLayout.cardW), float32(cardLayout.downH), float32(max(1, g.ss(1))), btnBorder, false)

			lbl := t("facedown")
			textW, textH := g.measureText(lbl, g.bodyTextScale())
			tx := cardLayout.cardX + (cardLayout.cardW-textW)/2
			ty := cardLayout.downY + (cardLayout.downH-textH)/2
			g.drawText(screen, lbl, tx, ty, g.bodyTextScale(), ivory)
		}
	}
}

func (g *Game) drawRaiseFlashOverlay(screen *ebiten.Image) {
	g.trucoFlashTimer++
	if (g.trucoFlashTimer/20)%2 == 0 {
		x, y, w, h := g.sx(40), g.sy(200), g.ss(440), g.ss(60)
		vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{180, 30, 30, 240}, false)
		vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

		requester := "Alguém"
		reqID := g.bundle.Match.CurrentHand.RaiseRequester
		if reqID >= 0 && reqID < len(g.bundle.Match.Players) {
			requester = g.bundle.Match.Players[reqID].Name
		}

		next := nextStake(g.bundle.Match.CurrentHand.Stake)
		msg := fmt.Sprintf("%s pediu %s", strings.ToUpper(requester), strings.ToUpper(raiseLabel(next)))
		g.drawCenteredText(screen, msg, x+w/2, y+g.ss(18), g.titleTextScale(), ivory)
	}
}

func (g *Game) drawGameOverOverlay(screen *ebiten.Image) {
	x, y, w, h := g.sx(30), g.sy(60), g.ss(460), g.ss(450)
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{10, 15, 20, 230}, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(g.ss(2)), gold, false)

	locale := g.bundle.Locale
	t := func(key string) string {
		if translations[locale] != nil {
			if val, ok := translations[locale][key]; ok {
				return val
			}
		}
		return key
	}

	winnerTeam := g.bundle.Match.WinnerTeam
	localSeat := g.bundle.UI.Actions.LocalPlayerID
	localTeam := 0
	if localSeat >= 0 && localSeat < len(g.bundle.Match.Players) {
		localTeam = g.bundle.Match.Players[localSeat].Team
	}

	msg := t("defeat")
	clr := accentRed
	if winnerTeam == localTeam {
		msg = t("winner")
		clr = accentGreen
	}

	g.drawCenteredText(screen, "PARTIDA ENCERRADA", x+w/2, g.sy(180), g.titleTextScale(), ivory)

	boxX, boxY, boxW, boxH := g.sx(100), g.sy(240), g.ss(320), g.ss(50)
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxW), float32(boxH), clr, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxW), float32(boxH), float32(g.ss(2)), gold, false)
	g.drawCenteredText(screen, msg, boxX+boxW/2, boxY+g.ss(14), g.titleTextScale(), ivory)

	g.drawCenteredText(screen, "Use SAIR para retornar ao menu.", x+w/2, g.sy(338), g.bodyTextScale(), color.RGBA{215, 215, 215, 255})
}

func (g *Game) drawPanel(screen *ebiten.Image, x, y, w, h int, label, value string, accent bool) {
	border := color.RGBA{255, 255, 255, 42}
	if accent {
		border = gold
	}
	sx, sy, sw, sh := g.sx(x), g.sy(y), g.ss(w), g.ss(h)
	vector.FillRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), panelBG, false)
	vector.StrokeRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), float32(max(1, g.ss(1))), border, false)
	g.drawText(screen, label, sx+g.ss(10), sy+g.ss(7), g.bodyTextScale(), ivory)
	valW, _ := g.measureText(value, g.textScale(1.1))
	g.drawText(screen, value, sx+sw-g.ss(12)-valW, sy+sh-g.ss(20), g.textScale(1.1), gold)
}

func (g *Game) drawCard(screen *ebiten.Image, x, y int, c truco.Card, label string, hoverOffset float64) {
	offsetY := int(math.Round(float64(g.ss(15)) * hoverOffset * -1))

	fx, fy := float32(x), float32(y+offsetY)
	cw, ch := float32(g.ss(cardWidth)), float32(g.ss(cardHeight))

	vector.FillRect(screen, fx+float32(g.ss(5)), fy+float32(g.ss(8)), cw, ch, cardShadow, false)
	vector.FillRect(screen, fx, fy, cw, ch, ivory, false)
	vector.FillRect(screen, fx+float32(g.ss(5)), fy+float32(g.ss(5)), cw-float32(g.ss(10)), ch-float32(g.ss(10)), color.RGBA{248, 236, 212, 255}, false)

	border := color.RGBA{118, 84, 43, 130}
	if hoverOffset > 0.0 {
		border = gold
	}
	vector.StrokeRect(screen, fx, fy, cw, ch, float32(g.ss(2)), border, false)
	vector.StrokeRect(screen, fx+float32(g.ss(6)), fy+float32(g.ss(6)), cw-float32(g.ss(12)), ch-float32(g.ss(12)), float32(max(1, g.ss(1))), color.RGBA{80, 55, 32, 36}, false)

	suitColor := color.RGBA{0, 0, 0, 255}
	if c.Suit == truco.Hearts || c.Suit == truco.Diamonds {
		suitColor = color.RGBA{190, 54, 45, 255}
	}

	g.drawText(screen, label, x+g.ss(8), y+offsetY+g.ss(8), g.bodyTextScale(), color.RGBA{32, 32, 32, 255})

	cardStr := c.String()
	parts := strings.Split(cardStr, " de ")
	if len(parts) == 2 {
		g.drawCenteredText(screen, parts[0], x+g.ss(cardWidth/2), y+offsetY+g.ss(cardHeight/2-22), g.bodyTextScale(), color.RGBA{35, 35, 35, 255})
		g.drawCenteredText(screen, parts[1], x+g.ss(cardWidth/2), y+offsetY+g.ss(cardHeight/2+4), g.bodyTextScale(), color.RGBA{35, 35, 35, 255})
	} else {
		g.drawCenteredText(screen, cardStr, x+g.ss(cardWidth/2), y+offsetY+g.ss(cardHeight/2-6), g.bodyTextScale(), color.RGBA{35, 35, 35, 255})
	}

	vector.FillCircle(screen, fx+cw-float32(g.ss(15)), fy+float32(g.ss(15)), float32(g.ss(6)), suitColor, true)
	vector.FillCircle(screen, fx+cw/2, fy+ch/2+float32(g.ss(24)), float32(g.ss(18)), color.RGBA{0, 0, 0, 12}, true)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.viewportW = max(outsideWidth, minWindowW)
	g.viewportH = max(outsideHeight, minWindowH)
	return g.viewportW, g.viewportH
}

func nextStake(s int) int {
	switch s {
	case 1:
		return 3
	case 3:
		return 6
	case 6:
		return 9
	case 9:
		return 12
	default:
		return s
	}
}

func raiseLabel(stake int) string {
	switch stake {
	case 3:
		return "truco"
	case 6:
		return "seis"
	case 9:
		return "nove"
	case 12:
		return "doze"
	default:
		return fmt.Sprintf("%d", stake)
	}
}

var translations = map[string]map[string]string{
	"en-US": {
		"play-offline": "Offline Match",
		"create-room":  "Host Room",
		"join-room":    "Join Room",
		"language":     "Language",
		"exit":         "Exit",
		"back":         "Back",
		"start":        "Start",
		"name":         "Name",
		"players":      "Players",
		"bind-address": "Bind Address",
		"relay-url":    "Relay URL",
		"invite-key":   "Invite Key",
		"desired-role": "Desired Role",
		"connect":      "Connect",
		"waiting":      "Waiting...",
		"start-match":  "Start Match",
		"leave":        "Leave",
		"copy":         "Copy",
		"send":         "Send",
		"you":          "You",
		"offline":      "offline",
		"host":         "host",
		"accept":       "Accept",
		"refuse":       "Refuse",
		"new-hand":     "New Hand",
		"diagnostics":  "Diag",
		"close-diag":   "Close Diag",
		"turn":         "YOUR TURN",
		"facedown":     "Face Down",
		"winner":       "WINNER!",
		"defeat":       "DEFEAT",
		"tie":          "TIE",
		"nos":          "NOS",
		"eles":         "ELES",
		"vale":         "VALE",
	},
	"pt-BR": {
		"play-offline": "Partida Offline",
		"create-room":  "Criar Host",
		"join-room":    "Entrar Online",
		"language":     "Idioma",
		"exit":         "Sair",
		"back":         "Voltar",
		"start":        "Iniciar",
		"name":         "Nome",
		"players":      "Jogadores",
		"bind-address": "Endereço Bind",
		"relay-url":    "Relay URL",
		"invite-key":   "Chave Convite",
		"desired-role": "Papel Desejado",
		"connect":      "Conectar",
		"waiting":      "Aguardando...",
		"start-match":  "Iniciar Partida",
		"leave":        "Sair",
		"copy":         "Copiar",
		"send":         "Enviar",
		"you":          "Você",
		"offline":      "offline",
		"host":         "host",
		"accept":       "Aceitar",
		"refuse":       "Recusar",
		"new-hand":     "Nova mão",
		"diagnostics":  "Diag",
		"close-diag":   "Fechar Diag",
		"turn":         "SUA VEZ",
		"facedown":     "Virada",
		"winner":       "VITORIA!",
		"defeat":       "DERROTA",
		"tie":          "EMPATE",
		"nos":          "NÓS",
		"eles":         "ELES",
		"vale":         "VALE",
	},
}

func (g *Game) updateMatchAnimations() {
	if g.bundle.Match == nil || g.currentScreen != "match" {
		return
	}

	// 1. Update hand hover offsets
	actions := g.bundle.UI.Actions
	localSeat := actions.LocalPlayerID
	if localSeat >= 0 && localSeat < len(g.bundle.Match.Players) {
		me := g.bundle.Match.Players[localSeat]
		hand := me.Hand
		layouts := g.matchCardLayout(len(hand))
		mx, my := ebiten.CursorPosition()
		for i := 0; i < 3; i++ {
			target := 0.0
			if i < len(hand) && i < len(layouts) {
				cardLayout := layouts[i]
				if mx >= cardLayout.cardX && mx <= cardLayout.cardX+cardLayout.cardW && my >= cardLayout.cardY && my <= cardLayout.cardY+cardLayout.cardH {
					target = 1.0
				}
			}
			diff := target - g.cardHoverOffsets[i]
			if diff > -0.01 && diff < 0.01 {
				g.cardHoverOffsets[i] = target
			} else {
				g.cardHoverOffsets[i] += diff * 0.25
			}
		}
	}

	// 2. Update played card animations
	if g.playedCardAnims == nil {
		g.playedCardAnims = make(map[string]*PlayedCardAnim)
	}

	numPlayers := g.bundle.Match.NumPlayers
	played := g.bundle.Match.CurrentHand.RoundCards

	// Keep track of active keys to clean up old ones
	activeKeys := make(map[string]bool)

	for _, pc := range played {
		relSeat := (pc.PlayerID - localSeat + numPlayers) % numPlayers
		if numPlayers == 2 && relSeat == 1 {
			relSeat = 2
		}

		key := fmt.Sprintf("%d-%s-%v", pc.PlayerID, pc.Card.String(), pc.FaceDown)
		activeKeys[key] = true

		if _, exists := g.playedCardAnims[key]; !exists {
			// Find start position based on player seat
			var startX, startY float64
			switch relSeat {
			case 0:
				startX, startY = float64(g.sx(260-45)), float64(g.sy(380))
			case 1:
				startX, startY = float64(g.sx(430-45)), float64(g.sy(260-65))
			case 2:
				startX, startY = float64(g.sx(260-45)), float64(g.sy(95-65))
			case 3:
				startX, startY = float64(g.sx(90-45)), float64(g.sy(260-65))
			}

			// Find target (end) position
			var endX, endY float64
			switch relSeat {
			case 0:
				endX, endY = float64(g.sx(260-45)), float64(g.sy(315))
			case 1:
				endX, endY = float64(g.sx(370)), float64(g.sy(215))
			case 2:
				endX, endY = float64(g.sx(260-45)), float64(g.sy(75))
			case 3:
				endX, endY = float64(g.sx(60)), float64(g.sy(215))
			}

			g.playedCardAnims[key] = &PlayedCardAnim{
				StartX:   startX,
				StartY:   startY,
				EndX:     endX,
				EndY:     endY,
				Progress: 0.0,
			}
		} else {
			anim := g.playedCardAnims[key]
			anim.Progress += 0.08
			if anim.Progress > 1.0 {
				anim.Progress = 1.0
			}
		}
	}

	// Clean up old animations
	for key := range g.playedCardAnims {
		if !activeKeys[key] {
			delete(g.playedCardAnims, key)
		}
	}
}

func main() {
	ebiten.SetWindowSize(defaultWindowW, defaultWindowH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSizeLimits(minWindowW, minWindowH, -1, -1)
	ebiten.SetWindowTitle("Truco Ebitengine - 2D Client")

	g := &Game{
		runtime:          appcore.NewRuntime(),
		currentScreen:    "home",
		setupNumPlayers:  4,
		setupTransport:   "auto",
		setupDesiredRole: "auto",
		playedCardAnims:  make(map[string]*PlayedCardAnim),
	}

	g.inputName = TextBox{X: 250, Y: 200, W: 300, H: 30, Placeholder: "Nome do Jogador"}
	g.inputBindAddr = TextBox{X: 250, Y: 235, W: 300, H: 30, Placeholder: "0.0.0.0:0"}
	g.inputRelayURL = TextBox{X: 250, Y: 285, W: 300, H: 30, Placeholder: "Relay URL (Opcional)"}
	g.inputInviteKey = TextBox{X: 250, Y: 200, W: 300, H: 30, Placeholder: "Chave de Convite"}
	g.inputChat = TextBox{X: 530, Y: 330, W: 250, H: 30, Placeholder: "Mensagem chat..."}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
