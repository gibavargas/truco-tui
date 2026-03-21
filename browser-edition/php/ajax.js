(() => {
  const viewRoot = document.getElementById("view-root");
  if (!viewRoot) return;

  const state = {
    view: viewRoot.dataset.view || detectView(),
    mode: viewRoot.dataset.mode || "idle",
    syncTimer: 0,
    retryTimer: 0,
    burstTimers: [],
    syncInFlight: false,
    lastInputAt: 0,
  };

  viewRoot.addEventListener("submit", handleSubmit);
  viewRoot.addEventListener("input", () => {
    state.lastInputAt = Date.now();
  }, true);
  document.addEventListener("visibilitychange", handleVisibilityChange);
  scheduleNextSync(250);

  document.addEventListener("mousemove", (event) => {
    const cardWrap = event.target.closest(".card-btn:not(:disabled) .card");
    if (!cardWrap) return;
    requestAnimationFrame(() => {
      const rect = cardWrap.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;
      const cx = rect.width / 2;
      const cy = rect.height / 2;
      const tiltX = ((y - cy) / cy) * -15;
      const tiltY = ((x - cx) / cx) * 15;
      cardWrap.style.transform = `scale(1.05) rotateX(${tiltX}deg) rotateY(${tiltY}deg)`;
    });
  });

  document.addEventListener("mouseout", (event) => {
    const cardWrap = event.target.closest(".card-btn .card");
    if (!cardWrap) return;
    requestAnimationFrame(() => {
      cardWrap.style.transform = "";
    });
  });

  async function handleSubmit(event) {
    const form = event.target.closest('form[data-ajax="true"]');
    if (!form) return;
    event.preventDefault();

    if (form.dataset.busy === "1") {
      return;
    }
    form.dataset.busy = "1";

    try {
      const payload = await submitAjaxForm(form);
      applyPayload(payload, { excludeForm: form });
      clearInlineError();
      if (isRefreshAction(getFormAction(form))) {
        scheduleNextSync();
      } else {
        queueBurstSync();
      }
    } catch (error) {
      console.error("AJAX form failed", error);
      showInlineError("Falha ao sincronizar a mesa. Tente novamente.");
      scheduleRetry();
    } finally {
      form.dataset.busy = "0";
    }
  }

  async function submitAjaxForm(form) {
    const method = (form.method || "POST").toUpperCase();
    const action = form.getAttribute("action") || window.location.pathname;
    const submission = new FormData(form);
    submission.set("ajax", "1");
    return sendRequest(action, method, submission);
  }

  async function runAutoSync() {
    const action = autoRefreshAction();
    if (!action) return;
    if (state.syncInFlight) return;
    if (shouldDeferSync()) {
      scheduleNextSync(650);
      return;
    }

    state.syncInFlight = true;
    try {
      const submission = new FormData();
      submission.set("action", action);
      submission.set("ajax", "1");
      const payload = await sendRequest(window.location.pathname, "POST", submission);
      applyPayload(payload, {});
      clearInlineError();
    } catch (error) {
      console.error("Auto sync failed", error);
      showInlineError("A mesa ficou sem atualização automática. O botão Atualizar continua disponível.");
      scheduleRetry();
    } finally {
      state.syncInFlight = false;
      scheduleNextSync();
    }
  }

  async function sendRequest(action, method, body) {
    const url = new URL(action, window.location.origin);
    url.searchParams.set("ajax", "1");

    const response = await fetch(url.toString(), {
      method,
      body,
      credentials: "same-origin",
    });

    if (!response.ok) {
      throw new Error(`Request failed (${response.status})`);
    }

    return response.json();
  }

  function applyPayload(payload, options) {
    if (!payload) return;

    const drafts = payload.viewHtml ? captureDrafts(options.excludeForm || null) : [];
    if (payload.viewHtml && payload.viewHtml !== viewRoot.innerHTML) {
      viewRoot.innerHTML = payload.viewHtml;
      restoreDrafts(drafts);
    }

    state.view = payload.view || detectView();
    state.mode = payload.mode || state.mode || "idle";
    viewRoot.dataset.view = state.view;
    viewRoot.dataset.mode = state.mode;

    if (!autoRefreshAction()) {
      clearAutoSync();
      clearBurstTimers();
    }
  }

  function captureDrafts(excludeForm) {
    return Array.from(viewRoot.querySelectorAll("input, textarea, select"))
      .filter((element) => {
        if (!element.name && !element.id) return false;
        if (excludeForm && element.closest("form") === excludeForm) return false;
        if (element.tagName === "INPUT") {
          const type = (element.getAttribute("type") || "text").toLowerCase();
          if (["hidden", "submit", "button", "file"].includes(type)) {
            return false;
          }
        }
        return true;
      })
      .map((element) => {
        const selector = element.id
          ? `#${CSS.escape(element.id)}`
          : `${element.tagName.toLowerCase()}[name="${CSS.escape(element.name)}"]`;
        return {
          selector,
          value: element.value,
          checked: typeof element.checked === "boolean" ? element.checked : null,
          focused: element === document.activeElement,
          selectionStart: typeof element.selectionStart === "number" ? element.selectionStart : null,
          selectionEnd: typeof element.selectionEnd === "number" ? element.selectionEnd : null,
        };
      });
  }

  function restoreDrafts(drafts) {
    drafts.forEach((draft) => {
      const element = viewRoot.querySelector(draft.selector);
      if (!element) return;
      if ("value" in element) {
        element.value = draft.value;
      }
      if (draft.checked !== null && "checked" in element) {
        element.checked = draft.checked;
      }
      if (draft.focused && typeof element.focus === "function") {
        element.focus({ preventScroll: true });
        if (
          typeof element.setSelectionRange === "function" &&
          draft.selectionStart !== null &&
          draft.selectionEnd !== null
        ) {
          element.setSelectionRange(draft.selectionStart, draft.selectionEnd);
        }
      }
    });
  }

  function autoRefreshAction() {
    if (state.view === "game") return "refreshGame";
    if (state.view === "lobby") return "refreshLobby";
    return "";
  }

  function autoRefreshDelay() {
    if (!autoRefreshAction()) return 0;
    return state.view === "game" ? 1200 : 1600;
  }

  function shouldDeferSync() {
    const active = document.activeElement;
    if (!active || !viewRoot.contains(active)) return false;
    if (!active.matches('input:not([type="hidden"]):not([type="submit"]):not([type="button"]), textarea')) {
      return false;
    }
    return Date.now() - state.lastInputAt < 900;
  }

  function scheduleNextSync(delay = autoRefreshDelay()) {
    clearTimeout(state.syncTimer);
    if (!autoRefreshAction() || delay <= 0) return;
    state.syncTimer = window.setTimeout(runAutoSync, delay);
  }

  function queueBurstSync() {
    clearBurstTimers();
    [180, 520, 1100, 2200].forEach((delay) => {
      const timerId = window.setTimeout(runAutoSync, delay);
      state.burstTimers.push(timerId);
    });
    scheduleNextSync();
  }

  function clearBurstTimers() {
    state.burstTimers.forEach((timerId) => window.clearTimeout(timerId));
    state.burstTimers = [];
  }

  function clearAutoSync() {
    window.clearTimeout(state.syncTimer);
    window.clearTimeout(state.retryTimer);
  }

  function scheduleRetry() {
    window.clearTimeout(state.retryTimer);
    if (!autoRefreshAction()) return;
    state.retryTimer = window.setTimeout(runAutoSync, 2200);
  }

  function handleVisibilityChange() {
    if (!autoRefreshAction()) return;
    if (document.hidden) {
      scheduleNextSync(autoRefreshDelay());
      return;
    }
    queueBurstSync();
  }

  function getFormAction(form) {
    return form.querySelector('input[name="action"]')?.value || "";
  }

  function isRefreshAction(action) {
    return action === "refreshGame" || action === "refreshLobby";
  }

  function detectView() {
    if (viewRoot.querySelector(".game-panel")) return "game";
    if (viewRoot.querySelector(".lobby-panel")) return "lobby";
    return "setup";
  }

  function showInlineError(message) {
    let errorNode = viewRoot.querySelector(".error-log.runtime-error");
    if (!errorNode) {
      errorNode = document.createElement("div");
      errorNode.className = "error-log runtime-error";
      viewRoot.appendChild(errorNode);
    }
    errorNode.textContent = message;
  }

  function clearInlineError() {
    const errorNode = viewRoot.querySelector(".error-log.runtime-error");
    if (errorNode) {
      errorNode.remove();
    }
  }
})();
