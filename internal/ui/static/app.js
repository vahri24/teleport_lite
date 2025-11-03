// internal/ui/static/app.js

document.addEventListener("DOMContentLoaded", () => {
  console.log("🚀 app.js loaded");

  const onDashboard = window.location.pathname === "/dashboard";

  // If user is on dashboard, check if token cookie exists
  if (onDashboard) {
    checkDashboardAccess();
  }

  // Handle login form
  const loginForm = document.getElementById("loginForm");
  if (loginForm) {
    loginForm.addEventListener("submit", async (e) => {
      e.preventDefault();

      const formData = new FormData(loginForm);
      const payload = Object.fromEntries(formData.entries());

      try {
        const res = await fetch("/api/v1/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
          credentials: "include", // ✅ store cookie
        });

        const data = await res.json();
        console.log("Login response:", data);

        if (res.ok && data.token) {
          console.log("✅ Login successful!");
          window.location.href = "/dashboard";
        } else {
          alert("❌ Login failed: " + (data.error || "Invalid credentials"));
        }
      } catch (err) {
        console.error("⚠️ Login error:", err);
        alert("Network error, please check console");
      }
    });
  }

  // Setup logout + user menu
  setupLogout();
  setupUserMenu();

  // Load dashboard data if on dashboard
  if (onDashboard) {
    loadTable("/api/v1/roles", "rolesTable", 3);
    loadTable("/api/v1/resources", "resourcesTable", 4);
    loadTable("/api/v1/audit", "auditTable", 5);
  }
});

// ----------- Auth Guard for Dashboard -----------
async function checkDashboardAccess() {
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

// ----------- Dynamic Table Loader -----------
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
        (r) =>
          `<tr class="border-b last:border-0">
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

// ----------- Logout + User Menu -----------
function setupLogout() {
  const logoutBtn = document.getElementById("logoutBtn");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", async () => {
      // Delete token cookie via backend (optional) or client-side redirect
      alert("👋 Logged out successfully!");
      document.cookie = "token=; Max-Age=0; path=/"; // clear cookie
      window.location.href = "/login";
    });
  }
}

function setupUserMenu() {
  const btn = document.getElementById("userMenuBtn");
  const dropdown = document.getElementById("menuDropdown");

  if (btn && dropdown) {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      dropdown.classList.toggle("hidden");
    });

    window.addEventListener("click", () => {
      dropdown.classList.add("hidden");
    });
  }
}


// ----------- Refresh Buttons for Dashboard -----------

document.addEventListener("DOMContentLoaded", () => {
  const refreshRoles = document.getElementById("refreshRoles");
  const refreshResources = document.getElementById("refreshResources");
  const refreshAudit = document.getElementById("refreshAudit");

  if (refreshRoles)
    refreshRoles.addEventListener("click", () => {
      console.log("Refreshing roles...");
      loadTable("/api/v1/roles", "rolesTable", 3);
    });

  if (refreshResources)
    refreshResources.addEventListener("click", () => {
      console.log("Refreshing resources...");
      loadTable("/api/v1/resources", "resourcesTable", 4);
    });

  if (refreshAudit)
    refreshAudit.addEventListener("click", () => {
      console.log("Refreshing audit logs...");
      loadTable("/api/v1/audit", "auditTable", 5);
    });
});