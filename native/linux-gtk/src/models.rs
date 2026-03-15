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

#[derive(Deserialize, Debug, Clone, Default)]
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
