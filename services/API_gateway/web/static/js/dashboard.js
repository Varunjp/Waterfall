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
  let currentUserRole      = null; // decoded from token
  let jobsAutoRefreshTimer = null; // 3-second auto-refresh while on jobs tab

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
      // Stop monitor polling when leaving monitor
      if (btn.dataset.section !== 'monitor') {
        clearInterval(monitorRefreshTimer);
        destroyCharts();
      }
      // Stop jobs auto-refresh when leaving jobs tab
      if (btn.dataset.section !== 'jobs') {
        clearInterval(jobsAutoRefreshTimer);
        jobsAutoRefreshTimer = null;
      }
      loadSection(currentSection);
    });
  });

  function loadSection(section) {
    const area = document.getElementById('content-area');
    area.style.animation = 'none';
    requestAnimationFrame(() => {
      area.style.animation = '';
      if (section === 'jobs')              loadJobs(0);
      else if (section === 'plans')        loadPlans();
      else if (section === 'users')        loadUsers();
      else if (section === 'subscription') loadSubscription();
      else if (section === 'monitor')      loadMonitor();
      else                                 renderPlaceholder(section);
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

    // Start 3-second silent auto-refresh while on jobs tab
    startJobsAutoRefresh();
  }

  function startJobsAutoRefresh() {
    clearInterval(jobsAutoRefreshTimer);
    jobsAutoRefreshTimer = setInterval(async () => {
      if (currentSection !== 'jobs') return;
      // Don't refresh if a modal is open
      if (document.getElementById('modal-overlay').classList.contains('open')) return;
      await silentRefreshJobs();
    }, 3000);
  }

  async function silentRefreshJobs() {
    try {
      const params = new URLSearchParams({ limit: LIMIT, offset: currentPage });
      if (currentFilter !== 'ALL')  params.set('status',    currentFilter);
      if (currentStartDate)         params.set('startDate', currentStartDate);
      if (currentEndDate)           params.set('endDate',   currentEndDate);
      const res  = await fetch(`/api/v1/jobs?${params}`, { headers: authHeader });
      const data = await res.json();
      const jobs = data.jobs || [];
      totalJobs  = data.total || jobs.length;
      renderStats(jobs);
      renderTable(jobs);
      renderPagination(currentPage, totalJobs);
    } catch { /* silent — don't disrupt the user */ }
  }

  function renderStats(jobs) {
    const c = { COMPLETED: 0, FAILED: 0, CANCELED: 0, CANCELLED: 0, PENDING: 0, RUNNING: 0, SCHEDULED: 0, QUEUED: 0 };
    jobs.forEach(j => { if (c[j.status] !== undefined) c[j.status]++; });

    const succeeded = c.COMPLETED;
    const failed    = c.FAILED + c.CANCELED + c.CANCELLED;
    const active    = c.PENDING + c.RUNNING + c.SCHEDULED + c.QUEUED;

    document.getElementById('stats-row').innerHTML = `
      <div class="stat-card">
        <p class="stat-label">Total Shown</p>
        <p class="stat-value">${jobs.length}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Completed</p>
        <p class="stat-value success">${succeeded}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Failed</p>
        <p class="stat-value failed">${failed}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Pending / Running</p>
        <p class="stat-value">${active}</p>
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

  /* Decode JWT payload — cached after first call */
  let _jwtCache = null;
  function decodeJwt() {
    if (_jwtCache) return _jwtCache;
    try {
      _jwtCache = JSON.parse(atob(token.split('.')[1]));
    } catch { _jwtCache = {}; }
    return _jwtCache;
  }

  function getMyRole() {
    if (currentUserRole) return currentUserRole;
    const p = decodeJwt();
    currentUserRole = p.role || p.Role || '';
    return currentUserRole;
  }

  function getMyAppId() {
    const p = decodeJwt();
    // Common JWT claim names for app_id
    return p.app_id || p.appId || p.AppId || p.app || p.sub || '';
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

  /* ══════════════════════════════════════════════════
     MONITOR  (Week 5: Charts, Workers, DLQ, Load Test)
  ══════════════════════════════════════════════════ */

  let monitorRefreshTimer = null;
  let loadTestTimer       = null;
  let loadTestRunning     = false;
  let chartInstances      = {};

  function destroyCharts() {
    Object.values(chartInstances).forEach(c => { try { c.destroy(); } catch {} });
    chartInstances = {};
  }

  async function loadMonitor() {
    destroyCharts();
    clearInterval(monitorRefreshTimer);

    const area = document.getElementById('content-area');
    area.innerHTML = `
      <div class="section-header">
        <div>
          <h2>MONITOR</h2>
        </div>
        <div style="display:flex;align-items:center;gap:12px">
          <span id="mon-generated-at" style="font-size:9px;letter-spacing:0.12em;text-transform:uppercase;color:var(--gray-dim)"></span>
          <button class="btn-monitor-refresh" id="btn-mon-refresh">
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" width="11" height="11">
              <path d="M13.5 8A5.5 5.5 0 1 1 10 3.07"/><path d="M10 1v3h3"/>
            </svg>
            Refresh
          </button>
        </div>
      </div>
      <div id="monitor-body"><div class="state-loading">Loading</div></div>`;

    document.getElementById('btn-mon-refresh').addEventListener('click', () => {
      refreshMonitor();
    });

    await renderMonitorShell();
    await refreshMonitor();

    // Auto-refresh every 15s
    monitorRefreshTimer = setInterval(refreshMonitor, 15000);
  }

  function renderMonitorShell() {
    document.getElementById('monitor-body').innerHTML = `

      <!-- ── Row 1: Key metrics ── -->
      <div class="stats-row" style="grid-template-columns:repeat(3,1fr);margin-bottom:16px">
        <div class="stat-card">
          <p class="stat-label">Jobs Created</p>
          <p class="stat-value" id="mon-created">—</p>
        </div>
        <div class="stat-card">
          <p class="stat-label">Succeeded</p>
          <p class="stat-value success" id="mon-success">—</p>
        </div>
        <div class="stat-card">
          <p class="stat-label">Failed</p>
          <p class="stat-value failed" id="mon-failed">—</p>
        </div>
      </div>

      <!-- ── Row 2: Status chart + Queue Lag + Worker count ── -->
      <div class="monitor-grid" style="grid-template-columns:1.2fr 0.8fr 1fr;gap:16px;margin-bottom:16px">

        <!-- Status donut chart -->
        <div class="monitor-card">
          <p class="monitor-card-label">
            Job Status Distribution
            <span class="live-pill">Live</span>
          </p>
          <div class="chart-container">
            <canvas id="chart-status"></canvas>
          </div>
        </div>

        <!-- Queue lag -->
        <div class="monitor-card">
          <p class="monitor-card-label">Queue Lag <span style="font-size:8px;letter-spacing:0.1em;color:var(--gray-dim)">(oldest ready job)</span></p>
          <div class="monitor-metric">
            <span class="monitor-metric-val" id="mon-lag-val">—</span>
            <span class="monitor-metric-unit" id="mon-lag-unit">ms</span>
          </div>
          <p class="monitor-metric-delta" id="mon-lag-status">—</p>
          <div style="margin-top:20px">
            <p class="monitor-card-label" style="margin-bottom:10px">Depth by Queue</p>
            <div class="queue-bar-list" id="queue-bar-list">
              <div style="color:var(--gray-dim);font-size:10px;letter-spacing:0.1em">Loading…</div>
            </div>
          </div>
        </div>

        <!-- Worker count -->
        <div class="monitor-card">
          <p class="monitor-card-label">Workers</p>
          <div class="worker-count-summary" id="worker-summary">
            <div class="worker-count-item">
              <span class="worker-count-num busy" id="wc-busy">—</span>
              <span style="font-size:9px;letter-spacing:0.12em;text-transform:uppercase;color:var(--status-running)">Busy</span>
            </div>
            <div class="worker-count-item">
              <span class="worker-count-num idle" id="wc-idle">—</span>
              <span style="font-size:9px;letter-spacing:0.12em;text-transform:uppercase;color:var(--gray-mid)">Idle</span>
            </div>
            <div class="worker-count-item">
              <span class="worker-count-num offline" id="wc-offline">—</span>
              <span style="font-size:9px;letter-spacing:0.12em;text-transform:uppercase;color:var(--status-failed)">Offline</span>
            </div>
          </div>
          <div class="worker-list" id="worker-list">
            <div style="color:var(--gray-dim);font-size:10px;letter-spacing:0.1em">Loading…</div>
          </div>
        </div>
      </div>

      <!-- ── Row 3: Throughput chart (full width) ── -->
      <div class="monitor-grid--wide">
        <div class="monitor-card">
          <p class="monitor-card-label">
            Job Throughput — Today (UTC)
            <span class="live-pill">Live</span>
          </p>
          <div class="chart-container--tall">
            <canvas id="chart-throughput"></canvas>
          </div>
        </div>
      </div>

      <!-- ── Row 4: Job Workflow full width ── -->
      <div class="monitor-grid--wide" style="margin-bottom:16px">
        <div class="monitor-card">
          <p class="monitor-card-label">Job Workflow</p>
          <div class="workflow-canvas" style="height:240px">
            <svg class="workflow-svg" id="workflow-svg" viewBox="0 0 700 210" xmlns="http://www.w3.org/2000/svg"></svg>
          </div>
        </div>
      </div>

      <!-- ── Row 5: DLQ management ── -->
      <div class="monitor-grid--wide" style="margin-bottom:16px">
        <div class="monitor-card">
          <div class="dlq-toolbar">
            <p class="monitor-card-label" style="margin-bottom:0">
              Failed Jobs
              <span class="dlq-count-badge" id="dlq-count-badge">— jobs</span>
            </p>
          </div>
          <div id="dlq-body"><div class="dlq-empty">Loading…</div></div>
          <div id="dlq-pagination" style="display:none;padding:12px 0 0 0;display:flex;align-items:center;justify-content:space-between">
            <span class="pagination-info" id="dlq-page-info">—</span>
            <div class="pagination-btns">
              <button class="btn-page" id="btn-dlq-prev" disabled>← Prev</button>
              <button class="btn-page" id="btn-dlq-next" disabled>Next →</button>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Row 6: Load / Spike testing ── -->
      <div class="monitor-grid--wide">
        <div class="monitor-card">
          <p class="monitor-card-label">Load &amp; Spike Simulator</p>

          <div class="load-test-controls">
            <div class="load-field-group">
              <label class="load-field-label">Job Count</label>
              <input type="number" class="load-field-input" id="lt-count" value="100" min="1" max="10000"/>
            </div>
            <div class="load-field-group">
              <label class="load-field-label">Concurrency</label>
              <input type="number" class="load-field-input" id="lt-concurrency" value="10" min="1" max="100"/>
            </div>
            <div class="load-field-group">
              <label class="load-field-label">Job Type</label>
              <input type="text" class="load-field-input" id="lt-type" value="load_test" placeholder="email, webhook…"/>
            </div>
            <div class="load-field-group">
              <label class="load-field-label">Mode</label>
              <select class="load-field-input filter-select" id="lt-mode">
                <option value="steady">Steady</option>
                <option value="spike">Spike Burst</option>
                <option value="backpressure">Backpressure</option>
              </select>
            </div>
          </div>

          <div class="load-test-actions">
            <button class="btn-load-run" id="btn-lt-run">▶ Run Simulation</button>
            <button class="btn-load-stop" id="btn-lt-stop" disabled>■ Stop</button>
          </div>

          <div class="load-status-bar" id="lt-status-bar">
            <span>Ready — configure and click Run</span>
            <span id="lt-elapsed"></span>
          </div>
          <div class="load-progress-track">
            <div class="load-progress-fill" id="lt-progress"></div>
          </div>

          <div class="load-result-grid" id="lt-results">
            <div class="load-result-cell"><p class="load-result-key">Dispatched</p><p class="load-result-val" id="lt-r-dispatched">—</p></div>
            <div class="load-result-cell"><p class="load-result-key">Accepted</p><p class="load-result-val" id="lt-r-accepted" style="color:var(--status-completed)">—</p></div>
            <div class="load-result-cell"><p class="load-result-key">Rejected</p><p class="load-result-val" id="lt-r-rejected" style="color:var(--status-failed)">—</p></div>
            <div class="load-result-cell"><p class="load-result-key">Throughput</p><p class="load-result-val" id="lt-r-tput">— /s</p></div>
          </div>
        </div>
      </div>`;

    renderWorkflowDiagram();
    bindLoadTestActions();
  }

  /* ── Refresh all live data — uses /api/v1/runtime/overview as source of truth ── */
  async function refreshMonitor() {
    const btn = document.getElementById('btn-mon-refresh');
    if (btn) btn.classList.add('spinning');
    try {
      // Always fetch today's full UTC day: 00:00:00Z → 23:59:59Z
      // This matches the API format: /api/v1/jobs/metrics?from=2026-04-04T00:00:00Z&to=2026-04-04T23:59:59Z&bucket=hour
      const nowUTC      = new Date();
      const todayUTC    = nowUTC.toISOString().slice(0, 10); // "YYYY-MM-DD"
      const metricsFrom = `${todayUTC}T00:00:00Z`;
      const metricsTo   = `${todayUTC}T23:59:59Z`;

      const [overview, jobsData, stats, metrics] = await Promise.all([
        fetch('/api/v1/runtime/overview', { headers: authHeader })
          .then(r => r.ok ? r.json() : null)
          .catch(() => null),
        fetch('/api/v1/jobs?limit=50&offset=0', { headers: authHeader })
          .then(r => r.json())
          .catch(() => ({ jobs: [], total: 0 })),
        fetch('/api/v1/jobs/stats', { headers: authHeader })
          .then(r => r.ok ? r.json() : null)
          .catch(() => null),
        fetch(`/api/v1/jobs/metrics?from=${metricsFrom}&to=${metricsTo}&bucket=hour`, { headers: authHeader })
          .then(r => r.ok ? r.json() : null)
          .catch(() => null),
      ]);
      applyOverviewMetrics(overview, jobsData, stats, metrics);
      await fetchDLQ();
    } finally {
      if (btn) btn.classList.remove('spinning');
    }
  }

  /* ── Map runtime/overview + jobs list + stats → all monitor panels ── */
  function applyOverviewMetrics(ov, jobsData, stats, metrics) {
    const jobs = jobsData?.jobs || [];

    // ── 1. Top stat cards — use /api/v1/jobs/stats when available ──
    // Stats API returns: { totalJobs, totalSuccessJobs, totalFailedJobs }
    const created = stats?.totalJobs        ?? (jobsData?.total || jobs.length);
    const success = stats?.totalSuccessJobs ?? jobs.filter(j => j.status === 'COMPLETED').length;
    const failed  = stats?.totalFailedJobs  ?? jobs.filter(j => j.status === 'FAILED' || j.status === 'CANCELED' || j.status === 'CANCELLED').length;

    setText('mon-created', created.toLocaleString());
    setText('mon-success', success.toLocaleString());
    setText('mon-failed',  failed.toLocaleString());

    // ── Still compute per-status counts from jobs list for charts/lag/workers ──
    let pending = 0, running = 0, scheduled = 0, queued = 0, cancelled = 0, successCount = 0, failedCount = 0;
    jobs.forEach(j => {
      if (j.status === 'COMPLETED')                              successCount++;
      else if (j.status === 'FAILED')                            failedCount++;
      else if (j.status === 'CANCELLED' || j.status === 'CANCELED') cancelled++;
      else if (j.status === 'PENDING')                           pending++;
      else if (j.status === 'RUNNING')                           running++;
      else if (j.status === 'SCHEDULED')                         scheduled++;
      else if (j.status === 'QUEUED')                            queued++;
    });

    // ── 2. Queue lag — oldestReadyAgeSeconds from runtime queues ──
    const queues = ov?.queues || [];
    const maxAge = queues.reduce(
      (max, q) => Math.max(max, parseInt(q.oldestReadyAgeSeconds || 0)), 0
    );
    const lagEl = document.getElementById('mon-lag-val');
    const lagSt = document.getElementById('mon-lag-status');
    // Swap unit label from ms → s since API gives seconds
    const lagUnit = document.getElementById('mon-lag-unit');
    if (lagUnit) lagUnit.textContent = 's';
    if (lagEl) lagEl.textContent = maxAge;
    if (lagSt) {
      if (maxAge === 0) {
        lagSt.textContent = '↓ No backlog';
        lagSt.className   = 'monitor-metric-delta up';
        lagSt.style.color = '';
      } else if (maxAge < 30) {
        lagSt.textContent = '↓ Healthy';
        lagSt.className   = 'monitor-metric-delta up';
        lagSt.style.color = '';
      } else if (maxAge < 120) {
        lagSt.textContent = '⚠ Elevated';
        lagSt.className   = 'monitor-metric-delta';
        lagSt.style.color = 'var(--status-pending)';
      } else {
        lagSt.textContent = '↑ High lag';
        lagSt.className   = 'monitor-metric-delta down';
        lagSt.style.color = '';
      }
    }

    // ── 3. Per-queue depth bars (one row per jobType) ──
    const barEl = document.getElementById('queue-bar-list');
    if (barEl) {
      if (queues.length) {
        const maxBar = Math.max(
          ...queues.map(q => parseInt(q.readyJobs || 0) + parseInt(q.runningJobs || 0)), 1
        );
        barEl.innerHTML = queues.map(q => {
          const ready   = parseInt(q.readyJobs    || 0);
          const run     = parseInt(q.runningJobs  || 0);
          const workers = parseInt(q.registeredWorkers || 0);
          const busy    = parseInt(q.busyWorkers  || 0);
          const total   = ready + run;
          const pct     = Math.round((total / maxBar) * 100);
          const color   = run > 0
            ? 'var(--status-running)'
            : ready > 0 ? 'var(--status-pending)' : 'var(--gray-dim)';
          return `
            <div class="queue-bar-item" title="${escHtml(q.jobType)}: ${ready} ready · ${run} running · ${workers} workers (${busy} busy)">
              <span class="queue-bar-label" style="width:90px;font-size:10px;overflow:hidden;text-overflow:ellipsis">${escHtml(q.jobType)}</span>
              <div class="queue-bar-track">
                <div class="queue-bar-fill" style="width:${pct}%;background:${color}"></div>
              </div>
              <span class="queue-bar-val" style="color:${color}">${total}</span>
            </div>`;
        }).join('');
      } else {
        // Fallback to job-list counts when no queue data
        const bars = [
          { label: 'Pending',   val: pending + queued, color: 'var(--status-pending)' },
          { label: 'Running',   val: running,           color: 'var(--status-running)' },
          { label: 'Scheduled', val: scheduled,         color: 'var(--gray-mid)' },
          { label: 'Failed',    val: failed,            color: 'var(--status-failed)' },
        ];
        const maxBar = Math.max(...bars.map(b => b.val), 1);
        barEl.innerHTML = bars.map(b => `
          <div class="queue-bar-item">
            <span class="queue-bar-label">${b.label}</span>
            <div class="queue-bar-track">
              <div class="queue-bar-fill" style="width:${Math.round((b.val/maxBar)*100)}%;background:${b.color}"></div>
            </div>
            <span class="queue-bar-val" style="color:${b.color}">${b.val}</span>
          </div>`).join('');
      }
    }

    // ── 4. Workers — from ov.workers array (runtime truth) ──
    if (ov) {
      const totalW  = parseInt(ov.totalWorkers  || 0);
      const onlineW = parseInt(ov.onlineWorkers || 0);
      const busyW   = parseInt(ov.busyWorkers   || 0);
      setText('wc-busy',    busyW);
      setText('wc-idle',    Math.max(0, onlineW - busyW));
      setText('wc-offline', Math.max(0, totalW  - onlineW));
      renderWorkersFromOverview(ov.workers || [], ov.queues || []);
    }

    // ── 5. generatedAt timestamp ──
    const genEl = document.getElementById('mon-generated-at');
    if (genEl && ov?.generatedAt) {
      const d = new Date(parseInt(ov.generatedAt) * 1000);
      genEl.textContent = `Updated ${d.toLocaleTimeString('en-GB', {
        hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false,
      })}`;
    }

    // ── 6. Charts ──
    drawStatusChart({ success, failed, pending: pending + queued, running, scheduled, cancelled });
    drawThroughputChart(metrics);
  }

  function setText(id, val) {
    const el = document.getElementById(id);
    if (el) el.textContent = val;
  }

  /* ── Status donut chart ── */
  function drawStatusChart({ success, failed, pending, running, scheduled, cancelled }) {
    const canvas = document.getElementById('chart-status');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    if (chartInstances['status']) { chartInstances['status'].destroy(); }

    chartInstances['status'] = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: ['Completed', 'Failed', 'Pending/Queued', 'Running', 'Scheduled', 'Cancelled'],
        datasets: [{
          data: [success, failed, pending, running, scheduled, cancelled],
          backgroundColor: [
            'rgba(168,240,184,0.85)',
            'rgba(240,168,168,0.85)',
            'rgba(240,230,168,0.85)',
            'rgba(168,212,240,0.85)',
            'rgba(200,200,200,0.4)',
            'rgba(120,120,120,0.4)',
          ],
          borderColor: 'rgba(15,15,15,0.6)',
          borderWidth: 2,
          hoverOffset: 6,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        cutout: '65%',
        plugins: {
          legend: {
            position: 'right',
            labels: {
              color: '#777770',
              font: { family: 'DM Sans', size: 10 },
              boxWidth: 10,
              padding: 10,
            },
          },
          tooltip: {
            backgroundColor: '#161616',
            borderColor: 'rgba(245,245,240,0.12)',
            borderWidth: 1,
            titleColor: '#f5f5f0',
            bodyColor: '#777770',
            titleFont: { family: 'Bebas Neue', size: 14, letterSpacing: '0.1em' },
          },
        },
      },
    });
  }

  /* ── Throughput line chart — uses GET /api/v1/jobs/metrics?from&to&bucket=hour ── */
  function drawThroughputChart(metrics) {
    const canvas = document.getElementById('chart-throughput');
    if (!canvas) return;
    if (chartInstances['throughput']) chartInstances['throughput'].destroy();

    const buckets = metrics?.buckets || [];

    // Slice "HH:MM" directly from the UTC ISO string — no Date() conversion,
    // so the label is exactly what the DB returned (e.g. "2026-04-04T10:00:00Z" → "10:00")
    const labels    = buckets.map(b => b.ts.slice(11, 16));
    const created   = buckets.map(b => Number(b.created  ?? 0));
    const succeeded = buckets.map(b => Number(b.completed ?? 0));
    const failedArr = buckets.map(b => Number(b.failed    ?? 0));

    const ctx = canvas.getContext('2d');

    if (!buckets.length) {
      chartInstances['throughput'] = new Chart(ctx, {
        type: 'line',
        data: { labels: ['No data'], datasets: [{ label: 'No data', data: [0], borderColor: 'rgba(245,245,240,0.2)' }] },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } } },
      });
      return;
    }

    chartInstances['throughput'] = new Chart(ctx, {
      type: 'line',
      data: {
        labels,
        datasets: [
          {
            label: 'Created',
            data: created,
            borderColor: 'rgba(245,245,240,0.6)',
            backgroundColor: 'rgba(245,245,240,0.04)',
            fill: true, tension: 0.3,
            pointRadius: 4, pointHoverRadius: 6,
            pointBackgroundColor: 'rgba(245,245,240,0.8)',
            borderWidth: 1.5,
            spanGaps: false,
          },
          {
            label: 'Completed',
            data: succeeded,
            borderColor: 'rgba(168,240,184,0.9)',
            backgroundColor: 'rgba(168,240,184,0.07)',
            fill: true, tension: 0.3,
            pointRadius: 4, pointHoverRadius: 6,
            pointBackgroundColor: 'rgba(168,240,184,0.9)',
            borderWidth: 1.5,
            spanGaps: false,
          },
          {
            label: 'Failed',
            data: failedArr,
            borderColor: 'rgba(240,168,168,0.9)',
            backgroundColor: 'rgba(240,168,168,0.07)',
            fill: true, tension: 0.3,
            pointRadius: 4, pointHoverRadius: 6,
            pointBackgroundColor: 'rgba(240,168,168,0.9)',
            borderWidth: 1.5,
            spanGaps: false,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: {
            labels: {
              color: '#777770',
              font: { family: 'DM Sans', size: 10 },
              boxWidth: 10, padding: 14,
            },
          },
          tooltip: {
            backgroundColor: '#161616',
            borderColor: 'rgba(245,245,240,0.12)',
            borderWidth: 1,
            titleColor: '#f5f5f0',
            bodyColor: '#777770',
            callbacks: {
              title: items => `${items[0].label} UTC`,
            },
          },
        },
        scales: {
          x: {
            grid: { color: 'rgba(245,245,240,0.04)' },
            ticks: { color: '#777770', font: { size: 9 }, maxTicksLimit: 13 },
          },
          y: {
            beginAtZero: true,
            grid: { color: 'rgba(245,245,240,0.04)' },
            ticks: { color: '#777770', font: { size: 9 }, precision: 0 },
          },
        },
      },
    });
  }


  /* ── Workers ── */
  /* ── renderWorkersFromOverview — called by applyOverviewMetrics ── */
  // Maps ov.workers (runtime schema) to the worker list panel.
  // Each worker from the API looks like:
  //   { workerId, jobTypes[], activeJobs, maxConcurrency, lastSeen, status }
  // status values: WORKER_STATUS_ONLINE | WORKER_STATUS_OFFLINE | WORKER_STATUS_BUSY
  function renderWorkersFromOverview(workers, queues) {
    const listEl = document.getElementById('worker-list');
    if (!listEl) return;

    if (!workers.length) {
      listEl.innerHTML = '<div style="color:var(--gray-dim);font-size:10px;letter-spacing:0.1em">No workers registered</div>';
      return;
    }

    listEl.innerHTML = workers.map(w => {
      // Normalise status string → simple key
      const rawStatus = (w.status || '').toUpperCase();
      let dotClass, labelClass, labelText;
      if (rawStatus.includes('OFFLINE')) {
        dotClass = 'offline'; labelClass = 'offline'; labelText = 'Offline';
      } else if (w.activeJobs > 0 || rawStatus.includes('BUSY')) {
        dotClass = 'busy';    labelClass = 'busy';    labelText = 'Busy';
      } else {
        dotClass = 'idle';    labelClass = 'idle';    labelText = 'Idle';
      }

      // Last seen: convert unix timestamp → human time
      const lastSeenTs = w.lastSeen ? new Date(parseInt(w.lastSeen) * 1000) : null;
      const lastSeenStr = lastSeenTs && !isNaN(lastSeenTs)
        ? lastSeenTs.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false })
        : '—';

      const jobTypesStr = (w.jobTypes || []).join(', ') || '—';
      const concLabel   = w.activeJobs != null && w.maxConcurrency != null
        ? `${w.activeJobs}/${w.maxConcurrency}`
        : '—';

      return `
        <div class="worker-row" title="Last seen: ${lastSeenStr} · Jobs: ${concLabel} · Types: ${escHtml(jobTypesStr)}">
          <div class="worker-status-dot ${dotClass}"></div>
          <span class="worker-id" title="${escHtml(w.workerId)}">${escHtml(w.workerId)}</span>
          <span class="worker-label ${labelClass}">${labelText}</span>
          <span class="worker-job" title="${escHtml(jobTypesStr)}">${escHtml(jobTypesStr)}</span>
          <span style="font-size:10px;color:var(--gray-mid);margin-left:auto;flex-shrink:0">${concLabel}</span>
        </div>`;
    }).join('');
  }

  /* ── DLQ ── */
  /* ── DLQ: paginated failed jobs from /api/v1/jobs?status=FAILED ── */
  let dlqPage  = 0;
  const DLQ_LIMIT = 10;
  let dlqTotal = 0;

  async function fetchDLQ(page) {
    if (page === undefined) page = dlqPage;
    dlqPage = page;

    const dlqBody = document.getElementById('dlq-body');
    if (!dlqBody) return;
    dlqBody.innerHTML = '<div class="dlq-empty" style="padding:24px">Loading…</div>';

    try {
      const params = new URLSearchParams({ status: 'FAILED', limit: DLQ_LIMIT, offset: page });
      const res  = await fetch(`/api/v1/jobs?${params}`, { headers: authHeader });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      const jobs = data.jobs || [];
      dlqTotal   = data.total || jobs.length;
      renderDLQ(jobs, page, dlqTotal);
    } catch (err) {
      dlqBody.innerHTML = `<div class="dlq-empty">Failed to load: ${escHtml(err.message)}</div>`;
    }
  }

  function renderDLQ(jobs, page, total) {
    const badge   = document.getElementById('dlq-count-badge');
    const dlqBody = document.getElementById('dlq-body');
    const pagEl   = document.getElementById('dlq-pagination');
    const infoEl  = document.getElementById('dlq-page-info');
    const prevBtn = document.getElementById('btn-dlq-prev');
    const nextBtn = document.getElementById('btn-dlq-next');
    if (!dlqBody) return;

    if (badge) badge.textContent = `${total} job${total !== 1 ? 's' : ''}`;

    if (!jobs.length) {
      dlqBody.innerHTML = '<div class="dlq-empty">✓ No failed jobs</div>';
      if (pagEl) pagEl.style.display = 'none';
      return;
    }

    dlqBody.innerHTML = `
      <table style="width:100%;border-collapse:collapse">
        <thead>
          <tr style="border-bottom:1px solid var(--border)">
            <th style="padding:8px 14px;font-size:9px;letter-spacing:0.18em;text-transform:uppercase;color:var(--gray-mid);text-align:left;font-weight:500">Job ID</th>
            <th style="padding:8px 14px;font-size:9px;letter-spacing:0.18em;text-transform:uppercase;color:var(--gray-mid);text-align:left;font-weight:500">Type</th>
            <th style="padding:8px 14px;font-size:9px;letter-spacing:0.18em;text-transform:uppercase;color:var(--gray-mid);text-align:left;font-weight:500">Retry</th>
            <th style="padding:8px 14px;font-size:9px;letter-spacing:0.18em;text-transform:uppercase;color:var(--gray-mid);text-align:left;font-weight:500">Failed At</th>
            <th style="padding:8px 14px;font-size:9px;letter-spacing:0.18em;text-transform:uppercase;color:var(--gray-mid);text-align:right;font-weight:500">Action</th>
          </tr>
        </thead>
        <tbody>
          ${jobs.map(j => `
            <tr style="border-bottom:1px solid var(--border)" onmouseover="this.style.background='rgba(245,245,240,0.02)'" onmouseout="this.style.background=''">
              <td style="padding:11px 14px">
                <span class="cell-id" title="${escHtml(j.jobId)}">${escHtml(j.jobId)}</span>
              </td>
              <td style="padding:11px 14px;font-size:12px;color:var(--gray-light)">${escHtml(j.type || '—')}</td>
              <td style="padding:11px 14px;font-size:12px;color:var(--status-failed)">${j.retry ?? 0} / ${j.maxRetry ?? 0}</td>
              <td style="padding:11px 14px;font-size:11px;color:var(--gray-mid)">${fmt(j.updatedAt)}</td>
              <td style="padding:11px 14px;text-align:right">
                <button class="btn-dlq-row-retry" onclick="window._dlqRetry('${escHtml(j.jobId)}', this)">↻ Retry</button>
              </td>
            </tr>`).join('')}
        </tbody>
      </table>`;

    // Pagination
    if (pagEl) pagEl.style.display = 'flex';
    const from = total === 0 ? 0 : page + 1;
    const to   = Math.min(page + DLQ_LIMIT, total);
    if (infoEl) infoEl.textContent = `${from}–${to} of ${total}`;
    if (prevBtn) {
      prevBtn.disabled = page <= 0;
      prevBtn.onclick  = () => fetchDLQ(page - DLQ_LIMIT);
    }
    if (nextBtn) {
      nextBtn.disabled = page + DLQ_LIMIT >= total;
      nextBtn.onclick  = () => fetchDLQ(page + DLQ_LIMIT);
    }
  }

  window._dlqRetry = async (jobId, btn) => {
    const orig = btn.textContent;
    btn.disabled    = true;
    btn.textContent = '…';
    try {
      const res = await fetch(`/api/v1/jobs/${jobId}/retry`, { method: 'POST', headers: authHeader });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      showToast('Job queued for retry', 'success');
      await fetchDLQ(dlqPage);
    } catch (err) {
      showToast(`Retry failed: ${err.message}`, 'error');
      btn.disabled    = false;
      btn.textContent = orig;
    }
  };

  /* ── Workflow visualization ── */
  function renderWorkflowDiagram() {
    const svg = document.getElementById('workflow-svg');
    if (!svg) return;

    const nodes = [
      { id: 'created',   x: 20,  y: 90, label: 'CREATED',   color: '#f5f5f0' },
      { id: 'scheduled', x: 140, y: 40, label: 'SCHEDULED', color: '#e0e0da' },
      { id: 'queued',    x: 260, y: 90, label: 'QUEUED',    color: '#f0e6a8' },
      { id: 'running',   x: 390, y: 90, label: 'RUNNING',   color: '#a8d4f0' },
      { id: 'completed', x: 530, y: 40, label: 'COMPLETED', color: '#a8f0b8' },
      { id: 'failed',    x: 530, y: 140, label: 'FAILED',   color: '#f0a8a8' },
      { id: 'retry',     x: 390, y: 155, label: 'RETRY',    color: '#f0e6a8' },
      { id: 'cancelled', x: 260, y: 155, label: 'CANCELLED', color: '#777770' },
    ];

    const edges = [
      { from: 'created',   to: 'queued',    label: '' },
      { from: 'created',   to: 'scheduled', label: 'sched' },
      { from: 'scheduled', to: 'queued',    label: '' },
      { from: 'queued',    to: 'running',   label: '' },
      { from: 'queued',    to: 'cancelled', label: 'cancel' },
      { from: 'running',   to: 'completed', label: 'ok' },
      { from: 'running',   to: 'failed',    label: 'err' },
      { from: 'failed',    to: 'retry',     label: '' },
      { from: 'retry',     to: 'queued',    label: '' },
    ];

    const nodeMap = {};
    nodes.forEach(n => { nodeMap[n.id] = n; });
    const W = 56; const H = 26;

    let edgesHtml = edges.map(e => {
      const from = nodeMap[e.from];
      const to   = nodeMap[e.to];
      const x1 = from.x + W;
      const y1 = from.y + H / 2;
      const x2 = to.x;
      const y2 = to.y + H / 2;
      const mx = (x1 + x2) / 2;
      const path = `M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}`;
      return `
        <path d="${path}" stroke="rgba(245,245,240,0.15)" stroke-width="1.5" fill="none" marker-end="url(#arrow)"/>
        ${e.label ? `<text x="${mx}" y="${(y1+y2)/2 - 4}" font-size="7" fill="#3a3a38" text-anchor="middle" font-family="DM Sans">${e.label}</text>` : ''}
      `;
    }).join('');

    let nodesHtml = nodes.map(n => `
      <g transform="translate(${n.x},${n.y})">
        <rect x="0" y="0" width="${W}" height="${H}" fill="rgba(15,15,15,0.6)" stroke="${n.color}" stroke-width="1" rx="0"/>
        <text x="${W/2}" y="${H/2+4}" font-size="7.5" fill="${n.color}" text-anchor="middle" font-family="Bebas Neue" letter-spacing="0.08em">${n.label}</text>
      </g>`).join('');

    svg.innerHTML = `
      <defs>
        <marker id="arrow" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
          <path d="M0,0 L6,3 L0,6 Z" fill="rgba(245,245,240,0.25)"/>
        </marker>
      </defs>
      ${edgesHtml}
      ${nodesHtml}`;
  }

  /* ── Load test simulation ── */
  function bindLoadTestActions() {
    const runBtn  = document.getElementById('btn-lt-run');
    const stopBtn = document.getElementById('btn-lt-stop');
    if (!runBtn || !stopBtn) return;

    runBtn.addEventListener('click', startLoadTest);
    stopBtn.addEventListener('click', stopLoadTest);
  }

  function startLoadTest() {
    if (loadTestRunning) return;
    loadTestRunning = true;

    const count       = parseInt(document.getElementById('lt-count').value)       || 100;
    const concurrency = parseInt(document.getElementById('lt-concurrency').value) || 10;
    const type        = document.getElementById('lt-type').value.trim() || 'load_test';
    const mode        = document.getElementById('lt-mode').value;

    const runBtn  = document.getElementById('btn-lt-run');
    const stopBtn = document.getElementById('btn-lt-stop');
    runBtn.disabled  = true;
    stopBtn.disabled = false;

    let dispatched = 0, accepted = 0, rejected = 0;
    const startTime = Date.now();

    const statusBar = document.getElementById('lt-status-bar');
    const progress  = document.getElementById('lt-progress');

    function updateUI() {
      const pct = Math.round((dispatched / count) * 100);
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      const tput  = elapsed > 0 ? (dispatched / elapsed).toFixed(1) : '0';

      if (progress)  progress.style.width = `${pct}%`;
      if (statusBar) {
        const modeLabel = mode === 'spike' ? 'Spike Burst' : mode === 'backpressure' ? 'Backpressure Test' : 'Steady Load';
        statusBar.innerHTML = `
          <span class="load-status-active">▶ ${modeLabel} — ${dispatched}/${count} dispatched (${pct}%)</span>
          <span id="lt-elapsed">${elapsed}s · ${tput} jobs/s</span>`;
      }

      setText('lt-r-dispatched', dispatched);
      setText('lt-r-accepted',   accepted);
      setText('lt-r-rejected',   rejected);
      document.getElementById('lt-r-tput').textContent =
        `${(dispatched / Math.max(0.001, (Date.now() - startTime) / 1000)).toFixed(1)}/s`;
    }

    // Batch config per mode:
    // spike       — send all jobs at once in one burst
    // backpressure — 1 job at a time, 500ms apart (stresses queue consumer)
    // steady       — concurrency jobs per batch, 150ms between batches
    const batchSize  = mode === 'spike' ? count : mode === 'backpressure' ? 1 : concurrency;
    const intervalMs = mode === 'spike' ? 0 : mode === 'backpressure' ? 500 : 150;

    async function dispatchBatch() {
      if (!loadTestRunning || dispatched >= count) {
        finishLoadTest(dispatched, accepted, rejected, Date.now() - startTime);
        return;
      }

      const batch = Math.min(batchSize, count - dispatched);

      // POST to /api/v1/jobs with the correct body shape
      const appId = getMyAppId();
      const promises = Array.from({ length: batch }, async () => {
        try {
          const body = {
            type,
            payload: JSON.stringify({ to: 'test@mail.com', body: 'Hello', source: 'load_test', index: dispatched, mode }),
          };
          if (appId) body.app_id = appId;

          const res = await fetch('/api/v1/jobs', {
            method:  'POST',
            headers: { ...authHeader, 'Content-Type': 'application/json' },
            body:    JSON.stringify(body),
          });
          if (res.ok) { accepted++; }
          else        { rejected++; }
        } catch {
          rejected++;
        }
        dispatched++;
      });

      await Promise.all(promises);
      updateUI();

      if (loadTestRunning && dispatched < count) {
        loadTestTimer = setTimeout(dispatchBatch, intervalMs);
      } else {
        finishLoadTest(dispatched, accepted, rejected, Date.now() - startTime);
      }
    }

    updateUI();
    dispatchBatch();
  }

  function stopLoadTest() {
    loadTestRunning = false;
    clearTimeout(loadTestTimer);
    const runBtn  = document.getElementById('btn-lt-run');
    const stopBtn = document.getElementById('btn-lt-stop');
    if (runBtn)  runBtn.disabled  = false;
    if (stopBtn) stopBtn.disabled = true;
    const statusBar = document.getElementById('lt-status-bar');
    if (statusBar) statusBar.innerHTML = '<span style="color:var(--status-pending)">■ Stopped by user</span><span></span>';
  }

  function finishLoadTest(dispatched, accepted, rejected, elapsedMs) {
    loadTestRunning = false;
    const runBtn  = document.getElementById('btn-lt-run');
    const stopBtn = document.getElementById('btn-lt-stop');
    if (runBtn)  runBtn.disabled  = false;
    if (stopBtn) stopBtn.disabled = true;

    const tput = (dispatched / (elapsedMs / 1000)).toFixed(1);
    const statusBar = document.getElementById('lt-status-bar');
    if (statusBar) {
      statusBar.innerHTML = `
        <span class="load-status-done">✓ Complete — ${dispatched} jobs in ${(elapsedMs/1000).toFixed(1)}s</span>
        <span>${tput} jobs/s avg</span>`;
    }
    const progress = document.getElementById('lt-progress');
    if (progress) {
      progress.style.background = 'var(--status-completed)';
      progress.style.width = '100%';
    }

    setText('lt-r-dispatched', dispatched);
    setText('lt-r-accepted',   accepted);
    setText('lt-r-rejected',   rejected);
    document.getElementById('lt-r-tput').textContent = `${tput}/s`;

    showToast(`Load test complete: ${accepted} accepted, ${rejected} rejected`, accepted > 0 ? 'success' : 'error');
  }

  window.fetchDLQ = fetchDLQ;

  document.getElementById('btn-logout').addEventListener('click', () => {
    document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
    localStorage.clear();
    sessionStorage.clear();
    window.location.href = '/login';
  });


  /* ── Boot ──────────────────────────────────────── */
  loadJobs(0);
})();
