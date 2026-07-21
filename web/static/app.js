const nsEl = document.getElementById("ns");
const headEl = document.getElementById("head");
const bodyEl = document.getElementById("body");
const ctxEl = document.getElementById("ctx");
const detailEl = document.getElementById("detail");
const detailKind = document.getElementById("detail-kind");
const detailTitle = document.getElementById("detail-title");
const detailMeta = document.getElementById("detail-meta");
const detailConditions = document.getElementById("detail-conditions");
const detailEvents = document.getElementById("detail-events");
const detailLogs = document.getElementById("detail-logs");
const detailLogPod = document.getElementById("detail-log-pod");
const detailClose = document.getElementById("detail-close");

let kind = "deployments";
let selected = null;

document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    kind = btn.dataset.kind;
    closeDetail();
    refreshTable();
  });
});

nsEl.addEventListener("change", () => {
  closeDetail();
  refreshTable();
});
detailClose.addEventListener("click", closeDetail);

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
      const items = data.items || [];
      bodyEl.innerHTML =
        items
          .map(
            (d) =>
              `<tr data-name="${escAttr(d.name)}" class="${selected === d.name ? "active" : ""}"><td>${esc(d.name)}</td><td>${esc(d.ready)}</td><td>${d.replicas ?? 0}</td><td>${esc(d.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="4">No deployments</td></tr>`;
    } else {
      headEl.innerHTML =
        "<tr><th>Name</th><th>Phase</th><th>Ready</th><th>Restarts</th><th>Node</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/pods`);
      const items = data.items || [];
      bodyEl.innerHTML =
        items
          .map(
            (p) =>
              `<tr data-name="${escAttr(p.name)}" class="${selected === p.name ? "active" : ""}"><td>${esc(p.name)}</td><td>${esc(p.phase)}</td><td>${esc(p.ready)}</td><td>${p.restarts ?? 0}</td><td>${esc(p.node || "—")}</td><td>${esc(p.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="6">No pods</td></tr>`;
    }
    bodyEl.querySelectorAll("tr[data-name]").forEach((tr) => {
      tr.addEventListener("click", () => openDetail(tr.dataset.name));
    });
  } catch (e) {
    bodyEl.innerHTML = `<tr><td colspan="6">${esc(String(e.message || e))}</td></tr>`;
  }
}

async function openDetail(name) {
  const ns = nsEl.value;
  selected = name;
  bodyEl.querySelectorAll("tr[data-name]").forEach((tr) => {
    tr.classList.toggle("active", tr.dataset.name === name);
  });
  detailEl.hidden = false;
  detailTitle.textContent = name;
  detailKind.textContent = kind === "deployments" ? "Deployment" : "Pod";
  detailMeta.textContent = "Loading…";
  detailConditions.innerHTML = "";
  detailEvents.innerHTML = "";
  detailLogs.textContent = "…";
  detailLogPod.textContent = "";
  try {
    const path =
      kind === "deployments"
        ? `/api/v1/namespaces/${encodeURIComponent(ns)}/deployments/${encodeURIComponent(name)}`
        : `/api/v1/namespaces/${encodeURIComponent(ns)}/pods/${encodeURIComponent(name)}`;
    const d = await fetchJSON(path);
    const bits = [`ns ${d.namespace}`, d.age ? `age ${d.age}` : null];
    if (d.ready) bits.push(`ready ${d.ready}`);
    if (d.phase) bits.push(d.phase);
    if (d.node) bits.push(`node ${d.node}`);
    if (d.restarts != null) bits.push(`restarts ${d.restarts}`);
    detailMeta.textContent = bits.filter(Boolean).join(" · ");
    detailConditions.innerHTML = renderKVList(d.conditions, (c) =>
      `${c.type}=${c.status}${c.reason ? ` (${c.reason})` : ""}${c.message ? ` — ${c.message}` : ""}`
    );
    detailEvents.innerHTML = renderKVList(d.events, (e) =>
      `[${e.type || "?"}] ${e.reason || ""} · ${e.age || ""} — ${e.message || ""}`
    );
    if (d.log_pod) detailLogPod.textContent = `(from ${d.log_pod})`;
    detailLogs.textContent = d.logs && String(d.logs).trim() ? d.logs : "(no logs)";
  } catch (e) {
    detailMeta.textContent = String(e.message || e);
    detailLogs.textContent = "—";
  }
}

function closeDetail() {
  selected = null;
  detailEl.hidden = true;
  bodyEl.querySelectorAll("tr.active").forEach((tr) => tr.classList.remove("active"));
}

function renderKVList(items, fmt) {
  if (!items || !items.length) return "<li>None</li>";
  return items.map((item) => `<li>${esc(fmt(item))}</li>`).join("");
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

function escAttr(s) {
  return esc(s).replaceAll("'", "&#39;");
}

init();
