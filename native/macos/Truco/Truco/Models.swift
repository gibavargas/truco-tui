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

enum JSONValue: Codable, Hashable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case object([String: JSONValue])
    case array([JSONValue])
    case null

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if container.decodeNil() {
            self = .null
        } else if let value = try? container.decode(Bool.self) {
            self = .bool(value)
        } else if let value = try? container.decode(Double.self) {
            self = .number(value)
        } else if let value = try? container.decode(String.self) {
            self = .string(value)
        } else if let value = try? container.decode([String: JSONValue].self) {
            self = .object(value)
        } else if let value = try? container.decode([JSONValue].self) {
            self = .array(value)
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unsupported JSON value")
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case .string(let value):
            try container.encode(value)
        case .number(let value):
            try container.encode(value)
        case .bool(let value):
            try container.encode(value)
        case .object(let value):
            try container.encode(value)
        case .array(let value):
            try container.encode(value)
        case .null:
            try container.encodeNil()
        }
    }
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
    let metadata: [String: JSONValue]?
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
    let last_error: AppErrorSnapshot?
    let last_event_sequence: Int64?
}

struct DiagnosticsSnapshot: Codable {
    let event_backlog: Int?
    let event_log: [String]?
}

struct AppErrorSnapshot: Codable {
    let code: String?
    let message: String?
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
    let code: String?
    let message: String?
}

// MARK: - Match Snapshot (matches Go truco.Snapshot)
struct MatchSnapshot: Codable {
    let Players: [Player]?
    let NumPlayers: Int?
    let CurrentHand: HandState?
    let LastTrickCards: [PlayedCard]?
    let TrickPiles: [TrickPile]?
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

struct TrickPile: Codable {
    let Winner: Int?
    let Team: Int?
    let Round: Int?
    let Cards: [PlayedCard]?
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
    
    // Derived property to calculate the highest card played so far
    var winningCardId: String? {
        guard let cards = RoundCards, !cards.isEmpty else { return nil }
        var bestId: String? = nil
        var bestPower = -1
        var isTie = false
        
        for pc in cards {
            let p = pc.Card.power(manilha: Manilha)
            if p > bestPower {
                bestPower = p
                bestId = pc.id
                isTie = false
            } else if p == bestPower {
                isTie = true
            }
        }
        
        return isTie ? nil : bestId
    }
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
    
    func power(manilha: String?) -> Int {
        let normalPower: [String: Int] = [
            "3": 10, "2": 9, "A": 8, "K": 7, "J": 6, "Q": 5, "7": 4, "6": 3, "5": 2, "4": 1
        ]
        let manilhaSuitPower: [String: Int] = [
            "Paus": 4, "Copas": 3, "Espadas": 2, "Ouros": 1
        ]
        
        if Rank == manilha {
            return 100 + (manilhaSuitPower[Suit] ?? 0)
        }
        return normalPower[Rank] ?? 0
    }
}
