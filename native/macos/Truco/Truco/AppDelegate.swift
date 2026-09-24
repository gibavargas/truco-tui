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
            CommandMenu("Mesa") {
                Button("Nova mesa offline") {
                    store.startOfflineDemo()
                }
                .keyboardShortcut("n")

                Button("Repetir última mesa") {
                    store.replayOfflineMatch()
                }
                .keyboardShortcut("r")

                Button("Começar partida hospedada") {
                    store.startHostedMatch()
                }
                .keyboardShortcut(.return, modifiers: [.command])
                .disabled(store.mode != "host_lobby")

                Button("Sair da mesa") {
                    store.closeSession()
                }
                .keyboardShortcut(".", modifiers: [.command])
                .disabled(!store.canCloseSession)
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
                .frame(
                    minWidth: 840,
                    idealWidth: launchWindowSize.width,
                    minHeight: 620,
                    idealHeight: launchWindowSize.height
                )
        } else {
            content
                .frame(minWidth: 840, minHeight: 620)
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
