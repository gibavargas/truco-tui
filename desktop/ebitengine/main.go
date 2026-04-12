package main

import (
	"fmt"
	"image/color"
	"log"

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

type Game struct {
	gameLogic  *truco.Game
	snapshot   truco.Snapshot
	message    string
	hoverIndex int
}

func (g *Game) Update() error {
	g.snapshot = g.gameLogic.Snapshot(0)

	x, y := ebiten.CursorPosition()
	g.updateHover(x, y)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleInput(x, y)
	}

	// Simulação básica: se não for a vez do jogador 0, a CPU joga (se implementado na lógica)
	// Como o truco.Game é síncrono, aqui apenas reagimos ao estado.
	return nil
}

func (g *Game) updateHover(x, y int) {
	g.hoverIndex = -1
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

func (g *Game) Draw(screen *ebiten.Image) {
	// Mesa de feltro verde com borda
	screen.Fill(color.RGBA{20, 80, 20, 255})
	vector.DrawFilledRect(screen, 10, 10, screenWidth-20, screenHeight-20, color.RGBA{30, 110, 30, 255}, false)

	// Desenha o Vira (posicionado à esquerda)
	g.drawCard(screen, 50, screenHeight/2-cardHeight/2, g.snapshot.CurrentHand.Vira, "VIRA", false)

	// Desenha as mãos dos oponentes (versos das cartas)
	g.drawOpponents(screen)

	// Desenha a mão do jogador local (Jogador 0)
	g.drawLocalHand(screen)

	// Desenha as cartas jogadas na rodada (no centro)
	g.drawPlayedCards(screen)

	// Placar e Interface
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SCORE - NÓS: %d | ELES: %d", g.snapshot.MatchPoints["0"], g.snapshot.MatchPoints["1"]), screenWidth-200, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mante: %d", g.snapshot.CurrentHand.Stake), screenWidth-200, 50)
	ebitenutil.DebugPrintAt(screen, g.message, screenWidth/2-50, screenHeight-20)
}

func (g *Game) drawCard(screen *ebiten.Image, x, y int, c truco.Card, label string, isHovered bool) {
	offsetY := 0
	if isHovered {
		offsetY = -15
	}

	fx, fy := float32(x), float32(y+offsetY)
	
	// Sombra
	vector.DrawFilledRect(screen, fx+4, fy+4, cardWidth, cardHeight, color.RGBA{0, 0, 0, 100}, false)
	// Fundo da carta
	vector.DrawFilledRect(screen, fx, fy, cardWidth, cardHeight, color.White, false)
	// Borda
	vector.StrokeRect(screen, fx, fy, cardWidth, cardHeight, 2, color.Black, false)

	// Cor do naipe
	suitColor := color.RGBA{0, 0, 0, 255}
	if c.Suit == truco.Hearts || c.Suit == truco.Diamonds {
		suitColor = color.RGBA{200, 0, 0, 255}
	}

	// Valor e Naipe
	ebitenutil.DebugPrintAt(screen, label, x+5, y+offsetY+5)
	
	cardStr := c.String()
	ebitenutil.DebugPrintAt(screen, cardStr, x+cardWidth/2-15, y+offsetY+cardHeight/2-5)
	
	// Desenha um pequeno círculo colorido para o naipe no canto
	vector.DrawFilledCircle(screen, fx+cardWidth-15, fy+15, 6, suitColor, true)
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
	startX := (screenWidth - (p2HandSize*40)) / 2
	for i := 0; i < p2HandSize; i++ {
		vector.DrawFilledRect(screen, float32(startX+i*40), 30, 35, 50, color.RGBA{150, 50, 50, 255}, false)
	}

	// Adversários Laterais (P1 e P3) - Simplificado
	vector.DrawFilledRect(screen, 30, float32(screenHeight/2-40), 40, 60, color.RGBA{50, 50, 150, 255}, false)
	vector.DrawFilledRect(screen, float32(screenWidth-70), float32(screenHeight/2-40), 40, 60, color.RGBA{50, 50, 150, 255}, false)
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
	logic := truco.NewGame(4)
	
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Truco Ebitengine 10/10")
	
	g := &Game{
		gameLogic: logic,
		snapshot:  logic.Snapshot(),
		message:   "Bem-vindo ao Truco!",
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
