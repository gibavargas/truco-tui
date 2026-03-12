import Foundation
import Combine

@MainActor
final class TrucoAppStore: ObservableObject {
    @Published private(set) var mode = "idle"
    @Published private(set) var status = "Pronto para jogar"
    @Published private(set) var snapshot: MatchSnapshot?
    @Published private(set) var bundle: SnapshotBundle?
    @Published private(set) var events: [AppEvent] = []

    private let bridge = TrucoCoreBridge()
    private var isPolling = false

    // MARK: - Offline

    func startOffline(names: [String], cpuFlags: [Bool]) {
        let namesJSON = names.map { "\"\($0)\"" }.joined(separator: ",")
        let cpuJSON = cpuFlags.map { $0 ? "true" : "false" }.joined(separator: ",")
        let intent = "{\"kind\":\"new_offline_game\",\"payload\":{\"player_names\":[\(namesJSON)],\"cpu_flags\":[\(cpuJSON)]}}"
        
        dispatchAndPoll(intent: intent, successStatus: "Partida offline")
    }

    func startOfflineDemo() {
        startOffline(names: ["Voce", "CPU-2"], cpuFlags: [false, true])
    }

    // MARK: - Online Host

    func createHost(name: String, numPlayers: Int, relayURL: String?) {
        var payload = "\"host_name\":\"\(name)\",\"num_players\":\(numPlayers)"
        if let relay = relayURL {
            payload += ",\"relay_url\":\"\(relay)\""
        }
        let intent = "{\"kind\":\"create_host_session\",\"payload\":{\(payload)}}"
        dispatchAndPoll(intent: intent, successStatus: "Sala criada")
    }

    // MARK: - Online Join

    func joinSession(name: String, key: String) {
        let intent = "{\"kind\":\"join_session\",\"payload\":{\"key\":\"\(key)\",\"player_name\":\"\(name)\"}}"
        dispatchAndPoll(intent: intent, successStatus: "Conectado à sala")
    }

    // MARK: - Game Actions

    func dispatchGameAction(action: String, cardIndex: Int = 0) {
        let json: String
        if action == "play" {
            json = "{\"kind\":\"game_action\",\"payload\":{\"action\":\"play\",\"card_index\":\(cardIndex)}}"
        } else {
            json = "{\"kind\":\"game_action\",\"payload\":{\"action\":\"\(action)\"}}"
        }
        Task {
            _ = bridge.dispatch(intentJSON: json)
            refreshSnapshot()
        }
    }
    
    func sendChat(text: String) {
        let json = "{\"kind\":\"send_chat\",\"payload\":{\"text\":\"\(text.replacingOccurrences(of: "\"", with: "\\\""))\"}}"
        dispatchIntent(json: json)
    }
    
    func voteHost(candidateSeat: Int) {
        let json = "{\"kind\":\"vote_host\",\"payload\":{\"candidate_seat\":\(candidateSeat)}}"
        dispatchIntent(json: json)
    }
    
    func requestReplacementInvite(targetSeat: Int) {
        let json = "{\"kind\":\"request_replacement_invite\",\"payload\":{\"target_seat\":\(targetSeat)}}"
        dispatchIntent(json: json)
    }
    
    func startHostedMatch() {
        dispatchIntent(json: "{\"kind\":\"start_hosted_match\"}")
    }
    
    func closeSession() {
        dispatchIntent(json: "{\"kind\":\"close_session\"}")
        DispatchQueue.main.async { [weak self] in
            self?.events.removeAll()
        }
    }

    func dispatchIntent(json: String) {
        Task {
            _ = bridge.dispatch(intentJSON: json)
            refreshSnapshot()
            // If mode changed to idle, stop polling
            if mode == "idle" {
                isPolling = false
                events.removeAll()
            }
        }
    }

    // MARK: - Internal

    private func dispatchAndPoll(intent: String, successStatus: String) {
        Task {
            let result = bridge.dispatch(intentJSON: intent)
            refreshSnapshot()

            if let result {
                status = "Erro: \(result)"
                return
            }
            status = successStatus

            if !isPolling {
                isPolling = true
                startPollingLoop()
            }
        }
    }

    nonisolated private func refreshSnapshot() {
        guard let snapshotStr = bridge.snapshotJSON(),
              let data = snapshotStr.data(using: .utf8),
              let parsed = try? JSONDecoder().decode(SnapshotBundle.self, from: data) else {
            return
        }
        Task { @MainActor [weak self] in
            self?.bundle = parsed
            self?.snapshot = parsed.match
            self?.mode = parsed.mode
        }
    }

    nonisolated private func startPollingLoop() {
        Task.detached { [weak self] in
            guard let self else { return }
            while await self.isPolling {
                if let eventJsonStr = self.bridge.pollEventJSON(),
                   let data = eventJsonStr.data(using: .utf8),
                   let event = try? JSONDecoder().decode(AppEvent.self, from: data) {
                    
                    Task { @MainActor [weak self] in
                        guard let self = self else { return }
                        self.events.append(event)
                        if self.events.count > 100 {
                            self.events.removeFirst()
                        }
                    }
                }
                self.refreshSnapshot()
                try? await Task.sleep(nanoseconds: 50_000_000) // 50ms
            }
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
        defer {
            if resultCStr != nil { TrucoCoreFreeString(resultCStr) }
        }

        if let resultCStr = resultCStr {
            return String(cString: resultCStr)
        }
        return nil
    }

    func snapshotJSON() -> String? {
        let snapCStr = TrucoCoreSnapshotJSON(self.handle)
        defer {
            if snapCStr != nil { TrucoCoreFreeString(snapCStr) }
        }

        if let snapCStr = snapCStr {
            return String(cString: snapCStr)
        }
        return nil
    }

    func pollEventJSON() -> String? {
        let eventCStr = TrucoCorePollEventJSON(self.handle)
        defer {
            if eventCStr != nil { TrucoCoreFreeString(eventCStr) }
        }

        if let eventCStr = eventCStr {
            return String(cString: eventCStr)
        }
        return nil
    }
}
