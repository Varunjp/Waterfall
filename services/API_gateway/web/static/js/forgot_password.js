(() => {
  /* ── State ───────────────────────────────────── */
  let userEmail   = '';
  let resetToken  = '';
  let timerHandle = null;
  const OTP_SECS  = 5 * 60; // 5 minutes
  const CIRCUMFERENCE = 2 * Math.PI * 24; // ≈ 150.8

  /* ── Step switching ──────────────────────────── */
  function showStep(id) {
    ['step-email', 'step-otp', 'step-reset', 'step-success'].forEach(s => {
      const el = document.getElementById(s);
      el.style.display = 'none';
      el.style.animation = 'none';
    });
    const target = document.getElementById(id);
    target.style.display = 'block';
    requestAnimationFrame(() => { target.style.animation = ''; });
  }

  /* ── Helpers ─────────────────────────────────── */
  function showError(id, msg) {
    const el = document.getElementById(id);
    el.textContent = msg;
    el.classList.add('visible');
  }

  function clearError(id) {
    const el = document.getElementById(id);
    el.textContent = '';
    el.classList.remove('visible');
  }

  function setLoading(btn, loading, label) {
    btn.disabled = loading;
    btn.querySelector('span').textContent = loading ? 'Please wait…' : label;
  }

  /* ════════════════════════════════════════════════
     STEP 1 — Email
  ════════════════════════════════════════════════ */
  document.getElementById('emailForm').onsubmit = async (e) => {
    e.preventDefault();
    clearError('email-error');
    const email = document.getElementById('fp-email').value.trim();

    if (!email) { showError('email-error', 'Please enter your email.'); return; }

    const btn = e.target.querySelector('button[type="submit"]');
    setLoading(btn, true, 'Send Reset Code');

    try {
      const res = await fetch('/api/v1/users/password/reset/request', {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ email }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);

      userEmail = email;
      document.getElementById('otp-email-display').textContent = email;
      showStep('step-otp');
      startOtpTimer();
      focusFirstOtpBox();
    } catch (err) {
      showError('email-error', err.message || 'Failed to send reset code. Try again.');
    } finally {
      setLoading(btn, false, 'Send Reset Code');
    }
  };

  /* ════════════════════════════════════════════════
     STEP 2 — OTP
  ════════════════════════════════════════════════ */

  /* ── OTP box behaviour ───────────────────────── */
  function focusFirstOtpBox() {
    const boxes = document.querySelectorAll('.otp-box');
    boxes[0].value = '';
    boxes[0].focus();
  }

  document.getElementById('otp-boxes').addEventListener('input', e => {
    const box = e.target;
    if (!/^\d$/.test(box.value)) { box.value = ''; return; }
    box.classList.add('filled');
    const boxes = [...document.querySelectorAll('.otp-box')];
    const idx   = boxes.indexOf(box);
    if (idx < boxes.length - 1) boxes[idx + 1].focus();
  });

  document.getElementById('otp-boxes').addEventListener('keydown', e => {
    const box   = e.target;
    const boxes = [...document.querySelectorAll('.otp-box')];
    const idx   = boxes.indexOf(box);
    if (e.key === 'Backspace') {
      if (box.value) { box.value = ''; box.classList.remove('filled'); }
      else if (idx > 0) { boxes[idx - 1].focus(); boxes[idx - 1].value = ''; boxes[idx - 1].classList.remove('filled'); }
    } else if (e.key === 'ArrowLeft'  && idx > 0) boxes[idx - 1].focus();
    else if (e.key === 'ArrowRight' && idx < boxes.length - 1) boxes[idx + 1].focus();
  });

  // Handle paste of full OTP code
  document.getElementById('otp-boxes').addEventListener('paste', e => {
    e.preventDefault();
    const pasted = (e.clipboardData || window.clipboardData).getData('text').replace(/\D/g, '').slice(0, 6);
    const boxes  = [...document.querySelectorAll('.otp-box')];
    pasted.split('').forEach((ch, i) => {
      if (boxes[i]) { boxes[i].value = ch; boxes[i].classList.add('filled'); }
    });
    const next = boxes[Math.min(pasted.length, 5)];
    if (next) next.focus();
  });

  function getOtpValue() {
    return [...document.querySelectorAll('.otp-box')].map(b => b.value).join('');
  }

  /* ── Timer ───────────────────────────────────── */
  function startOtpTimer() {
    clearInterval(timerHandle);
    let remaining = OTP_SECS;
    updateTimer(remaining);
    document.getElementById('resend-btn').disabled = true;

    timerHandle = setInterval(() => {
      remaining--;
      updateTimer(remaining);
      if (remaining <= 0) {
        clearInterval(timerHandle);
        document.getElementById('resend-btn').disabled = false;
      }
    }, 1000);
  }

  function updateTimer(secs) {
    const mins    = Math.floor(secs / 60);
    const s       = secs % 60;
    const label   = `${mins}:${String(s).padStart(2, '0')}`;
    const pct     = secs / OTP_SECS;
    const offset  = CIRCUMFERENCE * (1 - pct);

    const textEl  = document.getElementById('otp-timer-text');
    const ringEl  = document.getElementById('otp-ring-fill');

    textEl.textContent    = label;
    ringEl.style.strokeDashoffset = offset;

    // Colour states
    const urgent = secs <= 60;
    const warn   = secs <= 120 && !urgent;
    textEl.className = urgent ? 'otp-timer-text urgent' : warn ? 'otp-timer-text warn' : 'otp-timer-text';
    ringEl.className = urgent ? 'otp-ring-fill urgent'  : warn ? 'otp-ring-fill warn'  : 'otp-ring-fill';
  }

  /* ── Resend ──────────────────────────────────── */
  document.getElementById('resend-btn').addEventListener('click', async () => {
    clearError('otp-error');
    const btn = document.getElementById('resend-btn');
    btn.disabled = true;
    btn.textContent = 'Sending…';

    try {
      const res = await fetch('/api/v1/users/password/reset/request', {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ email: userEmail }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || 'Failed to resend.');

      // Clear boxes and restart timer
      document.querySelectorAll('.otp-box').forEach(b => { b.value = ''; b.classList.remove('filled'); });
      focusFirstOtpBox();
      startOtpTimer();
    } catch (err) {
      showError('otp-error', err.message);
      btn.disabled = false;
      btn.textContent = 'Resend';
    }
  });

  /* ── OTP submit ──────────────────────────────── */
  document.getElementById('otpForm').onsubmit = async (e) => {
    e.preventDefault();
    clearError('otp-error');
    const otp = getOtpValue();

    if (otp.length < 6) { showError('otp-error', 'Please enter all 6 digits.'); return; }

    const btn = document.getElementById('otp-submit-btn');
    setLoading(btn, true, 'Verify Code');

    try {
      const res = await fetch('/api/v1/users/password/reset/verify', {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ email: userEmail, otp }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
      if (!data.accessToken) throw new Error('Invalid response — no token received.');

      resetToken = data.accessToken;
      clearInterval(timerHandle);
      showStep('step-reset');
      document.getElementById('new-password').focus();
    } catch (err) {
      showError('otp-error', err.message || 'Invalid or expired code.');
    } finally {
      setLoading(btn, false, 'Verify Code');
    }
  };

  /* ════════════════════════════════════════════════
     STEP 3 — New password
  ════════════════════════════════════════════════ */

  /* ── Password strength ───────────────────────── */
  document.getElementById('new-password').addEventListener('input', e => {
    const val = e.target.value;
    const bar = document.getElementById('pw-strength');
    let score = 0;
    if (val.length >= 8)                   score++;
    if (val.length >= 12)                  score++;
    if (/[A-Z]/.test(val))                 score++;
    if (/[0-9]/.test(val))                 score++;
    if (/[^A-Za-z0-9]/.test(val))          score++;

    const levels = [
      { pct: '0%',   color: 'transparent' },
      { pct: '25%',  color: '#f0a8a8' },
      { pct: '50%',  color: '#f0e6a8' },
      { pct: '75%',  color: '#a8d4f0' },
      { pct: '90%',  color: '#a8f0b8' },
      { pct: '100%', color: '#a8f0b8' },
    ];
    const l = levels[score] || levels[0];
    bar.style.setProperty('--strength-pct',   l.pct);
    bar.style.setProperty('--strength-color', l.color);
  });

  /* ── Show / hide password toggles ───────────── */
  document.querySelectorAll('.pw-toggle').forEach(btn => {
    btn.addEventListener('click', () => {
      const input = document.getElementById(btn.dataset.target);
      input.type  = input.type === 'password' ? 'text' : 'password';
    });
  });

  /* ── Reset submit ────────────────────────────── */
  document.getElementById('resetForm').onsubmit = async (e) => {
    e.preventDefault();
    clearError('reset-error');

    const password = document.getElementById('new-password').value;
    const confirm  = document.getElementById('confirm-password').value;

    if (password.length < 8) {
      showError('reset-error', 'Password must be at least 8 characters.'); return;
    }
    if (password !== confirm) {
      showError('reset-error', 'Passwords do not match.'); return;
    }

    const btn = document.getElementById('reset-submit-btn');
    setLoading(btn, true, 'Set New Password');

    try {
      const res = await fetch('/api/v1/users/password/reset', {
        method:  'POST',
        headers: {
          'Content-Type':  'application/json',
          'Authorization': `Bearer ${resetToken}`,
        },
        body: JSON.stringify({ password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);

      showStep('step-success');
    } catch (err) {
      showError('reset-error', err.message || 'Failed to reset password. Please try again.');
    } finally {
      setLoading(btn, false, 'Set New Password');
    }
  };

})();
