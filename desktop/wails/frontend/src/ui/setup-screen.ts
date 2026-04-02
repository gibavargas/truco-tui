import type { LocaleCode } from "../types";

interface SetupScreenParams {
  locale: LocaleCode;
  playerName: string;
  relayURL: string;
  transportMode: string;
  t: (key: string, ...args: Array<string | number>) => string;
  escapeHtml: (value: string) => string;
  busyAttr: (formId: string) => string;
  buttonLabel: (formId: string, label: string) => string;
  transportOptions: (active: string) => string;
}

export function renderSetupScreen(params: SetupScreenParams): string {
  const { locale, playerName, relayURL, transportMode, t, escapeHtml, busyAttr, buttonLabel, transportOptions } = params;
  const displayName = playerName || t("name_placeholder");
  const teamOne = locale === "en-US" ? "Team 1" : "Time 1";
  const teamTwo = locale === "en-US" ? "Team 2" : "Time 2";

  return `
    <section class="setup10">
      <article class="surface-card setup10-hero">
        <div class="setup10-hero-copy">
          <p class="eyebrow">${escapeHtml(t("setup_title"))}</p>
          <h2>${escapeHtml(t("setup_title"))}</h2>
          <p class="setup10-lede" data-pretext-block="lock-height">${escapeHtml(t("setup_intro"))}</p>
          <p class="supporting-copy">${escapeHtml(t("setup_help"))}</p>
        </div>
        <div class="setup10-atlas">
          <div class="setup10-atlas-header">
            <span class="setup10-atlas-kicker">${escapeHtml(t("app_kicker"))}</span>
            <strong>${escapeHtml(t("setup_mode_offline"))} / ${escapeHtml(t("setup_mode_online"))}</strong>
          </div>
          <div class="setup10-atlas-board">
            <div class="setup10-atlas-seat setup10-atlas-seat-top">CPU-3</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-left">CPU-4</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-right">CPU-2</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-bottom">${escapeHtml(displayName)}</div>
            <div class="setup10-atlas-core">
              <span>${escapeHtml(t("game_vira"))}</span>
              <strong>A<span class="setup10-suit">♠</span></strong>
              <span>${escapeHtml(t("game_stake"))}</span>
              <strong>3</strong>
            </div>
          </div>
          <div class="setup10-atlas-notes">
            <div class="setup10-note">
              <span>${escapeHtml(teamOne)}</span>
              <strong>${escapeHtml(displayName)} · CPU-3</strong>
            </div>
            <div class="setup10-note">
              <span>${escapeHtml(teamTwo)}</span>
              <strong>CPU-2 · CPU-4</strong>
            </div>
          </div>
        </div>
      </article>

      <div class="setup10-grid">
        <article class="surface-card setup10-offline">
          <div class="setup10-section-head">
            <div>
              <span class="section-pill">${escapeHtml(t("setup_mode_offline"))}</span>
              <h3>${escapeHtml(t("setup_offline_title"))}</h3>
              <p>${escapeHtml(t("setup_offline_note"))}</p>
            </div>
            <strong class="form-emphasis">${escapeHtml(t("setup_offline_caption"))}</strong>
          </div>
          <form class="setup10-form" data-api-action="startGame" data-form-id="startGame">
            <div class="field-grid">
              <label>
                <span>${escapeHtml(t("setup_name"))}</span>
                <input name="name" type="text" value="${escapeHtml(displayName)}" autocomplete="off">
              </label>
              <label>
                <span>${escapeHtml(t("setup_players"))}</span>
                <select name="numPlayers">
                  <option value="2">2</option>
                  <option value="4">4</option>
                </select>
              </label>
            </div>
            <div class="setup10-roster">
              <div class="setup10-roster-seat">
                <span>1</span>
                <strong>${escapeHtml(displayName)}</strong>
                <em>${escapeHtml(teamOne)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>2</span>
                <strong>CPU-2</strong>
                <em>${escapeHtml(teamTwo)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>3</span>
                <strong>CPU-3</strong>
                <em>${escapeHtml(teamOne)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>4</span>
                <strong>CPU-4</strong>
                <em>${escapeHtml(teamTwo)}</em>
              </div>
            </div>
            <button class="primary-button setup10-launch" type="submit"${busyAttr("startGame")}>${buttonLabel("startGame", t("setup_start"))}</button>
          </form>
        </article>

        <article class="surface-card setup10-online">
          <div class="setup10-section-head">
            <div>
              <span class="section-pill section-pill-hot">${escapeHtml(t("setup_mode_online"))}</span>
              <h3>${escapeHtml(t("setup_online_title"))}</h3>
              <p>${escapeHtml(t("setup_online_note"))}</p>
            </div>
            <strong class="form-emphasis">${escapeHtml(t("setup_online_caption"))}</strong>
          </div>

          <div class="setup10-online-grid">
            <form class="setup10-pane" data-api-action="startOnlineHost" data-form-id="startOnlineHost">
              <div class="setup10-pane-head">
                <div>
                  <span class="eyebrow">${escapeHtml(t("setup_host"))}</span>
                  <h4>${escapeHtml(t("setup_host"))}</h4>
                </div>
              </div>
              <div class="field-grid">
                <label>
                  <span>${escapeHtml(t("setup_name"))}</span>
                  <input name="name" type="text" value="${escapeHtml(displayName)}" autocomplete="off">
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
                <span>${escapeHtml(t("setup_transport"))}</span>
                <select name="transport_mode">
                  ${transportOptions(transportMode)}
                </select>
              </label>
              <label>
                <span>${escapeHtml(t("setup_relay"))}</span>
                <input name="relay_url" type="text" value="${escapeHtml(relayURL)}" placeholder="${escapeHtml(t("relay_placeholder"))}" autocomplete="off">
              </label>
              <button class="secondary-button" type="submit"${busyAttr("startOnlineHost")}>${buttonLabel("startOnlineHost", t("setup_host"))}</button>
            </form>

            <form class="setup10-pane" data-api-action="joinOnline" data-form-id="joinOnline">
              <div class="setup10-pane-head">
                <div>
                  <span class="eyebrow">${escapeHtml(t("setup_join"))}</span>
                  <h4>${escapeHtml(t("setup_join"))}</h4>
                </div>
              </div>
              <div class="field-grid">
                <label>
                  <span>${escapeHtml(t("setup_name"))}</span>
                  <input name="name" type="text" value="${escapeHtml(displayName)}" autocomplete="off">
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
          <p class="supporting-copy setup10-support">${escapeHtml(t("setup_online_support"))}</p>
        </article>
      </div>
    </section>
  `;
}
