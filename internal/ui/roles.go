package ui

import (
	"strings"

	"truco-tui/internal/truco"
)

// SeatRoleInfo is a UI-level role contract derived from a game snapshot.
// It is intentionally generic so TUI and browser views can map it to copy/chips.
type SeatRoleInfo struct {
	TeamRelation     string
	TurnRole         string
	TrickRole        string
	ConnectivityRole string
	GovernanceRole   string
}

func deriveSeatRoles(s truco.Snapshot, localIdx int, isOnline bool) map[int]SeatRoleInfo {
	out := make(map[int]SeatRoleInfo, len(s.Players))
	if len(s.Players) == 0 {
		return out
	}

	if localIdx < 0 || localIdx >= len(s.Players) {
		localIdx = 0
	}
	localPlayerID := s.Players[localIdx].ID
	localTeam := s.Players[localIdx].Team
	roundStartIdx := playerIndexByID(s.Players, s.CurrentHand.RoundStart)
	if roundStartIdx < 0 {
		roundStartIdx = 0
	}
	n := len(s.Players)

	for _, p := range s.Players {
		info := SeatRoleInfo{
			TeamRelation: "opponent",
		}
		if p.ID == localPlayerID {
			info.TeamRelation = "self"
		} else if p.Team == localTeam {
			info.TeamRelation = "partner"
		}

		pi := playerIndexByID(s.Players, p.ID)
		if pi >= 0 && n > 0 {
			dist := (pi - roundStartIdx + n) % n
			if dist == 0 {
				info.TurnRole = "mao"
			}
			if dist == n-1 {
				info.TurnRole = joinRoles(info.TurnRole, "pe")
			}
		}

		if p.ID == s.CurrentHand.Turn {
			info.TrickRole = "turn"
		}

		if p.ProvisionalCPU {
			info.ConnectivityRole = "provisional_cpu"
		} else if p.CPU {
			info.ConnectivityRole = "cpu"
		}

		if isOnline && p.ID == 0 {
			info.GovernanceRole = "host"
		}

		out[p.ID] = info
	}
	return out
}

func joinRoles(existing, role string) string {
	if existing == "" {
		return role
	}
	return existing + "," + role
}

func roleBadgeKeys(info SeatRoleInfo) []string {
	keys := make([]string, 0, 6)
	switch info.TeamRelation {
	case "self":
		keys = append(keys, "ui_role_you")
	case "partner":
		keys = append(keys, "ui_role_partner")
	default:
		keys = append(keys, "ui_role_opponent")
	}

	if strings.Contains(info.TurnRole, "mao") {
		keys = append(keys, "ui_role_mao")
	}
	if strings.Contains(info.TurnRole, "pe") {
		keys = append(keys, "ui_role_pe")
	}
	if info.TrickRole == "turn" {
		keys = append(keys, "ui_role_turn")
	}
	if info.GovernanceRole == "host" {
		keys = append(keys, "ui_role_host")
	}
	if info.ConnectivityRole == "provisional_cpu" {
		keys = append(keys, "ui_role_cpu_prov")
	} else if info.ConnectivityRole == "cpu" {
		keys = append(keys, "ui_role_cpu")
	}
	return keys
}
