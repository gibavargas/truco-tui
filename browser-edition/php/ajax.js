(() => {
  const viewRoot = document.getElementById('view-root');
  if (!viewRoot) return;

  viewRoot.addEventListener('submit', async (event) => {
    const form = event.target.closest('form[data-ajax="true"]');
    if (!form) return;
    event.preventDefault();

    if (form.dataset.busy === '1') {
      return;
    }
    form.dataset.busy = '1';

    try {
      const payload = await submitAjaxForm(form);
      if (payload && payload.viewHtml) {
        viewRoot.innerHTML = payload.viewHtml;
      }
    } catch (error) {
      console.error('AJAX form failed', error);
    } finally {
      form.dataset.busy = '0';
    }
  });

  viewRoot.addEventListener('click', async (event) => {
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

  async function submitAjaxForm(form) {
    const method = (form.method || 'POST').toUpperCase();
    const action = form.action || window.location.pathname;
    const submission = new FormData(form);
    submission.set('ajax', '1');
    const url = new URL(action, window.location.origin);
    url.searchParams.set('ajax', '1');

    const response = await fetch(url.toString(), {
      method,
      body: submission,
      credentials: 'same-origin',
    });

    if (!response.ok) {
      throw new Error(`Request failed (${response.status})`);
    }

    return response.json();
  }

  // 3D Card Tilt Effect
  document.addEventListener('mousemove', (e) => {
    const cardWrap = e.target.closest('.card-btn:not(:disabled) .card');
    if (!cardWrap) return;
    
    // Using requestAnimationFrame to decouple calculation from layout thrashing
    requestAnimationFrame(() => {
        const rect = cardWrap.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const cx = rect.width / 2;
        const cy = rect.height / 2;
        
        const tiltX = ((y - cy) / cy) * -15; // Max 15 deg rotation
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

})();
