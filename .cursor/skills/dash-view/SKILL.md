---
name: dash-view
description: >-
  Add a read-first view to kprompt-dash: a client-go cluster read exposed through the
  internal API and rendered by the embedded web UI, keeping the local-bind + approval
  DNA. Use when adding a resource view or API endpoint to the dashboard.
---

# Dash view workflow

## 1. Wire the read

| Layer | Location |
|-------|----------|
| Cluster read | `internal/kube` (client-go) |
| API endpoint | `internal/api` |
| Bind / addressing | `internal/bind` (keep local default) |
| UI | `web/static` (rebuild embed via `web/embed.go`) |

## 2. Safety

- Read-only by default; if a mutation is unavoidable, gate it behind explicit approval — no silent apply.
- Do not widen the bind address; keep local-only unless an explicit flag + warning is added.
- Validate all request input; never echo kubeconfig contents or secrets.

## 3. Finish

- Rebuild embedded assets when UI changes.
- `go test ./... && go build ./cmd/kprompt-dash`.
- Commit / PR only when the user asks.
