# Releasing

`kprompt-dash` ships binaries via [GoReleaser](https://goreleaser.com) and GitHub Releases (same pattern as the main CLI).

## Tag a release

```bash
# on main, clean working tree
git pull origin main
VERSION=v0.1.0
git tag -a "$VERSION" -m "$VERSION"
git push origin "$VERSION"
```

The [release workflow](../.github/workflows/release.yml) builds:

| OS | Arch |
|----|------|
| darwin | amd64, arm64 |
| linux | amd64, arm64 |

Archive name: `kprompt-dash_<version>_<os>_<arch>.tar.gz` (version without leading `v`).

## Install

```bash
go install github.com/kprompt/kprompt-dash/cmd/kprompt-dash@latest
# or download a release asset and put it on PATH, then:
kprompt dash
```

## Local snapshot (no publish)

```bash
goreleaser release --snapshot --clean
```
