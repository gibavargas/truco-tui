import SwiftUI

@main
struct TrucoMacApp: App {
    @StateObject private var store = TrucoAppStore()

    var body: some Scene {
        WindowGroup("Truco") {
            NavigationSplitView {
                List {
                    Text("Lobby")
                    Text("Mesa")
                    Text("Diagnosticos")
                }
                .navigationTitle("Truco")
            } detail: {
                ContentView()
                    .environmentObject(store)
            }
        }
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("Nova Partida") {
                    store.startOfflineDemo()
                }
                .keyboardShortcut("n")
            }
        }

        Settings {
            Text("Configuracoes do Truco nativo")
                .padding()
        }
    }
}

struct ContentView: View {
    @EnvironmentObject var store: TrucoAppStore

    var body: some View {
        if store.mode == "offline_match" {
            GameView(snapshot: store.snapshot)
                .transition(.opacity)
                .animation(.easeInOut, value: store.mode)
        } else {
            VStack(alignment: .center, spacing: 24) {
                Image(systemName: "suit.spade.fill")
                    .font(.system(size: 80))
                    .foregroundColor(.white.opacity(0.2))
                    
                Text(store.status)
                    .font(.title3)
                    .foregroundStyle(.secondary)
                    
                Button("Criar partida offline de teste") {
                    store.startOfflineDemo()
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
                .tint(.yellow)
                .foregroundColor(.black)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color(white: 0.12))
        }
    }
}
