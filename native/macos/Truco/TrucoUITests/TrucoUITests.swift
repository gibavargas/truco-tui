//
//  TrucoUITests.swift
//  TrucoUITests
//
//  Created by Joao Vitor Guidi on 10/03/26.
//

import XCTest

final class TrucoUITests: XCTestCase {
    private let outputDir = FileManager.default.temporaryDirectory.appendingPathComponent("truco-macos-ui-review", isDirectory: true)

    override func setUpWithError() throws {
        continueAfterFailure = false
        try FileManager.default.createDirectory(at: outputDir, withIntermediateDirectories: true)
    }

    @MainActor
    func testRenderNativeScreens() throws {
        try captureMainLobby(width: 1320, height: 900, colorScheme: "light", name: "menu-large-light")
        try captureMainLobby(width: 860, height: 640, colorScheme: "dark", name: "menu-small-dark")
        try captureOfflineSetup(width: 1120, height: 840, colorScheme: "light", name: "offline-setup-light")
        try captureOnlineMenu(width: 1120, height: 840, colorScheme: "dark", name: "online-menu-dark")
        try captureOnlineLobby(width: 1320, height: 900, colorScheme: "dark", name: "online-lobby-dark")
        try captureOnlineLobby(width: 920, height: 700, colorScheme: "light", name: "online-lobby-small-light")
        try captureOfflineMatch(width: 1440, height: 980, colorScheme: "dark", name: "offline-match-large-dark")
        try captureOfflineMatch(width: 980, height: 720, colorScheme: "light", name: "offline-match-small-light")
    }

    private func launchApp(width: Int, height: Int, colorScheme: String) -> XCUIApplication {
        let app = XCUIApplication()
        app.launchEnvironment["TRUCO_WINDOW_WIDTH"] = "\(width)"
        app.launchEnvironment["TRUCO_WINDOW_HEIGHT"] = "\(height)"
        app.launchEnvironment["TRUCO_COLOR_SCHEME"] = colorScheme
        app.launch()
        return app
    }

    private func captureMainLobby(width: Int, height: Int, colorScheme: String, name: String) throws {
        let app = launchApp(width: width, height: height, colorScheme: colorScheme)
        XCTAssertTrue(app.buttons["Jogar Offline"].waitForExistence(timeout: 5))
        saveScreenshot(app, name: name)
        app.terminate()
    }

    private func captureOfflineSetup(width: Int, height: Int, colorScheme: String, name: String) throws {
        let app = launchApp(width: width, height: height, colorScheme: colorScheme)
        XCTAssertTrue(app.buttons["Jogar Offline"].waitForExistence(timeout: 5))
        app.buttons["Jogar Offline"].click()
        XCTAssertTrue(app.buttons["Iniciar Partida"].waitForExistence(timeout: 5))
        saveScreenshot(app, name: name)
        app.terminate()
    }

    private func captureOnlineMenu(width: Int, height: Int, colorScheme: String, name: String) throws {
        let app = launchApp(width: width, height: height, colorScheme: colorScheme)
        XCTAssertTrue(app.buttons["Jogar Online"].waitForExistence(timeout: 5))
        app.buttons["Jogar Online"].click()
        XCTAssertTrue(app.buttons["Criar Sala (Host)"].waitForExistence(timeout: 5))
        saveScreenshot(app, name: name)
        app.terminate()
    }

    private func captureOnlineLobby(width: Int, height: Int, colorScheme: String, name: String) throws {
        let app = launchApp(width: width, height: height, colorScheme: colorScheme)
        XCTAssertTrue(app.buttons["Jogar Online"].waitForExistence(timeout: 5))
        app.buttons["Jogar Online"].click()
        XCTAssertTrue(app.buttons["Criar Sala (Host)"].waitForExistence(timeout: 5))
        app.buttons["Criar Sala (Host)"].click()
        XCTAssertTrue(app.buttons["Criar"].waitForExistence(timeout: 5))
        app.buttons["Criar"].click()
        XCTAssertTrue(app.staticTexts["Chave de convite:"].waitForExistence(timeout: 5))
        saveScreenshot(app, name: name)
        app.terminate()
    }

    private func captureOfflineMatch(width: Int, height: Int, colorScheme: String, name: String) throws {
        let app = launchApp(width: width, height: height, colorScheme: colorScheme)
        XCTAssertTrue(app.buttons["Jogar Offline"].waitForExistence(timeout: 5))
        app.buttons["Jogar Offline"].click()
        XCTAssertTrue(app.buttons["Iniciar Partida"].waitForExistence(timeout: 5))
        app.buttons["Iniciar Partida"].click()
        XCTAssertTrue(app.buttons["Sair da Partida"].waitForExistence(timeout: 5))
        saveScreenshot(app, name: name)
        app.terminate()
    }

    private func saveScreenshot(_ app: XCUIApplication, name: String) {
        let screenshot = app.screenshot()
        let attachment = XCTAttachment(screenshot: screenshot)
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
        let url = outputDir.appendingPathComponent("\(name).png")
        try? screenshot.pngRepresentation.write(to: url)
    }
}
