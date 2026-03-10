# Truco Paulista TUI (Go)

Projeto base de Truco Paulista em terminal (TUI ASCII), com:

- Modo offline (2 ou 4 jogadores) com suporte a CPU.
- Modo multiplayer P2P por troca de chave (host/join) com lobby, chat e partida online completa sincronizada.
- Renderização Bubble Tea + Lipgloss com abas (`mesa`, `chat`, `log`) e atalhos de teclado.
- Reconexão automática de clientes em falhas transitórias de rede (com reentrada no mesmo slot durante a partida).
- Handoff automático de host em queda de conexão do processo host (cliente eleito assume e demais reconectam).
- Transporte P2P protegido com TLS efêmero e validação por fingerprint na chave de convite.
- Em desconexão de jogador remoto durante partida, o slot passa a CPU provisório até reconexão/substituição.
- Transferência democrática de host da mesa por votação (`/host`).
- Estrutura preparada para build cruzado Windows.
- Runtime compartilhado para clientes nativos via FFI em `internal/appcore` + `cmd/truco-core-ffi`.
- Scaffolds nativos em `native/` para macOS SwiftUI, Linux GTK4/libadwaita (Rust) e Windows WinUI 3.

## Requisitos

- Go 1.24+

## Rodar localmente

```bash
go run ./cmd/truco
```

### Relay QUIC (NAT restrito)

Para redes com NAT/firewall restritivo, rode o relay:

```bash
go run ./cmd/truco-relay
```

Variáveis de ambiente:

- `TRUCO_RELAY_HTTP_ADDR` (default: `127.0.0.1:9443`): endpoint HTTPS de controle.
- `TRUCO_RELAY_QUIC_ADDR` (default: `127.0.0.1:9444`): endpoint QUIC de tunelamento fallback.

Observabilidade do relay:

- `GET /healthz`: status + contadores agregados.
- `GET /metrics`: métricas em texto (`truco_relay_*`).

Em produção, exponha as portas públicas HTTPS+QUIC do relay e configure os hosts para criar sessão com `relay_url` no runtime (`create_host_session`).

## Build

### Build do seu SO atual

```bash
go build -o bin/truco ./cmd/truco
```

### Build da biblioteca compartilhada (macOS atual)

```bash
make ffi
```

### Build para Windows (executável `.exe`)

```bash
GOOS=windows GOARCH=amd64 go build -o bin/truco.exe ./cmd/truco
```

Opcional para 32 bits:

```bash
GOOS=windows GOARCH=386 go build -o bin/truco-386.exe ./cmd/truco
```

## Fluxo Multiplayer P2P

1. Um jogador entra em `Multiplayer P2P` e escolhe `Criar host`.
2. O host recebe uma chave codificada e compartilha com os convidados.
3. Quem entra em `Seguir partida` cola a chave e escolhe papel `partner`, `opponent` ou `auto`.
4. O host aloca o slot automaticamente:
   - Em 2 jogadores: slot restante.
   - Em 4 jogadores: prioridade para o papel escolhido quando disponível.
5. O lobby mostra slots em tempo real e possui chat integrado.
6. O host usa `start` para iniciar a partida.
7. Durante a partida online:
   - Host é autoritativo (aplica jogadas e distribui snapshots para todos).
   - Clientes enviam ações (`play`, `truco`, `accept`, `refuse`).
   - Chat segue funcionando durante o jogo.
   - Em perda de conexão, o cliente tenta reconectar automaticamente.
   - Se o processo host cair e não voltar, ocorre failover automático para um novo host eleito.

## Regras implementadas (Truco Paulista)

- Baralho de Truco (40 cartas, sem 8/9/10).
- Vira e manilha dinâmica (próxima carta da vira).
- Hierarquia normal: `3 > 2 > A > K > J > Q > 7 > 6 > 5 > 4`.
- Hierarquia de naipes para manilha: `Paus > Copas > Espadas > Ouros`.
- Aposta da mão: `1 -> 3 -> 6 -> 9 -> 12`.
- Pedido/aceite/corrida de truco.
- Partida até 12 pontos.

## Controles na partida

- `1`, `2`, `3`: jogar carta
- `t`: pedir truco
- `a` / `r`: aceitar / recusar truco
- `tab`: alternar aba `mesa` -> `chat` -> `log`
- `enter` na aba `chat`: enviar mensagem
- `q`: sair da partida

Comandos especiais no chat online:
- `/host <slot>`: votar para transferir o host da mesa (troca por maioria dos jogadores conectados)
- `/invite <slot>`: (host atual da mesa) gerar convite de substituição para slot desconectado em CPU provisório

## Testes

```bash
go test ./...
```

Também há workflow de CI (`.github/workflows/ci.yml`) com:
- `go mod verify`
- `go vet ./...`
- `go test ./...`
- `staticcheck ./...`

## Estrutura

- `cmd/truco/main.go`: bootstrap da aplicação.
- `cmd/truco-core-ffi`: exporta o runtime compartilhado em C ABI para clientes nativos.
- `cmd/truco-relay`: serviço relay QUIC/HTTPS (rendezvous + fallback forwarding).
- `internal/appcore`: runtime headless com intents/eventos JSON, snapshots e orquestração offline/online.
- `internal/netrelay`: cliente do plano de controle do relay + túnel QUIC.
- `internal/netquic`: adaptadores/utilitários QUIC.
- `internal/truco`: regras/cartas/motor/CPU.
- `internal/netp2p`: protocolo de chave + lobby/chat + sincronização da partida online.
- `internal/ui`: frontend Bubble Tea/Lipgloss, animações e modelos offline/online.
- `native/`: scaffolds dos clientes nativos desktop.
