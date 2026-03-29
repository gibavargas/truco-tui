import SwiftUI

struct GameView: View {
    let snapshot: MatchSnapshot?
    @EnvironmentObject var store: TrucoAppStore

    @State private var lastTrickSeqViewed: Int = -1
    @State private var showingTrickEndAnimation = false
    @State private var trickAnimOffset: CGSize = .zero
    @State private var trickAnimProgress: CGFloat = 0
    @State private var trickWinnerTeam: Int = -1
    @State private var trickTie: Bool = false
    @State private var chatMessage = ""

    fileprivate enum TrickPilePlacement {
        case top
        case bottom
        case leading
        case trailing

        var panelSize: CGSize {
            switch self {
            case .top, .bottom:
                return CGSize(width: 248, height: 132)
            case .leading, .trailing:
                return CGSize(width: 132, height: 248)
            }
        }

        var cardSize: CGSize {
            switch self {
            case .top, .bottom:
                return CGSize(width: 58, height: 84)
            case .leading, .trailing:
                return CGSize(width: 54, height: 78)
            }
        }

        func offset(for index: Int, count: Int) -> CGSize {
            let step = CGFloat(index)
            let total = CGFloat(max(count - 1, 1))
            switch self {
            case .top:
                return CGSize(width: (step - total / 2) * 42, height: step * 11)
            case .bottom:
                return CGSize(width: (step - total / 2) * 42, height: step * -11)
            case .leading:
                return CGSize(width: step * 11, height: (step - total / 2) * 42)
            case .trailing:
                return CGSize(width: step * -11, height: (step - total / 2) * 42)
            }
        }

        func rotation(for index: Int, count: Int) -> Double {
            let step = Double(index)
            let total = Double(max(count - 1, 1))
            let bias = step - total / 2
            switch self {
            case .top:
                return bias * 5.5
            case .bottom:
                return bias * -5.5
            case .leading:
                return bias * 4.5
            case .trailing:
                return bias * -4.5
            }
        }
    }

    private func seatPlayer(_ snap: MatchSnapshot, offset: Int) -> Player? {
        guard let players = snap.Players, let localID = snap.CurrentPlayerIdx, let count = snap.NumPlayers else { return nil }
        let seatID = (localID + offset) % count
        return players.first(where: { $0.playerID == seatID })
    }

    private func trickPiles(for snap: MatchSnapshot, playerID: Int) -> [TrickPile] {
        (snap.TrickPiles ?? [])
            .filter { ($0.Winner ?? -1) == playerID }
            .sorted {
                let leftRound = $0.Round ?? Int.max
                let rightRound = $1.Round ?? Int.max
                if leftRound == rightRound {
                    return ($0.Cards?.count ?? 0) < ($1.Cards?.count ?? 0)
                }
                return leftRound < rightRound
            }
    }

    fileprivate enum SeatRelation {
        case partner
        case opponent

        var label: String {
            switch self {
            case .partner:
                return "(PARCEIRO)"
            case .opponent:
                return "(ADVERSÁRIO)"
            }
        }

        var tint: Color {
            switch self {
            case .partner:
                return .green
            case .opponent:
                return .orange
            }
        }
    }

    private func seatRelation(for player: Player, localPlayerID: Int, localTeam: Int) -> SeatRelation {
        if player.playerID == localPlayerID {
            return .partner
        }
        return player.Team == localTeam ? .partner : .opponent
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
            let slotStates = store.bundle?.ui?.lobby_slots ?? []
            let connection = store.bundle?.connection
            let diagnostics = store.bundle?.diagnostics
            let isOnline = store.mode == "host_match" || store.mode == "client_match"
            let localPlayer = snap.Players?.first(where: { $0.playerID == snap.CurrentPlayerIdx })
            let localTeam = actions?.local_team ?? localPlayer?.Team ?? 0
            let lastTrickCards = snap.LastTrickCards ?? []
            let showLastTrickMonte = !lastTrickCards.isEmpty
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
                    HStack(alignment: .top) {
                        ScoreView(teamName: "Nós", points: snap.teamScore.us)
                        Spacer()
                        StakeInfoView(stake: snap.CurrentHand?.Stake ?? 1)
                        ScoreView(teamName: "Eles", points: snap.teamScore.them)
                    }
                    .padding(.horizontal, 50)
                    .padding(.top, 78)
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
                            .disabled(!store.canCloseSession)
                            .buttonStyle(.plain)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 8)
                            .background(Color.white.opacity(0.15))
                            .foregroundColor(.white)
                            .clipShape(Capsule())
                            .overlay(Capsule().stroke(Color.white.opacity(0.3), lineWidth: 1))
                            .shadow(radius: 4)
                            .padding(.leading, 30)
                            .padding(.top, 58)
                            
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
                    .padding(.top, 176)
                    .padding(.trailing, 44)
                    Spacer()
                }
                
                // Players & Center Table
                VStack(spacing: 0) {
                    if let opponent = seatPlayer(snap, offset: snap.NumPlayers == 4 ? 2 : 1) {
                        OpponentView(
                            player: opponent,
                            relation: seatRelation(
                                for: opponent,
                                localPlayerID: localPlayer?.playerID ?? opponent.playerID,
                                localTeam: localTeam
                            ),
                            trickPiles: trickPiles(for: snap, playerID: opponent.playerID),
                            placement: .top
                        )
                            .padding(.top, 96)
                    }
                    
                    Spacer()
                    
                    if let center = snap.CurrentHand {
                        CenterTableView(hand: center, players: snap.Players ?? [])
                    }

                    if showLastTrickMonte && snap.LastTrickTie == true {
                        MontePileView(title: "EMPATE", count: lastTrickCards.count)
                            .padding(.top, 10)
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
                            
                            PlayerHandView(
                                player: me,
                                isMyTurn: actions?.can_play_card == true,
                                currentRound: snap.CurrentHand?.Round ?? 1,
                                trickPiles: trickPiles(for: snap, playerID: me.playerID),
                                placement: .bottom
                            )
                        }
                        .padding(.bottom, 60)
                    }
                }

                if snap.NumPlayers == 4 {
                    HStack {
                        if let left = seatPlayer(snap, offset: 3) {
                            SideOpponentView(
                                player: left,
                                relation: seatRelation(
                                    for: left,
                                    localPlayerID: localPlayer?.playerID ?? left.playerID,
                                    localTeam: localTeam
                                ),
                                labelOnTrailingSide: true,
                                trickPiles: trickPiles(for: snap, playerID: left.playerID),
                                placement: .leading
                            )
                                .frame(maxWidth: 150)
                                .padding(.leading, 48)
                        }
                        Spacer()
                        if let right = seatPlayer(snap, offset: 1) {
                            SideOpponentView(
                                player: right,
                                relation: seatRelation(
                                    for: right,
                                    localPlayerID: localPlayer?.playerID ?? right.playerID,
                                    localTeam: localTeam
                                ),
                                labelOnTrailingSide: false,
                                trickPiles: trickPiles(for: snap, playerID: right.playerID),
                                placement: .trailing
                            )
                                .frame(maxWidth: 150)
                                .padding(.trailing, 48)
                        }
                    }
                    .padding(.vertical, 220)
                }

                if isOnline {
                    let network = connection?.network
                    VStack {
                        HStack(alignment: .top, spacing: 16) {
                            Spacer()
                            VStack(alignment: .leading, spacing: 12) {
                                VStack(alignment: .leading, spacing: 8) {
                                    Text("Conexão")
                                        .font(.caption.bold())
                                        .foregroundColor(.white.opacity(0.6))
                                    connectionLine("Status", connection?.status ?? store.mode)
                                    connectionLine("Modo", connection?.is_online == true ? "online" : "offline")
                                    if let network {
                                        connectionLine("Transporte", network.transportLabel)
                                        connectionLine("Protocolo", network.compatibilitySummary(isHost: connection?.is_host == true))
                                    }
                                    connectionLine("Fila", "\(diagnostics?.event_backlog ?? 0)")
                                    if let message = connection?.last_error?.message, !message.isEmpty {
                                        connectionLine("Erro", message, tint: .red.opacity(0.9))
                                    }
                                }
                                .padding(12)
                                .background(Color.black.opacity(0.32))
                                .cornerRadius(12)

                                if !slotStates.isEmpty {
                                    VStack(alignment: .leading, spacing: 8) {
                                        Text("Mesa Online")
                                            .font(.caption.bold())
                                            .foregroundColor(.white.opacity(0.6))
                                        ForEach(slotStates) { slot in
                                            VStack(alignment: .leading, spacing: 6) {
                                                HStack {
                                                    Text("Slot \(slot.seat + 1)")
                                                        .font(.caption.weight(.semibold))
                                                        .foregroundColor(.white)
                                                    Spacer()
                                                    Text(slot.name?.isEmpty == false ? slot.name! : "Aguardando...")
                                                        .font(.caption)
                                                        .foregroundColor(slot.is_empty ? .gray : .white.opacity(0.85))
                                                }
                                                HStack(spacing: 6) {
                                                    if slot.is_local { slotTag("você", color: .yellow) }
                                                    if slot.is_host { slotTag("host", color: .blue) }
                                                    slotTag(slot.is_connected ? "online" : "offline", color: slot.is_connected ? .green : .gray)
                                                    if slot.is_provisional_cpu { slotTag("cpu", color: .orange) }
                                                }
                                                HStack(spacing: 8) {
                                                    if slot.can_vote_host {
                                                        Button("Votar Host") {
                                                            store.voteHost(candidateSeat: slot.seat)
                                                        }
                                                        .font(.caption2)
                                                        .buttonStyle(.bordered)
                                                    }
                                                    if slot.can_request_replacement {
                                                        Button("Substituição") {
                                                            store.requestReplacementInvite(targetSeat: slot.seat)
                                                        }
                                                        .font(.caption2)
                                                        .buttonStyle(.borderedProminent)
                                                        .tint(.orange)
                                                    }
                                                }
                                            }
                                            .padding(10)
                                            .background(Color.white.opacity(0.04))
                                            .cornerRadius(10)
                                        }
                                    }
                                    .padding(12)
                                    .background(Color.black.opacity(0.32))
                                    .cornerRadius(12)
                                }

                                VStack(alignment: .leading, spacing: 10) {
                                    Text("Chat e Eventos")
                                        .font(.caption.bold())
                                        .foregroundColor(.white.opacity(0.6))
                                    ScrollView {
                                        VStack(alignment: .leading, spacing: 6) {
                                            ForEach(store.events.suffix(14)) { event in
                                                eventRow(event)
                                            }
                                        }
                                        .frame(maxWidth: .infinity, alignment: .leading)
                                    }
                                    .frame(height: 140)

                                    HStack {
                                        TextField("Digite uma mensagem...", text: $chatMessage)
                                            .textFieldStyle(.roundedBorder)
                                            .onSubmit {
                                                sendChatIfNeeded()
                                            }
                                        Button("Enviar") {
                                            sendChatIfNeeded()
                                        }
                                        .buttonStyle(.borderedProminent)
                                        .disabled(chatMessage.isEmpty)
                                    }
                                }
                                .padding(12)
                                .background(Color.black.opacity(0.32))
                                .cornerRadius(12)
                            }
                            .frame(width: 320)
                            .padding(.trailing, 24)
                        }
                        .padding(.top, 120)
                        .padding(.top, 26)
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
                        
                        let didWinMatch = snap.WinnerTeam == localTeam
                        Text(didWinMatch ? "VOCÊ VENCEU! 🏆" : "VOCÊ PERDEU 😢")
                            .font(.system(size: 28, weight: .bold, design: .rounded))
                            .foregroundColor(didWinMatch ? .yellow : .red)
                        
                        if isOnline {
                            Button("SAIR DA SESSÃO") {
                                store.closeSession()
                            }
                            .disabled(!store.canCloseSession)
                            .buttonStyle(.borderedProminent)
                            .tint(.yellow)
                            .foregroundColor(.black)
                            .controlSize(.large)
                            .font(.headline.weight(.black))
                        } else {
                            Button("NOVA PARTIDA") {
                                store.replayOfflineMatch()
                            }
                            .buttonStyle(.borderedProminent)
                            .tint(.yellow)
                            .foregroundColor(.black)
                            .controlSize(.large)
                            .font(.headline.weight(.black))
                        }
                    }
                }
                
                // Trick end animation overlay
                if showingTrickEndAnimation {
                    ZStack {
                        TrickTravelDeckView(
                            offset: trickAnimOffset,
                            progress: trickAnimProgress,
                            tie: trickTie
                        )
                        .zIndex(1)

                        TrickResultToast(
                            localTeam: localTeam,
                            winnerTeam: trickWinnerTeam,
                            tie: trickTie
                        )
                        .zIndex(10)
                    }
                    .zIndex(100)
                    .allowsHitTesting(false)
                    .transition(.opacity.animation(.easeInOut(duration: 0.2)))
                }
            }
            .onChange(of: snap.LastTrickSeq) {
                let newSeq = snap.LastTrickSeq
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
        trickAnimProgress = 0
        showingTrickEndAnimation = false

        DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) {
            withAnimation(.spring(response: 0.55, dampingFraction: 0.78)) {
                showingTrickEndAnimation = true
                trickAnimProgress = 1
            }
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.75) {
                withAnimation(.easeInOut(duration: 0.95)) {
                    trickAnimOffset = target
                }
            }
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 2.15) {
                withAnimation(.easeOut(duration: 0.35)) {
                    showingTrickEndAnimation = false
                }
            }
        }
    }

    private func sendChatIfNeeded() {
        let trimmed = chatMessage.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        store.sendChat(text: trimmed)
        chatMessage = ""
    }

    private func slotTag(_ label: String, color: Color) -> some View {
        Text(label)
            .font(.caption2.weight(.semibold))
            .foregroundColor(color)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(color.opacity(0.12))
            .clipShape(Capsule())
    }

    private func connectionLine(_ label: String, _ value: String, tint: Color = .white) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label.uppercased())
                .font(.caption2)
                .foregroundColor(.white.opacity(0.45))
            Text(value)
                .font(.caption)
                .foregroundColor(tint)
        }
    }

    @ViewBuilder
    private func eventRow(_ event: AppEvent) -> some View {
        switch event.kind {
        case "chat":
            Text("\(event.payload?.author ?? "?"): \(event.payload?.text ?? "")")
                .font(.caption)
                .foregroundColor(.white)
        case "system":
            Text(event.payload?.text ?? "")
                .font(.caption)
                .foregroundColor(.secondary)
        case "replacement_invite":
            Text("Link de subs: \(event.payload?.invite_key ?? "")")
                .font(.caption)
                .foregroundColor(.green)
        case "error":
            Text(event.payload?.message ?? event.payload?.text ?? "Erro")
                .font(.caption)
                .foregroundColor(.red.opacity(0.9))
        case "lobby_updated":
            Text("Lobby atualizado")
                .font(.caption)
                .foregroundColor(.white.opacity(0.65))
        case "match_updated":
            Text("Partida atualizada")
                .font(.caption)
                .foregroundColor(.white.opacity(0.65))
        default:
            Text(event.payload?.text ?? event.kind)
                .font(.caption)
                .foregroundColor(.white.opacity(0.7))
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

private struct StakeInfoView: View {
    let stake: Int

    var body: some View {
        VStack(alignment: .trailing, spacing: 6) {
            StakeBadge(stake: stake)

            if stake > 1 {
                VStack(alignment: .trailing, spacing: 2) {
                    Text("APOSTA EM CURSO")
                    Text("valor da mão: 6, 9, 12...")
                }
                .font(.caption2.weight(.semibold))
                .foregroundColor(.white.opacity(0.72))
                .multilineTextAlignment(.trailing)
                .frame(maxWidth: 180, alignment: .trailing)
            }
        }
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
        .frame(maxWidth: 250)
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
                            CardView(card: pc.Card, isFaceUp: pc.FaceDown != true)
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
    let relation: GameView.SeatRelation
    let trickPiles: [TrickPile]
    let placement: GameView.TrickPilePlacement
    
    var body: some View {
        VStack(spacing: 10) {
            HStack(alignment: .center, spacing: 14) {
                if !trickPiles.isEmpty {
                    TrickPilesView(piles: trickPiles, placement: placement)
                }

                HStack(spacing: -24) {
                    ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                        CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                            .shadow(color: .black.opacity(0.3), radius: 4, x: -2, y: 3)
                    }
                }

                playerBadge(alignment: .center)
            }
        }
    }

    @ViewBuilder
    private func playerBadge(alignment: HorizontalAlignment) -> some View {
        VStack(alignment: alignment, spacing: 6) {
            HStack(spacing: 6) {
                Text(player.Name.uppercased())
                    .font(.subheadline.bold())
                    .foregroundColor(.white)
                    .tracking(1)
                    .padding(.horizontal, 20)
                    .padding(.vertical, 6)
                    .background(Color.black.opacity(0.5))
                    .clipShape(Capsule())

                if player.CPU == true {
                    Text("CPU")
                        .font(.caption2.bold())
                        .foregroundColor(.yellow)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.black.opacity(0.55))
                        .clipShape(Capsule())
                }
            }

            relationBadge()
        }
    }

    @ViewBuilder
    private func relationBadge() -> some View {
        Text(relation.label)
            .font(.caption2.weight(.bold))
            .foregroundColor(relation.tint)
            .padding(.horizontal, 10)
            .padding(.vertical, 4)
            .background(relation.tint.opacity(0.14))
            .clipShape(Capsule())
    }
}

private struct SideOpponentView: View {
    let player: Player
    let relation: GameView.SeatRelation
    let labelOnTrailingSide: Bool
    let trickPiles: [TrickPile]
    let placement: GameView.TrickPilePlacement

    var body: some View {
        HStack(alignment: .center, spacing: 12) {
            if !trickPiles.isEmpty {
                TrickPilesView(piles: trickPiles, placement: placement)
            }

            if labelOnTrailingSide {
                cardStack
                playerBadge(alignment: .leading)
            } else {
                playerBadge(alignment: .trailing)
                cardStack
            }
        }
    }

    private var cardStack: some View {
        VStack(spacing: -20) {
            ForEach(0..<(player.Hand?.count ?? 3), id: \.self) { _ in
                CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                    .frame(width: 60, height: 88)
            }
        }
    }

    @ViewBuilder
    private func playerBadge(alignment: HorizontalAlignment) -> some View {
        VStack(alignment: alignment, spacing: 6) {
            HStack(spacing: 6) {
                Text(player.Name.uppercased())
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.88))
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(Color.black.opacity(0.55))
                    .clipShape(Capsule())

                if player.CPU == true {
                    Text("CPU")
                        .font(.caption2.bold())
                        .foregroundColor(.yellow)
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(Color.black.opacity(0.55))
                        .clipShape(Capsule())
                }
            }

            Text(relation.label)
                .font(.caption2.weight(.bold))
                .foregroundColor(relation.tint)
                .padding(.horizontal, 10)
                .padding(.vertical, 4)
                .background(relation.tint.opacity(0.14))
                .clipShape(Capsule())
        }
    }
}

private struct MontePileView: View {
    let title: String
    let count: Int

    var body: some View {
        VStack(spacing: 6) {
            Text(title)
                .font(.caption2.bold())
                .foregroundColor(.yellow.opacity(0.9))
                .tracking(1.5)
            HStack(spacing: -12) {
                ForEach(0..<max(1, min(count, 4)), id: \.self) { _ in
                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                        .frame(width: 34, height: 50)
                }
            }
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 6)
        .background(Color.black.opacity(0.25))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }
}

private struct TrickTravelDeckView: View {
    let offset: CGSize
    let progress: CGFloat
    let tie: Bool

    var body: some View {
        ZStack {
            ForEach(0..<3, id: \.self) { index in
                let lead = CGFloat(index)
                let driftX = (lead - 1) * 22
                let driftY = lead * -8
                let scale = 1 - lead * 0.05
                let opacity = tie ? 0.42 : 0.82

                CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                    .frame(width: 84, height: 120)
                    .rotationEffect(.degrees(Double((lead - 1) * 7 + progress * 12)))
                    .offset(
                        x: driftX + offset.width * progress,
                        y: driftY + offset.height * progress
                    )
                    .scaleEffect(scale + progress * 0.08)
                    .opacity(opacity)
                    .shadow(color: .black.opacity(0.35), radius: 12, x: 0, y: 6)
            }
        }
        .overlay(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .stroke(Color.white.opacity(0.12), lineWidth: 1)
        )
    }
}

private struct TrickResultToast: View {
    let localTeam: Int
    let winnerTeam: Int
    let tie: Bool

    var body: some View {
        VStack(spacing: 10) {
            if tie {
                Text("EMPATE")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.8))
                    .tracking(2)
                Text("😐")
                    .font(.system(size: 54))
            } else if winnerTeam == localTeam {
                Text("VOCÊ VENCEU")
                    .font(.headline.weight(.heavy))
                    .foregroundColor(.green)
                    .tracking(1.5)
                Text("a vaza")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.8))
            } else {
                Text("ELES VENCERAM")
                    .font(.headline.weight(.heavy))
                    .foregroundColor(.red)
                    .tracking(1.5)
                Text("a vaza")
                    .font(.caption.bold())
                    .foregroundColor(.white.opacity(0.8))
            }
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 14)
        .background(.black.opacity(0.82))
        .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 22, style: .continuous)
                .stroke(Color.white.opacity(0.14), lineWidth: 1)
        )
        .shadow(color: .black.opacity(0.3), radius: 20, x: 0, y: 10)
        .offset(y: 8)
    }
}

private struct TrickPilesView: View {
    let piles: [TrickPile]
    let placement: GameView.TrickPilePlacement

    var body: some View {
        ZStack {
            ForEach(Array(piles.enumerated()), id: \.offset) { index, pile in
                TrickPileCardView(
                    sequence: index + 1,
                    count: pile.Cards?.count ?? 0,
                    placement: placement
                )
                .offset(placement.offset(for: index, count: piles.count))
                .rotationEffect(.degrees(placement.rotation(for: index, count: piles.count)))
                .transition(.scale.combined(with: .opacity))
                .zIndex(Double(index))
            }
        }
        .frame(width: placement.panelSize.width, height: placement.panelSize.height)
        .padding(8)
        .background(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .fill(
                    LinearGradient(
                        colors: [
                            Color.black.opacity(0.42),
                            Color.black.opacity(0.18),
                            Color.green.opacity(0.08)
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
        )
        .overlay(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .stroke(Color.white.opacity(0.08), lineWidth: 1)
        )
    }
}

private struct TrickPileCardView: View {
    let sequence: Int
    let count: Int
    let placement: GameView.TrickPilePlacement

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 6) {
                Text("\(sequence)")
                    .font(.caption2.weight(.black))
                    .foregroundColor(.black)
                    .frame(width: 18, height: 18)
                    .background(Color.yellow)
                    .clipShape(Circle())

                Spacer(minLength: 0)

                Text("\(count)")
                    .font(.caption2.weight(.bold))
                    .foregroundColor(.white.opacity(0.88))
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.white.opacity(0.08))
                    .clipShape(Capsule())
            }

            ZStack {
                ForEach(0..<max(1, min(count, 4)), id: \.self) { index in
                    CardView(card: Card(Rank: "", Suit: ""), isFaceUp: false)
                        .frame(width: placement.cardSize.width, height: placement.cardSize.height)
                        .offset(x: CGFloat(index) * 5, y: CGFloat(index) * -4)
                        .rotationEffect(.degrees(Double(index) * 2.5))
                        .shadow(color: .black.opacity(0.25), radius: 4, x: 0, y: 2)
                }
            }
        }
        .padding(10)
        .frame(width: placement.cardSize.width + 34, height: placement.cardSize.height + 42)
        .background(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .fill(
                    LinearGradient(
                        colors: [
                            Color.white.opacity(0.08),
                            Color.white.opacity(0.02)
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
        )
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.yellow.opacity(0.26), lineWidth: 1)
        )
    }
}

private struct PlayerHandView: View {
    let player: Player
    let isMyTurn: Bool
    let currentRound: Int
    let trickPiles: [TrickPile]
    let placement: GameView.TrickPilePlacement
    @EnvironmentObject var store: TrucoAppStore
    
    @State private var hoveredCard: String?
    
    var body: some View {
        HStack(alignment: .bottom, spacing: 18) {
            if !trickPiles.isEmpty {
                TrickPilesView(piles: trickPiles, placement: placement)
                    .padding(.bottom, 12)
            }

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
                            VStack(spacing: 8) {
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

                                if isMyTurn && currentRound >= 2 {
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
}
