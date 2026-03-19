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
    private var lastOfflineNames: [String] = ["Voce", "CPU-2"]
    private var lastOfflineCPUFlags: [Bool] = [false, true]

    var canCloseSession: Bool {
        bundle?.ui?.actions?.can_close_session == true
    }

    // MARK: - Offline

    func startOffline(names: [String], cpuFlags: [Bool]) {
        lastOfflineNames = names
        lastOfflineCPUFlags = cpuFlags
        let intent = makeIntentJSON(kind: "new_offline_game", payload: [
            "player_names": names,
            "cpu_flags": cpuFlags,
        ])
        dispatchAndPoll(intent: intent, successStatus: "Partida offline")
    }

    func startOfflineDemo() {
        startOffline(names: ["Voce", "CPU-2"], cpuFlags: [false, true])
    }

    // MARK: - Online Host

    func createHost(name: String, numPlayers: Int, relayURL: String?) {
        var payload: [String: Any] = [
            "host_name": name,
            "num_players": numPlayers,
        ]
        if let relay = relayURL {
            payload["relay_url"] = relay
        }
        let intent = makeIntentJSON(kind: "create_host_session", payload: payload)
        dispatchAndPoll(intent: intent, successStatus: "Sala criada")
    }

    // MARK: - Online Join

    func joinSession(name: String, key: String, desiredRole: String) {
        let intent = makeIntentJSON(kind: "join_session", payload: [
            "key": key,
            "player_name": name,
            "desired_role": desiredRole,
        ])
        dispatchAndPoll(intent: intent, successStatus: "Conectado à sala")
    }

    // MARK: - Game Actions

    func dispatchGameAction(action: String, cardIndex: Int = 0) {
        var payload: [String: Any] = ["action": action]
        if action == "play" {
            payload["card_index"] = cardIndex
        }
        Task {
            if let result = bridge.dispatch(intentJSON: makeIntentJSON(kind: "game_action", payload: payload)) {
                self.status = self.runtimeErrorSummary(from: result) ?? "Erro: \(result)"
            }
            refreshSnapshot()
        }
    }
    
    func sendChat(text: String) {
        dispatchIntent(json: makeIntentJSON(kind: "send_chat", payload: ["text": text]))
    }
    
    func voteHost(candidateSeat: Int) {
        dispatchIntent(json: makeIntentJSON(kind: "vote_host", payload: ["candidate_seat": candidateSeat]))
    }
    
    func requestReplacementInvite(targetSeat: Int) {
        dispatchIntent(json: makeIntentJSON(kind: "request_replacement_invite", payload: ["target_seat": targetSeat]))
    }
    
    func startHostedMatch() {
        dispatchIntent(json: makeIntentJSON(kind: "start_hosted_match"))
    }
    
    func closeSession() {
        guard canCloseSession else { return }
        dispatchIntent(json: makeIntentJSON(kind: "close_session"))
        events.removeAll()
    }

    func replayOfflineMatch() {
        startOffline(names: lastOfflineNames, cpuFlags: lastOfflineCPUFlags)
    }

    func dispatchIntent(json: String) {
        Task {
            if let result = bridge.dispatch(intentJSON: json) {
                status = runtimeErrorSummary(from: result) ?? "Erro: \(result)"
            }
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
                status = runtimeErrorSummary(from: result) ?? "Erro: \(result)"
                return
            }
            status = successStatus

            if !isPolling {
                isPolling = true
                startPollingLoop()
            }
        }
    }

    private func refreshSnapshot(using bridge: TrucoCoreBridge? = nil) {
        let bridge = bridge ?? self.bridge
        guard let snapshotStr = bridge.snapshotJSON(),
              let data = snapshotStr.data(using: .utf8),
              let parsed = try? JSONDecoder().decode(SnapshotBundle.self, from: data) else {
            return
        }
        bundle = parsed
        snapshot = parsed.match
        mode = parsed.mode
    }

    private func runtimeErrorSummary(from response: String) -> String? {
        guard let data = response.data(using: .utf8),
              let runtimeError = try? JSONDecoder().decode(RuntimeErrorSnapshot.self, from: data),
              let message = runtimeError.message?.trimmingCharacters(in: .whitespacesAndNewlines),
              !message.isEmpty else {
            return nil
        }

        if let code = runtimeError.code?.trimmingCharacters(in: .whitespacesAndNewlines),
           !code.isEmpty {
            return "\(code): \(message)"
        }
        return message
    }

    private func startPollingLoop() {
        let bridge = self.bridge
        Task.detached { [weak self, bridge] in
            guard let self else { return }
            while await self.isPolling {
                if let eventJsonStr = bridge.pollEventJSON() {
                    await self.consumePolledEvent(eventJsonStr)
                }
                await self.refreshSnapshot(using: bridge)
                try? await Task.sleep(nanoseconds: 50_000_000) // 50ms
            }
        }
    }

    private func consumePolledEvent(_ eventJSON: String) {
        guard let data = eventJSON.data(using: .utf8),
              let event = try? JSONDecoder().decode(AppEvent.self, from: data) else {
            return
        }
        events.append(event)
        if events.count > 100 {
            events.removeFirst()
        }
    }

    private func makeIntentJSON(kind: String, payload: [String: Any]? = nil) -> String {
        var object: [String: Any] = ["kind": kind]
        if let payload {
            object["payload"] = payload
        }
        guard let data = try? JSONSerialization.data(withJSONObject: object),
              let json = String(data: data, encoding: .utf8) else {
            return "{\"kind\":\"\(kind)\"}"
        }
        return json
    }
}

final class TrucoCoreBridge: @unchecked Sendable {
    private static let requiredCoreAPIVersion = 1
    private static let requiredSnapshotSchemaVersion = 2
    private let handle: UInt

    init() {
        self.handle = UInt(TrucoCoreCreate())
        guard self.handle != 0 else {
            fatalError("Failed to initialize the shared Truco runtime.")
        }
        validateRuntimeVersion()
    }

    deinit {
        TrucoCoreDestroy(self.handle)
    }

    private func validateRuntimeVersion() {
        guard let versions = readRuntimeVersions() else {
            fatalError("The shared Truco runtime did not return version metadata.")
        }

        guard versions.core_api_version == Self.requiredCoreAPIVersion,
              versions.snapshot_schema_version == Self.requiredSnapshotSchemaVersion else {
            fatalError(
                "Incompatible Truco runtime. Expected core_api=\(Self.requiredCoreAPIVersion), snapshot_schema=\(Self.requiredSnapshotSchemaVersion); " +
                "found core_api=\(versions.core_api_version ?? -1), snapshot_schema=\(versions.snapshot_schema_version ?? -1)."
            )
        }
    }

    private func readRuntimeVersions() -> CoreVersions? {
        guard let cString = TrucoCoreVersionsJSON() else {
            return nil
        }
        defer { TrucoCoreFreeString(cString) }

        let json = String(cString: cString)
        let data = Data(json.utf8)
        return try? JSONDecoder().decode(CoreVersions.self, from: data)
    }

    nonisolated func dispatch(intentJSON: String) -> String? {
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

    nonisolated func snapshotJSON() -> String? {
        let snapCStr = TrucoCoreSnapshotJSON(self.handle)
        defer {
            if snapCStr != nil { TrucoCoreFreeString(snapCStr) }
        }

        if let snapCStr = snapCStr {
            return String(cString: snapCStr)
        }
        return nil
    }

    nonisolated func pollEventJSON() -> String? {
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

private struct RuntimeErrorSnapshot: Codable {
    let code: String?
    let message: String?
}
