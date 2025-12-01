// --------------------------- MAIN APP ENTRY --------------------------- //
document.addEventListener("DOMContentLoaded", () => {
  console.log("🚀 app.js loaded");

  const path = window.location.pathname;
  const onDashboard = path === "/dashboard";
  const onResources = path === "/resources";
  const onUsers = path === "/users";
  const onAudit = path === "/audit";
  const onRoles = path === "/roles";
  // Protect dashboard and resources routes
  if (onDashboard || onResources || onAudit || onUsers || onRoles) checkAuth();

  // Login handler
  const loginForm = document.getElementById("loginForm");
  if (loginForm) loginForm.addEventListener("submit", handleLogin);

  // Logout & dropdown menu
  setupLogout();
  setupUserMenu();

  // Dashboard loaders
  if (onDashboard) {
    loadTable("/api/v1/roles", "rolesTable", 3);
    loadTable("/api/v1/resources", "resourcesTable", 4);
    loadResourcesChart();
  }

  // Users loaders
  if (onUsers) {
    console.log("👤 Loading real Users...");
    loadUsers();
    setupAddUserModal();
  }

  // Resources page handlers
  if (onResources) {
    console.log("🔌 Loading real local resources...");
    loadLocalResources();
    setupAddResourceModal();
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

    const records =
      data.users || data.roles || data.resources || data.logs || [];

    if (records.length === 0) {
      table.innerHTML = `<tr><td colspan="${columns}" class="py-4 text-center text-slate-400">No data found</td></tr>`;
      return;
    }

    table.innerHTML = records
      .map(
        (r) => `
        <tr class="border-b last:border-0">
          ${Object.values(r)
            .slice(0, columns)
            .map(
              (v) =>
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
      table.innerHTML = `<tr><td colspan="5" class="py-4 text-center text-slate-400">No users found.</td></tr>`;
      return;
    }

    data.users.forEach((u, i) => {
      const row = `
        <tr class="border-b hover:bg-slate-50 transition">
          <td class="py-3 px-4">${i + 1}</td>
          <td class="py-3 px-4 font-medium text-slate-700">${u.Name}</td>
          <td class="py-3 px-4 text-slate-600">${u.Email}</td>
          <td class="py-3 px-4">${u.role || "-"}</td>
          <td class="py-3 px-4">${u.status || "active"}</td>
        </tr>`;
      table.insertAdjacentHTML("beforeend", row);
    });
  } catch (err) {
    console.error("Failed to load users:", err);
    table.innerHTML = `<tr><td colspan="5" class="py-4 text-center text-red-500">Failed to load users.</td></tr>`;
  }
}

// Hook refresh button
document.addEventListener("DOMContentLoaded", () => {
  if (document.getElementById("refreshUsers")) {
    loadUsers();
    document.getElementById("refreshUsers").addEventListener("click", loadUsers);
  }
});

// --------------------------- ADD User MODAL --------------------------- //
function setupAddUserModal() {
  const modal = document.getElementById("addUserModal");
  const showBtn = document.getElementById("showAddUser");
  const cancelBtn = document.getElementById("cancelAddUser");
  const form = document.getElementById("addUserForm");

  if (showBtn && modal) showBtn.addEventListener("click", () => modal.classList.remove("hidden"));
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
              <button class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded-lg connect-btn"
                      data-host="${r.Host}">
                Connect
              </button>
            </div>
            <p class="text-xs text-slate-600">${osVersion}</p>
          </div>
        `;
      })
      .join("");

    // Attach connect handlers
    document.querySelectorAll(".connect-btn").forEach((btn) => {
      btn.addEventListener("click", async (e) => {
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
    if (data.connect_user && data.connect_user.length > 0) {
      data.connect_user.forEach((u) => {
        const opt = document.createElement("option");
        opt.value = u;
        opt.textContent = u;
        select.appendChild(opt);
      });
    } else {
      const opt = document.createElement("option");
      opt.textContent = "No SSH users available";
      select.appendChild(opt);
    }
  } catch (err) {
    console.error("Failed to load SSH users:", err);
    select.innerHTML = "<option>Error loading users</option>";
  }

  cancelBtn.onclick = () => modal.classList.add("hidden");

  confirmBtn.onclick = () => {
    const selectedUser = select.value;
    modal.classList.add("hidden");
    openSSH(host, selectedUser);
  };
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
      alert("👋 Logged out successfully!");
      document.cookie = "token=; Max-Age=0; path=/";
      window.location.href = "/login";
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

  // Handle file download
  const downloadBtn = document.getElementById("downloadAgent");
  if (downloadBtn) {
    downloadBtn.addEventListener("click", async () => {
      try {
        const response = await fetch("/internal/shared/teleport-agent.tar.gz");
        if (!response.ok) throw new Error("File not found or inaccessible.");

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "teleport-agent.tar.gz";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

        downloadBtn.textContent = "Downloaded!";
        downloadBtn.classList.replace("bg-green-600", "bg-green-700");
        setTimeout(() => (downloadBtn.textContent = "Download"), 2000);
      } catch (err) {
        console.error("Download failed:", err);
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


