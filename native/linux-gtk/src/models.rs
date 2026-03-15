use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Deserialize, Debug, Clone, Default)]
pub struct SnapshotBundle {
    pub versions: Option<CoreVersions>,
    pub mode: Option<String>,
    pub locale: Option<String>,
    #[serde(rename = "match")]
    pub game: Option<GameSnapshot>,
    pub lobby: Option<LobbySnapshot>,
    pub ui: Option<UIStateSnapshot>,
    pub connection: Option<ConnectionSnapshot>,
    pub diagnostics: Option<DiagnosticsSnapshot>,
}

impl SnapshotBundle {
    pub fn from_json(json: &str) -> Option<Self> {
        serde_json::from_str(json).ok()
    }
}

#[derive(Deserialize, Debug, Clone, Default, PartialEq, Eq)]
pub struct CoreVersions {
    #[serde(rename = "core_api_version")]
    pub core_api_version: i32,
    #[serde(rename = "protocol_version")]
    pub protocol_version: i32,
    #[serde(rename = "snapshot_schema_version")]
    pub snapshot_schema_version: i32,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct ConnectionSnapshot {
    pub status: Option<String>,
    #[serde(rename = "is_online")]
    pub is_online: Option<bool>,
    #[serde(rename = "is_host")]
    pub is_host: Option<bool>,
    #[serde(rename = "last_error")]
    pub last_error: Option<AppError>,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct DiagnosticsSnapshot {
    #[serde(rename = "event_backlog")]
    pub event_backlog: Option<i32>,
    #[serde(rename = "event_log")]
    pub event_log: Option<Vec<String>>,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct AppError {
    pub code: Option<String>,
    pub message: Option<String>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct UIStateSnapshot {
    #[serde(rename = "lobby_slots")]
    pub lobby_slots: Option<Vec<LobbySlotState>>,
    #[serde(rename = "actions")]
    pub actions: Option<ActionSnapshot>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct LobbySlotState {
    #[serde(rename = "seat")]
    pub seat: i32,
    #[serde(rename = "name")]
    pub name: Option<String>,
    #[serde(rename = "status")]
    pub status: Option<String>,
    #[serde(rename = "is_empty")]
    pub is_empty: bool,
    #[serde(rename = "is_local")]
    pub is_local: bool,
    #[serde(rename = "is_host")]
    pub is_host: bool,
    #[serde(rename = "is_connected")]
    pub is_connected: bool,
    #[serde(rename = "is_occupied")]
    pub is_occupied: bool,
    #[serde(rename = "is_provisional_cpu")]
    pub is_provisional_cpu: bool,
    #[serde(rename = "can_vote_host")]
    pub can_vote_host: bool,
    #[serde(rename = "can_request_replacement")]
    pub can_request_replacement: bool,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ActionSnapshot {
    #[serde(rename = "local_player_id")]
    pub local_player_id: i32,
    #[serde(rename = "local_team")]
    pub local_team: i32,
    #[serde(rename = "can_play_card")]
    pub can_play_card: bool,
    #[serde(rename = "can_ask_or_raise")]
    pub can_ask_or_raise: bool,
    #[serde(rename = "must_respond")]
    pub must_respond: bool,
    #[serde(rename = "can_accept")]
    pub can_accept: bool,
    #[serde(rename = "can_refuse")]
    pub can_refuse: bool,
    #[serde(rename = "can_close_session")]
    pub can_close_session: bool,
}

#[derive(Deserialize, Debug, Clone)]
pub struct AppEvent {
    pub kind: String,
    pub sequence: i64,
    pub timestamp: String,
    pub payload: Option<Value>,
}

impl AppEvent {
    pub fn text(&self) -> Option<String> {
        match self.kind.as_str() {
            "chat" => {
                let author = self
                    .payload
                    .as_ref()
                    .and_then(|p| p.get("author"))
                    .and_then(|a| a.as_str())
                    .unwrap_or("?");
                let msg = self
                    .payload
                    .as_ref()
                    .and_then(|p| p.get("text"))
                    .and_then(|t| t.as_str())
                    .unwrap_or("");
                Some(format!("{author}: {msg}"))
            }
            "system" | "error" => self
                .payload
                .as_ref()
                .and_then(|p| p.get("text").or_else(|| p.get("message")))
                .and_then(|t| t.as_str())
                .map(str::to_string),
            "replacement_invite" => self
                .payload
                .as_ref()
                .and_then(|p| p.get("invite_key"))
                .and_then(|t| t.as_str())
                .map(|key| format!("Link de substituição: {key}")),
            _ => None,
        }
    }
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct LobbySnapshot {
    #[serde(rename = "invite_key")]
    pub invite_key: Option<String>,
    pub slots: Option<Vec<String>>,
    #[serde(rename = "assigned_seat")]
    pub assigned_seat: Option<i32>,
    #[serde(rename = "num_players")]
    pub num_players: Option<i32>,
    pub started: Option<bool>,
    #[serde(rename = "host_seat")]
    pub host_seat: Option<i32>,
    #[serde(rename = "connected_seats")]
    pub connected_seats: Option<HashMap<String, bool>>,
    pub role: Option<String>,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct GameSnapshot {
    #[serde(rename = "NumPlayers")]
    pub num_players: Option<i32>,
    #[serde(rename = "MatchPoints")]
    pub match_points: Option<HashMap<String, i32>>,
    #[serde(rename = "TurnPlayer")]
    pub turn_player: Option<i32>,
    #[serde(rename = "CurrentTeamTurn")]
    pub current_team_turn: Option<i32>,
    #[serde(rename = "CurrentHand")]
    pub current_hand: Option<HandState>,
    #[serde(rename = "Players")]
    pub players: Option<Vec<Player>>,
    #[serde(rename = "Logs")]
    pub logs: Option<Vec<String>>,
    #[serde(rename = "MatchFinished")]
    pub match_finished: Option<bool>,
    #[serde(rename = "WinnerTeam")]
    pub winner_team: Option<i32>,
    #[serde(rename = "CanAskTruco")]
    pub can_ask_truco: Option<bool>,
    #[serde(rename = "PendingRaiseFor")]
    pub pending_raise_for: Option<i32>,
    #[serde(rename = "PendingRaiseBy")]
    pub pending_raise_by: Option<i32>,
    #[serde(rename = "PendingRaiseTo")]
    pub pending_raise_to: Option<i32>,
    #[serde(rename = "CurrentPlayerIdx")]
    pub current_player_idx: Option<i32>,
    #[serde(rename = "LastTrickSeq")]
    pub last_trick_seq: Option<i32>,
    #[serde(rename = "LastTrickWinner")]
    pub last_trick_winner: Option<i32>,
    #[serde(rename = "LastTrickTeam")]
    pub last_trick_team: Option<i32>,
    #[serde(rename = "LastTrickTie")]
    pub last_trick_tie: Option<bool>,
    #[serde(rename = "LastTrickRound")]
    pub last_trick_round: Option<i32>,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct HandState {
    #[serde(rename = "Stake")]
    pub stake: Option<i32>,
    #[serde(rename = "Vira")]
    pub vira: Option<Card>,
    #[serde(rename = "Manilha")]
    pub manilha: Option<String>,
    #[serde(rename = "RoundCards")]
    pub round_cards: Option<Vec<PlayedCard>>,
    #[serde(rename = "Round")]
    pub round: Option<i32>,
    #[serde(rename = "Dealer")]
    pub dealer: Option<i32>,
    #[serde(rename = "Turn")]
    pub turn: Option<i32>,
    #[serde(rename = "TrucoByTeam")]
    pub truco_by_team: Option<i32>,
    #[serde(rename = "RaiseRequester")]
    pub raise_requester: Option<i32>,
    #[serde(rename = "WinnerTeam")]
    pub winner_team: Option<i32>,
    #[serde(rename = "Finished")]
    pub finished: Option<bool>,
    #[serde(rename = "TrickWins")]
    pub trick_wins: Option<HashMap<String, i32>>,
    #[serde(rename = "TrickResults")]
    pub trick_results: Option<Vec<i32>>,
}

impl HandState {
    pub fn winning_card_id(&self) -> Option<String> {
        let cards = self.round_cards.as_ref()?;
        if cards.is_empty() {
            return None;
        }

        let mut best_id: Option<String> = None;
        let mut best_power = -1;
        let mut is_tie = false;
        let manilha_ref = self.manilha.as_deref();

        for pc in cards {
            let power = pc.card.power(manilha_ref);
            if power > best_power {
                best_power = power;
                best_id = Some(format!("{}-{}-{}", pc.player_id, pc.card.rank, pc.card.suit));
                is_tie = false;
            } else if power == best_power {
                is_tie = true;
            }
        }

        if is_tie {
            None
        } else {
            best_id
        }
    }
}

#[derive(Deserialize, Debug, Clone)]
pub struct Player {
    #[serde(rename = "ID")]
    pub id: i32,
    #[serde(rename = "Name")]
    pub name: String,
    #[serde(rename = "Team")]
    pub team: i32,
    #[serde(rename = "Hand")]
    pub hand: Option<Vec<Card>>,
    #[serde(rename = "CPU")]
    pub cpu: Option<bool>,
    #[serde(rename = "ProvisionalCPU")]
    pub provisional_cpu: Option<bool>,
}

#[derive(Deserialize, Debug, Clone, Default)]
pub struct PlayedCard {
    #[serde(rename = "PlayerID")]
    pub player_id: i32,
    #[serde(rename = "Card")]
    pub card: Card,
}

#[derive(Deserialize, Serialize, Debug, Clone, Default)]
pub struct Card {
    #[serde(rename = "Rank")]
    pub rank: String,
    #[serde(rename = "Suit")]
    pub suit: String,
}

impl Card {
    pub fn suit_symbol(&self) -> &'static str {
        match self.suit.as_str() {
            "Espadas" => "♠",
            "Copas" => "♥",
            "Ouros" => "♦",
            "Paus" => "♣",
            _ => "",
        }
    }

    pub fn is_red(&self) -> bool {
        self.suit == "Copas" || self.suit == "Ouros"
    }

    pub fn power(&self, manilha: Option<&str>) -> i32 {
        let normal_power = match self.rank.as_str() {
            "3" => 10,
            "2" => 9,
            "A" => 8,
            "K" => 7,
            "J" => 6,
            "Q" => 5,
            "7" => 4,
            "6" => 3,
            "5" => 2,
            "4" => 1,
            _ => 0,
        };

        if let Some(target) = manilha {
            if self.rank == target {
                let suit_power = match self.suit.as_str() {
                    "Paus" => 4,
                    "Copas" => 3,
                    "Espadas" => 2,
                    "Ouros" => 1,
                    _ => 0,
                };
                return 100 + suit_power;
            }
        }

        normal_power
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parses_snapshot_bundle_with_versions() {
        let input = r#"{
          "versions":{"core_api_version":1,"protocol_version":2,"snapshot_schema_version":1},
          "mode":"idle",
          "locale":"pt-BR",
          "connection":{"status":"idle","is_online":false,"is_host":false},
          "diagnostics":{"event_backlog":0}
        }"#;
        let bundle = SnapshotBundle::from_json(input).expect("bundle");
        let versions = bundle.versions.expect("versions");
        assert_eq!(versions.core_api_version, 1);
        assert_eq!(versions.snapshot_schema_version, 1);
    }
}
