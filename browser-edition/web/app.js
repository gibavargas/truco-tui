/* ═══════════════════════════════════════════════════════════
   Truco Browser Edition — App Logic
   ═══════════════════════════════════════════════════════════ */

const $ = (id) => document.getElementById(id);

/* ── DOM refs ── */
const els = {
  setupPanel:  $("setup-panel"),
  gamePanel:   $("game-panel"),
  playerName:  $("player-name"),
  numPlayers:  $("num-players"),
  btnStart:    $("btn-start"),
  btnTruco:    $("btn-truco"),
  btnAccept:   $("btn-accept"),
  btnRefuse:   $("btn-refuse"),
  btnNewHand:  $("btn-new-hand"),
  scoreT1:     $("score-t1"),
  scoreT2:     $("score-t2"),
  stakeLabel:  $("stake-label"),
  turnLabel:   $("turn-label"),
  playersLine: $("players-line"),
  viraCard:    $("vira-card"),
  roundCards:  $("round-cards"),
  tricksRow:   $("tricks-row"),
  hand:        $("my-hand"),
  status:      $("status-line"),
  log:         $("log-box"),
  overlay:     $("match-overlay"),
  overlayTitle: $("overlay-title"),
  overlayDetail: $("overlay-detail"),
  btnPlayAgain: $("btn-play-again"),
};

let snap = null;
let cpuTimer = null;
let prevScores = [0, 0];

/* ── Suit helpers ── */
const SUIT_SYMBOL = {
  Ouros:    "♦",
  Espadas:  "♠",
  Copas:    "♥",
  Paus:     "♣",
};
const SUIT_COLOR = {
  Ouros:   "red",
  Copas:   "red",
  Espadas: "black",
  Paus:    "black",
};

/* ── Card DOM builder ── */
function buildCard(card, extraClasses = "") {
  if (!card) return null;
  const suit = card.Suit || "";
  const rank = card.Rank || "?";
  const symbol = SUIT_SYMBOL[suit] || "?";
  const color = SUIT_COLOR[suit] || "black";

  const el = document.createElement("div");
  el.className = `playing-card playing-card--${color} ${extraClasses}`.trim();

  const rankEl = document.createElement("span");
  rankEl.className = "playing-card__rank";
  rankEl.textContent = rank;

  const suitEl = document.createElement("span");
  suitEl.className = "playing-card__suit";
  suitEl.textContent = symbol;

  el.appendChild(rankEl);
  el.appendChild(suitEl);
  return el;
}

/* ── API wrapper ── */
function callApi(method, ...args) {
  if (!window.TrucoWasm || typeof window.TrucoWasm[method] !== "function") {
    return { ok: false, error: `API WASM indisponível: ${method}` };
  }
  return window.TrucoWasm[method](...args);
}

function parseSnap(payload) {
  if (!payload?.snapshot) return null;
  try { return JSON.parse(payload.snapshot); }
  catch { return null; }
}

function applyPayload(payload) {
  if (!payload?.ok) {
    setStatus(`Erro: ${payload?.error || "desconhecido"}`, true);
    return false;
  }
  const s = parseSnap(payload);
  if (!s) { setStatus("Erro: snapshot inválido", true); return false; }
  snap = s;
  render();
  return true;
}

/* ── Status ── */
function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle("status-bar--error", isError);
}

/* ═══════════════════════════════════════
   RENDERING
   ═══════════════════════════════════════ */
function render() {
  const s = snap;
  if (!s) return;

  /* Score */
  const t1 = s.MatchPoints?.["0"] ?? s.MatchPoints?.[0] ?? 0;
  const t2 = s.MatchPoints?.["1"] ?? s.MatchPoints?.[1] ?? 0;
  animateScore(els.scoreT1, prevScores[0], t1);
  animateScore(els.scoreT2, prevScores[1], t2);
  prevScores = [t1, t2];

  /* Stake */
  els.stakeLabel.textContent = s.CurrentHand?.Stake ?? 1;

  /* Turn */
  const turnId = s.CurrentHand?.Turn ?? -1;
  const turnPlayer = (s.Players || []).find(p => p.ID === turnId);
  const isMyTurn = turnId === 0;
  els.turnLabel.innerHTML = `<span class="turn-bar__dot"></span> Vez de ${turnPlayer?.Name || "?"}`;
  els.turnLabel.classList.toggle("turn-bar--mine", isMyTurn);

  /* Players */
  els.playersLine.innerHTML = "";
  (s.Players || []).forEach(p => {
    const chip = document.createElement("span");
    chip.className = `player-chip${p.ID === turnId ? " player-chip--active" : ""}`;
    chip.innerHTML = `<span class="player-chip__dot"></span>${esc(p.Name)} <small style="opacity:.5">T${(p.Team ?? 0) + 1}</small>`;
    els.playersLine.appendChild(chip);
  });

  /* Vira */
  els.viraCard.innerHTML = "";
  if (s.CurrentHand?.Vira) {
    const vc = buildCard(s.CurrentHand.Vira, "playing-card--vira");
    if (vc) els.viraCard.appendChild(vc);
  }

  /* Round cards (table) */
  els.roundCards.innerHTML = "";
  const played = s.CurrentHand?.RoundCards || [];
  if (played.length === 0) {
    const empty = document.createElement("span");
    empty.className = "muted";
    empty.textContent = "Mesa vazia";
    els.roundCards.appendChild(empty);
  } else {
    played.forEach((pc, i) => {
      const wrapper = document.createElement("div");
      wrapper.className = "table-card-wrapper";

      const nameEl = document.createElement("span");
      nameEl.className = "table-card-wrapper__name";
      const who = (s.Players || []).find(p => p.ID === pc.PlayerID)?.Name || `P${pc.PlayerID}`;
      nameEl.textContent = who;

      const card = buildCard(pc.Card, "playing-card--sm playing-card--deal");
      card.style.animationDelay = `${i * 0.08}s`;

      wrapper.appendChild(nameEl);
      wrapper.appendChild(card);
      els.roundCards.appendChild(wrapper);
    });
  }

  /* Trick results */
  els.tricksRow.innerHTML = "";
  const tricks = s.CurrentHand?.TrickResults || [];
  const myTeam = (s.Players || []).find(p => p.ID === 0)?.Team ?? 0;
  tricks.forEach((t, i) => {
    const badge = document.createElement("span");
    let cls = "trick-badge";
    let label;
    if (t === -1) { cls += " trick-badge--tie"; label = `R${i + 1}: Empate`; }
    else if (t === myTeam) { cls += " trick-badge--won"; label = `R${i + 1}: Ganhou`; }
    else { cls += " trick-badge--lost"; label = `R${i + 1}: Perdeu`; }
    badge.className = cls;
    badge.textContent = label;
    els.tricksRow.appendChild(badge);
  });

  /* Hand */
  els.hand.innerHTML = "";
  const me = (s.Players || []).find(p => p.ID === 0);
  const hand = me?.Hand || [];
  hand.forEach((c, idx) => {
    const card = buildCard(c, "playing-card--hand playing-card--deal");
    card.style.animationDelay = `${idx * 0.1}s`;
    card.addEventListener("click", () => {
      card.classList.add("playing-card--play");
      setTimeout(() => {
        applyPayload(callApi("play", idx));
        updatePending();
      }, 150);
    });
    els.hand.appendChild(card);
  });

  /* Log */
  els.log.textContent = (s.Logs || []).slice(-20).join("\n");

  /* Pending status */
  updatePending();

  /* Match finished? */
  if (s.MatchFinished) {
    showMatchEnd(s);
  }
}

/* ── Pending truco status ── */
function updatePending() {
  if (!snap) return;
  const s = snap;

  /* Remove previous pulse */
  els.btnAccept.classList.remove("btn--pulse");
  els.btnRefuse.classList.remove("btn--pulse");
  els.btnTruco.classList.remove("btn--pulse");

  if (s.MatchFinished) {
    setStatus(`Partida encerrada.`);
    return;
  }

  const myTeam = (s.Players || []).find(p => p.ID === 0)?.Team ?? 0;

  if (s.PendingRaiseFor === myTeam) {
    const target = pendingLabel(s);
    setStatus(`⚡ Pedido de ${target} pendente — Aceitar, Recusar ou Subir!`);
    els.btnAccept.classList.add("btn--pulse");
    els.btnRefuse.classList.add("btn--pulse");
    return;
  }
  if (s.PendingRaiseFor !== -1) {
    setStatus("⏳ Aguardando resposta adversária ao truco...");
    return;
  }

  const turnId = s.CurrentHand?.Turn ?? -1;
  if (turnId === 0) {
    setStatus("🎯 Sua vez — clique em uma carta para jogar!");
  } else {
    setStatus("Pronto.");
  }
}

function pendingLabel(s) {
  const pending = s.PendingRaiseTo || 0;
  if (pending > 0) return `${pending}`;
  const curr = s.CurrentHand?.Stake || 1;
  const map = { 1: "Truco (3)", 3: "Seis (6)", 6: "Nove (9)", 9: "Doze (12)" };
  return map[curr] || `${curr}`;
}

/* ── Score animation ── */
function animateScore(el, from, to) {
  el.textContent = to;
  if (from !== to) {
    el.classList.remove("score-pop");
    void el.offsetWidth; // reflow
    el.classList.add("score-pop");
  }
}

/* ── Match end overlay ── */
function showMatchEnd(s) {
  const t1 = s.MatchPoints?.["0"] ?? s.MatchPoints?.[0] ?? 0;
  const t2 = s.MatchPoints?.["1"] ?? s.MatchPoints?.[1] ?? 0;
  const myTeam = (s.Players || []).find(p => p.ID === 0)?.Team ?? 0;
  const won = s.WinnerTeam === myTeam;

  els.overlayTitle.textContent = won ? "🏆 Você Venceu!" : "😞 Você Perdeu";
  els.overlayDetail.textContent = `Placar final: Time 1 ${t1} × ${t2} Time 2`;
  els.overlay.classList.remove("hidden");
}

/* ── Escape HTML ── */
function esc(s) {
  const d = document.createElement("span");
  d.textContent = s ?? "";
  return d.innerHTML;
}

/* ═══════════════════════════════════════
   EVENT BINDING
   ═══════════════════════════════════════ */
function bindUi() {
  els.btnStart.addEventListener("click", () => {
    const payload = callApi("startGame", {
      numPlayers: Number(els.numPlayers.value || "2"),
      name: (els.playerName.value || "").trim() || "Você",
    });
    if (!applyPayload(payload)) return;
    els.setupPanel.classList.add("hidden");
    els.gamePanel.classList.remove("hidden");
    startCpuLoop();
  });

  els.btnTruco.addEventListener("click", () => applyPayload(callApi("truco")));
  els.btnAccept.addEventListener("click", () => applyPayload(callApi("accept")));
  els.btnRefuse.addEventListener("click", () => applyPayload(callApi("refuse")));
  els.btnNewHand.addEventListener("click", () => applyPayload(callApi("newHand")));

  els.btnPlayAgain.addEventListener("click", () => {
    els.overlay.classList.add("hidden");
    applyPayload(callApi("reset"));
    els.gamePanel.classList.add("hidden");
    els.setupPanel.classList.remove("hidden");
    prevScores = [0, 0];
    if (cpuTimer) { clearInterval(cpuTimer); cpuTimer = null; }
  });
}

/* ── CPU loop ── */
function startCpuLoop() {
  if (cpuTimer) clearInterval(cpuTimer);
  cpuTimer = setInterval(() => {
    const payload = callApi("autoCpuLoopTick");
    if (payload?.ok && payload?.changed) {
      applyPayload(payload);
    }
  }, 700);
}

/* ═══════════════════════════════════════
   BOOT
   ═══════════════════════════════════════ */
async function bootWasm() {
  if (!("Go" in window)) throw new Error("wasm_exec.js não carregado");
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch("./main.wasm"), go.importObject);
  go.run(result.instance);
}

async function main() {
  setStatus("Carregando runtime WASM...");
  try {
    await bootWasm();
    bindUi();
    setStatus("WASM carregado. Inicie uma partida!");
  } catch (err) {
    setStatus(`Falha ao iniciar WASM: ${err.message || err}`, true);
  }
}

main();
