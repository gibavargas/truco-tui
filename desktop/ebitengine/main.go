package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"truco-tui/internal/truco"
)

const (
	screenWidth  = 800
	screenHeight = 600
	cardWidth    = 90
	cardHeight   = 130
	cardSpacing  = 15
)

var (
	woodDark   = color.RGBA{45, 26, 14, 255}
	woodMid    = color.RGBA{95, 55, 30, 255}
	feltDark   = color.RGBA{12, 55, 38, 255}
	feltMid    = color.RGBA{22, 112, 76, 255}
	gold       = color.RGBA{233, 197, 122, 255}
	ivory      = color.RGBA{255, 250, 239, 255}
	cardShadow = color.RGBA{0, 0, 0, 92}
	panelBG    = color.RGBA{8, 13, 15, 186}
)

type Game struct {
	gameLogic  *truco.Game
	snapshot   truco.Snapshot
	message    string
	hoverIndex int
	lastCPUAct time.Time
}

func (g *Game) Update() error {
	// O jogador local é sempre ID 0 nesta implementação simplificada
	g.snapshot = g.gameLogic.Snapshot(0)

	x, y := ebiten.CursorPosition()
	g.updateHover(x, y)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleInput(x, y)
	}
	g.advanceCPUIfNeeded()

	return nil
}

func (g *Game) updateHover(x, y int) {
	g.hoverIndex = -1
	// Jogador local é o P0
	hand := g.snapshot.Players[0].Hand
	startX := (screenWidth - (len(hand)*cardWidth + (len(hand)-1)*cardSpacing)) / 2
	startY := screenHeight - cardHeight - 30

	for i := range hand {
		cx := startX + i*(cardWidth+cardSpacing)
		cy := startY
		if x >= cx && x <= cx+cardWidth && y >= cy && y <= cy+cardHeight {
			g.hoverIndex = i
			break
		}
	}
}

func (g *Game) handleInput(x, y int) {
	if g.hoverIndex != -1 {
		err := g.gameLogic.PlayCard(0, g.hoverIndex)
		if err != nil {
			g.message = fmt.Sprintf("Aviso: %v", err)
		} else {
			g.message = "Boa jogada!"
		}
	}
}

func (g *Game) advanceCPUIfNeeded() {
	if g.snapshot.MatchFinished || time.Since(g.lastCPUAct) < 650*time.Millisecond {
		return
	}
	isCPU, pid := g.gameLogic.IsCPUTurn()
	if !isCPU && g.snapshot.PendingRaiseFor != -1 {
		for _, player := range g.snapshot.Players {
			if player.CPU && player.Team == g.snapshot.PendingRaiseFor {
				isCPU = true
				pid = player.ID
				break
			}
		}
	}
	if !isCPU {
		return
	}
	action := truco.DecideCPUAction(g.gameLogic, pid)
	if err := applyCPUAction(g.gameLogic, pid, action); err != nil {
		g.message = fmt.Sprintf("CPU: %v", err)
	} else {
		g.message = "CPU jogou."
	}
	g.lastCPUAct = time.Now()
	g.snapshot = g.gameLogic.Snapshot(0)
}

func applyCPUAction(game *truco.Game, playerID int, action truco.CPUAction) error {
	switch action.Kind {
	case "play":
		return game.PlayCard(playerID, action.CardIndex)
	case "ask_truco":
		return game.AskTruco(playerID)
	case "raise":
		return game.RaiseTruco(playerID)
	case "accept":
		return game.RespondTruco(playerID, true)
	case "refuse":
		return game.RespondTruco(playerID, false)
	default:
		return fmt.Errorf("acao de CPU desconhecida: %s", action.Kind)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawRoom(screen)

	// Desenha o Vira (posicionado à esquerda)
	g.drawCard(screen, 50, screenHeight/2-cardHeight/2, g.snapshot.CurrentHand.Vira, "VIRA", false)

	// Desenha as mãos dos oponentes (versos das cartas)
	g.drawOpponents(screen)

	// Desenha a mão do jogador local (Jogador 0)
	g.drawLocalHand(screen)

	// Desenha as cartas jogadas na rodada (no centro)
	g.drawPlayedCards(screen)

	// Placar e Interface
	score0 := g.snapshot.MatchPoints[0]
	score1 := g.snapshot.MatchPoints[1]
	g.drawHUD(screen, score0, score1)
	g.drawMessage(screen)
}

func (g *Game) drawRoom(screen *ebiten.Image) {
	screen.Fill(woodDark)
	for i := 0; i < 8; i++ {
		x := float32(i * (screenWidth / 8))
		shade := woodMid
		if i%2 == 0 {
			shade = color.RGBA{75, 42, 22, 255}
		}
		vector.DrawFilledRect(screen, x, 0, screenWidth/8, screenHeight, shade, false)
		vector.DrawFilledRect(screen, x+float32(screenWidth/8)-2, 0, 2, screenHeight, color.RGBA{0, 0, 0, 52}, false)
	}

	vector.DrawFilledRect(screen, 24, 52, screenWidth-48, screenHeight-106, color.RGBA{57, 32, 17, 255}, false)
	vector.StrokeRect(screen, 24, 52, screenWidth-48, screenHeight-106, 4, gold, false)
	vector.DrawFilledRect(screen, 42, 70, screenWidth-84, screenHeight-142, feltDark, false)
	vector.DrawFilledRect(screen, 58, 86, screenWidth-116, screenHeight-174, feltMid, false)
	vector.StrokeRect(screen, 58, 86, screenWidth-116, screenHeight-174, 2, color.RGBA{255, 245, 214, 38}, false)

	for x := 70; x < screenWidth-70; x += 28 {
		vector.DrawFilledRect(screen, float32(x), 92, 1, screenHeight-188, color.RGBA{255, 244, 209, 10}, false)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image, score0, score1 int) {
	g.drawPanel(screen, 28, 18, 126, 44, "NOS", fmt.Sprintf("%d", score0), false)
	g.drawPanel(screen, screenWidth/2-58, 18, 116, 50, "VALE", fmt.Sprintf("%d", g.snapshot.CurrentHand.Stake), true)
	g.drawPanel(screen, screenWidth-154, 18, 126, 44, "ELES", fmt.Sprintf("%d", score1), false)
}

func (g *Game) drawPanel(screen *ebiten.Image, x, y, w, h int, label, value string, accent bool) {
	border := color.RGBA{255, 255, 255, 42}
	if accent {
		border = gold
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1.5, border, false)
	ebitenutil.DebugPrintAt(screen, label, x+10, y+8)
	ebitenutil.DebugPrintAt(screen, value, x+w-24, y+h-20)
}

func (g *Game) drawMessage(screen *ebiten.Image) {
	msgW := 360
	x := screenWidth/2 - msgW/2
	y := screenHeight - 28
	vector.DrawFilledRect(screen, float32(x), float32(y-8), float32(msgW), 26, panelBG, false)
	vector.StrokeRect(screen, float32(x), float32(y-8), float32(msgW), 26, 1, color.RGBA{255, 255, 255, 36}, false)
	ebitenutil.DebugPrintAt(screen, g.message, x+12, y)
}

func (g *Game) drawCard(screen *ebiten.Image, x, y int, c truco.Card, label string, isHovered bool) {
	offsetY := 0
	if isHovered {
		offsetY = -15
	}

	fx, fy := float32(x), float32(y+offsetY)

	// Sombra
	vector.DrawFilledRect(screen, fx+5, fy+8, cardWidth, cardHeight, cardShadow, false)
	// Fundo da carta
	vector.DrawFilledRect(screen, fx, fy, cardWidth, cardHeight, ivory, false)
	vector.DrawFilledRect(screen, fx+5, fy+5, cardWidth-10, cardHeight-10, color.RGBA{248, 236, 212, 255}, false)
	// Borda
	border := color.RGBA{118, 84, 43, 130}
	if isHovered {
		border = gold
	}
	vector.StrokeRect(screen, fx, fy, cardWidth, cardHeight, 2, border, false)
	vector.StrokeRect(screen, fx+6, fy+6, cardWidth-12, cardHeight-12, 1, color.RGBA{80, 55, 32, 36}, false)

	// Cor do naipe
	suitColor := color.RGBA{0, 0, 0, 255}
	if c.Suit == truco.Hearts || c.Suit == truco.Diamonds {
		suitColor = color.RGBA{190, 54, 45, 255}
	}

	// Valor e Naipe
	ebitenutil.DebugPrintAt(screen, label, x+5, y+offsetY+5)

	cardStr := c.String()
	ebitenutil.DebugPrintAt(screen, cardStr, x+cardWidth/2-15, y+offsetY+cardHeight/2-5)

	// Desenha um pequeno círculo colorido para o naipe no canto
	vector.DrawFilledCircle(screen, fx+cardWidth-15, fy+15, 6, suitColor, true)
	vector.DrawFilledCircle(screen, fx+cardWidth/2, fy+cardHeight/2+18, 18, color.RGBA{0, 0, 0, 12}, true)
}

func (g *Game) drawLocalHand(screen *ebiten.Image) {
	hand := g.snapshot.Players[0].Hand
	startX := (screenWidth - (len(hand)*cardWidth + (len(hand)-1)*cardSpacing)) / 2
	startY := screenHeight - cardHeight - 30

	for i, card := range hand {
		g.drawCard(screen, startX+i*(cardWidth+cardSpacing), startY, card, fmt.Sprintf("%d", i+1), i == g.hoverIndex)
	}
}

func (g *Game) drawOpponents(screen *ebiten.Image) {
	// Adversário Topo (P2)
	p2HandSize := len(g.snapshot.Players[2].Hand)
	startX := (screenWidth - (p2HandSize * 40)) / 2
	for i := 0; i < p2HandSize; i++ {
		g.drawCardBack(screen, startX+i*40, 112, 35, 50)
	}

	// Adversários Laterais (P1 e P3) - Simplificado
	g.drawCardBack(screen, 70, screenHeight/2-40, 40, 60)
	g.drawCardBack(screen, screenWidth-110, screenHeight/2-40, 40, 60)
}

func (g *Game) drawCardBack(screen *ebiten.Image, x, y, w, h int) {
	fx, fy := float32(x), float32(y)
	vector.DrawFilledRect(screen, fx+3, fy+4, float32(w), float32(h), cardShadow, false)
	vector.DrawFilledRect(screen, fx, fy, float32(w), float32(h), color.RGBA{18, 48, 62, 255}, false)
	vector.DrawFilledRect(screen, fx+4, fy+4, float32(w-8), float32(h-8), color.RGBA{28, 72, 84, 255}, false)
	vector.StrokeRect(screen, fx, fy, float32(w), float32(h), 1.5, color.RGBA{233, 197, 122, 120}, false)
	vector.StrokeRect(screen, fx+6, fy+6, float32(w-12), float32(h-12), 1, color.RGBA{255, 255, 255, 34}, false)
}

func (g *Game) drawPlayedCards(screen *ebiten.Image) {
	played := g.snapshot.CurrentHand.RoundCards
	centerX, centerY := screenWidth/2, screenHeight/2-30

	for i, pc := range played {
		// Distribui as cartas jogadas em volta do centro
		offsetX := (i%2)*110 - 55
		offsetY := (i/2)*140 - 70
		g.drawCard(screen, centerX+offsetX-cardWidth/2, centerY+offsetY-cardHeight/2, pc.Card, fmt.Sprintf("P%d", pc.PlayerID), false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Inicializa o jogo com 4 jogadores (2 times)
	playerNames := []string{"Você", "P1", "P2", "P3"}
	cpuFlags := []bool{false, true, true, true}
	logic, err := truco.NewGame(playerNames, cpuFlags)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Truco Ebitengine 10/10")

	g := &Game{
		gameLogic: logic,
		snapshot:  logic.Snapshot(0),
		message:   "Bem-vindo ao Truco!",
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
