import Foundation

@MainActor
final class TrucoAppStore: ObservableObject {
    @Published private(set) var mode = "idle"
    @Published private(set) var snapshot: GameSnapshot?

    private let bridge = TrucoCoreBridge()

    func startOfflineDemo() {
        let intent = """
        {
          "kind": "new_offline_game",
          "payload": {
            "player_names": ["Voce", "CPU-2"],
            "cpu_flags": [false, true],
            "seed_lo": 7,
            "seed_hi": 9
          }
        }
        """

        Task.detached {
            let result = self.bridge.dispatch(intentJSON: intent)
            let snapshotStr = self.bridge.snapshotJSON()
            
            var parsedSnapshot: GameSnapshot? = nil
            if let data = snapshotStr?.data(using: .utf8) {
                parsedSnapshot = try? JSONDecoder().decode(GameSnapshot.self, from: data)
            }
            
            await MainActor.run {
                if let result {
                    self.status = "Erro: \(result)"
                    return
                }
                self.snapshot = parsedSnapshot
                self.status = "Rodando partida offline"
                self.mode = "offline_match"
            }
        }
    }
}

final class TrucoCoreBridge {
    private let handle: UInt

    init() {
        self.handle = 0
    }

    func dispatch(intentJSON: String) -> String? {
        // Replace this stub with direct calls into the generated C ABI.
        _ = intentJSON
        return nil
    }

    func snapshotJSON() -> String? {
        // Return a realistic mock matching the PHP web version structure
        return """
        {
          "mode": "playing",
          "NumPlayers": 2,
          "MatchPoints": [8, 10],
          "TurnPlayer": 0,
          "PendingRaiseFor": -1,
          "MatchFinished": false,
          "CurrentHand": {
             "Stake": 3,
             "Vira": {"Rank": "3", "Suit": "Ouros"},
             "Manilha": "4",
             "TrickWins": [1, 0],
             "RoundCards": [
               {"PlayerID": 1, "Card": {"Rank": "Q", "Suit": "Espadas"}}
             ]
          },
          "Players": [
             {"ID": 0, "Name": "Voce", "Team": 0, "Hand": [
                {"Rank": "7", "Suit": "Copas"},
                {"Rank": "A", "Suit": "Paus"}
             ]},
             {"ID": 1, "Name": "CPU-2", "Team": 1, "Hand": [
                {"Rank": "2", "Suit": "Copas"},
                {"Rank": "5", "Suit": "Ouros"}
             ]}
          ],
          "Logs": ["CPU-2 jogou Q de Espadas", "É a sua vez!"]
        }
        """
    }
}
