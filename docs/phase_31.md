# Faz 31: Üçüncü Denetim — Kritik Düzeltmeler

**Durum:** ✅ Tamamlandı — 2026-11-04

## Özet

v0.4.1 üzerinde üçüncü tam denetim (runner elle; git, release/httputil, config/bumper, changelog/version/hook/notification/cli/githook paralel inceleme). Yedi kritik bulgu doğrulandı (geçici testler veya gerçek git ile) ve her biri regresyon testiyle ayrı commit olarak düzeltildi. Tetikleyici: kullanıcı `prepatch --preRelease` ile "tag already exists" hatasını canlı yaşadı (#2).

## Kritik Bulgular

| # | Bulgu | Yer | Etki | Durum |
|---|-------|-----|------|-------|
| 0 | Go 1.26.7 stdlib'inde 10 açık (`make check` vuln adımı kırmızı) | go.mod, Dockerfile | govulncheck kırmızı | ✅ 1.26.10 |
| 1 | `BREAKING CHANGE:` paragrafı son paragraf değilse atlanıyor (GitHub squash-merge `Co-authored-by:` ekler) | `changelog/parser.go` | major yerine minor bump, boş BREAKING CHANGES bölümü | ✅ |
| 2 | `prepatch` mevcut pre-release'de aynı (veya daha düşük) sürümü döndürüyor: `1.2.4-beta.0 → 1.2.4-beta.0` | `version/semver.go` | "tag already exists"; interaktif menü de aynı seçeneği sunuyor | ✅ `1.2.5-beta.0` (npm semver) |
| 2b | Runner farklı ID'de `"pre"+increment` ile taban sürümü kaydırıyordu (eski IncPatch hatası sayesinde gizliydi) | `runner.go` determineSemVer | `1.6.0-beta.3 → 1.7.0-rc.0` | ✅ pre-release latest → her zaman `prerelease` (npm) |
| 3 | Text bumper `ReplaceAll` sınır kontrolü yapmıyor | `bumper/writer.go` | `v11.0.0 → v12.0.0`, `1.0.0.1 → 2.0.0.1` | ✅ whole-version eşleşme |
| 4 | `commandExecutor` `CombinedOutput`: stderr parse edilen çıktıya karışıyor | `git/git.go`, `git/changelog.go` | `GIT_TRACE=1` ile branch/clean-tree/commit sayısı bozuk; tag rename uyarısı "latest tag" oluyor | ✅ ayrı stdout/stderr, `strconv.Atoi` |
| 5 | npm config'leri strict loader'da fatal: `$schema`, `github.update`, `preReleaseBase`, `hooks.after:git:bump`; `@release-it/bumper` sessizce düşüyor; top-level typo'da öneri yok | `config/compat.go`, `loader.go` | Migrasyon kırık | ✅ sessiz kabul / uyarı / bumper eşlemesi / öneri |
| 6 | GHES asset upload `api.github.com→uploads.github.com` replace ile türetiliyor (no-op); API'nin `upload_url`'i kullanılmıyor | `release/github.go` | Enterprise'da her asset 404 (tag push'tan sonra) | ✅ `upload_url` + `/api/uploads` fallback |
| 7 | `resolvePreReleaseBaseTag` raw tag'i `VersionFromTag` olmadan parse ediyor | `runner.go` | `release-${version}` ile beta serisi hiç devam etmiyor | ✅ |

## Kalan Bulgular (Faz 32 adayı)

**Orta:**
- `requireCleanWorkingDir` untracked dosyaları kirli sayıyor (npm `git diff --quiet HEAD`) — `prerequisites.go`
- `checkGitIdentity` yalnız `git config` bakıyor; `GIT_COMMITTER_*` env ile çalışan CI'lar reddediliyor (`git var GIT_COMMITTER_IDENT`)
- `git describe` `tagMatch`/`tagExclude` geçmiyor; fallback `--merged HEAD` yapmıyor (monorepo'da yanlış dal sürümü)
- `pushRepo` URL verilince token'lı URL log/hata mesajına yazılıyor; npm URL fallback'i yok
- Windows hook'larında çift tırnak `cmd.exe /C` kaçışında bozuluyor (`SysProcAttr.CmdLine`)
- Webhook URL'si bozuksa (`\n`) `http.NewRequest` `url.Error`'ı secret'ı sızdırıyor
- keep-a-changelog `## [Unreleased]` işlenmiyor; CRLF dosyada karışık satır sonu; header eşleşmesi satır başına bağlı değil
- GitLab `milestones` `${version}` render edilmiyor; project-uploads fallback linki kırık; upload'lara 30s `Client.Timeout`
- CalVer + `preReleaseId`/`increment` config'ten gelince `Validate()` sessiz
- Bumper `path` boşsa tüm doküman sürüm olarak okunuyor (npm default `version`)
- `Retry-After` sınırsız; contributors `%cn` (committer) → squash-merge'de "GitHub"
- scp-style SSH remote yalnız `git@` kabul ediyor (`org-123@github.com:` reddediliyor)

**Düşük:** INI inline comment/CRLF kaybı, `Revert "..."` commit'leri changelog'dan düşüyor, `hooks install` idempotent değil (tarih) + 0755 yeniden uygulanmıyor, `version` fallback `v0.0.0-<ts>-<hash>` formunu kaçırıyor, `RELEASE_CHANGELOG` env limiti, `version.go` ölü kod, proxy URL şifresi log'da, `PostComment` ölü kod, JSON fallback HTML-escape, `${changelog}` hook template'inde shell injection (env var yolu güvenli).
