import { layout, prepare } from "@chenglou/pretext";
import { translate } from "./i18n";
import type {
  ApiResult,
  Card,
  LocaleCode,
  LobbySlotState,
  MatchSnapshot,
  RuntimeEvent,
  SnapshotBundle,
} from "./types";

type ViewName = "setup" | "lobby" | "game";

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
}

const SESSION_KEY = "truco-browser-session-id";
const LOCALE_KEY = "truco-browser-locale";
const PLAYER_KEY = "truco-browser-player-name";
const RELAY_KEY = "truco-browser-relay-url";
const EVENT_LIMIT = 80;
const rootElement = document.querySelector<HTMLElement>("#app");

if (!rootElement) {
  throw new Error("missing #app root");
}

const root: HTMLElement = rootElement;

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
};

const preparedCache = new Map<string, ReturnType<typeof prepare>>();
let refreshTimer = 0;
let resizeTimer = 0;

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
  updateStateFromResult(localeResult, { replaceEvents: true });
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
    updateStateFromResult(result, { replaceEvents: true });
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

    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  } finally {
    state.busyForm = "";
    render();
  }
}

async function syncAfterAction(action: string): Promise<void> {
  switch (action) {
    case "pollEvents":
      return;
    case "setLocale":
      await syncSnapshot({ withEvents: currentView() !== "setup" });
      return;
    default:
      await syncSnapshot({ withEvents: isOnlineMode() });
  }
}

async function syncSnapshot(options: { withEvents: boolean }): Promise<void> {
  await ensureSession();
  const action = options.withEvents ? "pollEvents" : "snapshot";
  const result = await sendRequest(action, undefined);
  if (!result.ok) {
    throw new Error(result.error || "snapshot failed");
  }
  updateStateFromResult(result, { replaceEvents: action === "pollEvents" });
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

  const response = await fetch(`/api/${action}`, {
    method: "POST",
    headers,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "same-origin",
  });

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

function updateStateFromResult(result: ApiResult, options?: { replaceEvents?: boolean }): void {
  if (result.bundle) {
    state.bundle = result.bundle;
    state.locale = result.bundle.locale || state.locale;
    document.documentElement.lang = state.locale === "en-US" ? "en" : "pt-BR";
    localStorage.setItem(LOCALE_KEY, state.locale);
  }

  if (result.events) {
    state.events = options?.replaceEvents
      ? result.events.slice(-EVENT_LIMIT)
      : [...state.events, ...result.events].slice(-EVENT_LIMIT);
  }

  if (result.error) {
    state.error = result.error;
  }
}

function clearSession(): void {
  state.sessionId = "";
  sessionStorage.removeItem(SESSION_KEY);
  state.bundle = null;
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

function render(): void {
  root.innerHTML = renderApp();
  syncRefreshLoop();
  syncMeasuredBlocks();
}

function renderApp(): string {
  const locale = state.locale;
  const busy = state.busyForm !== "";
  const banner = state.error
    ? `<section class="runtime-banner" data-pretext-block="lock-height">${escapeHtml(state.error)}</section>`
    : "";

  return `
    <div class="page-shell">
      <div class="page-aura page-aura-left"></div>
      <div class="page-aura page-aura-right"></div>
      <main class="app-shell">
        <header class="hero-card">
          <div class="hero-copy">
            <p class="eyebrow">${escapeHtml(t("app_kicker"))}</p>
            <div class="hero-title-row">
              <h1>${escapeHtml(t("title_main"))}</h1>
              <span class="hero-stamp">${escapeHtml(t("app_stamp"))}</span>
            </div>
            <p class="hero-subtitle" data-pretext-block="lock-height">${escapeHtml(t("title_sub"))}</p>
          </div>
          <form class="locale-card" data-api-action="setLocale" data-form-id="setLocale">
            <label for="locale-select">${escapeHtml(t("locale_label"))}</label>
            <select id="locale-select" name="locale" ${busy ? "disabled" : ""}>
              ${localeOptions(locale)}
            </select>
          </form>
        </header>
        ${banner}
        ${renderView()}
      </main>
    </div>
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
  return `
    <section class="setup-shell">
      <article class="surface-card intro-card">
        <div class="card-head card-head-roomy">
          <div>
            <p class="eyebrow">${escapeHtml(t("setup_title"))}</p>
            <h2>${escapeHtml(t("setup_title"))}</h2>
          </div>
          <span class="section-pill">${escapeHtml(t("app_stamp"))}</span>
        </div>
        <p class="lede" data-pretext-block="lock-height">${escapeHtml(t("setup_intro"))}</p>
        <p class="supporting-copy">${escapeHtml(t("setup_help"))}</p>
        <div class="intro-grid">
          ${renderSetupNote(t("setup_signal_title"), t("setup_signal_body"))}
          ${renderSetupNote(t("setup_runtime_title"), t("setup_runtime_body"))}
        </div>
      </article>
      <div class="setup-stack">
        <article class="surface-card form-card primary-form-card">
          <div class="card-head card-head-roomy">
            <div>
              <span class="section-pill">${escapeHtml(t("setup_mode_offline"))}</span>
              <h3>${escapeHtml(t("setup_offline_title"))}</h3>
              <p>${escapeHtml(t("setup_offline_note"))}</p>
            </div>
            <strong class="form-emphasis">${escapeHtml(t("setup_offline_caption"))}</strong>
          </div>
          <form data-api-action="startGame" data-form-id="startGame">
            <div class="field-grid">
              <label>
                <span>${escapeHtml(t("setup_name"))}</span>
                <input name="name" type="text" value="${escapeHtml(state.playerName || t("name_placeholder"))}" autocomplete="off">
              </label>
              <label>
                <span>${escapeHtml(t("setup_players"))}</span>
                <select name="numPlayers">
                  <option value="2">2</option>
                  <option value="4">4</option>
                </select>
              </label>
            </div>
            <button class="primary-button" type="submit"${busyAttr("startGame")}>${buttonLabel("startGame", t("setup_start"))}</button>
          </form>
        </article>
        <article class="surface-card form-card online-form-card">
          <div class="card-head card-head-roomy">
            <div>
              <span class="section-pill section-pill-hot">${escapeHtml(t("setup_mode_online"))}</span>
              <h3>${escapeHtml(t("setup_online_title"))}</h3>
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
                  <input name="name" type="text" value="${escapeHtml(state.playerName || t("name_placeholder"))}" autocomplete="off">
                </label>
                <label>
                  <span>${escapeHtml(t("setup_players"))}</span>
                  <select name="numPlayers">
                    <option value="2">2</option>
                    <option value="4">4</option>
                  </select>
                </label>
              </div>
              <label>
                <span>${escapeHtml(t("setup_relay"))}</span>
                <input name="relay_url" type="text" value="${escapeHtml(state.relayURL)}" placeholder="${escapeHtml(t("relay_placeholder"))}" autocomplete="off">
              </label>
              <button class="secondary-button" type="submit"${busyAttr("startOnlineHost")}>${buttonLabel("startOnlineHost", t("setup_host"))}</button>
            </form>
            <form class="mode-form" data-api-action="joinOnline" data-form-id="joinOnline">
              <div class="mini-head">${escapeHtml(t("setup_join"))}</div>
              <div class="field-grid">
                <label>
                  <span>${escapeHtml(t("setup_name"))}</span>
                  <input name="name" type="text" value="${escapeHtml(state.playerName || t("name_placeholder"))}" autocomplete="off">
                </label>
                <label>
                  <span>${escapeHtml(t("setup_invite"))}</span>
                  <input name="key" type="text" autocomplete="off">
                </label>
              </div>
              <label>
                <span>${escapeHtml(t("setup_role"))}</span>
                <select name="role">
                  <option value="auto">${escapeHtml(t("role_auto"))}</option>
                  <option value="partner">${escapeHtml(t("role_partner"))}</option>
                  <option value="opponent">${escapeHtml(t("role_opponent"))}</option>
                </select>
              </label>
              <button class="secondary-button strong" type="submit"${busyAttr("joinOnline")}>${buttonLabel("joinOnline", t("setup_join"))}</button>
            </form>
          </div>
          <p class="supporting-copy inline-note">${escapeHtml(t("setup_online_support"))}</p>
        </article>
      </div>
    </section>
  `;
}

function renderSetupNote(title: string, copy: string): string {
  return `
    <div class="intro-note">
      <strong>${escapeHtml(title)}</strong>
      <p>${escapeHtml(copy)}</p>
    </div>
  `;
}

function renderLobby(): string {
  const bundle = requireBundle();
  const lobby = bundle.lobby;
  const network = bundle.connection.network;
  const slots = bundle.ui.lobby_slots || [];
  const invite = lobby?.invite_key || "";
  const isHost = bundle.connection.is_host;

  return `
    <section class="lobby-layout">
      <article class="surface-card invite-card lobby-lead-card">
        <div class="card-head card-head-roomy">
          <div>
            <p class="eyebrow">${escapeHtml(t("lobby_title"))}</p>
            <h2>${escapeHtml(t("lobby_title"))}</h2>
            <p class="supporting-copy">${escapeHtml(t("invite_hint"))}</p>
          </div>
          <span class="section-pill">${escapeHtml(isHost ? t("slot_host") : t("connection_online"))}</span>
        </div>
        <div class="invite-code-row invite-code-row-wide">
          <code class="invite-code">${escapeHtml(invite || "----")}</code>
          ${invite ? `<button type="button" class="ghost-button" data-copy-text="${escapeHtml(invite)}">${escapeHtml(t("invite_copy"))}</button>` : ""}
          ${isHost ? `<form data-api-action="startOnlineMatch" data-form-id="startOnlineMatch"><button class="primary-button" type="submit"${busyAttr("startOnlineMatch")}>${buttonLabel("startOnlineMatch", t("lobby_start"))}</button></form>` : ""}
        </div>
        <div class="telemetry-strip">
          ${renderMetric(t("connection_status"), bundle.connection.status)}
          ${renderMetric(t("connection_transport"), network?.transport || "-")}
          ${renderMetric(t("connection_protocol"), protocolLabel(network))}
        </div>
      </article>

      <article class="surface-card seat-card">
        <div class="card-head">
          <h3>${escapeHtml(t("lobby_slots"))}</h3>
          <span class="section-pill">${escapeHtml(String(lobby?.num_players || slots.length))}</span>
        </div>
        <div class="seat-list">
          ${slots.map((slot) => renderSeat(slot)).join("")}
        </div>
      </article>

      <article class="surface-card telemetry-card">
        <div class="telemetry-grid">
          ${renderMetric(t("connection_status"), bundle.connection.status)}
          ${renderMetric(t("connection_mode"), bundle.connection.is_online ? t("connection_online") : t("connection_offline"))}
          ${renderMetric(t("connection_transport"), network?.transport || "-")}
          ${renderMetric(t("connection_protocol"), protocolLabel(network))}
          ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
          ${bundle.lobby?.role ? renderMetric(t("connection_role"), bundle.lobby.role) : ""}
        </div>
      </article>

      <article class="surface-card event-card">
        <div class="card-head">
          <h3>${escapeHtml(t("lobby_events"))}</h3>
          <form data-api-action="pollEvents" data-form-id="pollEvents">
            <button class="ghost-button" type="submit"${busyAttr("pollEvents")}>${buttonLabel("pollEvents", t("lobby_refresh"))}</button>
          </form>
        </div>
        <pre class="event-feed" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed())}</pre>
      </article>

      <div class="lobby-side-stack">
        <article class="surface-card chat-card">
          <div class="card-head">
            <h3>${escapeHtml(t("lobby_chat"))}</h3>
          </div>
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
            <input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}">
            <button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat"))}</button>
          </form>
          <form data-api-action="closeSession" data-form-id="closeSession">
            <button class="ghost-button danger" type="submit"${busyAttr("closeSession")}>${buttonLabel("closeSession", t("lobby_leave"))}</button>
          </form>
        </article>
      </div>
    </section>
  `;
}

function renderSeat(slot: LobbySlotState): string {
  const tags = [
    slot.is_local ? t("slot_you") : "",
    slot.is_host ? t("slot_host") : "",
    slot.is_provisional_cpu ? t("slot_cpu") : "",
    slot.is_connected ? t("slot_online") : t("slot_offline"),
  ].filter(Boolean);

  return `
    <div class="seat-tile${slot.is_local ? " seat-tile-local" : ""}">
      <div class="seat-heading">
        <strong>${escapeHtml(slot.name || t("slot_empty"))}</strong>
        <span>#${slot.seat + 1}</span>
      </div>
      <div class="tag-row">${tags.map((tag) => `<span>${escapeHtml(tag)}</span>`).join("")}</div>
      <div class="seat-actions">
        ${slot.can_vote_host ? `<form data-api-action="sendHostVote" data-form-id="sendHostVote-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button" type="submit"${busyAttr(`sendHostVote-${slot.seat}`)}>${buttonLabel(`sendHostVote-${slot.seat}`, t("action_vote_host"))}</button></form>` : ""}
        ${slot.can_request_replacement ? `<form data-api-action="requestReplacementInvite" data-form-id="replacement-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button strong" type="submit"${busyAttr(`replacement-${slot.seat}`)}>${buttonLabel(`replacement-${slot.seat}`, t("action_replacement_invite"))}</button></form>` : ""}
      </div>
    </div>
  `;
}

function renderGame(): string {
  const bundle = requireBundle();
  const match = bundle.match;
  if (!match) {
    return renderSetup();
  }

  const localPlayer = match.Players.find((player) => player.ID === match.CurrentPlayerIdx) || match.Players[0];
  const localTeamId = localTeam(match);
  const pendingFor = match.PendingRaiseFor;
  const pendingTo = match.PendingRaiseTo || nextStake(match.CurrentHand.Stake);
  const canRespond = bundle.ui.actions.can_accept;
  const canRaise = bundle.ui.actions.can_ask_or_raise;
  const canPlay = bundle.ui.actions.can_play_card;
  const topLine = match.MatchFinished
    ? t("status_match_end")
    : pendingFor === localTeamId
      ? t("status_pending_you", raiseLabel(pendingTo))
      : pendingFor >= 0
        ? t("status_pending_other", playerName(match, match.TurnPlayer), raiseLabel(pendingTo))
        : match.TurnPlayer === match.CurrentPlayerIdx
          ? t("status_your_turn")
          : t("status_wait_turn", playerName(match, match.TurnPlayer));
  const bottomScore = teamScore(match, 0);
  const topScore = teamScore(match, 1);
  const tableTitle = isOnlineMode() ? t("game_title_online") : t("game_title_offline");

  return `
    <section class="game-layout">
      <article class="surface-card score-card">
        <div class="score-block${localTeamId === 0 ? " score-block-friendly" : ""}">
          <span>${escapeHtml(t("team_one"))}</span>
          <strong>${bottomScore}</strong>
        </div>
        <div class="score-center">
          <span>${escapeHtml(t("game_round"))} ${match.CurrentHand.Round}/3</span>
          <strong>${escapeHtml(t("game_stake"))} ${match.CurrentHand.Stake}</strong>
        </div>
        <div class="score-block${localTeamId === 1 ? " score-block-friendly" : ""}">
          <span>${escapeHtml(t("team_two"))}</span>
          <strong>${topScore}</strong>
        </div>
      </article>

      <div class="game-main-column">
        <article class="surface-card board-card">
          <div class="card-head board-head">
            <div>
              <p class="eyebrow">${escapeHtml(tableTitle)}</p>
              <h2>${escapeHtml(tableTitle)}</h2>
            </div>
            ${isOnlineMode() ? `<form data-api-action="closeSession" data-form-id="closeSession"><button class="ghost-button danger" type="submit"${busyAttr("closeSession")}>${buttonLabel("closeSession", t("lobby_leave"))}</button></form>` : ""}
          </div>
          <div class="status-band" data-pretext-block="lock-height">
            <span>${escapeHtml(t("game_status"))}</span>
            <strong>${escapeHtml(topLine)}</strong>
          </div>
          <div class="board-stage board-stage-${match.NumPlayers}">
            ${renderPlayers(match)}
            <div class="center-table">
              <div class="table-shell">
                <div class="table-chip">
                  <span>${escapeHtml(t("game_vira"))}</span>
                  ${renderCard(match.CurrentHand.Vira, "small")}
                </div>
                <div class="table-core">
                  <div class="table-rail">
                    <span>${escapeHtml(t("game_trick_track"))}</span>
                    <div class="trick-track">${renderTrickTrack(match)}</div>
                  </div>
                  <div class="round-pile">${renderRoundCards(match)}</div>
                </div>
                <div class="table-chip">
                  <span>${escapeHtml(t("game_manilha"))}</span>
                  <strong>${escapeHtml(match.CurrentHand.Manilha || "-")}</strong>
                </div>
              </div>
            </div>
          </div>
        </article>

        <article class="surface-card action-dock">
          <div class="card-head">
            <div>
              <p class="eyebrow">${escapeHtml(t("game_hand"))}</p>
              <h3>${escapeHtml(localPlayer?.Name || t("game_you"))}</h3>
            </div>
            <span class="section-pill">${localPlayer?.Hand.length || 0}</span>
          </div>
          <div class="action-dock-grid">
            <div class="hand-tray">
              <div class="hand-row">
                ${(localPlayer?.Hand || []).map((card, index) => renderPlayableCard(card, index, canPlay, match.CurrentHand.Round >= 2)).join("")}
              </div>
            </div>
            <div class="dock-controls">
              <div class="card-head compact-head">
                <h3>${escapeHtml(t("game_controls"))}</h3>
                <span class="section-pill">${escapeHtml(match.MatchFinished ? t("game_wait") : t("game_turn"))}</span>
              </div>
              <div class="control-stack">
                ${canRespond ? renderRespondControls(canRaise, bundle.ui.actions.can_accept, bundle.ui.actions.can_refuse, pendingTo) : renderTurnControls(canRaise, match.MatchFinished)}
                ${!match.MatchFinished ? `<form data-api-action="pollEvents" data-form-id="pollEvents"><button class="ghost-button" type="submit"${busyAttr("pollEvents")}>${buttonLabel("pollEvents", t("lobby_refresh"))}</button></form>` : ""}
              </div>
            </div>
          </div>
        </article>
      </div>

      <div class="game-side-column">
        <article class="surface-card activity-card">
          <div class="card-head">
            <h3>${escapeHtml(t("game_activity"))}</h3>
            <span class="section-pill">${escapeHtml(t("game_last_trick"))}</span>
          </div>
          <div class="activity-summary" data-pretext-block="lock-height">${escapeHtml(lastTrickCopy(match))}</div>
          <pre class="event-feed compact" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed(match.Logs.slice(-4)))}</pre>
        </article>
        <article class="surface-card table-note-card">
          <div class="card-head">
            <h3>${escapeHtml(t("game_table_notes"))}</h3>
            <span class="section-pill">${escapeHtml(t("game_round"))} ${match.CurrentHand.Round}/3</span>
          </div>
          <div class="telemetry-grid table-note-grid">
            ${renderMetric(t("game_vira"), cardLabel(match.CurrentHand.Vira))}
            ${renderMetric(t("game_manilha"), match.CurrentHand.Manilha || "-")}
            ${renderMetric(t("game_player_to_move"), playerName(match, match.TurnPlayer))}
          </div>
        </article>
        ${isOnlineMode() ? renderNetworkPanel(bundle) : ""}
      </div>
      ${match.MatchFinished ? renderOverlay(match, localTeamId) : ""}
    </section>
  `;
}

function renderPlayers(match: MatchSnapshot): string {
  const positions = seatPositions(match);
  const localTeamId = localTeam(match);
  return match.Players
    .filter((player) => player.ID !== match.CurrentPlayerIdx)
    .map((player) => {
      const position = positions.get(player.ID) || "top";
      const relation = player.Team === localTeamId ? t("game_partner") : t("game_opponent");
      const isTurn = player.ID === match.TurnPlayer;

      return `
        <div class="player-seat seat-${position}${isTurn ? " seat-turn" : ""}">
          <div class="player-head">
            <strong>${escapeHtml(player.Name)}</strong>
            <span>${escapeHtml(relation)}${player.CPU ? ` · ${escapeHtml(t("game_cpu"))}` : ""}</span>
          </div>
          <div class="player-cards">
            ${player.Hand.map(() => `<span class="card-back tiny"></span>`).join("")}
          </div>
        </div>
      `;
    })
    .join("");
}

function renderRoundCards(match: MatchSnapshot): string {
  const roundCards = match.CurrentHand.RoundCards || [];
  if (roundCards.length === 0) {
    return `<div class="round-card-placeholder">${escapeHtml(t("game_table_waiting"))}</div>`;
  }
  return roundCards.map((played) => `
      <div class="played-card">
        <span>${escapeHtml(playerName(match, played.PlayerID))}</span>
        ${played.FaceDown ? `<span class="card-back small"></span>` : renderCard(played.Card, "small")}
      </div>
    `).join("");
}

function renderTrickTrack(match: MatchSnapshot): string {
  return Array.from({ length: 3 }, (_, index) => {
    const result = match.CurrentHand.TrickResults[index];
    let label = t("game_trick_pending");
    let className = "trick-pill";
    if (result === -1) {
      label = t("game_trick_draw");
      className += " trick-pill-draw";
    } else if (result === 0 || result === 1) {
      label = t("game_trick_team", result + 1);
      className += ` trick-pill-team-${result + 1}`;
    }
    return `<span class="${className}">${escapeHtml(t("game_trick_label", index + 1))} · ${escapeHtml(label)}</span>`;
  }).join("");
}

function renderPlayableCard(card: Card, index: number, canPlay: boolean, canFaceDown: boolean): string {
  if (!canPlay) {
    return `<div class="play-card lock-card">${renderCard(card)}<span class="card-caption">${escapeHtml(cardLabel(card))}</span></div>`;
  }

  return `
    <div class="play-card">
      <form data-api-action="play" data-form-id="play-${index}">
        <input type="hidden" name="cardIndex" value="${index}">
        <button class="card-button" type="submit"${busyAttr(`play-${index}`)}>${renderCard(card)}</button>
      </form>
      <div class="play-card-actions">
        <span class="card-caption">${escapeHtml(cardLabel(card))}</span>
        ${canFaceDown ? `<form data-api-action="play" data-form-id="play-down-${index}"><input type="hidden" name="cardIndex" value="${index}"><input type="hidden" name="faceDown" value="true"><button class="ghost-button" type="submit"${busyAttr(`play-down-${index}`)}>${buttonLabel(`play-down-${index}`, t("game_face_down"))}</button></form>` : ""}
      </div>
    </div>
  `;
}

function renderTurnControls(canRaise: boolean, matchFinished: boolean): string {
  if (matchFinished) {
    return `<form data-api-action="reset" data-form-id="reset"><button class="primary-button" type="submit"${busyAttr("reset")}>${buttonLabel("reset", t("game_play_again"))}</button></form>`;
  }

  return `
    <div class="control-group">
      ${canRaise ? `<form data-api-action="truco" data-form-id="truco"><button class="primary-button" type="submit"${busyAttr("truco")}>${buttonLabel("truco", t("game_truco"))}</button></form>` : ""}
    </div>
  `;
}

function renderRespondControls(canRaise: boolean, canAccept: boolean, canRefuse: boolean, pendingTo: number): string {
  return `
    <div class="control-group">
      ${canRaise ? `<form data-api-action="truco" data-form-id="raise"><button class="primary-button" type="submit"${busyAttr("raise")}>${buttonLabel("raise", `${t("game_raise")} ${raiseLabel(nextStake(pendingTo))}`)}</button></form>` : ""}
      ${canAccept ? `<form data-api-action="accept" data-form-id="accept"><button class="secondary-button" type="submit"${busyAttr("accept")}>${buttonLabel("accept", t("game_accept"))}</button></form>` : ""}
      ${canRefuse ? `<form data-api-action="refuse" data-form-id="refuse"><button class="ghost-button danger" type="submit"${busyAttr("refuse")}>${buttonLabel("refuse", t("game_refuse"))}</button></form>` : ""}
    </div>
  `;
}

function renderNetworkPanel(bundle: SnapshotBundle): string {
  const network = bundle.connection.network;
  return `
    <article class="surface-card telemetry-card">
      <div class="card-head">
        <h3>${escapeHtml(t("game_network"))}</h3>
      </div>
      <div class="telemetry-grid">
        ${renderMetric(t("connection_status"), bundle.connection.status)}
        ${renderMetric(t("connection_mode"), bundle.connection.is_online ? t("connection_online") : t("connection_offline"))}
        ${renderMetric(t("connection_transport"), network?.transport || "-")}
        ${renderMetric(t("connection_protocol"), protocolLabel(network))}
        ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
        ${bundle.lobby?.role ? renderMetric(t("connection_role"), bundle.lobby.role) : ""}
      </div>
      <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
        <input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}">
        <button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat"))}</button>
      </form>
    </article>
  `;
}

function renderOverlay(match: MatchSnapshot, localTeamId: number): string {
  const youWon = match.WinnerTeam === localTeamId;
  return `
    <div class="overlay-layer">
      <div class="overlay-card">
        <p class="eyebrow">${escapeHtml(t("game_status"))}</p>
        <h3>${escapeHtml(youWon ? t("overlay_win") : t("overlay_loss"))}</h3>
        <p>${escapeHtml(t("overlay_score", teamScore(match, 0), teamScore(match, 1)))}</p>
        <form data-api-action="reset" data-form-id="reset">
          <button class="primary-button" type="submit"${busyAttr("reset")}>${buttonLabel("reset", t("game_play_again"))}</button>
        </form>
      </div>
    </div>
  `;
}

function renderCard(card: Card, size: "tiny" | "small" | "regular" = "regular"): string {
  const red = card.Suit === "Copas" || card.Suit === "Ouros";
  return `
    <span class="card-face card-face-${size}${red ? " card-face-red" : ""}">
      <span class="card-corner">${escapeHtml(card.Rank)}${escapeHtml(suitSymbol(card.Suit))}</span>
      <span class="card-center">${escapeHtml(suitSymbol(card.Suit))}</span>
      <span class="card-corner card-corner-bottom">${escapeHtml(card.Rank)}${escapeHtml(suitSymbol(card.Suit))}</span>
    </span>
  `;
}

function renderMetric(label: string, value: string): string {
  return `<div class="metric"><span>${escapeHtml(label)}</span><strong>${escapeHtml(value)}</strong></div>`;
}

function renderEventFeed(extraLines: string[] = []): string {
  const fromEvents = state.events.map(formatEventLine);
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
  const delay = view === "game" && !isOnlineMode() ? 850 : isOnlineMode() ? 1200 : 0;
  if (delay === 0) {
    return;
  }

  refreshTimer = window.setTimeout(async () => {
    try {
      await syncSnapshot({ withEvents: true });
      state.error = "";
      render();
    } catch (error) {
      state.error = errorMessage(error);
      render();
    }
  }, delay);
}

function syncMeasuredBlocks(): void {
  const blocks = root.querySelectorAll<HTMLElement>("[data-pretext-block]");
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
  if (!value || !navigator.clipboard) {
    return;
  }
  await navigator.clipboard.writeText(value);
  const label = button.textContent || t("invite_copy");
  button.textContent = t("copy_done");
  window.setTimeout(() => {
    button.textContent = label;
  }, 1000);
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

function cardLabel(card: Card): string {
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

function protocolLabel(network?: SnapshotBundle["connection"]["network"]): string {
  if (!network) {
    return "-";
  }
  if (network.negotiated_protocol_version) {
    return `v${network.negotiated_protocol_version}`;
  }
  const versions = Object.values(network.seat_protocol_versions || {}).filter((value) => value > 0);
  if (versions.length === 0) {
    return "-";
  }
  return Array.from(new Set(versions)).sort((a, b) => a - b).map((value) => `v${value}`).join(", ");
}

function seatPositions(match: MatchSnapshot): Map<number, string> {
  const positions = new Map<number, string>();
  const local = match.CurrentPlayerIdx;
  const layout = match.NumPlayers === 2 ? ["bottom", "top"] : ["bottom", "right", "top", "left"];
  for (const player of match.Players) {
    const offset = (player.ID - local + match.NumPlayers) % match.NumPlayers;
    positions.set(player.ID, layout[offset] || "top");
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
