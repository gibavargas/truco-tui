use serde::Serialize;
use serde_json::Value;

#[derive(Serialize)]
pub struct AppIntent<T>
where
    T: Serialize,
{
    pub kind: &'static str,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub payload: Option<T>,
}

impl<T> AppIntent<T>
where
    T: Serialize,
{
    pub fn with_payload(kind: &'static str, payload: T) -> Self {
        Self {
            kind,
            payload: Some(payload),
        }
    }
}

impl AppIntent<Value> {
    pub fn without_payload(kind: &'static str) -> Self {
        Self {
            kind,
            payload: None,
        }
    }
}

#[derive(Serialize)]
pub struct SetLocalePayload<'a> {
    pub locale: &'a str,
}

#[derive(Serialize)]
pub struct NewOfflineGamePayload {
    pub player_names: Vec<String>,
    pub cpu_flags: Vec<bool>,
}

#[derive(Serialize)]
pub struct CreateHostPayload<'a> {
    pub host_name: &'a str,
    pub num_players: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub relay_url: Option<&'a str>,
}

#[derive(Serialize)]
pub struct JoinSessionPayload<'a> {
    pub key: &'a str,
    pub player_name: &'a str,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub desired_role: Option<&'a str>,
}

#[derive(Serialize)]
pub struct GameActionPayload {
    pub action: &'static str,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub card_index: Option<usize>,
}

#[derive(Serialize)]
pub struct SendChatPayload<'a> {
    pub text: &'a str,
}

#[derive(Serialize)]
pub struct HostVotePayload {
    pub candidate_seat: usize,
}

#[derive(Serialize)]
pub struct ReplacementInvitePayload {
    pub target_seat: usize,
}

pub fn to_json<T>(intent: &AppIntent<T>) -> Option<String>
where
    T: Serialize,
{
    serde_json::to_string(intent).ok()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn serializes_offline_intent() {
        let intent = AppIntent::with_payload(
            "new_offline_game",
            NewOfflineGamePayload {
                player_names: vec!["Voce".into(), "CPU".into()],
                cpu_flags: vec![false, true],
            },
        );
        let json = to_json(&intent).expect("json");
        assert!(json.contains("\"kind\":\"new_offline_game\""));
        assert!(json.contains("\"player_names\":[\"Voce\",\"CPU\"]"));
    }

    #[test]
    fn serializes_empty_payload_intent() {
        let intent = AppIntent::without_payload("start_hosted_match");
        let json = to_json(&intent).expect("json");
        assert_eq!(json, "{\"kind\":\"start_hosted_match\"}");
    }
}
