// --------------------------- MAIN APP ENTRY --------------------------- //
document.addEventListener("DOMContentLoaded", async () => {
  console.log("🚀 app.js loaded");

  const path = window.location.pathname;
  const onDashboard = path === "/dashboard";
  const onResources = path === "/resources";
  const onUsers = path === "/users";
  const onAudit = path === "/audit";
  const onRoles = path === "/roles";
  // Protect dashboard and resources routes
  if (onDashboard || onResources || onAudit || onUsers || onRoles) checkAuth();

  // load current user permissions for client-side checks
  window.currentUserPermissions = [];
  window.hasResourceWrite = false;
  async function loadCurrentUser() {
    try {
      const res = await fetch('/api/v1/me', { credentials: 'include' });
      if (!res.ok) return;
      const data = await res.json();
      window.currentUserPermissions = data.permissions || [];
      window.hasResourceWrite = window.currentUserPermissions.includes('resources:write');
    } catch (err) {
      console.error('Failed to load current user permissions', err);
    }
  }
  if (onDashboard || onResources || onUsers) loadCurrentUser();

  // Login handler
  const loginForm = document.getElementById("loginForm");
  if (loginForm) loginForm.addEventListener("submit", handleLogin);

  // Logout & dropdown menu
  setupLogout();
  setupUserMenu();

  // Dashboard loaders
  if (onDashboard) {
    loadTable("/api/v1/roles", "rolesTable", 4);
    loadTable("/api/v1/resources", "resourcesTable", 4);
    loadResourcesChart();
  }

  // Users loaders
  if (onUsers) {
    console.log("👤 Loading real Users...");
    // ensure we have current user permissions before rendering actions
    await loadCurrentUser();
    loadUsers();
    setupAddUserModal();
  }

  // Resources page handlers
  if (onResources) {
    console.log("🔌 Loading real local resources...");
    loadLocalResources();
    setupAddResourceModal();
    setupNoAccessModal();
  }

  // Audit page handlers
  if (onAudit) {
    console.log("📜 Loading Audit");
    loadAuditLogs();
  }

  // Audit page handlers
  if (onRoles) {
    console.log("📜 Loading Roles");
    loadAssignUsers();
  }

});

// --------------------------- LOGIN HANDLER --------------------------- //
async function handleLogin(e) {
  e.preventDefault();

  const formData = new FormData(e.target);
  const payload = Object.fromEntries(formData.entries());

  try {
    const res = await fetch("/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      credentials: "include",
    });

    const data = await res.json();
    if (res.ok && data.token) {
      console.log("✅ Login successful");
      window.location.href = "/dashboard";
    } else {
      alert("❌ Login failed: " + (data.error || "Invalid credentials"));
    }
  } catch (err) {
    console.error("⚠️ Login error:", err);
    alert("Network error");
  }
}

// --------------------------- AUTH GUARD --------------------------- //
async function checkAuth() {
  try {
    const res = await fetch("/api/v1/roles", { credentials: "include" });
    if (res.status === 401) {
      console.warn("⚠️ Unauthorized — redirecting to login");
      window.location.href = "/login";
    }
  } catch (err) {
    console.error("Auth check failed:", err);
    window.location.href = "/login";
  }
}

// --------------------------- DASHBOARD TABLES --------------------------- //
async function loadTable(apiPath, tableId, columns) {
  try {
    const res = await fetch(apiPath, { credentials: "include" });
    const data = await res.json();
    const table = document.getElementById(tableId);
    if (!table) return;

    const records = data.users || data.roles || data.resources || data.logs || [];

    if (!records || records.length === 0) {
      table.innerHTML = `<tr><td colspan="${columns}" class="py-4 text-center text-slate-400">No data found</td></tr>`;
      return;
    }

    // Special-case rendering for roles table to ensure stable column order
    if (tableId === "rolesTable") {
      table.innerHTML = records
        .map((r) => {
          // API may return keys in different casing; prefer canonical lower-case keys
          const id = r.id ?? r.ID ?? r.created_at ?? "-";
          const name = r.name ?? r.Name ?? "-";
          const slug = r.slug ?? r.Slug ?? "-";
          const users = (r.users_count ?? r.usersCount ?? r.users ?? 0);
          return `
            <tr class="border-b last:border-0">
              <td class="py-2 px-2 whitespace-nowrap text-slate-700">${id ?? "-"}</td>
              <td class="py-2 px-2 whitespace-nowrap text-slate-700">${name ?? "-"}</td>
              <td class="py-2 px-2 whitespace-nowrap text-slate-700">${slug ?? "-"}</td>
              <td class="py-2 px-2 whitespace-nowrap text-slate-700">${users ?? 0}</td>
            </tr>`;
        })
        .join("");
      return;
    }

    // Generic renderer for other tables (keeps previous behavior)
    table.innerHTML = records
      .map(
        (r) => `
        <tr class="border-b last:border-0">
          ${Object.values(r)
            .slice(0, columns)
            .map((v) =>
              `<td class="py-2 px-2 whitespace-nowrap text-slate-700">${v ?? "-"}</td>`
            )
            .join("")}
        </tr>`
      )
      .join("");
  } catch (err) {
    console.error(`Error loading ${apiPath}:`, err);
  }
}

// ===== USERS PAGE =====
async function loadUsers() {
  const table = document.getElementById("usersTable");
  if (!table) return;

  try {
    const res = await fetch("/api/v1/users", { credentials: "include" });
    const data = await res.json();

    table.innerHTML = "";

    if (!data.users || data.users.length === 0) {
      table.innerHTML = `<tr><td colspan="6" class="py-4 text-center text-slate-400">No users found.</td></tr>`;
      return;
    }

    data.users.forEach((u, i) => {
      // Role may come as u.Roles (array) or u.role; handle both
      let roleText = "-";
      if (u.Roles && Array.isArray(u.Roles) && u.Roles.length > 0) {
        roleText = u.Roles.map(r => r.Name || r.name || r.slug || "-").join(", ");
      } else if (u.role) roleText = u.role;

      const status = u.Status || u.status || "active";

      const uid = u.ID ?? u.id ?? u.Id ?? u.ID;
      const showAction = false; // initial, visibility toggled after permissions load
      // Show Activate if suspended, otherwise Deactivate
      const isSuspended = (String(status).toLowerCase() === 'suspended');
      const deactivateBtn = `<button data-user-id="${uid}" class="deactivate-btn ${showAction ? '' : 'hidden'} px-3 py-1.5 rounded bg-red-600 text-white text-xs hover:bg-red-700">Deactivate</button>`;
      const activateBtn = `<button data-user-id="${uid}" class="activate-btn ${showAction ? '' : 'hidden'} px-3 py-1.5 rounded bg-green-600 text-white text-xs hover:bg-green-700">Activate</button>`;
      const changePwdBtn = `<button data-user-id="${uid}" class="change-pwd-btn ${showAction ? '' : 'hidden'} px-3 py-1.5 rounded bg-slate-600 text-white text-xs hover:bg-slate-700">Change Password</button>`;
      // wrap action buttons in a flex container to keep consistent alignment
      const actionBtn = `
        <div class="flex items-center gap-2">
          ${isSuspended ? (activateBtn + changePwdBtn) : (deactivateBtn + changePwdBtn)}
          <button data-user-id="${uid}" class="assign-access-btn ${showAction ? '' : 'hidden'} px-3 py-1.5 rounded bg-indigo-600 text-white text-xs hover:bg-indigo-700">Assign Access</button>
        </div>`;

      const row = `
        <tr class="border-b hover:bg-slate-50 transition">
          <td class="py-3 px-4">${i + 1}</td>
          <td class="py-3 px-4 font-medium text-slate-700">${u.Name || u.name || "-"}</td>
          <td class="py-3 px-4 text-slate-600">${u.Email || u.email || "-"}</td>
          <td class="py-3 px-4">${roleText}</td>
          <td class="py-3 px-4">${status}</td>
          <td class="py-3 px-4">${actionBtn}</td>
        </tr>`;
      table.insertAdjacentHTML("beforeend", row);
    });

    // After rendering, wire up action buttons and adjust visibility based on permissions
    updateUserActionVisibility();
    attachDeactivateHandlers();
    attachActivateHandlers();
    attachChangePasswordHandlers();
    attachAssignAccessHandlers();
  } catch (err) {
    console.error("Failed to load users:", err);
    table.innerHTML = `<tr><td colspan="6" class="py-4 text-center text-red-500">Failed to load users.</td></tr>`;
  }
}

function attachAssignAccessHandlers() {
  document.querySelectorAll('.assign-access-btn').forEach(btn => {
    btn.onclick = async (e) => {
      const userId = e.currentTarget.dataset.userId;
      if (!userId) return;
      await openAssignAccessModal(userId);
    };
  });
}

function updateUserActionVisibility() {
  const allowed = Array.isArray(window.currentUserPermissions) && window.currentUserPermissions.includes('users:assign-role');
  document.querySelectorAll('.deactivate-btn').forEach(btn => {
    if (allowed) btn.classList.remove('hidden');
    else btn.classList.add('hidden');
  });
  document.querySelectorAll('.activate-btn').forEach(btn => {
    if (allowed) btn.classList.remove('hidden');
    else btn.classList.add('hidden');
  });
  document.querySelectorAll('.change-pwd-btn').forEach(btn => {
    if (allowed) btn.classList.remove('hidden');
    else btn.classList.add('hidden');
  });
  document.querySelectorAll('.assign-access-btn').forEach(btn => {
    if (allowed) btn.classList.remove('hidden');
    else btn.classList.add('hidden');
  });
}

function attachDeactivateHandlers() {
  document.querySelectorAll('.deactivate-btn').forEach(btn => {
    btn.onclick = async (e) => {
      const userId = e.currentTarget.dataset.userId;
      if (!userId) return;
      if (!confirm('Are you sure you want to deactivate this user account?')) return;
      try {
        const res = await fetch(`/api/v1/users/${encodeURIComponent(userId)}/deactivate`, {
          method: 'POST',
          credentials: 'include',
        });
        const data = await res.json();
        if (res.ok) {
          alert('User deactivated');
          loadUsers();
        } else {
          alert('Failed to deactivate user: ' + (data.error || data.message || res.statusText));
        }
      } catch (err) {
        console.error('Deactivate failed', err);
        alert('Network error');
      }
    };
  });
}

function attachActivateHandlers() {
  document.querySelectorAll('.activate-btn').forEach(btn => {
    btn.onclick = async (e) => {
      const userId = e.currentTarget.dataset.userId;
      if (!userId) return;
      if (!confirm('Are you sure you want to activate this user account?')) return;
      try {
        const res = await fetch(`/api/v1/users/${encodeURIComponent(userId)}/activate`, {
          method: 'POST',
          credentials: 'include',
        });
        const data = await res.json();
        if (res.ok) {
          alert('User activated');
          loadUsers();
        } else {
          alert('Failed to activate user: ' + (data.error || data.message || res.statusText));
        }
      } catch (err) {
        console.error('Activate failed', err);
        alert('Network error');
      }
    };
  });
}

function attachChangePasswordHandlers() {
  document.querySelectorAll('.change-pwd-btn').forEach(btn => {
    btn.onclick = (e) => {
      const userId = e.currentTarget.dataset.userId;
      if (!userId) return;
      const modal = document.getElementById('changePasswordModal');
      const cpUser = document.getElementById('cpUserId');
      const cpPassword = document.getElementById('cpPassword');
      if (!modal || !cpUser || !cpPassword) return;
      cpUser.value = userId;
      cpPassword.value = '';
      modal.classList.remove('hidden');
    };
  });

  const cpCancel = document.getElementById('cpCancel');
  if (cpCancel) cpCancel.onclick = () => document.getElementById('changePasswordModal').classList.add('hidden');

  const cpForm = document.getElementById('changePasswordForm');
  if (cpForm) cpForm.onsubmit = async (e) => {
    e.preventDefault();
    const userId = document.getElementById('cpUserId').value;
    const password = document.getElementById('cpPassword').value;
    if (!userId || !password || password.length < 8) {
      alert('Password must be at least 8 characters');
      return;
    }
    try {
      const res = await fetch(`/api/v1/users/${encodeURIComponent(userId)}/password`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        credentials: 'include',
        body: JSON.stringify({password}),
      });
      const data = await res.json();
      if (res.ok) {
        alert('Password updated');
        document.getElementById('changePasswordModal').classList.add('hidden');
      } else {
        alert('Failed to update password: ' + (data.error || data.message || res.statusText));
      }
    } catch (err) {
      console.error('Change password failed', err);
      alert('Network error');
    }
  };
}

// Hook refresh button
document.addEventListener("DOMContentLoaded", () => {
  if (document.getElementById("refreshUsers")) {
    loadUsers();
    document.getElementById("refreshUsers").addEventListener("click", loadUsers);
  }
  if (document.getElementById("refreshRoles")) {
    loadTable("/api/v1/roles", "rolesTable", 4);
    document.getElementById("refreshRoles").addEventListener("click", () => loadTable("/api/v1/roles", "rolesTable", 4));
  }
});

// --------------------------- ADD User MODAL --------------------------- //
function setupAddUserModal() {
  const modal = document.getElementById("addUserModal");
  const showBtn = document.getElementById("showAddUser");
  const cancelBtn = document.getElementById("cancelAddUser");
  const form = document.getElementById("addUserForm");

  if (showBtn && modal) showBtn.addEventListener("click", () => {
    modal.classList.remove("hidden");
    try {
      const archSel = document.getElementById('agentArchSelect');
      const arch = archSel ? archSel.value : 'linux-amd64';
      const binaryName = `teleport-agent-${arch}`;
      const tarName = `${binaryName}.tar.gz`;
      const extractEl = document.getElementById('cmdExtract');
      const runEl = document.getElementById('cmdRun');
      if (extractEl) extractEl.innerText = `tar -xzf ${tarName}`;
      // include token if already created
      const tokenEl = document.getElementById('addGenTokenValue');
      const token = tokenEl ? tokenEl.textContent.trim() : '';
      if (runEl) runEl.innerText = token ? `CONTROLLER_URL=${window.location.origin} AGENT_REG_TOKEN=${token} ./${binaryName}` : `CONTROLLER_URL=${window.location.origin} ./${binaryName}`;
    } catch (e) {
      console.warn('Failed to set extract/run commands on modal open', e);
    }
  });
  if (cancelBtn && modal) cancelBtn.addEventListener("click", () => modal.classList.add("hidden"));

  if (form) {
    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      const formData = new FormData(form);
      const payload = Object.fromEntries(formData.entries());

      try {
        const res = await fetch("/api/v1/users", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify(payload),
        });

        const data = await res.json();
        if (res.ok) {
          console.log("✅ User added:", data);
          form.reset();
          modal.classList.add("hidden");
          loadLocalUsers();
        } else {
          alert("❌ Failed: " + data.error);
        }
      } catch (err) {
        console.error("Error adding User:", err);
      }
    });
  }
}


// --------------------------- RESOURCES: LOAD LOCAL INSTANCE --------------------------- //
async function loadLocalResources() {
  const grid = document.getElementById("resourcesGrid");
  if (!grid) return;

  try {
    const res = await fetch("/api/v1/resources", { credentials: "include" });
    const data = await res.json();

    if (!data.resources || data.resources.length === 0) {
      grid.innerHTML = `<p class="text-slate-400 text-center">No local resources found</p>`;
      return;
    }

    grid.innerHTML = data.resources
      .map((r) => {
        const osVersion = r.Metadata?.os || r.metadata?.os || "Unknown OS";
        const statusColor =
          r.Status === "online" ? "text-green-600" : "text-red-600";

        return `
          <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm hover:shadow-md transition">
            <div class="flex items-center justify-between mb-2">
              <div>
                <h3 class="text-slate-800 font-semibold">${r.Name}</h3>
                <p class="text-xs text-slate-500">${r.Host}</p>
              </div>
                  <div class="flex items-center gap-2">
                    <button class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded-lg connect-btn"
                            data-host="${r.Host}" data-resource-id="${r.ID}">
                      Connect
                    </button>
                  </div>
            </div>
            <p class="text-xs text-slate-600">${osVersion}</p>
          </div>
        `;
      })
      .join("");

    // Attach connect handlers (open modal even if no users; modal will disable Connect if none)
    document.querySelectorAll(".connect-btn").forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        // If user doesn't have resource write permission, show popup and don't navigate
        if (!window.hasResourceWrite) {
          showNoAccess("You don't have access for this resource. Please contact your admin.");
          return;
        }
        const host = e.currentTarget.dataset.host;
        await showUserSelectModal(host);
      });
    });
  } catch (err) {
    console.error("Error loading local resources:", err);
    grid.innerHTML = `<p class="text-red-500 text-center">Failed to load resources</p>`;
  }
}

// --------------------------- SELECT SSH USER MODAL --------------------------- //
async function showUserSelectModal(host) {
  const modal = document.getElementById("userModal");
  const select = document.getElementById("userSelect");
  const confirmBtn = document.getElementById("confirmUserSelect");
  const cancelBtn = document.getElementById("cancelUserSelect");

  modal.classList.remove("hidden");
  select.innerHTML = "<option>Loading allowed SSH users...</option>";

  try {
    const res = await fetch("/api/v1/users/connect-list", {
      credentials: "include",
    });
    const data = await res.json();

    select.innerHTML = "";
    const hasUsers = data.connect_user && data.connect_user.length > 0;
    if (hasUsers) {
      data.connect_user.forEach((u) => {
        const opt = document.createElement("option");
        opt.value = u;
        opt.textContent = u;
        select.appendChild(opt);
      });
      confirmBtn.disabled = false;
    } else {
      const opt = document.createElement("option");
      opt.textContent = "No SSH users available";
      select.appendChild(opt);
      confirmBtn.disabled = true;
    }
  } catch (err) {
    console.error("Failed to load SSH users:", err);
    select.innerHTML = "<option>Error loading users</option>";
    confirmBtn.disabled = true;
  }

  cancelBtn.onclick = () => modal.classList.add("hidden");

  confirmBtn.onclick = () => {
    if (confirmBtn.disabled) return;
    const selectedUser = select.value;
    modal.classList.add("hidden");
    openSSH(host, selectedUser);
  };
}

// --------------------------- ASSIGN ACCESS MODAL (ADMIN) --------------------------- //
async function openAssignAccessModal(userId) {
  const modal = document.getElementById('assignAccessModal');
  const info = document.getElementById('assignAccessInfo');
  const listEl = document.getElementById('assignAccessList');
  const saveBtn = document.getElementById('assignAccessSave');
  const cancelBtn = document.getElementById('assignAccessCancel');

  if (!modal || !listEl || !saveBtn || !cancelBtn) return;
  modal.classList.remove('hidden');
  info.textContent = `User ID: ${userId}`;
  listEl.innerHTML = 'Loading resources...';

  try {
    const [rRes, uRes] = await Promise.all([
      fetch('/api/v1/resources', { credentials: 'include' }),
      fetch('/api/v1/users', { credentials: 'include' }),
    ]);
    const rData = await rRes.json();
    const uData = await uRes.json();
    const resources = rData.resources || [];
    const users = uData.users || [];
    const user = users.find(uu => String(uu.ID) === String(userId) || String(uu.id) === String(userId));
    const assigned = user && (user.access_resources || user.resource_access || user.resources) ? (user.access_resources || user.resource_access || user.resources) : [];

    listEl.innerHTML = '';
    resources.forEach(r => {
      const id = `assign_res_${r.ID}`;
      // assigned may be array of ids or array of objects {resource_id, connect_user}
      let isChecked = false;
      let preUser = '';
      if (Array.isArray(assigned)) {
        assigned.forEach(a => {
          if (typeof a === 'number' || typeof a === 'string') {
            if (String(a) === String(r.ID)) isChecked = true;
          } else if (a && (a.resource_id || a.resourceId || a.id)) {
            const rid = a.resource_id || a.resourceId || a.id;
            if (String(rid) === String(r.ID)) {
              isChecked = true;
              preUser = a.connect_user || a.connectUser || a.username || '';
            }
          }
        });
      }

      const item = document.createElement('div');
      item.className = 'flex items-center gap-2 py-1';
      item.innerHTML = `
        <input type="checkbox" id="${id}" data-resource-id="${r.ID}" ${isChecked ? 'checked' : ''} />
        <label for="${id}" class="text-sm flex-1">${r.Name} (${r.Host})</label>
        <input type="text" placeholder="connect_user (optional)" class="connect-user-input border rounded px-2 py-1 text-sm w-36" data-resource-id="${r.ID}" value="${preUser}">
      `;
      listEl.appendChild(item);
    });

    cancelBtn.onclick = () => modal.classList.add('hidden');

    saveBtn.onclick = async () => {
      const checked = Array.from(listEl.querySelectorAll('input[type=checkbox]:checked'));
      const access = checked.map(c => {
        const rid = Number(c.getAttribute('data-resource-id'));
        const input = listEl.querySelector(`input.connect-user-input[data-resource-id="${rid}"]`);
        const username = input ? input.value.trim() : '';
        return { resource_id: rid, connect_user: username };
      }).filter(a => a.resource_id);
      saveBtn.disabled = true;
      saveBtn.textContent = 'Saving...';
      try {
        const res = await fetch(`/api/v1/users/${encodeURIComponent(userId)}/access`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ access }),
        });
        const data = await res.json();
        if (!res.ok) {
          alert('Failed to save: ' + (data.error || data.message || res.statusText));
          saveBtn.disabled = false;
          saveBtn.textContent = 'Save';
          return;
        }
        modal.classList.add('hidden');
        saveBtn.disabled = false;
        saveBtn.textContent = 'Save';
        // refresh users view to reflect changes
        await loadUsers();
        alert('Access updated');
      } catch (err) {
        console.error('Failed to save access', err);
        alert('Network error saving access');
        saveBtn.disabled = false;
        saveBtn.textContent = 'Save';
      }
    };

  } catch (err) {
    console.error('Failed to load resources/users for access assignment', err);
    listEl.innerHTML = '<div class="text-sm text-red-600">Failed to load resources</div>';
  }
}

// --------------------------- SSH CONNECT --------------------------- //
function openSSH(host, user) {
  const modal = document.getElementById("sshModal");
  const closeBtn = document.getElementById("closeSSH");
  const termEl = document.getElementById("terminal");

  modal.classList.remove("hidden");
  termEl.innerHTML = "";

  const term = new Terminal({
    cursorBlink: true,
    convertEol: true,
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
    fontSize: 14,
    theme: { background: "#0b1221", foreground: "#e5e7eb" },
  });

  const fit = new window.FitAddon.FitAddon();
  term.loadAddon(fit);
  term.open(termEl);
  fit.fit();

  const proto = location.protocol === "https:" ? "wss" : "ws";
  const url = `${proto}://${location.host}/api/v1/ws/ssh?host=${encodeURIComponent(
    host
  )}&port=22&user=${encodeURIComponent(user)}`;

  const ws = new WebSocket(url);
  ws.binaryType = "arraybuffer";

  ws.onopen = () => {
    const cols = term.cols || 120;
    const rows = term.rows || 32;
    ws.send(JSON.stringify({ op: "auth", cols, rows }));
  };

  ws.onmessage = (ev) => {
    if (ev.data instanceof ArrayBuffer) term.write(new Uint8Array(ev.data));
    else term.write(String(ev.data));
  };

  ws.onclose = () => term.write("\r\n\x1b[33m[Disconnected]\x1b[0m\r\n");
  ws.onerror = () => term.write("\r\n\x1b[31m[WebSocket error]\x1b[0m\r\n");

  term.onData((data) => {
    if (ws.readyState === WebSocket.OPEN)
      ws.send(new TextEncoder().encode(data));
  });

  window.addEventListener("resize", () => fit.fit());
  closeBtn.onclick = () => {
    try {
      ws.close();
    } catch {}
    modal.classList.add("hidden");
  };
}

// --------------------------- ADD RESOURCE MODAL --------------------------- //
function setupAddResourceModal() {
  const modal = document.getElementById("addResourceModal");
  const showBtn = document.getElementById("showAddResource");
  const cancelBtn = document.getElementById("cancelAddResource");

  if (showBtn && modal) showBtn.addEventListener("click", () => modal.classList.remove("hidden"));
  if (cancelBtn && modal) cancelBtn.addEventListener("click", () => modal.classList.add("hidden"));

  // Token UI inside Add Resource modal
  const showTokenSectionBtn = document.getElementById('showAddGenTokenSection');
  const tokenSection = document.getElementById('addGenTokenSection');
  const ttlInput = document.getElementById('addGenTokenTTL');
  const resultWrap = document.getElementById('addGenTokenResult');
  const resultVal = document.getElementById('addGenTokenValue');
  const createBtn = document.getElementById('createAddTokenBtn');
  const copyBtn = document.getElementById('copyAddGenToken');

  if (showTokenSectionBtn && tokenSection) {
    showTokenSectionBtn.addEventListener('click', async () => {
      // toggle visibility
      tokenSection.classList.toggle('hidden');

      // (resource selection removed) reset result area and TTL
      // reset result area
      if (resultWrap) resultWrap.classList.add('hidden');
      if (resultVal) resultVal.textContent = '';
      if (ttlInput) ttlInput.value = '60';
    });
  }

  if (createBtn) createBtn.addEventListener('click', async () => {
    const allowed = Array.isArray(window.currentUserPermissions) && window.currentUserPermissions.includes('resources:generate-token');
    if (!allowed) {
      alert('You are not authorized to create registration tokens');
      return;
    }

    const ttl = ttlInput ? parseInt(ttlInput.value || '0', 10) : 0;
    const payload = { resource_id: null, ttl_minutes: Number.isFinite(ttl) ? ttl : 0 };

    try {
      createBtn.disabled = true;
      createBtn.textContent = 'Creating...';
      const res = await fetch('/api/v1/agents/tokens', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (!res.ok) {
        // show server error message in-place if possible
        const errMsg = data && (data.error || data.message) ? (data.error || data.message) : (res.statusText || 'Failed to create token');
        if (resultWrap) {
          resultWrap.classList.remove('hidden');
          if (resultVal) resultVal.textContent = errMsg;
        } else {
          alert('Failed to create token: ' + errMsg);
        }
        createBtn.disabled = false;
        createBtn.textContent = 'Create Token';
        return;
      }

      const tokenText = data.token || data.Token || '';
      if (resultVal) resultVal.textContent = tokenText;
      if (resultWrap) resultWrap.classList.remove('hidden');
      // update extract and run command to include token and selected arch
      try {
        const archSel = document.getElementById('agentArchSelect');
        const arch = archSel ? archSel.value : 'linux-amd64';
        const binaryName = `teleport-agent-${arch}`;
        const tarName = `${binaryName}.tar.gz`;
        const extractEl = document.getElementById('cmdExtract');
        const runEl = document.getElementById('cmdRun');
        if (extractEl) extractEl.innerText = `tar -xzf ${tarName}`;
        if (runEl) runEl.innerText = `CONTROLLER_URL=${window.location.origin} AGENT_REG_TOKEN=${tokenText} ./${binaryName}`;
      } catch (e) {
        console.warn('Failed to update run commands with token', e);
      }
      createBtn.disabled = false;
      createBtn.textContent = 'Create Token';
    } catch (err) {
      console.error('Create token failed', err);
      alert('Network error creating token');
      createBtn.disabled = false;
      createBtn.textContent = 'Create Token';
    }
  });

  if (copyBtn) copyBtn.addEventListener('click', async () => {
    try {
      const text = resultVal ? resultVal.textContent.trim() : '';
      if (!text) return;
      await navigator.clipboard.writeText(text);
      copyBtn.textContent = 'Copied!';
      setTimeout(() => (copyBtn.textContent = 'Copy'), 2000);
    } catch (err) {
      console.error('Copy failed', err);
      copyBtn.textContent = 'Error';
      setTimeout(() => (copyBtn.textContent = 'Copy'), 2000);
    }
  });

}

// --------------------------- GENERATE TOKEN MODAL --------------------------- //
// previous standalone token modal removed — token UI is integrated into Add Resource modal

// --------------------------- NO ACCESS MODAL --------------------------- //
function setupNoAccessModal() {
  const modal = document.getElementById("noAccessModal");
  const closeBtn = document.getElementById("closeNoAccess");
  if (!modal) return;
  closeBtn.addEventListener("click", () => modal.classList.add("hidden"));
}

function showNoAccess(msg) {
  const modal = document.getElementById("noAccessModal");
  const msgEl = document.getElementById("noAccessMsg");
  if (!modal) {
    alert(msg);
    return;
  }
  if (msgEl) msgEl.textContent = msg;
  modal.classList.remove("hidden");
}

// ===== AUDIT TRAIL PAGE =====
async function loadAuditLogs() {
  const table = document.getElementById("auditTable");
  if (!table) return;

  try {
    const res = await fetch("/api/v1/audit", { credentials: "include" });
    const data = await res.json();

    table.innerHTML = "";

    const logs = data.logs || data.Audit || [];

    if (logs.length === 0) {
      table.innerHTML = `<tr><td colspan="6" class="py-4 text-center text-slate-400">No audit logs found.</td></tr>`;
      return;
    }
    
    logs.forEach((log, i) => {
      const row = `
        <tr class="border-b hover:bg-slate-50 transition">
          <td class="py-3 px-4">${i + 1}</td>
          <td class="py-3 px-4 text-slate-700">${log.initiator_name || log.User || "-"}</td>
          <td class="py-3 px-4 text-slate-700">${log.Action || "-"}</td>
          <td class="py-3 px-4 text-slate-700">${log.ResourceType || "-"}</td>
          <td class="py-3 px-4 text-slate-700">${log.IP || "-"}</td>
          <td class="py-3 px-4 text-slate-500">${new Date(log.CreatedAt).toLocaleString()}</td>
        </tr>`;
      table.insertAdjacentHTML("beforeend", row);
    });
  } catch (err) {
    console.error("Failed to load audit logs:", err);
    table.innerHTML = `<tr><td colspan="6" class="py-4 text-center text-red-500">Failed to load audit logs.</td></tr>`;
  }
}

// Hook refresh button
document.addEventListener("DOMContentLoaded", () => {
  if (document.getElementById("refreshAudit")) {
    loadAuditLogs();
    document.getElementById("refreshAudit").addEventListener("click", loadAuditLogs);
  }
});


// --------------------------- LOGOUT HANDLER --------------------------- //
function setupLogout() {
  const logoutBtn = document.getElementById("logoutBtn");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", () => {
      // Navigate to server-side logout which clears the HttpOnly cookie
      // and redirects back to the login page.
      window.location.href = "/logout";
    });
  }
}

// --------------------------- USER DROPDOWN --------------------------- //
function setupUserMenu() {
  const btn = document.getElementById("userMenuBtn");
  const dropdown = document.getElementById("menuDropdown");

  if (btn && dropdown) {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      dropdown.classList.toggle("hidden");
    });
    window.addEventListener("click", () => dropdown.classList.add("hidden"));
  }
}

// --------------------------- DASHBOARD RESOURCE CHART --------------------------- //
async function loadResourcesChart() {
  const canvas = document.getElementById("resourcesChart");
  if (!canvas) return;

  try {
    const res = await fetch("/api/v1/resources", {
      credentials: "include", // ✅ sends cookie with request
    });

    if (res.status === 401) {
      console.warn("⚠️ Unauthorized — redirecting to login");
      window.location.href = "/login";
      return;
    }

    const data = await res.json();
    const count = data.resources ? data.resources.length : 0;

    const ctx = canvas.getContext("2d");
    new Chart(ctx, {
      type: "doughnut",
      data: {
        labels: ["Resources"],
        datasets: [{
          data: [1],
          backgroundColor: ["#14b8a6", "#e5e7eb"], // teal + light gray
          cutout: "75%",
          borderWidth: 0,
        }],
      },
      options: {
        plugins: {
          legend: { display: false },
          tooltip: { enabled: false },
        },
      },
      plugins: [{
        id: "textCenter",
        beforeDraw(chart) {
          const { ctx, width } = chart;
          ctx.save();
          ctx.font = "bold 64px sans-serif";
          ctx.fillStyle = "#0f766e";
          ctx.textAlign = "center";
          ctx.textBaseline = "middle";
          ctx.fillText(count.toString(), width / 2, chart.chartArea.top + 80);
          ctx.restore();
        },
      }],
    });
  } catch (err) {
    console.error("Failed to load resources chart:", err);
  }
}

//Snippet
document.addEventListener("DOMContentLoaded", () => {
  // Detect controller URL automatically
  const controllerURL = window.location.origin;
  const cmdRun = document.getElementById("cmdRun");

  if (cmdRun) {
    cmdRun.innerText = `CONTROLLER_URL=${controllerURL} ./teleport-agent`;
  }

  // Handle file download (multiple architectures)
  const downloadBtn = document.getElementById("downloadAgent");
  const archSelect = document.getElementById("agentArchSelect");
  if (archSelect) {
    // Auto-detect user's arch and preselect if possible
    try {
      const ua = navigator.userAgent || "";
      const platform = navigator.platform || "";
      // simple heuristics
      if (/arm|aarch64/i.test(ua) || /arm|aarch64/i.test(platform)) archSelect.value = 'linux-arm64';
      else archSelect.value = 'linux-amd64';
    } catch (e) {}
    // update extract/run commands when arch changes
    archSelect.addEventListener('change', () => {
      const arch = archSelect.value || 'linux-amd64';
      const extractEl = document.getElementById('cmdExtract');
      const runEl = document.getElementById('cmdRun');
      const binaryName = `teleport-agent-${arch}`;
      const tarName = `${binaryName}.tar.gz`;
      if (extractEl) extractEl.innerText = `tar -xzf ${tarName}`;
      // if token present, include it
      const tokenEl = document.getElementById('addGenTokenValue');
      const token = tokenEl ? tokenEl.textContent.trim() : '';
      if (runEl) runEl.innerText = token ? `CONTROLLER_URL=${window.location.origin} AGENT_REG_TOKEN=${token} ./${binaryName}` : `CONTROLLER_URL=${window.location.origin} ./${binaryName}`;
    });
    // initialize extract/run commands for the current arch selection
    try {
      const arch0 = archSelect.value || 'linux-amd64';
      const binaryName0 = `teleport-agent-${arch0}`;
      const tarName0 = `${binaryName0}.tar.gz`;
      const extractEl0 = document.getElementById('cmdExtract');
      const runEl0 = document.getElementById('cmdRun');
      if (extractEl0) extractEl0.innerText = `tar -xzf ${tarName0}`;
      const tokenEl0 = document.getElementById('addGenTokenValue');
      const token0 = tokenEl0 ? tokenEl0.textContent.trim() : '';
      if (runEl0) runEl0.innerText = token0 ? `CONTROLLER_URL=${window.location.origin} AGENT_REG_TOKEN=${token0} ./${binaryName0}` : `CONTROLLER_URL=${window.location.origin} ./${binaryName0}`;
    } catch (e) {
      console.warn('Failed to initialize extract/run commands', e);
    }
  }

  if (downloadBtn) {
    downloadBtn.addEventListener("click", async () => {
      try {
        const arch = (archSelect && archSelect.value) ? archSelect.value : 'linux-amd64';
        // Map to file name convention
        const filename = `teleport-agent-${arch}.tar.gz`;
        const path = `/internal/shared/${filename}`;

        const response = await fetch(path);
        if (!response.ok) throw new Error("File not found or inaccessible: " + path);

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

        downloadBtn.textContent = "Downloaded!";
        downloadBtn.classList.replace("bg-green-600", "bg-green-700");
        setTimeout(() => (downloadBtn.textContent = "Download"), 2000);
      } catch (err) {
        console.error("Download failed:", err);
        alert('Failed to download agent: ' + (err.message || err));
        downloadBtn.textContent = "Error";
        downloadBtn.classList.replace("bg-green-600", "bg-red-600");
      }
    });
  }

  // Copy buttons functionality
  const copyButtons = [
    { btn: "copyExtract", code: "cmdExtract" },
    { btn: "copyRun", code: "cmdRun" },
  ];

  copyButtons.forEach(({ btn, code }) => {
    const button = document.getElementById(btn);
    const codeEl = document.getElementById(code);

    if (button && codeEl) {
      button.addEventListener("click", async () => {
        try {
          await navigator.clipboard.writeText(codeEl.innerText.trim());
          button.textContent = "Copied!";
          button.classList.remove("bg-blue-600");
          button.classList.add("bg-green-600");
          setTimeout(() => {
            button.textContent = "Copy";
            button.classList.remove("bg-green-600");
            button.classList.add("bg-blue-600");
          }, 2000);
        } catch (err) {
          console.error("Copy failed:", err);
          button.textContent = "Error";
        }
      });
    }
  });
});

async function loadAssignUsers() {
  const loading = document.getElementById("roles-loading");
  const table   = document.getElementById("roles-table");
  const empty   = document.getElementById("roles-empty-state");
  const tbody   = document.getElementById("roles-table-body");

  if (!tbody) return;

  // State awal
  if (loading) loading.classList.remove("hidden");
  if (table)   table.classList.add("hidden");
  if (empty)   empty.classList.add("hidden");

  try {
    const res = await fetch("/api/v1/assign/users", {
      method: "GET",
      credentials: "include",       // 🔑 penting biar cookie JWT ikut
      headers: { "Accept": "application/json" },
    });

    if (res.status === 401 || res.status === 403) {
      console.warn("Not authorized, redirecting to login");
      window.location.href = "/login";
      return;
    }

    if (!res.ok) {
      console.error("Failed to load assign users:", res.status, await res.text());
      if (loading) loading.textContent = "Failed to load data.";
      return;
    }

    const data = await res.json();
    const users = Array.isArray(data.users) ? data.users : data;

    tbody.innerHTML = "";

    if (!users || users.length === 0) {
      if (loading) loading.classList.add("hidden");
      if (empty)   empty.classList.remove("hidden");
      return;
    }

    users.forEach((u, idx) => {
      // adaptasi field tergantung JSON kamu
      const name  = u.name || u.Name || "-";
      const email = u.email || u.Email || "-";
      const roles = (u.roles || u.Roles || [])
        .map(r => r.name || r.Name || r.slug || r.Slug)
        .join(", ") || "-";

      const roleIds = (u.roles || u.Roles || []).map(r => r.id || r.ID || r.Id).filter(Boolean).join(",");

      const tr = document.createElement("tr");
      tr.className = "hover:bg-slate-50 transition";

      tr.innerHTML = `
        <td class="px-5 py-3 text-sm text-slate-500">${idx + 1}</td>
        <td class="px-5 py-3 text-sm font-medium text-slate-900">${name}</td>
        <td class="px-5 py-3 text-sm text-slate-700">${email}</td>
        <td class="px-5 py-3 text-sm text-slate-700">${roles}</td>
        <td class="px-5 py-3 text-right">
          <button
            class="assign-role-btn text-blue-600 hover:underline text-sm"
            data-user-id="${u.id || u.ID}"
            data-user-name="${name}"
            data-user-roles="${roleIds}"
          >
            Manage
          </button>
        </td>
      `;
      tbody.appendChild(tr);
    });

    if (loading) loading.classList.add("hidden");
    if (table)   table.classList.remove("hidden");
  } catch (err) {
    console.error("Error calling /api/v1/assign/users:", err);
    if (loading) loading.textContent = "Error loading data.";
  }
}

// Setup handlers for Manage buttons (call after loadAssignUsers completes)
function setupAssignRoleButtons() {
  document.querySelectorAll('.assign-role-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const userId = btn.getAttribute('data-user-id');
      const userName = btn.getAttribute('data-user-name');
      const userRolesCSV = btn.getAttribute('data-user-roles') || '';
      const currentRoleIds = userRolesCSV === '' ? [] : userRolesCSV.split(',').map(s => s.trim()).filter(Boolean).map(Number);

      // Show modal
      const modal = document.getElementById('manageRolesModal');
      const info = document.getElementById('manageRolesUserInfo');
      const list = document.getElementById('manageRolesList');
      modal.classList.remove('hidden');
      info.textContent = `Manage roles for: ${userName} (ID: ${userId})`;
      list.innerHTML = 'Loading roles...';

      try {
        const res = await fetch('/api/v1/roles', { credentials: 'include', headers: { 'Accept': 'application/json' } });
        if (!res.ok) {
          list.innerHTML = 'Failed to load roles.';
          return;
        }
        const data = await res.json();
        const roles = data.roles || [];
        list.innerHTML = '';
        roles.forEach(r => {
          const rid = r.id || r.ID || r.Id;
          const checked = currentRoleIds.includes(Number(rid));
          const id = `role_chk_${rid}`;
          const div = document.createElement('div');
          div.className = 'flex items-center gap-2';
          div.innerHTML = `
            <input type="checkbox" id="${id}" data-role-id="${rid}" ${checked ? 'checked' : ''} class="h-4 w-4">
            <label for="${id}" class="text-sm text-slate-700">${r.name || r.Name || r.slug || r.Slug}</label>
          `;
          list.appendChild(div);
        });

        // attach save handler (remove previous)
        const saveBtn = document.getElementById('saveManageRoles');
        saveBtn.onclick = async () => {
          const checked = Array.from(list.querySelectorAll('input[type=checkbox]:checked')).map(cb => Number(cb.getAttribute('data-role-id')));
          try {
            const resp = await fetch(`/api/v1/users/${userId}/roles`, {
              method: 'POST',
              credentials: 'include',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ role_ids: checked }),
            });
            if (!resp.ok) {
              alert('Failed to save roles: ' + resp.status);
              return;
            }
            modal.classList.add('hidden');
            // refresh list
            loadAssignUsers();
          } catch (err) {
            console.error('Save roles error', err);
            alert('Error saving roles');
          }
        };

        // cancel handler
        document.getElementById('cancelManageRoles').onclick = () => modal.classList.add('hidden');

      } catch (err) {
        console.error('Failed to load roles list', err);
        list.innerHTML = 'Error loading roles.';
      }
    });
  });
}

// Ensure we wire buttons after DOM updates: simple interval observer
const assignObserver = setInterval(() => {
  if (document.querySelectorAll && document.querySelectorAll('.assign-role-btn').length > 0) {
    setupAssignRoleButtons();
    clearInterval(assignObserver);
  }
}, 300);


