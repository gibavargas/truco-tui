package truco

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	randv2 "math/rand/v2"
	"sort"
	"sync"
)

const (
	TargetPoints = 12
)

// Player representa um jogador humano ou CPU.
type Player struct {
	ID             int
	Name           string
	CPU            bool
	ProvisionalCPU bool
	Team           int
	Hand           []Card
	Score          int
}

// PlayedCard mantém o histórico da vazA atual.
type PlayedCard struct {
	PlayerID int
	Card     Card
}

// HandState mantém o estado da mão atual.
type HandState struct {
	Vira            Card
	Manilha         Rank
	Stake           int
	TrucoByTeam     int // -1 quando não existe pedido pendente
	RaiseRequester  int // jogador que pediu truco; -1 quando não há
	Dealer          int
	Turn            int
	Round           int
	RoundStart      int
	RoundCards      []PlayedCard
	TrickResults    []int // team id ou -1 para empate
	TrickWins       map[int]int
	WinnerTeam      int
	Finished        bool
	PendingRaiseFor int // time que precisa responder, -1 quando não há
}

// Snapshot fornece uma visão pronta para renderização na TUI.
type Snapshot struct {
	Players          []Player
	NumPlayers       int
	CurrentHand      HandState
	MatchPoints      map[int]int
	TurnPlayer       int
	CurrentTeamTurn  int
	Logs             []string
	WinnerTeam       int
	MatchFinished    bool
	CanAskTruco      bool
	PendingRaiseFor  int
	PendingRaiseBy   int
	PendingRaiseTo   int
	CurrentPlayerIdx int
	LastTrickSeq     int
	LastTrickTeam    int
	LastTrickWinner  int
	LastTrickTie     bool
	LastTrickRound   int
}

// Game concentra regras e sincronização para partida local/offline.
type Game struct {
	mu sync.Mutex

	rng        *randv2.Rand
	players    []Player
	numPlayers int
	points     map[int]int

	hand HandState
	logs []string

	winnerTeam int
	ended      bool

	lastTrickSeq    int
	lastTrickTeam   int
	lastTrickWinner int
	lastTrickTie    bool
	lastTrickRound  int
}

func NewGame(playerNames []string, cpuFlags []bool) (*Game, error) {
	if len(playerNames) != 2 && len(playerNames) != 4 {
		return nil, fmt.Errorf("quantidade de jogadores inválida: %d", len(playerNames))
	}
	if len(playerNames) != len(cpuFlags) {
		return nil, errors.New("playerNames e cpuFlags devem ter o mesmo tamanho")
	}

	rng, err := newSecureRNG()
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar RNG seguro: %w", err)
	}

	g := &Game{
		rng:             rng,
		numPlayers:      len(playerNames),
		points:          map[int]int{0: 0, 1: 0},
		winnerTeam:      -1,
		lastTrickTeam:   -1,
		lastTrickWinner: -1,
	}

	for i := range playerNames {
		team := i % 2
		g.players = append(g.players, Player{
			ID:             i,
			Name:           playerNames[i],
			CPU:            cpuFlags[i],
			ProvisionalCPU: false,
			Team:           team,
		})
	}
	g.startHandLocked(0)
	return g, nil
}

func NewGameFromSnapshot(s Snapshot) (*Game, error) {
	if s.NumPlayers != 2 && s.NumPlayers != 4 {
		return nil, fmt.Errorf("quantidade de jogadores inválida: %d", s.NumPlayers)
	}
	if len(s.Players) != s.NumPlayers {
		return nil, errors.New("snapshot inválido: jogadores inconsistentes")
	}
	rng, err := newSecureRNG()
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar RNG seguro: %w", err)
	}
	g := &Game{
		rng:             rng,
		numPlayers:      s.NumPlayers,
		players:         make([]Player, len(s.Players)),
		points:          map[int]int{0: 0, 1: 0},
		winnerTeam:      s.WinnerTeam,
		ended:           s.MatchFinished,
		lastTrickSeq:    s.LastTrickSeq,
		lastTrickTeam:   s.LastTrickTeam,
		lastTrickWinner: s.LastTrickWinner,
		lastTrickTie:    s.LastTrickTie,
		lastTrickRound:  s.LastTrickRound,
	}
	copy(g.players, s.Players)
	for i := range g.players {
		g.players[i].ID = i
		g.players[i].Hand = append([]Card(nil), s.Players[i].Hand...)
	}
	g.hand = s.CurrentHand
	g.hand.RoundCards = append([]PlayedCard(nil), s.CurrentHand.RoundCards...)
	g.hand.TrickResults = append([]int(nil), s.CurrentHand.TrickResults...)
	if s.CurrentHand.TrickWins != nil {
		g.hand.TrickWins = make(map[int]int, len(s.CurrentHand.TrickWins))
		for k, v := range s.CurrentHand.TrickWins {
			g.hand.TrickWins[k] = v
		}
	}
	g.logs = append([]string(nil), s.Logs...)
	if s.MatchPoints != nil {
		g.points[0] = s.MatchPoints[0]
		g.points[1] = s.MatchPoints[1]
	}
	return g, nil
}

func newSecureRNG() (*randv2.Rand, error) {
	var seed [16]byte
	if _, err := cryptorand.Read(seed[:]); err != nil {
		return nil, err
	}
	lo := binary.LittleEndian.Uint64(seed[:8])
	hi := binary.LittleEndian.Uint64(seed[8:])
	return randv2.New(randv2.NewPCG(lo, hi)), nil
}

func (g *Game) addLogLocked(msg string) {
	g.logs = append(g.logs, msg)
	if len(g.logs) > 120 {
		g.logs = g.logs[len(g.logs)-120:]
	}
}

func (g *Game) StartNewHand() {
	g.mu.Lock()
	defer g.mu.Unlock()
	dealer := (g.hand.Dealer + 1) % g.numPlayers
	g.startHandLocked(dealer)
}

func (g *Game) startHandLocked(dealer int) {
	deck := NewDeck()
	deck.Shuffle(g.rng)

	for i := range g.players {
		g.players[i].Hand = g.players[i].Hand[:0]
	}
	for c := 0; c < 3; c++ {
		for i := 0; i < g.numPlayers; i++ {
			card, ok := deck.Draw()
			if !ok {
				panic("baralho sem cartas durante distribuição")
			}
			g.players[i].Hand = append(g.players[i].Hand, card)
		}
	}
	for i := range g.players {
		sort.Slice(g.players[i].Hand, func(a, b int) bool {
			return normalPower[g.players[i].Hand[a].Rank] > normalPower[g.players[i].Hand[b].Rank]
		})
	}

	vira, ok := deck.Draw()
	if !ok {
		panic("não foi possível revelar a vira")
	}

	start := (dealer + 1) % g.numPlayers
	g.hand = HandState{
		Vira:            vira,
		Manilha:         NextRank(vira.Rank),
		Stake:           1,
		TrucoByTeam:     -1,
		RaiseRequester:  -1,
		Dealer:          dealer,
		Turn:            start,
		Round:           1,
		RoundStart:      start,
		RoundCards:      []PlayedCard{},
		TrickResults:    []int{},
		TrickWins:       map[int]int{0: 0, 1: 0},
		WinnerTeam:      -1,
		Finished:        false,
		PendingRaiseFor: -1,
	}
	g.addLogLocked(fmt.Sprintf("Nova mão: vira %s, manilha %s.", g.hand.Vira, g.hand.Manilha))
}

func (g *Game) nextSeatLocked(current int) int {
	return (current + 1) % g.numPlayers
}

func (g *Game) nextSeatByTeamLocked(fromPlayerID, team int) int {
	seat := fromPlayerID
	for i := 0; i < g.numPlayers; i++ {
		seat = g.nextSeatLocked(seat)
		pi := g.playerIndexByIDLocked(seat)
		if pi >= 0 && g.players[pi].Team == team {
			return seat
		}
	}
	return fromPlayerID
}

func (g *Game) playerIndexByIDLocked(id int) int {
	for i := range g.players {
		if g.players[i].ID == id {
			return i
		}
	}
	return -1
}

func (g *Game) PlayCard(playerID, handIndex int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.ended {
		return errors.New("partida já terminou")
	}
	if g.hand.Finished {
		return errors.New("mão já finalizada")
	}
	if g.hand.PendingRaiseFor != -1 {
		return errors.New("existe um pedido de truco pendente")
	}
	if g.hand.Turn != playerID {
		return errors.New("não é a vez do jogador")
	}

	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return errors.New("jogador inválido")
	}
	if handIndex < 0 || handIndex >= len(g.players[pi].Hand) {
		return errors.New("índice de carta inválido")
	}

	card := g.players[pi].Hand[handIndex]
	g.players[pi].Hand = append(g.players[pi].Hand[:handIndex], g.players[pi].Hand[handIndex+1:]...)
	g.hand.RoundCards = append(g.hand.RoundCards, PlayedCard{PlayerID: playerID, Card: card})
	g.addLogLocked(fmt.Sprintf("%s jogou %s.", g.players[pi].Name, card))

	if len(g.hand.RoundCards) < g.numPlayers {
		g.hand.Turn = g.nextSeatLocked(g.hand.Turn)
		return nil
	}

	g.resolveTrickLocked()
	g.checkHandEndLocked()

	if g.hand.Finished {
		g.finishHandLocked()
	}
	return nil
}

func (g *Game) resolveTrickLocked() {
	roundNum := g.hand.Round
	best := g.hand.RoundCards[0]
	isTie := false

	for i := 1; i < len(g.hand.RoundCards); i++ {
		cmp := CompareCards(g.hand.RoundCards[i].Card, best.Card, g.hand.Manilha)
		if cmp > 0 {
			best = g.hand.RoundCards[i]
			isTie = false
		} else if cmp == 0 {
			isTie = true
		}
	}

	result := -1
	if !isTie {
		winnerIdx := g.playerIndexByIDLocked(best.PlayerID)
		team := g.players[winnerIdx].Team
		g.hand.TrickWins[team]++
		result = team
		g.lastTrickTeam = team
		g.lastTrickWinner = best.PlayerID
		g.lastTrickTie = false
		g.lastTrickRound = roundNum
		g.lastTrickSeq++
		g.hand.RoundStart = best.PlayerID
		g.hand.Turn = best.PlayerID
		g.addLogLocked(fmt.Sprintf("Vaza %d: %s venceu.", g.hand.Round, g.players[winnerIdx].Name))
	} else {
		g.lastTrickTeam = -1
		g.lastTrickWinner = -1
		g.lastTrickTie = true
		g.lastTrickRound = roundNum
		g.lastTrickSeq++
		g.addLogLocked(fmt.Sprintf("Vaza %d empatou.", g.hand.Round))
		g.hand.Turn = g.hand.RoundStart
	}

	g.hand.TrickResults = append(g.hand.TrickResults, result)
	g.hand.Round++
	g.hand.RoundCards = g.hand.RoundCards[:0]
}

func (g *Game) checkHandEndLocked() {
	if g.hand.TrickWins[0] >= 2 {
		g.hand.WinnerTeam = 0
		g.hand.Finished = true
		return
	}
	if g.hand.TrickWins[1] >= 2 {
		g.hand.WinnerTeam = 1
		g.hand.Finished = true
		return
	}

	results := g.hand.TrickResults
	if len(results) == 2 {
		if results[0] == -1 && results[1] != -1 {
			g.hand.WinnerTeam = results[1]
			g.hand.Finished = true
			return
		}
		if results[0] != -1 && results[1] == -1 {
			g.hand.WinnerTeam = results[0]
			g.hand.Finished = true
			return
		}
	}

	if len(results) >= 3 {
		if g.hand.TrickWins[0] > g.hand.TrickWins[1] {
			g.hand.WinnerTeam = 0
		} else if g.hand.TrickWins[1] > g.hand.TrickWins[0] {
			g.hand.WinnerTeam = 1
		} else {
			firstNonTie := -1
			for _, r := range results {
				if r != -1 {
					firstNonTie = r
					break
				}
			}
			if firstNonTie != -1 {
				g.hand.WinnerTeam = firstNonTie
			} else {
				starterTeam := g.players[g.playerIndexByIDLocked(g.hand.RoundStart)].Team
				g.hand.WinnerTeam = starterTeam
			}
		}
		g.hand.Finished = true
	}
}

func (g *Game) finishHandLocked() {
	team := g.hand.WinnerTeam
	if team < 0 {
		return
	}
	g.points[team] += g.hand.Stake
	g.addLogLocked(fmt.Sprintf("Mão encerrada: time %d ganhou %d ponto(s). Placar: %d x %d.", team+1, g.hand.Stake, g.points[0], g.points[1]))

	if g.points[team] >= TargetPoints {
		g.ended = true
		g.winnerTeam = team
		g.addLogLocked(fmt.Sprintf("Partida encerrada: time %d venceu!", team+1))
		return
	}

	dealer := (g.hand.Dealer + 1) % g.numPlayers
	g.startHandLocked(dealer)
}

func (g *Game) AskTruco(playerID int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.ended || g.hand.Finished {
		return errors.New("não é possível pedir truco agora")
	}
	if g.hand.PendingRaiseFor != -1 {
		return errors.New("já existe pedido de truco pendente")
	}
	if g.hand.Turn != playerID {
		return errors.New("só o jogador da vez pode pedir truco")
	}
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return errors.New("jogador inválido")
	}
	team := g.players[pi].Team
	if team == g.hand.TrucoByTeam {
		return errors.New("seu time já aumentou a aposta por último")
	}
	if g.hand.Stake >= 12 {
		return errors.New("aposta já está no máximo")
	}

	opponentTeam := 1 - team
	g.hand.PendingRaiseFor = opponentTeam
	g.hand.TrucoByTeam = team
	g.hand.RaiseRequester = playerID
	g.hand.Turn = g.nextSeatByTeamLocked(playerID, opponentTeam)
	g.addLogLocked(fmt.Sprintf("%s pediu %s!", g.players[pi].Name, raiseLabel(nextStake(g.hand.Stake))))
	return nil
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

func (g *Game) RespondTruco(playerID int, accept bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.hand.PendingRaiseFor == -1 {
		return errors.New("não há pedido de truco pendente")
	}
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return errors.New("jogador inválido")
	}
	team := g.players[pi].Team
	if team != g.hand.PendingRaiseFor {
		return errors.New("seu time não deve responder o truco")
	}

	if accept {
		g.hand.Stake = nextStake(g.hand.Stake)
		g.addLogLocked(fmt.Sprintf("%s aceitou %s. Aposta agora vale %d.", g.players[pi].Name, raiseLabel(g.hand.Stake), g.hand.Stake))
		g.hand.PendingRaiseFor = -1
		if g.playerIndexByIDLocked(g.hand.RaiseRequester) >= 0 {
			g.hand.Turn = g.hand.RaiseRequester
		}
		g.hand.RaiseRequester = -1
		return nil
	}

	winner := 1 - g.hand.PendingRaiseFor
	g.addLogLocked(fmt.Sprintf("%s correu do truco.", g.players[pi].Name))
	g.hand.WinnerTeam = winner
	g.hand.Finished = true
	g.hand.PendingRaiseFor = -1
	g.hand.RaiseRequester = -1
	g.finishHandLocked()
	return nil
}

// RaiseTruco aceita o pedido pendente atual e já solicita o próximo degrau
// (ex.: pedido de 3 -> resposta sobe para 6).
func (g *Game) RaiseTruco(playerID int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.ended || g.hand.Finished {
		return errors.New("não é possível aumentar o truco agora")
	}
	if g.hand.PendingRaiseFor == -1 {
		return errors.New("não há pedido de truco pendente")
	}
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return errors.New("jogador inválido")
	}
	team := g.players[pi].Team
	if team != g.hand.PendingRaiseFor {
		return errors.New("seu time não deve responder o truco")
	}

	acceptedStake := nextStake(g.hand.Stake)
	if acceptedStake <= g.hand.Stake {
		return errors.New("pedido atual já está no máximo")
	}
	raisedStake := nextStake(acceptedStake)
	if raisedStake <= acceptedStake {
		return errors.New("aposta já está no máximo para novo aumento")
	}

	// Ao aumentar, o time aceita o valor atual e imediatamente pede o próximo.
	g.hand.Stake = acceptedStake
	opponentTeam := 1 - team
	g.hand.PendingRaiseFor = opponentTeam
	g.hand.TrucoByTeam = team
	g.hand.RaiseRequester = playerID
	g.hand.Turn = g.nextSeatByTeamLocked(playerID, opponentTeam)
	g.addLogLocked(fmt.Sprintf("%s aumentou para %s!", g.players[pi].Name, raiseLabel(raisedStake)))
	return nil
}

func (g *Game) Snapshot(forPlayer int) Snapshot {
	g.mu.Lock()
	defer g.mu.Unlock()

	players := make([]Player, len(g.players))
	copy(players, g.players)
	logs := make([]string, len(g.logs))
	copy(logs, g.logs)

	currTeam := -1
	currIdx := g.playerIndexByIDLocked(g.hand.Turn)
	if currIdx >= 0 {
		currTeam = g.players[currIdx].Team
	}

	pendingBy := -1
	pendingTo := 0
	if g.hand.PendingRaiseFor != -1 {
		pendingBy = 1 - g.hand.PendingRaiseFor
		pendingTo = nextStake(g.hand.Stake)
	}

	return Snapshot{
		Players:          players,
		NumPlayers:       g.numPlayers,
		CurrentHand:      g.hand,
		MatchPoints:      map[int]int{0: g.points[0], 1: g.points[1]},
		TurnPlayer:       g.hand.Turn,
		CurrentTeamTurn:  currTeam,
		Logs:             logs,
		WinnerTeam:       g.winnerTeam,
		MatchFinished:    g.ended,
		CanAskTruco:      g.hand.PendingRaiseFor == -1 && g.hand.Stake < 12,
		PendingRaiseFor:  g.hand.PendingRaiseFor,
		PendingRaiseBy:   pendingBy,
		PendingRaiseTo:   pendingTo,
		CurrentPlayerIdx: g.playerIndexByIDLocked(forPlayer),
		LastTrickSeq:     g.lastTrickSeq,
		LastTrickTeam:    g.lastTrickTeam,
		LastTrickWinner:  g.lastTrickWinner,
		LastTrickTie:     g.lastTrickTie,
		LastTrickRound:   g.lastTrickRound,
	}
}

func (g *Game) IsCPUTurn() (bool, int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(g.hand.Turn)
	if pi < 0 {
		return false, -1
	}
	return g.players[pi].CPU, g.players[pi].ID
}

func (g *Game) TeamOfPlayer(playerID int) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return -1
	}
	return g.players[pi].Team
}

// CanAskTrucoByPlayer indica se o jogador pode pedir truco no estado atual.
// Essa checagem é usada pela CPU para evitar loops de tentativa inválida.
func (g *Game) CanAskTrucoByPlayer(playerID int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.ended || g.hand.Finished {
		return false
	}
	if g.hand.PendingRaiseFor != -1 {
		return false
	}
	if g.hand.Turn != playerID {
		return false
	}
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return false
	}
	team := g.players[pi].Team
	if team == g.hand.TrucoByTeam {
		return false
	}
	if g.hand.Stake >= 12 {
		return false
	}
	return true
}

func (g *Game) PendingTeamToRespond() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.hand.PendingRaiseFor
}

func (g *Game) PlayerName(playerID int) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return ""
	}
	return g.players[pi].Name
}

func (g *Game) HandCards(playerID int) []Card {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return nil
	}
	out := make([]Card, len(g.players[pi].Hand))
	copy(out, g.players[pi].Hand)
	return out
}

func (g *Game) SetPlayerCPU(playerID int, cpu bool, provisional bool) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return false
	}
	changed := g.players[pi].CPU != cpu || g.players[pi].ProvisionalCPU != provisional
	g.players[pi].CPU = cpu
	g.players[pi].ProvisionalCPU = provisional
	return changed
}

func (g *Game) PlayerCPU(playerID int) (bool, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	pi := g.playerIndexByIDLocked(playerID)
	if pi < 0 {
		return false, false
	}
	return g.players[pi].CPU, g.players[pi].ProvisionalCPU
}

func (g *Game) MatchEnded() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.ended
}
