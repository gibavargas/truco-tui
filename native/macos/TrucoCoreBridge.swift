import Foundation

@MainActor
final class TrucoAppStore: ObservableObject {
    @Published private(set) var mode = "idle"
    @Published private(set) var status = "Aguardando..."
    @Published private(set) var snapshot: MatchSnapshot?
    @Published private(set) var bundle: SnapshotBundle?

    private let bridge = TrucoCoreBridge()

    private var isPolling = false

    func startOfflineGame(playerNames: [String], cpuFlags: [Bool]) {
        guard let namesData = try? JSONSerialization.data(withJSONObject: playerNames),
              let namesStr = String(data: namesData, encoding: .utf8),
              let flagsData = try? JSONSerialization.data(withJSONObject: cpuFlags),
              let flagsStr = String(data: flagsData, encoding: .utf8) else { return }
        
        let intent = """
        {"kind":"new_offline_game","payload":{"player_names":\(namesStr),"cpu_flags":\(flagsStr)}}
        """

        Task.detached { [bridge, weak self] in
            let result = bridge.dispatch(intentJSON: intent)
            
            await MainActor.run {
                guard let self = self else { return }
                if let result {
                    self.status = "Erro: \(result)"
                    return
                }
                self.status = "Rodando partida offline (\(playerNames.count)p)"
                self.mode = "offline_match"
                self.refreshSnapshot()
                
                if !self.isPolling {
                    self.isPolling = true
                    self.startPollingLoop()
                }
            }
        }
    }
    
    func startOfflineDemo() {
        startOfflineGame(
            playerNames: ["Voce", "CPU-2"],
            cpuFlags: [false, true]
        )
    }
    
    func setLocale(_ locale: String) {
        let intent = """
        {"kind":"set_locale","payload":{"locale":"\(locale)"}}
        """
        Task.detached { [bridge] in
            _ = bridge.dispatch(intentJSON: intent)
        }
    }
    
    func refreshSnapshot() {
        guard let snapshotStr = bridge.snapshotJSON(),
              let data = snapshotStr.data(using: .utf8) else {
            return
        }
        
        if let parsedBundle = try? JSONDecoder().decode(SnapshotBundle.self, from: data) {
            self.bundle = parsedBundle
            self.snapshot = parsedBundle.match
            if let m = parsedBundle.mode {
                self.mode = m
            }
        }
    }
    
    private func startPollingLoop() {
        Task.detached { [bridge, weak self] in
            while true {
                guard let self = self else { return }
                let stillPolling = await MainActor.run { self.isPolling }
                guard stillPolling else { return }
                
                if let _ = bridge.pollEventJSON() {
                    await MainActor.run {
                        self.refreshSnapshot()
                    }
                }
                await MainActor.run {
                    self.refreshSnapshot()
                }
                try? await Task.sleep(nanoseconds: 50_000_000) // 50ms
            }
        }
    }
    
    func dispatchGameAction(_ action: String, cardIndex: Int? = nil) {
        var payload: [String: Any] = ["action": action]
        if let idx = cardIndex {
            payload["card_index"] = idx
        }
        
        guard let payloadData = try? JSONSerialization.data(withJSONObject: payload),
              let payloadStr = String(data: payloadData, encoding: .utf8) else { return }
        
        let intent = """
        {"kind":"game_action","payload":\(payloadStr)}
        """
        
        Task.detached { [bridge] in
            _ = bridge.dispatch(intentJSON: intent)
        }
    }
}

final class TrucoCoreBridge: @unchecked Sendable {
    private let handle: UInt

    init() {
        self.handle = UInt(TrucoCoreCreate())
    }
    
    deinit {
        TrucoCoreDestroy(self.handle)
    }

    func dispatch(intentJSON: String) -> String? {
        let resultCStr = intentJSON.withCString { cStr in
            TrucoCoreDispatchIntentJSON(self.handle, UnsafeMutablePointer(mutating: cStr))
        }
        defer { TrucoCoreFreeString(resultCStr) }
        
        if let resultCStr = resultCStr {
            return String(cString: resultCStr)
        }
        return nil
    }

    func snapshotJSON() -> String? {
        let snapCStr = TrucoCoreSnapshotJSON(self.handle)
        defer { TrucoCoreFreeString(snapCStr) }
        
        if let snapCStr = snapCStr {
            return String(cString: snapCStr)
        }
        return nil
    }
    
    func pollEventJSON() -> String? {
        let eventCStr = TrucoCorePollEventJSON(self.handle)
        defer { TrucoCoreFreeString(eventCStr) }
        
        if let eventCStr = eventCStr {
            return String(cString: eventCStr)
        }
        return nil
    }
}
