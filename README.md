# kprompt-dash

**Simple, read-only, localhost cluster inventory** for [kprompt](https://github.com/kprompt/kprompt) — not a Lens/Headlamp replacement.

```bash
go run ./cmd/kprompt-dash
# open http://127.0.0.1:7474
```

Uses your local **kubeconfig** (same as `kubectl` / `kprompt`). Binds to **`127.0.0.1` only** by default.

## What it is

- Namespace + Deployment + Pod tables
- Detail: events + short log tail (coming in D-005)
- Handoff to `kprompt "…"` for plan → approve → apply (D-006)

## What it is not

- Hosted multi-tenant dashboard (`app.kprompt.ai` stays governance/Insights)
- In-browser mutate / apply
- Full CRD explorer

Architecture: [ADR-0011](https://github.com/kprompt/kprompt-architecture/blob/main/decisions/ADR-0011-local-cluster-dash.md) · tasks [DASH-TASKS.md](https://github.com/kprompt/kprompt-architecture/blob/main/DASH-TASKS.md)

## License

Apache-2.0
