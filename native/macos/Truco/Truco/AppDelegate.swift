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
    private let launchWindowSize = LaunchConfiguration.windowSize
    private let launchColorScheme = LaunchConfiguration.colorScheme

    var body: some Scene {
        WindowGroup("Truco") {
            configuredRootView
        }
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("Nova Partida Offline") {
                    store.startOfflineDemo()
                }
                .keyboardShortcut("n")
                
                Button("Sair da Sala") {
                    store.closeSession()
                }
                .keyboardShortcut("w")
            }
        }
    }

    @ViewBuilder
    private var configuredRootView: some View {
        let content = ContentView()
            .environmentObject(store)
            .preferredColorScheme(launchColorScheme)

        if let launchWindowSize {
            content
                .frame(width: launchWindowSize.width, height: launchWindowSize.height)
        } else {
            content
                .frame(minWidth: 800, minHeight: 600)
        }
    }
}

private enum LaunchConfiguration {
    static var windowSize: CGSize? {
        let env = ProcessInfo.processInfo.environment
        guard
            let widthText = env["TRUCO_WINDOW_WIDTH"],
            let heightText = env["TRUCO_WINDOW_HEIGHT"],
            let width = Double(widthText),
            let height = Double(heightText)
        else {
            return nil
        }
        return CGSize(width: width, height: height)
    }

    static var colorScheme: ColorScheme? {
        switch ProcessInfo.processInfo.environment["TRUCO_COLOR_SCHEME"]?.lowercased() {
        case "light":
            return .light
        case "dark":
            return .dark
        default:
            return nil
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
