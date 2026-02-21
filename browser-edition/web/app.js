const $ = (id) => document.getElementById(id);

const els = {
  localeSelect: $("locale-select"),

  setupPanel: $("setup-panel"),
  gamePanel: $("game-panel"),
  playerName: $("player-name"),
  numPlayers: $("num-players"),
  btnStart: $("btn-start"),

  scoreT1: $("score-t1"),
  scoreT2: $("score-t2"),
  stakeValue: $("stake-value"),
  stakeLadder: $("stake-ladder"),
  turnLine: $("turn-line"),

  seatTop: $("seat-top"),
  seatLeft: $("seat-left"),
  seatRight: $("seat-right"),
  seatBottom: $("seat-bottom"),
  tableShell: $("table-shell"),

  viraCard: $("vira-card"),
  manilhaCard: $("manilha-card"),
  playedLayer: $("played-layer"),
  animLayer: $("anim-layer"),

  statusLine: $("status-line"),
  trickScore: $("trick-score"),
  trickHistory: $("trick-history"),
  logBox: $("log-box"),
  errorLog: $("error-log"),

  btnTruco: $("btn-truco"),
  btnTrucoText: $("btn-truco-text"),
  btnAccept: $("btn-accept"),
  btnRefuse: $("btn-refuse"),
  btnNewHand: $("btn-new-hand"),
  hand: $("my-hand"),

  trucoOverlay: $("truco-overlay"),
  trucoOverlayTitle: $("truco-overlay-title"),
  trucoOverlaySub: $("truco-overlay-sub"),
  trickOverlay: $("trick-overlay"),
  trickOverlayTitle: $("trick-overlay-title"),
  trickOverlayText: $("trick-overlay-text"),
  matchOverlay: $("match-overlay"),
  matchOverlayTitle: $("match-overlay-title"),
  matchOverlayDetail: $("match-overlay-detail"),
  btnPlayAgain: $("btn-play-again"),

  // labels
  titleMain: $("title-main"),
  titleSub: $("title-sub"),
  localeLabel: $("locale-label"),
  setupTitle: $("setup-title"),
  setupNameLabel: $("setup-name-label"),
  setupPlayersLabel: $("setup-players-label"),
  setupStartLabel: $("setup-start-label"),
  setupHelp: $("setup-help"),
  team1Label: $("team1-label"),
  team2Label: $("team2-label"),
  stakeMainLabel: $("stake-main-label"),
  metaViraLabel: $("meta-vira-label"),
  metaManilhaLabel: $("meta-manilha-label"),
  statusTitle: $("status-title"),
  tricksTitle: $("tricks-title"),
  helpTitle: $("help-title"),
  logTitle: $("log-title"),
  helpPlay: $("help-play"),
  helpTruco: $("help-truco"),
  helpAnswer: $("help-answer"),
  btnAcceptText: $("btn-accept-text"),
  btnRefuseText: $("btn-refuse-text"),
  btnNewHandText: $("btn-new-hand-text"),
  handTitle: $("hand-title"),
  btnPlayAgainText: $("btn-play-again-text"),
};

const POS_COORDS = {
  top: { x: 50, y: 24 },
  right: { x: 73, y: 50 },
  bottom: { x: 50, y: 76 },
  left: { x: 27, y: 50 },
};

const SUIT_SYMBOL = {
  Ouros: "♦",
  Espadas: "♠",
  Copas: "♥",
  Paus: "♣",
};

let locale = "pt-BR";
let snap = null;
let prevSnap = null;
let cpuLoopTimer = null;
let uiFrameTimer = null;
let spinnerFrame = 0;
let errorTrail = [];
let trickOverlayTimer = null;
let trucoOverlayTimer = null;

const i18n = {
  "pt-BR": {
    title_main: "Truco Browser Edition",
    title_sub: "Compatível com regras da TUI (WASM)",
    locale_label: "Idioma",
    setup_title: "Nova partida",
    setup_name: "Seu nome",
    setup_players: "Jogadores",
    setup_start: "Iniciar partida",
    setup_help: "Atalhos: 1/2/3 jogar, T truco/subir, A aceitar, R recusar.",
    team1: "Time 1",
    team2: "Time 2",
    stake: "Aposta",
    vira: "Vira",
    manilha: "Manilha",
    status_title: "Status",
    tricks_title: "Vazas",
    help_title: "Controles",
    log_title: "Log da partida",
    help_play: "jogar carta",
    help_truco: "truco/subir",
    help_answer: "responder",
    btn_truco: "Truco",
    btn_raise: "Subir",
    btn_accept: "Aceitar",
    btn_refuse: "Recusar",
    btn_new_hand: "Nova mão",
    hand_title: "Sua mão",
    btn_play_again: "Jogar novamente",
    status_ready: "Pronto.",
    status_your_turn: "Sua vez. Escolha uma carta.",
    status_wait_cpu: "Aguardando {0}...",
    status_pending_you: "{0} ({1}) pendente: aceitar, recusar ou subir.",
    status_pending_other: "Aguardando resposta de {0} para {1} ({2}).",
    status_match_end: "Partida encerrada.",
    overlay_truco_title: "TRUCO!",
    overlay_truco_sub: "Aposta em disputa",
    overlay_trick_title: "Fim da vaza",
    overlay_trick_collecting: "Recolhendo vaza...",
    overlay_trick_collecting_by: "Recolhendo vaza para {0}...",
    overlay_trick_tie: "Vaza {0} empatou.",
    overlay_trick_own: "Vaza {0}: ponto da sua equipe.",
    overlay_trick_enemy: "Vaza {0}: ponto da equipe adversária.",
    overlay_match_title_win: "Você venceu!",
    overlay_match_title_loss: "Você perdeu.",
    overlay_match_detail: "Placar final: Time 1 {0} x {1} Time 2",
    trick_short: "V",
    trick_tie: "empate",
    loading_wasm: "Carregando runtime WASM...",
    wasm_ready: "WASM carregado. Inicie uma partida.",
    wasm_fail: "Falha ao iniciar WASM: {0}",
    error_prefix: "Erro: ",
    turn_of: "Vez de {0}",
    score_line: "T1 {0} x {1} T2",
    round_empty: "V1 ... | V2 ... | V3 ...",
    cpu_tag: "CPU",
  },
  "en-US": {
    title_main: "Truco Browser Edition",
    title_sub: "Compatible with TUI rules (WASM)",
    locale_label: "Language",
    setup_title: "New match",
    setup_name: "Your name",
    setup_players: "Players",
    setup_start: "Start match",
    setup_help: "Shortcuts: 1/2/3 play, T truco/raise, A accept, R refuse.",
    team1: "Team 1",
    team2: "Team 2",
    stake: "Stake",
    vira: "Flip",
    manilha: "Trump",
    status_title: "Status",
    tricks_title: "Tricks",
    help_title: "Controls",
    log_title: "Match log",
    help_play: "play card",
    help_truco: "truco/raise",
    help_answer: "respond",
    btn_truco: "Truco",
    btn_raise: "Raise",
    btn_accept: "Accept",
    btn_refuse: "Refuse",
    btn_new_hand: "New hand",
    hand_title: "Your hand",
    btn_play_again: "Play again",
    status_ready: "Ready.",
    status_your_turn: "Your turn. Pick a card.",
    status_wait_cpu: "Waiting for {0}...",
    status_pending_you: "{0} ({1}) pending: accept, refuse or raise.",
    status_pending_other: "Waiting for {0} to answer {1} ({2}).",
    status_match_end: "Match ended.",
    overlay_truco_title: "TRUCO!",
    overlay_truco_sub: "Stake in dispute",
    overlay_trick_title: "Trick ended",
    overlay_trick_collecting: "Collecting trick...",
    overlay_trick_collecting_by: "Collecting trick to {0}...",
    overlay_trick_tie: "Trick {0} tied.",
    overlay_trick_own: "Trick {0}: point for your team.",
    overlay_trick_enemy: "Trick {0}: point for opponent team.",
    overlay_match_title_win: "You won!",
    overlay_match_title_loss: "You lost.",
    overlay_match_detail: "Final score: Team 1 {0} x {1} Team 2",
    trick_short: "T",
    trick_tie: "tie",
    loading_wasm: "Loading WASM runtime...",
    wasm_ready: "WASM loaded. Start a match.",
    wasm_fail: "Failed to start WASM: {0}",
    error_prefix: "Error: ",
    turn_of: "Turn: {0}",
    score_line: "T1 {0} x {1} T2",
    round_empty: "T1 ... | T2 ... | T3 ...",
    cpu_tag: "CPU",
  },
};

function tr(key, ...args) {
  const dict = i18n[locale] || i18n["pt-BR"];
  let out = dict[key] || key;
  args.forEach((arg, i) => {
    out = out.replaceAll(`{${i}}`, String(arg));
  });
  return out;
}

function setLocale(nextLocale) {
  locale = i18n[nextLocale] ? nextLocale : "pt-BR";
  if (els.localeSelect.value !== locale) {
    els.localeSelect.value = locale;
  }
  renderStaticLabels();
  if (snap) {
    renderSnapshot();
  } else {
    setStatus(tr("status_ready"), false);
  }
}

function renderStaticLabels() {
  els.titleMain.textContent = tr("title_main");
  els.titleSub.textContent = tr("title_sub");
  els.localeLabel.textContent = tr("locale_label");
  els.setupTitle.textContent = tr("setup_title");
  els.setupNameLabel.textContent = tr("setup_name");
  els.setupPlayersLabel.textContent = tr("setup_players");
  els.setupStartLabel.textContent = tr("setup_start");
  els.setupHelp.textContent = tr("setup_help");
  els.team1Label.textContent = tr("team1");
  els.team2Label.textContent = tr("team2");
  els.stakeMainLabel.textContent = tr("stake");
  els.metaViraLabel.textContent = tr("vira");
  els.metaManilhaLabel.textContent = tr("manilha");
  els.statusTitle.textContent = tr("status_title");
  els.tricksTitle.textContent = tr("tricks_title");
  els.helpTitle.textContent = tr("help_title");
  els.logTitle.textContent = tr("log_title");
  els.helpPlay.textContent = tr("help_play");
  els.helpTruco.textContent = tr("help_truco");
  els.helpAnswer.textContent = tr("help_answer");
  els.btnAcceptText.textContent = tr("btn_accept");
  els.btnRefuseText.textContent = tr("btn_refuse");
  els.btnNewHandText.textContent = tr("btn_new_hand");
  els.handTitle.textContent = tr("hand_title");
  els.btnPlayAgainText.textContent = tr("btn_play_again");
  els.trucoOverlayTitle.textContent = tr("overlay_truco_title");
  els.trucoOverlaySub.textContent = tr("overlay_truco_sub");
  els.trickOverlayTitle.textContent = tr("overlay_trick_title");
}

function callApi(method, ...args) {
  if (!window.TrucoWasm || typeof window.TrucoWasm[method] !== "function") {
    return { ok: false, error: `WASM API unavailable: ${method}` };
  }
  return window.TrucoWasm[method](...args);
}

function parseSnapshot(payload) {
  if (!payload || !payload.snapshot) return null;
  try {
    return JSON.parse(payload.snapshot);
  } catch {
    return null;
  }
}

function playerByID(state, id) {
  return (state?.Players || []).find((p) => p.ID === id);
}

function teamForID(state, id) {
  const p = playerByID(state, id);
  return p?.Team ?? -1;
}

function positionForPlayer(playerID, numPlayers) {
  if (numPlayers === 2) {
    return playerID === 0 ? "bottom" : "top";
  }
  if (playerID === 0) return "bottom";
  if (playerID === 1) return "right";
  if (playerID === 2) return "top";
  if (playerID === 3) return "left";
  return "top";
}

function seatElementByPos(pos) {
  if (pos === "top") return els.seatTop;
  if (pos === "left") return els.seatLeft;
  if (pos === "right") return els.seatRight;
  return els.seatBottom;
}

function suitSymbol(suit) {
  return SUIT_SYMBOL[suit] || "?";
}

function suitColorClass(suit) {
  return suit === "Copas" || suit === "Ouros" ? "red" : "black";
}

function createFaceCard(card, opts = {}) {
  const { small = false, hand = false, keyHint = "" } = opts;
  const el = document.createElement("div");
  el.className = `card ${small ? "small" : ""} ${hand ? "hand" : ""} ${suitColorClass(card?.Suit)}`.trim();

  const top = document.createElement("div");
  top.className = "corner";
  top.innerHTML = `<span>${card?.Rank || "?"}</span><span>${suitSymbol(card?.Suit)}</span>`;

  const bottom = document.createElement("div");
  bottom.className = "corner bottom";
  bottom.innerHTML = `<span>${card?.Rank || "?"}</span><span>${suitSymbol(card?.Suit)}</span>`;

  const center = document.createElement("div");
  center.className = "suit";
  center.textContent = suitSymbol(card?.Suit);

  el.append(top, center, bottom);

  if (keyHint) {
    const hint = document.createElement("div");
    hint.className = "hotkey";
    hint.textContent = keyHint;
    el.appendChild(hint);
  }
  return el;
}

function createBackCard(small = true) {
  const el = document.createElement("div");
  el.className = `card ${small ? "small" : ""} card-back`;
  return el;
}

function stakeLadderLabel(current, pending) {
  const steps = [1, 3, 6, 9, 12];
  return steps
    .map((s) => {
      if (pending === s) return `{${s}}`;
      if (current === s) return `[${s}]`;
      return `${s}`;
    })
    .join(">");
}

function nextStake(curr) {
  if (curr === 1) return 3;
  if (curr === 3) return 6;
  if (curr === 6) return 9;
  if (curr === 9) return 12;
  return curr;
}

function raiseLabel(stake) {
  if (stake === 3) return locale === "en-US" ? "Truco" : "Truco";
  if (stake === 6) return locale === "en-US" ? "Six" : "Seis";
  if (stake === 9) return locale === "en-US" ? "Nine" : "Nove";
  if (stake === 12) return locale === "en-US" ? "Twelve" : "Doze";
  return `${stake}`;
}

function turnMarker(pos) {
  if (pos === "top") return "▼";
  if (pos === "bottom") return "▲";
  if (pos === "left") return "▶";
  if (pos === "right") return "◀";
  return "▶";
}

function setStatus(text, isError = false) {
  els.statusLine.textContent = text;
  els.statusLine.classList.toggle("error", !!isError);
}

function pushError(msg) {
  if (!msg) return;
  const ts = new Date();
  const stamp = `${String(ts.getHours()).padStart(2, "0")}:${String(ts.getMinutes()).padStart(2, "0")}`;
  errorTrail.unshift(`${stamp} ${msg}`);
  errorTrail = errorTrail.slice(0, 4);
  els.errorLog.textContent = errorTrail.join("\n");
}

function applyPayload(payload) {
  if (!payload?.ok) {
    const msg = `${tr("error_prefix")}${payload?.error || "?"}`;
    setStatus(msg, true);
    pushError(msg);
    return false;
  }
  prevSnap = snap;
  snap = parseSnapshot(payload);
  if (!snap) {
    const msg = `${tr("error_prefix")}snapshot inválido`;
    setStatus(msg, true);
    pushError(msg);
    return false;
  }
  renderSnapshot();
  runTransitions(prevSnap, snap);
  return true;
}

function renderSnapshot() {
  const s = snap;
  if (!s) return;

  renderHud(s);
  renderSeats(s);
  renderCenter(s);
  renderSidebar(s);
  renderHand(s);
  updateControlsAndStatus(s);

  if (s.MatchFinished) {
    showMatchOverlay(s);
  } else {
    els.matchOverlay.classList.add("hidden");
  }
}

function renderHud(s) {
  const t1 = s.MatchPoints?.[0] ?? s.MatchPoints?.["0"] ?? 0;
  const t2 = s.MatchPoints?.[1] ?? s.MatchPoints?.["1"] ?? 0;
  els.scoreT1.textContent = `${t1}`;
  els.scoreT2.textContent = `${t2}`;
  const stake = s.CurrentHand?.Stake ?? 1;
  els.stakeValue.textContent = `${stake}`;
  els.stakeLadder.textContent = stakeLadderLabel(stake, s.PendingRaiseTo || 0);

  const turnPlayer = playerByID(s, s.CurrentHand?.Turn ?? -1);
  const isMyTurn = (s.CurrentHand?.Turn ?? -1) === 0;
  const isCpuTurn = !!turnPlayer?.CPU;
  let turnText = tr("turn_of", turnPlayer?.Name || "?");
  if (isCpuTurn && !s.MatchFinished) {
    const frames = ["▶", "▷", "▹", "▸"];
    turnText += ` ${frames[spinnerFrame % frames.length]}`;
  }
  els.turnLine.textContent = turnText;
  els.turnLine.classList.toggle("my-turn", isMyTurn);
}

function renderSeat(pos, contentHTML) {
  const el = seatElementByPos(pos);
  el.innerHTML = contentHTML;
}

function renderSeats(s) {
  const players = [...(s.Players || [])].sort((a, b) => a.ID - b.ID);
  const used = new Set();

  for (const p of players) {
    const pos = positionForPlayer(p.ID, s.NumPlayers);
    used.add(pos);

    const isTurn = p.ID === s.CurrentHand?.Turn;
    const marker = isTurn ? `${turnMarker(pos)} ` : "";
    const cpuTag = p.CPU ? ` · ${tr("cpu_tag")}` : "";
    const label = `${marker}${escapeHtml(p.Name)} (T${(p.Team ?? 0) + 1}${cpuTag})`;

    let cards = "";
    if (p.ID !== 0) {
      const total = (p.Hand || []).length;
      for (let i = 0; i < total; i++) {
        cards += `<span class="tiny-back"></span>`;
      }
    }

    const html = `
      <div class="seat-pill ${isTurn ? "active" : ""}" data-player="${p.ID}">${label}</div>
      <div class="seat-meta">${cards}</div>
    `;
    renderSeat(pos, html);
  }

  for (const pos of ["top", "left", "right", "bottom"]) {
    const el = seatElementByPos(pos);
    if (used.has(pos)) {
      el.classList.remove("hidden");
    } else {
      el.classList.add("hidden");
      el.innerHTML = "";
    }
  }
}

function renderCenter(s) {
  els.viraCard.innerHTML = "";
  if (s.CurrentHand?.Vira) {
    els.viraCard.appendChild(createFaceCard(s.CurrentHand.Vira, { small: true }));
  }
  els.manilhaCard.textContent = `${s.CurrentHand?.Manilha || "-"}`;

  els.playedLayer.innerHTML = "";
  const cards = s.CurrentHand?.RoundCards || [];
  if (cards.length === 0) {
    return;
  }

  for (const pc of cards) {
    const pos = positionForPlayer(pc.PlayerID, s.NumPlayers);
    const anchor = POS_COORDS[pos] || POS_COORDS.top;
    const wrap = document.createElement("div");
    wrap.className = "played-card";
    wrap.style.left = `${anchor.x}%`;
    wrap.style.top = `${anchor.y}%`;

    const owner = document.createElement("div");
    owner.className = "owner";
    owner.textContent = playerByID(s, pc.PlayerID)?.Name || `P${pc.PlayerID}`;

    const card = createFaceCard(pc.Card, { small: true });
    wrap.append(owner, card);
    els.playedLayer.appendChild(wrap);
  }
}

function formatTrickHistory(results) {
  if (!Array.isArray(results) || results.length === 0) {
    return tr("round_empty");
  }
  const myTeam = teamForID(snap, 0);
  const out = [];
  for (let i = 0; i < 3; i++) {
    const r = i < results.length ? results[i] : null;
    if (r === null) {
      out.push(`${tr("trick_short")}${i + 1} ...`);
      continue;
    }
    if (r === -1) {
      out.push(`${tr("trick_short")}${i + 1} ${tr("trick_tie")}`);
      continue;
    }
    out.push(`${tr("trick_short")}${i + 1} ${r === myTeam ? "✓" : "✗"}`);
  }
  return out.join(" | ");
}

function renderSidebar(s) {
  const wins = s.CurrentHand?.TrickWins || { 0: 0, 1: 0 };
  const t1 = wins[0] ?? wins["0"] ?? 0;
  const t2 = wins[1] ?? wins["1"] ?? 0;
  els.trickScore.textContent = tr("score_line", t1, t2);
  els.trickHistory.textContent = formatTrickHistory(s.CurrentHand?.TrickResults || []);

  const logs = s.Logs || [];
  els.logBox.textContent = logs.slice(-22).join("\n");
}

function canPlayCard(s) {
  return !s.MatchFinished && s.CurrentHand?.Turn === 0 && s.PendingRaiseFor === -1;
}

function renderHand(s) {
  els.hand.innerHTML = "";
  const me = playerByID(s, 0);
  const hand = me?.Hand || [];
  const allowPlay = canPlayCard(s);

  hand.forEach((card, idx) => {
    const cardEl = createFaceCard(card, { hand: true, keyHint: `${idx + 1}` });
    if (!allowPlay) {
      cardEl.style.opacity = "0.56";
      cardEl.style.pointerEvents = "none";
    } else {
      cardEl.addEventListener("click", () => {
        applyPayload(callApi("play", idx));
      });
    }
    els.hand.appendChild(cardEl);
  });
}

function updateControlsAndStatus(s) {
  const pendingTeam = s.PendingRaiseFor;
  const myTeam = teamForID(s, 0);
  const pendingForMe = pendingTeam !== -1 && pendingTeam === myTeam;
  const hasPending = pendingTeam !== -1;
  const canTruco = !s.MatchFinished && ((s.CurrentHand?.Turn === 0 && !hasPending) || pendingForMe);

  const pendingTo = s.PendingRaiseTo || nextStake(s.CurrentHand?.Stake || 1);

  els.btnTruco.disabled = !canTruco;
  els.btnAccept.disabled = !pendingForMe || s.MatchFinished;
  els.btnRefuse.disabled = !pendingForMe || s.MatchFinished;
  els.btnNewHand.disabled = s.MatchFinished;

  els.btnTruco.classList.remove("pulse");
  els.btnAccept.classList.remove("pulse");
  els.btnRefuse.classList.remove("pulse");

  if (pendingForMe) {
    els.btnTrucoText.textContent = tr("btn_raise");
    els.btnTruco.classList.add("pulse");
    els.btnAccept.classList.add("pulse");
    els.btnRefuse.classList.add("pulse");
  } else {
    els.btnTrucoText.textContent = tr("btn_truco");
  }

  if (s.MatchFinished) {
    setStatus(tr("status_match_end"), false);
    return;
  }

  if (pendingForMe) {
    setStatus(tr("status_pending_you", raiseLabel(pendingTo), pendingTo), false);
    return;
  }

  if (hasPending) {
    const byName = playerByID(s, s.CurrentHand?.RaiseRequester)?.Name || "?";
    setStatus(tr("status_pending_other", byName, raiseLabel(pendingTo), pendingTo), false);
    return;
  }

  if (s.CurrentHand?.Turn === 0) {
    setStatus(tr("status_your_turn"), false);
  } else {
    const turnName = playerByID(s, s.CurrentHand?.Turn ?? -1)?.Name || "?";
    setStatus(tr("status_wait_cpu", turnName), false);
  }
}

function showMatchOverlay(s) {
  const t1 = s.MatchPoints?.[0] ?? s.MatchPoints?.["0"] ?? 0;
  const t2 = s.MatchPoints?.[1] ?? s.MatchPoints?.["1"] ?? 0;
  const myTeam = teamForID(s, 0);
  const won = s.WinnerTeam === myTeam;

  els.matchOverlayTitle.textContent = won ? tr("overlay_match_title_win") : tr("overlay_match_title_loss");
  els.matchOverlayDetail.textContent = tr("overlay_match_detail", t1, t2);
  els.matchOverlay.classList.remove("hidden");
}

function showTrucoOverlay(pendingTo) {
  els.trucoOverlaySub.textContent = `${tr("overlay_truco_sub")} · ${raiseLabel(pendingTo)} (${pendingTo})`;
  els.trucoOverlay.classList.remove("hidden");
  clearTimeout(trucoOverlayTimer);
  trucoOverlayTimer = setTimeout(() => {
    els.trucoOverlay.classList.add("hidden");
  }, 1000);
}

function showTrickOverlay(state) {
  let msg = tr("overlay_trick_collecting");
  if (state.LastTrickTie) {
    msg = tr("overlay_trick_tie", state.LastTrickRound || 1);
  } else if (typeof state.LastTrickWinner === "number" && state.LastTrickWinner >= 0) {
    const winner = playerByID(state, state.LastTrickWinner)?.Name || "?";
    msg = tr("overlay_trick_collecting_by", winner);
  } else if (state.LastTrickTeam === teamForID(state, 0)) {
    msg = tr("overlay_trick_own", state.LastTrickRound || 1);
  } else {
    msg = tr("overlay_trick_enemy", state.LastTrickRound || 1);
  }

  els.trickOverlayText.textContent = msg;
  els.trickOverlay.classList.remove("hidden");
  clearTimeout(trickOverlayTimer);
  trickOverlayTimer = setTimeout(() => {
    els.trickOverlay.classList.add("hidden");
  }, 980);
}

function seatAnchorPos(pos) {
  const seatEl = seatElementByPos(pos);
  if (!seatEl) return { x: 0, y: 0 };
  const shellRect = els.tableShell.getBoundingClientRect();
  const seatRect = seatEl.getBoundingClientRect();
  return {
    x: seatRect.left - shellRect.left + seatRect.width / 2,
    y: seatRect.top - shellRect.top + seatRect.height / 2,
  };
}

function centerAnchorPos(pos) {
  const shellRect = els.tableShell.getBoundingClientRect();
  const coord = POS_COORDS[pos] || POS_COORDS.top;
  return {
    x: (coord.x / 100) * shellRect.width,
    y: (coord.y / 100) * shellRect.height,
  };
}

function animateFlight(playerID, card, numPlayers) {
  const pos = positionForPlayer(playerID, numPlayers);
  const from = seatAnchorPos(pos);
  const to = centerAnchorPos(pos);

  const floatWrap = document.createElement("div");
  floatWrap.className = "float-card";
  floatWrap.style.left = `${from.x}px`;
  floatWrap.style.top = `${from.y}px`;
  const cardEl = createFaceCard(card, { small: true });
  floatWrap.appendChild(cardEl);
  els.animLayer.appendChild(floatWrap);

  const dx = to.x - from.x;
  const dy = to.y - from.y;
  requestAnimationFrame(() => {
    floatWrap.style.transition = "transform 460ms cubic-bezier(.2,.8,.2,1), opacity 460ms ease";
    floatWrap.style.transform = `translate(${dx}px, ${dy}px) scale(.94)`;
    floatWrap.style.opacity = "0.05";
  });

  setTimeout(() => {
    floatWrap.remove();
  }, 520);
}

function animateSweep(lastWinner, numPlayers) {
  const pile = document.createElement("div");
  pile.className = "float-card";
  pile.style.left = "50%";
  pile.style.top = "50%";

  const stack = document.createElement("div");
  stack.style.position = "relative";
  stack.appendChild(createBackCard(true));
  pile.appendChild(stack);
  els.animLayer.appendChild(pile);

  const winnerPos = lastWinner >= 0 ? positionForPlayer(lastWinner, numPlayers) : "top";
  const target = centerAnchorPos(winnerPos);
  const shellRect = els.tableShell.getBoundingClientRect();
  const from = { x: shellRect.width / 2, y: shellRect.height / 2 };

  const dx = target.x - from.x;
  const dy = target.y - from.y;
  requestAnimationFrame(() => {
    pile.style.transition = "transform 420ms cubic-bezier(.2,.8,.2,1), opacity 420ms ease";
    pile.style.transform = `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(.76)`;
    pile.style.opacity = "0.08";
  });

  setTimeout(() => pile.remove(), 470);
}

function runTransitions(prev, next) {
  if (!prev || !next) return;

  if (prev.PendingRaiseFor === -1 && next.PendingRaiseFor !== -1) {
    const pendingTo = next.PendingRaiseTo || nextStake(next.CurrentHand?.Stake || 1);
    showTrucoOverlay(pendingTo);
  }

  const prevRound = prev.CurrentHand?.Round ?? 0;
  const nextRound = next.CurrentHand?.Round ?? 0;
  const prevLen = prev.CurrentHand?.RoundCards?.length ?? 0;
  const nextLen = next.CurrentHand?.RoundCards?.length ?? 0;
  if (prevRound === nextRound && nextLen === prevLen + 1) {
    const played = next.CurrentHand.RoundCards[nextLen - 1];
    if (played) {
      animateFlight(played.PlayerID, played.Card, next.NumPlayers);
    }
  }

  if ((next.LastTrickSeq || 0) > (prev.LastTrickSeq || 0)) {
    showTrickOverlay(next);
    animateSweep(next.LastTrickWinner ?? -1, next.NumPlayers);
  }
}

function doAction(method, ...args) {
  const payload = callApi(method, ...args);
  applyPayload(payload);
}

function bindUI() {
  els.localeSelect.addEventListener("change", () => {
    setLocale(els.localeSelect.value);
  });

  els.btnStart.addEventListener("click", () => {
    const payload = callApi("startGame", {
      numPlayers: Number(els.numPlayers.value || "2"),
      name: (els.playerName.value || "").trim() || (locale === "en-US" ? "You" : "Você"),
    });
    if (!applyPayload(payload)) return;
    els.setupPanel.classList.add("hidden");
    els.gamePanel.classList.remove("hidden");
    startCpuLoop();
  });

  els.btnTruco.addEventListener("click", () => doAction("truco"));
  els.btnAccept.addEventListener("click", () => doAction("accept"));
  els.btnRefuse.addEventListener("click", () => doAction("refuse"));
  els.btnNewHand.addEventListener("click", () => doAction("newHand"));

  els.btnPlayAgain.addEventListener("click", () => {
    callApi("reset");
    snap = null;
    prevSnap = null;
    els.matchOverlay.classList.add("hidden");
    els.trickOverlay.classList.add("hidden");
    els.trucoOverlay.classList.add("hidden");
    els.gamePanel.classList.add("hidden");
    els.setupPanel.classList.remove("hidden");
    setStatus(tr("status_ready"));
    if (cpuLoopTimer) {
      clearInterval(cpuLoopTimer);
      cpuLoopTimer = null;
    }
  });

  window.addEventListener("keydown", (ev) => {
    if (!snap || els.gamePanel.classList.contains("hidden")) return;
    if (ev.repeat) return;
    const key = ev.key.toLowerCase();
    if (key === "1" || key === "2" || key === "3") {
      doAction("play", Number(key) - 1);
      return;
    }
    if (key === "t") {
      doAction("truco");
      return;
    }
    if (key === "a") {
      doAction("accept");
      return;
    }
    if (key === "r") {
      doAction("refuse");
    }
  });
}

function startCpuLoop() {
  if (cpuLoopTimer) clearInterval(cpuLoopTimer);
  cpuLoopTimer = setInterval(() => {
    const payload = callApi("autoCpuLoopTick");
    if (payload?.ok && payload?.changed) {
      applyPayload(payload);
    }
  }, 700);
}

function startUiTicker() {
  if (uiFrameTimer) clearInterval(uiFrameTimer);
  uiFrameTimer = setInterval(() => {
    spinnerFrame = (spinnerFrame + 1) % 4;
    if (snap && !snap.MatchFinished) {
      renderHud(snap);
    }
  }, 160);
}

async function bootWasm() {
  if (!("Go" in window)) {
    throw new Error("wasm_exec.js not loaded");
  }
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch("./main.wasm"), go.importObject);
  go.run(result.instance);
}

function escapeHtml(value) {
  const span = document.createElement("span");
  span.textContent = value ?? "";
  return span.innerHTML;
}

async function main() {
  setLocale("pt-BR");
  setStatus(tr("loading_wasm"));
  try {
    await bootWasm();
    bindUI();
    startUiTicker();
    setStatus(tr("wasm_ready"));
  } catch (err) {
    const msg = tr("wasm_fail", err?.message || String(err));
    setStatus(msg, true);
    pushError(msg);
  }
}

main();
