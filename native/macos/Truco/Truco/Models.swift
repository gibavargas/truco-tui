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

struct TrucoCopy {
    let locale: String?

    var isEnglish: Bool {
        locale?.lowercased().hasPrefix("en") == true
    }

    func text(_ ptBR: String, _ enUS: String) -> String {
        isEnglish ? enUS : ptBR
    }

    func seatLabel(_ seat: Int) -> String {
        text("Assento \(seat + 1)", "Seat \(seat + 1)")
    }

    var waitingForPlayer: String { text("Aguardando...", "Waiting...") }
    var youTag: String { text("você", "you") }
    var hostTag: String { text("host", "host") }
    var onlineTag: String { text("online", "online") }
    var offlineTag: String { text("offline", "offline") }
    var cpuTag: String { text("cpu", "cpu") }

    func slotStatusLabel(_ status: String) -> String {
        switch status {
        case "occupied_online":
            return text("ocupado", "occupied")
        case "occupied_offline":
            return text("desconectado", "disconnected")
        case "provisional_cpu":
            return text("cpu provisória", "provisional cpu")
        default:
            return text("vazio", "empty")
        }
    }

    func roleLabel(_ role: String) -> String {
        switch role {
        case "partner":
            return text("Parceiro", "Partner")
        case "opponent":
            return text("Adversário", "Opponent")
        case "host":
            return text("Host", "Host")
        case "guest":
            return text("Convidado", "Guest")
        default:
            return role
        }
    }

    func eventSummary(_ event: AppEvent) -> String {
        switch event.kind {
        case "chat":
            let author = event.payload?.author ?? text("Alguém", "Someone")
            return "\(author): \(event.payload?.text ?? "")"
        case "system":
            return event.payload?.text ?? text("Atualização do sistema", "System update")
        case "replacement_invite":
            let key = event.payload?.invite_key ?? "-"
            return text("Convite de substituição: \(key)", "Replacement invite: \(key)")
        case "error":
            return event.payload?.message ?? event.payload?.text ?? text("Erro", "Error")
        case "lobby_updated":
            return text("Lobby atualizado", "Lobby updated")
        case "match_updated":
            return text("Partida atualizada", "Match updated")
        default:
            return event.payload?.text ?? event.kind
        }
    }
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
    let network: NetworkSnapshot?
    let last_error: AppErrorSnapshot?
    let last_event_sequence: Int64?
}

struct DiagnosticsSnapshot: Codable {
    let event_backlog: Int?
    let replay_seed_lo: UInt64?
    let replay_seed_hi: UInt64?
    let event_log: [String]?
}

struct NetworkSnapshot: Codable {
    let transport: String?
    let requested_transport: String?
    let direct_path_known: Bool?
    let direct_path: Bool?
    let relay_fallback: Bool?
    let coordinator_status: String?
    let coordinator_url: String?
    let tailnet_node: String?
    let tailnet_authority: String?
    let tailnet_service_port: Int?
    let fallback_reason: String?
    let supported_protocol_versions: [Int]?
    let negotiated_protocol_version: Int?
    let seat_protocol_versions: [String: Int]?
    let mixed_protocol_session: Bool?

    var transportLabel: String {
        switch transport {
        case "tailnet_tsnet_v1": return "Tailnet"
        case "relay_quic_v2": return "Relay QUIC v2"
        default: return "TCP + TLS"
        }
    }

    var supportedVersionsLabel: String {
        guard let versions = supported_protocol_versions, !versions.isEmpty else { return "-" }
        return versions.map { "v\($0)" }.joined(separator: "/")
    }

    func protocolVersion(for seat: Int) -> Int? {
        seat_protocol_versions?["\(seat)"]
    }

    func compatibilitySummary(isHost: Bool) -> String {
        if isHost {
            var unique: [Int] = []
            for version in (seat_protocol_versions ?? [:]).values.filter({ $0 > 0 }).sorted(by: >) {
                if !unique.contains(version) {
                    unique.append(version)
                }
            }
            let summary = unique.isEmpty ? supportedVersionsLabel : unique.map { "v\($0)" }.joined(separator: "/")
            return mixed_protocol_session == true ? "Sessão mista \(summary)" : summary
        }
        if let negotiatedProtocolVersion = negotiated_protocol_version, negotiatedProtocolVersion > 0 {
            return "Negociado v\(negotiatedProtocolVersion)"
        }
        return supportedVersionsLabel
    }

    func routeSummary(copy: TrucoCopy) -> String {
        if relay_fallback == true {
            return copy.text("Relay ativo", "Relay active")
        }
        if direct_path_known == true {
            return direct_path == true
                ? copy.text("Conexão direta confirmada", "Direct path confirmed")
                : copy.text("Sem caminho direto", "No direct path")
        }
        if transport == "tailnet_tsnet_v1" {
            return copy.text("Rota via Tailnet", "Tailnet route")
        }
        return copy.text("Aguardando rota", "Waiting for route")
    }

    func diagnosticsLines(copy: TrucoCopy) -> [(String, String)] {
        var lines: [(String, String)] = [
            (copy.text("Rede", "Network"), transportLabel),
            (copy.text("Rota", "Route"), routeSummary(copy: copy)),
        ]

        if let requested_transport, !requested_transport.isEmpty, requested_transport != transport {
            lines.append((copy.text("Preferência", "Preference"), requested_transport))
        }
        if let coordinator_status, !coordinator_status.isEmpty {
            lines.append((copy.text("Coordenador", "Coordinator"), coordinator_status))
        }
        if let fallback_reason, !fallback_reason.isEmpty {
            lines.append((copy.text("Fallback", "Fallback"), fallback_reason))
        }
        if let tailnet_node, !tailnet_node.isEmpty {
            lines.append((copy.text("Nó Tailnet", "Tailnet node"), tailnet_node))
        }
        return lines
    }
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
        if cards.contains(where: { $0.FaceDown == true }) {
            return nil
        }
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
    let FaceDown: Bool?
    
    var id: String { "\(PlayerID)-\(Card.Rank)-\(Card.Suit)-\(FaceDown == true ? 1 : 0)" }
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

    var accessibilityLabel: String {
        if Rank.isEmpty || Suit.isEmpty {
            return "Carta virada para baixo"
        }
        return "\(spokenRank) de \(spokenSuit)"
    }

    private var spokenRank: String {
        switch Rank {
        case "A": return "ás"
        case "K": return "rei"
        case "Q": return "dama"
        case "J": return "valete"
        default: return Rank
        }
    }

    private var spokenSuit: String {
        switch Suit {
        case "Espadas": return "espadas"
        case "Copas": return "copas"
        case "Ouros": return "ouros"
        case "Paus": return "paus"
        default: return Suit.lowercased()
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
