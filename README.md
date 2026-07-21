# kprompt-dash

**Simple, read-only, localhost cluster inventory** for [kprompt](https://github.com/kprompt/kprompt) — not a Lens/Headlamp replacement.

See the cluster → copy a `kprompt "…"` command → plan → approve → apply in the CLI.

## Quickstart

```bash
# install
go install github.com/kprompt/kprompt-dash/cmd/kprompt-dash@latest

# run (uses your kubeconfig)
kprompt-dash -open
# or via the main CLI:
kprompt dash
```

Opens **http://127.0.0.1:7474**. Bind is **localhost only** by default.

### Flags

| Flag | Default | Notes |
|------|---------|--------|
| `-addr` | `127.0.0.1:7474` | Listen address |
| `-allow-remote` | off | Required for non-loopback binds (no auth — dangerous) |
| `-context` | current | kubeconfig context |
| `-open` | off | Open the UI in a browser |

## What it is

- Sidebar: Cluster · Nodes · Deployments · ReplicaSets · Pods
- Detail: events + short log tail (workloads)
- Prompt handoff: copy a `kprompt "…"` command with `-n` / `--context`
- List APIs cap at 200 items by default (`?limit=` up to 500); responses include `truncated` when capped

## What it is not

- Hosted multi-tenant dashboard (`app.kprompt.ai` stays governance/Insights)
- In-browser mutate / apply
- Full CRD / RBAC / Helm explorer
- A Lens or Headlamp replacement

## Security

- **No HTTP authentication.** Anyone who can reach the port can use your kubeconfig privileges (read-only MVP, but still sensitive: logs, labels, events).
- Default bind is **loopback**. Non-loopback requires an explicit `-allow-remote`.
- Bind-all forms like `:7474` are refused.
- Prefer `kprompt` plan → approve for mutations; dash does not apply.

## Install from releases

GitHub Releases ship platform binaries (same GoReleaser story as the CLI). Tag `v*` on this repo to publish.

Architecture: [ADR-0011](https://github.com/kprompt/kprompt-architecture/blob/main/decisions/ADR-0011-local-cluster-dash.md) · tasks [DASH-TASKS.md](https://github.com/kprompt/kprompt-architecture/blob/main/DASH-TASKS.md)

## License

Apache-2.0
