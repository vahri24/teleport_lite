// internal/ui/static/app.js

document.addEventListener("DOMContentLoaded", () => {
  console.log("🚀 app.js loaded");

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
        });

        const data = await res.json();
        if (res.ok && data.token) {
          localStorage.setItem("token", data.token);
          alert("✅ Login successful!");
          window.location.href =
            "/dashboard?name=" +
            encodeURIComponent(data.user.name || "Admin") +
            "&org_id=" +
            encodeURIComponent(data.user.org_id || 1);
        } else {
          alert("❌ Login failed: " + (data.error || "Invalid credentials"));
        }
      } catch (err) {
        console.error("Login error:", err);
        alert("Network error, check console for details");
      }
    });
  }

  // Load dashboard data if applicable
  const rolesTable = document.getElementById("rolesTable");
  if (rolesTable) loadRoles();

  const resourcesTable = document.getElementById("resourcesTable");
  if (resourcesTable) loadResources();

  const auditTable = document.getElementById("auditTable");
  if (auditTable) loadAuditLogs();
});

// ----------- Data Loader Functions -----------

async function loadRoles() {
  const token = localStorage.getItem("token");
  if (!token) {
    alert("⚠️ You must login first");
    window.location.href = "/login";
    return;
  }

  try {
    const res = await fetch("/api/v1/roles", {
      headers: { Authorization: "Bearer " + token },
    });

    const data = await res.json();
    const table = document.getElementById("rolesTable");
    if (data.roles && data.roles.length > 0) {
      table.innerHTML = data.roles
        .map(
          (r) =>
            `<tr><td>${r.ID}</td><td>${r.Name}</td><td>${r.Slug}</td></tr>`
        )
        .join("");
    } else {
      table.innerHTML = "<tr><td colspan='3'>No roles found</td></tr>";
    }
  } catch (err) {
    console.error(err);
  }
}

async function loadResources() {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const res = await fetch("/api/v1/resources", {
      headers: { Authorization: "Bearer " + token },
    });
    const data = await res.json();

    const table = document.getElementById("resourcesTable");
    if (data.resources && data.resources.length > 0) {
      table.innerHTML = data.resources
        .map(
          (r) =>
            `<tr><td>${r.ID}</td><td>${r.Name}</td><td>${r.Type}</td><td>${r.ExternalRef || "-"}</td></tr>`
        )
        .join("");
    } else {
      table.innerHTML = "<tr><td colspan='4'>No resources found</td></tr>";
    }
  } catch (err) {
    console.error("Error loading resources:", err);
  }
}

async function loadAuditLogs() {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const res = await fetch("/api/v1/audit", {
      headers: { Authorization: "Bearer " + token },
    });
    const data = await res.json();

    const table = document.getElementById("auditTable");
    if (data.logs && data.logs.length > 0) {
      table.innerHTML = data.logs
        .map(
          (a) =>
            `<tr><td>${a.ID}</td><td>${a.UserID}</td><td>${a.Action}</td><td>${a.ResourceType}</td><td>${a.CreatedAt}</td></tr>`
        )
        .join("");
    } else {
      table.innerHTML = "<tr><td colspan='5'>No audit logs found</td></tr>";
    }
  } catch (err) {
    console.error("Error loading audit logs:", err);
  }
}
