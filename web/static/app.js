const nsEl = document.getElementById("ns");
const nsWrap = document.getElementById("ns-wrap");
const headEl = document.getElementById("head");
const bodyEl = document.getElementById("body");
const tableWrap = document.getElementById("table-wrap");
const overviewEl = document.getElementById("overview");
const ctxEl = document.getElementById("ctx");
const viewTitle = document.getElementById("view-title");
const viewEyebrow = document.getElementById("view-eyebrow");
const detailEl = document.getElementById("detail");
const detailKind = document.getElementById("detail-kind");
const detailTitle = document.getElementById("detail-title");
const detailMeta = document.getElementById("detail-meta");
const detailConditions = document.getElementById("detail-conditions");
const detailEvents = document.getElementById("detail-events");
const detailLogs = document.getElementById("detail-logs");
const detailLogPod = document.getElementById("detail-log-pod");
const logsWrap = document.getElementById("logs-wrap");
const detailClose = document.getElementById("detail-close");
const handoffPresets = document.getElementById("handoff-presets");
const handoffPrompt = document.getElementById("handoff-prompt");
const handoffCopy = document.getElementById("handoff-copy");
const handoffCmd = document.getElementById("handoff-cmd");
const handoffStatus = document.getElementById("handoff-status");

const VIEWS = {
  cluster: { title: "Cluster", eyebrow: "Overview", namespaced: false },
  nodes: { title: "Nodes", eyebrow: "Cluster scope", namespaced: false },
  deployments: { title: "Deployments", eyebrow: "Workloads", namespaced: true },
  replicasets: { title: "ReplicaSets", eyebrow: "Workloads", namespaced: true },
  pods: { title: "Pods", eyebrow: "Workloads", namespaced: true },
};

let view = "cluster";
let selected = null;
let kubeContext = "";

document.querySelectorAll(".nav-item").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".nav-item").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    view = btn.dataset.view;
    closeDetail();
    renderView();
  });
});

nsEl.addEventListener("change", () => {
  closeDetail();
  if (VIEWS[view].namespaced) refreshTable();
});
detailClose.addEventListener("click", closeDetail);
handoffPrompt.addEventListener("input", refreshHandoffCmd);
handoffCopy.addEventListener("click", async () => {
  const cmd = handoffCmd.textContent || "";
  if (!cmd || cmd === "—") return;
  try {
    await navigator.clipboard.writeText(cmd);
    handoffStatus.textContent = "Copied — paste in a terminal";
  } catch {
    handoffStatus.textContent = "Copy failed — select the command manually";
  }
});

async function init() {
  const health = await fetchJSON("/api/v1/healthz");
  kubeContext = health.context || "";
  ctxEl.textContent = kubeContext || "—";
  const ns = await fetchJSON("/api/v1/namespaces");
  nsEl.innerHTML = "";
  const items = ns.items || [];
  for (const n of items) {
    const opt = document.createElement("option");
    opt.value = n.name;
    opt.textContent = n.name;
    nsEl.appendChild(opt);
  }
  if (items.length) {
    const preferred = items.find((n) => n.name === "default") || items[0];
    nsEl.value = preferred.name;
  }
  await renderView();
}

async function renderView() {
  const meta = VIEWS[view];
  viewTitle.textContent = meta.title;
  viewEyebrow.textContent = meta.eyebrow;
  nsWrap.hidden = !meta.namespaced;
  closeDetail();

  if (view === "cluster") {
    tableWrap.hidden = true;
    overviewEl.hidden = false;
    await loadOverview();
    return;
  }

  overviewEl.hidden = true;
  tableWrap.hidden = false;
  await refreshTable();
}

async function loadOverview() {
  overviewEl.innerHTML = `<div class="stat"><p class="label">Loading</p><p class="value">…</p></div>`;
  try {
    const o = await fetchJSON("/api/v1/overview");
    overviewEl.innerHTML = [
      stat("Context", o.context || "—"),
      stat("Nodes ready", `${o.nodes_ready ?? "—"} / ${o.nodes ?? "—"}`),
      stat("Namespaces", String(o.namespaces ?? "—")),
    ].join("");
  } catch (e) {
    overviewEl.innerHTML = `<div class="stat"><p class="label">Error</p><p class="value">${esc(String(e.message || e))}</p></div>`;
  }
}

function stat(label, value) {
  return `<div class="stat"><p class="label">${esc(label)}</p><p class="value">${esc(value)}</p></div>`;
}

async function refreshTable() {
  bodyEl.innerHTML = `<tr><td colspan="6">Loading…</td></tr>`;
  try {
    if (view === "nodes") {
      headEl.innerHTML =
        "<tr><th>Name</th><th>Ready</th><th>Roles</th><th>Version</th><th>OS</th><th>Age</th></tr>";
      const data = await fetchJSON("/api/v1/nodes");
      bodyEl.innerHTML =
        (data.items || [])
          .map(
            (n) =>
              `<tr data-name="${escAttr(n.name)}"><td>${esc(n.name)}</td><td>${n.ready ? "True" : "False"}</td><td>${esc(n.roles)}</td><td>${esc(n.version)}</td><td>${esc(n.os)}/${esc(n.arch)}</td><td>${esc(n.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="6">No nodes</td></tr>`;
    } else if (view === "deployments") {
      const ns = nsEl.value;
      headEl.innerHTML =
        "<tr><th>Name</th><th>Ready</th><th>Replicas</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/deployments`);
      bodyEl.innerHTML =
        (data.items || [])
          .map(
            (d) =>
              `<tr data-name="${escAttr(d.name)}"><td>${esc(d.name)}</td><td>${esc(d.ready)}</td><td>${d.replicas ?? 0}</td><td>${esc(d.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="4">No deployments</td></tr>`;
    } else if (view === "replicasets") {
      const ns = nsEl.value;
      headEl.innerHTML =
        "<tr><th>Name</th><th>Ready</th><th>Owner</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/replicasets`);
      bodyEl.innerHTML =
        (data.items || [])
          .map(
            (rs) =>
              `<tr data-name="${escAttr(rs.name)}"><td>${esc(rs.name)}</td><td>${esc(rs.ready)}</td><td>${esc(rs.owner)}</td><td>${esc(rs.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="4">No replicasets</td></tr>`;
    } else if (view === "pods") {
      const ns = nsEl.value;
      headEl.innerHTML =
        "<tr><th>Name</th><th>Phase</th><th>Ready</th><th>Restarts</th><th>Node</th><th>Age</th></tr>";
      const data = await fetchJSON(`/api/v1/namespaces/${encodeURIComponent(ns)}/pods`);
      bodyEl.innerHTML =
        (data.items || [])
          .map(
            (p) =>
              `<tr data-name="${escAttr(p.name)}"><td>${esc(p.name)}</td><td>${esc(p.phase)}</td><td>${esc(p.ready)}</td><td>${p.restarts ?? 0}</td><td>${esc(p.node || "—")}</td><td>${esc(p.age)}</td></tr>`
          )
          .join("") || `<tr><td colspan="6">No pods</td></tr>`;
    }
    bodyEl.querySelectorAll("tr[data-name]").forEach((tr) => {
      tr.addEventListener("click", () => openDetail(tr.dataset.name));
      if (selected === tr.dataset.name) tr.classList.add("active");
    });
  } catch (e) {
    bodyEl.innerHTML = `<tr><td colspan="6">${esc(String(e.message || e))}</td></tr>`;
  }
}

async function openDetail(name) {
  selected = name;
  bodyEl.querySelectorAll("tr[data-name]").forEach((tr) => {
    tr.classList.toggle("active", tr.dataset.name === name);
  });
  detailEl.hidden = false;
  detailTitle.textContent = name;
  detailKind.textContent = detailKindLabel();
  detailMeta.textContent = "Loading…";
  detailConditions.innerHTML = "";
  detailEvents.innerHTML = "";
  detailLogs.textContent = "…";
  detailLogPod.textContent = "";
  handoffStatus.textContent = "";
  const showLogs = view === "pods" || view === "deployments";
  logsWrap.hidden = !showLogs;
  setupHandoff(name);
  try {
    const d = await fetchJSON(detailPath(name));
    const bits = [];
    if (d.namespace) bits.push(`ns ${d.namespace}`);
    if (d.age) bits.push(`age ${d.age}`);
    if (d.ready != null && d.ready !== true && d.ready !== false) bits.push(`ready ${d.ready}`);
    if (d.ready === true || d.ready === false) bits.push(d.ready ? "Ready" : "NotReady");
    if (d.phase) bits.push(d.phase);
    if (d.node) bits.push(`node ${d.node}`);
    if (d.roles) bits.push(`roles ${d.roles}`);
    if (d.owner) bits.push(`owner ${d.owner}`);
    if (d.restarts != null) bits.push(`restarts ${d.restarts}`);
    if (d.info?.kubelet) bits.push(d.info.kubelet);
    detailMeta.textContent = bits.join(" · ") || "—";
    detailConditions.innerHTML = renderKVList(d.conditions, (c) =>
      `${c.type}=${c.status}${c.reason ? ` (${c.reason})` : ""}${c.message ? ` — ${c.message}` : ""}`
    );
    detailEvents.innerHTML = renderKVList(d.events, (e) =>
      `[${e.type || "?"}] ${e.reason || ""} · ${e.age || ""} — ${e.message || ""}`
    );
    if (showLogs) {
      if (d.log_pod) detailLogPod.textContent = `(from ${d.log_pod})`;
      detailLogs.textContent = d.logs && String(d.logs).trim() ? d.logs : "(no logs)";
    }
  } catch (e) {
    detailMeta.textContent = String(e.message || e);
    detailLogs.textContent = "—";
  }
}

function detailKindLabel() {
  return (
    {
      nodes: "Node",
      deployments: "Deployment",
      replicasets: "ReplicaSet",
      pods: "Pod",
    }[view] || "Resource"
  );
}

function detailPath(name) {
  const ns = nsEl.value;
  if (view === "nodes") return `/api/v1/nodes/${encodeURIComponent(name)}`;
  if (view === "deployments")
    return `/api/v1/namespaces/${encodeURIComponent(ns)}/deployments/${encodeURIComponent(name)}`;
  if (view === "replicasets")
    return `/api/v1/namespaces/${encodeURIComponent(ns)}/replicasets/${encodeURIComponent(name)}`;
  return `/api/v1/namespaces/${encodeURIComponent(ns)}/pods/${encodeURIComponent(name)}`;
}

function setupHandoff(name) {
  const ns = nsEl.value;
  let presets = [];
  if (view === "nodes") {
    presets = [`how many pods are on node ${name}`, `describe node ${name}`];
  } else if (view === "deployments") {
    presets = [
      `explain why ${name} is not ready`,
      `logs ${name}`,
      `describe ${name}`,
      `scale ${name} to 2`,
    ];
  } else if (view === "replicasets") {
    presets = [`describe replicaset ${name}`, `list pods for ${name}`];
  } else if (view === "pods") {
    presets = [`explain why pod ${name} is failing`, `logs ${name}`, `describe pod ${name}`];
  }
  handoffPresets.innerHTML = "";
  for (const p of presets) {
    const b = document.createElement("button");
    b.type = "button";
    b.textContent = p;
    b.addEventListener("click", () => {
      handoffPrompt.value = p;
      refreshHandoffCmd();
    });
    handoffPresets.appendChild(b);
  }
  handoffPrompt.value = presets[0] || "";
  refreshHandoffCmd();
  void ns;
}

function refreshHandoffCmd() {
  const prompt = (handoffPrompt.value || "").trim();
  if (!prompt || !selected) {
    handoffCmd.textContent = "—";
    return;
  }
  let cmd = `kprompt ${shellQuote(prompt)}`;
  if (VIEWS[view].namespaced && nsEl.value) cmd += ` -n ${shellQuote(nsEl.value)}`;
  if (kubeContext) cmd += ` --context ${shellQuote(kubeContext)}`;
  handoffCmd.textContent = cmd;
}

function shellQuote(s) {
  if (/^[A-Za-z0-9_./:@+=,-]+$/.test(s)) return s;
  return `'${String(s).replaceAll("'", `'\"'\"'`)}'`;
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
