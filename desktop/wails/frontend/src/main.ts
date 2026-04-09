import { layout, prepare } from "@chenglou/pretext";
import { translate } from "./i18n";
import { renderGameScreen } from "./ui/game-screen";
import { renderLobbyScreen } from "./ui/lobby-screen";
import { renderSetupScreen } from "./ui/setup-screen";
import {
  expectedModesForAction,
  recoveryStateForBundle,
  shouldApplyIncomingSequence,
  viewForMode,
} from "./runtime-state";
import type { ViewName } from "./runtime-state";
import type { GamePanelTab } from "./ui/game-screen";
import type { LobbyPanelTab } from "./ui/lobby-screen";
import type {
  Card,
  LocaleCode,
  LobbySlotState,
  MatchSnapshot,
  RuntimeEvent,
  RuntimeUpdate,
  SnapshotBundle,
} from "./types";
import { copyText, invoke, onRuntimeUpdate, snapshot } from "./wails";

type ClientAction = "refresh" | "toggle-diagnostics";
type MutationExpectation = { modes: string[]; timeoutLabel: string };
type RefreshState = "idle" | "ok" | "error" | "stale";

interface AppState {
  locale: LocaleCode;
  bundle: SnapshotBundle | null;
  events: RuntimeEvent[];
  playerName: string;
  relayURL: string;
  transportMode: string;
  error: string;
  busyForm: string;
  initialized: boolean;
  diagnosticsOpen: boolean;
  lobbyPanelTab: LobbyPanelTab;
  gamePanelTab: GamePanelTab;
  pendingAction: string;
  lastSubmittedAction: string;
  lastExpectedModes: string[];
  lastSeenSequence: number;
  lastMode: string;
  lastRefreshState: RefreshState;
  lastRefreshError: string;
  lastRenderError: string;
}

const LOCALE_KEY = "truco-wails-locale";
const PLAYER_KEY = "truco-wails-player-name";
const RELAY_KEY = "truco-wails-relay-url";
const TRANSPORT_KEY = "truco-wails-transport-mode";
const EVENT_LIMIT = 80;
const ACTION_TIMEOUT_MS = 4000;
const SNAPSHOT_TIMEOUT_MS = 2500;
const rootElement = document.querySelector<HTMLElement>("#app");

if (!rootElement) {
  throw new Error("missing #app root");
}

const root = rootElement;
const preparedCache = new Map<string, ReturnType<typeof prepare>>();
let resizeTimer = 0;

const state: AppState = {
  locale: readLocale(localStorage.getItem(LOCALE_KEY)),
  bundle: null,
  events: [],
  playerName: localStorage.getItem(PLAYER_KEY) || "",
  relayURL: localStorage.getItem(RELAY_KEY) || "",
  transportMode: localStorage.getItem(TRANSPORT_KEY) || "",
  error: "",
  busyForm: "",
  initialized: false,
  diagnosticsOpen: false,
  lobbyPanelTab: "pulse",
  gamePanelTab: "pulse",
  pendingAction: "",
  lastSubmittedAction: "",
  lastExpectedModes: [],
  lastSeenSequence: 0,
  lastMode: "idle",
  lastRefreshState: "idle",
  lastRefreshError: "",
  lastRenderError: "",
};

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

  const clientAction = target.closest<HTMLElement>("[data-client-action]");
  if (clientAction) {
    event.preventDefault();
    void runClientAction(clientAction.dataset.clientAction as ClientAction | undefined);
    return;
  }

  const panelTab = target.closest<HTMLElement>("[data-panel-tab]");
  if (panelTab) {
    event.preventDefault();
    setPanelTab(panelTab.dataset.panelTab || "");
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
    await onRuntimeUpdate((update) => {
      applyRuntimeUpdate(update);
      if (!update.bundle.connection.last_error) {
        state.error = "";
      }
      render();
    });
    await refreshSnapshot();
    state.initialized = true;
    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  }
  render();
}

async function refreshSnapshot(): Promise<void> {
  const bundle = await snapshot();
  applyBundle(bundle, "snapshot");
}

function applyRuntimeUpdate(update: RuntimeUpdate): void {
  if (!applyBundle(update.bundle, "event")) {
    return;
  }

  if (update.events && update.events.length > 0) {
    state.events = [...state.events, ...update.events].slice(-EVENT_LIMIT);
  }
}

async function applyLocale(locale: LocaleCode): Promise<void> {
  state.locale = locale;
  localStorage.setItem(LOCALE_KEY, locale);
  document.documentElement.lang = locale === "en-US" ? "en" : "pt-BR";
  render();

  try {
    const result = await invoke("setLocale", { locale });
    if (result?.message) {
      await refreshSnapshot();
      throw new Error(result.message);
    }
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
  const expectation = mutationExpectation(action);
  state.busyForm = formId;
  state.pendingAction = action;
  state.lastSubmittedAction = action;
  state.lastExpectedModes = expectation.modes;
  state.lastRenderError = "";
  state.lastRefreshState = "idle";
  state.lastRefreshError = "";
  state.error = "";
  render();

  try {
    persistInputs(payload);
    const result = await withTimeout(
      invoke(action, payload),
      ACTION_TIMEOUT_MS,
      t("status_action_timeout", expectation.timeoutLabel),
    );
    if (result?.message) {
      await reconcileSnapshot(action, expectation, true);
      throw new Error(result.message);
    }
    await reconcileSnapshot(action, expectation, false);
    state.error = "";
  } catch (error) {
    state.error = errorMessage(error);
  } finally {
    state.busyForm = "";
    state.pendingAction = "";
    render();
  }
}

async function runClientAction(action: ClientAction | undefined): Promise<void> {
  if (!action) {
    return;
  }
  switch (action) {
    case "refresh":
      try {
        await refreshSnapshot();
        state.error = "";
      } catch (error) {
        state.error = errorMessage(error);
      }
      render();
      return;
    case "toggle-diagnostics":
      state.diagnosticsOpen = !state.diagnosticsOpen;
      render();
      return;
    default:
      return;
  }
}

function setPanelTab(raw: string): void {
  const [surface, value] = raw.split(":");
  if (surface === "lobby" && (value === "pulse" || value === "network" || value === "chat")) {
    state.lobbyPanelTab = value;
    render();
    return;
  }
  if (surface === "game" && (value === "pulse" || value === "network" || value === "chat")) {
    state.gamePanelTab = value;
    render();
  }
}

function currentView(): ViewName {
  return viewForMode(state.bundle?.mode || "idle");
}

function isOnlineMode(): boolean {
  const mode = state.bundle?.mode || "";
  return mode.startsWith("host_") || mode.startsWith("client_");
}

function render(): void {
  root.innerHTML = renderApp();
  syncMeasuredBlocks();
}

function renderApp(): string {
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
          <div class="hero-tools">
            <form class="locale-card">
              <label for="locale-select">${escapeHtml(t("locale_label"))}</label>
              <select id="locale-select" name="locale">
                ${localeOptions(state.locale)}
              </select>
            </form>
            <div class="desktop-action-row">
              <button class="ghost-button" type="button" data-client-action="refresh">${escapeHtml(t("header_resync"))}</button>
              <button class="ghost-button strong" type="button" data-client-action="toggle-diagnostics">
                ${escapeHtml(state.diagnosticsOpen ? t("header_hide_diagnostics") : t("header_diagnostics"))}
              </button>
            </div>
          </div>
        </header>
        ${renderRuntimeBanner()}
        ${renderSafeView()}
        ${renderDiagnostics()}
      </main>
    </div>
  `;
}

function renderRuntimeBanner(): string {
  const classes = ["runtime-banner"];
  let message = "";
  const eventSignal = recentEventSignal();

  if (state.pendingAction) {
    classes.push("runtime-banner-info");
    message = t("status_action_pending", describeAction(state.pendingAction));
  } else if (state.lastRenderError) {
    classes.push("runtime-banner-warning");
    message = `${t("status_render_recovery")} ${state.lastRenderError}`;
  } else if (activeError()) {
    classes.push("runtime-banner-danger");
    message = activeError() || "";
  } else if (state.lastRefreshState === "stale") {
    classes.push("runtime-banner-warning");
    message = t("status_refresh_stale", formatExpectedModes(state.lastExpectedModes));
  } else if (state.lastRefreshState === "error" && state.lastRefreshError) {
    classes.push("runtime-banner-warning");
    message = state.lastRefreshError;
  } else if (eventSignal) {
    classes.push("runtime-banner-info");
    message = eventSignal;
  }

  if (!message) {
    return "";
  }

  return `
    <section class="${classes.join(" ")}" data-pretext-block="lock-height">
      <div class="runtime-banner-copy">${escapeHtml(message)}</div>
      <div class="runtime-banner-actions">
        <button class="ghost-button strong" type="button" data-client-action="refresh">${escapeHtml(t("header_resync"))}</button>
        <button class="ghost-button" type="button" data-client-action="toggle-diagnostics">
          ${escapeHtml(state.diagnosticsOpen ? t("header_hide_diagnostics") : t("header_diagnostics"))}
        </button>
      </div>
    </section>
  `;
}

function recentEventSignal(): string {
  const recent = [...state.events].reverse().find((event) =>
    event.kind === "failover_promoted" ||
    event.kind === "failover_rejoined" ||
    event.kind === "replacement_invite",
  );
  if (!recent) {
    return "";
  }
  switch (recent.kind) {
    case "failover_promoted":
      return t("signal_failover_promoted");
    case "failover_rejoined":
      return t("signal_failover_rejoined");
    case "replacement_invite":
      return t("signal_replacement_ready");
    default:
      return "";
  }
}

function renderSafeView(): string {
  try {
    return renderView();
  } catch (error) {
    state.lastRenderError = errorMessage(error);
    return renderRenderRecoveryCard();
  }
}

function renderView(): string {
  if (!state.initialized && !state.bundle && !activeError()) {
    return `<section class="surface-card loading-card"><div class="loading-pip"></div><strong>${escapeHtml(t("button_busy"))}</strong></section>`;
  }

  const recoveryState = recoveryStateForBundle(state.bundle);
  if (recoveryState === "waiting_lobby") {
    return renderStateRecoveryCard(t("lobby_title"), t("status_waiting_lobby"));
  }
  if (recoveryState === "waiting_match") {
    return renderStateRecoveryCard(t("game_title_offline"), t("status_waiting_match"));
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
  return renderSetupScreen({
    locale: state.locale,
    playerName: state.playerName,
    relayURL: state.relayURL,
    transportMode: state.transportMode,
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    transportOptions,
  });
}

function renderLobby(): string {
  const bundle = requireBundle();
  return renderLobbyScreen({
    bundle,
    panelTab: state.lobbyPanelTab,
    events: state.events,
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    renderMetric,
    renderEventFeed,
    protocolLabel,
  });
}

function renderGame(): string {
  const bundle = requireBundle();
  return renderGameScreen({
    bundle,
    panelTab: state.gamePanelTab,
    events: state.events,
    isOnlineMode: isOnlineMode(),
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    renderMetric,
    renderEventFeed,
    renderCard,
    protocolLabel,
    cardLabel,
    playerName,
    teamScore,
    localTeam,
    nextStake,
    raiseLabel,
    lastTrickCopy,
    seatPositions,
  });
}

function renderPlayers(match: MatchSnapshot, bundle: SnapshotBundle): string {
  const localPlayerID = bundle.ui.actions.local_player_id >= 0 ? bundle.ui.actions.local_player_id : match.CurrentPlayerIdx;
  const positions = seatPositions(match, bundle);
  const localTeamId = localTeam(match, bundle);
  return (match.Players || [])
    .filter((player) => player.ID !== localPlayerID)
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
            ${(player.Hand || []).map(() => `<span class="card-back tiny"></span>`).join("")}
          </div>
        </div>
      `;
    })
    .join("");
}

function renderRoundCards(match: MatchSnapshot): string {
  const roundCards = match.CurrentHand?.RoundCards || [];
  if (roundCards.length === 0) {
    return `<div class="round-card-placeholder">${escapeHtml(t("game_table_waiting"))}</div>`;
  }
  return roundCards
    .map((played) => `
      <div class="played-card">
        <span>${escapeHtml(playerName(match, played.PlayerID))}</span>
        ${played.FaceDown ? `<span class="card-back small"></span>` : renderCard(played.Card, "small")}
      </div>
    `)
    .join("");
}

function renderTrickTrack(match: MatchSnapshot): string {
  const trickResults = match.CurrentHand?.TrickResults || [];
  return Array.from({ length: 3 }, (_, index) => {
    const result = trickResults[index];
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

function renderStateRecoveryCard(title: string, copy: string): string {
  return `
    <section class="surface-card loading-card">
      <div class="card-head">
        <div>
          <p class="eyebrow">${escapeHtml(t("game_runtime_stale_title"))}</p>
          <h3>${escapeHtml(title)}</h3>
        </div>
      </div>
      <p class="supporting-copy" data-pretext-block="lock-height">${escapeHtml(copy)}</p>
      <div class="desktop-action-row">
        <button class="ghost-button strong" type="button" data-client-action="refresh">${escapeHtml(t("header_resync"))}</button>
        <button class="ghost-button" type="button" data-client-action="toggle-diagnostics">${escapeHtml(t("header_diagnostics"))}</button>
      </div>
    </section>
  `;
}

function renderRenderRecoveryCard(): string {
  return renderStateRecoveryCard(t("game_runtime_stale_title"), t("status_render_recovery_copy"));
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

function renderDiagnostics(): string {
  if (!state.diagnosticsOpen || !state.bundle) {
    return "";
  }
  const bundle = state.bundle;
  const error = bundle.connection.last_error;
  const eventLog = bundle.diagnostics.event_log || [];

  return `
    <section class="surface-card diagnostics-card">
      <div class="card-head">
        <div>
          <p class="eyebrow">${escapeHtml(t("header_diagnostics"))}</p>
          <h3>${escapeHtml(t("diagnostics_title"))}</h3>
        </div>
        <div class="desktop-action-row">
          <button class="ghost-button" type="button" data-client-action="refresh">${escapeHtml(t("header_resync"))}</button>
          <form data-api-action="tick" data-form-id="tick">
            <input type="hidden" name="maxSteps" value="12">
            <button class="ghost-button strong" type="submit"${busyAttr("tick")}>${buttonLabel("tick", t("diagnostics_force_tick"))}</button>
          </form>
        </div>
      </div>
      <div class="diagnostics-grid">
        ${renderMetric(t("diagnostics_versions"), `core ${bundle.versions.core_api_version} · protocol ${bundle.versions.protocol_version} · schema ${bundle.versions.snapshot_schema_version}`)}
        ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
        ${renderMetric(t("connection_status"), bundle.connection.status)}
        ${renderMetric(t("connection_transport"), bundle.connection.network?.transport || "-")}
        ${renderMetric(t("diagnostics_mode"), state.lastMode || bundle.mode || "idle")}
        ${renderMetric(t("diagnostics_sequence"), String(state.lastSeenSequence))}
        ${renderMetric(t("diagnostics_last_action"), state.lastSubmittedAction ? describeAction(state.lastSubmittedAction) : "-")}
        ${renderMetric(t("diagnostics_refresh"), state.lastRefreshState)}
        ${error ? renderMetric(t("event_error"), `${error.code}: ${error.message}`) : ""}
        ${state.lastRefreshError ? renderMetric(t("diagnostics_refresh_error"), state.lastRefreshError) : ""}
        ${state.lastRenderError ? renderMetric(t("diagnostics_render_error"), state.lastRenderError) : ""}
      </div>
      <div class="diagnostics-log-shell">
        <strong>${escapeHtml(t("diagnostics_event_log"))}</strong>
        <pre class="diagnostics-log" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(eventLog.length > 0 ? eventLog.slice(-16).join("\n") : t("diagnostics_none"))}</pre>
      </div>
    </section>
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

function applyBundle(bundle: SnapshotBundle, source: "snapshot" | "event"): boolean {
  const sequence = bundle.connection?.last_event_sequence || 0;
  if (!shouldApplyIncomingSequence(state.lastSeenSequence, sequence, source)) {
    return false;
  }

  const previousMode = state.lastMode;
  state.bundle = bundle;
  state.locale = bundle.locale || state.locale;
  state.lastSeenSequence = Math.max(state.lastSeenSequence, sequence);
  state.lastMode = bundle.mode || "idle";
  document.documentElement.lang = state.locale === "en-US" ? "en" : "pt-BR";
  localStorage.setItem(LOCALE_KEY, state.locale);

  if (state.lastMode !== previousMode) {
    if (state.lastMode === "idle" || state.lastMode === "host_lobby" || state.lastMode === "client_lobby") {
      state.lobbyPanelTab = "pulse";
    }
    if (state.lastMode === "idle" || state.lastMode === "offline_match" || state.lastMode === "host_match" || state.lastMode === "client_match") {
      state.gamePanelTab = "pulse";
    }
  }

  if (state.lastMode === "idle") {
    state.pendingAction = "";
  }

  return true;
}

async function reconcileSnapshot(
  action: string,
  expectation: MutationExpectation,
  isErrorRecovery: boolean,
): Promise<void> {
  try {
    await withTimeout(refreshSnapshot(), SNAPSHOT_TIMEOUT_MS, t("status_snapshot_timeout"));
    state.lastRefreshState = "ok";
    state.lastRefreshError = "";
  } catch (error) {
    state.lastRefreshState = "error";
    state.lastRefreshError = t("status_refresh_failed", describeAction(action), errorMessage(error));
    throw new Error(state.lastRefreshError);
  }

  if (isErrorRecovery) {
    return;
  }

  const mode = state.bundle?.mode || "idle";
  if (expectation.modes.length > 0 && !expectation.modes.includes(mode)) {
    state.lastRefreshState = "stale";
    throw new Error(t("status_transition_failed", describeAction(action), formatExpectedModes(expectation.modes), mode));
  }
}

function mutationExpectation(action: string): MutationExpectation {
  return { modes: expectedModesForAction(action), timeoutLabel: describeAction(action) };
}

function describeAction(action: string): string {
  switch (action) {
    case "startGame":
      return t("setup_offline_title");
    case "startOnlineHost":
      return t("setup_host");
    case "joinOnline":
      return t("setup_join");
    case "startOnlineMatch":
      return t("lobby_start");
    case "sendChat":
      return t("lobby_chat");
    case "closeSession":
    case "reset":
      return t("lobby_leave");
    default:
      return action;
  }
}

function formatExpectedModes(modes: string[]): string {
  if (modes.length === 0) {
    return "-";
  }
  return modes.join(", ");
}

async function withTimeout<T>(promise: Promise<T>, timeoutMs: number, message: string): Promise<T> {
  let timer = 0;
  try {
    return await Promise.race([
      promise,
      new Promise<T>((_, reject) => {
        timer = window.setTimeout(() => reject(new Error(message)), timeoutMs);
      }),
    ]);
  } finally {
    window.clearTimeout(timer);
  }
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

function transportOptions(active: string): string {
  return [
    ["", t("transport_auto")],
    ["tcp_tls", t("transport_direct")],
    ["relay_quic_v2", t("transport_relay")],
  ]
    .map(([value, label]) => `<option value="${value}"${active === value ? " selected" : ""}>${escapeHtml(label)}</option>`)
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
  if (typeof payload.transport_mode === "string") {
    state.transportMode = payload.transport_mode;
    localStorage.setItem(TRANSPORT_KEY, state.transportMode);
  }
}

async function copyInvite(button: HTMLElement): Promise<void> {
  const value = button.dataset.copyText || "";
  if (!value) {
    return;
  }
  await copyText(value);
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

function activeError(): string | null {
  if (state.error) {
    return state.error;
  }
  const runtimeError = state.bundle?.connection.last_error;
  if (!runtimeError?.message) {
    return null;
  }
  return runtimeError.code ? `${runtimeError.code}: ${runtimeError.message}` : runtimeError.message;
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

function localTeam(match: MatchSnapshot, bundle: SnapshotBundle): number {
  const localPlayerID = bundle.ui.actions.local_player_id >= 0 ? bundle.ui.actions.local_player_id : match.CurrentPlayerIdx;
  return match.Players.find((player) => player.ID === localPlayerID)?.Team || 0;
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

function seatPositions(match: MatchSnapshot, bundle: SnapshotBundle): Map<number, string> {
  const positions = new Map<number, string>();
  const local = bundle.ui.actions.local_player_id >= 0 ? bundle.ui.actions.local_player_id : match.CurrentPlayerIdx;
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
