(() => {
  const UI_MODE_KEY = 'truco-browser-ui-mode';
  const viewRoot = document.getElementById('view-root');
  if (!viewRoot) return;
  viewRoot.tabIndex = -1;

  const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
  const uiState = window.__trucoBrowserUI || (window.__trucoBrowserUI = {
    lastTrickSeq: null,
    trickTimers: [],
    lastFocusTarget: 'auto',
    motionReduced: motionQuery.matches,
  });

  const syncMotionPreference = () => {
    uiState.motionReduced = motionQuery.matches;
  };

  syncMotionPreference();
  if (typeof motionQuery.addEventListener === 'function') {
    motionQuery.addEventListener('change', syncMotionPreference);
  } else if (typeof motionQuery.addListener === 'function') {
    motionQuery.addListener(syncMotionPreference);
  }

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
    uiState.lastFocusTarget = focusTarget;
    setBusyState(form, submitter, true);

    try {
      const payload = await submitAjaxForm(form);
      if (payload && payload.viewHtml) {
        clearTrickTimers();
        viewRoot.innerHTML = payload.viewHtml;
        syncUiModeButtons();
        syncTrickPresentation();
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
  syncTrickPresentation();
  uiState.syncTrickPresentation = syncTrickPresentation;
  uiState.startTrickAnimation = startTrickAnimation;
  uiState.clearTrickTimers = clearTrickTimers;

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

  function clearTrickTimers() {
    while (uiState.trickTimers.length > 0) {
      window.clearTimeout(uiState.trickTimers.pop());
    }
  }

  function queueTrickTimer(fn, delay) {
    const timerId = window.setTimeout(() => {
      uiState.trickTimers = uiState.trickTimers.filter((id) => id !== timerId);
      fn();
    }, delay);
    uiState.trickTimers.push(timerId);
    return timerId;
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

  function syncTrickPresentation() {
    const gamePanel = viewRoot.querySelector('.game-panel[data-last-trick-seq]');
    if (!gamePanel) {
      clearTrickTimers();
      uiState.lastTrickSeq = null;
      return;
    }

    const nextSeq = parseInt(gamePanel.dataset.lastTrickSeq || '0', 10) || 0;
    if (uiState.lastTrickSeq === null) {
      uiState.lastTrickSeq = nextSeq;
      return;
    }

    if (nextSeq <= 0) {
      uiState.lastTrickSeq = nextSeq;
      return;
    }

    if (nextSeq < uiState.lastTrickSeq) {
      clearTrickTimers();
      uiState.lastTrickSeq = nextSeq;
      return;
    }

    if (nextSeq === uiState.lastTrickSeq) {
      return;
    }

    clearTrickTimers();
    uiState.lastTrickSeq = nextSeq;

    queueTrickTimer(() => {
      startTrickAnimation(gamePanel);
    }, 16);
  }

  function startTrickAnimation(gamePanel) {
    const overlay = gamePanel.querySelector('[data-trick-overlay]');
    const stage = gamePanel.querySelector('.board-stage');
    if (!overlay || !stage) return;

    const isTie = gamePanel.dataset.lastTrickTie === '1';
    const winnerTeam = parseInt(gamePanel.dataset.lastTrickTeam || '-1', 10);
    const winnerId = parseInt(gamePanel.dataset.lastTrickWinner || '-1', 10);
    const localPlayerId = parseInt(gamePanel.dataset.localPlayerId || '-1', 10);
    const roundLabel = gamePanel.dataset.lastTrickRoundLabel || '';
    const deckTarget = { x: 0, y: 0 };
    const localSeat = gamePanel.querySelector(`.board-seat[data-player-id="${localPlayerId}"]`);
    const localTeam = localSeat ? parseInt(localSeat.dataset.team || '-1', 10) : -1;
    const result = isTie ? 'tie' : (localTeam >= 0 && winnerTeam === localTeam ? 'win' : 'loss');

    overlay.classList.remove('is-visible', 'is-traveling', 'is-reduced', 'is-win', 'is-loss', 'is-tie');
    overlay.setAttribute('aria-hidden', 'true');
    overlay.style.setProperty('--trick-target-x', '0px');
    overlay.style.setProperty('--trick-target-y', '0px');

    const kicker = overlay.querySelector('[data-trick-overlay-kicker]');
    const title = overlay.querySelector('[data-trick-overlay-title]');
    const caption = overlay.querySelector('[data-trick-overlay-caption]');
    if (kicker) {
      kicker.textContent = roundLabel;
    }
    if (title) {
      if (result === 'win') {
        title.textContent = overlay.dataset.trickToastWin || '';
      } else if (result === 'loss') {
        title.textContent = overlay.dataset.trickToastLoss || '';
      } else {
        title.textContent = overlay.dataset.trickToastTie || '';
      }
    }
    if (caption) {
      caption.textContent = overlay.dataset.trickToastCaption || '';
    }

    if (!isTie) {
      const targetSeat = winnerId >= 0
        ? gamePanel.querySelector(`.board-seat[data-player-id="${winnerId}"]`)
        : null;
      const seatRect = targetSeat ? targetSeat.getBoundingClientRect() : null;
      const stageRect = stage.getBoundingClientRect();
      if (seatRect && stageRect.width > 0 && stageRect.height > 0) {
        deckTarget.x = Math.round((seatRect.left + seatRect.width / 2) - (stageRect.left + stageRect.width / 2));
        deckTarget.y = Math.round((seatRect.top + seatRect.height / 2) - (stageRect.top + stageRect.height / 2));
      }
    }
    overlay.classList.add(`is-${result}`);
    if (uiState.motionReduced) {
      overlay.classList.add('is-reduced');
    }

    queueTrickTimer(() => {
      overlay.setAttribute('aria-hidden', 'false');
      overlay.classList.add('is-visible');
    }, 150);

    if (!uiState.motionReduced && !isTie) {
      queueTrickTimer(() => {
        overlay.style.setProperty('--trick-target-x', `${deckTarget.x}px`);
        overlay.style.setProperty('--trick-target-y', `${deckTarget.y}px`);
        overlay.classList.add('is-traveling');
      }, 900);
    }

    queueTrickTimer(() => {
      overlay.classList.remove('is-visible', 'is-traveling');
      overlay.setAttribute('aria-hidden', 'true');
    }, uiState.motionReduced ? 1200 : 2150);
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
