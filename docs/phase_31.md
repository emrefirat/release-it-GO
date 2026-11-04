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

## Orta Seviye Bulgular (aynı gün, ikinci tur)

Kullanıcı onayıyla 13 orta bulgu da düzeltildi; dört paket-grubu commit'i, her biri regresyon testli.

| # | Bulgu | Yer | Durum |
|---|-------|-----|-------|
| 1 | `requireCleanWorkingDir` untracked dosyaları kirli sayıyordu (npm `git diff --quiet HEAD`) | `prerequisites.go` | ✅ `--untracked-files=no`; gerçek git testi |
| 2 | Identity yalnız `git config`'ten; `GIT_COMMITTER_*` env ile CI reddediliyordu | `prerequisites.go` | ✅ `git var GIT_AUTHOR_IDENT`/`GIT_COMMITTER_IDENT` |
| 3 | `git describe` filtre geçmiyor; fallback erişilebilirliğe bakmıyordu | `tag.go` | ✅ `--match`/`--exclude`; explicit `tagMatch` eşleşmezse ilk release |
| 4 | `pushRepo` URL → token log/hata mesajında; npm URL fallback yok | `repo.go`, `git.go` | ✅ URL doğrudan parse; `gitError` ile redaction (errors.Is korunur) |
| 5 | Windows hook çift tırnak | `hook.go` + `cmdline_windows.go` | ✅ `SysProcAttr.CmdLine` `cmd.exe /S /C "..."` (GOOS=windows vet) |
| 6 | Bozuk webhook URL'si `http.NewRequest` `url.Error`'ında sızıyordu | `notification.go` | ✅ TrimSpace + `urlErr.Err` |
| 7 | keep-a-changelog `## [Unreleased]`, CRLF, header satır-başı | `generator.go` | ✅ |
| 8 | GitLab `milestones` `${version}` render edilmiyordu | `runner.go`, `gitlab.go` | ✅ `ReleaseOptions.Milestones` |
| 9 | Upload'lara 30s `Client.Timeout` | `release.go` `newUploadClient` | ✅ ayrı client, `ResponseHeaderTimeout` ≥ 1m; GitLab fallback linki de düzeltildi |
| 10 | CalVer + `preReleaseId`/`increment` config'ten sessiz | `validate.go` | ✅ + `bumper.*.file` zorunlu |
| 11 | Bumper boş `path` → tüm doküman | `bumper/formats.go` | ✅ JSON/YAML/TOML için `version` default'u |
| 12 | `Retry-After` sınırsız; contributors `%cn` | `httputil/retry.go`, `git/changelog.go` | ✅ 60s cap; `%an` |
| 13 | scp-style remote yalnız `git@` | `repo.go` | ✅ herhangi kullanıcı |

**Davranış değişiklikleri (not):** untracked dosyalar artık release'i engellemez (npm paritesi); env ile verilen identity kabul edilir; explicit `tagMatch` ile erişilebilir eşleşme yoksa paket ilk release'inden başlar (başka daldan tag miras almaz).

## Kalan Bulgular

Düşük seviye liste Faz 32 olarak aynı gün kapatıldı → `docs/phase_32.md`. Açık kalan tek karar: CalVer `time.Now()` yerel saat kullanıyor (ürün kararı, değiştirilmedi).
