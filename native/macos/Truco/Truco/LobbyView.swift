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
                colors: [Color(red: 0.06, green: 0.08, blue: 0.12), Color(red: 0.10, green: 0.14, blue: 0.20)],
                startPoint: .top,
                endPoint: .bottom
            )
            .ignoresSafeArea()
            
            // Subtle card suit pattern
            GeometryReader { geo in
                ForEach(0..<12, id: \.self) { i in
                    Text(["♠", "♥", "♦", "♣"][i % 4])
                        .font(.system(size: 60))
                        .foregroundColor(.white.opacity(0.03))
                        .position(
                            x: CGFloat.random(in: 0...geo.size.width),
                            y: CGFloat.random(in: 0...geo.size.height)
                        )
                }
            }
            
            VStack(spacing: 0) {
                Spacer()
                
                // Title
                VStack(spacing: 12) {
                    HStack(spacing: 16) {
                        Text("♠").font(.system(size: 40)).foregroundColor(.white.opacity(0.4))
                        Text("TRUCO")
                            .font(.system(size: 56, weight: .black, design: .rounded))
                            .foregroundColor(.white)
                            .tracking(6)
                        Text("♣").font(.system(size: 40)).foregroundColor(.white.opacity(0.4))
                    }
                    
                    Text("PAULISTA")
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundColor(.yellow.opacity(0.7))
                        .tracking(8)
                    
                    Rectangle()
                        .fill(LinearGradient(colors: [.clear, .yellow.opacity(0.4), .clear], startPoint: .leading, endPoint: .trailing))
                        .frame(width: 200, height: 2)
                        .padding(.top, 8)
                }
                
                Spacer().frame(height: 50)
                
                // Menu buttons
                VStack(spacing: 16) {
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
                .frame(maxWidth: 320)
                
                Spacer().frame(height: 30)
                
                // Status
                if store.status != "Pronto para jogar" {
                    Text(store.status)
                        .font(.caption)
                        .foregroundColor(.yellow.opacity(0.8))
                        .padding(8)
                        .background(Color.black.opacity(0.3))
                        .cornerRadius(8)
                }
                
                Spacer()
                
                // Footer
                Text("v1.0 — Truco Paulista Nativo macOS")
                    .font(.caption2)
                    .foregroundColor(.white.opacity(0.25))
                    .padding(.bottom, 16)
            }
            .padding(.horizontal, 40)
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
                    .fill(Color.white.opacity(isHovered ? 0.12 : 0.06))
            )
            .overlay(
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .stroke(color.opacity(isHovered ? 0.5 : 0.2), lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
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
            Text("Partida Offline")
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
                
                Button("Iniciar Partida") {
                    startGame()
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
            }
        }
        .padding()
        .frame(width: 400, height: numPlayers == 4 ? 520 : 380)
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
                        Text("Criar Sala (Host)")
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
                        Text("Entrar em Sala (Join)")
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
        .frame(width: 400, height: 240)
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
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Criar Sala")
                .font(.title2.bold())
            
            Form {
                TextField("Seu nome", text: $hostName, prompt: Text("Host"))
                
                Picker("Jogadores", selection: $numPlayers) {
                    Text("2").tag(2)
                    Text("4").tag(4)
                }
                .pickerStyle(.segmented)
                
                TextField("Relay URL (opcional)", text: $relayURL, prompt: Text("Deixe vazio para P2P direto"))
            }
            .formStyle(.grouped)
            
            HStack(spacing: 16) {
                Button("Cancelar") { dismiss() }
                    .buttonStyle(.bordered)
                
                Button("Criar") {
                    store.createHost(
                        name: hostName.isEmpty ? "Host" : hostName,
                        numPlayers: numPlayers,
                        relayURL: relayURL.isEmpty ? nil : relayURL
                    )
                    dismiss()
                    // Dismiss parent sheets too
                }
                .buttonStyle(.borderedProminent)
                .tint(.blue)
            }
        }
        .padding()
        .frame(width: 400, height: 300)
    }
}

// MARK: - Join Setup

struct JoinSetupSheet: View {
    @EnvironmentObject var store: TrucoAppStore
    @Environment(\.dismiss) var dismiss
    
    @State private var playerName = ""
    @State private var inviteKey = ""
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Entrar na Sala")
                .font(.title2.bold())
            
            Form {
                TextField("Seu nome", text: $playerName, prompt: Text("Jogador"))
                TextField("Chave de convite", text: $inviteKey, prompt: Text("Cole a chave aqui"))
            }
            .formStyle(.grouped)
            
            HStack(spacing: 16) {
                Button("Cancelar") { dismiss() }
                    .buttonStyle(.bordered)
                
                Button("Entrar") {
                    store.joinSession(
                        name: playerName.isEmpty ? "Jogador" : playerName,
                        key: inviteKey
                    )
                    dismiss()
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .disabled(inviteKey.isEmpty)
            }
        }
        .padding()
        .frame(width: 400, height: 280)
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
        .frame(width: 350, height: 300)
    }
}

struct OnlineLobbyView: View {
    @EnvironmentObject var store: TrucoAppStore
    @State private var chatMessage = ""
    
    var body: some View {
        ZStack {
            Color(red: 0.06, green: 0.08, blue: 0.12)
                .ignoresSafeArea()
            
            VStack(spacing: 24) {
                Text(store.mode.contains("host") ? "🏠 Sala Criada" : "🔗 Conectado")
                    .font(.title.bold())
                    .foregroundColor(.white)
                
                HStack(alignment: .top, spacing: 30) {
                    
                    // Left Column: Players and Info
                    VStack(spacing: 20) {
                        if let lobby = store.bundle?.lobby {
                            if let key = lobby.invite_key, !key.isEmpty {
                                VStack(spacing: 8) {
                                    Text("Chave de convite:")
                                        .font(.caption)
                                        .foregroundColor(.white.opacity(0.6))
                                    
                                    HStack {
                                        Text(key)
                                            .font(.system(.body, design: .monospaced))
                                            .foregroundColor(.yellow)
                                            .textSelection(.enabled)
                                        
                                        Button {
                                            NSPasteboard.general.clearContents()
                                            NSPasteboard.general.setString(key, forType: .string)
                                        } label: {
                                            Image(systemName: "doc.on.doc")
                                        }
                                        .buttonStyle(.bordered)
                                    }
                                    .padding(12)
                                    .background(Color.black.opacity(0.3))
                                    .cornerRadius(10)
                                }
                            }
                            
                            // Slots display
                            if let slots = lobby.slots {
                                VStack(spacing: 8) {
                                    Text("Jogadores (\(slots.filter { !$0.isEmpty }.count)/\(lobby.num_players ?? slots.count)):")
                                        .font(.caption.bold())
                                        .foregroundColor(.white.opacity(0.6))
                                    
                                    ForEach(Array(slots.enumerated()), id: \.offset) { index, name in
                                        HStack {
                                            Circle()
                                                .fill(name.isEmpty ? Color.gray : Color.green)
                                                .frame(width: 10, height: 10)
                                            Text(name.isEmpty ? "Aguardando..." : name)
                                                .foregroundColor(name.isEmpty ? .gray : .white)
                                                .frame(width: 120, alignment: .leading)
                                            
                                            if index == lobby.assigned_seat {
                                                Text("(você)")
                                                    .font(.caption)
                                                    .foregroundColor(.yellow)
                                            } else if name.isEmpty && store.mode.contains("host") {
                                                Button("Convidar CPU") {
                                                    store.requestReplacementInvite(targetSeat: index)
                                                }
                                                .font(.caption)
                                                .buttonStyle(.bordered)
                                            } else if !name.isEmpty {
                                                Button("Votar Host") {
                                                    store.voteHost(candidateSeat: index)
                                                }
                                                .font(.caption)
                                                .buttonStyle(.bordered)
                                            }
                                        }
                                        .padding(.vertical, 4)
                                    }
                                }
                                .padding()
                                .background(Color.white.opacity(0.05))
                                .cornerRadius(12)
                            }
                        }
                        
                        // Start Match
                        if store.mode == "host_lobby" {
                            Button("Iniciar Partida") {
                                store.dispatchIntent(json: "{\"kind\":\"start_hosted_match\"}")
                            }
                            .buttonStyle(.borderedProminent)
                            .tint(.green)
                            .controlSize(.large)
                            .font(.headline.weight(.black))
                        } else {
                            Text("Aguardando o host iniciar a partida...")
                                .font(.caption)
                                .foregroundColor(.white.opacity(0.5))
                        }
                    }
                    .frame(width: 380)
                    
                    // Right Column: Chat and Events
                    VStack(alignment: .leading, spacing: 10) {
                        Text("Chat e Eventos")
                            .font(.headline)
                            .foregroundColor(.white)
                        
                        ScrollViewReader { proxy in
                            ScrollView {
                                VStack(alignment: .leading, spacing: 6) {
                                    ForEach(store.events) { ev in
                                        if ev.kind == "chat" {
                                            HStack(alignment: .top) {
                                                Text("\(ev.payload?.author ?? "Alguém"):")
                                                    .font(.caption.bold())
                                                    .foregroundColor(.cyan)
                                                Text(ev.payload?.text ?? "")
                                                    .font(.caption)
                                                    .foregroundColor(.white)
                                            }
                                        } else if ev.kind == "system" {
                                            Text(ev.payload?.text ?? "")
                                                .font(.caption.italic())
                                                .foregroundColor(.gray)
                                        } else if ev.kind == "replacement_invite" {
                                            Text("Link de Substituição (\(ev.payload?.target_seat ?? 0)): \(ev.payload?.invite_key ?? "")")
                                                .font(.caption.italic())
                                                .foregroundColor(.yellow)
                                        }
                                    }
                                }
                                .padding(8)
                                .frame(maxWidth: .infinity, alignment: .leading)
                            }
                            .background(Color.black.opacity(0.4))
                            .cornerRadius(8)
                            .onChange(of: store.events.count) { _ in
                                if let last = store.events.last {
                                    withAnimation { proxy.scrollTo(last.id, anchor: .bottom) }
                                }
                            }
                        }
                        
                        HStack {
                            TextField("Digite uma mensagem...", text: $chatMessage)
                                .textFieldStyle(.roundedBorder)
                                .onSubmit {
                                    if !chatMessage.isEmpty {
                                        store.sendChat(text: chatMessage)
                                        chatMessage = ""
                                    }
                                }
                            Button("Enviar") {
                                if !chatMessage.isEmpty {
                                    store.sendChat(text: chatMessage)
                                    chatMessage = ""
                                }
                            }
                            .buttonStyle(.borderedProminent)
                            .disabled(chatMessage.isEmpty)
                        }
                    }
                    .frame(width: 320, height: 350)
                }
                
                Button("Sair da sala") {
                    store.dispatchIntent(json: "{\"kind\":\"close_session\"}")
                }
                .buttonStyle(.bordered)
                .tint(.red)
            }
            .padding(40)
        }
    }
}
