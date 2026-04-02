import type { LobbySlotState, RuntimeEvent, SnapshotBundle } from "../types";

export type LobbyPanelTab = "pulse" | "network" | "chat";

interface LobbyScreenParams {
  bundle: SnapshotBundle;
  panelTab: LobbyPanelTab;
  events: RuntimeEvent[];
  t: (key: string, ...args: Array<string | number>) => string;
  escapeHtml: (value: string) => string;
  busyAttr: (formId: string) => string;
  buttonLabel: (formId: string, label: string) => string;
  renderMetric: (label: string, value: string) => string;
  renderEventFeed: (extraLines?: string[]) => string;
  protocolLabel: (network?: SnapshotBundle["connection"]["network"]) => string;
}

export function renderLobbyScreen(params: LobbyScreenParams): string {
  const { bundle, panelTab, events, t, escapeHtml, busyAttr, buttonLabel, renderMetric, renderEventFeed, protocolLabel } = params;
  const lobby = bundle.lobby;
  if (!lobby) {
    return "";
  }
  const slots = bundle.ui.lobby_slots || [];
  const network = bundle.connection.network;
  const invite = lobby.invite_key || "";
  const occupiedSeats = slots.filter((slot) => !slot.is_empty).length;
  const isHost = bundle.connection.is_host;
  const failoverSignal = latestFailoverSignal(events, t);

  return `
    <section class="lobby10">
      <article class="surface-card lobby10-lead">
        <div class="lobby10-banner">
          <div>
            <p class="eyebrow">${escapeHtml(isHost ? t("lobby_host_headline") : t("lobby_join_headline"))}</p>
            <h2>${escapeHtml(t("lobby_title"))}</h2>
            <p class="supporting-copy" data-pretext-block="lock-height">${escapeHtml(failoverSignal || t("invite_hint"))}</p>
          </div>
          <div class="lobby10-banner-actions">
            ${isHost ? `<form data-api-action="startOnlineMatch" data-form-id="startOnlineMatch"><button class="primary-button" type="submit"${busyAttr("startOnlineMatch")}>${buttonLabel("startOnlineMatch", t("lobby_start"))}</button></form>` : ""}
            <form data-api-action="closeSession" data-form-id="closeSession">
              <button class="ghost-button danger" type="submit"${busyAttr("closeSession")}>${buttonLabel("closeSession", t("lobby_leave"))}</button>
            </form>
          </div>
        </div>
        <div class="lobby10-invite-band">
          <div class="lobby10-invite">
            <span>${escapeHtml(t("setup_invite"))}</span>
            <code>${escapeHtml(invite || "----")}</code>
          </div>
          ${invite ? `<button type="button" class="ghost-button strong" data-copy-text="${escapeHtml(invite)}">${escapeHtml(t("invite_copy"))}</button>` : ""}
          <div class="lobby10-network-strip">
            ${renderMetric(t("connection_transport"), network?.transport || "-")}
            ${renderMetric(t("connection_protocol"), protocolLabel(network))}
            ${renderMetric(t("connection_status"), bundle.connection.status)}
          </div>
        </div>
      </article>

      <div class="lobby10-grid">
        <article class="surface-card lobby10-seats">
          <div class="card-head">
            <div>
              <p class="eyebrow">${escapeHtml(t("lobby_slots"))}</p>
              <h3>${escapeHtml(t("lobby_slots"))}</h3>
            </div>
            <span class="section-pill">${occupiedSeats}/${lobby.num_players || slots.length}</span>
          </div>
          <div class="lobby10-seat-grid">
            ${slots.map((slot) => renderLobbySeat(slot, t, escapeHtml, busyAttr, buttonLabel)).join("")}
          </div>
        </article>

        <aside class="surface-card lobby10-side">
          <div class="card-head">
            <div>
              <p class="eyebrow">${escapeHtml(t("lobby_events"))}</p>
              <h3>${escapeHtml(tabTitle(panelTab, t))}</h3>
            </div>
            <div class="panel-tabs">
              ${renderPanelTab("pulse", panelTab, t("lobby_events"), escapeHtml)}
              ${renderPanelTab("network", panelTab, t("game_network"), escapeHtml)}
              ${renderPanelTab("chat", panelTab, t("lobby_chat"), escapeHtml)}
            </div>
          </div>
          ${renderLobbyPanel(panelTab, bundle, events, t, escapeHtml, renderMetric, renderEventFeed, protocolLabel, busyAttr, buttonLabel)}
        </aside>
      </div>
    </section>
  `;
}

function renderLobbySeat(
  slot: LobbySlotState,
  t: (key: string, ...args: Array<string | number>) => string,
  escapeHtml: (value: string) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
): string {
  const tags = [
    slot.is_local ? t("slot_you") : "",
    slot.is_host ? t("slot_host") : "",
    slot.is_provisional_cpu ? t("slot_cpu") : "",
    slot.is_connected ? t("slot_online") : t("slot_offline"),
  ].filter(Boolean);

  return `
    <section class="lobby10-seat${slot.is_local ? " lobby10-seat-local" : ""}">
      <div class="lobby10-seat-head">
        <div>
          <strong>${escapeHtml(slot.name || t("slot_empty"))}</strong>
          <span>#${slot.seat + 1}</span>
        </div>
        <em>${escapeHtml(slotStatusLabel(slot.status, t))}</em>
      </div>
      <div class="tag-row">${tags.map((tag) => `<span>${escapeHtml(tag)}</span>`).join("")}</div>
      <div class="lobby10-seat-actions">
        ${slot.can_vote_host ? `<form data-api-action="sendHostVote" data-form-id="sendHostVote-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="ghost-button" type="submit"${busyAttr(`sendHostVote-${slot.seat}`)}>${buttonLabel(`sendHostVote-${slot.seat}`, t("action_vote_host"))}</button></form>` : ""}
        ${slot.can_request_replacement ? `<form data-api-action="requestReplacementInvite" data-form-id="replacement-${slot.seat}"><input type="hidden" name="slot" value="${slot.seat}"><button class="secondary-button strong" type="submit"${busyAttr(`replacement-${slot.seat}`)}>${buttonLabel(`replacement-${slot.seat}`, t("action_replacement_invite"))}</button></form>` : ""}
      </div>
    </section>
  `;
}

function renderLobbyPanel(
  panelTab: LobbyPanelTab,
  bundle: SnapshotBundle,
  events: RuntimeEvent[],
  t: (key: string, ...args: Array<string | number>) => string,
  escapeHtml: (value: string) => string,
  renderMetric: (label: string, value: string) => string,
  renderEventFeed: (extraLines?: string[]) => string,
  protocolLabel: (network?: SnapshotBundle["connection"]["network"]) => string,
  busyAttr: (formId: string) => string,
  buttonLabel: (formId: string, label: string) => string,
): string {
  const network = bundle.connection.network;
  switch (panelTab) {
    case "network":
      return `
        <div class="lobby10-panel">
          <div class="telemetry-grid">
            ${renderMetric(t("connection_status"), bundle.connection.status)}
            ${renderMetric(t("connection_mode"), bundle.connection.is_online ? t("connection_online") : t("connection_offline"))}
            ${renderMetric(t("connection_transport"), network?.transport || "-")}
            ${renderMetric(t("connection_protocol"), protocolLabel(network))}
            ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
            ${bundle.lobby?.role ? renderMetric(t("connection_role"), bundle.lobby.role) : ""}
          </div>
          ${bundle.connection.last_error ? `<div class="lobby10-inline-error">${escapeHtml(`${bundle.connection.last_error.code}: ${bundle.connection.last_error.message}`)}</div>` : ""}
          ${latestFailoverSignal(events, t) ? `<p class="supporting-copy">${escapeHtml(latestFailoverSignal(events, t) || "")}</p>` : ""}
        </div>
      `;
    case "chat":
      return `
        <div class="lobby10-panel">
          <pre class="event-feed" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed())}</pre>
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
            <input name="message" type="text" autocomplete="off" placeholder="${escapeHtml(t("chat_placeholder"))}">
            <button class="secondary-button" type="submit"${busyAttr("sendChat")}>${buttonLabel("sendChat", t("lobby_chat"))}</button>
          </form>
        </div>
      `;
    default:
      return `
        <div class="lobby10-panel">
          <pre class="event-feed" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${escapeHtml(renderEventFeed())}</pre>
          <div class="lobby10-signal-grid">
            ${renderMetric(t("connection_status"), bundle.connection.status)}
            ${renderMetric(t("lobby_slots"), `${bundle.ui.lobby_slots.filter((slot) => !slot.is_empty).length}/${bundle.lobby?.num_players || bundle.ui.lobby_slots.length}`)}
            ${renderMetric(t("connection_backlog"), String(bundle.diagnostics.event_backlog || 0))}
          </div>
        </div>
      `;
  }
}

function renderPanelTab(value: LobbyPanelTab, active: LobbyPanelTab, label: string, escapeHtml: (value: string) => string): string {
  return `<button class="panel-tab${active === value ? " panel-tab-active" : ""}" type="button" data-panel-tab="lobby:${value}">${escapeHtml(label)}</button>`;
}

function tabTitle(tab: LobbyPanelTab, t: (key: string, ...args: Array<string | number>) => string): string {
  switch (tab) {
    case "network":
      return t("game_network");
    case "chat":
      return t("lobby_chat");
    default:
      return t("lobby_events");
  }
}

function latestFailoverSignal(events: RuntimeEvent[], t: (key: string, ...args: Array<string | number>) => string): string {
  const recent = [...events].reverse().find((event) => event.kind === "failover_promoted" || event.kind === "failover_rejoined");
  if (!recent) {
    return "";
  }
  return recent.kind === "failover_promoted" ? t("signal_failover_promoted") : t("signal_failover_rejoined");
}

function slotStatusLabel(status: string, t: (key: string, ...args: Array<string | number>) => string): string {
  switch (status) {
    case "occupied_online":
      return t("slot_online");
    case "occupied_offline":
      return t("slot_offline");
    case "provisional_cpu":
      return t("slot_cpu");
    default:
      return t("slot_empty");
  }
}
