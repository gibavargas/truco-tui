package netp2p

import (
	"encoding/json"
	"errors"
	"unicode/utf8"

	"truco-tui/internal/truco"
)

const (
	maxOutboundSnapshotLogs = 80
	maxOutboundLogRuneLen   = 180
)

func cloneSnapshot(in truco.Snapshot) truco.Snapshot {
	out := in

	out.Players = append([]truco.Player(nil), in.Players...)
	for i := range out.Players {
		if in.Players[i].Hand != nil {
			out.Players[i].Hand = append([]truco.Card(nil), in.Players[i].Hand...)
		}
	}

	out.CurrentHand.RoundCards = append([]truco.PlayedCard(nil), in.CurrentHand.RoundCards...)
	out.CurrentHand.TrickResults = append([]int(nil), in.CurrentHand.TrickResults...)
	if in.CurrentHand.TrickWins != nil {
		out.CurrentHand.TrickWins = make(map[int]int, len(in.CurrentHand.TrickWins))
		for k, v := range in.CurrentHand.TrickWins {
			out.CurrentHand.TrickWins[k] = v
		}
	}
	out.LastTrickCards = append([]truco.PlayedCard(nil), in.LastTrickCards...)
	out.TrickPiles = make([]truco.TrickPile, len(in.TrickPiles))
	for i, pile := range in.TrickPiles {
		out.TrickPiles[i] = pile
		out.TrickPiles[i].Cards = append([]truco.PlayedCard(nil), pile.Cards...)
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

func truncateRunes(in string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(in) <= max {
		return in
	}
	out := make([]rune, 0, max)
	for _, r := range in {
		out = append(out, r)
		if len(out) >= max {
			break
		}
	}
	return string(out)
}

func snapshotWireSize(s truco.Snapshot) int {
	msg := Message{Type: "game_state", ProtocolVersion: protocolVersion, State: &s}
	b, err := json.Marshal(msg)
	if err != nil {
		return scannerMaxBuffer + 1
	}
	return len(b) + 1 // '\n' adicionado em writeMessage
}

func trimSnapshotForWire(in truco.Snapshot, maxPayload int) truco.Snapshot {
	snap := cloneSnapshot(in)

	if len(snap.Logs) > maxOutboundSnapshotLogs {
		snap.Logs = append([]string(nil), snap.Logs[len(snap.Logs)-maxOutboundSnapshotLogs:]...)
	}
	for i := range snap.Logs {
		snap.Logs[i] = truncateRunes(snap.Logs[i], maxOutboundLogRuneLen)
	}
	if snapshotWireSize(snap) <= maxPayload {
		return snap
	}

	for len(snap.Logs) > 0 && snapshotWireSize(snap) > maxPayload {
		snap.Logs = snap.Logs[1:]
	}
	if snapshotWireSize(snap) <= maxPayload {
		return snap
	}

	// Fallback: remove histórico textual e detalhes transitórios de rodada.
	snap.Logs = nil
	snap.CurrentHand.RoundCards = nil
	snap.CurrentHand.TrickResults = nil
	snap.CurrentHand.TrickWins = nil
	return snap
}

func trimSnapshotForFailover(in truco.Snapshot) truco.Snapshot {
	snap := cloneSnapshot(in)
	// Estado autoritativo precisa das mãos e da rodada; logs ficam no snapshot visual.
	snap.Logs = nil
	snap.CurrentPlayerIdx = -1
	return snap
}

func rotateSeat(idx, pivot, n int) int {
	if n <= 0 {
		return idx
	}
	return (idx - pivot + n) % n
}

func unrotateSeat(idx, pivot, n int) int {
	if n <= 0 {
		return idx
	}
	return (idx + pivot) % n
}

func RotateSeatSlice(in []string, pivot int) []string {
	n := len(in)
	out := make([]string, n)
	for newSeat := 0; newSeat < n; newSeat++ {
		oldSeat := unrotateSeat(newSeat, pivot, n)
		out[newSeat] = in[oldSeat]
	}
	return out
}

func RotateSeatMapString(in map[int]string, pivot, n int) map[int]string {
	out := make(map[int]string, len(in))
	for oldSeat, value := range in {
		if oldSeat < 0 || oldSeat >= n {
			continue
		}
		out[rotateSeat(oldSeat, pivot, n)] = value
	}
	return out
}

func RotateFailoverSnapshot(in truco.Snapshot, pivot int) (truco.Snapshot, error) {
	n := in.NumPlayers
	if n != 2 && n != 4 {
		return truco.Snapshot{}, errors.New("snapshot inválido para rotação")
	}
	if pivot < 0 || pivot >= n {
		return truco.Snapshot{}, errors.New("pivot inválido para rotação")
	}
	out := cloneSnapshot(in)
	out.Players = make([]truco.Player, n)
	for newSeat := 0; newSeat < n; newSeat++ {
		oldSeat := unrotateSeat(newSeat, pivot, n)
		out.Players[newSeat] = in.Players[oldSeat]
		out.Players[newSeat].ID = newSeat
	}

	remapPlayer := func(pid int) int {
		if pid < 0 || pid >= n {
			return pid
		}
		return rotateSeat(pid, pivot, n)
	}
	out.CurrentHand.Dealer = remapPlayer(in.CurrentHand.Dealer)
	out.CurrentHand.Turn = remapPlayer(in.CurrentHand.Turn)
	out.CurrentHand.RoundStart = remapPlayer(in.CurrentHand.RoundStart)
	out.CurrentHand.RaiseRequester = remapPlayer(in.CurrentHand.RaiseRequester)
	out.TurnPlayer = remapPlayer(in.TurnPlayer)

	out.CurrentHand.RoundCards = make([]truco.PlayedCard, len(in.CurrentHand.RoundCards))
	for i, rc := range in.CurrentHand.RoundCards {
		out.CurrentHand.RoundCards[i] = rc
		out.CurrentHand.RoundCards[i].PlayerID = remapPlayer(rc.PlayerID)
	}
	out.LastTrickCards = make([]truco.PlayedCard, len(in.LastTrickCards))
	for i, rc := range in.LastTrickCards {
		out.LastTrickCards[i] = rc
		out.LastTrickCards[i].PlayerID = remapPlayer(rc.PlayerID)
	}
	out.TrickPiles = make([]truco.TrickPile, len(in.TrickPiles))
	for i, pile := range in.TrickPiles {
		out.TrickPiles[i] = pile
		out.TrickPiles[i].Winner = remapPlayer(pile.Winner)
		out.TrickPiles[i].Cards = make([]truco.PlayedCard, len(pile.Cards))
		for j, rc := range pile.Cards {
			out.TrickPiles[i].Cards[j] = rc
			out.TrickPiles[i].Cards[j].PlayerID = remapPlayer(rc.PlayerID)
		}
	}
	if in.LastTrickWinner >= 0 {
		out.LastTrickWinner = remapPlayer(in.LastTrickWinner)
	}
	if in.CurrentPlayerIdx >= 0 {
		out.CurrentPlayerIdx = rotateSeat(in.CurrentPlayerIdx, pivot, n)
	}
	return out, nil
}
