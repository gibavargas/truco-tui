import Foundation

// MARK: - SnapshotBundle (top-level response from FFI)
struct SnapshotBundle: Codable {
    let mode: String?
    let locale: String?
    let match: MatchSnapshot?
    let lobby: LobbySnapshot?
    let connection: ConnectionSnapshot?
    let diagnostics: DiagnosticsSnapshot?
}

// MARK: - Match Snapshot (the actual game state)
struct MatchSnapshot: Codable {
    let NumPlayers: Int?
    let MatchPoints: [String: Int]?
    let TurnPlayer: Int?
    let CurrentTeamTurn: Int?
    let PendingRaiseFor: Int?
    let PendingRaiseBy: Int?
    let PendingRaiseTo: Int?
    let MatchFinished: Bool?
    let WinnerTeam: Int?
    let CanAskTruco: Bool?
    let CurrentPlayerIdx: Int?
    let CurrentHand: HandState?
    let Players: [Player]?
    let Logs: [String]?
    let LastTrickSeq: Int?
    let LastTrickTeam: Int?
    let LastTrickWinner: Int?
    let LastTrickTie: Bool?
    let LastTrickRound: Int?
}

struct HandState: Codable {
    let Stake: Int?
    let Vira: Card?
    let Manilha: String?
    let TrickWins: [String: Int]?
    let RoundCards: [PlayedCard]?
    let TrickResults: [Int]?
    let Round: Int?
    let RoundStart: Int?
    let Dealer: Int?
    let Turn: Int?
    let TrucoByTeam: Int?
    let RaiseRequester: Int?
    let WinnerTeam: Int?
    let Finished: Bool?
}

struct Player: Codable, Identifiable {
    let ID: Int
    let Name: String
    let Team: Int
    let CPU: Bool?
    let ProvisionalCPU: Bool?
    let Hand: [Card]?
    
    var id: Int { ID }
}

struct PlayedCard: Codable, Identifiable {
    let PlayerID: Int
    let Card: Card
    
    var id: String { "\(PlayerID)-\(Card.Rank)-\(Card.Suit)" }
}

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

// MARK: - Supporting Snapshots
struct LobbySnapshot: Codable {
    let invite_key: String?
    let slots: [String]?
    let assigned_seat: Int?
    let num_players: Int?
    let started: Bool?
    let host_seat: Int?
}

struct ConnectionSnapshot: Codable {
    let status: String?
    let is_online: Bool?
    let is_host: Bool?
}

struct DiagnosticsSnapshot: Codable {
    let event_backlog: Int?
    let event_log: [String]?
}
