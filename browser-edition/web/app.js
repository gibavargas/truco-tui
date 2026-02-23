const $ = (id) => document.getElementById(id);

const els = {
  localeSelect: $("locale-select"),

  setupPanel: $("setup-panel"),
  lobbyPanel: $("lobby-panel"),
  gamePanel: $("game-panel"),
  playerName: $("player-name"),
  modeSelect: $("mode-select"),
  onlineSetup: $("online-setup"),
  onlineAction: $("online-action"),
  onlineRole: $("online-role"),
  onlineKey: $("online-key"),
  numPlayers: $("num-players"),
  btnStart: $("btn-start"),
  setupModeLabel: $("setup-mode-label"),
  setupOnlineActionLabel: $("setup-online-action-label"),
  setupOnlineRoleLabel: $("setup-online-role-label"),
  setupOnlineKeyLabel: $("setup-online-key-label"),

  lobbyTitle: $("lobby-title"),
  lobbyMode: $("lobby-mode"),
  lobbyKeyLabel: $("lobby-key-label"),
  lobbyKey: $("lobby-key"),
  lobbySlotsTitle: $("lobby-slots-title"),
  lobbySlots: $("lobby-slots"),
  lobbyEventsTitle: $("lobby-events-title"),
  lobbyEvents: $("lobby-events"),
  lobbyChatInput: $("lobby-chat-input"),
  btnLobbyChat: $("btn-lobby-chat"),
  btnLobbyChatText: $("btn-lobby-chat-text"),
  btnLobbyStart: $("btn-lobby-start"),
  btnLobbyStartText: $("btn-lobby-start-text"),
  btnLobbyRefresh: $("btn-lobby-refresh"),
  btnLobbyRefreshText: $("btn-lobby-refresh-text"),
  btnLobbyLeave: $("btn-lobby-leave"),
  btnLobbyLeaveText: $("btn-lobby-leave-text"),

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
  btnHostVote: $("btn-host-vote"),
  btnHostVoteText: $("btn-host-vote-text"),
  btnReplaceInvite: $("btn-replace-invite"),
  btnReplaceInviteText: $("btn-replace-invite-text"),
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
const STAKE_STEPS = [1, 3, 6, 9, 12];

let locale = "pt-BR";
let snap = null;
let prevSnap = null;
let cpuLoopTimer = null;
let uiFrameTimer = null;
let onlinePollTimer = null;
let spinnerFrame = 0;
let errorTrail = [];
let trickOverlayTimer = null;
let trucoOverlayTimer = null;
let mode = "offline";
let onlineSession = null;
let onlineEventTrail = [];
let handFocusIndex = 0;

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
    setup_mode: "Modo",
    setup_mode_offline: "Offline",
    setup_mode_online: "Online (Alpha)",
    setup_online_action: "Ação online",
    setup_online_role: "Papel",
    setup_online_key: "Convite",
    online_action_host: "Criar host",
    online_action_join: "Entrar com convite",
    online_role_auto: "Auto",
    online_role_partner: "Parceiro",
    online_role_opponent: "Adversário",
    setup_start_online: "Abrir lobby",
    lobby_title: "Lobby online",
    lobby_mode_host: "Host",
    lobby_mode_client: "Cliente",
    lobby_invite: "Convite",
    lobby_slots_title: "Slots",
    lobby_events_title: "Eventos",
    lobby_chat_send: "Enviar",
    lobby_start: "Iniciar partida",
    lobby_enter: "Entrar na partida",
    lobby_refresh: "Atualizar",
    lobby_leave: "Sair do lobby",
    btn_host_vote: "Votar host",
    btn_replace_invite: "Convite reposição",
    prompt_slot_vote: "Slot para voto de host (1-{0})",
    prompt_slot_replace: "Slot para convite de reposição (1-{0})",
    online_invite_generated: "Convite gerado para slot {0}: {1}",
    lobby_slot_empty: "vazio",
    lobby_events_empty: "Sem eventos.",
    status_invalid_snapshot: "snapshot inválido",
    role_you: "Você",
    role_partner: "Parceiro",
    role_opponent: "Adversário",
    role_mao: "Mão",
    role_pe: "Pé",
    role_turn: "Vez",
    role_host: "Host",
    role_cpu_prov: "CPU provisório",
    role_cpu: "CPU",
    seat_slot: "Slot {0}",
    online_mode_notice_host: "Lobby host criado. Compartilhe a chave.",
    online_mode_notice_join: "Lobby cliente conectado.",
    online_alpha_notice: "Sessão online alpha em runtime local.",
    online_api_missing: "API online indisponível: {0}",
    api_unavailable: "API indisponível: {0}",
    card_aria_label: "Carta {0}: {1} de {2}",
    online_reconnecting: "Reconectando...",
    online_degraded: "Conexão degradada",
    online_handoff: "Handoff de host detectado",
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
    setup_mode: "Mode",
    setup_mode_offline: "Offline",
    setup_mode_online: "Online (Alpha)",
    setup_online_action: "Online action",
    setup_online_role: "Role",
    setup_online_key: "Invite",
    online_action_host: "Create host",
    online_action_join: "Join with invite",
    online_role_auto: "Auto",
    online_role_partner: "Partner",
    online_role_opponent: "Opponent",
    setup_start_online: "Open lobby",
    lobby_title: "Online lobby",
    lobby_mode_host: "Host",
    lobby_mode_client: "Client",
    lobby_invite: "Invite",
    lobby_slots_title: "Slots",
    lobby_events_title: "Events",
    lobby_chat_send: "Send",
    lobby_start: "Start match",
    lobby_enter: "Enter match",
    lobby_refresh: "Refresh",
    lobby_leave: "Leave lobby",
    btn_host_vote: "Vote host",
    btn_replace_invite: "Replacement invite",
    prompt_slot_vote: "Slot for host vote (1-{0})",
    prompt_slot_replace: "Slot for replacement invite (1-{0})",
    online_invite_generated: "Invite generated for slot {0}: {1}",
    lobby_slot_empty: "empty",
    lobby_events_empty: "No events.",
    status_invalid_snapshot: "invalid snapshot",
    role_you: "You",
    role_partner: "Partner",
    role_opponent: "Opponent",
    role_mao: "Mao",
    role_pe: "Pe",
    role_turn: "Turn",
    role_host: "Host",
    role_cpu_prov: "Provisional CPU",
    role_cpu: "CPU",
    seat_slot: "Slot {0}",
    online_mode_notice_host: "Host lobby created. Share the key.",
    online_mode_notice_join: "Client lobby connected.",
    online_alpha_notice: "Online alpha session in local runtime.",
    online_api_missing: "Online API unavailable: {0}",
    api_unavailable: "API unavailable: {0}",
    card_aria_label: "Card {0}: {1} of {2}",
    online_reconnecting: "Reconnecting...",
    online_degraded: "Degraded connection",
    online_handoff: "Host handoff detected",
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
  } else if (onlineSession) {
    renderLobby();
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
  els.setupModeLabel.textContent = tr("setup_mode");
  els.setupPlayersLabel.textContent = tr("setup_players");
  els.setupOnlineActionLabel.textContent = tr("setup_online_action");
  els.setupOnlineRoleLabel.textContent = tr("setup_online_role");
  els.setupOnlineKeyLabel.textContent = tr("setup_online_key");
  els.setupStartLabel.textContent = mode === "online" ? tr("setup_start_online") : tr("setup_start");
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
  els.btnHostVoteText.textContent = tr("btn_host_vote");
  els.btnReplaceInviteText.textContent = tr("btn_replace_invite");
  els.handTitle.textContent = tr("hand_title");
  els.btnPlayAgainText.textContent = tr("btn_play_again");
  els.trucoOverlayTitle.textContent = tr("overlay_truco_title");
  els.trucoOverlaySub.textContent = tr("overlay_truco_sub");
  els.trickOverlayTitle.textContent = tr("overlay_trick_title");
  els.lobbyTitle.textContent = tr("lobby_title");
  els.lobbyKeyLabel.textContent = tr("lobby_invite");
  els.lobbySlotsTitle.textContent = tr("lobby_slots_title");
  els.lobbyEventsTitle.textContent = tr("lobby_events_title");
  els.btnLobbyChatText.textContent = tr("lobby_chat_send");
  els.btnLobbyRefreshText.textContent = tr("lobby_refresh");
  els.btnLobbyLeaveText.textContent = tr("lobby_leave");
  updateOnlineSelectLabels();
  updateOnlineActionUI();
  updateLobbyLabels();
}

function callApi(method, ...args) {
  if (!window.TrucoWasm || typeof window.TrucoWasm[method] !== "function") {
    return { ok: false, error: tr("api_unavailable", method) };
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

function teamPoints(state, team) {
  return state?.MatchPoints?.[team] ?? state?.MatchPoints?.[String(team)] ?? 0;
}

function localPlayerID(state) {
  const idx = state?.CurrentPlayerIdx ?? 0;
  return state?.Players?.[idx]?.ID ?? 0;
}

function localTeamID(state) {
  return teamForID(state, localPlayerID(state));
}

function playerIndexByID(state, playerID) {
  return (state?.Players || []).findIndex((p) => p.ID === playerID);
}

function deriveSeatRoles(state, localID = 0, hostSeat = 0) {
  const out = {};
  const players = state?.Players || [];
  if (players.length === 0) return out;
  const localPlayer = playerByID(state, localID) || players[0];
  const localTeam = localPlayer?.Team ?? 0;
  const roundStart = state?.CurrentHand?.RoundStart ?? players[0].ID;
  const roundStartIdx = Math.max(0, playerIndexByID(state, roundStart));
  const n = players.length;

  for (const p of players) {
    const info = {
      teamRelation: "opponent",
      turnRole: "",
      trickRole: "",
      connectivityRole: "",
      governanceRole: "",
    };
    if (p.ID === localPlayer.ID) info.teamRelation = "self";
    else if (p.Team === localTeam) info.teamRelation = "partner";

    const idx = playerIndexByID(state, p.ID);
    if (idx >= 0) {
      const dist = (idx - roundStartIdx + n) % n;
      if (dist === 0) info.turnRole = "mao";
      if (dist === n - 1) info.turnRole = info.turnRole ? `${info.turnRole},pe` : "pe";
    }
    if (p.ID === state?.CurrentHand?.Turn) info.trickRole = "turn";
    if (p.ProvisionalCPU) info.connectivityRole = "provisional_cpu";
    else if (p.CPU) info.connectivityRole = "cpu";
    if (hostSeat >= 0 && p.ID === hostSeat) info.governanceRole = "host";
    out[p.ID] = info;
  }
  return out;
}

function roleLabels(info) {
  const out = [];
  if (!info) return out;
  if (info.teamRelation === "self") out.push(tr("role_you"));
  else if (info.teamRelation === "partner") out.push(tr("role_partner"));
  else out.push(tr("role_opponent"));
  if ((info.turnRole || "").includes("mao")) out.push(tr("role_mao"));
  if ((info.turnRole || "").includes("pe")) out.push(tr("role_pe"));
  if (info.trickRole === "turn") out.push(tr("role_turn"));
  if (info.governanceRole === "host") out.push(tr("role_host"));
  if (info.connectivityRole === "provisional_cpu") out.push(tr("role_cpu_prov"));
  else if (info.connectivityRole === "cpu") out.push(tr("role_cpu"));
  return out;
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

function renderStakeLadder(current, pending) {
  const safeCurrent = STAKE_STEPS.includes(current) ? current : 1;
  const safePending = STAKE_STEPS.includes(pending) ? pending : 0;
  const html = STAKE_STEPS.map((step) => {
    let stateClass = "future";
    if (safePending === step) {
      stateClass = "pending";
    } else if (step === safeCurrent) {
      stateClass = "current";
    } else if (step < safeCurrent) {
      stateClass = "done";
    }
    return `<span class="stake-step ${stateClass}" data-step="${step}"><span class="stake-dot"></span><span class="stake-label">${step}</span></span>`;
  }).join("");
  els.stakeLadder.innerHTML = html;
  els.stakeLadder.dataset.current = `${safeCurrent}`;
  els.stakeLadder.dataset.pending = `${safePending}`;
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

function triggerAnimationClass(el, className) {
  if (!el) return;
  el.classList.remove(className);
  void el.offsetWidth;
  el.classList.add(className);
}

function seatAvatar(player) {
  if (player?.ID === 0) return "★";
  if (player?.CPU) return "🤖";
  const name = (player?.Name || "").trim();
  return (name[0] || "?").toUpperCase();
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
  if (payload?.session) {
    onlineSession = payload.session;
  }
  prevSnap = snap;
  snap = parseSnapshot(payload);
  if (!snap) {
    const msg = `${tr("error_prefix")}${tr("status_invalid_snapshot")}`;
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
  const t1 = teamPoints(s, 0);
  const t2 = teamPoints(s, 1);
  els.scoreT1.textContent = `${t1}`;
  els.scoreT2.textContent = `${t2}`;
  const stake = s.CurrentHand?.Stake ?? 1;
  const pendingStake = s.PendingRaiseTo || 0;
  els.stakeValue.textContent = `${stake}`;
  const ladderCurrent = Number(els.stakeLadder.dataset.current || "-1");
  const ladderPending = Number(els.stakeLadder.dataset.pending || "-1");
  if (ladderCurrent !== stake || ladderPending !== pendingStake) {
    renderStakeLadder(stake, pendingStake);
  }

  const turnPlayer = playerByID(s, s.CurrentHand?.Turn ?? -1);
  const myID = localPlayerID(s);
  const isMyTurn = (s.CurrentHand?.Turn ?? -1) === myID;
  const isCpuTurn = !!turnPlayer?.CPU;
  let turnText = tr("turn_of", turnPlayer?.Name || "?");
  if (isCpuTurn && !s.MatchFinished && (s.PendingRaiseFor ?? -1) === -1) {
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
  const hostSeat = onlineSession ? (onlineSession.HostSeat ?? onlineSession.hostSeat ?? 0) : -1;
  const rolesByPlayer = deriveSeatRoles(s, localPlayerID(s), hostSeat);

  for (const p of players) {
    const pos = positionForPlayer(p.ID, s.NumPlayers);
    used.add(pos);

    const isTurn = p.ID === s.CurrentHand?.Turn;
    const teamNum = (p.Team ?? 0) + 1;
    const cpuTag = p.CPU ? ` · ${escapeHtml(tr("cpu_tag"))}` : "";
    const turn = isTurn ? `<span class="turn-flag">${turnMarker(pos)}</span>` : "";
    const roles = roleLabels(rolesByPlayer[p.ID] || {});
    const roleHtml = roles
      .map((r, idx) => `<span class="seat-role ${idx === 0 ? "primary" : ""}">${escapeHtml(r)}</span>`)
      .join("");

    let cards = "";
    if (p.ID !== 0) {
      const total = (p.Hand || []).length;
      for (let i = 0; i < total; i++) {
        cards += `<span class="tiny-back"></span>`;
      }
    }

    const html = `
      <div class="seat-head">
        <div class="seat-avatar team-${teamNum}" data-player="${p.ID}">${escapeHtml(seatAvatar(p))}</div>
        <div class="seat-pill team-${teamNum} ${isTurn ? "active" : ""}" data-player="${p.ID}">
          <span class="seat-name">${escapeHtml(p.Name)}</span>
          <span class="seat-team">T${teamNum}${cpuTag}</span>
        </div>
      </div>
      <div class="seat-meta">${turn}${cards}</div>
      <div class="seat-roles">${roleHtml}</div>
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
  const manilha = `${s.CurrentHand?.Manilha || "-"}`;
  if (manilha === "-") {
    els.manilhaCard.textContent = manilha;
    els.manilhaCard.classList.remove("hot");
  } else {
    els.manilhaCard.innerHTML = `<span class="spark">✦</span><strong>${escapeHtml(manilha)}</strong><span class="spark">✦</span>`;
    els.manilhaCard.classList.add("hot");
  }

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

function renderTrickHistory(results) {
  const myTeam = localTeamID(snap);
  const out = [];
  for (let i = 0; i < 3; i++) {
    const r = i < results.length ? results[i] : null;
    const prefix = `${tr("trick_short")}${i + 1}`;
    if (r === null) {
      out.push(`<span class="trick-badge pending">${prefix} •••</span>`);
      continue;
    }
    if (r === -1) {
      out.push(`<span class="trick-badge tie">${prefix} =</span>`);
      continue;
    }
    out.push(`<span class="trick-badge ${r === myTeam ? "win" : "loss"}">${prefix} ${r === myTeam ? "✓" : "✗"}</span>`);
  }
  return out.join("");
}

function renderSidebar(s) {
  const wins = s.CurrentHand?.TrickWins || { 0: 0, 1: 0 };
  const t1 = wins[0] ?? wins["0"] ?? 0;
  const t2 = wins[1] ?? wins["1"] ?? 0;
  els.trickScore.textContent = tr("score_line", t1, t2);
  els.trickHistory.innerHTML = renderTrickHistory(s.CurrentHand?.TrickResults || []);

  const logs = s.Logs || [];
  els.logBox.textContent = logs.slice(-22).join("\n");
}

function canPlayCard(s) {
  return !s.MatchFinished && s.CurrentHand?.Turn === localPlayerID(s) && s.PendingRaiseFor === -1;
}

function playCardAtIndex(idx) {
  const payload = callApi("play", idx);
  if (applyPayload(payload)) {
    handFocusIndex = Math.max(0, idx - 1);
  }
}

function setHandFocus(index) {
  const buttons = [...els.hand.querySelectorAll(".card-btn:not(:disabled)")];
  if (buttons.length === 0) return;
  handFocusIndex = Math.max(0, Math.min(index, buttons.length - 1));
  buttons.forEach((btn, idx) => {
    btn.tabIndex = idx === handFocusIndex ? 0 : -1;
  });
  buttons[handFocusIndex].focus();
}

function renderHand(s) {
  els.hand.innerHTML = "";
  const me = playerByID(s, localPlayerID(s));
  const hand = me?.Hand || [];
  const allowPlay = canPlayCard(s);
  if (handFocusIndex >= hand.length) {
    handFocusIndex = Math.max(0, hand.length - 1);
  }

  hand.forEach((card, idx) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "card-btn";
    btn.setAttribute("role", "listitem");
    btn.setAttribute("aria-label", tr("card_aria_label", idx + 1, card?.Rank || "?", suitSymbol(card?.Suit)));
    btn.dataset.index = `${idx}`;
    btn.tabIndex = idx === handFocusIndex ? 0 : -1;
    const cardEl = createFaceCard(card, { hand: true, keyHint: `${idx + 1}` });
    if (!allowPlay) {
      cardEl.style.opacity = "0.56";
      btn.disabled = true;
    } else {
      btn.addEventListener("click", () => playCardAtIndex(idx));
      btn.addEventListener("focus", () => {
        handFocusIndex = idx;
      });
      btn.addEventListener("keydown", (ev) => {
        if (ev.key === "ArrowRight") {
          ev.preventDefault();
          setHandFocus(idx + 1);
          return;
        }
        if (ev.key === "ArrowLeft") {
          ev.preventDefault();
          setHandFocus(idx - 1);
          return;
        }
        if (ev.key === " " || ev.key === "Enter") {
          ev.preventDefault();
          playCardAtIndex(idx);
        }
      });
    }
    btn.appendChild(cardEl);
    els.hand.appendChild(btn);
  });
}

function updateControlsAndStatus(s) {
  const pendingTeam = s.PendingRaiseFor;
  const myTeam = localTeamID(s);
  const pendingForMe = pendingTeam !== -1 && pendingTeam === myTeam;
  const hasPending = pendingTeam !== -1;
  const canTruco = !s.MatchFinished && ((s.CurrentHand?.Turn === localPlayerID(s) && !hasPending) || pendingForMe);

  const pendingTo = s.PendingRaiseTo || nextStake(s.CurrentHand?.Stake || 1);

  els.btnTruco.disabled = !canTruco;
  els.btnAccept.disabled = !pendingForMe || s.MatchFinished;
  els.btnRefuse.disabled = !pendingForMe || s.MatchFinished;
  els.btnNewHand.disabled = s.MatchFinished;
  const showOnlineGovernance = !!onlineSession;
  els.btnHostVote.classList.toggle("hidden", !showOnlineGovernance);
  els.btnReplaceInvite.classList.toggle("hidden", !showOnlineGovernance);
  els.btnHostVote.disabled = !showOnlineGovernance || s.MatchFinished;
  els.btnReplaceInvite.disabled = !showOnlineGovernance || s.MatchFinished;
  els.btnTruco.classList.toggle("armed", canTruco && !pendingForMe && !s.MatchFinished);

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

  if (s.CurrentHand?.Turn === localPlayerID(s)) {
    setStatus(tr("status_your_turn"), false);
  } else {
    const turnName = playerByID(s, s.CurrentHand?.Turn ?? -1)?.Name || "?";
    setStatus(tr("status_wait_cpu", turnName), false);
  }
}

function showMatchOverlay(s) {
  const t1 = teamPoints(s, 0);
  const t2 = teamPoints(s, 1);
  const myTeam = localTeamID(s);
  const won = s.WinnerTeam === myTeam;

  els.matchOverlayTitle.textContent = won ? tr("overlay_match_title_win") : tr("overlay_match_title_loss");
  els.matchOverlayDetail.textContent = tr("overlay_match_detail", t1, t2);
  els.matchOverlay.classList.remove("hidden");
}

function showTrucoOverlay(pendingTo) {
  els.trucoOverlaySub.textContent = `${tr("overlay_truco_sub")} · ${raiseLabel(pendingTo)} (${pendingTo})`;
  els.trucoOverlay.classList.remove("hidden");
  triggerAnimationClass(els.trucoOverlay, "overlay-boom");
  triggerAnimationClass(els.tableShell, "truco-shake");
  setTimeout(() => {
    els.tableShell.classList.remove("truco-shake");
  }, 420);
  clearTimeout(trucoOverlayTimer);
  trucoOverlayTimer = setTimeout(() => {
    els.trucoOverlay.classList.add("hidden");
    els.trucoOverlay.classList.remove("overlay-boom");
  }, 1100);
}

function showTrickOverlay(state) {
  let msg = tr("overlay_trick_collecting");
  if (state.LastTrickTie) {
    msg = tr("overlay_trick_tie", state.LastTrickRound || 1);
  } else if (typeof state.LastTrickWinner === "number" && state.LastTrickWinner >= 0) {
    const winner = playerByID(state, state.LastTrickWinner)?.Name || "?";
    msg = tr("overlay_trick_collecting_by", winner);
  } else if (state.LastTrickTeam === localTeamID(state)) {
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

  if (teamPoints(prev, 0) !== teamPoints(next, 0)) {
    triggerAnimationClass(els.scoreT1, "score-pop");
  }
  if (teamPoints(prev, 1) !== teamPoints(next, 1)) {
    triggerAnimationClass(els.scoreT2, "score-pop");
  }
  if ((prev.CurrentHand?.Stake ?? 1) !== (next.CurrentHand?.Stake ?? 1)) {
    triggerAnimationClass(els.stakeValue, "score-pop");
    triggerAnimationClass(els.stakeLadder, "ladder-flash");
  }
  if ((prev.PendingRaiseTo || 0) !== (next.PendingRaiseTo || 0)) {
    triggerAnimationClass(els.stakeLadder, "ladder-flash");
  }

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

function updateOnlineSelectLabels() {
  if (els.modeSelect?.options?.length >= 2) {
    els.modeSelect.options[0].textContent = tr("setup_mode_offline");
    els.modeSelect.options[1].textContent = tr("setup_mode_online");
  }
  if (els.onlineAction?.options?.length >= 2) {
    els.onlineAction.options[0].textContent = tr("online_action_host");
    els.onlineAction.options[1].textContent = tr("online_action_join");
  }
  if (els.onlineRole?.options?.length >= 3) {
    els.onlineRole.options[0].textContent = tr("online_role_auto");
    els.onlineRole.options[1].textContent = tr("online_role_partner");
    els.onlineRole.options[2].textContent = tr("online_role_opponent");
  }
}

function updateOnlineActionUI() {
  const showKey = (els.onlineAction.value || "host") === "join";
  const wrapper = els.onlineKey?.parentElement?.parentElement;
  if (wrapper) {
    wrapper.classList.toggle("hidden", !showKey);
  }
}

function updateLobbyLabels() {
  if (!onlineSession) {
    els.btnLobbyStartText.textContent = tr("lobby_start");
    return;
  }
  const modeLabel = onlineSession.mode === "host" ? tr("lobby_mode_host") : tr("lobby_mode_client");
  els.lobbyMode.textContent = modeLabel;
  els.lobbyKey.textContent = onlineSession.inviteKey || "-";
  els.btnLobbyStartText.textContent = onlineSession.mode === "host" ? tr("lobby_start") : tr("lobby_enter");
}

function startOfflineFlow() {
  const payload = callApi("startGame", {
    numPlayers: Number(els.numPlayers.value || "2"),
    name: (els.playerName.value || "").trim() || (locale === "en-US" ? "You" : "Você"),
  });
  if (!applyPayload(payload)) return;
  onlineSession = null;
  stopOnlinePolling();
  els.setupPanel.classList.add("hidden");
  els.lobbyPanel.classList.add("hidden");
  els.gamePanel.classList.remove("hidden");
  startCpuLoop();
}

function startOnlineFlow() {
  const action = els.onlineAction.value || "host";
  const method = action === "join" ? "joinOnline" : "startOnlineHost";
  const payload = callApi(method, {
    numPlayers: Number(els.numPlayers.value || "2"),
    name: (els.playerName.value || "").trim() || (locale === "en-US" ? "You" : "Você"),
    key: (els.onlineKey.value || "").trim(),
    role: els.onlineRole.value || "auto",
  });
  if (!payload?.ok) {
    const msg = `${tr("error_prefix")}${payload?.error || tr("online_api_missing", method)}`;
    setStatus(msg, true);
    pushError(msg);
    return;
  }
  onlineSession = payload?.session || null;
  if (!onlineSession) {
    const msg = `${tr("error_prefix")}${tr("online_api_missing", method)}`;
    setStatus(msg, true);
    pushError(msg);
    return;
  }
  onlineSession.mode = action === "join" ? "client" : "host";
  els.setupPanel.classList.add("hidden");
  els.gamePanel.classList.add("hidden");
  els.lobbyPanel.classList.remove("hidden");
  setStatus(action === "join" ? tr("online_mode_notice_join") : tr("online_mode_notice_host"), false);
  refreshOnlineState(true);
  refreshOnlineEvents(true);
  startOnlinePolling();
}

function startOnlineMatchFlow() {
  const payload = callApi("startOnlineMatch");
  if (!payload?.ok) {
    const msg = `${tr("error_prefix")}${payload?.error || "?"}`;
    setStatus(msg, true);
    pushError(msg);
    return;
  }
  if (!applyPayload(payload)) return;
  stopOnlinePolling();
  els.lobbyPanel.classList.add("hidden");
  els.gamePanel.classList.remove("hidden");
  startCpuLoop();
  setStatus(`${tr("online_alpha_notice")} ${tr("status_ready")}`, false);
}

function refreshOnlineState(showErrors = false) {
  const payload = callApi("onlineState");
  if (!payload?.ok) {
    if (showErrors) pushError(`${tr("error_prefix")}${payload?.error || "?"}`);
    return;
  }
  if (Object.prototype.hasOwnProperty.call(payload, "session")) {
    onlineSession = payload.session;
  }
  renderLobby();
}

function refreshOnlineEvents(showErrors = false) {
  const payload = callApi("pullOnlineEvents");
  if (!payload?.ok) {
    if (showErrors) pushError(`${tr("error_prefix")}${payload?.error || "?"}`);
    return;
  }
  const events = payload?.events || [];
  for (const ev of events) {
    const stamp = new Date(ev.timestamp || Date.now());
    const hh = `${String(stamp.getHours()).padStart(2, "0")}:${String(stamp.getMinutes()).padStart(2, "0")}`;
    const line = `[${hh}] ${ev.type || "event"} · ${ev.message || ""}`;
    onlineEventTrail.unshift(line);
  }
  onlineEventTrail = onlineEventTrail.slice(0, 24);
  renderLobby();
}

function renderLobby() {
  updateLobbyLabels();
  const session = onlineSession;
  if (!session) {
    els.lobbySlots.innerHTML = "";
    els.lobbyEvents.textContent = tr("lobby_events_empty");
    return;
  }

  const slots = session.slots || [];
  const html = slots.map((name, idx) => {
    const safeName = escapeHtml((name || "").trim() || tr("lobby_slot_empty"));
    const roleList = [];
    if (idx === (session.assignedSeat ?? 0)) roleList.push(tr("role_you"));
    if (idx === (session.hostSeat ?? 0)) roleList.push(tr("role_host"));
    return `
      <div class="lobby-slot">
        <div class="top">
          <strong>${escapeHtml(tr("seat_slot", idx + 1))}</strong>
          <span>${safeName}</span>
        </div>
        <div class="roles">
          ${roleList.map((r) => `<span>${escapeHtml(r)}</span>`).join("")}
        </div>
      </div>
    `;
  }).join("");
  els.lobbySlots.innerHTML = html;
  els.lobbyEvents.textContent = onlineEventTrail.length ? onlineEventTrail.join("\n") : tr("lobby_events_empty");
}

function startOnlinePolling() {
  stopOnlinePolling();
  onlinePollTimer = setInterval(() => {
    refreshOnlineState(false);
    refreshOnlineEvents(false);
  }, 900);
}

function stopOnlinePolling() {
  if (!onlinePollTimer) return;
  clearInterval(onlinePollTimer);
  onlinePollTimer = null;
}

function bindUI() {
  els.localeSelect.addEventListener("change", () => {
    setLocale(els.localeSelect.value);
  });

  els.modeSelect.addEventListener("change", () => {
    mode = els.modeSelect.value === "online" ? "online" : "offline";
    els.onlineSetup.classList.toggle("hidden", mode !== "online");
    updateOnlineActionUI();
    renderStaticLabels();
  });
  els.onlineAction.addEventListener("change", () => {
    updateOnlineActionUI();
  });

  els.btnStart.addEventListener("click", () => {
    if (mode === "online") {
      startOnlineFlow();
      return;
    }
    startOfflineFlow();
  });

  els.btnLobbyStart.addEventListener("click", () => {
    startOnlineMatchFlow();
  });
  els.btnLobbyRefresh.addEventListener("click", () => {
    refreshOnlineState(true);
  });
  els.btnLobbyLeave.addEventListener("click", () => {
    callApi("leaveSession");
    stopOnlinePolling();
    onlineSession = null;
    onlineEventTrail = [];
    els.lobbyPanel.classList.add("hidden");
    els.setupPanel.classList.remove("hidden");
    setStatus(tr("status_ready"));
  });
  els.btnLobbyChat.addEventListener("click", () => {
    const msg = (els.lobbyChatInput.value || "").trim();
    if (!msg) return;
    const payload = callApi("sendChat", msg);
    if (!payload?.ok) {
      pushError(`${tr("error_prefix")}${payload?.error || "?"}`);
      return;
    }
    els.lobbyChatInput.value = "";
    refreshOnlineEvents(true);
  });

  els.btnTruco.addEventListener("click", () => doAction("truco"));
  els.btnAccept.addEventListener("click", () => doAction("accept"));
  els.btnRefuse.addEventListener("click", () => doAction("refuse"));
  els.btnNewHand.addEventListener("click", () => doAction("newHand"));
  els.btnHostVote.addEventListener("click", () => {
    if (!onlineSession) return;
    const raw = window.prompt(tr("prompt_slot_vote", onlineSession.numPlayers || 2), "1");
    const slot = Number(raw || "0");
    if (!Number.isFinite(slot) || slot < 1) return;
    const payload = callApi("sendHostVote", slot);
    if (!payload?.ok) {
      pushError(`${tr("error_prefix")}${payload?.error || "?"}`);
      return;
    }
    refreshOnlineEvents(true);
  });
  els.btnReplaceInvite.addEventListener("click", () => {
    if (!onlineSession) return;
    const raw = window.prompt(tr("prompt_slot_replace", onlineSession.numPlayers || 2), "2");
    const slot = Number(raw || "0");
    if (!Number.isFinite(slot) || slot < 1) return;
    const payload = callApi("requestReplacementInvite", slot);
    if (!payload?.ok) {
      pushError(`${tr("error_prefix")}${payload?.error || "?"}`);
      return;
    }
    if (payload?.inviteKey) {
      pushError(tr("online_invite_generated", slot, payload.inviteKey));
    }
    refreshOnlineEvents(true);
  });

  els.btnPlayAgain.addEventListener("click", () => {
    callApi("reset");
    snap = null;
    prevSnap = null;
    onlineSession = null;
    onlineEventTrail = [];
    stopOnlinePolling();
    els.matchOverlay.classList.add("hidden");
    els.trickOverlay.classList.add("hidden");
    els.trucoOverlay.classList.add("hidden");
    els.trucoOverlay.classList.remove("overlay-boom");
    els.tableShell.classList.remove("truco-shake");
    els.gamePanel.classList.add("hidden");
    els.lobbyPanel.classList.add("hidden");
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
    const activeTag = document.activeElement?.tagName?.toLowerCase();
    if (activeTag === "input" || activeTag === "textarea" || activeTag === "select") return;
    const key = ev.key.toLowerCase();
    if (key === "1" || key === "2" || key === "3") {
      playCardAtIndex(Number(key) - 1);
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
  mode = els.modeSelect.value === "online" ? "online" : "offline";
  els.onlineSetup.classList.toggle("hidden", mode !== "online");
  els.lobbyEvents.textContent = tr("lobby_events_empty");
  setLocale("pt-BR");
  renderStakeLadder(1, 0);
  setStatus(tr("loading_wasm"));
  try {
    await bootWasm();
    els.trucoOverlay.setAttribute("aria-live", "assertive");
    els.matchOverlay.setAttribute("aria-live", "assertive");
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
