# release-it-go

> [release-it](https://github.com/release-it/release-it) npm aracinin Go ile yeniden yazimi. npm/Node.js bagimliligi olmadan Git tag + GitHub/GitLab release + changelog olusturma.

## Mimari

| Karar | Secim |
|-------|-------|
| Dil / Minimum | Go 1.24+ |
| CLI | Cobra (`internal/cli`) |
| Config | JSON + YAML + TOML (Viper) |
| Versiyon | Git tag (birincil), VERSION dosyasi (ikincil) |
| Changelog | Conventional commits (Angular preset) + keep-a-changelog |
| Release | GitHub + GitLab API |
| UI | Bubbletea + Lipgloss (interaktif), CI auto-detect |
| Dagitim | GitHub Releases, `go install`, Docker |
| Plugin | Yok - hepsi built-in |

## Proje Yapisi

```
cmd/release-it-go/main.go      # Giris noktasi (ldflags: version/commit/date)
internal/
  cli/          # Cobra CLI komutlari (root, init, version, completion)
  config/       # Config yukleme, merge, write, migration, compat
  version/      # Semver parse/increment, CalVer, branch-aware pre-release
  git/          # Git CLI wrapper (commit, tag, push, prerequisites, repo info)
  changelog/    # Conventional commit parser, analyzer, changelog renderer
  release/      # GitHub + GitLab API client, asset upload
  bumper/       # Coklu dosya versiyon guncelleme (JSON/YAML/TOML/INI/text)
  hook/         # Lifecycle hook runner (before:/after: events)
  notification/ # Webhook bildirim (Slack + Teams)
  ui/           # Terminal UI (prompt, spinner, colors, CI mode)
  runner/       # Pipeline orchestrator (tum adimlari koordine eder)
  log/          # Structured logger (slog, verbose levels)
test/integration/               # Gercek git repo ile entegrasyon testleri
docs/phase_*.md                 # Faz PRD dokumanlari (1-15)
```

## Pipeline Sirasi

```
init → prerequisites → commitlint → version → bump → changelog → git:release → github:release → gitlab:release → notification
```

Her adim `before:` / `after:` hook destekler. Dry-run tum adimlar icin desteklenir.

## Hizli Komutlar

```bash
make check          # fmt + vet + lint + vuln + test + build (commit oncesi)
make build          # Binary derle (ldflags ile)
make test           # Tum testler (-v -cover -race)
make test-unit      # Sadece unit testler
make test-integration  # Sadece integration testler
make coverage       # HTML coverage raporu
make vuln           # govulncheck guvenlik taramasi
make lint           # golangci-lint
make tidy           # go mod tidy + verify
make docker-build   # Docker image (VERSION=x.y.z ile)
```

## Test Mocking Patterni

Git islemleri `commandExecutor` fonksiyon degiskeni uzerinden calisir:
```go
// Testlerde mock:
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    return "mock output", nil
}
```

## Onemli Dosyalar

| Dosya | Aciklama |
|-------|----------|
| `PROGRESS.md` | Faz takibi, bug log, degisiklik gecmisi - **her oturum sonunda guncelle** |
| `Makefile` | Build, test, lint, docker - ldflags ile versiyon enjeksiyonu |
| `.goreleaser.yaml` | Multi-platform release (linux/darwin/windows x amd64/arm64) |
| `Dockerfile` | Multi-stage build, non-root user, static binary |
| `.github/workflows/` | CI (matrix test), build (GoReleaser), release (self-hosting) |

## Zorunlu Uygulamalar

1. **Context7 kullan**: Kod yazarken veya kutuphane kullanirken `resolve-library-id` + `query-docs` ile guncel dokumantasyonu kontrol et. Tahmine dayali kod yazma.
2. **PROGRESS.md guncelle**: Her gelistirme oturumu sonunda yapilanlari, tamamlanan maddeleri ve bug'lari `PROGRESS.md`'ye yaz. Yeni faz eklenirse `docs/phase_N.md` de olustur.
3. **Guvenlik taramasi yap**: Commit oncesi `make check` (govulncheck dahil) calistir. Guvenlik acigi suphesi varsa `gosec ./...` ile statik analiz yap.
4. **Commit oncesi checklist**: `make check` = fmt + vet + lint + vuln + test + build. Basarisiz olursa commit atma.

## Kurallar

Detayli kurallar `.claude/rules/` altinda modular dosyalarda:
- `git-workflow.md` — Commit, branch, conventional commits
- `testing.md` — Test zorunluluklari ve pattern'ler
- `code-quality.md` — Go best practices, naming, error handling
- `security.md` — Guvenlik kurallari, gosec, govulncheck
- `progress-tracking.md` — PROGRESS.md ve bug takibi
- `best-practices.md` — Go best practices, dependency, concurrency, pipeline, Docker

**Bu kurallar zorunludur.**
