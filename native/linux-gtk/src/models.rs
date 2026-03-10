use serde::Deserialize;

#[derive(Deserialize, Debug, Clone)]
pub struct GameSnapshot {
    pub mode: Option<String>,
    #[serde(rename = "NumPlayers")]
    pub num_players: Option<i32>,
    #[serde(rename = "MatchPoints")]
    pub match_points: Option<Vec<i32>>,
    #[serde(rename = "TurnPlayer")]
    pub turn_player: Option<i32>,
    #[serde(rename = "CurrentHand")]
    pub current_hand: Option<HandState>,
    #[serde(rename = "Players")]
    pub players: Option<Vec<Player>>,
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
