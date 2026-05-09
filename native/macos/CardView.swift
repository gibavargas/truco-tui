import SwiftUI

struct CardView: View {
    let card: Card
    let isFaceUp: Bool
    
    init(card: Card, isFaceUp: Bool = true) {
        self.card = card
        self.isFaceUp = isFaceUp
    }
    
    var body: some View {
        ZStack {
            cardBack
                .rotation3DEffect(.degrees(180), axis: (x: 0, y: 1, z: 0))
                .opacity(isFaceUp ? 0 : 1)
            
            cardFront
                .opacity(isFaceUp ? 1 : 0)
        }
        .rotation3DEffect(.degrees(isFaceUp ? 0 : 180), axis: (x: 0, y: 1, z: 0))
        .shadow(color: Color.black.opacity(0.34), radius: 12, x: 0, y: 8)
        .frame(width: 86, height: 124)
        .animation(.spring(response: 0.5, dampingFraction: 0.7), value: isFaceUp)
    }
    
    private var cardFront: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .fill(
                    LinearGradient(
                        colors: [
                            Color(red: 1.0, green: 0.99, blue: 0.95),
                            Color(red: 0.94, green: 0.88, blue: 0.76)
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
            
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(Color(red: 0.76, green: 0.58, blue: 0.32).opacity(0.42), lineWidth: 1)

            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .stroke(Color.black.opacity(0.07), lineWidth: 1)
                .padding(6)

            Text("TP")
                .font(.system(size: 32, weight: .black, design: .serif))
                .foregroundColor(Color.black.opacity(0.045))
                .tracking(2)
            
            // Top-left corner
            VStack {
                HStack {
                    VStack(spacing: -4) {
                        Text(card.Rank)
                            .font(.system(size: 18, weight: .bold, design: .rounded))
                        Text(card.suitSymbol)
                            .font(.system(size: 16))
                    }
                    Spacer()
                }
                .padding(8)
                Spacer()
            }
            
            // Center suit big symbol
            Text(card.suitSymbol)
                .font(.system(size: 46))
                .shadow(color: Color.white.opacity(0.4), radius: 0, x: 0, y: 1)
            
            // Bottom-right corner
            VStack {
                Spacer()
                HStack {
                    Spacer()
                    VStack(spacing: -4) {
                        Text(card.Rank)
                            .font(.system(size: 18, weight: .bold, design: .rounded))
                        Text(card.suitSymbol)
                            .font(.system(size: 16))
                    }
                    .rotationEffect(.degrees(180))
                }
                .padding(8)
            }
        }
        .foregroundColor(card.isRed ? Color(red: 0.85, green: 0.15, blue: 0.2) : Color.black)
    }
    
    private var cardBack: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .fill(
                    LinearGradient(
                        colors: [
                            Color(red: 0.09, green: 0.24, blue: 0.31),
                            Color(red: 0.06, green: 0.15, blue: 0.20)
                        ],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
            
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(Color(red: 0.86, green: 0.70, blue: 0.38).opacity(0.42), lineWidth: 1)
            
            RoundedRectangle(cornerRadius: 8)
                .stroke(Color.white.opacity(0.15), style: StrokeStyle(lineWidth: 2, dash: [4, 4]))
                .padding(6)

            RoundedRectangle(cornerRadius: 5, style: .continuous)
                .stroke(Color.white.opacity(0.08), lineWidth: 1)
                .padding(14)
            
            VStack(spacing: 4) {
                Image(systemName: "suit.spade.fill")
                    .font(.system(size: 18))
                Text("TRUCO")
                    .font(.system(size: 10, weight: .black, design: .rounded))
                    .tracking(2)
            }
            .foregroundColor(Color(red: 0.86, green: 0.70, blue: 0.38).opacity(0.34))
        }
    }
}
