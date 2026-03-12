import SwiftUI

@main
struct TrucoMacApp: App {
    @StateObject private var store = TrucoAppStore()
    @AppStorage("darkMode") private var darkMode = true

    var body: some Scene {
        WindowGroup("Truco") {
            ContentView()
                .environmentObject(store)
                .preferredColorScheme(darkMode ? .dark : .light)
        }
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("Nova Partida") {
                    store.mode = "idle"
                }
                .keyboardShortcut("n")
            }
            CommandGroup(after: .appSettings) {
                Toggle("Modo Escuro", isOn: $darkMode)
                    .keyboardShortcut("d", modifiers: [.command, .shift])
            }
        }

        Settings {
            VStack(spacing: 16) {
                Toggle("Modo Escuro", isOn: $darkMode)
                Text("Truco Paulista — Edição Nativa macOS")
                    .font(.caption).foregroundColor(.secondary)
            }
            .padding(24)
            .frame(width: 300)
        }
    }
}

struct ContentView: View {
    @EnvironmentObject var store: TrucoAppStore

    var body: some View {
        ZStack {
            if store.mode == "offline_match" || store.mode == "host_match" || store.mode == "client_match" || store.mode == "match_over" {
                TabView {
                    GameView(snapshot: store.snapshot)
                        .tabItem { Label("Mesa", systemImage: "suit.spade.fill") }
                        .tag("mesa")
                    
                    LogPanelView()
                        .tabItem { Label("Log", systemImage: "doc.text") }
                        .tag("log")
                }
                .transition(.opacity)
                .animation(.easeInOut, value: store.mode)
            } else if store.mode == "host_lobby" || store.mode == "client_lobby" {
                OnlineLobbyView()
                    .transition(.opacity)
                    .animation(.easeInOut, value: store.mode)
            } else {
                SetupView()
                    .transition(.opacity)
                    .animation(.easeInOut, value: store.mode)
            }
        }
    }
}

// MARK: - Log Panel
struct LogPanelView: View {
    @EnvironmentObject var store: TrucoAppStore
    
    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack {
                Image(systemName: "doc.text")
                    .foregroundColor(.yellow)
                Text("EVENT LOG")
                    .font(.headline.bold())
                    .tracking(2)
                Spacer()
                Text("\(logEntries.count) eventos")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            .padding()
            .background(Color.black.opacity(0.3))
            
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 4) {
                        ForEach(Array(logEntries.enumerated()), id: \.offset) { idx, entry in
                            HStack(alignment: .top, spacing: 8) {
                                Text("\(idx + 1)")
                                    .font(.system(size: 10, design: .monospaced))
                                    .foregroundColor(.gray)
                                    .frame(width: 30, alignment: .trailing)
                                
                                Text(entry)
                                    .font(.system(size: 12, design: .monospaced))
                                    .foregroundColor(entryColor(entry))
                            }
                            .padding(.horizontal)
                            .padding(.vertical, 2)
                            .id(idx)
                        }
                    }
                    .padding(.vertical, 8)
                }
                .onChange(of: logEntries.count) { _ in
                    if let last = logEntries.indices.last {
                        proxy.scrollTo(last, anchor: .bottom)
                    }
                }
            }
        }
        .background(Color(white: 0.08))
    }
    
    private var logEntries: [String] {
        var entries: [String] = []
        if let logs = store.snapshot?.Logs {
            entries.append(contentsOf: logs)
        }
        if let diag = store.bundle?.diagnostics?.event_log {
            entries.append(contentsOf: diag)
        }
        return entries
    }
    
    private func entryColor(_ entry: String) -> Color {
        if entry.lowercased().contains("erro") || entry.lowercased().contains("error") { return .red }
        if entry.lowercased().contains("truco") || entry.lowercased().contains("seis") || entry.lowercased().contains("nove") || entry.lowercased().contains("doze") { return .yellow }
        if entry.lowercased().contains("vitória") || entry.lowercased().contains("ganhou") { return .green }
        return .white.opacity(0.8)
    }
}

// MARK: - Setup View
struct SetupView: View {
    @EnvironmentObject var store: TrucoAppStore
    @AppStorage("darkMode") private var darkMode = true
    
    @State private var numPlayers = 2
    @State private var playerName = "Voce"
    @State private var selectedLocale = "pt-BR"
    @State private var inviteKey = ""
    @State private var logoScale: CGFloat = 0.8
    @State private var logoOpacity: Double = 0
    
    private let locales = [("pt-BR", "Português"), ("en-US", "English")]
    
    var body: some View {
        ZStack {
            Color(white: darkMode ? 0.08 : 0.92).ignoresSafeArea()
            
            VStack(spacing: 28) {
                Spacer()
                
                // Logo with entrance animation
                VStack(spacing: 12) {
                    HStack(spacing: 8) {
                        Text("♠").font(.system(size: 50))
                        Text("♥").font(.system(size: 50)).foregroundColor(.red)
                    }
                    Text("TRUCO PAULISTA")
                        .font(.system(size: 36, weight: .black, design: .rounded))
                        .foregroundColor(darkMode ? .white : .black)
                        .tracking(4)
                    Text("EDIÇÃO NATIVA macOS")
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .tracking(2)
                }
                .scaleEffect(logoScale)
                .opacity(logoOpacity)
                .onAppear {
                    withAnimation(.spring(response: 0.8, dampingFraction: 0.6)) {
                        logoScale = 1.0
                        logoOpacity = 1.0
                    }
                }
                
                // Setup Card
                VStack(spacing: 16) {
                    setupField(label: "SEU NOME") {
                        TextField("Nome", text: $playerName)
                            .textFieldStyle(.roundedBorder)
                    }
                    
                    setupField(label: "JOGADORES") {
                        Picker("", selection: $numPlayers) {
                            Text("2 Jogadores").tag(2)
                            Text("4 Jogadores").tag(4)
                        }
                        .pickerStyle(.segmented)
                    }
                    
                    setupField(label: "IDIOMA") {
                        Picker("", selection: $selectedLocale) {
                            ForEach(locales, id: \.0) { code, label in
                                Text(label).tag(code)
                            }
                        }
                        .pickerStyle(.segmented)
                    }
                    
                    // Theme toggle
                    HStack {
                        Text("TEMA")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(.secondary)
                            .tracking(1)
                        Spacer()
                        Toggle(darkMode ? "🌙 Escuro" : "☀️ Claro", isOn: $darkMode)
                            .toggleStyle(.switch)
                            .tint(.yellow)
                    }
                    
                    // Player Roster
                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(0..<numPlayers, id: \.self) { i in
                            HStack {
                                Image(systemName: i == 0 ? "person.fill" : "cpu")
                                    .foregroundColor(i == 0 ? .yellow : .gray)
                                    .frame(width: 20)
                                Text(playerNameFor(index: i))
                                    .font(.system(size: 13))
                                Spacer()
                                Text(i == 0 ? "Humano" : "CPU")
                                    .font(.system(size: 10))
                                    .foregroundColor(.secondary)
                                Text("Time \(i % 2 + 1)")
                                    .font(.system(size: 10, weight: .bold))
                                    .foregroundColor(i % 2 == 0 ? .blue : .orange)
                                    .padding(.horizontal, 6).padding(.vertical, 2)
                                    .background((i % 2 == 0 ? Color.blue : Color.orange).opacity(0.15))
                                    .cornerRadius(4)
                            }
                            .padding(.horizontal, 10).padding(.vertical, 5)
                            .background(Color.primary.opacity(0.04))
                            .cornerRadius(6)
                        }
                    }
                }
                .frame(maxWidth: 280)
                .padding(28)
                .background(.ultraThinMaterial)
                .cornerRadius(20)
                .shadow(color: .black.opacity(0.2), radius: 20)
                
                // Start / Online Actions
                VStack(spacing: 12) {
                    Button(action: startGame) {
                        HStack(spacing: 8) {
                            Image(systemName: "play.fill")
                            Text("JOGAR OFFLINE")
                                .font(.headline.weight(.black))
                        }
                        .foregroundColor(.black)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(Color.yellow)
                        .cornerRadius(8)
                    }
                    .buttonStyle(.plain)
                    
                    Button(action: {
                        let players = numPlayers == 2 ? 2 : 4
                        store.createHost(name: playerName.isEmpty ? "Voce" : playerName, numPlayers: players, relayURL: nil)
                    }) {
                        Text("CRIAR SALA (HOST)")
                            .font(.headline.weight(.bold))
                            .foregroundColor(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 12)
                            .background(Color.white.opacity(0.1))
                            .cornerRadius(8)
                    }
                    .buttonStyle(.plain)
                    
                    HStack {
                        TextField("Código", text: $inviteKey)
                            .textFieldStyle(.roundedBorder)
                        Button("Entrar") {
                            store.joinSession(name: playerName.isEmpty ? "Voce" : playerName, key: inviteKey)
                        }
                        .disabled(inviteKey.isEmpty)
                        .buttonStyle(.bordered)
                    }
                }
                .frame(maxWidth: 280)
                .scaleEffect(logoScale)
                .animation(.spring(response: 0.3), value: logoScale)
                
                Spacer()
                
                Text(store.status)
                    .font(.caption).foregroundColor(.secondary)
                    .padding(.bottom, 16)
            }
        }
        .onAppear { store.setLocale(selectedLocale) }
        .onChange(of: selectedLocale) { nv in store.setLocale(nv) }
    }
    
    @ViewBuilder
    private func setupField<Content: View>(label: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.system(size: 10, weight: .bold))
                .foregroundColor(.secondary)
                .tracking(1)
            content()
        }
    }
    
    private func playerNameFor(index: Int) -> String {
        if index == 0 { return playerName.isEmpty ? "Voce" : playerName }
        if numPlayers == 2 { return "CPU-Oponente" }
        switch index {
        case 1: return "CPU-Direita"
        case 2: return "CPU-Parceiro"
        case 3: return "CPU-Esquerda"
        default: return "CPU-\(index)"
        }
    }
    
    private func startGame() {
        let names = (0..<numPlayers).map { playerNameFor(index: $0) }
        let cpus = (0..<numPlayers).map { $0 != 0 }
        store.setLocale(selectedLocale)
        store.startOfflineGame(playerNames: names, cpuFlags: cpus)
    }
}
