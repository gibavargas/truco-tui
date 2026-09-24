import type { Card, MatchSnapshot, RuntimeEvent, SnapshotBundle } from "../types";
import {
  activeTransportLabel,
  connectionPathLabel,
  fallbackReasonLabel,
  requestedTransportLabel,
  seatProtocolLabel,
} from "./network-copy";

export type GamePanelTab = "pulse" | "network" | "chat";

interface GameScreenParams {
  bundle: SnapshotBundle;
  panelTab: GamePanelTab;
  events: RuntimeEvent[];
  isOnlineMode: boolean;
  t: (key: string, ...args: Array<string | number>) => string;
  escapeHtml: (value: string) => string;
  busyAttr: (formId: string) => string;
  buttonLabel: (formId: string, label: string) => string;
  renderMetric: (label: string, value: string) => string;
  renderEventFeed: (extraLines?: string[]) => string;
  renderCard: (card: Card, size?: "tiny" | "small" | "regular") => string;
  protocolLabel: (network?: SnapshotBundle["connection"]["network"]) => string;
  cardLabel: (card: Card) => string;
  playerName: (match: MatchSnapshot, playerId: number) => string;
  teamScore: (match: MatchSnapshot, team: number) => number;
  localTeam: (match: MatchSnapshot, bundle: SnapshotBundle) => number;
  nextStake: (current: number) => number;
  raiseLabel: (value: number) => string;
  lastTrickCopy: (match: MatchSnapshot) => string;
  seatPositions: (match: MatchSnapshot, bundle: SnapshotBundle) => Map<number, string>;
}

export function renderGameScreen(params: GameScreenParams): string {
  const {
    bundle,
    panelTab,
    events,
    isOnlineMode,
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
  } = params;

  const match = bundle.match;
  if (!match || !match.CurrentHand || !Array.isArray(match.Players) || match.Players.length === 0) {
    return "";
  }

  const localPlayerID = bundle.ui.actions.local_player_id >= 0 ? bundle.ui.actions.local_player_id : match.CurrentPlayerIdx;
  const localPlayer = match.Players.find((player) => player.ID === localPlayerID) || match.Players[0];
  const localTeamID = localTeam(match, bundle);
  const pendingFor = match.PendingRaiseFor;
  const pendingTo = match.PendingRaiseTo || nextStake(match.CurrentHand.Stake);
  const canRespond = bundle.ui.actions.must_respond;
  const canRaise = bundle.ui.actions.can_ask_or_raise;
  const canPlay = bundle.ui.actions.can_play_card;
  const canStartNewHand = canStartNewHandForBundle(bundle, match);
  const isYourTurn = match.TurnPlayer === localPlayerID;
  const feltPlayerClass = `game10-felt-${Math.max(2, Math.min(4, match.Players.length))}`;
  const tableTitle = isOnlineMode ? t("game_title_online") : t("game_title_offline");
  const signal = match.MatchFinished
    ? t("status_match_end")
    : pendingFor === localTeamID
      ? t("status_pending_you", raiseLabel(pendingTo))
      : pendingFor >= 0
        ? t("status_pending_other", playerName(match, match.TurnPlayer), raiseLabel(pendingTo))
        : isYourTurn
          ? t("status_your_turn")
          : t("status_wait_turn", playerName(match, match.TurnPlayer));

  const opponentTeamID = localTeamID === 0 ? 1 : 0;

  return `
    <section class="game10" data-player-count="${match.NumPlayers}" aria-labelledby="game-table-title">
      <!-- Premium Score Board -->
      <article class="surface-card game10-hud">
        <div class="game10-hud-cluster">
          <div class="game10-score-badge${localTeamID === 0 ? " game10-score-badge-friendly" : ""}">
            <span>${escapeHtml(t("team_us"))}</span>
            <strong>${teamScore(match, localTeamID)}</strong>
            <div class="game10-trick-indicators" aria-label="${escapeHtml(t("team_us"))} vazas">
              ${renderTrickIndicators(match, localTeamID, t)}
            </div>
          </div>

          <div class="game10-center-signal">
            <span class="game10-match-round">${escapeHtml(t("game_round"))} ${match.CurrentHand.Round}/3</span>
            <div class="game10-stake-badge">${escapeHtml(pendingFor >= 0 ? `${t("game_pending_raise")} ${raiseLabel(pendingTo)}` : `${t("game_stake")} ${match.CurrentHand.Stake}`)}</div>
          </div>

          <div class="game10-score-badge${localTeamID === 1 ? " game10-score-badge-friendly" : ""}">
            <span>${escapeHtml(t("team_them"))}</span>
            <strong>${teamScore(match, opponentTeamID)}</strong>
            <div class="game10-trick-indicators" aria-label="${escapeHtml(t("team_them"))} vazas">
              ${renderTrickIndicators(match, opponentTeamID, t)}
            </div>
          </div>
        </div>
      </article>

      <div class="game10-grid">
        <article class="surface-card game10-table-wrap">
          <div class="game10-table-head">
            <div>
              <p class="eyebrow">${escapeHtml(t("game_status"))}</p>
              <h3 id="game-table-title">${escapeHtml(tableTitle)}</h3>
            </div>
            <div class="desktop-action-row game10-head-actions">
              <button class="ghost-button" type="button" data-client-action="refresh">${escapeHtml(t("header_resync"))}</button>
              ${canStartNewHand ? `<form data-api-action="newHand" data-form-id="newHand"><button class="ghost-button" type="submit"${busyAttr("newHand")}>${buttonLabel("newHand", t("game_new_hand"))}</button></form>` : ""}
              ${bundle.ui.actions.can_close_session ? `<form data-api-action="${isOnlineMode ? "closeSession" : "reset"}" data-form-id="${isOnlineMode ? "closeSession" : "reset"}"><button class="ghost-button danger" type="submit"${busyAttr(isOnlineMode ? "closeSession" : "reset")}>${buttonLabel(isOnlineMode ? "closeSession" : "reset", isOnlineMode ? t("lobby_leave") : t("game_play_again"))}</button></form>` : ""}
            </div>
          </div>

          <div class="game10-status-ribbon" role="status" aria-live="polite">
            <span>${escapeHtml(heroSignalLabel(match, isYourTurn, canRespond, t))}</span>
            <strong>${escapeHtml(signal)}</strong>
          </div>

          <div class="game10-felt ${feltPlayerClass}">
            <!-- 3D Felt Circular Table Surface -->
            <div class="game10-table-surface">
              <!-- Center area: Deck & Vira -->
              <div class="game10-table-center-deck">
                <div class="game10-card-deck-stack">
                  <div class="card-back small game10-deck-card-1"></div>
                  <div class="card-back small game10-deck-card-2"></div>
                </div>
                <div class="game10-table-vira-card">
                  <span class="game10-vira-label">${escapeHtml(t("game_vira"))}</span>
                  ${renderCard(match.CurrentHand.Vira, "small")}
                </div>
                <div class="game10-manilha-badge">
                  <span class="game10-manilha-label">${escapeHtml(t("game_manilha"))}</span>
                  <strong class="game10-manilha-value">${escapeHtml(match.CurrentHand.Manilha || "-")}</strong>
                </div>
              </div>

              <!-- Played Cards around the deck -->
              <div class="game10-table-played-cards">
                ${renderTablePlayedCards(match, bundle, seatPositions, playerName, renderCard, escapeHtml)}
              </div>
            </div>

            <!-- Player Seats positioned around the table -->
            ${renderPlayerSeats(match, bundle, seatPositions, localTeamID, t, escapeHtml)}

            <div class="game10-bottom-area">
              <div class="game10-action-band">
                ${canRespond ? renderRespondControls(canRaise, bundle.ui.actions.can_accept, bundle.ui.actions.can_refuse, raiseLabel(nextStake(pendingTo)), t, busyAttr, buttonLabel) : renderTurnControls(canRaise, match.MatchFinished, t, busyAttr, buttonLabel, raiseLabel(nextStake(match.CurrentHand.Stake)))}
              </div>
              <div class="game10-hand-stage">
                <div class="game10-hand-head">
                  <div>
                    <p class="eyebrow">${escapeHtml(t("game_hand"))}</p>
                    <h3 id="game-hand-title">${escapeHtml(localPlayer?.Name || t("game_you"))}</h3>
                  </div>
                  <div class="game10-chip-row">
                    ${renderRoleChip(roleBadge(match, localPlayerID), escapeHtml)}
                    ${isYourTurn ? `<span class="game10-turn-badge">${escapeHtml(t("game_turn"))}</span>` : ""}
                  </div>
                </div>
                <div class="game10-hand-row" role="list" aria-labelledby="game-hand-title">
                  ${(localPlayer?.Hand || []).map((card, index) => renderPlayableCard(card, index, canPlay, match.CurrentHand.Round >= 2, t, escapeHtml, busyAttr, buttonLabel, renderCard, cardLabel)).join("")}
                </div>
              </div>
            </div>
          </div>
        </article>

        <aside class="surface-card game10-side">
          <div class="card-head">
            <div>
              <p class="eyebrow">${escapeHtml(t("game_activity"))}</p>
              <h3 id="game-panel-title">${escapeHtml(tabTitle(panelTab, t))}</h3>
            </div>
            <div class="panel-tabs" role="tablist" aria-label="${escapeHtml(t("game_activity"))}">
              ${renderPanelTab("pulse", panelTab, t("game_activity"), escapeHtml)}
              ${renderPanelTab("network", panelTab, t("game_network"), escapeHtml)}
              ${renderPanelTab("chat", panelTab, t("lobby_chat"), escapeHtml)}
            </div>
          </div>
          ${renderGamePanel(panelTab, bundle, events, isOnlineMode, t, escapeHtml, renderMetric, renderEventFeed, protocolLabel, lastTrickCopy, busyAttr, buttonLabel, match, playerName)}
        </aside>
      </div>

      ${match.MatchFinished ? renderOverlay(match, localTeamID, t, escapeHtml, busyAttr, buttonLabel, isOnlineMode) : ""}
    </section>
  `;
}

function canStartNewHandForBundle(bundle: SnapshotBundle, match: MatchSnapshot): boolean {
  return !match.MatchFinished && (bundle.mode === "offline_match" || bundle.mode === "host_match");
}

function renderTrickIndicators(match: MatchSnapshot, teamId: number, t: any): string {
  const results = match.CurrentHand?.TrickResults || [];
  return Array.from({ length: 3 }, (_, index) => {
    const result = results[index];
    let stateClass = "game10-trick-dot-pending";
    let title = t("game_trick_pending");
    if (result === teamId) {
      stateClass = "game10-trick-dot-win";
      title = "Vencedor";
    } else if (result === -1) {
      stateClass = "game10-trick-dot-tie";
      title = t("game_trick_draw");
    } else if (result !== undefined) {
      stateClass = "game10-trick-dot-lose";
      title = "Perdedor";
    }
    return `<span class="game10-trick-dot ${stateClass}" title="${t("game_trick_label", index + 1)}: ${title}"></span>`;
  }).join("");
}

function renderTablePlayedCards(
  match: MatchSnapshot,
  bundle: SnapshotBundle,
  seatPositions: any,
  playerName: any,
  renderCard: any,
  escapeHtml: any,
): string {
  const roundCards = match.CurrentHand?.RoundCards || [];
  const positions = seatPositions(match, bundle);
  if (roundCards.length === 0) {
    return "";
  }
  return roundCards.map((played) => {
    const position = positions.get(played.PlayerID) || "top";
    return `
      <div class="game10-table-played-card game10-played-on-${position} game10-played-card-enter">
        ${played.FaceDown ? `<span class="card-back small"></span>` : renderCard(played.Card, "small")}
        <span class="game10-played-card-owner">${escapeHtml(playerName(match, played.PlayerID).split(" ")[0])}</span>
      </div>
    `;
  }).join("");
}

function renderPlayerSeats(
  match: MatchSnapshot,
  bundle: SnapshotBundle,
  seatPositions: any,
  localTeamID: number,
  t: any,
  escapeHtml: any,
): string {
  const positions = seatPositions(match, bundle);
  return match.Players.map((player) => {
    const position = positions.get(player.ID) || "top";
    const localPlayerID = bundle.ui.actions.local_player_id >= 0 ? bundle.ui.actions.local_player_id : match.CurrentPlayerIdx;
    const isLocal = player.ID === localPlayerID;
    const relation = isLocal
      ? t("game_you")
      : player.Team === localTeamID ? t("game_partner") : t("game_opponent");
    const isTurn = player.ID === match.TurnPlayer;

    return `
      <div class="game10-seat game10-seat-${position}${isTurn ? " game10-seat-turn" : ""}${isLocal ? " game10-seat-local" : ""}">
        <div class="game10-player-avatar">
          <span class="game10-avatar-letter">${escapeHtml(player.Name[0].toUpperCase())}</span>
          ${isTurn ? `<span class="game10-turn-ring"></span>` : ""}
        </div>
        <div class="game10-seat-label">
          <strong>${escapeHtml(player.Name)}</strong>
          <span>${escapeHtml(relation)}${player.CPU ? ` · ${escapeHtml(t("game_cpu"))}` : ""}</span>
        </div>
        <div class="game10-seat-meta">
          ${renderRoleChip(roleBadge(match, player.ID), escapeHtml)}
        </div>
        ${!isLocal ? `
          <div class="player-cards">
            ${player.Hand.map(() => `<span class="card-back tiny"></span>`).join("")}
          </div>
        ` : ""}
      </div>
    `;
  }).join("");
}

function renderPlayableCard(
  card: Card,
  index: number,
  canPlay: boolean,
  canFaceDown: boolean,
  t: (key: string, ...args: Array<string | number>) => string,
  escapeHtml: (value: string) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
  renderCard: (card: Card, size?: "tiny" | "small" | "regular") => string,
  cardLabel: (card: Card) => string,
): string {
  if (!canPlay) {
    return `<div class="game10-hand-card game10-hand-card-locked" role="listitem">${renderCard(card)}<span class="card-caption">${escapeHtml(cardLabel(card))}</span></div>`;
  }

  return `
    <div class="game10-hand-card" role="listitem">
      <form data-api-action="play" data-form-id="play-${index}">
        <input type="hidden" name="cardIndex" value="${index}">
        <button class="card-button game10-card-button" type="submit" aria-label="${escapeHtml(`${t("game_play")} ${cardLabel(card)}`)}"${busyAttr(`play-${index}`)}>${renderCard(card)}</button>
      </form>
      <span class="card-caption">${escapeHtml(cardLabel(card))}</span>
      ${canFaceDown ? `<form data-api-action="play" data-form-id="play-down-${index}"><input type="hidden" name="cardIndex" value="${index}"><input type="hidden" name="faceDown" value="true"><button class="ghost-button game10-face-down" type="submit" aria-label="${escapeHtml(`${t("game_face_down")} ${cardLabel(card)}`)}"${busyAttr(`play-down-${index}`)}>${buttonLabel(`play-down-${index}`, t("game_face_down"))}</button></form>` : ""}
    </div>
  `;
}

function renderTurnControls(
  canRaise: boolean,
  matchFinished: boolean,
  t: (key: string, ...args: Array<string | number>) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
  raiseLabel: string,
): string {
  if (matchFinished) {
    return `<form data-api-action="reset" data-form-id="reset"><button class="primary-button" type="submit"${busyAttr("reset")}>${buttonLabel("reset", t("game_play_again"))}</button></form>`;
  }
  return `
    <div class="game10-control-row">
      ${canRaise ? `<form data-api-action="truco" data-form-id="truco"><button class="primary-button" type="submit"${busyAttr("truco")}>${buttonLabel("truco", raiseLabel)}</button></form>` : ""}
    </div>
  `;
}

function renderRespondControls(
  canRaise: boolean,
  canAccept: boolean,
  canRefuse: boolean,
  nextRaiseLabel: string,
  t: (key: string, ...args: Array<string | number>) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
): string {
  return `
    <div class="game10-control-row">
      ${canAccept ? `<form data-api-action="accept" data-form-id="accept"><button class="secondary-button" type="submit"${busyAttr("accept")}>${buttonLabel("accept", t("game_accept"))}</button></form>` : ""}
      ${canRaise ? `<form data-api-action="truco" data-form-id="raise"><button class="primary-button" type="submit"${busyAttr("raise")}>${buttonLabel("raise", `${t("game_raise")} ${nextRaiseLabel}`)}</button></form>` : ""}
      ${canRefuse ? `<form data-api-action="refuse" data-form-id="refuse"><button class="ghost-button danger" type="submit"${busyAttr("refuse")}>${buttonLabel("refuse", t("game_refuse"))}</button></form>` : ""}
    </div>
  `;
}

function renderGamePanel(
  panelTab: GamePanelTab,
  bundle: SnapshotBundle,
  events: RuntimeEvent[],
  isOnlineMode: boolean,
  t: (key: string, ...args: Array<string | number>) => string,
  escapeHtml: (value: string) => string,
  renderMetric: (label: string, value: string) => string,
  renderEventFeed: (extraLines?: string[]) => string,
  protocolLabel: (network?: SnapshotBundle["connection"]["network"]) => string,
  lastTrickCopy: (match: MatchSnapshot) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
  match: MatchSnapshot,
  playerName: (match: MatchSnapshot, playerId: number) => string,
): string {
  const network = bundle.connection.network;
  const fallback = fallbackReasonLabel(network, t);
  const seatProtocols = seatProtocolLabel(network);
  switch (panelTab) {
    case "network":
      return `
        <section class="game10-panel panel-region" id="game-panel-network" role="tabpanel" tabindex="0" aria-labelledby="game-tab-network" aria-label="${escapeHtml(t("game_network"))}">
          <div class="telemetry-grid">
            ${renderMetric(t("connection_status"), bundle.connection.status)}
            ${renderMetric(t("connection_mode"), bundle.connection.is_online ? t("connection_online") : t("connection_offline"))}
            ${renderMetric(t("connection_transport"), activeTransportLabel(network, t))}
            ${renderMetric(t("connection_requested_transport"), requestedTransportLabel(network, t))}
            ${renderMetric(t("connection_path"), connectionPathLabel(network, t))}
            ${renderMetric(t("connection_protocol"), protocolLabel(network))}
            ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
            ${bundle.lobby?.role ? renderMetric(t("connection_role"), bundle.lobby.role) : ""}
            ${seatProtocols ? renderMetric(t("connection_protocol_seats"), seatProtocols) : ""}
          </div>
          ${renderSeatStrip(bundle.ui.lobby_slots || [], escapeHtml, t, busyAttr, buttonLabel)}
          ${bundle.connection.last_error ? `<div class="lobby10-inline-error" role="alert">${escapeHtml(`${bundle.connection.last_error.code}: ${bundle.connection.last_error.message}`)}</div>` : ""}
          ${fallback ? `<p class="supporting-copy network-note">${escapeHtml(`${t("connection_fallback")}: ${fallback}`)}</p>` : ""}
          ${latestFailoverSignal(events, t) ? `<p class="supporting-copy">${escapeHtml(latestFailoverSignal(events, t) || "")}</p>` : ""}
        </section>
      `;
    case "chat":
      return `
        <section class="game10-panel panel-region" id="game-panel-chat" role="tabpanel" tabindex="0" aria-labelledby="game-tab-chat" aria-label="${escapeHtml(t("lobby_chat"))}">
          <pre class="event-feed compact" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed((match.Logs || []).slice(-4)))}</pre>
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat"><input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}"><button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat"))}</button></form>
        </section>
      `;
    default:
      return `
        <section class="game10-panel panel-region" id="game-panel-pulse" role="tabpanel" tabindex="0" aria-labelledby="game-tab-pulse" aria-label="${escapeHtml(t("game_overview"))}">
          <div class="game10-pulse-stack">
            <div class="game10-pulse-card">
              <span>${escapeHtml(t("game_last_trick"))}</span>
              <strong>${escapeHtml(lastTrickCopy(match))}</strong>
            </div>
            <div class="game10-pulse-card">
              <span>${escapeHtml(t("game_player_to_move"))}</span>
              <strong>${escapeHtml(playerName(match, match.TurnPlayer))}</strong>
            </div>
          </div>
          <pre class="event-feed compact" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed((match.Logs || []).slice(-6)))}</pre>
        </section>
      `;
  }
}

function renderSeatStrip(
  slots: SnapshotBundle["ui"]["lobby_slots"],
  escapeHtml: (value: string) => string,
  t: (key: string, ...args: Array<string | number>) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
): string {
  if (slots.length === 0) {
    return "";
  }
  return `
    <div class="game10-seat-strip" role="list" aria-label="${escapeHtml(t("lobby_slots"))}">
      ${slots.map((slot) => `
        <div class="game10-seat-pill${slot.is_local ? " game10-seat-pill-local" : ""}" role="listitem">
          <strong>${escapeHtml(slot.name || t("slot_empty"))}</strong>
          <span>${escapeHtml(slot.is_connected ? t("slot_online") : t("slot_offline"))}</span>
          ${slot.can_vote_host ? `<form data-api-action="sendHostVote" data-form-id="match-sendHostVote-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button" type="submit"${busyAttr(`match-sendHostVote-${slot.seat}`)}>${buttonLabel(`match-sendHostVote-${slot.seat}`, t("action_vote_host"))}</button></form>` : ""}
          ${slot.can_request_replacement ? `<form data-api-action="requestReplacementInvite" data-form-id="match-replacement-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="secondary-button strong" type="submit"${busyAttr(`match-replacement-${slot.seat}`)}>${buttonLabel(`match-replacement-${slot.seat}`, t("action_replacement_invite"))}</button></form>` : ""}
        </div>
      `).join("")}
    </div>
  `;
}

function renderPanelTab(value: GamePanelTab, active: GamePanelTab, label: string, escapeHtml: (value: string) => string): string {
  return `<button class="panel-tab${active === value ? " panel-tab-active" : ""}" type="button" id="game-tab-${value}" role="tab" tabindex="${active === value ? "0" : "-1"}" aria-selected="${active === value ? "true" : "false"}" aria-controls="game-panel-${value}" data-panel-tab="game:${value}" data-focus-key="game-tab:${value}">${escapeHtml(label)}</button>`;
}

function tabTitle(tab: GamePanelTab, t: (key: string, ...args: Array<string | number>) => string): string {
  switch (tab) {
    case "network":
      return t("game_network");
    case "chat":
      return t("lobby_chat");
    default:
      return t("game_overview");
  }
}

function latestFailoverSignal(events: RuntimeEvent[], t: (key: string, ...args: Array<string | number>) => string): string {
  const recent = [...events].reverse().find((event) => event.kind === "failover_promoted" || event.kind === "failover_rejoined");
  if (!recent) {
    return "";
  }
  return recent.kind === "failover_promoted" ? t("signal_failover_promoted") : t("signal_failover_rejoined");
}

function heroSignalLabel(
  match: MatchSnapshot,
  isYourTurn: boolean,
  canRespond: boolean,
  t: (key: string, ...args: Array<string | number>) => string,
): string {
  if (match.MatchFinished) {
    return t("status_match_end");
  }
  if (canRespond) {
    return t("game_raise");
  }
  return isYourTurn ? t("status_your_turn") : t("game_wait");
}

function roleBadge(match: MatchSnapshot, playerID: number): string {
  const dealer = match.CurrentHand?.Dealer ?? -1;
  const total = match.NumPlayers || 2;
  if (playerID === dealer) {
    return "🃏";
  }
  if (playerID === (dealer + 1) % total) {
    return "✋";
  }
  if (total === 4 && playerID === (dealer + total - 1) % total) {
    return "🦶";
  }
  return "";
}

function renderRoleChip(value: string, escapeHtml: (value: string) => string): string {
  if (!value) {
    return "";
  }
  return `<span class="game10-role-chip">${escapeHtml(value)}</span>`;
}

function renderOverlay(
  match: MatchSnapshot,
  localTeamID: number,
  t: (key: string, ...args: Array<string | number>) => string,
  escapeHtml: (value: string) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
  isOnlineMode: boolean,
): string {
  const youWon = match.WinnerTeam === localTeamID;
  return `
    <div class="overlay-layer">
      <div class="overlay-card">
        <p class="eyebrow">${escapeHtml(t("game_status"))}</p>
        <h3>${escapeHtml(youWon ? t("overlay_win") : t("overlay_loss"))}</h3>
        <p>${escapeHtml(t("overlay_score", String(match.MatchPoints["0"] || 0), String(match.MatchPoints["1"] || 0)))}</p>
        <form data-api-action="${isOnlineMode ? "closeSession" : "reset"}" data-form-id="${isOnlineMode ? "closeSession" : "reset"}">
          <button class="primary-button" type="submit"${busyAttr(isOnlineMode ? "closeSession" : "reset")}>${buttonLabel(isOnlineMode ? "closeSession" : "reset", isOnlineMode ? t("lobby_leave") : t("game_play_again"))}</button>
        </form>
      </div>
    </div>
  `;
}
