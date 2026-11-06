import { layout, prepare } from "@chenglou/pretext";
import { translate } from "./i18n";
import type {
  ApiResult,
  Card,
  LocaleCode,
  LobbySlotState,
  MatchSnapshot,
  PlayedCard,
  RuntimeEvent,
  SnapshotBundle,
} from "./types";

type ViewName = "setup" | "lobby" | "game";

interface TrickRevealState {
  seq: number;
  cards: PlayedCard[];
  winner: number;
  team: number;
  tie: boolean;
  round: number;
  manilha: string;
  until: number;
  holdMatch: MatchSnapshot | null;
}

interface PendingHandBanner {
  title: string;
  sub: string;
  kind: string;
  durationMs: number;
  confettiMs: number;
}

interface UIFxState {
  animatedCardKeys: Set<string>;
  animatedViraKey: string;
  trucoCall: { id: number; text: string; sub: string; until: number } | null;
  lastRaiseSignal: string;
  announcement: { id: number; title: string; sub: string; kind: string; until: number } | null;
  confetti: { id: number; strength: number; until: number } | null;
  trickReveal: TrickRevealState | null;
  pendingHandBanner: PendingHandBanner | null;
}

interface AppState {
  sessionId: string;
  locale: LocaleCode;
  bundle: SnapshotBundle | null;
  events: RuntimeEvent[];
  playerName: string;
  relayURL: string;
  error: string;
  busyForm: string;
  initialized: boolean;
  viewSig: string;
  fxSig: string;
  lastRenderedView: ViewName | "";
  lastMatchSig: string;
  lastMatchChangeAt: number;
  drawerOpen: boolean;
  uiFx: UIFxState;
}

const SESSION_KEY = "truco-browser-session-id";
const LOCALE_KEY = "truco-browser-locale";
const PLAYER_KEY = "truco-browser-player-name";
const RELAY_KEY = "truco-browser-relay-url";
const EVENT_LIMIT = 80;

const NORMAL_RANK_POWER: Record<string, number> = {
  "4": 1,
  "5": 2,
  "6": 3,
  "7": 4,
  Q: 5,
  J: 6,
  K: 7,
  A: 8,
  "2": 9,
  "3": 10,
};

const MANILHA_SUIT_POWER: Record<string, number> = {
  Ouros: 1,
  Espadas: 2,
  Copas: 3,
  Paus: 4,
};

const rootElement = document.querySelector<HTMLElement>("#app");
if (!rootElement) {
  throw new Error("missing #app root");
}
const root: HTMLElement = rootElement;

root.innerHTML = `
  <div id="view-layer" class="page-shell"></div>
  <div id="fx-layer" class="fx-root"></div>
`;

const viewLayer = root.querySelector<HTMLElement>("#view-layer")!;
const fxLayer = root.querySelector<HTMLElement>("#fx-layer")!;

let fxSeq = 0;
let refreshTimer = 0;
let resizeTimer = 0;
let fxTimer = 0;
let shakeTimeout = 0;
let pulseTimeout = 0;

const state: AppState = {
  sessionId: sessionStorage.getItem(SESSION_KEY) || "",
  locale: readLocale(localStorage.getItem(LOCALE_KEY)),
  bundle: null,
  events: [],
  playerName: localStorage.getItem(PLAYER_KEY) || "",
  relayURL: localStorage.getItem(RELAY_KEY) || "",
  error: "",
  busyForm: "",
  initialized: false,
  viewSig: "",
  fxSig: "",
  lastRenderedView: "",
  lastMatchSig: "",
  lastMatchChangeAt: 0,
  drawerOpen: false,
  uiFx: {
    animatedCardKeys: new Set<string>(),
    animatedViraKey: "",
    trucoCall: null,
    lastRaiseSignal: "",
    announcement: null,
    confetti: null,
    trickReveal: null,
    pendingHandBanner: null,
  },
};

const preparedCache = new Map<string, ReturnType<typeof prepare>>();

root.addEventListener("submit", (event) => {
  const form = (event.target as HTMLElement | null)?.closest<HTMLFormElement>("form[data-api-action]");
  if (!form) {
    return;
  }
  event.preventDefault();
  void submitForm(form);
});

root.addEventListener("click", (event) => {
  const target = event.target as HTMLElement | null;
  if (!target) {
    return;
  }

  const copyButton = target.closest<HTMLElement>("[data-copy-text]");
  if (copyButton) {
    void copyInvite(copyButton);
    return;
  }

  const drawerToggle = target.closest<HTMLElement>("[data-toggle-drawer]");
  if (drawerToggle) {
    state.drawerOpen = !state.drawerOpen;
    render();
    return;
  }

  const dismissBanner = target.closest<HTMLElement>("[data-dismiss-error]");
  if (dismissBanner) {
    state.error = "";
    render();
  }
});

root.addEventListener("input", (event) => {
  const input = event.target as HTMLInputElement | null;
  if (!input) {
    return;
  }
  if (input.name === "name" || input.dataset.syncPlayerName === "true") {
    const val = input.value;
    state.playerName = val;
    if (val.trim() !== "") {
      localStorage.setItem(PLAYER_KEY, val.trim());
    }
    for (const other of root.querySelectorAll<HTMLInputElement>('input[name="name"], input[data-sync-player-name="true"]')) {
      if (other !== input && other.value !== val) {
        other.value = val;
      }
    }
  }
});

root.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && state.drawerOpen) {
    state.drawerOpen = false;
    render();
    return;
  }
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  const target = event.target as HTMLElement | null;
  const copyTarget = target?.closest<HTMLElement>("[data-copy-text][role='button']");
  if (copyTarget) {
    event.preventDefault();
    void copyInvite(copyTarget);
  }
});

root.addEventListener("change", (event) => {
  const select = event.target as HTMLSelectElement | null;
  if (!select || select.name !== "locale") {
    return;
  }
  void applyLocale(readLocale(select.value));
});

window.addEventListener("resize", () => {
  window.clearTimeout(resizeTimer);
  resizeTimer = window.setTimeout(() => syncMeasuredBlocks(), 90);
});

document.addEventListener("visibilitychange", () => {
  if (document.hidden || !state.bundle) {
    return;
  }
  void (async () => {
    try {
      await syncSnapshot({ withEvents: true });
      state.error = "";
    } catch (error) {
      state.error = errorMessage(error);
    }
    render();
  })();
});

void bootstrap();

async function bootstrap(): Promise<void> {
  render();
  try {
    await ensureSession();
    await syncSnapshot({ withEvents: true });
    state.initialized = true;
    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  }
  render();
}

async function ensureSession(force = false): Promise<void> {
  if (!force && state.sessionId) {
    return;
  }

  const result = await sendRequest("createSession", undefined, { withSession: false, retryOnSessionMiss: false });
  if (!result.ok || !result.sessionId) {
    throw new Error(result.error || "failed to create browser session");
  }

  state.sessionId = result.sessionId;
  sessionStorage.setItem(SESSION_KEY, state.sessionId);

  const localeResult = await sendRequest("setLocale", { locale: state.locale }, { retryOnSessionMiss: false });
  updateStateFromResult(localeResult);
}

async function applyLocale(locale: LocaleCode): Promise<void> {
  state.locale = locale;
  localStorage.setItem(LOCALE_KEY, locale);
  document.documentElement.lang = locale === "en-US" ? "en" : "pt-BR";

  try {
    if (!state.sessionId) {
      await ensureSession();
    }
    const result = await sendRequest("setLocale", { locale });
    updateStateFromResult(result);
    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  }

  render();
}

async function submitForm(form: HTMLFormElement): Promise<void> {
  const action = form.dataset.apiAction;
  if (!action || state.busyForm) {
    return;
  }

  // If the user explicitly interacts during a visual trick hold, flush it immediately
  flushPendingVisualHold();

  const viewBeforeSubmit = currentView();
  let viewChanged = false;
  const payload = formPayload(form);
  const formId = form.dataset.formId || action;
  state.busyForm = formId;
  render();

  try {
    persistInputs(payload);
    await ensureSession();
    const result = await sendRequest(action, payload);

    if (!result.ok) {
      throw new Error(result.error || action);
    }

    if (action === "closeSession") {
      clearSession();
      await ensureSession(true);
      state.events = [];
      await syncSnapshot({ withEvents: true });
    } else {
      updateStateFromResult(result);
      await syncAfterAction(action);
    }

    viewChanged = currentView() !== viewBeforeSubmit;
    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  } finally {
    state.busyForm = "";
    render();
    if (viewChanged) {
      window.scrollTo({ top: 0, left: 0 });
    }
  }
}

async function syncAfterAction(action: string): Promise<void> {
  // Most actions already return the updated SnapshotBundle directly in their POST response.
  // Only poll events separately when in online mode for actions that don't drain events.
  if (action === "pollEvents" || action === "setLocale" || action === "reset") {
    return;
  }
  if (isOnlineMode() && (action === "play" || action === "truco" || action === "accept" || action === "refuse")) {
    await syncSnapshot({ withEvents: true });
  }
}

async function syncSnapshot(options: { withEvents: boolean }): Promise<void> {
  await ensureSession();
  const action = options.withEvents ? "pollEvents" : "snapshot";
  const result = await sendRequest(action, undefined);
  if (!result.ok) {
    throw new Error(result.error || "snapshot failed");
  }
  updateStateFromResult(result);
}

async function sendRequest(
  action: string,
  body?: Record<string, unknown>,
  options?: { withSession?: boolean; retryOnSessionMiss?: boolean },
): Promise<ApiResult> {
  const withSession = options?.withSession ?? true;
  const retryOnSessionMiss = options?.retryOnSessionMiss ?? true;
  const headers: HeadersInit = { "Content-Type": "application/json" };
  if (withSession && state.sessionId) {
    headers["X-Session-ID"] = state.sessionId;
  }

  let response: Response;
  try {
    response = await fetch(`/api/${action}`, {
      method: "POST",
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: "same-origin",
    });
  } catch {
    throw new Error(t("error_fetch"));
  }

  const result = normalizeApiResult(await response.json());
  if (response.ok) {
    return result;
  }

  if (
    retryOnSessionMiss &&
    result.error_code === "session_not_found" &&
    action !== "createSession"
  ) {
    clearSession();
    await ensureSession(true);
    return sendRequest(action, body, { withSession, retryOnSessionMiss: false });
  }

  return result;
}

function normalizeApiResult(raw: unknown): ApiResult {
  const source = isRecord(raw) ? raw : {};
  return {
    ok: source.ok === true,
    error: stringValue(source.error),
    error_code: stringValue(source.error_code),
    sessionId: stringValue(source.sessionId),
    bundle: source.bundle ? normalizeBundle(source.bundle) : undefined,
    mode: stringValue(source.mode),
    session: source.session ? normalizeSessionView(source.session) : undefined,
    events: normalizeEvents(source.events),
    snapshot: stringValue(source.snapshot),
    sessionClosed: source.sessionClosed === true,
  };
}

function normalizeBundle(raw: unknown): SnapshotBundle {
  const source = isRecord(raw) ? raw : {};
  return {
    versions: {
      core_api_version: numberValue(recordValue(source, "versions"), "core_api_version", 0),
      protocol_version: numberValue(recordValue(source, "versions"), "protocol_version", 0),
      snapshot_schema_version: numberValue(recordValue(source, "versions"), "snapshot_schema_version", 0),
    },
    mode: stringValue(source.mode) || "idle",
    locale: readLocale(stringValue(source.locale) || null),
    match: source.match ? normalizeMatch(source.match) : undefined,
    lobby: source.lobby ? normalizeLobby(source.lobby) : undefined,
    ui: normalizeUIState(source.ui),
    connection: normalizeConnection(source.connection),
    diagnostics: normalizeDiagnostics(source.diagnostics),
  };
}

function normalizeMatch(raw: unknown): MatchSnapshot {
  const source = isRecord(raw) ? raw : {};
  return {
    Players: normalizePlayers(source.Players),
    NumPlayers: numberValue(source, "NumPlayers", 0),
    CurrentHand: normalizeHandState(source.CurrentHand),
    LastTrickCards: normalizePlayedCards(source.LastTrickCards),
    TrickPiles: normalizeTrickPiles(source.TrickPiles),
    MatchPoints: normalizeNumberMap(source.MatchPoints),
    TurnPlayer: numberValue(source, "TurnPlayer", -1),
    CurrentTeamTurn: numberValue(source, "CurrentTeamTurn", -1),
    Logs: normalizeStringArray(source.Logs),
    WinnerTeam: numberValue(source, "WinnerTeam", -1),
    MatchFinished: source.MatchFinished === true,
    CanAskTruco: source.CanAskTruco === true,
    PendingRaiseFor: numberValue(source, "PendingRaiseFor", -1),
    PendingRaiseBy: numberValue(source, "PendingRaiseBy", -1),
    PendingRaiseTo: numberValue(source, "PendingRaiseTo", 0),
    CurrentPlayerIdx: numberValue(source, "CurrentPlayerIdx", -1),
    LastTrickSeq: numberValue(source, "LastTrickSeq", 0),
    LastTrickTeam: numberValue(source, "LastTrickTeam", -1),
    LastTrickWinner: numberValue(source, "LastTrickWinner", -1),
    LastTrickTie: source.LastTrickTie === true,
    LastTrickRound: numberValue(source, "LastTrickRound", 0),
  };
}

function normalizePlayers(raw: unknown): MatchSnapshot["Players"] {
  return normalizeRecordArray(raw).map((player) => ({
    ID: numberValue(player, "ID", -1),
    Name: stringValue(player.Name) || "",
    CPU: player.CPU === true,
    ProvisionalCPU: player.ProvisionalCPU === true,
    Team: numberValue(player, "Team", 0),
    Hand: normalizeCards(player.Hand),
  }));
}

function normalizeCards(raw: unknown): Card[] {
  return normalizeRecordArray(raw).map((card) => ({
    Rank: stringValue(card.Rank) || "",
    Suit: stringValue(card.Suit) || "",
  }));
}

function normalizePlayedCards(raw: unknown): MatchSnapshot["CurrentHand"]["RoundCards"] {
  return normalizeRecordArray(raw).map((played) => ({
    PlayerID: numberValue(played, "PlayerID", -1),
    Card: normalizeCards([played.Card])[0] || { Rank: "", Suit: "" },
    FaceDown: played.FaceDown === true,
  }));
}

function normalizeTrickPiles(raw: unknown): MatchSnapshot["TrickPiles"] {
  return normalizeRecordArray(raw).map((pile) => ({
    Winner: numberValue(pile, "Winner", -1),
    Team: numberValue(pile, "Team", -1),
    Round: numberValue(pile, "Round", 0),
    Cards: normalizePlayedCards(pile.Cards),
  }));
}

function normalizeHandState(raw: unknown): MatchSnapshot["CurrentHand"] {
  const source = isRecord(raw) ? raw : {};
  return {
    Vira: normalizeCards([source.Vira])[0] || { Rank: "", Suit: "" },
    Manilha: stringValue(source.Manilha) || "",
    Stake: numberValue(source, "Stake", 1),
    TrucoByTeam: numberValue(source, "TrucoByTeam", -1),
    RaiseRequester: numberValue(source, "RaiseRequester", -1),
    Dealer: numberValue(source, "Dealer", -1),
    Turn: numberValue(source, "Turn", 0),
    Round: numberValue(source, "Round", 1),
    RoundStart: numberValue(source, "RoundStart", 0),
    RoundCards: normalizePlayedCards(source.RoundCards),
    TrickResults: normalizeNumberArray(source.TrickResults),
    TrickWins: normalizeNumberMap(source.TrickWins),
    WinnerTeam: numberValue(source, "WinnerTeam", -1),
    Finished: source.Finished === true,
    PendingRaiseFor: numberValue(source, "PendingRaiseFor", -1),
  };
}

function normalizeLobby(raw: unknown): NonNullable<SnapshotBundle["lobby"]> {
  const source = isRecord(raw) ? raw : {};
  return {
    invite_key: stringValue(source.invite_key),
    slots: normalizeStringArray(source.slots),
    assigned_seat: numberValue(source, "assigned_seat", -1),
    num_players: numberValue(source, "num_players", 0),
    started: source.started === true,
    host_seat: numberValue(source, "host_seat", -1),
    connected_seats: normalizeBooleanMap(source.connected_seats),
    role: stringValue(source.role),
  };
}

function normalizeUIState(raw: unknown): SnapshotBundle["ui"] {
  const source = isRecord(raw) ? raw : {};
  const actions = recordValue(source, "actions");
  return {
    lobby_slots: normalizeRecordArray(source.lobby_slots).map((slot) => ({
      seat: numberValue(slot, "seat", 0),
      name: stringValue(slot.name),
      status: stringValue(slot.status) || "",
      is_empty: slot.is_empty === true,
      is_local: slot.is_local === true,
      is_host: slot.is_host === true,
      is_connected: slot.is_connected === true,
      is_occupied: slot.is_occupied === true,
      is_provisional_cpu: slot.is_provisional_cpu === true,
      can_vote_host: slot.can_vote_host === true,
      can_request_replacement: slot.can_request_replacement === true,
    })),
    actions: {
      local_player_id: numberValue(actions, "local_player_id", -1),
      local_team: numberValue(actions, "local_team", -1),
      can_play_card: boolValue(actions, "can_play_card", false),
      can_ask_or_raise: boolValue(actions, "can_ask_or_raise", false),
      must_respond: boolValue(actions, "must_respond", false),
      can_accept: boolValue(actions, "can_accept", false),
      can_refuse: boolValue(actions, "can_refuse", false),
      can_close_session: boolValue(actions, "can_close_session", false),
    },
  };
}

function normalizeConnection(raw: unknown): SnapshotBundle["connection"] {
  const source = isRecord(raw) ? raw : {};
  const networkSource = recordValue(source, "network");
  const lastErrorSource = recordValue(source, "last_error");
  return {
    status: stringValue(source.status) || "",
    is_online: source.is_online === true,
    is_host: source.is_host === true,
    network: networkSource
      ? {
          transport: stringValue(networkSource.transport),
          supported_protocol_versions: normalizeNumberArray(networkSource.supported_protocol_versions),
          negotiated_protocol_version: numberValue(networkSource, "negotiated_protocol_version", 0) || undefined,
          seat_protocol_versions: normalizeNumberMap(networkSource.seat_protocol_versions),
          mixed_protocol_session: networkSource.mixed_protocol_session === true,
        }
      : undefined,
    last_error: lastErrorSource
      ? {
          code: stringValue(lastErrorSource.code) || "",
          message: stringValue(lastErrorSource.message) || "",
        }
      : undefined,
    last_event_sequence: numberValue(source, "last_event_sequence", 0),
  };
}

function normalizeDiagnostics(raw: unknown): SnapshotBundle["diagnostics"] {
  const source = isRecord(raw) ? raw : {};
  return {
    event_backlog: numberValue(source, "event_backlog", 0),
    replay_seed_lo: numberValue(source, "replay_seed_lo", 0) || undefined,
    replay_seed_hi: numberValue(source, "replay_seed_hi", 0) || undefined,
    event_log: normalizeStringArray(source.event_log),
  };
}

function normalizeSessionView(raw: unknown): NonNullable<ApiResult["session"]> {
  const source = isRecord(raw) ? raw : {};
  const networkSource = recordValue(source, "network");
  return {
    mode: stringValue(source.mode),
    inviteKey: stringValue(source.inviteKey),
    numPlayers: numberValue(source, "numPlayers", 0) || undefined,
    assignedSeat: numberValue(source, "assignedSeat", 0) || undefined,
    hostSeat: numberValue(source, "hostSeat", 0) || undefined,
    slots: normalizeStringArray(source.slots),
    connected: normalizeBooleanArray(source.connected),
    started: source.started === true,
    role: stringValue(source.role),
    network: networkSource
      ? {
          transport: stringValue(networkSource.transport),
          supported_protocol_versions: normalizeNumberArray(networkSource.supported_protocol_versions),
          negotiated_protocol_version: numberValue(networkSource, "negotiated_protocol_version", 0) || undefined,
          seat_protocol_versions: normalizeNumberMap(networkSource.seat_protocol_versions),
          mixed_protocol_session: networkSource.mixed_protocol_session === true,
        }
      : undefined,
  };
}

function normalizeEvents(raw: unknown): RuntimeEvent[] {
  return normalizeRecordArray(raw).map((event) => ({
    kind: stringValue(event.kind) || "system",
    sequence: numberValue(event, "sequence", 0),
    timestamp: stringValue(event.timestamp) || "",
    payload: recordValue(event, "payload") || {},
  }));
}

function normalizeRecordArray(raw: unknown): Array<Record<string, unknown>> {
  return Array.isArray(raw) ? raw.filter(isRecord) : [];
}

function normalizeStringArray(raw: unknown): string[] {
  return Array.isArray(raw) ? raw.filter((value): value is string => typeof value === "string") : [];
}

function normalizeNumberArray(raw: unknown): number[] {
  if (!Array.isArray(raw)) {
    return [];
  }
  return raw
    .map((value) => (typeof value === "number" && Number.isFinite(value) ? value : Number.NaN))
    .filter((value) => Number.isFinite(value));
}

function normalizeBooleanArray(raw: unknown): boolean[] {
  return Array.isArray(raw) ? raw.map((value) => value === true) : [];
}

function normalizeNumberMap(raw: unknown): Record<string, number> {
  const source = isRecord(raw) ? raw : {};
  const out: Record<string, number> = {};
  for (const [key, value] of Object.entries(source)) {
    if (typeof value === "number" && Number.isFinite(value)) {
      out[key] = value;
    }
  }
  return out;
}

function normalizeBooleanMap(raw: unknown): Record<string, boolean> {
  const source = isRecord(raw) ? raw : {};
  const out: Record<string, boolean> = {};
  for (const [key, value] of Object.entries(source)) {
    out[key] = value === true;
  }
  return out;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function recordValue(source: unknown, key: string): Record<string, unknown> | null {
  if (!isRecord(source) || !isRecord(source[key])) {
    return null;
  }
  return source[key];
}

function stringValue(value: unknown): string | undefined {
  return typeof value === "string" ? value : undefined;
}

function numberValue(source: unknown, key: string, fallback: number): number {
  return isRecord(source) && typeof source[key] === "number" && Number.isFinite(source[key]) ? source[key] : fallback;
}

function boolValue(source: unknown, key: string, fallback: boolean): boolean {
  return isRecord(source) && typeof source[key] === "boolean" ? source[key] : fallback;
}

function updateStateFromResult(result: ApiResult): void {
  const previousMatch = state.bundle?.match;

  if (result.bundle) {
    state.bundle = result.bundle;
    state.locale = result.bundle.locale || state.locale;
    document.documentElement.lang = state.locale === "en-US" ? "en" : "pt-BR";
    localStorage.setItem(LOCALE_KEY, state.locale);
  }

  if (result.events) {
    mergeEvents(result.events);
  }

  if (result.error) {
    state.error = result.error;
  }

  trackMatchChange(state.bundle?.match);
  updateVisualFx(previousMatch, state.bundle?.match, result.events || []);
}

function mergeEvents(incoming: RuntimeEvent[]): void {
  if (incoming.length === 0) {
    return;
  }
  const seen = new Set(state.events.map((event) => event.sequence));
  const fresh = incoming.filter((event) => !seen.has(event.sequence));
  if (fresh.length === 0) {
    return;
  }
  state.events = [...state.events, ...fresh]
    .sort((a, b) => a.sequence - b.sequence)
    .slice(-EVENT_LIMIT);
}

function matchSignature(match: MatchSnapshot | undefined): string {
  if (!match) {
    return "";
  }
  return JSON.stringify([
    match.TurnPlayer,
    match.CurrentHand.Round,
    match.CurrentHand.RoundCards.length,
    match.CurrentHand.TrickResults,
    match.CurrentHand.Stake,
    match.PendingRaiseFor,
    match.MatchPoints,
    match.MatchFinished,
    match.LastTrickSeq,
    match.CurrentHand.Vira.Rank,
    match.CurrentHand.Vira.Suit,
    match.Logs.length,
  ]);
}

function trackMatchChange(nextMatch: MatchSnapshot | undefined): void {
  const sig = matchSignature(nextMatch);
  if (sig !== state.lastMatchSig) {
    state.lastMatchSig = sig;
    state.lastMatchChangeAt = Date.now();
  }
}

function clearSession(): void {
  state.sessionId = "";
  sessionStorage.removeItem(SESSION_KEY);
  state.bundle = null;
  state.events = [];
  state.lastMatchSig = "";
  state.lastMatchChangeAt = 0;
  state.drawerOpen = false;
  state.uiFx.trickReveal = null;
  state.uiFx.pendingHandBanner = null;
  state.uiFx.announcement = null;
  state.uiFx.trucoCall = null;
  state.uiFx.confetti = null;
  state.uiFx.animatedCardKeys.clear();
}

function flushPendingVisualHold(): void {
  if (state.uiFx.trickReveal) {
    state.uiFx.trickReveal = null;
  }
  if (state.uiFx.pendingHandBanner) {
    const banner = state.uiFx.pendingHandBanner;
    state.uiFx.pendingHandBanner = null;
    showAnnouncement(banner.title, banner.sub, banner.kind, banner.durationMs);
    if (banner.confettiMs > 0) {
      triggerConfetti(banner.confettiMs, 1);
    }
  }
}

function scheduleFxRefresh(): void {
  window.clearTimeout(fxTimer);
  const now = Date.now();

  if (state.uiFx.trickReveal && state.uiFx.trickReveal.until <= now) {
    state.uiFx.trickReveal = null;
    if (state.uiFx.pendingHandBanner) {
      const banner = state.uiFx.pendingHandBanner;
      state.uiFx.pendingHandBanner = null;
      showAnnouncement(banner.title, banner.sub, banner.kind, banner.durationMs);
      if (banner.confettiMs > 0) {
        triggerConfetti(banner.confettiMs, 1);
      }
    }
  }

  if (state.uiFx.trucoCall && state.uiFx.trucoCall.until <= now) {
    state.uiFx.trucoCall = null;
  }
  if (state.uiFx.announcement && state.uiFx.announcement.until <= now) {
    state.uiFx.announcement = null;
  }
  if (state.uiFx.confetti && state.uiFx.confetti.until <= now) {
    state.uiFx.confetti = null;
  }

  const deadlines = [
    state.uiFx.trickReveal?.until || 0,
    state.uiFx.trucoCall?.until || 0,
    state.uiFx.announcement?.until || 0,
    state.uiFx.confetti?.until || 0,
  ].filter((time) => time > now);

  if (deadlines.length === 0) {
    return;
  }
  const nextExpiry = Math.min(...deadlines);
  fxTimer = window.setTimeout(() => render(), Math.max(24, nextExpiry - now + 20));
}

function currentView(): ViewName {
  const mode = state.bundle?.mode || "idle";
  if (mode.includes("lobby")) {
    return "lobby";
  }
  if (mode.includes("match")) {
    return "game";
  }
  return "setup";
}

function isOnlineMode(): boolean {
  const mode = state.bundle?.mode || "";
  return mode.startsWith("host_") || mode.startsWith("client_");
}

function activeDisplayMatch(): MatchSnapshot | undefined {
  const liveMatch = state.bundle?.match;
  if (!liveMatch) {
    return undefined;
  }
  const now = Date.now();
  const reveal = state.uiFx.trickReveal;
  if (reveal && reveal.until > now && reveal.holdMatch) {
    return reveal.holdMatch;
  }
  return liveMatch;
}

function animateScoreChange(container: HTMLElement, team: string, prevScoreStr: string | undefined | null): void {
  const el = container.querySelector<HTMLElement>(`[data-team="${team}"] [data-score]`);
  if (!el) return;
  const current = el.getAttribute("data-score");
  if (!prevScoreStr || !current || prevScoreStr === current) return;
  el.classList.add("score-updating");
  el.addEventListener("animationend", () => el.classList.remove("score-updating"), { once: true });
}

function render(): void {
  scheduleFxRefresh();

  const view = currentView();
  const nextViewSig = computeViewSignature();
  const nextFxSig = computeFxSignature();

  if (nextViewSig !== state.viewSig) {
    const focusState = captureFocusState();
    const detailsState = captureDetailsState();
    const prevUsScore = viewLayer.querySelector('[data-team="us"] [data-score]')?.getAttribute("data-score");
    const prevThemScore = viewLayer.querySelector('[data-team="them"] [data-score]')?.getAttribute("data-score");

    viewLayer.className = `page-shell page-shell-${view}`;
    viewLayer.innerHTML = renderAppContent();

    animateScoreChange(viewLayer, "us", prevUsScore);
    animateScoreChange(viewLayer, "them", prevThemScore);
    restoreFocusState(focusState);
    restoreDetailsState(detailsState);
    state.viewSig = nextViewSig;
    state.lastRenderedView = view;
    syncMeasuredBlocks();
    scrollEventFeeds();
  }

  if (nextFxSig !== state.fxSig) {
    fxLayer.innerHTML = renderFxLayers();
    state.fxSig = nextFxSig;
  }

  syncRefreshLoop();
}

function captureDetailsState(): Record<string, boolean> {
  const open: Record<string, boolean> = {};
  for (const details of viewLayer.querySelectorAll<HTMLDetailsElement>("details[data-details-key]")) {
    open[details.dataset.detailsKey || ""] = details.open;
  }
  return open;
}

function restoreDetailsState(open: Record<string, boolean>): void {
  for (const details of viewLayer.querySelectorAll<HTMLDetailsElement>("details[data-details-key]")) {
    const key = details.dataset.detailsKey || "";
    if (key in open) {
      details.open = open[key];
    }
  }
}

function scrollEventFeeds(): void {
  for (const feed of viewLayer.querySelectorAll<HTMLElement>(".event-feed")) {
    feed.scrollTop = feed.scrollHeight;
  }
}

interface FocusState {
  selector: string;
  value?: string;
  selectionStart: number | null;
  selectionEnd: number | null;
}

function captureFocusState(): FocusState | null {
  const active = document.activeElement;
  if (!(active instanceof HTMLInputElement || active instanceof HTMLTextAreaElement || active instanceof HTMLSelectElement)) {
    return null;
  }
  if (!viewLayer.contains(active)) {
    return null;
  }
  const key = active.id || active.name || active.getAttribute("data-focus-key");
  if (!key) {
    return null;
  }
  const formId = active.closest("form")?.getAttribute("data-form-id");
  const selector = formId && active.name
    ? `form[data-form-id="${CSS.escape(formId)}"] [name="${CSS.escape(active.name)}"]`
    : active.id
      ? `#${CSS.escape(active.id)}`
      : `[name="${CSS.escape(active.name || key)}"]`;
  return {
    selector,
    value: active.value,
    selectionStart: "selectionStart" in active ? active.selectionStart : null,
    selectionEnd: "selectionEnd" in active ? active.selectionEnd : null,
  };
}

function restoreFocusState(stateToRestore: FocusState | null): void {
  if (!stateToRestore) {
    return;
  }
  const next = viewLayer.querySelector<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(stateToRestore.selector);
  if (!next) {
    return;
  }
  if (stateToRestore.value !== undefined && next.value !== stateToRestore.value) {
    next.value = stateToRestore.value;
  }
  next.focus({ preventScroll: true });
  if (
    (next instanceof HTMLInputElement || next instanceof HTMLTextAreaElement) &&
    stateToRestore.selectionStart !== null &&
    stateToRestore.selectionEnd !== null
  ) {
    next.setSelectionRange(stateToRestore.selectionStart, stateToRestore.selectionEnd);
  }
}

function renderAppContent(): string {
  const locale = state.locale;
  const busy = state.busyForm !== "";
  const view = currentView();
  const banner = state.error
    ? `<section class="runtime-banner" role="alert" aria-live="assertive">
        <span>${escapeHtml(state.error)}</span>
        <button type="button" class="banner-dismiss" data-dismiss-error="true" aria-label="Fechar">×</button>
      </section>`
    : "";
  const chrome = view === "game" ? "" : `
        <header class="hero-card">
          <div class="hero-copy">
            <div class="hero-title-row">
              <h1>${escapeHtml(t("title_main"))}</h1>
              <span class="hero-stamp">${escapeHtml(t("app_stamp"))}</span>
            </div>
            <p class="hero-subtitle">${escapeHtml(t("title_sub"))}</p>
          </div>
          <form class="locale-card" data-api-action="setLocale" data-form-id="setLocale">
            <label for="locale-select">${escapeHtml(t("locale_label"))}</label>
            <select id="locale-select" name="locale" ${busy ? "disabled" : ""}>
              ${localeOptions(locale)}
            </select>
          </form>
        </header>`;

  return `
    <main class="app-shell app-shell-${view}">
      ${chrome}
      ${banner}
      ${renderView()}
    </main>
  `;
}

function renderView(): string {
  if (!state.initialized && !state.bundle && !state.error) {
    return `<section class="surface-card loading-card"><div class="loading-pip"></div><strong>${escapeHtml(t("button_busy"))}</strong></section>`;
  }

  switch (currentView()) {
    case "lobby":
      return renderLobby();
    case "game":
      return renderGame();
    default:
      return renderSetup();
  }
}

function renderSetup(): string {
  const defaultName = state.playerName || t("name_placeholder");
  return `
    <section class="setup-shell">
      <div class="setup-stack">
        <article class="surface-card form-card primary-form-card">
          <div class="card-head">
            <div>
              <span class="section-pill">${escapeHtml(t("setup_mode_offline"))}</span>
              <h2>${escapeHtml(t("setup_offline_title"))}</h2>
              <p>${escapeHtml(t("setup_offline_note"))}</p>
            </div>
            <strong class="form-emphasis">${escapeHtml(t("setup_offline_caption"))}</strong>
          </div>
          <form class="setup-main-form" data-api-action="startGame" data-form-id="startGame">
            <div class="field-grid">
              <label>
                <span>${escapeHtml(t("setup_name"))}</span>
                <input name="name" type="text" value="${escapeHtml(defaultName)}" autocomplete="off" data-sync-player-name="true">
              </label>
              <label>
                <span>${escapeHtml(t("setup_players"))}</span>
                <select name="numPlayers">
                  <option value="2">${escapeHtml(t("setup_players_2"))}</option>
                  <option value="4">${escapeHtml(t("setup_players_4"))}</option>
                </select>
              </label>
            </div>
            <button class="primary-button setup-cta" type="submit"${busyAttr("startGame")}>${buttonLabel("startGame", t("setup_start"))}</button>
          </form>
        </article>

        <article class="surface-card form-card online-form-card">
          <div class="card-head">
            <div>
              <span class="section-pill section-pill-hot">${escapeHtml(t("setup_mode_online"))}</span>
              <h2>${escapeHtml(t("setup_online_title"))}</h2>
              <p>${escapeHtml(t("setup_online_note"))}</p>
            </div>
            <strong class="form-emphasis">${escapeHtml(t("setup_online_caption"))}</strong>
          </div>
          <div class="online-flow-grid">
            <form class="mode-form" data-api-action="startOnlineHost" data-form-id="startOnlineHost">
              <div class="mini-head">${escapeHtml(t("setup_host"))}</div>
              <div class="field-grid">
                <label>
                  <span>${escapeHtml(t("setup_name"))}</span>
                  <input name="name" type="text" value="${escapeHtml(defaultName)}" autocomplete="off" data-sync-player-name="true">
                </label>
                <label>
                  <span>${escapeHtml(t("setup_players"))}</span>
                  <select name="numPlayers">
                    <option value="2">${escapeHtml(t("setup_players_2"))}</option>
                    <option value="4">${escapeHtml(t("setup_players_4"))}</option>
                  </select>
                </label>
              </div>
              <details class="details-advanced" data-details-key="advanced-relay">
                <summary>${escapeHtml(t("setup_advanced"))}</summary>
                <div class="advanced-fields">
                  <label>
                    <span>${escapeHtml(t("setup_relay"))}</span>
                    <input name="relay_url" type="text" value="${escapeHtml(state.relayURL)}" placeholder="${escapeHtml(t("relay_placeholder"))}" autocomplete="off" title="${escapeHtml(t("setup_relay_hint"))}">
                  </label>
                  <label>
                    <span>${escapeHtml(t("setup_transport"))}</span>
                    <select name="transport_mode">
                      <option value="">${escapeHtml(t("transport_auto"))}</option>
                      <option value="tcp_tls">${escapeHtml(t("transport_direct"))}</option>
                      <option value="tailnet_tsnet_v1">${escapeHtml(t("transport_tailnet"))}</option>
                      <option value="relay_quic_v2">${escapeHtml(t("transport_relay"))}</option>
                    </select>
                  </label>
                </div>
              </details>
              <button class="secondary-button" type="submit"${busyAttr("startOnlineHost")}>${buttonLabel("startOnlineHost", t("setup_host"))}</button>
            </form>

            <form class="mode-form" data-api-action="joinOnline" data-form-id="joinOnline">
              <div class="mini-head">${escapeHtml(t("setup_join"))}</div>
              <div class="field-grid">
                <label>
                  <span>${escapeHtml(t("setup_name"))}</span>
                  <input name="name" type="text" value="${escapeHtml(defaultName)}" autocomplete="off" data-sync-player-name="true">
                </label>
                <label>
                  <span>${escapeHtml(t("setup_role"))}</span>
                  <select name="role">
                    <option value="auto">${escapeHtml(t("role_auto"))}</option>
                    <option value="partner">${escapeHtml(t("role_partner"))}</option>
                    <option value="opponent">${escapeHtml(t("role_opponent"))}</option>
                  </select>
                </label>
              </div>
              <label>
                <span>${escapeHtml(t("setup_invite"))}</span>
                <input name="key" type="text" autocomplete="off" placeholder="${escapeHtml(t("setup_invite_placeholder"))}">
              </label>
              <button class="secondary-button strong" type="submit"${busyAttr("joinOnline")}>${buttonLabel("joinOnline", t("setup_join"))}</button>
            </form>
          </div>
        </article>
      </div>

      <article class="surface-card rules-banner-card">
        <div class="rules-grid">
          <div class="rules-item">
            <strong>${escapeHtml(t("setup_rules_title"))}</strong>
            <p>${escapeHtml(t("setup_rules_body"))}</p>
          </div>
          <div class="rules-item">
            <strong>${escapeHtml(t("setup_rules_manilha"))}</strong>
            <p><code>${escapeHtml(t("setup_rules_manilha_detail"))}</code></p>
            <strong class="rules-sub-head">${escapeHtml(t("setup_rules_power"))}</strong>
            <p><code>${escapeHtml(t("setup_rules_power_detail"))}</code></p>
          </div>
        </div>
      </article>
    </section>
  `;
}

function renderLobby(): string {
  const bundle = requireBundle();
  const lobby = bundle.lobby;
  const slots = bundle.ui.lobby_slots || [];
  const invite = lobby?.invite_key || "";
  const isHost = bundle.connection.is_host;
  const statusLabel = formatConnectionStatus(bundle.connection.status);

  return `
    <section class="lobby-layout">
      <article class="surface-card invite-card lobby-lead-card">
        <div class="card-head">
          <div>
            <p class="eyebrow">${escapeHtml(isHost ? t("slot_host") : t("connection_online"))}</p>
            <h2>${escapeHtml(t("lobby_title"))}</h2>
            <p class="supporting-copy">${escapeHtml(t("invite_hint"))}</p>
          </div>
          <div class="lobby-top-actions">
            ${isHost ? `<form data-api-action="startOnlineMatch" data-form-id="startOnlineMatch"><button class="primary-button" type="submit"${busyAttr("startOnlineMatch")}>${buttonLabel("startOnlineMatch", t("lobby_start"))}</button></form>` : ""}
            <form data-api-action="closeSession" data-form-id="closeSession">
              <button class="ghost-button danger" type="submit"${busyAttr("closeSession")}>${buttonLabel("closeSession", t("lobby_leave"))}</button>
            </form>
          </div>
        </div>
        ${invite ? `
          <div class="invite-code-row invite-code-row-wide">
            <code class="invite-code" data-copy-text="${escapeHtml(invite)}" role="button" tabindex="0" title="${escapeHtml(t("invite_click_copy"))}">${escapeHtml(invite)}</code>
            <button type="button" class="secondary-button strong" data-copy-text="${escapeHtml(invite)}">${escapeHtml(t("invite_copy"))}</button>
          </div>
          <p class="supporting-copy invite-copy-hint">${escapeHtml(t("invite_click_copy"))}</p>
        ` : ""}
        <div class="telemetry-strip">
          ${renderMetric(t("connection_status"), statusLabel)}
          ${renderMetric(t("connection_mode"), bundle.connection.is_online ? t("connection_online") : t("connection_offline"))}
          ${bundle.lobby?.role ? renderMetric(t("connection_role"), formatRoleLabel(bundle.lobby.role)) : renderMetric(t("lobby_slots"), String(lobby?.num_players || slots.length))}
        </div>
      </article>

      <article class="surface-card seat-card">
        <div class="card-head">
          <h3>${escapeHtml(t("lobby_slots"))}</h3>
          <span class="section-pill">${escapeHtml(`${slots.filter((s) => !s.is_empty).length} / ${lobby?.num_players || slots.length}`)}</span>
        </div>
        <div class="seat-list">
          ${slots.map((slot) => renderSeat(slot)).join("")}
        </div>
      </article>

      <div class="lobby-side-stack">
        <article class="surface-card chat-card">
          <div class="card-head">
            <h3>${escapeHtml(t("lobby_chat"))}</h3>
            <form data-api-action="pollEvents" data-form-id="pollEvents">
              <button class="ghost-button" type="submit"${busyAttr("pollEvents")}>${buttonLabel("pollEvents", t("lobby_refresh"))}</button>
            </form>
          </div>
          <pre class="event-feed" role="log" aria-live="polite">${escapeHtml(renderEventFeed())}</pre>
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
            <input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}">
            <button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat_send"))}</button>
          </form>
        </article>
      </div>
    </section>
  `;
}

function renderSeat(slot: LobbySlotState): string {
  const teamLabel = slot.seat % 2 === 0 ? t("lobby_team_1") : t("lobby_team_2");
  const tags = [
    teamLabel,
    slot.is_local ? t("slot_you") : "",
    slot.is_host ? t("slot_host") : "",
    slot.is_provisional_cpu ? t("slot_cpu") : "",
    !slot.is_empty ? (slot.is_connected ? t("slot_online") : t("slot_offline")) : "",
  ].filter(Boolean);
  const presence = slot.is_empty ? "empty" : slot.is_provisional_cpu ? "cpu" : slot.is_connected ? "online" : "offline";

  return `
    <div class="seat-tile${slot.is_local ? " seat-tile-local" : ""}">
      <div class="seat-heading">
        <strong><i class="presence-dot presence-${presence}" aria-hidden="true"></i>${escapeHtml(slot.name || t("slot_empty"))}</strong>
        <span class="seat-number">#${slot.seat + 1}</span>
      </div>
      <div class="tag-row">${tags.map((tag) => `<span>${escapeHtml(tag)}</span>`).join("")}</div>
      ${slot.can_vote_host || slot.can_request_replacement ? `
        <div class="seat-actions">
          ${slot.can_vote_host ? `<form data-api-action="sendHostVote" data-form-id="sendHostVote-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button" type="submit"${busyAttr(`sendHostVote-${slot.seat}`)}>${buttonLabel(`sendHostVote-${slot.seat}`, t("action_vote_host"))}</button></form>` : ""}
          ${slot.can_request_replacement ? `<form data-api-action="requestReplacementInvite" data-form-id="replacement-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button strong" type="submit"${busyAttr(`replacement-${slot.seat}`)}>${buttonLabel(`replacement-${slot.seat}`, t("action_replacement_invite"))}</button></form>` : ""}
        </div>
      ` : ""}
    </div>
  `;
}

function renderGame(): string {
  const bundle = requireBundle();
  const liveMatch = bundle.match;
  const match = activeDisplayMatch();
  if (!match || !liveMatch) {
    return renderSetup();
  }

  const now = Date.now();
  const trickHoldActive = state.uiFx.trickReveal !== null && state.uiFx.trickReveal.until > now;
  const localPlayer = match.Players.find((player) => player.ID === match.CurrentPlayerIdx) || match.Players[0];
  const localTeamId = localTeam(match);
  const opponentTeamId = localTeamId === 0 ? 1 : 0;
  const pendingFor = trickHoldActive ? -1 : match.PendingRaiseFor;
  const pendingTo = match.PendingRaiseTo || nextStake(match.CurrentHand.Stake);

  // During a trick resolution hold, pause interactive card/raise actions so the user sees the trick finish cleanly
  const canRespond = !trickHoldActive && bundle.ui.actions.can_accept;
  const canRaise = !trickHoldActive && bundle.ui.actions.can_ask_or_raise;
  const canPlay = !trickHoldActive && bundle.ui.actions.can_play_card;

  const topLine = trickHoldActive && state.uiFx.trickReveal
    ? formatTrickRevealHeadline(match, state.uiFx.trickReveal)
    : match.MatchFinished
      ? t("status_match_end")
      : pendingFor === localTeamId
        ? t("status_pending_you", raiseLabel(pendingTo))
        : pendingFor >= 0
          ? t("status_pending_other", playerName(match, match.TurnPlayer), raiseLabel(pendingTo))
          : match.TurnPlayer === match.CurrentPlayerIdx
            ? t("status_your_turn")
            : t("status_wait_turn", playerName(match, match.TurnPlayer));

  const usScore = teamScore(match, localTeamId);
  const themScore = teamScore(match, opponentTeamId);
  const currentStake = match.CurrentHand.Stake;
  const stakeLabel = pendingFor >= 0
    ? `${t("game_pending_raise")} ${raiseLabel(pendingTo)} (${pendingTo})`
    : currentStake === 1
      ? t("game_hud_worth_1")
      : t("game_hud_worth_n", currentStake);

  const tableTitle = isOnlineMode() ? t("game_title_online") : t("game_title_offline");
  const waitingOnOthers = !match.MatchFinished && !canPlay && !canRespond;
  const handStatusClasses = [
    "hand-status",
    trickHoldActive ? "hand-status-reveal" : waitingOnOthers ? "hand-status-waiting" : "hand-status-action",
  ].filter(Boolean).join(" ");

  const enterBoardClass = state.lastRenderedView !== "game" ? " board-fade-in" : "";
  const manilhaRank = match.CurrentHand.Manilha || "-";
  const vira = match.CurrentHand.Vira;

  return `
    <section class="game-layout">
      <header class="game-topbar surface-card score-card">
        <div class="topbar-left">
          <div class="topbar-brand">
            <span class="topbar-mode-pill">${escapeHtml(tableTitle)}</span>
            ${vira && vira.Rank ? `
              <div class="topbar-trump-pill" title="${escapeHtml(t("game_manilha_hint"))}: ${escapeHtml(t("game_manilha_detail"))}">
                <span>${escapeHtml(t("game_vira"))}: <strong>${escapeHtml(vira.Rank)}${escapeHtml(suitSymbol(vira.Suit))}</strong></span>
                <span class="trump-divider">·</span>
                <span class="trump-highlight">★ ${escapeHtml(t("game_manilha"))}: <strong>${escapeHtml(manilhaRank)}</strong></span>
              </div>
            ` : ""}
          </div>
        </div>

        <div class="topbar-scoreboard">
          <div class="score-block score-block-friendly" data-team="us">
            <span>${escapeHtml(t("team_us"))}</span>
            <strong data-score="${usScore}">${usScore}</strong>
          </div>
          <div class="score-center">
            <span class="score-round-tag">${escapeHtml(t("game_round"))} ${Math.min(match.CurrentHand.Round, 3)}/3 · ${escapeHtml(t("game_target"))}</span>
            <strong class="score-stake-badge${pendingFor >= 0 ? " score-stake-pending" : ""}">${escapeHtml(stakeLabel)}</strong>
            <div class="trick-dots" aria-label="${escapeHtml(t("game_trick_track"))}">${renderTrickDots(match)}</div>
          </div>
          <div class="score-block score-block-foe" data-team="them">
            <span>${escapeHtml(t("team_them"))}</span>
            <strong data-score="${themScore}">${themScore}</strong>
          </div>
        </div>

        <div class="topbar-right">
          <button type="button" class="ghost-button topbar-drawer-btn${state.drawerOpen ? " is-active" : ""}" data-toggle-drawer="true" aria-expanded="${state.drawerOpen ? "true" : "false"}">
            ${escapeHtml(t("game_drawer_toggle"))}
          </button>
          <form data-api-action="closeSession" data-form-id="closeSession">
            <button class="ghost-button danger topbar-leave-btn" type="submit"${busyAttr("closeSession")}>${buttonLabel("closeSession", isOnlineMode() ? t("lobby_leave") : t("game_leave"))}</button>
          </form>
        </div>
      </header>

      <div class="game-main-column">
        <article class="board-card">
          <div class="board-stage board-stage-${match.NumPlayers}${enterBoardClass}">
            ${renderViraCorner(match)}
            ${renderPlayers(match, trickHoldActive)}
            ${renderLiveFeedTicker(match)}
            <div class="center-table">
              <div class="table-core">
                ${renderCenterTableArena(match)}
              </div>
            </div>
            ${renderTrickPiles(match)}
          </div>
        </article>

        <article class="surface-card action-dock${canPlay || canRespond ? " action-dock-turn" : ""}">
          ${renderCallResponseBanner(match, localTeamId, canRespond, canRaise, bundle.ui.actions.can_accept, bundle.ui.actions.can_refuse, pendingFor, pendingTo)}
          <div class="action-dock-grid">
            <div class="hand-tray">
              <div class="hand-tray-header">
                <div class="hand-owner">
                  <strong>${escapeHtml(localPlayer?.Name || t("game_you"))}</strong>
                  <span class="team-chip team-chip-ally">${escapeHtml(t("team_us"))}</span>
                </div>
                <div class="${handStatusClasses}" role="status" aria-live="polite">
                  ${waitingOnOthers && !trickHoldActive ? `<i class="status-dot" aria-hidden="true"></i>` : ""}
                  <span>${escapeHtml(topLine)}</span>
                </div>
              </div>
              <div class="hand-row">
                ${(localPlayer?.Hand || []).map((card, index, hand) => renderPlayableCard(match, card, index, hand.length, canPlay, match.CurrentHand.Round >= 2, match.CurrentHand.Manilha)).join("")}
              </div>
            </div>
            ${!canRespond && (canRaise || match.MatchFinished) ? `
              <div class="dock-controls">
                <div class="control-stack">
                  ${renderTurnControls(canRaise, match.MatchFinished && !trickHoldActive, currentStake)}
                </div>
              </div>
            ` : ""}
          </div>
        </article>
      </div>

      ${renderGameSideDrawer(bundle, match)}
      ${!trickHoldActive && liveMatch.MatchFinished ? renderOverlay(liveMatch, localTeamId) : ""}
    </section>
  `;
}

function renderCallResponseBanner(
  match: MatchSnapshot,
  localTeamId: number,
  canRespond: boolean,
  canRaise: boolean,
  canAccept: boolean,
  canRefuse: boolean,
  pendingFor: number,
  pendingTo: number,
): string {
  if (pendingFor < 0 || match.MatchFinished) {
    return "";
  }
  const requesterName = match.CurrentHand.RaiseRequester >= 0
    ? playerName(match, match.CurrentHand.RaiseRequester)
    : pendingFor === localTeamId
      ? t("game_opponent")
      : t("team_us");
  const callName = raiseLabel(pendingTo);
  const currentStake = match.CurrentHand.Stake;

  if (canRespond && pendingFor === localTeamId) {
    const nextLevel = nextStake(pendingTo);
    const canEscalate = canRaise && nextLevel > pendingTo;
    return `
      <div class="call-banner call-banner-urgent" role="region" aria-live="assertive">
        <div class="call-banner-copy">
          <strong>🔥 ${escapeHtml(t("game_call_banner_title", requesterName, callName.toUpperCase()))}</strong>
          <span>${escapeHtml(t("game_call_banner_sub", pendingTo, currentStake, currentStake > 1 ? "s" : ""))}</span>
        </div>
        <div class="call-banner-actions">
          ${canAccept ? `<form data-api-action="accept" data-form-id="accept"><button class="secondary-button call-btn-accept" type="submit"${busyAttr("accept")}>${buttonLabel("accept", t("game_accept_btn", pendingTo))}</button></form>` : ""}
          ${canEscalate ? `<form data-api-action="truco" data-form-id="raise"><button class="primary-button call-btn-raise" type="submit"${busyAttr("raise")}>${buttonLabel("raise", t("game_ask_raise_btn", raiseLabel(nextLevel), nextLevel))}</button></form>` : ""}
          ${canRefuse ? `<form data-api-action="refuse" data-form-id="refuse"><button class="ghost-button danger call-btn-refuse" type="submit"${busyAttr("refuse")}>${buttonLabel("refuse", t("game_refuse_btn", currentStake))}</button></form>` : ""}
        </div>
      </div>
    `;
  }

  const responderName = playerName(match, match.TurnPlayer);
  return `
    <div class="call-banner call-banner-waiting" role="status" aria-live="polite">
      <div class="call-banner-copy">
        <strong>⏳ ${escapeHtml(t("game_call_waiting_title", callName.toUpperCase()))}</strong>
        <span>${escapeHtml(t("game_call_waiting_sub", responderName, pendingTo))}</span>
      </div>
    </div>
  `;
}

function renderLiveFeedTicker(match: MatchSnapshot): string {
  const recentLogs = match.Logs.slice(-3).map((line) => localizeLogLine(line));
  if (recentLogs.length === 0) {
    return "";
  }
  return `
    <aside class="table-live-ticker" aria-label="${escapeHtml(t("game_activity"))}">
      ${recentLogs.map((line, idx) => `<div class="ticker-line${idx === recentLogs.length - 1 ? " ticker-line-latest" : ""}">${escapeHtml(line)}</div>`).join("")}
    </aside>
  `;
}

function renderGameSideDrawer(bundle: SnapshotBundle, match: MatchSnapshot): string {
  if (!state.drawerOpen) {
    return "";
  }
  const manilhaRank = match.CurrentHand.Manilha || "-";
  return `
    <aside class="game-drawer-backdrop" data-toggle-drawer="true">
      <article class="surface-card game-drawer" onclick="event.stopPropagation()">
        <div class="card-head">
          <h3>${escapeHtml(t("game_activity"))}</h3>
          <button type="button" class="ghost-button" data-toggle-drawer="true">×</button>
        </div>
        <div class="telemetry-grid table-note-grid">
          ${renderMetric(t("game_vira"), cardLabel(match.CurrentHand.Vira))}
          ${renderMetric(t("game_manilha"), `${manilhaRank} (♣ > ♥ > ♠ > ♦)`)}
          ${renderMetric(t("game_player_to_move"), playerName(match, match.TurnPlayer))}
        </div>
        <div class="activity-summary">${escapeHtml(lastTrickCopy(match))}</div>
        <pre class="event-feed compact" role="log" aria-live="polite">${escapeHtml(renderEventFeed(match.Logs.slice(-10)))}</pre>
        ${isOnlineMode() ? `
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
            <input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}">
            <button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat_send"))}</button>
          </form>
        ` : ""}
      </article>
    </aside>
  `;
}

function renderPlayers(match: MatchSnapshot, trickHoldActive: boolean): string {
  const positions = seatPositions(match);
  const localTeamId = localTeam(match);
  return match.Players
    .filter((player) => player.ID !== match.CurrentPlayerIdx)
    .map((player) => {
      const position = positions.get(player.ID) || "top";
      const isAlly = player.Team === localTeamId;
      const relation = isAlly ? t("game_partner") : t("game_opponent");
      const teamClass = isAlly ? "team-ally" : "team-foe";
      const teamBadge = isAlly ? "🛡" : "⚔";
      const isTurn = !trickHoldActive && !match.MatchFinished && player.ID === match.TurnPlayer;

      return `
        <div class="player-seat seat-${position} ${teamClass}${isTurn ? " seat-turn" : ""}">
          <div class="player-head">
            <div class="player-title-row">
              <strong><span class="team-badge team-badge-${isAlly ? "ally" : "foe"}" aria-hidden="true">${teamBadge}</span>${escapeHtml(player.Name)}</strong>
              ${isTurn ? `<span class="turn-indicator-pill">${escapeHtml(t("game_turn_badge"))}</span>` : ""}
            </div>
            <span class="team-label team-label-${isAlly ? "ally" : "foe"}">${escapeHtml(relation)}${player.CPU ? ` · ${escapeHtml(t("game_cpu"))}` : ""}</span>
          </div>
          <div class="player-cards" aria-label="${player.Hand.length} cartas">
            ${player.Hand.map(() => `<span class="card-back tiny" aria-hidden="true"></span>`).join("")}
          </div>
        </div>
      `;
    })
    .join("");
}

function renderCenterTableArena(match: MatchSnapshot): string {
  const now = Date.now();
  const reveal = state.uiFx.trickReveal;
  const positions = seatPositions(match);

  if (reveal && reveal.until > now && reveal.cards.length > 0) {
    const headline = formatTrickRevealHeadline(match, reveal);
    const cardsHtml = reveal.cards.map((played, index) => {
      const seatPos = positions.get(played.PlayerID) || "top";
      const isWinner = !reveal.tie && played.PlayerID === reveal.winner;
      const animKey = `reveal:${reveal.seq}:${index}:${played.PlayerID}`;
      const shouldEnter = index === reveal.cards.length - 1 && consumeCardAnimationKey(animKey);
      const classes = [
        "played-card",
        `played-from-${seatPos}`,
        "played-card-replay",
        isWinner ? "played-card-winner" : "",
        reveal.tie ? "played-card-tied" : "",
        shouldEnter ? "played-card-enter" : "",
      ].filter(Boolean).join(" ");

      const statusBadge = isWinner
        ? `<span class="played-card-status played-card-status-win">★ ${escapeHtml(t("trick_winner_card_badge"))}</span>`
        : reveal.tie
          ? `<span class="played-card-status played-card-status-tie">${escapeHtml(t("trick_tied_card_badge"))}</span>`
          : "";

      return `
        <div class="${classes}">
          <span class="played-card-owner">${escapeHtml(playerName(match, played.PlayerID))}</span>
          ${played.FaceDown ? `<span class="card-back small" aria-hidden="true"></span>` : renderCard(played.Card, "small", played.Card.Rank === reveal.manilha)}
          ${statusBadge}
        </div>
      `;
    }).join("");

    return `
      <div class="arena-headline arena-headline-${reveal.tie ? "tie" : reveal.team === localTeam(match) ? "us" : "them"}">
        ${escapeHtml(headline)}
      </div>
      <div class="round-pile round-pile-spatial">${cardsHtml}</div>
    `;
  }

  const roundCards = match.CurrentHand.RoundCards || [];
  if (roundCards.length === 0) {
    const isMyTurn = match.TurnPlayer === match.CurrentPlayerIdx && match.PendingRaiseFor < 0;
    const waitText = isMyTurn
      ? t("game_table_waiting")
      : t("game_table_waiting_other", playerName(match, match.TurnPlayer));
    return `
      <div class="round-pile round-pile-empty">
        <div class="empty-felt-prompt">
          <span class="empty-felt-card-slot" aria-hidden="true"></span>
          <p>${escapeHtml(waitText)}</p>
        </div>
      </div>
    `;
  }

  const leadInfo = evaluateLeadingPlayedCard(roundCards, match.CurrentHand.Manilha);

  const cardsHtml = roundCards.map((played, index) => {
    const seatPos = positions.get(played.PlayerID) || "top";
    const key = roundCardKey(match, played, index);
    const shouldAnimate = consumeCardAnimationKey(key);
    const isLeading = roundCards.length > 1 && !leadInfo.isTie && leadInfo.winnerIndex === index;
    const isTiedLead = roundCards.length > 1 && leadInfo.isTie && leadInfo.tiedIndices.has(index);

    const classes = [
      "played-card",
      `played-from-${seatPos}`,
      shouldAnimate ? "played-card-enter" : "",
      isLeading ? "played-card-leading" : "",
    ].filter(Boolean).join(" ");

    const faceDownClasses = [
      "card-back",
      "small",
      shouldAnimate && played.FaceDown ? "card-back-flip-in" : "",
    ].filter(Boolean).join(" ");

    const statusBadge = isLeading
      ? `<span class="played-card-status played-card-status-lead">${escapeHtml(t("trick_leading_card_badge"))}</span>`
      : isTiedLead
        ? `<span class="played-card-status played-card-status-tie">${escapeHtml(t("trick_tied_card_badge"))}</span>`
        : "";

    return `
      <div class="${classes}">
        <span class="played-card-owner">${escapeHtml(playerName(match, played.PlayerID))}</span>
        ${played.FaceDown ? `<span class="${faceDownClasses}" aria-hidden="true"></span>` : renderCard(played.Card, "small", played.Card.Rank === match.CurrentHand.Manilha)}
        ${statusBadge}
      </div>
    `;
  }).join("");

  return `<div class="round-pile round-pile-spatial">${cardsHtml}</div>`;
}

function evaluateLeadingPlayedCard(
  roundCards: PlayedCard[],
  manilha: string,
): { winnerIndex: number; isTie: boolean; tiedIndices: Set<number> } {
  let bestPower = -1;
  let winnerIndex = -1;
  let isTie = false;
  const tiedIndices = new Set<number>();

  roundCards.forEach((played, idx) => {
    if (played.FaceDown || !played.Card.Rank) {
      return;
    }
    const power = cardPower(played.Card, manilha);
    if (power > bestPower) {
      bestPower = power;
      winnerIndex = idx;
      isTie = false;
      tiedIndices.clear();
      tiedIndices.add(idx);
    } else if (power === bestPower && bestPower > 0) {
      isTie = true;
      tiedIndices.add(idx);
    }
  });

  return { winnerIndex, isTie, tiedIndices };
}

function cardPower(card: Card, manilha: string): number {
  if (manilha && card.Rank === manilha) {
    return 100 + (MANILHA_SUIT_POWER[card.Suit] || 0);
  }
  return NORMAL_RANK_POWER[card.Rank] || 0;
}

function formatTrickRevealHeadline(match: MatchSnapshot, reveal: TrickRevealState): string {
  if (reveal.tie) {
    return t("trick_tie", reveal.round);
  }
  return t("trick_win", playerName(match, reveal.winner), reveal.round);
}

function renderTrickDots(match: MatchSnapshot): string {
  const localTeamId = localTeam(match);
  return Array.from({ length: 3 }, (_, index) => {
    const result = match.CurrentHand.TrickResults[index];
    let label = t("game_trick_pending");
    let className = "trick-dot";
    if (result === -1) {
      label = t("game_trick_draw");
      className += " trick-dot-draw";
    } else if (result === 0 || result === 1) {
      label = result === localTeamId ? t("team_us") : t("team_them");
      className += result === localTeamId ? " trick-dot-us" : " trick-dot-them";
    }
    return `<span class="${className}" title="${escapeHtml(t("game_trick_label", index + 1))} · ${escapeHtml(label)}"></span>`;
  }).join("");
}

function renderViraCorner(match: MatchSnapshot): string {
  const vira = match.CurrentHand.Vira;
  if (!vira || !vira.Rank) {
    return "";
  }
  const manilha = match.CurrentHand.Manilha || "-";
  const viraKey = `${match.MatchPoints["0"] || 0}:${match.MatchPoints["1"] || 0}:${vira.Rank}:${vira.Suit}`;
  const isNewVira = state.uiFx.animatedViraKey !== viraKey;
  if (isNewVira) {
    state.uiFx.animatedViraKey = viraKey;
  }
  const revealClass = isNewVira ? " vira-card-reveal" : "";
  const shimmerClass = isNewVira ? " vira-shimmer" : "";
  return `
    <div class="table-vira-corner${shimmerClass}">
      <span class="table-vira-label">${escapeHtml(t("game_vira"))}</span>
      <div class="vira-deck-stack">
        <span class="card-back small vira-deck-back" aria-hidden="true"></span>
        <span class="vira-card-wrap${revealClass}">${renderCard(vira, "small")}</span>
      </div>
      <span class="manilha-badge" title="${escapeHtml(t("game_manilha_hint"))}: ${escapeHtml(t("game_manilha_detail"))}">★ ${escapeHtml(t("game_manilha"))} ${escapeHtml(manilha)}</span>
    </div>
  `;
}

function renderTrickPiles(match: MatchSnapshot): string {
  const localTeamId = localTeam(match);
  const piles = match.TrickPiles || [];
  const renderSide = (team: number, side: string, label: string): string => {
    const teamPiles = piles.filter((pile) => pile.Team === team && pile.Cards.length > 0);
    if (teamPiles.length === 0) {
      return "";
    }
    const packets = teamPiles
      .map((pile, pileIndex) => {
        const cards = pile.Cards.map((played, cardIndex) => {
          const face = played.FaceDown
            ? `<span class="card-back tiny" aria-hidden="true"></span>`
            : renderCard(played.Card, "tiny", played.Card.Rank === match.CurrentHand.Manilha);
          return `<span class="pile-card" style="--pile-card:${cardIndex}">${face}</span>`;
        }).join("");
        return `<div class="trick-pile-packet" style="--pile-index:${pileIndex}" title="${escapeHtml(t("game_trick_label", pile.Round))} · ${escapeHtml(playerName(match, pile.Winner))}">${cards}</div>`;
      })
      .join("");
    return `
      <div class="trick-pile-side trick-pile-${side}">
        <span class="trick-pile-label">${escapeHtml(label)} (${teamPiles.length})</span>
        <div class="trick-pile-stack">${packets}</div>
      </div>
    `;
  };
  return renderSide(localTeamId, "us", t("game_piles_us")) + renderSide(localTeamId === 0 ? 1 : 0, "them", t("game_piles_them"));
}

function renderPlayableCard(
  match: MatchSnapshot,
  card: Card,
  index: number,
  total: number,
  canPlay: boolean,
  canFaceDown: boolean,
  manilhaRank: string,
): string {
  const key = handCardKey(match, card);
  const shouldDealIn = consumeCardAnimationKey(key);
  const isManilha = manilhaRank !== "" && card.Rank === manilhaRank;
  const classes = [
    "play-card",
    canPlay ? "play-card-playable" : "play-card-locked",
    shouldDealIn ? "deal-in" : "",
  ].filter(Boolean).join(" ");
  const style = `style="--card-index:${index}; --card-total:${total};"`;
  const cardFace = renderCard(card, "regular", isManilha);

  if (!canPlay) {
    return `
      <div class="${classes}" ${style}>
        <div class="lock-card">${cardFace}</div>
        <div class="play-card-actions">
          <span class="card-caption">${escapeHtml(cardLabel(card))}</span>
        </div>
      </div>
    `;
  }

  return `
    <div class="${classes}" ${style}>
      <form data-api-action="play" data-form-id="play-${index}">
        <input type="hidden" name="cardIndex" value="${index}">
        <button class="card-button" type="submit" aria-label="${escapeHtml(`${t("game_play")} ${cardLabel(card)}`)}"${busyAttr(`play-${index}`)}>${cardFace}</button>
      </form>
      <div class="play-card-actions">
        ${canFaceDown
          ? `<form data-api-action="play" data-form-id="play-down-${index}"><input type="hidden" name="cardIndex" value="${index}"><input type="hidden" name="faceDown" value="true"><button class="face-down-pill-btn" type="submit" aria-label="${escapeHtml(`${t("game_face_down")} ${cardLabel(card)}`)}"${busyAttr(`play-down-${index}`)}>${buttonLabel(`play-down-${index}`, t("game_face_down"))}</button></form>`
          : `<span class="card-caption">${escapeHtml(cardLabel(card))}</span>`}
      </div>
    </div>
  `;
}

function renderTurnControls(canRaise: boolean, matchFinished: boolean, currentStake: number): string {
  if (matchFinished) {
    return `<form data-api-action="reset" data-form-id="reset"><button class="primary-button" type="submit"${busyAttr("reset")}>${buttonLabel("reset", t("game_play_again"))}</button></form>`;
  }

  if (!canRaise) {
    return "";
  }

  const targetStake = nextStake(currentStake);
  const label = t("game_ask_raise_btn", raiseLabel(targetStake), targetStake);
  return `
    <div class="control-group">
      <form data-api-action="truco" data-form-id="truco">
        <button class="primary-button truco-call-button" type="submit"${busyAttr("truco")}>🔥 ${buttonLabel("truco", label)}</button>
      </form>
    </div>
  `;
}

function renderOverlay(match: MatchSnapshot, localTeamId: number): string {
  const youWon = match.WinnerTeam === localTeamId;
  const usScore = teamScore(match, localTeamId);
  const themScore = teamScore(match, localTeamId === 0 ? 1 : 0);
  return `
    <div class="overlay-layer overlay-layer-end">
      <div class="overlay-card overlay-card-${youWon ? "win" : "loss"}" role="dialog" aria-modal="true" aria-live="assertive" aria-labelledby="match-result-title">
        <p class="eyebrow">${escapeHtml(t("status_match_end"))}</p>
        <h3 id="match-result-title">${escapeHtml(youWon ? t("overlay_win") : t("overlay_loss"))}</h3>
        <p class="overlay-subtitle">${escapeHtml(youWon ? t("overlay_win_sub") : t("overlay_loss_sub"))}</p>
        <div class="overlay-final-score">
          <div class="overlay-score-block overlay-score-us">
            <span>${escapeHtml(t("team_us"))}</span>
            <strong>${usScore}</strong>
          </div>
          <span class="overlay-score-divider" aria-hidden="true">×</span>
          <div class="overlay-score-block">
            <span>${escapeHtml(t("team_them"))}</span>
            <strong>${themScore}</strong>
          </div>
        </div>
        <div class="overlay-actions">
          ${!isOnlineMode() ? `
            <form data-api-action="startGame" data-form-id="rematchGame">
              <input type="hidden" name="name" value="${escapeHtml(state.playerName || t("name_placeholder"))}">
              <input type="hidden" name="numPlayers" value="${match.NumPlayers}">
              <button class="primary-button" type="submit"${busyAttr("rematchGame")}>${buttonLabel("rematchGame", t("game_new_match_btn"))}</button>
            </form>
          ` : ""}
          <form data-api-action="reset" data-form-id="reset">
            <button class="${isOnlineMode() ? "primary-button" : "ghost-button"}" type="submit"${busyAttr("reset")}>${buttonLabel("reset", t("game_play_again"))}</button>
          </form>
        </div>
      </div>
    </div>
  `;
}

function renderFxLayers(): string {
  const now = Date.now();
  const layers: string[] = [];

  if (state.uiFx.trucoCall && state.uiFx.trucoCall.until > now) {
    layers.push(`
      <div class="truco-flash-overlay" aria-live="polite" aria-atomic="true">
        <div class="truco-flash-card">
          <div class="truco-flash-text">${escapeHtml(state.uiFx.trucoCall.text)}</div>
          ${state.uiFx.trucoCall.sub ? `<div class="truco-flash-sub">${escapeHtml(state.uiFx.trucoCall.sub)}</div>` : ""}
        </div>
      </div>
    `);
  }

  if (state.uiFx.announcement && state.uiFx.announcement.until > now && !state.bundle?.match?.MatchFinished) {
    const announcement = state.uiFx.announcement;
    const remaining = Math.max(400, announcement.until - now);
    layers.push(`
      <div class="announcement-layer" aria-live="polite" aria-atomic="true">
        <div class="announcement-card announcement-${escapeHtml(announcement.kind)}" style="animation-duration:${remaining}ms">
          <strong>${escapeHtml(announcement.title)}</strong>
          ${announcement.sub ? `<span>${escapeHtml(announcement.sub)}</span>` : ""}
        </div>
      </div>
    `);
  }

  if (state.uiFx.confetti && state.uiFx.confetti.until > now) {
    const count = state.uiFx.confetti.strength > 1 ? 48 : 28;
    const pieces = Array.from({ length: count }, (_, index) => `
      <span class="confetti-piece" style="--piece:${index}; --hue:${(index * 37) % 360}; --drift:${index % 2 === 0 ? 1 : -1};"></span>
    `).join("");
    layers.push(`<div class="confetti-layer" aria-hidden="true">${pieces}</div>`);
  }

  return layers.join("");
}

function renderCard(card: Card, size: "tiny" | "small" | "regular" = "regular", manilha = false): string {
  const red = card.Suit === "Copas" || card.Suit === "Ouros";
  const classes = [
    "card-face",
    `card-face-${size}`,
    red ? "card-face-red" : "",
    manilha ? "card-face-manilha" : "",
  ].filter(Boolean).join(" ");
  return `
    <span class="${classes}" role="img" aria-label="${escapeHtml(cardLabel(card))}">
      <span class="card-corner" aria-hidden="true"><b>${escapeHtml(card.Rank)}</b><i>${escapeHtml(suitSymbol(card.Suit))}</i></span>
      <span class="card-center" aria-hidden="true">${escapeHtml(suitSymbol(card.Suit))}</span>
      <span class="card-corner card-corner-bottom" aria-hidden="true"><b>${escapeHtml(card.Rank)}</b><i>${escapeHtml(suitSymbol(card.Suit))}</i></span>
    </span>
  `;
}

function renderMetric(label: string, value: string): string {
  return `<div class="metric"><span>${escapeHtml(label)}</span><strong>${escapeHtml(value)}</strong></div>`;
}

const HIDDEN_EVENT_KINDS = new Set([
  "match_updated",
  "lobby_updated",
  "tick",
  "session_ready",
  "session_closed",
  "locale_changed",
]);

function renderEventFeed(extraLines: string[] = []): string {
  const fromEvents = state.events
    .filter((event) => !HIDDEN_EVENT_KINDS.has(event.kind))
    .map(formatEventLine);
  const localizedExtra = extraLines.map((line) => localizeLogLine(line));
  const lines = localizedExtra.length > 0 ? [...localizedExtra, ...fromEvents.slice(-8)] : fromEvents;
  if (lines.length === 0) {
    return t("game_no_events");
  }
  return lines.slice(-12).join("\n");
}

function formatEventLine(event: RuntimeEvent): string {
  const time = event.timestamp ? event.timestamp.slice(11, 19) : "--:--:--";
  const payload = event.payload || {};
  const kindLabel = t(`event_${event.kind}`);
  const author = typeof payload.author === "string" ? payload.author : "";
  const text = typeof payload.text === "string" ? payload.text : "";
  const message = typeof payload.message === "string" ? payload.message : "";
  const invite = typeof payload.invite_key === "string" ? payload.invite_key : "";
  const detail = [author && text ? `${author}: ${text}` : "", text && !author ? text : "", message, invite]
    .filter(Boolean)
    .join(" · ");
  return `[${time}] ${kindLabel}${detail ? ` · ${detail}` : ""}`;
}

function localizeLogLine(line: string): string {
  const newHand = line.match(/^Nova mão: vira (.+), manilha (.+)\.$/);
  if (newHand) {
    return t("game_log_new_hand", localizeCardText(newHand[1]), newHand[2]);
  }

  const playedFaceDown = line.match(/^(.+) jogou carta virada\.$/);
  if (playedFaceDown) {
    return t("game_log_face_down", playedFaceDown[1]);
  }

  const playedCard = line.match(/^(.+) jogou (.+)\.$/);
  if (playedCard) {
    return t("game_log_played", playedCard[1], localizeCardText(playedCard[2]));
  }

  const acceptedRaise = line.match(/^(.+) aceitou (.+)\. Aposta agora vale (\d+)\.$/);
  if (acceptedRaise) {
    return t("game_log_accept", acceptedRaise[1], localizeRaiseText(acceptedRaise[2]), acceptedRaise[3]);
  }

  const foldedRaise = line.match(/^(.+) correu do truco\.$/);
  if (foldedRaise) {
    return t("game_log_fold", foldedRaise[1]);
  }

  return line;
}

function localizeCardText(raw: string): string {
  const card = raw.match(/^(.+) de (Ouros|Espadas|Copas|Paus)$/);
  if (!card) {
    return raw;
  }
  return t("card_of", card[1], t(`suit_${card[2]}`));
}

function localizeRaiseText(raw: string): string {
  switch (raw.toLowerCase()) {
    case "truco":
      return t("raise_3");
    case "seis":
      return t("raise_6");
    case "nove":
      return t("raise_9");
    case "doze":
      return t("raise_12");
    default:
      return raw;
  }
}

function syncRefreshLoop(): void {
  window.clearTimeout(refreshTimer);
  if (!state.bundle) {
    return;
  }

  const view = currentView();
  const delay = view === "game" && !isOnlineMode() ? 420 : isOnlineMode() ? 850 : 0;
  if (delay === 0) {
    return;
  }

  refreshTimer = window.setTimeout(async () => {
    try {
      if (document.hidden) {
        syncRefreshLoop();
        return;
      }
      // Do not poll new snapshots from the backend while a trick resolution hold is playing out on screen
      if (state.uiFx.trickReveal && state.uiFx.trickReveal.until > Date.now()) {
        syncRefreshLoop();
        return;
      }
      const prevViewSig = state.viewSig;
      const prevFxSig = state.fxSig;
      if (shouldNudgeCpu()) {
        const result = await sendRequest("autoCpuLoopTick", undefined);
        if (result.ok) {
          updateStateFromResult(result);
        }
      } else {
        await syncSnapshot({ withEvents: true });
      }
      state.error = "";
      if (computeViewSignature() !== prevViewSig || computeFxSignature() !== prevFxSig) {
        render();
      } else {
        syncRefreshLoop();
      }
    } catch (error) {
      state.error = errorMessage(error);
      render();
    }
  }, delay);
}

function shouldNudgeCpu(): boolean {
  if (isOnlineMode() || currentView() !== "game") {
    return false;
  }
  const match = state.bundle?.match;
  if (!match || match.MatchFinished) {
    return false;
  }
  const localId = match.CurrentPlayerIdx;
  const localTeamId = localTeam(match);
  if (match.PendingRaiseFor === localTeamId) {
    return false;
  }
  const actor = match.Players.find((player) => player.ID === match.TurnPlayer);
  if (!actor || !actor.CPU || actor.ID === localId) {
    return false;
  }
  return Date.now() - state.lastMatchChangeAt > 4200;
}

function computeViewSignature(): string {
  const now = Date.now();
  const trickHoldActive = state.uiFx.trickReveal !== null && state.uiFx.trickReveal.until > now;
  return JSON.stringify([
    state.bundle?.mode,
    activeDisplayMatch(),
    state.bundle?.lobby,
    state.bundle?.ui,
    state.bundle?.connection,
    trickHoldActive ? state.uiFx.trickReveal?.seq : 0,
    state.events.length,
    state.error,
    state.busyForm,
    state.locale,
    state.initialized,
    state.drawerOpen,
  ]);
}

function computeFxSignature(): string {
  const now = Date.now();
  return JSON.stringify([
    state.uiFx.trucoCall && state.uiFx.trucoCall.until > now ? state.uiFx.trucoCall.id : 0,
    state.uiFx.announcement && state.uiFx.announcement.until > now ? state.uiFx.announcement.id : 0,
    state.uiFx.confetti && state.uiFx.confetti.until > now ? state.uiFx.confetti.id : 0,
    state.bundle?.match?.MatchFinished === true,
  ]);
}

function syncMeasuredBlocks(): void {
  const blocks = viewLayer.querySelectorAll<HTMLElement>("[data-pretext-block]");
  for (const block of blocks) {
    const width = block.clientWidth;
    if (width <= 0) {
      continue;
    }
    const style = window.getComputedStyle(block);
    const whiteSpace = block.dataset.pretextWhitespace === "pre-wrap" ? "pre-wrap" : "normal";
    const key = `${style.font}|${whiteSpace}|${block.textContent || ""}`;
    let preparedText = preparedCache.get(key);
    if (!preparedText) {
      preparedText = prepare(block.textContent || "", style.font, { whiteSpace });
      preparedCache.set(key, preparedText);
    }
    const lineHeight = numericLineHeight(style);
    const result = layout(preparedText, width, lineHeight);
    block.style.minHeight = `${Math.ceil(result.height)}px`;
    block.dataset.pretextLines = String(result.lineCount);
  }
}

function updateVisualFx(
  previousMatch: MatchSnapshot | undefined,
  nextMatch: MatchSnapshot | undefined,
  newEvents: RuntimeEvent[],
): void {
  if (newEvents.some((event) => isIntenseEvent(event))) {
    triggerShake(420);
    triggerPulse(480);
  }

  if (!nextMatch) {
    return;
  }

  const nextRound = nextMatch.CurrentHand.RoundCards || [];
  const previousRound = previousMatch?.CurrentHand.RoundCards || [];
  if (nextRound.length > previousRound.length) {
    triggerPulse(340);
  }

  if (nextMatch.PendingRaiseFor >= 0) {
    const pendingTo = nextMatch.PendingRaiseTo || nextStake(nextMatch.CurrentHand.Stake);
    const signal = `${nextMatch.PendingRaiseFor}:${pendingTo}:${nextMatch.CurrentHand.Round}:${nextMatch.CurrentHand.Turn}`;
    const pendingChanged = previousMatch?.PendingRaiseFor !== nextMatch.PendingRaiseFor
      || previousMatch?.PendingRaiseTo !== nextMatch.PendingRaiseTo;
    if (pendingChanged && signal !== state.uiFx.lastRaiseSignal) {
      const requester = nextMatch.CurrentHand.RaiseRequester >= 0
        ? playerName(nextMatch, nextMatch.CurrentHand.RaiseRequester)
        : "";
      showTrucoCallFlash(raiseLabel(pendingTo), requester);
      state.uiFx.lastRaiseSignal = signal;
    }
  }

  if (!previousMatch) {
    return;
  }

  const localTeamId = localTeam(nextMatch);
  const opponentTeamId = localTeamId === 0 ? 1 : 0;
  const usDelta = teamScore(nextMatch, localTeamId) - teamScore(previousMatch, localTeamId);
  const themDelta = teamScore(nextMatch, opponentTeamId) - teamScore(previousMatch, opponentTeamId);
  const handScored = usDelta > 0 || themDelta > 0;
  const trickResolved = nextMatch.LastTrickSeq > previousMatch.LastTrickSeq && nextMatch.LastTrickCards.length > 0;

  if (trickResolved) {
    const holdMatch = handScored ? buildTrickHoldSnapshot(previousMatch, nextMatch) : null;
    state.uiFx.trickReveal = {
      seq: nextMatch.LastTrickSeq,
      cards: nextMatch.LastTrickCards,
      winner: nextMatch.LastTrickWinner,
      team: nextMatch.LastTrickTeam,
      tie: nextMatch.LastTrickTie,
      round: nextMatch.LastTrickRound,
      manilha: previousMatch.CurrentHand.Manilha,
      until: Date.now() + 1500,
      holdMatch,
    };

    if (handScored) {
      const sub = nextMatch.MatchFinished
        ? ""
        : t("banner_new_hand", cardLabel(nextMatch.CurrentHand.Vira), nextMatch.CurrentHand.Manilha || "-");
      state.uiFx.pendingHandBanner = {
        title: usDelta > 0 ? t("banner_hand_us", usDelta) : t("banner_hand_them", themDelta),
        sub,
        kind: usDelta > 0 ? "hand-us" : "hand-them",
        durationMs: 2800,
        confettiMs: nextMatch.MatchFinished ? 3200 : usDelta > 0 ? 1600 : 0,
      };
    } else {
      const trickTitle = nextMatch.LastTrickTie
        ? t("trick_tie", nextMatch.LastTrickRound)
        : t("trick_win", playerName(nextMatch, nextMatch.LastTrickWinner), nextMatch.LastTrickRound);
      showAnnouncement(trickTitle, "", nextMatch.LastTrickTie ? "tie" : "trick", 1500);
    }
    return;
  }

  if (handScored) {
    // Hand ended without a trick resolution (someone folded to Truco)
    const sub = nextMatch.MatchFinished
      ? ""
      : t("banner_new_hand", cardLabel(nextMatch.CurrentHand.Vira), nextMatch.CurrentHand.Manilha || "-");
    if (usDelta > 0) {
      showAnnouncement(t("banner_hand_us", usDelta), sub, "hand-us", 2800);
      triggerConfetti(nextMatch.MatchFinished ? 3200 : 1600, nextMatch.MatchFinished ? 2 : 1);
    } else {
      showAnnouncement(t("banner_hand_them", themDelta), sub, "hand-them", 2800);
    }
  } else if (!previousMatch.MatchFinished && nextMatch.MatchFinished) {
    triggerConfetti(3200, 2);
    triggerShake(480);
  }
}

function buildTrickHoldSnapshot(previousMatch: MatchSnapshot, nextMatch: MatchSnapshot): MatchSnapshot {
  const lastCards = nextMatch.LastTrickCards || [];
  const finalPlayed = lastCards[lastCards.length - 1];
  const players = previousMatch.Players.map((player) => {
    if (!finalPlayed || player.ID !== finalPlayed.PlayerID) {
      return { ...player, Hand: [...player.Hand] };
    }
    const remaining = [...player.Hand];
    const matchIdx = remaining.findIndex(
      (c) => c.Rank === finalPlayed.Card.Rank && c.Suit === finalPlayed.Card.Suit,
    );
    if (matchIdx >= 0) {
      remaining.splice(matchIdx, 1);
    } else if (remaining.length > 0) {
      remaining.pop();
    }
    return { ...player, Hand: remaining };
  });

  const trickResult = nextMatch.LastTrickTie ? -1 : nextMatch.LastTrickTeam;
  return {
    ...previousMatch,
    Players: players,
    CurrentHand: {
      ...previousMatch.CurrentHand,
      RoundCards: [],
      TrickResults: [...previousMatch.CurrentHand.TrickResults, trickResult],
      PendingRaiseFor: -1,
    },
    PendingRaiseFor: -1,
    LastTrickCards: nextMatch.LastTrickCards,
    LastTrickSeq: nextMatch.LastTrickSeq,
    LastTrickTeam: nextMatch.LastTrickTeam,
    LastTrickWinner: nextMatch.LastTrickWinner,
    LastTrickTie: nextMatch.LastTrickTie,
    LastTrickRound: nextMatch.LastTrickRound,
    Logs: nextMatch.Logs,
  };
}

function consumeCardAnimationKey(key: string): boolean {
  if (state.uiFx.animatedCardKeys.has(key)) {
    return false;
  }
  state.uiFx.animatedCardKeys.add(key);
  if (state.uiFx.animatedCardKeys.size > 240) {
    const first = state.uiFx.animatedCardKeys.values().next().value;
    if (first) {
      state.uiFx.animatedCardKeys.delete(first);
    }
  }
  return true;
}

function isIntenseEvent(event: RuntimeEvent): boolean {
  const payload = event.payload || {};
  const author = typeof payload.author === "string" ? payload.author : "";
  const message = typeof payload.message === "string" ? payload.message : "";
  const text = typeof payload.text === "string" ? payload.text : "";
  const content = `${event.kind} ${author} ${message} ${text}`.toLowerCase();
  return content.includes("truco")
    || content.includes("seis")
    || content.includes("doze")
    || content.includes("raise")
    || content.includes("refuse")
    || content.includes("accept");
}

function triggerShake(durationMs: number): void {
  window.clearTimeout(shakeTimeout);
  viewLayer.classList.add("fx-shake");
  shakeTimeout = window.setTimeout(() => {
    viewLayer.classList.remove("fx-shake");
  }, durationMs);
}

function triggerPulse(durationMs: number): void {
  window.clearTimeout(pulseTimeout);
  viewLayer.classList.add("fx-pulse");
  pulseTimeout = window.setTimeout(() => {
    viewLayer.classList.remove("fx-pulse");
  }, durationMs);
}

function triggerConfetti(durationMs: number, strength: number): void {
  fxSeq++;
  state.uiFx.confetti = {
    id: fxSeq,
    strength,
    until: Date.now() + durationMs,
  };
}

function showTrucoCallFlash(callLabel: string, requesterName: string): void {
  fxSeq++;
  state.uiFx.trucoCall = {
    id: fxSeq,
    text: `${callLabel.toUpperCase()}!`,
    sub: requesterName ? t("game_call_banner_title", requesterName, callLabel) : "",
    until: Date.now() + 1450,
  };
  triggerShake(380);
  triggerPulse(520);
}

function showAnnouncement(title: string, sub: string, kind: string, durationMs: number): void {
  fxSeq++;
  state.uiFx.announcement = {
    id: fxSeq,
    title,
    sub,
    kind,
    until: Date.now() + durationMs,
  };
}

function numericLineHeight(style: CSSStyleDeclaration): number {
  const parsed = Number.parseFloat(style.lineHeight);
  if (Number.isFinite(parsed)) {
    return parsed;
  }
  const fontSize = Number.parseFloat(style.fontSize);
  return Number.isFinite(fontSize) ? fontSize * 1.3 : 22;
}

function localeOptions(active: LocaleCode): string {
  return [
    ["pt-BR", "Português (BR)"],
    ["en-US", "English (US)"],
  ]
    .map(([value, label]) => `<option value="${value}"${active === value ? " selected" : ""}>${label}</option>`)
    .join("");
}

function formPayload(form: HTMLFormElement): Record<string, unknown> {
  const payload: Record<string, unknown> = {};
  const data = new FormData(form);
  data.forEach((value, key) => {
    if (typeof value !== "string") {
      return;
    }
    if (value === "true") {
      payload[key] = true;
      return;
    }
    if (value === "false") {
      payload[key] = false;
      return;
    }
    if (/^-?\d+$/.test(value)) {
      payload[key] = Number.parseInt(value, 10);
      return;
    }
    payload[key] = value;
  });
  return payload;
}

function persistInputs(payload: Record<string, unknown>): void {
  if (typeof payload.name === "string" && payload.name.trim() !== "") {
    state.playerName = payload.name.trim();
    localStorage.setItem(PLAYER_KEY, state.playerName);
  }
  if (typeof payload.relay_url === "string") {
    state.relayURL = payload.relay_url;
    localStorage.setItem(RELAY_KEY, state.relayURL);
  }
}

async function copyInvite(button: HTMLElement): Promise<void> {
  const value = button.dataset.copyText || "";
  if (!value) {
    return;
  }
  const copied = await writeClipboard(value);
  if (!copied) {
    return;
  }
  button.classList.add("copy-flash");
  window.setTimeout(() => {
    button.classList.remove("copy-flash");
  }, 1000);
  if (button.tagName === "BUTTON") {
    const label = button.textContent || t("invite_copy");
    button.textContent = t("copy_done");
    window.setTimeout(() => {
      button.textContent = label;
    }, 1000);
  }
}

async function writeClipboard(value: string): Promise<boolean> {
  if (navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(value);
      return true;
    } catch {
      // fall through to the legacy path
    }
  }
  const helper = document.createElement("textarea");
  helper.value = value;
  helper.setAttribute("readonly", "");
  helper.style.position = "fixed";
  helper.style.opacity = "0";
  document.body.appendChild(helper);
  helper.select();
  let ok = false;
  try {
    ok = document.execCommand("copy");
  } catch {
    ok = false;
  }
  helper.remove();
  return ok;
}

function busyAttr(formId: string): string {
  return state.busyForm === formId ? " disabled" : "";
}

function buttonLabel(formId: string, label: string): string {
  return state.busyForm === formId ? t("button_busy") : label;
}

function lastTrickCopy(match: MatchSnapshot): string {
  if (match.LastTrickRound <= 0) {
    return t("status_idle");
  }
  if (match.LastTrickTie) {
    return t("trick_tie", match.LastTrickRound);
  }
  return t("trick_win", playerName(match, match.LastTrickWinner), match.LastTrickRound);
}

function handIdentity(match: MatchSnapshot): string {
  return `${match.MatchPoints["0"] || 0}:${match.MatchPoints["1"] || 0}:${match.CurrentHand.Vira.Rank}:${match.CurrentHand.Vira.Suit}`;
}

function handCardKey(match: MatchSnapshot, card: Card): string {
  return `deal:${handIdentity(match)}:${card.Rank}:${card.Suit}`;
}

function roundCardKey(match: MatchSnapshot, played: PlayedCard, index: number): string {
  return `play:${handIdentity(match)}:${match.CurrentHand.Round}:${index}:${played.PlayerID}:${played.Card.Rank}:${played.Card.Suit}:${played.FaceDown ? "down" : "up"}`;
}

function playerName(match: MatchSnapshot, playerId: number): string {
  return match.Players.find((player) => player.ID === playerId)?.Name || "?";
}

function teamScore(match: MatchSnapshot, team: number): number {
  return match.MatchPoints[String(team)] || 0;
}

function localTeam(match: MatchSnapshot): number {
  return match.Players.find((player) => player.ID === match.CurrentPlayerIdx)?.Team || 0;
}

function nextStake(current: number): number {
  switch (current) {
    case 1:
      return 3;
    case 3:
      return 6;
    case 6:
      return 9;
    case 9:
      return 12;
    default:
      return current;
  }
}

function raiseLabel(value: number): string {
  switch (value) {
    case 3:
      return t("raise_3");
    case 6:
      return t("raise_6");
    case 9:
      return t("raise_9");
    case 12:
      return t("raise_12");
    default:
      return String(value);
  }
}

function formatConnectionStatus(status: string): string {
  switch (status) {
    case "host_lobby":
      return t("lobby_status_host_lobby");
    case "client_lobby":
      return t("lobby_status_client_lobby");
    case "host_match":
    case "client_match":
      return t("lobby_status_online_match");
    default:
      return status;
  }
}

function formatRoleLabel(role: string): string {
  switch (role) {
    case "partner":
      return t("role_partner");
    case "opponent":
      return t("role_opponent");
    default:
      return t("role_auto");
  }
}

function cardLabel(card: Card): string {
  if (!card || !card.Rank) {
    return t("game_covered_card");
  }
  return t("card_of", card.Rank, t(`suit_${card.Suit}`));
}

function suitSymbol(suit: string): string {
  switch (suit) {
    case "Ouros":
      return "♦";
    case "Espadas":
      return "♠";
    case "Copas":
      return "♥";
    case "Paus":
      return "♣";
    default:
      return "?";
  }
}

function seatPositions(match: MatchSnapshot): Map<number, string> {
  const positions = new Map<number, string>();
  const local = match.CurrentPlayerIdx;
  const layoutOrder = match.NumPlayers === 2 ? ["bottom", "top"] : ["bottom", "right", "top", "left"];
  for (const player of match.Players) {
    const offset = (player.ID - local + match.NumPlayers) % match.NumPlayers;
    positions.set(player.ID, layoutOrder[offset] || "top");
  }
  return positions;
}

function readLocale(raw: string | null): LocaleCode {
  return raw === "en-US" ? "en-US" : "pt-BR";
}

function requireBundle(): SnapshotBundle {
  if (!state.bundle) {
    throw new Error("bundle unavailable");
  }
  return state.bundle;
}

function t(key: string, ...args: Array<string | number>): string {
  return translate(state.locale, key, ...args);
}

function errorMessage(value: unknown): string {
  if (value instanceof Error) {
    return value.message;
  }
  return String(value || "unknown error");
}

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
