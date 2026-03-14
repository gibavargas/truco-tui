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
                    let slotStates = store.bundle?.ui?.lobby_slots ?? []
                    
                    ForEach(slotStates) { slot in
                        let name = slot.name ?? ""
                        let isEmpty = slot.is_empty
                        
                        HStack {
                            Text(isEmpty ? "Aguardando..." : name)
                                .foregroundColor(isEmpty ? .secondary : .primary)
                            
                            if slot.is_local {
                                Text("(você)")
                                    .font(.caption.bold())
                                    .foregroundColor(.yellow)
                            }
                            if slot.is_host {
                                Text("host")
                                    .font(.caption2.bold())
                                    .foregroundColor(.blue)
                            }
                            if slot.is_provisional_cpu {
                                Text("cpu")
                                    .font(.caption2.bold())
                                    .foregroundColor(.orange)
                            }
                            
                            Spacer()
                            
                            if slot.can_request_replacement {
                                Button("Convite") {
                                    store.requestReplacementInvite(targetSeat: slot.seat)
                                }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                            } else if slot.can_vote_host {
                                Button("Votar Host") {
                                    store.voteHost(candidateSeat: slot.seat)
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
