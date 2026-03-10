import Foundation

// MARK: - API Snapshot Models
struct GameSnapshot: Codable {
    let mode: String?
    let NumPlayers: Int?
    let MatchPoints: [Int]?
    let TurnPlayer: Int?
    let PendingRaiseFor: Int?
    let MatchFinished: Bool?
    let CurrentHand: HandState?
    let Players: [Player]?
    let Logs: [String]?
}

struct HandState: Codable {
    let Stake: Int?
    let Vira: Card?
    let Manilha: String?
    let TrickWins: [Int]?
    let RoundCards: [PlayedCard]?
}

struct Player: Codable, Identifiable {
    let ID: Int
    let Name: String
    let Team: Int
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
