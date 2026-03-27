import SwiftUI

struct GameView: View {
    let snapshot: MatchSnapshot?
    @EnvironmentObject var store: TrucoAppStore
    @State private var highlightedPlayerID: Int?
    @State private var trickHighlightToken = 0

    private func seatPlayer(_ snap: MatchSnapshot, offset: Int) -> Player? {
        guard let players = snap.Players, let localID = snap.CurrentPlayerIdx, let count = snap.NumPlayers else { return nil }
        let seatID = (localID + offset) % count
        return players.first(where: { $0.playerID == seatID })
    }
    
    private func raiseLabel(for stake: Int) -> String {
        switch stake {
        case 3: return "SEIS!"
        case 6: return "NOVE!"
        case 9: return "DOZE!"
        case 12: return "DOZE!"
        default: return "TRUCO!"
        }
    }

    private func canPlayCard(_ snap: MatchSnapshot, player: Player) -> Bool {
        snap.MatchFinished != true &&
        snap.TurnPlayer == player.playerID &&
        (snap.PendingRaiseFor ?? -1) == -1
    }

    private func resolvedHighlightedPlayerID(_ snap: MatchSnapshot) -> Int? {
        highlightedPlayerID ?? snap.TurnPlayer
    }

    private func updateTrickHighlight(for snap: MatchSnapshot) {
        guard snap.LastTrickSeq ?? 0 > 0 else {
            highlightedPlayerID = nil
            return
        }
        if snap.LastTrickTie == true || (snap.LastTrickWinner ?? -1) < 0 {
            highlightedPlayerID = nil
            return
        }
        trickHighlightToken += 1
        let token = trickHighlightToken
        highlightedPlayerID = snap.LastTrickWinner
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.4) {
            if trickHighlightToken == token {
                highlightedPlayerID = nil
            }
        }
    }
    
    var body: some View {
        if let snap = snapshot {
            let highlightedID = resolvedHighlightedPlayerID(snap)
            let isOnlineMatch = store.bundle?.connection?.is_online == true || store.mode == "host_match" || store.mode == "client_match"
            ZStack {
                // Background wood panels
                HStack(spacing: 0) {
                    ForEach(0..<8, id: \.self) { i in
                        Rectangle()
                            .fill(LinearGradient(
                                colors: [
                                    Color(red: 0.38 - Double(i % 3) * 0.02, green: 0.24 - Double(i % 3) * 0.02, blue: 0.14 - Double(i % 3) * 0.01),
                                    Color(red: 0.28, green: 0.16, blue: 0.08)
                                ],
                                startPoint: .top,
                                endPoint: .bottom
                            ))
                            .overlay(
                                Rectangle()
                                    .fill(
                                        LinearGradient(
                                            colors: [Color.white.opacity(i % 2 == 0 ? 0.05 : 0), Color.clear],
                                            startPoint: .leading,
                                            endPoint: .trailing
                                        )
                                    )
                            )
                            .overlay(
                                HStack {
                                    Spacer()
                                    Rectangle().fill(Color.black.opacity(0.3)).frame(width: 2)
                                }
                            )
                    }
                }
                .ignoresSafeArea()
                
                // Table Felt
                RoundedRectangle(cornerRadius: 120, style: .continuous)
                    .fill(RadialGradient(
                        colors: [Color(red: 0.12, green: 0.45, blue: 0.22), Color(red: 0.05, green: 0.20, blue: 0.10)],
                        center: .center,
                        startRadius: 50,
                        endRadius: 500
                    ))
                    .padding(32)
                    .shadow(color: .black.opacity(0.6), radius: 40, x: 0, y: 20)
                    .overlay(
                        RoundedRectangle(cornerRadius: 120, style: .continuous)
                            .strokeBorder(Color.white.opacity(0.15), lineWidth: 2)
                            .padding(32)
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: 120, style: .continuous)
                            .stroke(Color.black.opacity(0.4), lineWidth: 20)
                            .blur(radius: 12)
                            .clipShape(RoundedRectangle(cornerRadius: 120, style: .continuous))
                            .padding(32)
                    )
                
                // HUD
                VStack {
                    HStack {
                        ScoreView(teamName: "Nós", points: snap.teamScore.us)
                        Spacer()
                        StakeBadge(stake: snap.CurrentHand?.Stake ?? 1)
                        Spacer()
                        ScoreView(teamName: "Eles", points: snap.teamScore.them)
                    }
                    .padding(50)
                    Spacer()
                }
                
                // Game Log (top-right)
                VStack {
                    HStack {
                        Spacer()
                        MatchActivityPanel(logs: snap.Logs ?? [], events: store.events, isOnline: isOnlineMatch)
                    }
                    .padding(.top, 120)
                    .padding(.trailing, 50)
                    Spacer()
                }
                
                // Players & Center Table
                VStack(spacing: 0) {
                    if let opponent = seatPlayer(snap, offset: snap.NumPlayers == 4 ? 2 : 1) {
                        OpponentView(player: opponent, isHighlighted: highlightedID == opponent.playerID)
                            .padding(.top, 60)
                    }
                    
                    Spacer()
                    
                    if let center = snap.CurrentHand {
                        CenterTableView(hand: center)
                    }
                    
                    Spacer()
                    
                    if let me = snap.Players?.first(where: { $0.playerID == snap.CurrentPlayerIdx }) {
                        VStack(spacing: 24) {
                            // Action Buttons (Truco, Accept, Refuse)
                            if snap.PendingRaiseFor == me.Team {
                                HStack(spacing: 20) {
                                    Button("ACEITAR") {
                                        store.dispatchGameAction(action: "accept")
                                    }
                                    .buttonStyle(.borderedProminent)
                                    .tint(.green)
                                    .controlSize(.large)
                                    .font(.headline.weight(.black))
                                    
                                    Button("CORRER") {
                                        store.dispatchGameAction(action: "refuse")
                                    }
                                    .buttonStyle(.borderedProminent)
                                    .tint(.red)
                                    .controlSize(.large)
                                    .font(.headline.weight(.black))
                                }
                                .padding(.top, 10)
                            } else if snap.TurnPlayer == snap.CurrentPlayerIdx && snap.CanAskTruco == true {
                                let label = snap.PendingRaiseTo != nil ? raiseLabel(for: snap.PendingRaiseTo!) : raiseLabel(for: snap.CurrentHand?.Stake ?? 1)
                                Button(label) {
                                    store.dispatchGameAction(action: "truco")
                                }
                                .buttonStyle(.borderedProminent)
                                .tint(.yellow)
                                .foregroundColor(.black)
                                .controlSize(.large)
                                .font(.headline.weight(.black))
                                .padding(.top, 10)
                            }
                            
                            PlayerHandView(
                                player: me,
                                canPlayCard: canPlayCard(snap, player: me),
                                currentRound: snap.CurrentHand?.Round ?? 1,
                                isHighlighted: highlightedID == me.playerID
                            )
                        }
                        .padding(.bottom, 60)
                    }
                }

                if snap.NumPlayers == 4 {
                    HStack {
                        if let left = seatPlayer(snap, offset: 3) {
                            SideOpponentView(player: left, isHighlighted: highlightedID == left.playerID)
                                .frame(maxWidth: 120)
                                .padding(.leading, 48)
                        }
                        Spacer()
                        if let right = seatPlayer(snap, offset: 1) {
                            SideOpponentView(player: right, isHighlighted: highlightedID == right.playerID)
                                .frame(maxWidth: 120)
                                .padding(.trailing, 48)
                        }
                    }
                    .padding(.vertical, 220)
                }
                
                // Match finished overlay
                if snap.MatchFinished == true {
                    Color.black.opacity(0.7)
                        .ignoresSafeArea()
                    
                    VStack(spacing: 24) {
                        Text("FIM DE JOGO")
                            .font(.system(size: 48, weight: .black, design: .rounded))
                            .foregroundColor(.yellow)
                        
                        let scores = snap.teamScore
                        Text("\(scores.us) x \(scores.them)")
                            .font(.system(size: 36, weight: .bold, design: .rounded))
                            .foregroundColor(.white)
                        
                        Text(snap.WinnerTeam == 0 ? "VOCÊ VENCEU! 🏆" : "VOCÊ PERDEU 😢")
                            .font(.system(size: 28, weight: .bold, design: .rounded))
                            .foregroundColor(snap.WinnerTeam == 0 ? .yellow : .red)
                        
                        Button("NOVA PARTIDA") {
                            store.startOfflineDemo()
                        }
                        .buttonStyle(.borderedProminent)
                        .tint(.yellow)
                        .foregroundColor(.black)
                        .controlSize(.large)
                        .font(.headline.weight(.black))
                    }
                }
            }
            .onAppear {
                highlightedPlayerID = nil
            }
            .onChange(of: snap.LastTrickSeq) { _ in
                updateTrickHighlight(for: snap)
            }
        } else {
            ProgressView("Carregando snapshot...")
                .scaleEffect(1.5)
        }
    }
}

// MARK: - Sub-views

private struct ScoreView: View {
    let teamName: String
    let points: Int
    
    var body: some View {
        VStack {
            Text(teamName.uppercased())
                .font(.caption)
                .fontWeight(.bold)
                .foregroundColor(.white.opacity(0.7))
            Text("\(points)")
                .font(.system(size: 32, weight: .black, design: .rounded))
                .foregroundColor(.white)
        }
        .frame(width: 80, height: 80)
        .background(Color.black.opacity(0.4))
        .cornerRadius(16)
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(Color.white.opacity(0.1), lineWidth: 1)
        )
    }
}

private struct StakeBadge: View {
    let stake: Int
    
    var body: some View {
        VStack {
            Text("VALE")
                .font(.caption)
                .fontWeight(.bold)
                .foregroundColor(.white.opacity(0.7))
            Text("\(stake)")
                .font(.system(size: 28, weight: .black, design: .rounded))
                .foregroundColor(.yellow)
        }
        .frame(width: 90, height: 80)
        .background(Color.black.opacity(0.5))
        .cornerRadius(16)
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(Color.yellow.opacity(0.6), lineWidth: 2)
                .shadow(color: .yellow.opacity(0.4), radius: 5)
        )
    }
}

private struct MatchActivityPanel: View {
    let logs: [String]
    let events: [AppEvent]
    let isOnline: Bool
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(isOnline ? "Atividade Online" : "Log")
                .font(.caption.bold())
                .foregroundColor(.white.opacity(0.7))
            ForEach(entries, id: \.self) { line in
                Text(line)
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.78))
                    .lineLimit(2)
            }
        }
        .padding(12)
        .background(Color.black.opacity(0.3))
        .cornerRadius(12)
        .frame(maxWidth: 280)
    }

    private var entries: [String] {
        var out = Array(logs.suffix(isOnline ? 4 : 6))
        if isOnline {
            for ev in events.suffix(6) {
                if ev.kind == "chat" {
                    let author = ev.payload?.author ?? "Alguém"
                    let text = ev.payload?.text ?? ""
                    out.append("[chat] \(author): \(text)")
                } else if let text = ev.payload?.text, !text.isEmpty {
                    out.append("[\(ev.kind)] \(text)")
                } else {
                    out.append("[\(ev.kind)]")
                }
            }
        }
        return Array(out.suffix(8))
    }
}

private struct CenterTableView: View {
    let hand: HandState
    
    var body: some View {
        HStack(spacing: 60) {
            // Vira
            VStack(spacing: 12) {
                Text("VIRA")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.6))
                    .tracking(2)
                if let vira = hand.Vira {
                    CardView(card: vira)
                        .rotationEffect(.degrees(-6))
                } else {
                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                }
            }
            
            // Played Cards Layer
            ZStack {
                Circle()
                    .fill(Color.white.opacity(0.04))
                    .frame(width: 160, height: 160)
                
                if let played = hand.RoundCards {
                    ForEach(Array(played.enumerated()), id: \.element.id) { index, pc in
                        CardView(card: pc.Card, isFaceUp: pc.FaceDown != true)
                            .rotationEffect(.degrees(Double(index * 15 - 10)))
                            .offset(x: CGFloat(index * 15 - 5), y: CGFloat(index * -10))
                            .transition(.move(edge: .bottom).combined(with: .opacity))
                    }
                }
            }
            .animation(.spring(response: 0.5, dampingFraction: 0.7), value: hand.RoundCards?.count)
            
            // Manilha
            VStack(spacing: 12) {
                Text("MANILHA")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.6))
                    .tracking(2)
                if let manilha = hand.Manilha, manilha != "-" {
                    Text(manilha)
                        .font(.system(size: 38, weight: .black, design: .rounded))
                        .foregroundColor(.yellow)
                        .frame(width: 86, height: 124)
                        .background(Color.yellow.opacity(0.15))
                        .cornerRadius(12)
                        .overlay(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(Color.yellow, lineWidth: 2)
                                .shadow(color: .yellow, radius: 4)
                        )
                } else {
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color.white.opacity(0.05))
                        .frame(width: 86, height: 124)
                }
            }
        }
    }
}

private struct OpponentView: View {
    let player: Player
    let isHighlighted: Bool
    
    var body: some View {
        VStack(spacing: 16) {
            Text(player.Name.uppercased())
                .font(.subheadline.bold())
                .foregroundColor(isHighlighted ? .black : .white)
                .tracking(1)
                .padding(.horizontal, 20)
                .padding(.vertical, 6)
                .background(isHighlighted ? Color.yellow : Color.black.opacity(0.5))
                .clipShape(Capsule())
            
            HStack(spacing: -24) {
                ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                        .shadow(color: .black.opacity(0.3), radius: 4, x: -2, y: 3)
                }
            }
        }
    }
}

private struct SideOpponentView: View {
    let player: Player
    let isHighlighted: Bool

    var body: some View {
        VStack(spacing: 14) {
            Text(player.Name.uppercased())
                .font(.caption.bold())
                .foregroundColor(isHighlighted ? .yellow : .white.opacity(0.85))
                .rotationEffect(.degrees(90))

            VStack(spacing: -20) {
                ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                        .frame(width: 60, height: 88)
                }
            }
        }
    }
}

private struct PlayerHandView: View {
    let player: Player
    let canPlayCard: Bool
    let currentRound: Int
    let isHighlighted: Bool
    @EnvironmentObject var store: TrucoAppStore
    
    @State private var hoveredCard: String?
    
    var body: some View {
        VStack(spacing: 20) {
            Text(player.Name.uppercased())
                .font(.caption.bold())
                .foregroundColor(isHighlighted ? .black : .white.opacity(0.9))
                .tracking(1)
                .padding(.horizontal, 16)
                .padding(.vertical, 6)
                .background(isHighlighted ? Color.yellow : Color.black.opacity(0.45))
                .clipShape(Capsule())

            if canPlayCard {
                Text("SUA VEZ")
                    .font(.caption.bold())
                    .foregroundColor(Color(red: 0.35, green: 0.21, blue: 0.12))
                    .tracking(1.5)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 6)
                    .background(Color.yellow)
                    .clipShape(Capsule())
                    .shadow(color: .yellow.opacity(0.4), radius: 8)
            }
            
	            HStack(spacing: 16) {
	                if let hand = player.Hand {
	                    ForEach(Array(hand.enumerated()), id: \.element) { index, card in
	                        VStack(spacing: 8) {
                            CardView(card: card)
                                .offset(y: hoveredCard == card.Rank + card.Suit ? -30 : 0)
                                .animation(.interactiveSpring(response: 0.3, dampingFraction: 0.6), value: hoveredCard)
                                .onHover { isHovered in
                                    if isHovered && canPlayCard {
                                        hoveredCard = card.Rank + card.Suit
                                    } else if hoveredCard == card.Rank + card.Suit {
                                        hoveredCard = nil
                                    }
                                }
                                .onTapGesture {
                                    if canPlayCard {
                                        store.dispatchGameAction(action: "play", cardIndex: index)
                                    }
                                }

                            if canPlayCard && currentRound >= 2 {
                                Button("Virada") {
                                    store.dispatchGameAction(action: "play", cardIndex: index, faceDown: true)
                                }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                                .tint(.white.opacity(0.25))
	                    }
	                }
	            }
	        }
	    }
	}
    }
}
