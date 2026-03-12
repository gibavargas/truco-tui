import SwiftUI

struct GameView: View {
    @EnvironmentObject var store: TrucoAppStore
    let snapshot: MatchSnapshot?
    
    // Animation states
    @State private var trucoFlashOpacity: Double = 0
    @State private var trucoFlashScale: CGFloat = 0.5
    @State private var dealAnimationProgress: [Bool] = [false, false, false]
    @State private var lastPendingRaise: Int? = nil
    
    var body: some View {
        GeometryReader { geo in
            if let snap = snapshot {
                ZStack {
                    // Background
                    tableBackground
                    
                    // HUD
                    VStack {
                        scoreBar(snap: snap)
                        Spacer()
                    }
                    
                    // Game Layout
                    gameLayout(snap: snap, geo: geo)
                    
                    // Match Over Overlay
                    if snap.MatchFinished == true {
                        matchOverOverlay(snap: snap)
                    }
                    
                    // Truco Flash Overlay
                    trucoFlashView(snap: snap)
                }
                .onChange(of: snap.PendingRaiseTo) { newVal in
                    if newVal != nil && newVal != lastPendingRaise {
                        triggerTrucoFlash()
                    }
                    lastPendingRaise = newVal
                }
                .onAppear { animateDeal() }
                .onChange(of: snap.CurrentHand?.Round) { _ in animateDeal() }
            } else {
                ProgressView("Carregando...")
                    .scaleEffect(1.5)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
    }
    
    // MARK: - Truco Flash
    @ViewBuilder
    private func trucoFlashView(snap: MatchSnapshot) -> some View {
        if trucoFlashOpacity > 0 {
            let raiseTo = snap.PendingRaiseTo ?? 3
            let label = raiseLabelFor(stake: raiseTo)
            
            VStack(spacing: 8) {
                Text(label)
                    .font(.system(size: 72, weight: .black, design: .rounded))
                    .foregroundColor(.yellow)
                    .shadow(color: .yellow.opacity(0.8), radius: 20)
                
                if let byPlayer = snap.PendingRaiseBy,
                   let name = snap.Players?.first(where: { $0.ID == byPlayer })?.Name {
                    Text("por \(name)")
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.7))
                }
            }
            .scaleEffect(trucoFlashScale)
            .opacity(trucoFlashOpacity)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color.red.opacity(0.3 * trucoFlashOpacity))
            .allowsHitTesting(false)
        }
    }
    
    private func triggerTrucoFlash() {
        trucoFlashOpacity = 0
        trucoFlashScale = 0.5
        withAnimation(.spring(response: 0.3, dampingFraction: 0.5)) {
            trucoFlashOpacity = 1
            trucoFlashScale = 1.2
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
            withAnimation(.easeOut(duration: 0.5)) {
                trucoFlashOpacity = 0
                trucoFlashScale = 0.8
            }
        }
    }
    
    private func animateDeal() {
        dealAnimationProgress = [false, false, false]
        for i in 0..<3 {
            DispatchQueue.main.asyncAfter(deadline: .now() + Double(i) * 0.15) {
                withAnimation(.spring(response: 0.4, dampingFraction: 0.7)) {
                    dealAnimationProgress[i] = true
                }
            }
        }
    }
    
    // MARK: - Table Background
    private var tableBackground: some View {
        ZStack {
            HStack(spacing: 0) {
                ForEach(0..<8, id: \.self) { i in
                    Rectangle()
                        .fill(LinearGradient(
                            colors: [
                                Color(red: 0.38 - Double(i % 3) * 0.02, green: 0.24 - Double(i % 3) * 0.02, blue: 0.14 - Double(i % 3) * 0.01),
                                Color(red: 0.28, green: 0.16, blue: 0.08)
                            ],
                            startPoint: .top, endPoint: .bottom
                        ))
                        .overlay(HStack { Spacer(); Rectangle().fill(Color.black.opacity(0.3)).frame(width: 2) })
                }
            }
            .ignoresSafeArea()
            
            RoundedRectangle(cornerRadius: 120, style: .continuous)
                .fill(RadialGradient(
                    colors: [Color(red: 0.12, green: 0.45, blue: 0.22), Color(red: 0.05, green: 0.20, blue: 0.10)],
                    center: .center, startRadius: 50, endRadius: 500
                ))
                .padding(32)
                .shadow(color: .black.opacity(0.6), radius: 40, x: 0, y: 20)
        }
    }
    
    // MARK: - Score Bar
    @ViewBuilder
    private func scoreBar(snap: MatchSnapshot) -> some View {
        let stake = snap.CurrentHand?.Stake ?? 1
        let round = snap.CurrentHand?.Round ?? 1
        
        HStack(spacing: 0) {
            ScoreView(teamName: "NÓS", points: snap.MatchPoints?["0"] ?? 0, color: .blue)
            Spacer()
            VStack(spacing: 4) {
                HStack(spacing: 12) {
                    VStack(spacing: 2) {
                        Text("VALE").font(.system(size: 9, weight: .bold)).foregroundColor(.white.opacity(0.5))
                        Text("\(stake)").font(.system(size: 22, weight: .black, design: .rounded)).foregroundColor(.yellow)
                    }
                    HStack(spacing: 3) {
                        ForEach([1, 3, 6, 9, 12], id: \.self) { level in
                            Text(stakeName(level))
                                .font(.system(size: 8, weight: .bold))
                                .foregroundColor(stake >= level ? .yellow : .white.opacity(0.2))
                                .padding(.horizontal, 3).padding(.vertical, 1)
                                .background(stake >= level ? Color.yellow.opacity(0.2) : Color.clear)
                                .cornerRadius(3)
                        }
                    }
                    VStack(spacing: 2) {
                        Text("RODADA").font(.system(size: 9, weight: .bold)).foregroundColor(.white.opacity(0.5))
                        Text("\(round)/3").font(.system(size: 16, weight: .bold, design: .rounded)).foregroundColor(.white)
                    }
                }
                if let tp = snap.TurnPlayer, let player = snap.Players?.first(where: { $0.ID == tp }) {
                    HStack(spacing: 4) {
                        if player.CPU == true {
                            ProgressView().controlSize(.mini).tint(.yellow)
                        }
                        Text("Vez: \(player.Name)")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(.black)
                    }
                    .padding(.horizontal, 12).padding(.vertical, 3)
                    .background(Color.yellow)
                    .clipShape(Capsule())
                }
            }
            Spacer()
            ScoreView(teamName: "ELES", points: snap.MatchPoints?["1"] ?? 0, color: .orange)
        }
        .padding(.horizontal, 50).padding(.top, 40)
    }
    
    // MARK: - Layout
    @ViewBuilder
    private func gameLayout(snap: MatchSnapshot, geo: GeometryProxy) -> some View {
        let localIdx = snap.CurrentPlayerIdx ?? 0
        let n = snap.NumPlayers ?? 2
        let compact = geo.size.width < 800
        
        if n == 4 {
            VStack(spacing: 0) {
                if let top = playerAt(snap: snap, offset: 2, localIdx: localIdx, n: n) {
                    OpponentView(player: top, isTurn: snap.TurnPlayer == top.ID, position: .top, snap: snap, compact: compact)
                        .padding(.top, compact ? 70 : 90)
                }
                Spacer()
                HStack(spacing: 0) {
                    if let left = playerAt(snap: snap, offset: 3, localIdx: localIdx, n: n) {
                        OpponentView(player: left, isTurn: snap.TurnPlayer == left.ID, position: .left, snap: snap, compact: compact)
                            .frame(width: compact ? 90 : 120)
                    }
                    Spacer()
                    if let center = snap.CurrentHand {
                        CenterTableView(hand: center, turnPlayer: snap.TurnPlayer, players: snap.Players, compact: compact)
                    }
                    Spacer()
                    if let right = playerAt(snap: snap, offset: 1, localIdx: localIdx, n: n) {
                        OpponentView(player: right, isTurn: snap.TurnPlayer == right.ID, position: .right, snap: snap, compact: compact)
                            .frame(width: compact ? 90 : 120)
                    }
                }
                Spacer()
                myHandSection(snap: snap, localIdx: localIdx, compact: compact)
            }
        } else {
            VStack(spacing: 0) {
                if let opp = playerAt(snap: snap, offset: 1, localIdx: localIdx, n: n) {
                    OpponentView(player: opp, isTurn: snap.TurnPlayer == opp.ID, position: .top, snap: snap, compact: compact)
                        .padding(.top, compact ? 70 : 90)
                }
                Spacer()
                if let center = snap.CurrentHand {
                    CenterTableView(hand: center, turnPlayer: snap.TurnPlayer, players: snap.Players, compact: compact)
                }
                Spacer()
                myHandSection(snap: snap, localIdx: localIdx, compact: compact)
            }
        }
    }
    
    @ViewBuilder
    private func myHandSection(snap: MatchSnapshot, localIdx: Int, compact: Bool) -> some View {
        if let me = snap.Players?.first(where: { $0.ID == localIdx }) {
            VStack(spacing: compact ? 10 : 16) {
                actionButtons(snap: snap, localIdx: localIdx)
                PlayerHandView(
                    player: me,
                    isMyTurn: snap.TurnPlayer == localIdx && (snap.PendingRaiseFor ?? -1) == -1,
                    store: store,
                    dealProgress: dealAnimationProgress,
                    snap: snap,
                    compact: compact
                )
            }
            .padding(.bottom, compact ? 24 : 40)
        }
    }
    
    // MARK: - Match Over
    @ViewBuilder
    private func matchOverOverlay(snap: MatchSnapshot) -> some View {
        let myTeam = snap.Players?.first(where: { $0.ID == (snap.CurrentPlayerIdx ?? 0) })?.Team ?? 0
        VStack(spacing: 24) {
            Text(snap.WinnerTeam == myTeam ? "VITÓRIA! 🎉" : "DERROTA 😢")
                .font(.system(size: 56, weight: .black, design: .rounded))
                .foregroundColor(.yellow)
                .shadow(color: .yellow.opacity(0.4), radius: 20)
            
            Button("NOVA PARTIDA") { store.mode = "idle" }
                .buttonStyle(.borderedProminent).controlSize(.large).tint(.yellow).foregroundColor(.black)
                .font(.headline.weight(.black))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.black.opacity(0.7))
        .transition(.opacity)
    }
    
    // MARK: - Action Buttons
    @ViewBuilder
    private func actionButtons(snap: MatchSnapshot, localIdx: Int) -> some View {
        let myTeam = snap.Players?.first(where: { $0.ID == localIdx })?.Team ?? 0
        let pendingFor = snap.PendingRaiseFor ?? -1
        let canAsk = snap.CanAskTruco ?? false
        let isMyTurn = snap.TurnPlayer == localIdx
        let matchOver = snap.MatchFinished ?? false
        
        if !matchOver {
            if pendingFor == myTeam {
                HStack(spacing: 16) {
                    Button("ACEITAR") { store.dispatchGameAction("accept") }
                        .buttonStyle(.borderedProminent).tint(.green).controlSize(.large).font(.headline.weight(.black))
                    let cs = snap.CurrentHand?.Stake ?? 1
                    if cs < 9 {
                        Button(raiseLabelFor(stake: nextStake(cs))) { store.dispatchGameAction("truco") }
                            .buttonStyle(.borderedProminent).tint(.orange).controlSize(.large).font(.headline.weight(.black))
                    }
                    Button("CORRER") { store.dispatchGameAction("refuse") }
                        .buttonStyle(.borderedProminent).tint(.red).controlSize(.large).font(.headline.weight(.black))
                }
                .transition(.scale.combined(with: .opacity))
            } else if isMyTurn && canAsk {
                let cs = snap.CurrentHand?.Stake ?? 1
                Button(raiseLabelFor(stake: nextStake(cs))) { store.dispatchGameAction("truco") }
                    .buttonStyle(.borderedProminent).tint(.yellow).foregroundColor(.black).controlSize(.large).font(.headline.weight(.black))
                    .transition(.scale.combined(with: .opacity))
            }
        }
    }
    
    // MARK: - Helpers
    private func playerAt(snap: MatchSnapshot, offset: Int, localIdx: Int, n: Int) -> Player? {
        snap.Players?.first(where: { $0.ID == (localIdx + offset) % n })
    }
    private func stakeName(_ l: Int) -> String { [1:"1",3:"T",6:"6",9:"9",12:"12"][l] ?? "\(l)" }
    private func nextStake(_ s: Int) -> Int { [1:3,3:6,6:9,9:12][s] ?? s }
    private func raiseLabelFor(stake: Int) -> String { [3:"TRUCO!",6:"SEIS!",9:"NOVE!",12:"DOZE!"][stake] ?? "TRUCO!" }
    
    private func roleBadge(snap: MatchSnapshot, playerID: Int) -> String? {
        guard let hand = snap.CurrentHand else { return nil }
        if hand.Dealer == playerID { return "🃏" }
        let n = snap.NumPlayers ?? 2
        let dealerIdx = hand.Dealer ?? 0
        let maoIdx = (dealerIdx + 1) % n
        if playerID == maoIdx { return "✋" }
        if n == 4 {
            let peIdx = (dealerIdx + n - 1) % n
            if playerID == peIdx { return "🦶" }
        }
        return nil
    }
}

// MARK: - Subviews

enum PlayerPosition { case top, left, right }

private struct ScoreView: View {
    let teamName: String; let points: Int; let color: Color
    var body: some View {
        VStack(spacing: 2) {
            Text(teamName).font(.system(size: 10, weight: .bold)).foregroundColor(color.opacity(0.8))
            Text("\(points)").font(.system(size: 28, weight: .black, design: .rounded)).foregroundColor(.white)
        }
        .frame(width: 70, height: 60)
        .background(Color.black.opacity(0.4)).cornerRadius(12)
        .overlay(RoundedRectangle(cornerRadius: 12).stroke(color.opacity(0.4), lineWidth: 1))
    }
}

private struct CenterTableView: View {
    let hand: HandState; let turnPlayer: Int?; let players: [Player]?; let compact: Bool
    var body: some View {
        HStack(spacing: compact ? 20 : 40) {
            VStack(spacing: 6) {
                Text("VIRA").font(.system(size: 9, weight: .bold)).foregroundColor(.white.opacity(0.5)).tracking(2)
                if let vira = hand.Vira {
                    CardView(card: vira).rotationEffect(.degrees(-6)).scaleEffect(compact ? 0.8 : 1)
                }
            }
            ZStack {
                Circle().fill(Color.white.opacity(0.04)).frame(width: compact ? 120 : 160, height: compact ? 120 : 160)
                if let played = hand.RoundCards {
                    ForEach(Array(played.enumerated()), id: \.element.id) { idx, pc in
                        VStack(spacing: 2) {
                            Text(players?.first(where: { $0.ID == pc.PlayerID })?.Name ?? "?")
                                .font(.system(size: 9)).foregroundColor(.white.opacity(0.7))
                            CardView(card: pc.Card).scaleEffect(compact ? 0.75 : 0.9)
                        }
                        .rotationEffect(.degrees(Double(idx * 15 - 10)))
                        .offset(x: CGFloat(idx * 15 - 5), y: CGFloat(idx * -10))
                        .transition(.asymmetric(
                            insertion: .scale(scale: 0.3).combined(with: .opacity).combined(with: .offset(y: 80)),
                            removal: .opacity
                        ))
                    }
                }
            }
            .animation(.spring(response: 0.5, dampingFraction: 0.7), value: hand.RoundCards?.count)
            VStack(spacing: 6) {
                Text("MANILHA").font(.system(size: 9, weight: .bold)).foregroundColor(.white.opacity(0.5)).tracking(2)
                if let m = hand.Manilha, m != "-" {
                    Text(m).font(.system(size: compact ? 24 : 30, weight: .black, design: .rounded)).foregroundColor(.yellow)
                        .frame(width: compact ? 56 : 70, height: compact ? 80 : 100)
                        .background(Color.yellow.opacity(0.15)).cornerRadius(10)
                        .overlay(RoundedRectangle(cornerRadius: 10).stroke(Color.yellow, lineWidth: 2))
                }
            }
        }
    }
}

private struct OpponentView: View {
    let player: Player; let isTurn: Bool; let position: PlayerPosition; let snap: MatchSnapshot; let compact: Bool
    
    var body: some View {
        let isVertical = position == .left || position == .right
        VStack(spacing: 8) {
            HStack(spacing: 4) {
                // Role badge
                if let badge = roleBadgeFor(playerID: player.ID) {
                    Text(badge).font(.system(size: 12))
                }
                Text(player.Name.uppercased()).font(.system(size: compact ? 9 : 11, weight: .bold)).foregroundColor(.white).tracking(0.5)
                if player.CPU == true {
                    if isTurn {
                        ProgressView().controlSize(.mini).tint(.yellow)
                    } else {
                        Text("🤖").font(.system(size: 10))
                    }
                }
                if isTurn && player.CPU != true {
                    Text("◀").font(.system(size: 9, weight: .bold)).foregroundColor(.yellow)
                }
            }
            .padding(.horizontal, 10).padding(.vertical, 4)
            .background(isTurn ? Color.yellow.opacity(0.3) : Color.black.opacity(0.5))
            .clipShape(Capsule())
            
            Text("Time \(player.Team + 1)").font(.system(size: 9, weight: .bold)).foregroundColor(player.Team == 0 ? .blue : .orange)
            
            if isVertical {
                VStack(spacing: -30) {
                    ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                        CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false).scaleEffect(compact ? 0.55 : 0.7)
                    }
                }
            } else {
                HStack(spacing: compact ? -24 : -20) {
                    ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                        CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                            .scaleEffect(compact ? 0.8 : 1)
                            .shadow(color: .black.opacity(0.3), radius: 4, x: -2, y: 3)
                    }
                }
            }
        }
    }
    
    private func roleBadgeFor(playerID: Int) -> String? {
        guard let hand = snap.CurrentHand else { return nil }
        let n = snap.NumPlayers ?? 2
        let d = hand.Dealer ?? 0
        if playerID == d { return "🃏" }
        if playerID == (d + 1) % n { return "✋" }
        if n == 4 && playerID == (d + n - 1) % n { return "🦶" }
        return nil
    }
}

private struct PlayerHandView: View {
    let player: Player; let isMyTurn: Bool; let store: TrucoAppStore
    let dealProgress: [Bool]; let snap: MatchSnapshot; let compact: Bool
    @State private var hoveredCard: String?
    
    var body: some View {
        VStack(spacing: compact ? 8 : 14) {
            HStack(spacing: 6) {
                if let badge = roleBadge { Text(badge).font(.system(size: 14)) }
                Text(player.Name.uppercased()).font(.system(size: 12, weight: .bold)).foregroundColor(.white).tracking(1)
                Text("Time \(player.Team + 1)").font(.system(size: 9, weight: .bold))
                    .foregroundColor(player.Team == 0 ? .blue : .orange)
                    .padding(.horizontal, 5).padding(.vertical, 2)
                    .background((player.Team == 0 ? Color.blue : Color.orange).opacity(0.2)).cornerRadius(4)
            }
            if isMyTurn {
                Text("SUA VEZ")
                    .font(.system(size: 10, weight: .black)).foregroundColor(Color(red: 0.35, green: 0.21, blue: 0.12))
                    .tracking(1.5).padding(.horizontal, 14).padding(.vertical, 4)
                    .background(Color.yellow).clipShape(Capsule())
                    .shadow(color: .yellow.opacity(0.4), radius: 8)
            }
            HStack(spacing: compact ? 8 : 14) {
                if let hand = player.Hand {
                    ForEach(Array(hand.enumerated()), id: \.element) { index, card in
                        CardView(card: card)
                            .scaleEffect(compact ? 0.85 : 1)
                            .offset(y: hoveredCard == card.Rank + card.Suit ? -30 : 0)
                            .opacity(index < dealProgress.count && dealProgress[index] ? 1 : 0.3)
                            .scaleEffect(index < dealProgress.count && dealProgress[index] ? 1 : 0.8)
                            .animation(.interactiveSpring(response: 0.3, dampingFraction: 0.6), value: hoveredCard)
                            .onHover { h in
                                if h && isMyTurn { hoveredCard = card.Rank + card.Suit }
                                else if hoveredCard == card.Rank + card.Suit { hoveredCard = nil }
                            }
                            .onTapGesture { if isMyTurn { store.dispatchGameAction("play", cardIndex: index) } }
                    }
                }
            }
        }
    }
    
    private var roleBadge: String? {
        guard let hand = snap.CurrentHand else { return nil }
        let n = snap.NumPlayers ?? 2
        let d = hand.Dealer ?? 0
        if player.ID == d { return "🃏" }
        if player.ID == (d + 1) % n { return "✋" }
        if n == 4 && player.ID == (d + n - 1) % n { return "🦶" }
        return nil
    }
}
