import Foundation

// MARK: - Top-level SnapshotBundle (matches Go appcore.SnapshotBundle)
struct SnapshotBundle: Codable {
    let versions: CoreVersions?
    let mode: String
    let locale: String?
    let match: MatchSnapshot?
    let lobby: LobbySnapshot?
    let ui: UIStateSnapshot?
    let connection: ConnectionSnapshot?
    let diagnostics: DiagnosticsSnapshot?
}

struct CoreVersions: Codable {
    let core_api_version: Int?
    let protocol_version: Int?
    let snapshot_schema_version: Int?
}

struct LobbySnapshot: Codable {
    let invite_key: String?
    let slots: [String]?
    let assigned_seat: Int?
    let num_players: Int?
    let started: Bool?
    let host_seat: Int?
    let connected_seats: [String: Bool]?
    let role: String?
}

struct UIStateSnapshot: Codable {
    let lobby_slots: [LobbySlotState]?
    let actions: ActionSnapshot?
}

struct LobbySlotState: Codable, Identifiable {
    let seat: Int
    let name: String?
    let status: String
    let is_empty: Bool
    let is_local: Bool
    let is_host: Bool
    let is_connected: Bool
    let is_occupied: Bool
    let is_provisional_cpu: Bool
    let can_vote_host: Bool
    let can_request_replacement: Bool

    var id: Int { seat }
}

struct ActionSnapshot: Codable {
    let local_player_id: Int
    let local_team: Int
    let can_play_card: Bool
    let can_ask_or_raise: Bool
    let must_respond: Bool
    let can_accept: Bool
    let can_refuse: Bool
    let can_close_session: Bool
}

struct ConnectionSnapshot: Codable {
    let status: String?
    let is_online: Bool?
    let is_host: Bool?
}

struct DiagnosticsSnapshot: Codable {
    let event_backlog: Int?
}

// MARK: - AppEvents definition
struct AppEvent: Codable, Identifiable {
    let kind: String
    let sequence: Int64
    let timestamp: String
    let payload: EventPayload?

    var id: Int64 { sequence }
}

struct EventPayload: Codable {
    let text: String?
    let author: String?
    let target_seat: Int?
    let invite_key: String?
}

// MARK: - Match Snapshot (matches Go truco.Snapshot)
struct MatchSnapshot: Codable {
    let Players: [Player]?
    let NumPlayers: Int?
    let CurrentHand: HandState?
    let MatchPoints: [String: Int]?   // Go map[int]int serializes as {"0":0,"1":0}
    let TurnPlayer: Int?
    let CurrentTeamTurn: Int?
    let Logs: [String]?
    let WinnerTeam: Int?
    let MatchFinished: Bool?
    let CanAskTruco: Bool?
    let PendingRaiseFor: Int?
    let PendingRaiseBy: Int?
    let PendingRaiseTo: Int?
    let CurrentPlayerIdx: Int?
    let LastTrickSeq: Int?
    let LastTrickTeam: Int?
    let LastTrickWinner: Int?
    let LastTrickTie: Bool?
    let LastTrickRound: Int?
    
    /// Helper to get team scores
    var teamScore: (us: Int, them: Int) {
        let us = MatchPoints?["0"] ?? 0
        let them = MatchPoints?["1"] ?? 0
        return (us, them)
    }
}

// MARK: - Hand State (matches Go truco.HandState)
struct HandState: Codable {
    let Vira: Card?
    let Manilha: String?    // Go Rank type serializes as string
    let Stake: Int?
    let TrucoByTeam: Int?
    let RaiseRequester: Int?
    let Dealer: Int?
    let Turn: Int?
    let Round: Int?
    let RoundStart: Int?
    let RoundCards: [PlayedCard]?
    let TrickResults: [Int]?
    let TrickWins: [String: Int]?  // Go map[int]int
    let WinnerTeam: Int?
    let Finished: Bool?
    let PendingRaiseFor: Int?
}

// MARK: - Player (matches Go truco.Player)
struct Player: Codable, Identifiable {
    let playerID: Int
    let Name: String
    let CPU: Bool?
    let Team: Int
    let Hand: [Card]?
    
    var id: Int { playerID }
    
    enum CodingKeys: String, CodingKey {
        case playerID = "ID"
        case Name, CPU, Team, Hand
    }
}

// MARK: - PlayedCard (matches Go truco.PlayedCard)
struct PlayedCard: Codable, Identifiable {
    let PlayerID: Int
    let Card: Card
    
    var id: String { "\(PlayerID)-\(Card.Rank)-\(Card.Suit)" }
}

// MARK: - Card (matches Go truco.Card)
struct Card: Codable, Equatable, Hashable {
    let Rank: String
    let Suit: String
    
    var isRed: Bool {
        return Suit == "Copas" || Suit == "Ouros"
    }
    
    var suitSymbol: String {
        switch Suit {
        case "Espadas": return "♠"
        case "Copas": return "♥"
        case "Ouros": return "♦"
        case "Paus": return "♣"
        default: return ""
        }
    }
}
