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
        .shadow(color: Color.black.opacity(0.35), radius: 6, x: 0, y: 5)
        .frame(width: 86, height: 124)
        .animation(.spring(response: 0.5, dampingFraction: 0.7), value: isFaceUp)
    }
    
    private var cardFront: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .fill(Color.white)
            
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(Color(white: 0.9), lineWidth: 1)
            
            // Top-left corner
            VStack {
                HStack {
                    VStack(spacing: -4) {
                        Text(card.rank)
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
            
            // Bottom-right corner
            VStack {
                Spacer()
                HStack {
                    Spacer()
                    VStack(spacing: -4) {
                        Text(card.rank)
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
                .fill(Color(red: 0.12, green: 0.25, blue: 0.35))
            
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(Color.white.opacity(0.4), lineWidth: 1)
            
            RoundedRectangle(cornerRadius: 8)
                .stroke(Color.white.opacity(0.15), style: StrokeStyle(lineWidth: 2, dash: [4, 4]))
                .padding(6)
            
            Image(systemName: "suit.spade.fill")
                .font(.system(size: 24))
                .foregroundColor(Color.white.opacity(0.1))
        }
    }
}
