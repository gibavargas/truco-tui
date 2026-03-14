import SwiftUI

struct GameView: View {
    let snapshot: MatchSnapshot?
    @EnvironmentObject var store: TrucoAppStore

    @State private var lastTrickSeqViewed: Int = -1
    @State private var showingTrickEndAnimation = false
    @State private var trickAnimOffset: CGSize = .zero
    @State private var trickWinnerTeam: Int = -1
    @State private var trickTie: Bool = false

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
    
    var body: some View {
        if let snap = snapshot {
            let actions = store.bundle?.ui?.actions
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
                .zIndex(50)
                .overlay(
                    VStack {
                        HStack {
                            Button(action: {
                                store.closeSession()
                            }) {
                                HStack {
                                    Image(systemName: "chevron.left")
                                    Text("Sair da Partida")
                                        .fontWeight(.bold)
                                }
                            }
                            .buttonStyle(.plain)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 8)
                            .background(Color.white.opacity(0.15))
                            .foregroundColor(.white)
                            .clipShape(Capsule())
                            .overlay(Capsule().stroke(Color.white.opacity(0.3), lineWidth: 1))
                            .shadow(radius: 4)
                            .padding(.leading, 30)
                            .padding(.top, 30)
                            
                            Spacer()
                        }
                        Spacer()
                    }, alignment: .topLeading
                )
                
                // Game Log (top-right)
                VStack {
                    HStack {
                        Spacer()
                        LogView(logs: snap.Logs ?? [])
                    }
                    .padding(.top, 120)
                    .padding(.trailing, 50)
                    Spacer()
                }
                
                // Players & Center Table
                VStack(spacing: 0) {
                    if let opponent = seatPlayer(snap, offset: snap.NumPlayers == 4 ? 2 : 1) {
                        OpponentView(player: opponent)
                            .padding(.top, 60)
                    }
                    
                    Spacer()
                    
                    if let center = snap.CurrentHand {
                        CenterTableView(hand: center, players: snap.Players ?? [])
                    }
                    
                    Spacer()
                    
                    if let me = snap.Players?.first(where: { $0.playerID == snap.CurrentPlayerIdx }) {
                        VStack(spacing: 24) {
                            // Action Buttons (Truco, Accept, Refuse)
                            if actions?.must_respond == true {
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
                            } else if actions?.can_ask_or_raise == true {
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
                            
                            PlayerHandView(player: me, isMyTurn: actions?.can_play_card == true)
                        }
                        .padding(.bottom, 60)
                    }
                }

                if snap.NumPlayers == 4 {
                    HStack {
                        if let left = seatPlayer(snap, offset: 3) {
                            SideOpponentView(player: left)
                                .frame(maxWidth: 120)
                                .padding(.leading, 48)
                        }
                        Spacer()
                        if let right = seatPlayer(snap, offset: 1) {
                            SideOpponentView(player: right)
                                .frame(maxWidth: 120)
                                .padding(.trailing, 48)
                        }
                    }
                    .padding(.vertical, 220)
                }

                if store.mode == "host_match" || store.mode == "client_match" {
                    VStack {
                        HStack {
                            Spacer()
                            ScrollView {
                                VStack(alignment: .leading, spacing: 6) {
                                    ForEach(store.events.suffix(10)) { event in
                                        if event.kind == "chat" {
                                            Text("\(event.payload?.author ?? "?"): \(event.payload?.text ?? "")")
                                                .font(.caption)
                                        } else if event.kind == "system" {
                                            Text(event.payload?.text ?? "")
                                                .font(.caption)
                                                .foregroundColor(.secondary)
                                        } else if event.kind == "replacement_invite" {
                                            Text("Link de subs: \(event.payload?.invite_key ?? "")")
                                                .font(.caption)
                                                .foregroundColor(.green)
                                        }
                                    }
                                }
                                .frame(maxWidth: .infinity, alignment: .leading)
                            }
                            .frame(width: 250, height: 140)
                            .padding(10)
                            .background(Color.black.opacity(0.28))
                            .cornerRadius(12)
                            .padding(.trailing, 24)
                        }
                        .padding(.top, 120)
                        Spacer()
                    }
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
                
                // Trick end animation overlay
                if showingTrickEndAnimation {
                    ZStack {
                        // Cards gathering and flying
                        if !trickTie {
                            ZStack {
                                ForEach(0..<4, id: \.self) { i in
                                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                                        .rotationEffect(.degrees(Double(i * 12 - 18)))
                                        .offset(x: CGFloat(i * 6 - 9), y: CGFloat(i * -4))
                                }
                            }
                            .offset(trickAnimOffset)
                            .opacity(trickAnimOffset == .zero ? 1 : 0)
                            .scaleEffect(trickAnimOffset == .zero ? 1 : 0.4)
                            .padding(.top, 100) // Start closer to the center table
                            .zIndex(1)
                        }
                        
                        // Emoji message in the center
                        VStack {
                            if trickTie {
                                Text("😐").font(.system(size: 80))
                                Text("EMPATE!").font(.title.weight(.heavy)).foregroundColor(.white)
                            } else if let myPlayer = snap.Players?.first(where: { $0.playerID == snap.CurrentPlayerIdx }), trickWinnerTeam == myPlayer.Team {
                                Text("🎉").font(.system(size: 80))
                                Text("VOCÊ VENCEU A VAZA!").font(.title.weight(.heavy)).foregroundColor(.green)
                            } else {
                                Text("😢").font(.system(size: 80))
                                Text("ELES VENCERAM").font(.title.weight(.heavy)).foregroundColor(.red)
                            }
                        }
                        .padding(24)
                        .background(Color.black.opacity(0.85))
                        .cornerRadius(24)
                        .overlay(
                            RoundedRectangle(cornerRadius: 24)
                                .stroke(Color.white.opacity(0.2), lineWidth: 2)
                        )
                        .shadow(radius: 20)
                        .zIndex(10)
                    }
                    .zIndex(100)
                    .allowsHitTesting(false)
                    .transition(.opacity.animation(.easeInOut(duration: 0.2)))
                }
            }
            .onChange(of: snap.LastTrickSeq) { newSeq in
                guard let seq = newSeq, seq > 0 else { return }
                if lastTrickSeqViewed == -1 {
                    lastTrickSeqViewed = seq
                    return
                }
                if seq > lastTrickSeqViewed {
                    lastTrickSeqViewed = seq
                    triggerTrickAnimation(snap: snap)
                }
            }
        } else {
            ProgressView("Carregando snapshot...")
                .scaleEffect(1.5)
        }
    }
    
    private func triggerTrickAnimation(snap: MatchSnapshot) {
        let localId = snap.CurrentPlayerIdx ?? 0
        let winnerId = snap.LastTrickWinner ?? -1
        let numPlayers = snap.NumPlayers ?? 2
        
        trickWinnerTeam = snap.LastTrickTeam ?? -1
        trickTie = snap.LastTrickTie ?? false
        
        var target: CGSize = .zero
        if !trickTie && winnerId >= 0 {
            let diff = (winnerId - localId + numPlayers) % numPlayers
            if numPlayers == 2 {
                target = (diff == 0) ? CGSize(width: 0, height: 400) : CGSize(width: 0, height: -400)
            } else {
                switch diff {
                case 0: target = CGSize(width: 0, height: 500)
                case 1: target = CGSize(width: 600, height: 0)
                case 2: target = CGSize(width: 0, height: -500)
                case 3: target = CGSize(width: -600, height: 0)
                default: break
                }
            }
        }
        
        trickAnimOffset = .zero
        showingTrickEndAnimation = false
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
            withAnimation(.spring(response: 0.4, dampingFraction: 0.7)) {
                showingTrickEndAnimation = true
            }
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.2) {
                withAnimation(.easeIn(duration: 0.4)) {
                    trickAnimOffset = target
                }
            }
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.8) {
                withAnimation(.easeOut(duration: 0.2)) {
                    showingTrickEndAnimation = false
                }
            }
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

private struct LogView: View {
    let logs: [String]
    
    var body: some View {
        VStack(alignment: .trailing, spacing: 4) {
            ForEach(logs.suffix(5), id: \.self) { log in
                Text(log)
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.7))
                    .lineLimit(1)
            }
        }
        .padding(12)
        .background(Color.black.opacity(0.3))
        .cornerRadius(12)
        .frame(maxWidth: 280)
    }
}

private struct CenterTableView: View {
    let hand: HandState
    let players: [Player]
    
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
                    let winId = hand.winningCardId
                    ForEach(Array(played.enumerated()), id: \.element.id) { index, pc in
                        let isWinning = (pc.id == winId)
                        let playerName = players.first(where: { $0.playerID == pc.PlayerID })?.Name ?? "Jogador"
                        
                        ZStack {
                            CardView(card: pc.Card)
                                .overlay(
                                    RoundedRectangle(cornerRadius: 12)
                                        .stroke(Color.yellow, lineWidth: isWinning ? 3 : 0)
                                        .shadow(color: .yellow, radius: isWinning ? 8 : 0)
                                )
                                .rotationEffect(.degrees(Double(index * 15 - 10)))
                            
                            Text((isWinning ? "🏆 " : "") + playerName.uppercased())
                                .font(.caption2.bold())
                                .foregroundColor(isWinning ? .yellow : .white)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 4)
                                .background(Color.black.opacity(0.8))
                                .clipShape(Capsule())
                                .shadow(color: .black, radius: 2)
                                .offset(y: -65) // Float slightly above the card
                                .zIndex(20)
                        }
                        .offset(x: CGFloat(index * 15 - 5), y: CGFloat(index * -10))
                        .transition(.move(edge: .bottom).combined(with: .opacity))
                        .zIndex(isWinning ? 10 : Double(index))
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
    
    var body: some View {
        VStack(spacing: 16) {
            Text(player.Name.uppercased())
                .font(.subheadline.bold())
                .foregroundColor(.white)
                .tracking(1)
                .padding(.horizontal, 20)
                .padding(.vertical, 6)
                .background(Color.black.opacity(0.5))
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

    var body: some View {
        VStack(spacing: 14) {
            Text(player.Name.uppercased())
                .font(.caption.bold())
                .foregroundColor(.white.opacity(0.85))
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
    let isMyTurn: Bool
    @EnvironmentObject var store: TrucoAppStore
    
    @State private var hoveredCard: String?
    
    var body: some View {
        VStack(spacing: 20) {
            if isMyTurn {
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
                        CardView(card: card)
                            .offset(y: hoveredCard == card.Rank + card.Suit ? -30 : 0)
                            .animation(.interactiveSpring(response: 0.3, dampingFraction: 0.6), value: hoveredCard)
                            .onHover { isHovered in
                                if isHovered && isMyTurn {
                                    hoveredCard = card.Rank + card.Suit
                                } else if hoveredCard == card.Rank + card.Suit {
                                    hoveredCard = nil
                                }
                            }
                            .onTapGesture {
                                if isMyTurn {
                                    store.dispatchGameAction(action: "play", cardIndex: index)
                                }
                            }
                    }
                }
            }
        }
    }
}
