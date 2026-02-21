const byId = (id) => document.getElementById(id);

const els = {
  setupPanel: byId("setup-panel"),
  gamePanel: byId("game-panel"),
  playerName: byId("player-name"),
  numPlayers: byId("num-players"),
  btnStart: byId("btn-start"),
  btnTruco: byId("btn-truco"),
  btnAccept: byId("btn-accept"),
  btnRefuse: byId("btn-refuse"),
  btnNewHand: byId("btn-new-hand"),
  scoreLabel: byId("score-label"),
  stakeLabel: byId("stake-label"),
  turnLabel: byId("turn-label"),
  playersLine: byId("players-line"),
  roundCards: byId("round-cards"),
  hand: byId("my-hand"),
  status: byId("status-line"),
  log: byId("log-box"),
};

let currentSnapshot = null;
let cpuTimer = null;

function cardLabel(card) {
  if (!card) return "??";
  return `${card.Rank || "?"}${card.Suit || "?"}`;
}

function safeText(s) {
  return String(s ?? "");
}

function parseSnapshot(payload) {
  if (!payload || !payload.snapshot) return null;
  try {
    return JSON.parse(payload.snapshot);
  } catch (_) {
    return null;
  }
}

function callApi(method, ...args) {
  if (!window.TrucoWasm || typeof window.TrucoWasm[method] !== "function") {
    return { ok: false, error: `API WASM indisponível: ${method}` };
  }
  return window.TrucoWasm[method](...args);
}

function applyPayload(payload) {
  if (!payload || !payload.ok) {
    const msg = payload?.error || "erro desconhecido";
    setStatus(`Erro: ${msg}`, true);
    return false;
  }
  const snap = parseSnapshot(payload);
  if (!snap) {
    setStatus("Erro: snapshot inválido", true);
    return false;
  }
  currentSnapshot = snap;
  renderSnapshot();
  return true;
}

function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle("error", isError);
}

function teamOf(player) {
  return player?.Team ?? -1;
}

function stakePendingLabel(snapshot) {
  const pending = snapshot.PendingRaiseTo || 0;
  if (pending > 0) return pending;
  const curr = snapshot.CurrentHand?.Stake || 1;
  if (curr === 1) return 3;
  if (curr === 3) return 6;
  if (curr === 6) return 9;
  if (curr === 9) return 12;
  return curr;
}

function renderSnapshot() {
  const s = currentSnapshot;
  if (!s) return;

  const t1 = s.MatchPoints?.["0"] ?? s.MatchPoints?.[0] ?? 0;
  const t2 = s.MatchPoints?.["1"] ?? s.MatchPoints?.[1] ?? 0;
  els.scoreLabel.textContent = `T1 ${t1} x ${t2} T2`;
  els.stakeLabel.textContent = `Aposta ${s.CurrentHand?.Stake ?? 1}`;

  const turnId = s.CurrentHand?.Turn ?? -1;
  const turnPlayer = (s.Players || []).find((p) => p.ID === turnId);
  els.turnLabel.textContent = `Vez de ${turnPlayer?.Name || "?"}`;

  els.playersLine.innerHTML = "";
  (s.Players || []).forEach((p) => {
    const chip = document.createElement("span");
    chip.className = "chip";
    if (p.ID === turnId) chip.classList.add("active");
    chip.textContent = `${p.Name} (T${teamOf(p) + 1})`;
    els.playersLine.appendChild(chip);
  });

  els.roundCards.innerHTML = "";
  const played = s.CurrentHand?.RoundCards || [];
  if (played.length === 0) {
    const empty = document.createElement("span");
    empty.className = "muted";
    empty.textContent = "Mesa vazia";
    els.roundCards.appendChild(empty);
  } else {
    played.forEach((pc) => {
      const card = document.createElement("div");
      card.className = "table-card";
      const who = (s.Players || []).find((p) => p.ID === pc.PlayerID)?.Name || `P${pc.PlayerID}`;
      card.textContent = `${who}: ${cardLabel(pc.Card)}`;
      els.roundCards.appendChild(card);
    });
  }

  els.hand.innerHTML = "";
  const me = (s.Players || []).find((p) => p.ID === 0);
  const hand = me?.Hand || [];
  hand.forEach((c, idx) => {
    const btn = document.createElement("button");
    btn.className = "card-btn";
    btn.textContent = cardLabel(c);
    btn.addEventListener("click", () => {
      const payload = callApi("play", idx);
      applyPayload(payload);
      updatePendingStatus();
    });
    els.hand.appendChild(btn);
  });

  els.log.textContent = (s.Logs || []).slice(-14).join("\n");
  updatePendingStatus();
}

function updatePendingStatus() {
  const s = currentSnapshot;
  if (!s) return;
  if (s.MatchFinished) {
    setStatus(`Partida encerrada. Time ${s.WinnerTeam + 1} venceu.`);
    return;
  }
  const me = (s.Players || []).find((p) => p.ID === 0);
  const myTeam = me?.Team ?? 0;
  if (s.PendingRaiseFor === myTeam) {
    setStatus(`⚠ Pedido pendente para ${stakePendingLabel(s)}. Use Aceitar/Recusar/Subir.`);
    return;
  }
  if (s.PendingRaiseFor !== -1) {
    setStatus("Aguardando resposta adversária ao truco...");
    return;
  }
  setStatus("Pronto.");
}

function bindUi() {
  els.btnStart.addEventListener("click", () => {
    const payload = callApi("startGame", {
      numPlayers: Number(els.numPlayers.value || "2"),
      name: safeText(els.playerName.value).trim() || "Você",
    });
    if (!applyPayload(payload)) return;
    els.setupPanel.classList.add("hidden");
    els.gamePanel.classList.remove("hidden");
    startCpuLoop();
  });

  els.btnTruco.addEventListener("click", () => {
    const payload = callApi("truco");
    applyPayload(payload);
  });
  els.btnAccept.addEventListener("click", () => {
    const payload = callApi("accept");
    applyPayload(payload);
  });
  els.btnRefuse.addEventListener("click", () => {
    const payload = callApi("refuse");
    applyPayload(payload);
  });
  els.btnNewHand.addEventListener("click", () => {
    const payload = callApi("newHand");
    applyPayload(payload);
  });
}

function startCpuLoop() {
  if (cpuTimer) clearInterval(cpuTimer);
  cpuTimer = setInterval(() => {
    const payload = callApi("autoCpuLoopTick");
    if (payload?.ok && payload?.changed) {
      applyPayload(payload);
    }
  }, 700);
}

async function bootWasm() {
  if (!("Go" in window)) {
    throw new Error("wasm_exec.js não carregado");
  }
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch("./main.wasm"), go.importObject);
  go.run(result.instance);
}

async function main() {
  setStatus("Carregando runtime WASM...");
  try {
    await bootWasm();
    bindUi();
    setStatus("WASM carregado. Inicie uma partida.");
  } catch (err) {
    setStatus(`Falha ao iniciar WASM: ${err.message || err}`, true);
  }
}

main();
