package ui

import "truco-tui/internal/truco"

// requestOrRaiseTruco aplica o atalho [t]:
// - se o time do jogador está respondendo um truco pendente, sobe a aposta;
// - caso contrário, inicia um novo pedido de truco.
func requestOrRaiseTruco(g *truco.Game, playerID int) error {
	pendingTeam := g.PendingTeamToRespond()
	if pendingTeam != -1 && g.TeamOfPlayer(playerID) == pendingTeam {
		return g.RaiseTruco(playerID)
	}
	return g.AskTruco(playerID)
}
