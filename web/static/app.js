const nsEl = document.getElementById("ns");
const headEl = document.getElementById("head");
const bodyEl = document.getElementById("body");
const ctxEl = document.getElementById("ctx");
let kind = "deployments";

document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    kind = btn.dataset.kind;
    refreshTable();
  });
});

nsEl.addEventListener("change", refreshTable);

async function init() {
  const health = await fetchJSON("/api/v1/healthz");
  ctxEl.textContent = `context: ${health.context || "—"}`;
  const ns = await fetchJSON("/api/v1/namespaces");
  nsEl.innerHTML = "";
  const items = ns.items || [];
  for (const n of items) {
    const opt = document.createElement("option");
    opt.value = n.name;
    opt.textContent = n.name;
    nsEl.appendChild(opt);
  }
  if (!items.length) {
    bodyEl.innerHTML = `<tr><td colspan="6">No namespaces (RBAC or empty cluster)</td></tr>`;
    return;
  }
  const preferred = items.find((n) => n.name === "default") || items[0];
  nsEl.value = preferred.name;
  await refreshTable();
}

async function refreshTable() {
  const ns = nsEl.value;
  if (!ns) return;
  bodyEl.innerHTML = `<tr><td colspan="6">Loading…</td></tr>`;
  try {
    if (kind === "deployments") {
      headEl.innerHTML =
        "<tr><th>Name</th><th>Ready</th><th>Replicas</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/deployments`);
      bodyEl.innerHTML = (data.items || [])
        .map(
          (d) =>
            `<tr><td>${esc(d.name)}</td><td>${esc(d.ready)}</td><td>${d.replicas ?? 0}</td><td>${esc(d.age)}</td></tr>`
        )
        .join("") || `<tr><td colspan="4">No deployments</td></tr>`;
    } else {
      headEl.innerHTML =
        "<tr><th>Name</th><th>Phase</th><th>Ready</th><th>Restarts</th><th>Node</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/pods`);
      bodyEl.innerHTML = (data.items || [])
        .map(
          (p) =>
            `<tr><td>${esc(p.name)}</td><td>${esc(p.phase)}</td><td>${esc(p.ready)}</td><td>${p.restarts ?? 0}</td><td>${esc(p.node || "—")}</td><td>${esc(p.age)}</td></tr>`
        )
        .join("") || `<tr><td colspan="6">No pods</td></tr>`;
    }
  } catch (e) {
    bodyEl.innerHTML = `<tr><td colspan="6">${esc(String(e.message || e))}</td></tr>`;
  }
}

async function fetchJSON(path) {
  const res = await fetch(path);
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || res.statusText);
  return body;
}

function esc(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

init();
