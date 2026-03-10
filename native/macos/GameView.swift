import SwiftUI

struct GameView: View {
    let snapshot: GameSnapshot?
    
    var body: some View {
        if let snap = snapshot {
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
                        ScoreView(teamName: "Nós", points: snap.MatchPoints?[0] ?? 0)
                        Spacer()
                        StakeBadge(stake: snap.CurrentHand?.Stake ?? 1)
                        Spacer()
                        ScoreView(teamName: "Eles", points: snap.MatchPoints?[1] ?? 0)
                    }
                    .padding(50)
                    Spacer()
                }
                
                // Players & Center Table
                VStack(spacing: 0) {
                    if let opponent = snap.Players?.first(where: { $0.ID != 0 }) {
                        OpponentView(player: opponent)
                            .padding(.top, 60)
                    }
                    
                    Spacer()
                    
                    if let center = snap.CurrentHand {
                        CenterTableView(hand: center)
                    }
                    
                    Spacer()
                    
                    if let me = snap.Players?.first(where: { $0.ID == 0 }) {
                        PlayerHandView(player: me, isMyTurn: snap.TurnPlayer == 0)
                            .padding(.bottom, 60)
                    }
                }
            }
        } else {
            ProgressView("Carregando snapshot...")
                .scaleEffect(1.5)
        }
    }
}

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
                        CardView(card: pc.Card)
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

private struct PlayerHandView: View {
    let player: Player
    let isMyTurn: Bool
    
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
                    ForEach(hand, id: \.self) { card in
                        CardView(card: card)
                            .offset(y: hoveredCard == card.Rank + card.Suit ? -30 : 0)
                            .animation(.interactiveSpring(response: 0.3, dampingFraction: 0.6), value: hoveredCard)
                            .onHover { isHovered in
                                if isHovered {
                                    hoveredCard = card.Rank + card.Suit
                                } else if hoveredCard == card.Rank + card.Suit {
                                    hoveredCard = nil
                                }
                            }
                    }
                }
            }
        }
    }
}
