(() => {
  /* ── Auth ─────────────────────────────────────────── */
  const token = (() => {
    const m = document.cookie.match(/admin_token=([^;]+)/);
    return m ? m[1] : null;
  })();
  if (!token) { window.location.href = '/admin/login'; }
  const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };
  const authH = { Authorization: `Bearer ${token}` };

  /* ── State ────────────────────────────────────────── */
  const LIMIT = 10;
  let state = {
    section: 'overview',
    jobs:    { page: 0, filter: 'ALL', start: null, end: null },
    payments:{ page: 0, appId: '', start: null, end: null },
    apps:    { page: 0 },
    subs:    { page: 0 },
  };

  /* ── Helpers ──────────────────────────────────────── */
  const $ = id => document.getElementById(id);
  const esc = s => String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');

  function fmt(d) {
    if (!d) return '—';
    const dt = new Date(d);
    if (isNaN(dt)) return d;
    return dt.toLocaleString('en-GB', { day:'2-digit', month:'short', year:'numeric', hour:'2-digit', minute:'2-digit', hour12:false });
  }

  function fmtDate(d) {
    if (!d) return '—';
    const dt = new Date(d);
    return dt.toLocaleDateString('en-GB', { day:'2-digit', month:'short', year:'numeric' });
  }

  function fmtMoney(cents) {
    return (cents).toLocaleString('en-IN', { style:'currency', currency:'INR', maximumFractionDigits:2 });
  }

  function badge(status) {
    const cls = status ? `badge-${status.toUpperCase().replace(/\s/g,'_')}` : '';
    return `<span class="badge ${cls}">${esc(status || '—')}</span>`;
  }

  function toast(msg, type = 'success') {
    let el = $('admin-toast');
    if (!el) { el = document.createElement('div'); el.id = 'admin-toast'; el.className = 'toast'; document.body.appendChild(el); }
    el.textContent = msg;
    el.className = `toast ${type}`;
    el.classList.add('show');
    clearTimeout(el._t);
    el._t = setTimeout(() => el.classList.remove('show'), 3200);
  }

  function setArea(html) {
    const a = $('content-area');
    a.style.animation = 'none';
    requestAnimationFrame(() => { a.style.animation = ''; a.innerHTML = html; });
  }

  function paginationHtml(page, total, prevFn, nextFn) {
    const from = total === 0 ? 0 : page + 1;
    const to   = Math.min(page + LIMIT, total);
    return `<div class="pagination">
      <span class="pagination-info">${from}–${to} of ${total}</span>
      <div class="pagination-btns">
        <button class="btn-page" onclick="${prevFn}()" ${page <= 0 ? 'disabled' : ''}>← Prev</button>
        <button class="btn-page" onclick="${nextFn}()" ${page + LIMIT >= total ? 'disabled' : ''}>Next →</button>
      </div>
    </div>`;
  }

  /* ── Navigation ───────────────────────────────────── */
  document.querySelectorAll('.nav-btn[data-section]').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.nav-btn[data-section]').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      state.section = btn.dataset.section;
      $('topbar-title').textContent = btn.dataset.section.toUpperCase();
      loadSection(state.section);
    });
  });

  function loadSection(s) {
    const map = { overview, apps: loadApps, jobs: loadJobs, plans: loadPlans, payments: loadPayments, subscribers: loadSubscribers };
    if (map[s]) map[s]();
  }

  /* ══════════════════════════════════════════════════
     OVERVIEW
  ══════════════════════════════════════════════════ */
  async function overview() {
    setArea(`<div class="section-header"><div><p class="section-sub">System Health</p><h2>OVERVIEW</h2></div></div>
      <div class="kpi-grid" id="kpi-grid">
        <div class="kpi-card"><p class="kpi-label">Total Users</p><p class="kpi-value" id="kv-users">—</p><p class="kpi-meta">Registered accounts</p></div>
        <div class="kpi-card"><p class="kpi-label">Revenue This Month</p><p class="kpi-value" id="kv-revenue">—</p><p class="kpi-meta" id="kv-delta"><span class="kpi-delta" id="kv-delta-val"></span>&nbsp;vs last month</p></div>
        <div class="kpi-card"><p class="kpi-label">Active Subscribers</p><p class="kpi-value" id="kv-subs">—</p><p class="kpi-meta">Paying plans</p></div>
      </div>
      <div class="kpi-grid" id="kpi-grid2">
        <div class="kpi-card"><p class="kpi-label">Total Apps</p><p class="kpi-value" id="kv-apps">—</p></div>
        <div class="kpi-card"><p class="kpi-label">Jobs Today</p><p class="kpi-value" id="kv-jobs">—</p></div>
        <div class="kpi-card"><p class="kpi-label">Failed Jobs</p><p class="kpi-value" id="kv-failed" style="color:var(--red)">—</p></div>
      </div>`);

    try {
      const [statsRes, subsRes, appsRes] = await Promise.all([
        fetch('/api/v1/admin/stats', { headers: authH }).catch(() => null),
        fetch('/api/v1/admin/subscribers?limit=1', { headers: authH }).catch(() => null),
        fetch('/api/v1/admin/apps?limit=1', { headers: authH }).catch(() => null),
      ]);

      if (statsRes?.ok) {
        const d = await statsRes.json();
        if ($('kv-users'))   $('kv-users').textContent   = (d.totalUsers   ?? '—').toLocaleString?.() ?? d.totalUsers ?? '—';
        if ($('kv-revenue')) $('kv-revenue').textContent = d.revenueMonth  != null ? fmtMoney(d.revenueMonth) : '—';
        if ($('kv-jobs'))    $('kv-jobs').textContent    = (d.jobsToday    ?? '—').toLocaleString?.() ?? '—';
        if ($('kv-failed'))  $('kv-failed').textContent  = (d.jobsFailed   ?? '—').toLocaleString?.() ?? '—';

        if (d.revenueMonth != null && d.revenueLastMonth != null) {
          const diff  = d.revenueMonth - d.revenueLastMonth;
          const sign  = diff >= 0 ? '+' : '';
          const cls   = diff >= 0 ? 'up' : 'down';
          if ($('kv-delta-val')) { $('kv-delta-val').textContent = `${sign}${fmtMoney(Math.abs(diff))}`; $('kv-delta-val').className = `kpi-delta ${cls}`; }
        }
      }

      if (subsRes?.ok) {
        const d = await subsRes.json();
        if ($('kv-subs')) $('kv-subs').textContent = (d.total ?? d.subscribers?.length ?? '—').toLocaleString?.() ?? '—';
      }

      if (appsRes?.ok) {
        const d = await appsRes.json();
        if ($('kv-apps')) $('kv-apps').textContent = (d.total ?? d.apps?.length ?? '—').toLocaleString?.() ?? '—';
      }
    } catch { /* KPIs remain — */ }
  }

  /* ══════════════════════════════════════════════════
     APPS
  ══════════════════════════════════════════════════ */
  async function loadApps(page = state.apps.page) {
    state.apps.page = page;
    setArea(`<div class="section-header"><div><p class="section-sub">Tenants</p><h2>APPS</h2></div></div>
      <div class="table-wrap">
        <div class="table-toolbar"><div class="toolbar-left"><span class="toolbar-label">All Applications</span></div><span class="toolbar-label" id="apps-count">—</span></div>
        <div id="apps-body"><div class="state-loading">Loading</div></div>
      </div>`);

    try {
      const res  = await fetch(`/api/v1/admin/apps?limit=${LIMIT}&offset=${page}`, { headers: authH });
      const data = await res.json();
      const apps = data.apps || data || [];
      const total = data.total || apps.length;
      $('apps-count').textContent = `${total} app${total !== 1 ? 's' : ''}`;

      if (!apps.length) { $('apps-body').innerHTML = '<div class="state-empty">No apps found</div>'; return; }

      const rows = apps.map(a => `<tr>
        <td><span class="cell-id" title="${esc(a.appId || a.id)}">${esc(a.appId || a.id)}</span></td>
        <td>${esc(a.name || a.appName || '—')}</td>
        <td>${esc(a.email || '—')}</td>
        <td>${badge(a.status)}</td>
        <td>${esc(a.planName || '—')}</td>
        <td>${fmtDate(a.planEndDate || a.currentPeriodEnd)}</td>
        <td class="action-cell">
          ${a.status === 'active' || a.status === 'Active'
            ? `<button class="btn-action-sm btn-block" onclick="toggleApp('${esc(a.appId||a.id)}','block',this)">Block</button>`
            : `<button class="btn-action-sm btn-unblock" onclick="toggleApp('${esc(a.appId||a.id)}','unblock',this)">Unblock</button>`}
          <button class="btn-action-sm btn-neutral" onclick="openAppUsers('${esc(a.appId||a.id)}','${esc(a.name||a.appName||a.email||'')}')">Users</button>
        </td>
      </tr>`).join('');

      $('apps-body').innerHTML = `<table>
        <thead><tr><th>App ID</th><th>Name</th><th>Email</th><th>Status</th><th>Plan</th><th>Plan Ends</th><th>Actions</th></tr></thead>
        <tbody>${rows}</tbody>
      </table>
      ${paginationHtml(page, total, '_appsPrev', '_appsNext')}`;
    } catch {
      $('apps-body').innerHTML = '<div class="state-empty">Failed to load apps</div>';
    }
  }

  window._appsPrev = () => loadApps(state.apps.page - LIMIT);
  window._appsNext = () => loadApps(state.apps.page + LIMIT);

  window.toggleApp = async (appId, action, btn) => {
    const orig = btn.textContent;
    btn.disabled = true; btn.textContent = '…';
    try {
      const res = await fetch(`/api/v1/admin/apps/${appId}/${action}`, { method: 'POST', headers: authH });
      if (!res.ok) throw new Error((await res.json().catch(()=>({}))).message || 'Failed');
      toast(`App ${action}ed successfully`);
      loadApps(state.apps.page);
    } catch (err) {
      toast(err.message, 'error');
      btn.disabled = false; btn.textContent = orig;
    }
  };

  /* App Users modal */
  window.openAppUsers = async (appId, label) => {
    $('app-users-title').textContent = esc(label) || 'USERS';
    $('app-users-body').innerHTML = '<div class="state-loading">Loading</div>';
    $('app-users-overlay').classList.add('open');
    try {
      const res  = await fetch(`/api/v1/admin/apps/${appId}/users`, { headers: authH });
      const data = await res.json();
      const users = data.users || data || [];
      if (!users.length) { $('app-users-body').innerHTML = '<div class="state-empty">No users</div>'; return; }
      const rows = users.map(u => `<tr>
        <td>${esc(u.email)}</td>
        <td><span class="role-chip">${esc(u.role)}</span></td>
        <td>${badge(u.status)}</td>
      </tr>`).join('');
      $('app-users-body').innerHTML = `<table>
        <thead><tr><th>Email</th><th>Role</th><th>Status</th></tr></thead>
        <tbody>${rows}</tbody></table>`;
    } catch {
      $('app-users-body').innerHTML = '<div class="state-empty">Failed to load users</div>';
    }
  };

  window.closeAppUsers = () => $('app-users-overlay').classList.remove('open');
  $('app-users-overlay').addEventListener('click', e => { if (e.target === $('app-users-overlay')) closeAppUsers(); });

  /* ══════════════════════════════════════════════════
     JOBS
  ══════════════════════════════════════════════════ */
  async function loadJobs(page = 0, filter = state.jobs.filter, start = state.jobs.start, end = state.jobs.end) {
    state.jobs = { page, filter, start, end };
    const today = new Date().toLocaleDateString('en-CA');

    setArea(`<div class="section-header"><div><p class="section-sub">Queue</p><h2>JOBS</h2></div></div>
      <div class="table-wrap">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <span class="toolbar-label">Status</span>
            <select class="filter-select" id="j-status">
              ${['ALL','COMPLETED','FAILED','PENDING','SCHEDULED','QUEUED','RUNNING','CANCELLED'].map(s =>
                `<option value="${s}" ${s===filter?'selected':''}>${s==='ALL'?'All Statuses':s}</option>`).join('')}
            </select>
            <span class="toolbar-divider"></span>
            <span class="toolbar-label">From</span>
            <input type="date" class="filter-date" id="j-start" max="${today}" value="${start?start.substring(0,10):''}" />
            <span class="toolbar-label">To</span>
            <input type="date" class="filter-date" id="j-end"   max="${today}" value="${end?end.substring(0,10):''}" />
            <button class="btn-filter" id="j-apply">Apply</button>
            <button class="btn-filter-outline" id="j-clear">Clear</button>
          </div>
          <span class="toolbar-label" id="j-count">—</span>
        </div>
        <div id="jobs-body"><div class="state-loading">Loading</div></div>
      </div>`);

    const startEl = $('j-start'), endEl = $('j-end');
    startEl.addEventListener('change', () => { endEl.min = startEl.value || ''; if(endEl.value && endEl.value < startEl.value) endEl.value = startEl.value; });
    endEl.addEventListener('change',   () => { if(startEl.value && endEl.value < startEl.value) endEl.value = startEl.value; });

    $('j-apply').addEventListener('click', () => {
      const s = startEl.value, e = endEl.value;
      loadJobs(0, $('j-status').value, s ? `${s}T00:00:00Z` : null, e ? `${e}T23:59:59Z` : null);
    });
    $('j-clear').addEventListener('click', () => { startEl.value=''; endEl.value=''; loadJobs(0, $('j-status').value, null, null); });
    $('j-status').addEventListener('change', e => loadJobs(0, e.target.value, state.jobs.start, state.jobs.end));

    try {
      const p = new URLSearchParams({ limit: LIMIT, offset: page });
      if (filter !== 'ALL') p.set('status', filter);
      if (start) p.set('startDate', start);
      if (end)   p.set('endDate', end);

      const res  = await fetch(`/api/v1/admin/jobs?${p}`, { headers: authH });
      const data = await res.json();
      const jobs = data.jobs || [];
      const total = data.total || jobs.length;
      $('j-count').textContent = `${total} record${total!==1?'s':''}`;

      if (!jobs.length) { $('jobs-body').innerHTML = '<div class="state-empty">No jobs found</div>'; return; }

      const rows = jobs.map(j => `<tr>
        <td><span class="cell-id" title="${esc(j.jobId)}">${esc(j.jobId)}</span></td>
        <td>${esc(j.type||'—')}</td>
        <td>${badge(j.status)}</td>
        <td>${j.retry??0} / ${j.maxRetry??0}</td>
        <td>${fmt(j.scheduleAt)}</td>
        <td>${fmt(j.createdAt)}</td>
      </tr>`).join('');

      $('jobs-body').innerHTML = `<table>
        <thead><tr><th>Job ID</th><th>Type</th><th>Status</th><th>Retry</th><th>Scheduled</th><th>Created</th></tr></thead>
        <tbody>${rows}</tbody>
      </table>
      ${paginationHtml(page, total, '_jobsPrev', '_jobsNext')}`;
    } catch {
      $('jobs-body').innerHTML = '<div class="state-empty">Failed to load jobs</div>';
    }
  }

  window._jobsPrev = () => loadJobs(state.jobs.page - LIMIT, state.jobs.filter, state.jobs.start, state.jobs.end);
  window._jobsNext = () => loadJobs(state.jobs.page + LIMIT, state.jobs.filter, state.jobs.start, state.jobs.end);

  /* ══════════════════════════════════════════════════
     PLANS
  ══════════════════════════════════════════════════ */
  let allPlans = [];

  async function loadPlans() {
    setArea(`<div class="section-header">
        <div><p class="section-sub">Pricing</p><h2>PLANS</h2></div>
        <button class="btn-create" onclick="openPlanModal()">+ New Plan</button>
      </div>
      <div id="plans-area"><div class="state-loading">Loading</div></div>`);

    try {
      const res  = await fetch('/api/v1/plans', { headers: authH });
      const data = await res.json();
      allPlans   = data.plans || [];
      renderPlans();
    } catch {
      $('plans-area').innerHTML = '<div class="state-empty">Failed to load plans</div>';
    }
  }

  function renderPlans() {
    if (!allPlans.length) { $('plans-area').innerHTML = '<div class="state-empty">No plans found</div>'; return; }
    $('plans-area').innerHTML = `<div class="admin-plans-grid">${
      allPlans.map(p => `<div class="admin-plan-card">
        <p class="admin-plan-name">${esc(p.planName)}</p>
        <p class="admin-plan-price">${fmtMoney(p.planprice)} / mo</p>
        <p class="admin-plan-limit">${Number(p.monthlyLimit).toLocaleString()} jobs / month</p>
        <p class="admin-plan-id">${esc(p.planId)}</p>
        <button class="btn-action-sm btn-amber" onclick="openPlanModal('${esc(p.planId)}')">Edit Plan</button>
      </div>`).join('')
    }</div>`;
  }

  window.openPlanModal = (planId) => {
    const plan = planId ? allPlans.find(p => p.planId === planId) : null;
    $('plan-modal-tag').textContent   = plan ? 'Edit Plan' : 'New Plan';
    $('plan-modal-title').textContent = plan ? plan.planName : 'CREATE';
    $('pm-name').value   = plan?.planName    || '';
    $('pm-price').value  = plan?.planprice   || '';
    $('pm-limit').value  = plan?.monthlyLimit|| '';
    $('plan-modal-error').textContent = '';
    $('plan-modal-overlay').dataset.planId = planId || '';
    $('plan-modal-overlay').classList.add('open');
  };

  window.closePlanModal = () => $('plan-modal-overlay').classList.remove('open');
  $('plan-modal-overlay').addEventListener('click', e => { if(e.target===$('plan-modal-overlay')) closePlanModal(); });

  $('plan-modal-save').addEventListener('click', async () => {
    const planId = $('plan-modal-overlay').dataset.planId;
    const body   = { planName: $('pm-name').value.trim(), planprice: Number($('pm-price').value), monthlyLimit: Number($('pm-limit').value) };
    const errEl  = $('plan-modal-error');
    const btn    = $('plan-modal-save');
    errEl.textContent = '';

    if (!body.planName) { errEl.textContent = 'Plan name is required.'; return; }
    if (!body.planprice || body.planprice < 0) { errEl.textContent = 'Valid price required.'; return; }
    if (!body.monthlyLimit || body.monthlyLimit < 1) { errEl.textContent = 'Job limit must be ≥ 1.'; return; }

    btn.disabled = true; btn.textContent = 'Saving…';
    try {
      const url    = planId ? `/api/v1/admin/plans/${planId}` : '/api/v1/admin/plans';
      const method = planId ? 'PUT' : 'POST';
      const res    = await fetch(url, { method, headers: H, body: JSON.stringify(body) });
      const data   = await res.json().catch(()=>({}));
      if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
      closePlanModal();
      toast(planId ? 'Plan updated' : 'Plan created');
      loadPlans();
    } catch(err) {
      errEl.textContent = err.message;
    } finally {
      btn.disabled = false; btn.textContent = 'Save Plan';
    }
  });

  /* ══════════════════════════════════════════════════
     PAYMENTS
  ══════════════════════════════════════════════════ */
  async function loadPayments(page = 0, appId = state.payments.appId, start = state.payments.start, end = state.payments.end) {
    state.payments = { page, appId, start, end };
    const today = new Date().toLocaleDateString('en-CA');

    setArea(`<div class="section-header"><div><p class="section-sub">Billing</p><h2>PAYMENTS</h2></div></div>
      <div class="table-wrap">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <span class="toolbar-label">App</span>
            <input type="text" class="filter-select" id="pay-app" placeholder="App ID or name" style="width:160px" value="${esc(appId)}" />
            <span class="toolbar-divider"></span>
            <span class="toolbar-label">From</span>
            <input type="date" class="filter-date" id="pay-start" max="${today}" value="${start?start.substring(0,10):''}" />
            <span class="toolbar-label">To</span>
            <input type="date" class="filter-date" id="pay-end"   max="${today}" value="${end?end.substring(0,10):''}" />
            <button class="btn-filter" id="pay-apply">Apply</button>
            <button class="btn-filter-outline" id="pay-clear">Clear</button>
          </div>
          <span class="toolbar-label" id="pay-count">—</span>
        </div>
        <div id="pay-body"><div class="state-loading">Loading</div></div>
      </div>`);

    const startEl = $('pay-start'), endEl = $('pay-end');
    startEl.addEventListener('change', () => { endEl.min = startEl.value||''; if(endEl.value && endEl.value < startEl.value) endEl.value = startEl.value; });
    endEl.addEventListener('change',   () => { if(startEl.value && endEl.value < startEl.value) endEl.value = startEl.value; });

    $('pay-apply').addEventListener('click', () => {
      const s = startEl.value, e = endEl.value;
      loadPayments(0, $('pay-app').value.trim(), s?`${s}T00:00:00Z`:null, e?`${e}T23:59:59Z`:null);
    });
    $('pay-clear').addEventListener('click', () => { $('pay-app').value=''; startEl.value=''; endEl.value=''; loadPayments(0,'',null,null); });

    try {
      const p = new URLSearchParams({ limit: LIMIT, offset: page });
      if (appId) p.set('appId', appId);
      if (start) p.set('startDate', start);
      if (end)   p.set('endDate', end);

      const res  = await fetch(`/api/v1/admin/payments?${p}`, { headers: authH });
      const data = await res.json();
      const payments = data.payments || data || [];
      const total    = data.total || payments.length;
      $('pay-count').textContent = `${total} record${total!==1?'s':''}`;

      if (!payments.length) { $('pay-body').innerHTML = '<div class="state-empty">No payments found</div>'; return; }

      const rows = payments.map(p => `<tr>
        <td><span class="cell-id" title="${esc(p.paymentId||p.id)}">${esc(p.paymentId||p.id||'—')}</span></td>
        <td>${esc(p.appName||p.appId||'—')}</td>
        <td>${esc(p.planName||'—')}</td>
        <td style="color:var(--green);font-family:'Bebas Neue',sans-serif;font-size:14px;letter-spacing:.04em">${p.amount!=null?fmtMoney(p.amount):'—'}</td>
        <td>${badge(p.status||'paid')}</td>
        <td>${fmt(p.createdAt||p.paidAt)}</td>
        <td><button class="btn-invoice" onclick="downloadInvoice('${esc(p.paymentId||p.id)}')">↓ Invoice</button></td>
      </tr>`).join('');

      $('pay-body').innerHTML = `<table>
        <thead><tr><th>Payment ID</th><th>App</th><th>Plan</th><th>Amount</th><th>Status</th><th>Date</th><th>Invoice</th></tr></thead>
        <tbody>${rows}</tbody>
      </table>
      ${paginationHtml(page, total, '_payPrev', '_payNext')}`;
    } catch {
      $('pay-body').innerHTML = '<div class="state-empty">Failed to load payments</div>';
    }
  }

  window._payPrev = () => loadPayments(state.payments.page - LIMIT, state.payments.appId, state.payments.start, state.payments.end);
  window._payNext = () => loadPayments(state.payments.page + LIMIT, state.payments.appId, state.payments.start, state.payments.end);

  window.downloadInvoice = async (paymentId) => {
    try {
      const res = await fetch(`/api/v1/admin/payments/${paymentId}/invoice`, { headers: authH });
      if (!res.ok) throw new Error('Invoice not available');
      const blob = await res.blob();
      const url  = URL.createObjectURL(blob);
      const a    = document.createElement('a');
      a.href = url; a.download = `invoice_${paymentId}.pdf`;
      document.body.appendChild(a); a.click();
      document.body.removeChild(a); URL.revokeObjectURL(url);
    } catch(err) {
      toast(err.message || 'Failed to download invoice', 'error');
    }
  };

  /* ══════════════════════════════════════════════════
     SUBSCRIBERS
  ══════════════════════════════════════════════════ */
  async function loadSubscribers(page = state.subs.page) {
    state.subs.page = page;
    setArea(`<div class="section-header"><div><p class="section-sub">Active Plans</p><h2>SUBSCRIBERS</h2></div></div>
      <div class="table-wrap">
        <div class="table-toolbar"><div class="toolbar-left"><span class="toolbar-label">Active Subscriptions</span></div><span class="toolbar-label" id="subs-count">—</span></div>
        <div id="subs-body"><div class="state-loading">Loading</div></div>
      </div>`);

    try {
      const res  = await fetch(`/api/v1/admin/subscribers?limit=${LIMIT}&offset=${page}`, { headers: authH });
      const data = await res.json();
      const subs = data.subscribers || data || [];
      const total = data.total || subs.length;
      $('subs-count').textContent = `${total} subscriber${total!==1?'s':''}`;

      if (!subs.length) { $('subs-body').innerHTML = '<div class="state-empty">No active subscribers</div>'; return; }

      const rows = subs.map(s => {
        const start  = new Date(s.currentPeriodStart || s.startDate);
        const end    = new Date(s.currentPeriodEnd   || s.endDate);
        const now    = new Date();
        const total  = end - start;
        const elapsed= Math.min(now - start, total);
        const pct    = total > 0 ? Math.round((elapsed / total) * 100) : 0;
        const daysLeft = Math.max(0, Math.round((end - now) / 86400000));

        return `<tr>
          <td><span class="cell-id" title="${esc(s.appId||s.id)}">${esc(s.appId||s.id)}</span></td>
          <td>${esc(s.appName||s.email||'—')}</td>
          <td>${esc(s.planName||'—')}</td>
          <td>${badge('active')}</td>
          <td>${fmtDate(s.currentPeriodStart||s.startDate)}</td>
          <td>${fmtDate(s.currentPeriodEnd||s.endDate)}</td>
          <td>
            <div class="duration-bar-wrap" title="${pct}% elapsed, ${daysLeft}d left">
              <div class="duration-bar-track"><div class="duration-bar-fill" style="width:${pct}%"></div></div>
            </div>
          </td>
          <td style="font-size:11px;color:var(--gray-mid)">${daysLeft}d left</td>
        </tr>`;
      }).join('');

      $('subs-body').innerHTML = `<table>
        <thead><tr><th>App ID</th><th>App</th><th>Plan</th><th>Status</th><th>Start</th><th>End</th><th>Duration</th><th></th></tr></thead>
        <tbody>${rows}</tbody>
      </table>
      ${paginationHtml(page, total, '_subsPrev', '_subsNext')}`;
    } catch {
      $('subs-body').innerHTML = '<div class="state-empty">Failed to load subscribers</div>';
    }
  }

  window._subsPrev = () => loadSubscribers(state.subs.page - LIMIT);
  window._subsNext = () => loadSubscribers(state.subs.page + LIMIT);

  /* ── Logout ───────────────────────────────────────── */
  $('btn-logout').addEventListener('click', () => {
    document.cookie = 'admin_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
    window.location.href = '/admin/login';
  });

  /* ── Boot ─────────────────────────────────────────── */
  overview();
})();
