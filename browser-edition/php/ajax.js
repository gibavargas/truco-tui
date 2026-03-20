(() => {
  const UI_MODE_KEY = 'truco-browser-ui-mode';
  const viewRoot = document.getElementById('view-root');
  if (!viewRoot) return;
  viewRoot.tabIndex = -1;

  const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');

  applyUiMode(localStorage.getItem(UI_MODE_KEY) || 'polished');

  viewRoot.addEventListener('submit', async (event) => {
    const form = event.target.closest('form[data-ajax="true"]');
    if (!form) return;
    event.preventDefault();

    if (form.dataset.busy === '1') {
      return;
    }
    const submitter = event.submitter || form.querySelector('button[type="submit"]');
    const focusTarget = inferFocusTarget(form);
    setBusyState(form, submitter, true);

    try {
      const payload = await submitAjaxForm(form);
      if (payload && payload.viewHtml) {
        viewRoot.innerHTML = payload.viewHtml;
        syncUiModeButtons();
        restoreFocus(focusTarget);
      } else {
        throw new Error('Empty AJAX response');
      }
    } catch (error) {
      showRuntimeNotice(error, true);
      console.error('AJAX form failed', error);
    } finally {
      setBusyState(form, submitter, false);
    }
  });

  viewRoot.addEventListener('click', async (event) => {
    const modeButton = event.target.closest('[data-ui-mode]');
    if (modeButton) {
      const nextMode = modeButton.dataset.uiMode === 'wireframe' ? 'wireframe' : 'polished';
      localStorage.setItem(UI_MODE_KEY, nextMode);
      applyUiMode(nextMode);
      syncUiModeButtons();
      return;
    }

    const button = event.target.closest('.btn-copy');
    if (!button) return;
    const text = button.dataset.copyText || '';
    if (!text || !navigator.clipboard) return;
    try {
      await navigator.clipboard.writeText(text);
      const oldText = button.textContent;
      button.textContent = 'OK';
      setTimeout(() => {
        button.textContent = oldText;
      }, 1200);
    } catch (error) {
      console.error('Copy failed', error);
    }
  });

  syncUiModeButtons();

  async function submitAjaxForm(form) {
    const method = (form.method || 'POST').toUpperCase();
    const action = form.getAttribute('action') || window.location.pathname;
    const submission = new FormData(form);
    submission.set('ajax', '1');
    const url = new URL(action, window.location.origin);
    url.searchParams.set('ajax', '1');

    const response = await fetch(url.toString(), {
      method,
      body: submission,
      credentials: 'same-origin',
    });

    const text = await response.text();
    let payload = null;
    try {
      payload = text ? JSON.parse(text) : null;
    } catch (parseError) {
      if (!response.ok) {
        throw new Error(`Request failed (${response.status})`);
      }
      throw parseError;
    }

    if (!response.ok) {
      const message = payload && payload.error ? payload.error : `Request failed (${response.status})`;
      throw new Error(message);
    }

    return payload;
  }

  function applyUiMode(mode) {
    document.body.classList.toggle('ui-wireframe', mode === 'wireframe');
    document.body.classList.toggle('ui-polished', mode !== 'wireframe');
  }

  function syncUiModeButtons() {
    const mode = document.body.classList.contains('ui-wireframe') ? 'wireframe' : 'polished';
    document.querySelectorAll('[data-ui-mode]').forEach((button) => {
      button.classList.toggle('is-active', button.dataset.uiMode === mode);
      button.setAttribute('aria-pressed', button.dataset.uiMode === mode ? 'true' : 'false');
    });
  }

  function setBusyState(form, submitter, busy) {
    form.dataset.busy = busy ? '1' : '0';
    form.setAttribute('aria-busy', busy ? 'true' : 'false');
    form.classList.toggle('is-busy', busy);
    if (submitter) {
      submitter.disabled = busy;
    }
  }

  function inferFocusTarget(form) {
    const actionField = form.querySelector('input[name="action"]');
    const action = actionField ? actionField.value : '';
    switch (action) {
      case 'play':
        return 'player-hand';
      case 'truco':
      case 'accept':
      case 'refuse':
        return 'board-callout';
      case 'sendChat':
        return 'lobby-events';
      case 'refreshLobby':
      case 'refreshGame':
      case 'setLocale':
      case 'startGame':
      case 'startOnlineHost':
      case 'joinOnline':
      case 'startOnlineMatch':
      case 'leaveLobby':
      case 'reset':
      case 'voteHost':
      case 'requestReplacementInvite':
      default:
        return 'auto';
    }
  }

  function restoreFocus(targetName) {
    requestAnimationFrame(() => {
      let target = null;
      if (targetName === 'auto') {
        target = viewRoot.querySelector('[data-focus-target="board-callout"]')
          || viewRoot.querySelector('[data-focus-target="player-hand"]')
          || viewRoot.querySelector('[data-focus-target="lobby-events"]')
          || viewRoot;
      } else if (targetName && targetName !== 'view-root') {
        target = viewRoot.querySelector(`[data-focus-target="${targetName}"]`);
      } else {
        target = viewRoot;
      }
      const focusTarget = target || viewRoot;
      if (focusTarget && typeof focusTarget.focus === 'function') {
        focusTarget.focus({ preventScroll: false });
      }
    });
  }

  function showRuntimeNotice(error, stale) {
    const notice = viewRoot.querySelector('#runtime-notice');
    if (!notice) return;

    const locale = (document.documentElement.getAttribute('lang') || 'pt-BR').toLowerCase();
    const isPt = locale.startsWith('pt');
    const title = stale
      ? (isPt ? 'Conexão perdida' : 'Connection lost')
      : (isPt ? 'Erro' : 'Error');
    let message = '';
    if (error && error.message) {
      message = error.message;
    } else if (stale) {
      message = isPt ? 'Atualize para sincronizar a mesa.' : 'Refresh to resync the table.';
    } else {
      message = isPt ? 'Falha na requisição. Tente novamente.' : 'Request failed. Try again.';
    }

    notice.classList.toggle('stale', stale);
    notice.classList.toggle('error', !stale);
    notice.innerHTML = '';

    const titleEl = document.createElement('strong');
    titleEl.textContent = title;
    const messageEl = document.createElement('span');
    messageEl.textContent = message;
    notice.append(titleEl, messageEl);
    notice.focus({ preventScroll: false });
  }

  // 3D Card Tilt Effect
  if (!motionQuery.matches) {
    document.addEventListener('mousemove', (e) => {
      const cardWrap = e.target.closest('.card-btn:not(:disabled) .card');
      if (!cardWrap) return;

      requestAnimationFrame(() => {
        const rect = cardWrap.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const cx = rect.width / 2;
        const cy = rect.height / 2;

        const tiltX = ((y - cy) / cy) * -15;
        const tiltY = ((x - cx) / cx) * 15;

        cardWrap.style.transform = `scale(1.05) rotateX(${tiltX}deg) rotateY(${tiltY}deg)`;
      });
    });

    document.addEventListener('mouseout', (e) => {
      const cardWrap = e.target.closest('.card-btn .card');
      if (cardWrap) {
        requestAnimationFrame(() => {
          cardWrap.style.transform = '';
        });
      }
    });
  }

})();
