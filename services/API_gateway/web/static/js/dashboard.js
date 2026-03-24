(() => {
  /* ── Auth ──────────────────────────────────────── */
  const token = (() => {
    const m = document.cookie.match(/token=([^;]+)/);
    return m ? m[1] : null;
  })();

  if (!token) { window.location.href = '/login'; }

  const authHeader = { Authorization: 'Bearer ' + token };

  /* ── State ─────────────────────────────────────── */
  let currentPage    = 0;
  let currentSection = 'jobs';
  let currentFilter  = 'ALL';
  let totalJobs      = 0;
  const LIMIT        = 10;

  /* ── Helpers ───────────────────────────────────── */
  function fmt(dateStr) {
    if (!dateStr) return '—';
    const d = new Date(dateStr);
    if (isNaN(d)) return dateStr;
    return d.toLocaleString('en-GB', {
      day: '2-digit', month: 'short', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
      hour12: false, timeZone: 'Asia/Kolkata',
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
      if (section === 'jobs')       loadJobs(0);
      else if (section === 'plans') loadPlans();
      else                          renderPlaceholder(section);
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
          class="btn-purchase ${featured ? 'btn-purchase--featured' : ''}"
          data-plan-id="${escHtml(plan.planId)}"
          data-plan-name="${escHtml(plan.planName)}"
        >Get ${escHtml(plan.planName)}</button>

      </div>`;
  }

  // Delegate purchase clicks from dynamically rendered cards
  document.getElementById('content-area').addEventListener('click', e => {
    const btn = e.target.closest('.btn-purchase');
    if (!btn) return;
    purchasePlan(btn.dataset.planId, btn.dataset.planName, btn);
  });

  async function purchasePlan(planId, planName, btn) {
    const original  = btn.textContent;
    btn.disabled    = true;
    btn.textContent = 'Processing…';

    try {
      const res  = await fetch('/billing/checkout', {
        method:  'POST',
        headers: { ...authHeader, 'Content-Type': 'application/json' },
        body:    JSON.stringify({ plan_id: planId }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.message || `HTTP ${res.status}`);

      if (data.url || data.checkoutUrl || data.redirect) {
        window.location.href = data.url || data.checkoutUrl || data.redirect;
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
  async function loadJobs(page = 0, filter = currentFilter) {
    currentPage   = page;
    currentFilter = filter;

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
            <span class="toolbar-label">Filter</span>
            <select class="filter-select" id="filterSelect">
              <option value="ALL">All Statuses</option>
              <option value="COMPLETED">Completed</option>
              <option value="FAILED">Failed</option>
              <option value="PENDING">Pending</option>
              <option value="RUNNING">Running</option>
            </select>
          </div>
          <span class="toolbar-label" id="record-count">—</span>
        </div>
        <div id="table-body"><div class="state-loading">Loading</div></div>
        <div class="pagination" id="pagination"></div>
      </div>`;

    document.getElementById('filterSelect').value = filter;
    document.getElementById('filterSelect').addEventListener('change', e => {
      loadJobs(0, e.target.value);
    });

    try {
      const url = filter === 'ALL'
        ? `/api/v1/jobs?limit=${LIMIT}&offset=${page}`
        : `/api/v1/jobs?limit=${LIMIT}&offset=${page}&status=${filter}`;

      const res  = await fetch(url, { headers: authHeader });
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
    const c = { COMPLETED: 0, FAILED: 0, PENDING: 0, RUNNING: 0 };
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
        <p class="stat-value failed">${c.FAILED}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Pending / Running</p>
        <p class="stat-value">${c.PENDING + c.RUNNING}</p>
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

  window._prevPage = () => loadJobs(currentPage - LIMIT, currentFilter);
  window._nextPage = () => loadJobs(currentPage + LIMIT, currentFilter);

  window.retryJob = async (jobId) => {
    try {
      await fetch(`/api/v1/jobs/${jobId}/retry`, { method: 'POST', headers: authHeader });
      loadJobs(currentPage, currentFilter);
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

    if (job.status === 'FAILED') {
      retryBtn.style.display = 'block';
      retryBtn.onclick       = () => { closeModal(); window.retryJob(job.jobId); };
      logsBtn.style.display  = 'flex';
      logsBtn.onclick = () => {
        const isOpen = logsPanel.style.display !== 'none';
        logsPanel.style.display = isOpen ? 'none' : 'block';
        logsBtn.classList.toggle('active', !isOpen);
        if (!isOpen) fetchLogs(job.jobId);
      };
      refreshBtn.onclick = () => fetchLogs(job.jobId);
    } else {
      retryBtn.style.display = 'none';
      logsBtn.style.display  = 'none';
    }

    document.getElementById('modal-overlay').classList.add('open');
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
