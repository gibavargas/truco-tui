# Browser Edition (WASM)

Fork web da versão TUI, compilado em WebAssembly para deploy estático.

## Estrutura

- `browser-edition/cmd/wasm/main.go`: bridge Go/WASM com regras do `internal/truco`.
- `browser-edition/web/`: frontend estático (HTML/CSS/JS).
- `browser-edition/scripts/build-web.sh`: gera `browser-edition/dist/`.
- `browser-edition/dist/`: artefatos prontos para deploy.

## Build local

```bash
bash browser-edition/scripts/build-web.sh
```

Isso gera:

- `browser-edition/dist/index.html`
- `browser-edition/dist/app.js`
- `browser-edition/dist/style.css`
- `browser-edition/dist/wasm_exec.js`
- `browser-edition/dist/main.wasm`

## Rodar local

Use qualquer servidor estático apontando para `browser-edition/dist/`.

Exemplo simples (se tiver Python):

```bash
cd browser-edition/dist
python3 -m http.server 8080
```

## Deploy (estático)

Qualquer host estático funciona:

- Netlify: publicar pasta `browser-edition/dist`.
- Vercel: projeto estático com output em `browser-edition/dist`.
- GitHub Pages: publicar conteúdo de `browser-edition/dist`.

## API exposta pelo WASM

Objeto global: `window.TrucoWasm`

- `startGame({ numPlayers, name })`
- `snapshot()`
- `play(cardIndex)`
- `truco()`
- `accept()`
- `refuse()`
- `cpuStep()`
- `autoCpuLoopTick()`
- `newHand()`
- `reset()`
- `startOnlineHost({ numPlayers, name })`
- `joinOnline({ numPlayers, name, key, role })`
- `onlineState()`
- `startOnlineMatch()`
- `sendChat(msg)`
- `sendHostVote(slotNumber)`
- `requestReplacementInvite(slotNumber)`
- `pullOnlineEvents()`
- `leaveSession()`

Retorno padrão:

```json
{ "ok": true, "snapshot": "{...json...}" }
```

ou erro:

```json
{ "ok": false, "error": "mensagem" }
```
