# Browser Edition (PHP + Go API)

Versão web do Truco, com frontend PHP 100% server-rendered (sem JavaScript)
e backend Go HTTP API expondo a lógica do `internal/truco`.

## Arquitetura

```
Browser  ──form POST──▶  PHP Frontend  ──HTTP JSON──▶  Go HTTP API
                         (server-rendered HTML)          (game logic)
```

- **Go microservice** (`browser-edition/cmd/httpapi/`): HTTP server wrapping `internal/truco`.
  Sessões em memória, API JSON com `X-Session-ID` header.
- **PHP frontend** (`browser-edition/php/`): renderiza todo o HTML server-side.
  Ações via `<form>` POST + PRG (Post-Redirect-Get). Zero JavaScript.

## Estrutura

- `browser-edition/cmd/httpapi/main.go`: Go HTTP API server.
- `browser-edition/cmd/httpapi/main_test.go`: testes de integração (26 testes).
- `browser-edition/php/index.php`: router principal PHP.
- `browser-edition/php/api_client.php`: cliente HTTP para a API Go.
- `browser-edition/php/i18n.php`: strings de tradução (pt-BR, en-US).
- `browser-edition/php/views/`: templates renderizados server-side.
- `browser-edition/php/style.css`: estilos visuais.
- `browser-edition/scripts/build-web.sh`: compila a API Go e copia PHP para `dist/`.

## Requisitos

- Go 1.21+
- PHP 8.1+ com extensão curl

## Build

```bash
bash browser-edition/scripts/build-web.sh
```

## Rodar localmente

Em dois terminais:

```bash
# Terminal 1: Go API
go run ./browser-edition/cmd/httpapi/

# Terminal 2: PHP
php -S localhost:8080 -t browser-edition/php
```

Abrir: `http://localhost:8080/index.php`

## Testes

```bash
# Core game logic
go test ./internal/truco/... -v

# HTTP API integration
go test ./browser-edition/cmd/httpapi/... -v
```

## API do Go Microservice

Endpoint: `POST /api/{action}` com `X-Session-ID` header.

### Ações disponíveis

| Ação | Descrição |
|------|-----------|
| `createSession` | Cria sessão nova |
| `startGame` | Inicia partida offline |
| `snapshot` | Retorna estado atual |
| `play` | Joga carta (`cardIndex`) |
| `truco` | Pede truco ou sobe |
| `accept` | Aceita truco |
| `refuse` | Recusa truco |
| `autoCpuLoopTick` | Executa turnos da CPU |
| `newHand` | Inicia nova mão |
| `reset` | Reseta sessão |
| `startOnlineHost` | Cria lobby online |
| `joinOnline` | Entra no lobby |
| `onlineState` | Estado do lobby |
| `startOnlineMatch` | Inicia partida online |
| `sendChat` | Envia mensagem |
| `sendHostVote` | Voto de transferência |
| `requestReplacementInvite` | Convite de reposição |
| `pullOnlineEvents` | Puxa eventos |
| `leaveSession` | Sai da sessão |

Retorno padrão:

```json
{ "ok": true, "snapshot": "{...}" }
```
