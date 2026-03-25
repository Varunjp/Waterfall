(() => {
  /* ── Auth ──────────────────────────────────────── */
  const token = (() => {
    const m = document.cookie.match(/token=([^;]+)/);
    return m ? m[1] : null;
  })();

  if (!token) { window.location.href = '/login'; }

  const authHeader = { Authorization: 'Bearer ' + token };

  /* ── State ─────────────────────────────────────── */
  let currentPage      = 0;
  let currentSection   = 'jobs';
  let currentFilter    = 'ALL';
  let currentStartDate = null;
  let currentEndDate   = null;
  let totalJobs        = 0;
  const LIMIT          = 10;
  let currentUserRole  = null; // decoded from token

  /* ── Helpers ───────────────────────────────────── */
  function fmt(dateStr) {
    if (!dateStr) return '—';
    const d = new Date(dateStr);
    if (isNaN(d)) return dateStr;
    return d.toLocaleString('en-GB', {
      day: '2-digit', month: 'short', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
      hour12: false, timeZone: 'UTC',
    });
  }

  function fmtPayload(raw) {
    try { return JSON.stringify(JSON.parse(raw), null, 2); }
    catch { return raw || '—'; }
  }

  function statusBadge(status) {
    return `<span class="badge badge-${status}">${status}</span>`;
  }

  function escHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;')
      .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  /* ── Navigation ────────────────────────────────── */
  document.querySelectorAll('.nav-btn[data-section]').forEach(btn => {
    btn.addEventListener('click', () => {
      currentSection = btn.dataset.section;
      document.querySelectorAll('.nav-btn[data-section]').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      document.getElementById('topbar-title').textContent = btn.dataset.section.toUpperCase();
      loadSection(currentSection);
    });
  });

  function loadSection(section) {
    const area = document.getElementById('content-area');
    area.style.animation = 'none';
    requestAnimationFrame(() => {
      area.style.animation = '';
      if (section === 'jobs')        loadJobs(0);
      else if (section === 'plans')  loadPlans();
      else if (section === 'users')        loadUsers();
      else if (section === 'subscription') loadSubscription();
      else                               renderPlaceholder(section);
    });
  }

  function renderPlaceholder(section) {
    document.getElementById('content-area').innerHTML = `
      <div class="placeholder-panel">
        <p class="ph-title">${section.toUpperCase()}</p>
        <p class="ph-sub">Section coming soon</p>
      </div>`;
  }

  /* ══════════════════════════════════════════════════
     PLANS
  ══════════════════════════════════════════════════ */
  async function loadPlans() {
    const area = document.getElementById('content-area');
    area.innerHTML = `
      <div class="section-header">
        <div>
          <p class="section-sub">Billing &amp; Subscriptions</p>
          <h2>PLANS</h2>
        </div>
      </div>
      <div id="plans-body"><div class="state-loading">Loading</div></div>`;

    try {
      const res   = await fetch('/api/v1/plans', { headers: authHeader });
      const data  = await res.json();
      const plans = data.plans || [];

      if (!plans.length) {
        document.getElementById('plans-body').innerHTML =
          `<div class="state-empty">No plans available</div>`;
        return;
      }

      // Sort cheapest → most expensive
      plans.sort((a, b) => a.planprice - b.planprice);

      document.getElementById('plans-body').innerHTML = `
        <div class="plans-grid">
          ${plans.map((p, i) => renderPlanCard(p, i, plans.length)).join('')}
        </div>`;

    } catch {
      document.getElementById('plans-body').innerHTML =
        `<div class="state-empty">Failed to load plans</div>`;
    }
  }

  function renderPlanCard(plan, idx, total) {
    // Middle card is "featured"; if only 2, the second one is featured
    const featuredIdx = total > 2 ? Math.floor(total / 2) : total - 1;
    const featured    = idx === featuredIdx;

    const price = (plan.planprice).toLocaleString('en-IN', {
      style: 'currency', currency: 'INR', maximumFractionDigits: 0,
    });
    const limit = Number(plan.monthlyLimit).toLocaleString();

    return `
      <div class="plan-card ${featured ? 'plan-card--featured' : ''}">

        ${featured ? '<span class="plan-badge">Popular</span>' : ''}

        <div class="plan-header">
          <p class="plan-name">${escHtml(plan.planName)}</p>
          <div class="plan-price-block">
            <span class="plan-price-amount">${price}</span>
            <span class="plan-price-period">/ mo</span>
          </div>
        </div>

        <div class="plan-divider"></div>

        <ul class="plan-features">
          <li class="plan-feature">
            <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" width="12" height="12">
              <polyline points="2 7 5.5 10.5 12 3.5"/>
            </svg>
            <span><strong>${limit}</strong> jobs / month</span>
          </li>
          <li class="plan-feature">
            <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" width="12" height="12">
              <polyline points="2 7 5.5 10.5 12 3.5"/>
            </svg>
            <span>Priority queue access</span>
          </li>
          <li class="plan-feature">
            <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" width="12" height="12">
              <polyline points="2 7 5.5 10.5 12 3.5"/>
            </svg>
            <span>Retry &amp; scheduling</span>
          </li>
          <li class="plan-feature">
            <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" width="12" height="12">
              <polyline points="2 7 5.5 10.5 12 3.5"/>
            </svg>
            <span>Full API access</span>
          </li>
        </ul>

        <div class="plan-meta">
          <span class="plan-meta-key">Plan ID</span>
          <span class="plan-meta-val">${escHtml(plan.planId)}</span>
        </div>

        <button
          class="btn-purchase ${featured ? 'btn-purchase--featured' : ''} ${getMyRole() === 'viewer' ? 'btn-purchase--disabled' : ''}"
          data-plan-id="${escHtml(plan.planId)}"
          data-plan-name="${escHtml(plan.planName)}"
          ${getMyRole() === 'viewer' ? 'title="Viewers cannot purchase plans"' : ''}
        >${getMyRole() === 'viewer' ? 'Unavailable' : `Get ${escHtml(plan.planName)}`}</button>

      </div>`;
  }

  // Delegate purchase clicks from dynamically rendered cards
  document.getElementById('content-area').addEventListener('click', e => {
    const btn = e.target.closest('.btn-purchase');
    if (!btn) return;
    purchasePlan(btn.dataset.planId, btn.dataset.planName, btn);
  });

  async function purchasePlan(planId, planName, btn) {
    if (getMyRole() === 'viewer') {
      showToast('Viewers cannot purchase plans. Contact your admin.', 'error');
      return;
    }
    const original  = btn.textContent;
    btn.disabled    = true;
    btn.textContent = 'Processing…';

    try {
      const res = await fetch('/billing/checkout', {
        method:  'POST',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify({ plan_id: planId }),
      });

      // If server redirects us directly (3xx), follow it
      if (res.redirected) {
        window.location.href = res.url;
        return;
      }

      // Try to parse JSON — fall back gracefully if response isn't JSON
      const contentType = res.headers.get('content-type') || '';
      let data = {};
      if (contentType.includes('application/json')) {
        data = await res.json();
      } else {
        // Non-JSON body — read as text for error reporting only
        const text = await res.text();
        if (!res.ok) throw new Error(text || `Server error (${res.status})`);
      }

      if (!res.ok) throw new Error(data.message || `Server error (${res.status})`);

      // Follow redirect URL returned in body
      const redirectTo = data.checkout_url || data.url || data.checkoutUrl || data.redirect;
      if (redirectTo) {
        window.location.href = redirectTo;
        return;
      }

      showToast(`${planName} activated!`, 'success');
    } catch (err) {
      showToast(err.message || 'Purchase failed. Please try again.', 'error');
    } finally {
      btn.disabled    = false;
      btn.textContent = original;
    }
  }

  /* ══════════════════════════════════════════════════
     JOBS
  ══════════════════════════════════════════════════ */
  async function loadJobs(page = 0, filter = currentFilter, startDate = currentStartDate, endDate = currentEndDate) {
    currentPage      = page;
    currentFilter    = filter;
    currentStartDate = startDate;
    currentEndDate   = endDate;

    const area = document.getElementById('content-area');
    area.innerHTML = `
      <div class="section-header">
        <div>
          <p class="section-sub">Queue Overview</p>
          <h2>JOBS</h2>
        </div>
      </div>
      <div class="stats-row" id="stats-row"></div>
      <div class="table-wrap">
        <div class="table-toolbar">
          <div class="table-toolbar-left">
            <span class="toolbar-label">Status</span>
            <select class="filter-select" id="filterSelect">
              <option value="ALL">All</option>
              <option value="COMPLETED">Completed</option>
              <option value="FAILED">Failed</option>
              <option value="CANCELLED">Cancelled</option>
              <option value="PENDING">Pending</option>
              <option value="SCHEDULED">Scheduled</option>
              <option value="QUEUED">Queued</option>
              <option value="RUNNING">Running</option>
            </select>
            <span class="toolbar-divider"></span>
            <span class="toolbar-label">From</span>
            <input type="date" class="filter-date" id="startDate" />
            <span class="toolbar-label">To</span>
            <input type="date" class="filter-date" id="endDate" />
            <button class="btn-filter-apply" id="btnApply">Apply</button>
            <button class="btn-filter-clear" id="btnClear">Clear</button>
          </div>
          <span class="toolbar-label" id="record-count">—</span>
        </div>
        <div id="table-body"><div class="state-loading">Loading</div></div>
        <div class="pagination" id="pagination"></div>
      </div>`;

    // Today in YYYY-MM-DD (local) — hard max for both pickers
    const todayStr = new Date().toLocaleDateString('en-CA');

    const startEl = document.getElementById('startDate');
    const endEl   = document.getElementById('endDate');

    // Block future dates on both inputs
    startEl.max = todayStr;
    endEl.max   = todayStr;

    // Restore saved filter values
    document.getElementById('filterSelect').value = filter;
    if (startDate) startEl.value = startDate.substring(0, 10);
    if (endDate)   endEl.value   = endDate.substring(0, 10);

    // When start changes: end must be >= start
    startEl.addEventListener('change', () => {
      endEl.min = startEl.value || '';
      if (endEl.value && startEl.value && endEl.value < startEl.value) {
        endEl.value = startEl.value;
      }
    });

    // When end changes: prevent end < start
    endEl.addEventListener('change', () => {
      if (startEl.value && endEl.value && endEl.value < startEl.value) {
        endEl.value = startEl.value;
      }
    });

    // Status dropdown — reset page on change
    document.getElementById('filterSelect').addEventListener('change', e => {
      loadJobs(0, e.target.value, currentStartDate, currentEndDate);
    });

    // Apply button — convert YYYY-MM-DD picker value to ISO UTC format for API
    document.getElementById('btnApply').addEventListener('click', () => {
      const toISO = (val, endOfDay = false) => {
        if (!val) return null;
        return endOfDay ? `${val}T23:59:59Z` : `${val}T00:00:00Z`;
      };
      loadJobs(0, currentFilter, toISO(startEl.value), toISO(endEl.value, true));
    });

    // Clear button — wipe dates and reload
    document.getElementById('btnClear').addEventListener('click', () => {
      startEl.value = '';
      endEl.value   = '';
      loadJobs(0, currentFilter, null, null);
    });

    try {
      const params = new URLSearchParams({ limit: LIMIT, offset: page });
      if (filter !== 'ALL')  params.set('status',    filter);
      if (startDate)         params.set('startDate', startDate);
      if (endDate)           params.set('endDate',   endDate);

      const res  = await fetch(`/api/v1/jobs?${params}`, { headers: authHeader });
      const data = await res.json();
      const jobs = data.jobs || [];
      totalJobs  = data.total || jobs.length;

      renderStats(jobs);
      renderTable(jobs);
      renderPagination(page, totalJobs);
    } catch {
      document.getElementById('table-body').innerHTML =
        `<div class="state-empty">Failed to load jobs</div>`;
    }
  }

  function renderStats(jobs) {
    const c = { COMPLETED: 0, FAILED: 0, CANCELED: 0, PENDING: 0, RUNNING: 0, SCHEDULED: 0, QUEUED: 0 };
    jobs.forEach(j => { if (c[j.status] !== undefined) c[j.status]++; });

    document.getElementById('stats-row').innerHTML = `
      <div class="stat-card">
        <p class="stat-label">Total Shown</p>
        <p class="stat-value">${jobs.length}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Completed</p>
        <p class="stat-value success">${c.COMPLETED}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Failed</p>
        <p class="stat-value failed">${c.FAILED + c.CANCELED}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Pending / Running</p>
        <p class="stat-value">${c.PENDING + c.RUNNING + c.SCHEDULED + c.QUEUED}</p>
      </div>`;
  }

  function renderTable(jobs) {
    if (!jobs.length) {
      document.getElementById('table-body').innerHTML =
        `<div class="state-empty">No jobs found</div>`;
      return;
    }

    const rows = jobs.map(j => `
      <tr data-job='${JSON.stringify(j).replace(/'/g, "&#39;")}'>
        <td><span class="cell-id" title="${escHtml(j.jobId)}">${escHtml(j.jobId)}</span></td>
        <td>${escHtml(j.type || '—')}</td>
        <td>${statusBadge(j.status)}</td>
        <td>${j.retry ?? 0} / ${j.maxRetry ?? 0}</td>
        <td>${fmt(j.scheduleAt)}</td>
        <td>${fmt(j.createdAt)}</td>
        <td>
          ${j.status === 'FAILED'
            ? `<button class="btn-retry" onclick="event.stopPropagation(); window.retryJob('${escHtml(j.jobId)}')">Retry</button>`
            : '—'}
        </td>
      </tr>`).join('');

    document.getElementById('table-body').innerHTML = `
      <table>
        <thead>
          <tr>
            <th>Job ID</th><th>Type</th><th>Status</th><th>Retry</th>
            <th>Scheduled At</th><th>Created At</th><th>Action</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>`;

    document.querySelectorAll('tbody tr').forEach(row => {
      row.addEventListener('click', () => {
        try { openModal(JSON.parse(row.dataset.job)); } catch {}
      });
    });

    document.getElementById('record-count').textContent =
      `${jobs.length} record${jobs.length !== 1 ? 's' : ''}`;
  }

  function renderPagination(page, total) {
    const from = total === 0 ? 0 : page + 1;
    const to   = Math.min(page + LIMIT, total);
    document.getElementById('pagination').innerHTML = `
      <span class="pagination-info">${from}–${to} of ${total}</span>
      <div class="pagination-btns">
        <button class="btn-page" onclick="window._prevPage()" ${page <= 0 ? 'disabled' : ''}>← Prev</button>
        <button class="btn-page" onclick="window._nextPage()" ${page + LIMIT >= total ? 'disabled' : ''}>Next →</button>
      </div>`;
  }

  window._prevPage = () => loadJobs(currentPage - LIMIT, currentFilter, currentStartDate, currentEndDate);
  window._nextPage = () => loadJobs(currentPage + LIMIT, currentFilter, currentStartDate, currentEndDate);

  window.retryJob = async (jobId) => {
    try {
      await fetch(`/api/v1/jobs/${jobId}/retry`, { method: 'POST', headers: authHeader });
      loadJobs(currentPage, currentFilter, currentStartDate, currentEndDate);
    } catch {
      showToast('Retry failed', 'error');
    }
  };

  /* ── Job Detail Modal ──────────────────────────── */
  function openModal(job) {
    document.getElementById('modal-job-id').textContent       = job.jobId;
    document.getElementById('modal-type').textContent         = job.type || '—';
    document.getElementById('modal-status').innerHTML         = statusBadge(job.status);
    document.getElementById('modal-retry').textContent        = `${job.retry ?? 0} / ${job.maxRetry ?? 0}`;
    document.getElementById('modal-manual-retry').textContent = job.manualRetry ?? 0;
    document.getElementById('modal-schedule').textContent     = fmt(job.scheduleAt);
    document.getElementById('modal-created').textContent      = fmt(job.createdAt);
    document.getElementById('modal-updated').textContent      = fmt(job.updatedAt);
    document.getElementById('modal-payload').textContent      = fmtPayload(job.payload);

    const logsPanel  = document.getElementById('logs-panel');
    const logsBody   = document.getElementById('logs-body');
    const logsBtn    = document.getElementById('modal-logs-btn');
    const refreshBtn = document.getElementById('btn-logs-refresh');
    const retryBtn   = document.getElementById('modal-retry-btn');

    logsPanel.style.display = 'none';
    logsBody.innerHTML      = '<div class="logs-empty">Click refresh to load logs</div>';
    logsBtn.classList.remove('active');

    // Retry — FAILED only
    if (job.status === 'FAILED') {
      retryBtn.style.display = 'block';
      retryBtn.onclick       = () => { closeModal(); window.retryJob(job.jobId); };
    } else {
      retryBtn.style.display = 'none';
    }

    // Update + Cancel — PENDING / SCHEDULED / QUEUED only
    const isPending   = ['PENDING','SCHEDULED','QUEUED'].includes(job.status);
    const updateBtn   = document.getElementById('modal-update-btn');
    const cancelJobBtn = document.getElementById('modal-cancel-job-btn');
    const editPanel   = document.getElementById('job-edit-panel');

    if (isPending) {
      // Pre-fill edit fields
      try {
        const parsed = JSON.parse(job.payload || '{}');
        document.getElementById('edit-payload').value = JSON.stringify(parsed, null, 2);
      } catch {
        document.getElementById('edit-payload').value = job.payload || '';
      }
      // Pre-fill: show UTC time directly; store original to detect user changes
      const scheduleEl = document.getElementById('edit-schedule');
      if (job.scheduleAt) {
        const utcVal = job.scheduleAt.replace('Z', '').substring(0, 16);
        scheduleEl.value = utcVal;
        scheduleEl.dataset.original = utcVal; // remember what we set
      } else {
        scheduleEl.value = '';
        scheduleEl.dataset.original = '';
      }
      document.getElementById('edit-schedule').max = '';  // allow future for scheduling
      document.getElementById('job-edit-error').textContent = '';
      editPanel.style.display  = 'block';
      updateBtn.style.display  = 'flex';
      cancelJobBtn.style.display = 'flex';
      updateBtn.onclick = () => updateJob(job.jobId);
      cancelJobBtn.onclick = () => cancelJob(job.jobId);
    } else {
      editPanel.style.display   = 'none';
      updateBtn.style.display   = 'none';
      cancelJobBtn.style.display = 'none';
    }

    // Logs — all jobs
    logsBtn.style.display = 'flex';
    logsBtn.onclick = () => {
      const isOpen = logsPanel.style.display !== 'none';
      logsPanel.style.display = isOpen ? 'none' : 'block';
      logsBtn.classList.toggle('active', !isOpen);
      if (!isOpen) fetchLogs(job.jobId);
    };
    refreshBtn.onclick = () => fetchLogs(job.jobId);

    document.getElementById('modal-overlay').classList.add('open');
  }

  /* ── Update job ────────────────────────────────── */
  async function updateJob(jobId) {
    const payloadRaw  = document.getElementById('edit-payload').value.trim();
    const scheduleVal = document.getElementById('edit-schedule').value;
    const errEl       = document.getElementById('job-edit-error');
    const updateBtn   = document.getElementById('modal-update-btn');
    errEl.textContent = '';

    // Build body with only set fields
    const body = {};
    if (payloadRaw) {
      try { JSON.parse(payloadRaw); } // validate JSON
      catch { errEl.textContent = 'Payload must be valid JSON.'; return; }
      body.payload = payloadRaw;
    }
    // Only send schedule_at if user actually changed it from the pre-filled value
    const scheduleEl    = document.getElementById('edit-schedule');
    const originalVal   = scheduleEl.dataset.original || '';
    const scheduleChanged = scheduleVal !== originalVal;
    if (scheduleVal && scheduleChanged) {
      body.schedule_at = scheduleVal.length === 16 ? `${scheduleVal}:00Z` : `${scheduleVal}Z`;
    }
    if (!Object.keys(body).length) {
      errEl.textContent = 'Nothing to update — enter a payload or schedule time.';
      return;
    }

    updateBtn.disabled = true;
    updateBtn.querySelector('span').textContent = 'Saving…';
    try {
      const res = await fetch(`/api/v1/jobs/${jobId}`, {
        method:  'PUT',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify(body),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
      closeModal();
      showToast('Job updated — takes effect only if not yet picked.', 'success');
      loadJobs(currentPage, currentFilter, currentStartDate, currentEndDate);
    } catch (err) {
      errEl.textContent = err.message;
    } finally {
      updateBtn.disabled = false;
      updateBtn.querySelector('span').textContent = 'Update Job';
    }
  }

  /* ── Cancel job ────────────────────────────────── */
  async function cancelJob(jobId) {
    const cancelJobBtn = document.getElementById('modal-cancel-job-btn');
    cancelJobBtn.disabled = true;
    cancelJobBtn.querySelector('span').textContent = 'Cancelling…';
    try {
      const res = await fetch(`/api/v1/jobs/${jobId}`, {
        method:  'DELETE',
        headers: authHeader,
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
      closeModal();
      showToast('Job cancelled — only if not yet picked.', 'success');
      loadJobs(currentPage, currentFilter, currentStartDate, currentEndDate);
    } catch (err) {
      document.getElementById('job-edit-error').textContent = err.message;
      cancelJobBtn.disabled = false;
      cancelJobBtn.querySelector('span').textContent = 'Cancel Job';
    }
  }

  /* ── Logs ──────────────────────────────────────── */
  async function fetchLogs(jobId) {
    const logsBody   = document.getElementById('logs-body');
    const refreshBtn = document.getElementById('btn-logs-refresh');

    logsBody.innerHTML = '<div class="logs-loading">Fetching logs</div>';
    refreshBtn.classList.add('spinning');

    try {
      const res  = await fetch(`/api/v1/jobs/${jobId}/logs`, { headers: authHeader });
      const data = await res.json();
      refreshBtn.classList.remove('spinning');

      const logs = Array.isArray(data) ? data : (data.logs ?? []);
      if (!logs.length) {
        logsBody.innerHTML = '<div class="logs-empty">No logs available</div>';
        return;
      }
      logsBody.innerHTML = logs.map(renderLogEntry).join('');
    } catch (err) {
      refreshBtn.classList.remove('spinning');
      logsBody.innerHTML =
        `<div class="log-error-raw">Failed to fetch logs: ${escHtml(err.message)}</div>`;
    }
  }

  function renderLogEntry(entry) {
    const ts     = entry.timestamp || entry.time || entry.createdAt || '';
    const status = (entry.status || '').toUpperCase();
    const msg    = entry.errorMessage || entry.message || entry.msg || '';

    let cls = '';
    if (status === 'FAILED')                                 cls = 'log-error';
    else if (status === 'JOB_RETRY')                         cls = 'log-warn';
    else if (status === 'COMPLETED' || status === 'SUCCESS') cls = 'log-success';

    const tsHtml     = ts     ? `<span class="log-ts">${fmt(ts)}</span>` : '';
    const statusHtml = status ? `<span class="log-status-tag log-status-${status}">${status}</span>` : '';
    const msgHtml    = msg    ? `<span class="log-msg">${escHtml(msg)}</span>` : '';

    return `<div class="log-line ${cls}">${tsHtml}${statusHtml}${msgHtml}</div>`;
  }

  function closeModal() {
    document.getElementById('modal-overlay').classList.remove('open');
    document.getElementById('logs-panel').style.display = 'none';
    document.getElementById('modal-logs-btn').classList.remove('active');
    document.getElementById('logs-body').innerHTML =
      '<div class="logs-empty">Click refresh to load logs</div>';
  }

  window.closeModal = closeModal;

  document.getElementById('modal-overlay').addEventListener('click', e => {
    if (e.target === document.getElementById('modal-overlay')) closeModal();
  });
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeModal();
  });


  /* ══════════════════════════════════════════════════
     USERS
  ══════════════════════════════════════════════════ */

  /* Decode JWT payload to get the logged-in user's role */
  function getMyRole() {
    if (currentUserRole) return currentUserRole;
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      currentUserRole = payload.role || payload.Role || '';
    } catch { currentUserRole = ''; }
    return currentUserRole;
  }

  async function loadUsers() {
    const area = document.getElementById('content-area');
    const isSuperAdmin = getMyRole() === 'super_admin';

    area.innerHTML = `
      <div class="section-header">
        <div>
          <p class="section-sub">Account Management</p>
          <h2>USERS</h2>
        </div>
        ${isSuperAdmin ? `<button class="btn-create-user" onclick="openCreateUserModal()">+ Create User</button>` : ''}
      </div>
      <div class="table-wrap">
        <div class="table-toolbar">
          <div class="table-toolbar-left">
            <span class="toolbar-label">User Directory</span>
          </div>
          <span class="toolbar-label" id="user-count">—</span>
        </div>
        <div id="user-table-body"><div class="state-loading">Loading</div></div>
      </div>`;

    try {
      const res   = await fetch('/api/v1/users', { headers: authHeader });
      const data  = await res.json();
      const users = data.users || [];

      document.getElementById('user-count').textContent =
        `${users.length} user${users.length !== 1 ? 's' : ''}`;

      if (!users.length) {
        document.getElementById('user-table-body').innerHTML =
          '<div class="state-empty">No users found</div>';
        return;
      }

      const rows = users.map(u => `
        <tr>
          <td><span class="cell-id" title="${escHtml(u.id)}">${escHtml(u.id)}</span></td>
          <td>${escHtml(u.email)}</td>
          <td><span class="role-tag role-${escHtml(u.role)}">${escHtml(u.role)}</span></td>
          <td>${userStatusBadge(u.status)}</td>
          <td class="user-actions">
            ${isSuperAdmin ? `
              <button class="btn-user-toggle ${u.status === 'active' ? 'btn-user-disable' : 'btn-user-enable'}"
                onclick="toggleUser('${escHtml(u.id)}', '${escHtml(u.status)}', this)">
                ${u.status === 'active' ? 'Disable' : 'Enable'}
              </button>
              <button class="btn-user-edit"
                onclick="openUserModal('${escHtml(u.id)}', '${escHtml(u.email)}', '${escHtml(u.role)}')">
                Edit
              </button>
            ` : '<span class="user-actions-readonly">—</span>'}
          </td>
        </tr>`).join('');

      document.getElementById('user-table-body').innerHTML = `
        <table>
          <thead>
            <tr>
              <th>User ID</th>
              <th>Email</th>
              <th>Role</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>`;

    } catch {
      document.getElementById('user-table-body').innerHTML =
        '<div class="state-empty">Failed to load users</div>';
    }
  }

  function userStatusBadge(status) {
    const cls = status === 'active' ? 'badge-COMPLETED' : 'badge-FAILED';
    return `<span class="badge ${cls}">${escHtml(status)}</span>`;
  }

  /* ── Toggle disable / enable ───────────────────── */
  window.toggleUser = async (userId, currentStatus, btn) => {
    const newStatus  = currentStatus === 'active' ? 'blocked' : 'active';
    const origText   = btn.textContent;
    btn.disabled     = true;
    btn.textContent  = '…';

    try {
      const res = await fetch(`/api/v1/users/${userId}/status`, {
        method:  'POST',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify({ status: newStatus }),
      });
      if (!res.ok) throw new Error((await res.json()).message || 'Request failed');
      showToast(`User ${newStatus === 'active' ? 'enabled' : 'blocked'} successfully`, 'success');
      loadUsers();
    } catch (err) {
      btn.disabled    = false;
      btn.textContent = origText;
      showToast(err.message || 'Failed to update user status', 'error');
    }
  };

  /* ── User edit modal ───────────────────────────── */
  window.openUserModal = (userId, email, role) => {
    document.getElementById('user-modal-email').textContent  = email;
    document.getElementById('user-modal-role').value         = role;
    document.getElementById('user-modal-password').value     = '';
    document.getElementById('user-modal-pw-confirm').value   = '';
    document.getElementById('user-modal-error').textContent  = '';
    document.getElementById('user-modal-overlay').dataset.userId = userId;
    document.getElementById('user-modal-overlay').classList.add('open');
  };

  window.closeUserModal = () => {
    document.getElementById('user-modal-overlay').classList.remove('open');
  };

  document.getElementById('user-modal-overlay').addEventListener('click', e => {
    if (e.target === document.getElementById('user-modal-overlay')) closeUserModal();
  });

  document.getElementById('user-modal-save').addEventListener('click', async () => {
    const overlay  = document.getElementById('user-modal-overlay');
    const userId   = overlay.dataset.userId;
    const role     = document.getElementById('user-modal-role').value;
    const password = document.getElementById('user-modal-password').value.trim();
    const confirm  = document.getElementById('user-modal-pw-confirm').value.trim();
    const errEl    = document.getElementById('user-modal-error');
    const saveBtn  = document.getElementById('user-modal-save');

    errEl.textContent = '';

    if (password && password !== confirm) {
      errEl.textContent = 'Passwords do not match.';
      return;
    }

    if (password && password.length < 8) {
      errEl.textContent = 'Password must be at least 8 characters.';
      return;
    }

    const body = { role };
    if (password) body.password = password;

    saveBtn.disabled    = true;
    saveBtn.textContent = 'Saving…';

    try {
      const res = await fetch(`/api/v1/users/${userId}`, {
        method:  'PATCH',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify(body),
      });
      if (!res.ok) throw new Error((await res.json()).message || 'Update failed');
      closeUserModal();
      showToast('User updated successfully', 'success');
      loadUsers();
    } catch (err) {
      errEl.textContent = err.message || 'Something went wrong.';
    } finally {
      saveBtn.disabled    = false;
      saveBtn.textContent = 'Save Changes';
    }
  });


  /* ══════════════════════════════════════════════════
     SUBSCRIPTION
  ══════════════════════════════════════════════════ */
  async function loadSubscription() {
    const area = document.getElementById('content-area');
    area.innerHTML = `
      <div class="section-header">
        <div>
          <p class="section-sub">Billing &amp; Usage</p>
          <h2>SUBSCRIPTION</h2>
        </div>
      </div>
      <div id="sub-body"><div class="state-loading">Loading</div></div>`;

    try {
      const res  = await fetch('/billing/subscription', { headers: authHeader });
      const data = await res.json();
      renderSubscription(data);
    } catch {
      document.getElementById('sub-body').innerHTML =
        '<div class="state-empty">Failed to load subscription</div>';
    }
  }

  function renderSubscription(d) {
    const used    = d.CurrentLimit  ?? 0;
    const limit   = d.PlanLimit     ?? 1;
    const pct     = Math.min(100, Math.round((used / limit) * 100));

    // Colour ramp: green → amber → red
    let usageColor, usageBg;
    if (pct < 60) {
      usageColor = 'var(--status-completed)';
      usageBg    = 'rgba(168,240,184,0.12)';
    } else if (pct < 85) {
      usageColor = 'var(--status-pending)';
      usageBg    = 'rgba(240,230,168,0.12)';
    } else {
      usageColor = 'var(--status-failed)';
      usageBg    = 'rgba(240,168,168,0.12)';
    }

    const statusCls = d.Status === 'active' ? 'badge-COMPLETED' : 'badge-FAILED';

    document.getElementById('sub-body').innerHTML = `
      <div class="sub-layout">

        <!-- ── Plan hero ── -->
        <div class="sub-hero">
          <div class="sub-hero-left">
            <p class="sub-hero-label">Current Plan</p>
            <h3 class="sub-hero-name">${escHtml(d.PlanName)}</h3>
            <span class="badge ${statusCls}" style="margin-top:10px">${escHtml(d.Status)}</span>
          </div>
          <div class="sub-hero-right">
            <p class="sub-hero-label">Subscription ID</p>
            <p class="sub-hero-id">${escHtml(d.StripeSubscriptionID || '—')}</p>
          </div>
        </div>

        <!-- ── Usage bar card ── -->
        <div class="sub-card sub-usage-card">
          <div class="sub-usage-header">
            <div>
              <p class="sub-card-label">Monthly Usage</p>
              <p class="sub-usage-numbers">
                <span class="sub-usage-used" style="color:${usageColor}">${used.toLocaleString()}</span>
                <span class="sub-usage-sep">/</span>
                <span class="sub-usage-limit">${limit.toLocaleString()} jobs</span>
              </p>
            </div>
            <div class="sub-pct-badge" style="color:${usageColor};background:${usageBg}">
              ${pct}%
            </div>
          </div>
          <div class="sub-bar-track">
            <div class="sub-bar-fill" style="width:${pct}%;background:${usageColor}"></div>
          </div>
          <p class="sub-bar-caption" style="color:${usageColor}">
            ${limit - used > 0
              ? `${(limit - used).toLocaleString()} jobs remaining this period`
              : 'Monthly limit reached'}
          </p>
        </div>

        <!-- ── Meta grid ── -->
        <div class="sub-meta-grid">
          <div class="sub-meta-cell">
            <p class="sub-card-label">Plan Limit</p>
            <p class="sub-meta-val">${limit.toLocaleString()} <span>jobs / mo</span></p>
          </div>
          <div class="sub-meta-cell">
            <p class="sub-card-label">Jobs Used</p>
            <p class="sub-meta-val" style="color:${usageColor}">${used.toLocaleString()}</p>
          </div>
          <div class="sub-meta-cell">
            <p class="sub-card-label">Period Start</p>
            <p class="sub-meta-val">${fmt(d.CurrentPeriodStart)}</p>
          </div>
          <div class="sub-meta-cell">
            <p class="sub-card-label">Period End</p>
            <p class="sub-meta-val">${fmt(d.CurrentPeriodEnd)}</p>
          </div>
          <div class="sub-meta-cell">
            <p class="sub-card-label">App ID</p>
            <p class="sub-meta-val sub-meta-mono">${escHtml(d.AppID || '—')}</p>
          </div>
          <div class="sub-meta-cell">
            <p class="sub-card-label">Subscribed On</p>
            <p class="sub-meta-val">${fmt(d.CreatedAt)}</p>
          </div>
        </div>

      </div>`;

    // Animate bar in after render
    requestAnimationFrame(() => {
      const fill = document.querySelector('.sub-bar-fill');
      if (fill) { fill.style.transition = 'width 0.9s cubic-bezier(0.16,1,0.3,1)'; }
    });
  }


  /* ── Create user modal ─────────────────────────── */
  window.openCreateUserModal = () => {
    document.getElementById('create-user-email').value    = '';
    document.getElementById('create-user-password').value = '';
    document.getElementById('create-user-role').value     = 'viewer';
    document.getElementById('create-user-appid').value    = '';
    document.getElementById('create-user-error').textContent = '';
    document.getElementById('create-user-overlay').classList.add('open');
  };

  window.closeCreateUserModal = () => {
    document.getElementById('create-user-overlay').classList.remove('open');
  };

  document.getElementById('create-user-overlay').addEventListener('click', e => {
    if (e.target === document.getElementById('create-user-overlay')) closeCreateUserModal();
  });

  document.getElementById('create-user-save').addEventListener('click', async () => {
    const email    = document.getElementById('create-user-email').value.trim();
    const password = document.getElementById('create-user-password').value.trim();
    const role     = document.getElementById('create-user-role').value;
    const appId    = document.getElementById('create-user-appid').value.trim();
    const errEl    = document.getElementById('create-user-error');
    const saveBtn  = document.getElementById('create-user-save');

    errEl.textContent = '';
    if (!email)    { errEl.textContent = 'Email is required.';    return; }
    if (!password) { errEl.textContent = 'Password is required.'; return; }
    if (password.length < 6) { errEl.textContent = 'Password must be at least 6 characters.'; return; }

    saveBtn.disabled    = true;
    saveBtn.textContent = 'Creating…';

    try {
      const body = { email, password, role };
      if (appId) body.app_id = appId;

      const res = await fetch('/api/v1/users', {
        method:  'POST',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify(body),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
      closeCreateUserModal();
      showToast('User created successfully', 'success');
      loadUsers();
    } catch (err) {
      errEl.textContent = err.message || 'Failed to create user.';
    } finally {
      saveBtn.disabled    = false;
      saveBtn.textContent = 'Create User';
    }
  });

  /* ── Toast ─────────────────────────────────────── */
  function showToast(msg, type = 'success') {
    let toast = document.getElementById('app-toast');
    if (!toast) {
      toast    = document.createElement('div');
      toast.id = 'app-toast';
      document.body.appendChild(toast);
    }
    toast.className   = `plan-toast plan-toast--${type}`;
    toast.textContent = msg;
    toast.classList.add('plan-toast--visible');
    clearTimeout(toast._t);
    toast._t = setTimeout(() => toast.classList.remove('plan-toast--visible'), 3500);
  }

  /* ── Logout ────────────────────────────────────── */
  document.getElementById('btn-logout').addEventListener('click', () => {
    document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
    localStorage.clear();
    sessionStorage.clear();
    window.location.href = '/login';
  });

  /* ── Boot ──────────────────────────────────────── */
  loadJobs(0);
})();
