import SwiftUI

struct OnlineLobbyView: View {
    @EnvironmentObject var store: TrucoAppStore
    @State private var chatMessage: String = ""
    
    var body: some View {
        HStack(spacing: 0) {
            // Left Column: Lobby state and players
            VStack(alignment: .leading, spacing: 20) {
                Text(store.mode == "host_lobby" ? "Sala Criada" : "Conectado")
                    .font(.largeTitle.weight(.black))
                
                if let key = store.bundle?.lobby?.invite_key {
                    HStack {
                        Text("Chave:")
                            .foregroundColor(.secondary)
                        Text(key)
                            .font(.system(size: 18, weight: .bold, design: .monospaced))
                            .foregroundColor(.yellow)
                    }
                    .padding()
                    .background(Color.black.opacity(0.2))
                    .cornerRadius(8)
                }
                
                Text("Jogadores")
                    .font(.headline)
                
                VStack(spacing: 8) {
                    let slots = store.bundle?.lobby?.slots ?? []
                    let assignedSeat = store.bundle?.lobby?.assigned_seat ?? -1
                    let isHost = store.mode == "host_lobby"
                    
                    ForEach(0..<slots.count, id: \.self) { i in
                        let name = slots[i]
                        let isEmpty = name.isEmpty
                        
                        HStack {
                            Text(isEmpty ? "Aguardando..." : name)
                                .foregroundColor(isEmpty ? .secondary : .primary)
                            
                            if i == assignedSeat {
                                Text("(você)")
                                    .font(.caption.bold())
                                    .foregroundColor(.yellow)
                            }
                            
                            Spacer()
                            
                            if isEmpty && isHost {
                                Button("Convidar CPU") {
                                    store.requestReplacementInvite(targetSeat: i)
                                }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                            } else if !isEmpty && i != assignedSeat {
                                Button("Votar Host") {
                                    store.voteHost(candidateSeat: i)
                                }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                            }
                        }
                        .padding()
                        .background(Color.primary.opacity(0.05))
                        .cornerRadius(8)
                    }
                }
                
                Spacer()
                
                if store.mode == "host_lobby" {
                    Button(action: {
                        store.startHostedMatch()
                    }) {
                        Text("Iniciar Partida")
                            .font(.headline.bold())
                            .frame(maxWidth: .infinity)
                            .padding()
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(.green)
                }
                
                Button(action: {
                    store.closeSession()
                }) {
                    Text("Sair da Sala")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                }
                .buttonStyle(.borderedProminent)
                .tint(.red)
            }
            .padding(32)
            .frame(maxWidth: 400)
            
            Divider()
            
            // Right Column: Chat
            VStack(spacing: 0) {
                Text("Chat")
                    .font(.headline)
                    .padding()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.black.opacity(0.2))
                
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 8) {
                            ForEach(store.events) { event in
                                if event.kind == "chat" {
                                    let author = event.payload?.author ?? "?"
                                    let msg = event.payload?.text ?? ""
                                    Text("\(author): \(msg)")
                                        .padding(.horizontal)
                                } else if event.kind == "system" {
                                    let msg = event.payload?.text ?? ""
                                    Text(msg)
                                        .italic()
                                        .foregroundColor(.secondary)
                                        .padding(.horizontal)
                                } else if event.kind == "replacement_invite" {
                                    let link = event.payload?.invite_key ?? ""
                                    Text("Link de subs: \(link)")
                                        .foregroundColor(.green)
                                        .padding(.horizontal)
                                }
                            }
                        }
                        .padding(.vertical)
                    }
                    .onChange(of: store.events.count) { _ in
                        if let last = store.events.last {
                            proxy.scrollTo(last.id, anchor: .bottom)
                        }
                    }
                }
                
                HStack {
                    TextField("Mensagem...", text: $chatMessage)
                        .textFieldStyle(.roundedBorder)
                        .onSubmit {
                            sendChat()
                        }
                    
                    Button("Enviar") {
                        sendChat()
                    }
                    .keyboardShortcut(.defaultAction)
                }
                .padding()
                .background(Color.black.opacity(0.2))
            }
            .frame(maxWidth: .infinity)
        }
        .frame(minWidth: 700, minHeight: 450)
        .background(Color(white: 0.12))
    }
    
    private func sendChat() {
        guard !chatMessage.isEmpty else { return }
        store.sendChat(text: chatMessage)
        chatMessage = ""
    }
}
