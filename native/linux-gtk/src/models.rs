use serde::Deserialize;

// Top-level bundle returned by FFI
#[derive(Deserialize, Debug, Clone)]
pub struct SnapshotBundle {
    pub mode: Option<String>,
    pub locale: Option<String>,
    pub match_snapshot: Option<GameSnapshot>,
    // Serde rename for the "match" key (reserved word in Rust)
    pub lobby: Option<LobbySnapshot>,
    pub connection: Option<ConnectionSnapshot>,
    pub diagnostics: Option<DiagnosticsSnapshot>,
}

// Custom deserializer to handle "match" as a key name
impl SnapshotBundle {
    pub fn from_json(json: &str) -> Option<Self> {
        let v: serde_json::Value = serde_json::from_str(json).ok()?;
        let mode = v.get("mode").and_then(|m| m.as_str()).map(String::from);
        let locale = v.get("locale").and_then(|l| l.as_str()).map(String::from);
        let match_snapshot = v.get("match").and_then(|m| {
            serde_json::from_value::<GameSnapshot>(m.clone()).ok()
        });
        let lobby = v.get("lobby").and_then(|value| serde_json::from_value::<LobbySnapshot>(value.clone()).ok());
        let connection = v.get("connection").and_then(|value| serde_json::from_value::<ConnectionSnapshot>(value.clone()).ok());
        let diagnostics = v.get("diagnostics").and_then(|value| serde_json::from_value::<DiagnosticsSnapshot>(value.clone()).ok());
        
        Some(SnapshotBundle {
            mode,
            locale,
            match_snapshot,
            lobby,
            connection,
            diagnostics,
        })
    }
}

#[derive(Deserialize, Debug, Clone)]
pub struct AppEvent {
    pub kind: String,
    pub sequence: i64,
    pub timestamp: String,
    pub payload: Option<serde_json::Value>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct LobbySnapshot {
    #[serde(rename = "invite_key")]
    pub invite_key: Option<String>,
    #[serde(rename = "slots")]
    pub slots: Option<Vec<String>>,
    #[serde(rename = "assigned_seat")]
    pub assigned_seat: Option<i32>,
    #[serde(rename = "num_players")]
    pub num_players: Option<i32>,
    #[serde(rename = "started")]
    pub started: Option<bool>,
    #[serde(rename = "host_seat")]
    pub host_seat: Option<i32>,
    #[serde(rename = "connected_seats")]
    pub connected_seats: Option<std::collections::HashMap<String, bool>>,
    #[serde(rename = "role")]
    pub role: Option<String>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ConnectionSnapshot {
    #[serde(rename = "status")]
    pub status: Option<String>,
    #[serde(rename = "is_online")]
    pub is_online: Option<bool>,
    #[serde(rename = "is_host")]
    pub is_host: Option<bool>,
    #[serde(rename = "network")]
    pub network: Option<NetworkSnapshot>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct NetworkSnapshot {
    #[serde(rename = "transport")]
    pub transport: Option<String>,
    #[serde(rename = "supported_protocol_versions")]
    pub supported_protocol_versions: Option<Vec<i32>>,
    #[serde(rename = "negotiated_protocol_version")]
    pub negotiated_protocol_version: Option<i32>,
    #[serde(rename = "seat_protocol_versions")]
    pub seat_protocol_versions: Option<std::collections::HashMap<String, i32>>,
    #[serde(rename = "mixed_protocol_session")]
    pub mixed_protocol_session: Option<bool>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct DiagnosticsSnapshot {
    #[serde(rename = "event_backlog")]
    pub event_backlog: Option<i32>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct GameSnapshot {
    #[serde(rename = "NumPlayers")]
    pub num_players: Option<i32>,
    #[serde(rename = "MatchPoints")]
    pub match_points: Option<std::collections::HashMap<String, i32>>,
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

#[derive(Deserialize, Debug, Clone)]
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
    pub trick_wins: Option<std::collections::HashMap<String, i32>>,
    #[serde(rename = "TrickResults")]
    pub trick_results: Option<Vec<i32>>,
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

#[derive(Deserialize, Debug, Clone)]
pub struct PlayedCard {
    #[serde(rename = "PlayerID")]
    pub player_id: i32,
    #[serde(rename = "Card")]
    pub card: Card,
}

#[derive(Deserialize, Debug, Clone)]
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
