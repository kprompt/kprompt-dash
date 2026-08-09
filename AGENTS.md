# AGENTS.md — kprompt-dash

**Local cluster dashboard** for kprompt (OSS, ADR-0011). Go 1.23 (module `github.com/kprompt/kprompt-dash`) serving an embedded web UI for a read-first local view of the cluster. A developer tool you run locally — not a hosted service.

Plan lives in the private `kprompt-architecture` repo ([DASH-TASKS.md], ADR-0011). Do not invent conflicting contracts here.

## Product DNA (do not contradict)

- **Read-first** local dashboard. Any mutation must follow the same plan -> approve DNA as the CLI; no silent cluster writes.
- kubeconfig-based access via `client-go`; binds locally by default (careful with exposing the bind address).
- Honesty: label experimental views; do not fake cluster data.

## Layout

| Path | Role |
|------|------|
| `cmd/kprompt-dash` | Binary entry |
| `internal/api` | HTTP API serving the UI |
| `internal/kube` | client-go cluster reads |
| `internal/bind` | Local bind / addressing |
| `web/` | Embedded static UI (`embed.go` + `static/`) |

## Build / test

```bash
go test ./...
go build ./cmd/kprompt-dash
```

## Working rules

1. Read paths first; if adding a mutation, route it through explicit approval — never silent apply.
2. Default to a local bind; do not widen the exposed address without an explicit, documented flag + warning.
3. `gofmt`; packages under `internal/`; treat all request input as untrusted.
4. Rebuild embedded assets when the UI changes; keep `go test ./...` green.
5. No secrets or kubeconfigs committed.
6. Commit / open PRs only when asked.

## Cursor rules / skills

- Always-on: `.cursor/rules/project.mdc`.
- Skill: `dash-view` — add a read view (kube read -> API -> embedded UI) safely.
