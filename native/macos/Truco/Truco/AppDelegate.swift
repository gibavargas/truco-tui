//
//  AppDelegate.swift
//  Truco
//
//  Created by Joao Vitor Guidi on 10/03/26.
//

import SwiftUI

@main
struct TrucoApp: App {
    @StateObject private var store = TrucoAppStore()

    var body: some Scene {
        WindowGroup("Truco") {
            ContentView()
                .environmentObject(store)
                .frame(minWidth: 800, minHeight: 600)
        }
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("Nova Partida Offline") {
                    store.startOfflineDemo()
                }
                .keyboardShortcut("n")
                
                Button("Sair da Sala") {
                    store.dispatchIntent(json: "{\"kind\":\"close_session\"}")
                }
                .keyboardShortcut("w")
            }
        }
    }
}

struct ContentView: View {
    @EnvironmentObject var store: TrucoAppStore

    var body: some View {
        Group {
            if store.mode.contains("match") {
                // In-game (offline_match, host_match, client_match)
                GameView(snapshot: store.snapshot)
                    .environmentObject(store)
            } else if store.mode.contains("lobby") {
                // Online lobby (host_lobby, client_lobby)
                OnlineLobbyView()
                    .environmentObject(store)
            } else {
                // idle
                LobbyView()
                    .environmentObject(store)
            }
        }
        .transition(.opacity)
        .animation(.easeInOut(duration: 0.3), value: store.mode)
    }
}
