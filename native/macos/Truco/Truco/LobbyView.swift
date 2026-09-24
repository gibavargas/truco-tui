import SwiftUI

// MARK: - Main Lobby (mirrors TUI menu)

struct LobbyView: View {
    @EnvironmentObject var store: TrucoAppStore
    @State private var showOfflineSetup = false
    @State private var showOnlineMenu = false
    @State private var showLanguage = false
    
    var body: some View {
        ZStack {
            // Background
            LinearGradient(
                colors: [
                    Color(red: 0.055, green: 0.075, blue: 0.09),
                    Color(red: 0.10, green: 0.14, blue: 0.13),
                    Color(red: 0.12, green: 0.08, blue: 0.045)
                ],
                startPoint: .top,
                endPoint: .bottom
            )
            .ignoresSafeArea()
            
            LobbySuitPattern()

            GeometryReader { geometry in
                let compact = geometry.size.width < 760 || geometry.size.height < 720
                ScrollView {
                    VStack(spacing: 0) {
                        Spacer(minLength: compact ? 28 : 54)

                        VStack(spacing: 12) {
                            HStack(spacing: compact ? 10 : 16) {
                                Text("♠")
                                    .font(.system(size: compact ? 28 : 40))
                                    .foregroundColor(.white.opacity(0.4))
                                Text("TRUCO")
                                    .font(.system(size: compact ? 42 : 56, weight: .black, design: .rounded))
                                    .foregroundColor(.white)
                                    .tracking(compact ? 4 : 6)
                                Text("♣")
                                    .font(.system(size: compact ? 28 : 40))
                                    .foregroundColor(.white.opacity(0.4))
                            }

                            Text("PAULISTA")
                                .font(.system(size: compact ? 14 : 18, weight: .bold, design: .rounded))
                                .foregroundColor(.yellow.opacity(0.7))
                                .tracking(compact ? 5 : 8)

                            Rectangle()
                                .fill(LinearGradient(colors: [.clear, .yellow.opacity(0.46), .clear], startPoint: .leading, endPoint: .trailing))
                                .frame(width: compact ? 150 : 200, height: 2)
                                .padding(.top, 8)
                        }

                        Spacer().frame(height: compact ? 30 : 50)

                        VStack(spacing: compact ? 12 : 16) {
                            LobbyButton(title: "Jogar Offline", icon: "person.fill", color: .green) {
                                showOfflineSetup = true
                            }

                            LobbyButton(title: "Jogar Online", icon: "network", color: .blue) {
                                showOnlineMenu = true
                            }

                            LobbyButton(title: "Idioma / Language", icon: "globe", color: .orange) {
                                showLanguage = true
                            }
                        }
                        .frame(maxWidth: compact ? 360 : 320)

                        Spacer().frame(height: 30)

                        if store.status != "Pronto para jogar" {
                            Text(store.status)
                                .font(.caption)
                                .foregroundColor(.yellow.opacity(0.8))
                                .padding(8)
                                .background(Color.black.opacity(0.3))
                                .cornerRadius(8)
                                .textSelection(.enabled)
                        }

                        Spacer(minLength: compact ? 28 : 50)

                        Text("v1.0 — Truco Paulista Nativo macOS")
                            .font(.caption2)
                            .foregroundColor(.white.opacity(0.25))
                            .padding(.bottom, 16)
                    }
                    .padding(.horizontal, compact ? 24 : 40)
                    .frame(maxWidth: .infinity, minHeight: geometry.size.height)
                }
            }
        }
        .sheet(isPresented: $showOfflineSetup) {
            OfflineSetupSheet()
                .environmentObject(store)
        }
        .sheet(isPresented: $showOnlineMenu) {
            OnlineMenuSheet()
                .environmentObject(store)
        }
        .sheet(isPresented: $showLanguage) {
            LanguageSheet()
                .environmentObject(store)
        }
    }
}

private struct LobbySuitPattern: View {
    private let suits = ["♠", "♥", "♦", "♣"]

    var body: some View {
        GeometryReader { geo in
            ForEach(0..<16, id: \.self) { index in
                let column = CGFloat(index % 4)
                let row = CGFloat(index / 4)
                Text(suits[index % suits.count])
                    .font(.system(size: 46 + CGFloat((index % 3) * 10), weight: .bold, design: .serif))
                    .foregroundColor(.white.opacity(index % 2 == 0 ? 0.035 : 0.022))
                    .rotationEffect(.degrees(Double(index % 2 == 0 ? 8 : -11)))
                    .position(
                        x: geo.size.width * (0.12 + column * 0.25),
                        y: geo.size.height * (0.12 + row * 0.24)
                    )
            }
        }
        .allowsHitTesting(false)
    }
}

// MARK: - Lobby Button

private struct LobbyButton: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void
    
    @State private var isHovered = false
    
    var body: some View {
        Button(action: action) {
            HStack(spacing: 14) {
                Image(systemName: icon)
                    .font(.title2)
                    .foregroundColor(color)
                    .frame(width: 36)
                
                Text(title)
                    .font(.headline)
                    .foregroundColor(.white)
                
                Spacer()
                
                Image(systemName: "chevron.right")
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.3))
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 16)
            .background(
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .fill(
                        LinearGradient(
                            colors: [
                                color.opacity(isHovered ? 0.22 : 0.12),
                                Color.white.opacity(isHovered ? 0.10 : 0.055)
                            ],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
            )
            .overlay(
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .stroke(color.opacity(isHovered ? 0.5 : 0.2), lineWidth: 1)
            )
            .shadow(color: color.opacity(isHovered ? 0.22 : 0.08), radius: isHovered ? 16 : 8, x: 0, y: 8)
            .scaleEffect(isHovered ? 1.018 : 1)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(title)
        .onHover { h in
            withAnimation(.easeInOut(duration: 0.15)) { isHovered = h }
        }
    }
}

// MARK: - Offline Setup Sheet

struct OfflineSetupSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    @State private var playerName = "Você"
    @State private var numPlayers = 2
    @State private var player2IsCPU = true
    @State private var player2Name = "CPU-2"
    @State private var player3IsCPU = true
    @State private var player3Name = "CPU-3"
    @State private var player4IsCPU = true
    @State private var player4Name = "CPU-4"
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Mesa offline")
                .font(.title2.bold())
            
            Form {
                TextField("Seu nome", text: $playerName)
                
                Picker("Jogadores", selection: $numPlayers) {
                    Text("2 jogadores").tag(2)
                    Text("4 jogadores").tag(4)
                }
                .pickerStyle(.segmented)
                
                Section("Jogador 2") {
                    Toggle("CPU", isOn: $player2IsCPU)
                    if !player2IsCPU {
                        TextField("Nome", text: $player2Name)
                    }
                }
                
                if numPlayers == 4 {
                    Section("Jogador 3") {
                        Toggle("CPU", isOn: $player3IsCPU)
                        if !player3IsCPU {
                            TextField("Nome", text: $player3Name)
                        }
                    }
                    
                    Section("Jogador 4") {
                        Toggle("CPU", isOn: $player4IsCPU)
                        if !player4IsCPU {
                            TextField("Nome", text: $player4Name)
                        }
                    }
                }
            }
            .formStyle(.grouped)
            
            HStack(spacing: 16) {
                Button("Cancelar") { dismiss() }
                    .buttonStyle(.bordered)
                
                Button("Começar partida") {
                    startGame()
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
            }
        }
        .padding()
        .frame(
            minWidth: 400,
            idealWidth: 420,
            minHeight: numPlayers == 4 ? 500 : 380,
            idealHeight: numPlayers == 4 ? 540 : 400
        )
    }
    
    private func startGame() {
        var names = [playerName.isEmpty ? "Você" : playerName]
        var cpus = [false]
        
        names.append(player2IsCPU ? "CPU-2" : (player2Name.isEmpty ? "Jogador 2" : player2Name))
        cpus.append(player2IsCPU)
        
        if numPlayers == 4 {
            names.append(player3IsCPU ? "CPU-3" : (player3Name.isEmpty ? "Jogador 3" : player3Name))
            cpus.append(player3IsCPU)
            names.append(player4IsCPU ? "CPU-4" : (player4Name.isEmpty ? "Jogador 4" : player4Name))
            cpus.append(player4IsCPU)
        }
        
        store.startOffline(names: names, cpuFlags: cpus)
    }
}

// MARK: - Online Menu Sheet

struct OnlineMenuSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    @State private var showHost = false
    @State private var showJoin = false
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Jogar Online")
                .font(.title2.bold())
            
            VStack(spacing: 12) {
                Button {
                    showHost = true
                } label: {
                    HStack {
                        Image(systemName: "antenna.radiowaves.left.and.right")
                        Text("Criar mesa")
                    }
                    .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .tint(.blue)
                .controlSize(.large)
                
                Button {
                    showJoin = true
                } label: {
                    HStack {
                        Image(systemName: "arrow.right.circle")
                        Text("Entrar com convite")
                    }
                    .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .controlSize(.large)
            }
            .frame(maxWidth: 280)
            
            Button("Voltar") { dismiss() }
                .buttonStyle(.bordered)
        }
        .padding(30)
        .frame(minWidth: 380, idealWidth: 420, minHeight: 240, idealHeight: 260)
        .sheet(isPresented: $showHost) {
            HostSetupSheet()
                .environmentObject(store)
        }
        .sheet(isPresented: $showJoin) {
            JoinSetupSheet()
                .environmentObject(store)
        }
    }
}

// MARK: - Host Setup

struct HostSetupSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    @State private var hostName = ""
    @State private var numPlayers = 2
    @State private var relayURL = ""
    @State private var transportMode = "auto"
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Criar mesa")
                .font(.title2.bold())
            
            Form {
                TextField("Seu nome", text: $hostName, prompt: Text("Host"))
                
                Picker("Jogadores", selection: $numPlayers) {
                    Text("2").tag(2)
                    Text("4").tag(4)
                }
                .pickerStyle(.segmented)

                Picker("Transporte", selection: $transportMode) {
                    Text("Auto").tag("auto")
                    Text("Direto").tag("tcp_tls")
                    Text("Tailnet").tag("tailnet_tsnet_v1")
                    Text("Relay").tag("relay_quic_v2")
                }
                .pickerStyle(.segmented)
                
                TextField("Relay URL (opcional)", text: $relayURL, prompt: Text("Deixe em branco para conexão direta"))
            }
            .formStyle(.grouped)
            
            HStack(spacing: 16) {
                Button("Cancelar") { dismiss() }
                    .buttonStyle(.bordered)
                
                Button("Criar") {
                    store.createHost(
                        name: hostName.isEmpty ? "Host" : hostName,
                        numPlayers: numPlayers,
                        relayURL: relayURL.isEmpty ? nil : relayURL,
                        transportMode: transportMode
                    )
                    dismiss()
                    // Dismiss parent sheets too
                }
                .buttonStyle(.borderedProminent)
                .tint(.blue)
            }
        }
        .padding()
        .frame(minWidth: 400, idealWidth: 440, minHeight: 300, idealHeight: 330)
    }
}

// MARK: - Join Setup

struct JoinSetupSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    @State private var playerName = ""
    @State private var inviteKey = ""
    @State private var desiredRole = "auto"
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Entrar com convite")
                .font(.title2.bold())
            
            Form {
                TextField("Seu nome", text: $playerName, prompt: Text("Jogador"))
                TextField("Chave de convite", text: $inviteKey, prompt: Text("Cole a chave aqui"))
                Picker("Papel", selection: $desiredRole) {
                    Text("Auto").tag("auto")
                    Text("Parceiro").tag("partner")
                    Text("Adversário").tag("opponent")
                }
                .pickerStyle(.segmented)
            }
            .formStyle(.grouped)
            
            HStack(spacing: 16) {
                Button("Cancelar") { dismiss() }
                    .buttonStyle(.bordered)
                
                Button("Entrar") {
                    store.joinSession(
                        name: playerName.isEmpty ? "Jogador" : playerName,
                        key: inviteKey,
                        desiredRole: desiredRole
                    )
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .disabled(inviteKey.isEmpty)
            }
        }
        .padding()
        .frame(minWidth: 400, idealWidth: 430, minHeight: 320, idealHeight: 360)
    }
}

// MARK: - Language Sheet

struct LanguageSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Idioma / Language")
                .font(.title2.bold())
            
            VStack(spacing: 12) {
                Button("🇧🇷 Português") {
                    store.dispatchIntent(json: "{\"kind\":\"set_locale\",\"payload\":{\"locale\":\"pt-BR\"}}")
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .controlSize(.large)
                
                Button("🇺🇸 English") {
                    store.dispatchIntent(json: "{\"kind\":\"set_locale\",\"payload\":{\"locale\":\"en-US\"}}")
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.blue)
                .controlSize(.large)
                
            }
            .frame(maxWidth: 240)
            
            Button("Voltar") { dismiss() }
                .buttonStyle(.bordered)
        }
        .padding(30)
        .frame(minWidth: 320, idealWidth: 360, minHeight: 280, idealHeight: 310)
    }
}

struct OnlineLobbyView: View {
    @EnvironmentObject var store: TrucoAppStore
    @State private var chatMessage = ""

    private var copy: TrucoCopy {
        TrucoCopy(locale: store.bundle?.locale)
    }
    
    var body: some View {
        let lobby = store.bundle?.lobby
        let slotStates = store.bundle?.ui?.lobby_slots ?? []
        let connection = store.bundle?.connection
        let diagnostics = store.bundle?.diagnostics

        ZStack {
            Color(red: 0.06, green: 0.08, blue: 0.12)
                .ignoresSafeArea()

            GeometryReader { geometry in
                let isCompact = geometry.size.width < 980 || geometry.size.height < 760
                ScrollView {
                    VStack(spacing: isCompact ? 18 : 24) {
                        Text(store.mode.contains("host") ? copy.text("🏠 Mesa criada", "🏠 Table created") : copy.text("🔗 Conectado por convite", "🔗 Connected by invite"))
                            .font(.system(size: isCompact ? 30 : 34, weight: .black, design: .rounded))
                            .foregroundColor(.white)

                        if isCompact {
                            VStack(spacing: 16) {
                                lobbyPrimaryColumn(lobby: lobby, slotStates: slotStates, connection: connection, diagnostics: diagnostics, compact: true)
                                    .frame(maxWidth: 720)

                                lobbyEventsColumn(compact: true)
                                    .frame(maxWidth: 720, minHeight: 200)
                            }
                        } else {
                            HStack(alignment: .top, spacing: 30) {
                                lobbyPrimaryColumn(lobby: lobby, slotStates: slotStates, connection: connection, diagnostics: diagnostics, compact: false)
                                    .frame(width: 400)

                                lobbyEventsColumn(compact: false)
                                    .frame(width: 340, height: 370)
                            }
                        }

                        if isCompact {
                            HStack(spacing: 12) {
                                if store.mode == "host_lobby" {
                                    Button(copy.text("Começar partida", "Start match")) {
                                        store.startHostedMatch()
                                    }
                                    .buttonStyle(.borderedProminent)
                                    .tint(.green)
                                    .controlSize(.regular)
                                }

                                Button(copy.text("Sair da mesa", "Leave table")) {
                                    store.closeSession()
                                }
                                .disabled(!store.canCloseSession)
                                .buttonStyle(.borderedProminent)
                                .tint(.red)
                                .controlSize(.regular)
                            }
                        } else {
                            Button(copy.text("Sair da mesa", "Leave table")) {
                                store.closeSession()
                            }
                            .disabled(!store.canCloseSession)
                            .buttonStyle(.borderedProminent)
                            .tint(.red)
                            .controlSize(.large)
                        }
                    }
                    .padding(.horizontal, isCompact ? 28 : 40)
                    .padding(.top, isCompact ? 34 : 56)
                    .padding(.bottom, isCompact ? 24 : 32)
                    .frame(maxWidth: .infinity, minHeight: geometry.size.height, alignment: .top)
                }
            }
        }
    }

    @ViewBuilder
    private func lobbyPrimaryColumn(
        lobby: LobbySnapshot?,
        slotStates: [LobbySlotState],
        connection: ConnectionSnapshot?,
        diagnostics: DiagnosticsSnapshot?,
        compact: Bool
    ) -> some View {
        VStack(spacing: compact ? 14 : 20) {
            if let lobby {
                if let key = lobby.invite_key, !key.isEmpty {
                    VStack(spacing: 8) {
                        Text(copy.text("Chave de convite", "Invite key"))
                            .font(.footnote.weight(.semibold))
                            .foregroundColor(.white.opacity(0.7))

                        HStack(alignment: .center, spacing: 10) {
                            Text(key)
                                .font(.system(.body, design: .monospaced))
                                .foregroundColor(.yellow)
                                .textSelection(.enabled)
                                .lineLimit(2)
                                .frame(maxWidth: .infinity, alignment: .leading)

                            Button {
                                NSPasteboard.general.clearContents()
                                NSPasteboard.general.setString(key, forType: .string)
                            } label: {
                                Image(systemName: "doc.on.doc")
                            }
                            .buttonStyle(.bordered)
                        }
                        .padding(compact ? 10 : 12)
                        .background(Color.black.opacity(0.34))
                        .cornerRadius(10)
                        .accessibilityLabel(copy.text("Chave de convite", "Invite key"))
                    }
                }

                if !slotStates.isEmpty {
                    VStack(spacing: 10) {
                        Text("\(copy.text("Assentos", "Seats")) (\(slotStates.filter { !$0.is_empty }.count)/\(lobby.num_players ?? slotStates.count)):")
                            .font(.footnote.weight(.bold))
                            .foregroundColor(.white.opacity(0.7))

                        ForEach(slotStates) { slot in
                            VStack(alignment: .leading, spacing: 10) {
                                HStack(spacing: 10) {
                                    Circle()
                                        .fill(slotBadgeColor(for: slot))
                                        .frame(width: 10, height: 10)
                                    VStack(alignment: .leading, spacing: 4) {
                                        Text(slot.name?.isEmpty == false ? slot.name! : copy.waitingForPlayer)
                                            .font(.headline)
                                            .foregroundColor(slot.is_empty ? .gray : .white)
                                        Text(copy.slotStatusLabel(slot.status))
                                            .font(.footnote)
                                            .foregroundColor(.white.opacity(0.68))
                                    }
                                    Spacer()
                                    Text(copy.seatLabel(slot.seat))
                                        .font(.caption.weight(.semibold))
                                        .foregroundColor(.white.opacity(0.56))
                                }

                                HStack(spacing: 6) {
                                    if slot.is_local { slotTag(copy.youTag, color: .yellow) }
                                    if slot.is_host { slotTag(copy.hostTag, color: .blue) }
                                    slotTag(slot.is_connected ? copy.onlineTag : copy.offlineTag, color: slot.is_connected ? .green : .gray)
                                    if slot.is_provisional_cpu { slotTag(copy.cpuTag, color: .orange) }
                                }

                                HStack(spacing: 8) {
                                    if slot.can_vote_host {
                                        Button(copy.text("Votar host", "Vote host")) {
                                            store.voteHost(candidateSeat: slot.seat)
                                        }
                                        .font(.caption)
                                        .buttonStyle(.bordered)
                                    }
                                    if slot.can_request_replacement {
                                        Button(copy.text("Chamar substituto", "Invite substitute")) {
                                            store.requestReplacementInvite(targetSeat: slot.seat)
                                        }
                                        .font(.caption)
                                        .buttonStyle(.borderedProminent)
                                        .tint(.orange)
                                    }
                                }
                            }
                            .padding(compact ? 12 : 14)
                            .background(Color.white.opacity(0.06))
                            .overlay(
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(Color.white.opacity(0.08), lineWidth: 1)
                            )
                            .cornerRadius(12)
                            .accessibilityElement(children: .combine)
                        }
                    }
                }
            }

            VStack(alignment: .leading, spacing: 12) {
                let network = connection?.network
                Text(copy.text("Detalhes da mesa", "Table details"))
                    .font(.footnote.bold())
                    .foregroundColor(.white.opacity(0.7))
                connectionLine(copy.text("Estado", "Status"), connection?.status ?? store.mode)
                connectionLine(copy.text("Modo", "Mode"), connection?.is_online == true ? copy.onlineTag : copy.offlineTag)
                if let role = lobby?.role, !role.isEmpty {
                    connectionLine(copy.text("Papel", "Role"), copy.roleLabel(role))
                }
                if let network {
                    connectionLine(copy.text("Compatibilidade", "Compatibility"), network.compatibilitySummary(isHost: connection?.is_host == true))
                    ForEach(Array(network.diagnosticsLines(copy: copy).dropFirst()), id: \.0) { line in
                        connectionLine(line.0, line.1)
                    }
                }
                connectionLine(copy.text("Eventos", "Events"), "\(diagnostics?.event_backlog ?? 0)")
                if let message = connection?.last_error?.message, !message.isEmpty {
                    connectionLine(copy.text("Erro", "Error"), message, tint: .red.opacity(0.95))
                }
                if let entries = diagnostics?.event_log?.suffix(4), !entries.isEmpty {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(copy.text("Diagnóstico", "Diagnostics"))
                            .font(.caption.weight(.semibold))
                            .foregroundColor(.white.opacity(0.6))
                        ForEach(Array(entries.enumerated()), id: \.offset) { _, line in
                            Text(line)
                                .font(.caption2.monospaced())
                                .foregroundColor(.white.opacity(0.72))
                                .textSelection(.enabled)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding()
            .background(Color.white.opacity(0.06))
            .cornerRadius(12)

            if !compact && store.mode == "host_lobby" {
                Button(copy.text("Começar partida", "Start match")) {
                    store.startHostedMatch()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .controlSize(.large)
                .font(.headline.weight(.black))
            } else if !compact {
                Text(copy.text("Aguardando o host iniciar a partida...", "Waiting for the host to start the match..."))
                    .font(.footnote)
                    .foregroundColor(.white.opacity(0.6))
            }
        }
    }

    @ViewBuilder
    private func lobbyEventsColumn(compact: Bool) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(copy.text("Atualizações", "Updates"))
                .font(.headline)
                .foregroundColor(.white)

            ScrollViewReader { proxy in
                ScrollView {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(store.events) { ev in
                            eventRow(ev)
                        }
                    }
                    .padding(10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                }
                .background(Color.black.opacity(0.48))
                .cornerRadius(10)
                .onChange(of: store.events.count) {
                    if let last = store.events.last {
                        withAnimation { proxy.scrollTo(last.id, anchor: .bottom) }
                    }
                }
            }

            HStack {
                TextField(copy.text("Digite uma mensagem...", "Type a message..."), text: $chatMessage)
                    .textFieldStyle(.roundedBorder)
                    .onSubmit {
                        sendChatIfNeeded()
                    }
                Button(copy.text("Enviar", "Send")) {
                    sendChatIfNeeded()
                }
                .buttonStyle(.borderedProminent)
                .disabled(chatMessage.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }
        }
        .frame(maxHeight: compact ? 250 : .infinity, alignment: .top)
    }

    private func slotBadgeColor(for slot: LobbySlotState) -> Color {
        switch slot.status {
        case "occupied_online":
            return .green
        case "occupied_offline":
            return .red
        case "provisional_cpu":
            return .orange
        default:
            return .gray
        }
    }

    private func slotTag(_ label: String, color: Color) -> some View {
        Text(label)
            .font(.caption.weight(.semibold))
            .foregroundColor(color)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(color.opacity(0.12))
            .clipShape(Capsule())
    }

    private func connectionLine(_ label: String, _ value: String, tint: Color = .white) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label.uppercased())
                .font(.caption.weight(.semibold))
                .foregroundColor(.white.opacity(0.52))
            Text(value)
                .font(.footnote)
                .foregroundColor(tint)
        }
    }

    private func sendChatIfNeeded() {
        let trimmed = chatMessage.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        store.sendChat(text: trimmed)
        chatMessage = ""
    }

    @ViewBuilder
    private func eventRow(_ ev: AppEvent) -> some View {
        switch ev.kind {
        case "chat":
            HStack(alignment: .top) {
                Text("\(ev.payload?.author ?? copy.text("Alguém", "Someone")):")
                    .font(.footnote.bold())
                    .foregroundColor(.cyan)
                Text(ev.payload?.text ?? "")
                    .font(.footnote)
                    .foregroundColor(.white)
            }
        case "system":
            Text(ev.payload?.text ?? "")
                .font(.footnote.italic())
                .foregroundColor(.gray)
        case "replacement_invite":
            Text("\(copy.text("Convite de substituição", "Replacement invite")) (\(copy.seatLabel(max(0, ev.payload?.target_seat ?? 0)))): \(ev.payload?.invite_key ?? "")")
                .font(.footnote.italic())
                .foregroundColor(.yellow)
                .textSelection(.enabled)
        case "error":
            Text(ev.payload?.message ?? ev.payload?.text ?? copy.text("Erro", "Error"))
                .font(.footnote)
                .foregroundColor(.red.opacity(0.9))
        case "lobby_updated":
            Text(copy.text("Lobby atualizado", "Lobby updated"))
                .font(.footnote)
                .foregroundColor(.white.opacity(0.6))
        case "match_updated":
            Text(copy.text("Partida atualizada", "Match updated"))
                .font(.footnote)
                .foregroundColor(.white.opacity(0.6))
        default:
            Text(copy.eventSummary(ev))
                .font(.footnote)
                .foregroundColor(.white.opacity(0.7))
        }
    }
}
