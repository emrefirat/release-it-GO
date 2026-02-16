# PROGRESS.md - release-it-go Proje Ilerleme Takibi

> Bu dosya, projenin genel ilerlemesini ve her fazin durumunu takip eder.
> Her gelistirme oturumu sonunda guncellenmelidir.

---

## Genel Durum

| Faz | Baslik | Durum | Ilerleme |
|-----|--------|-------|----------|
| 1 | Core Foundation | Tamamlandi | 100% |
| 2 | Git Operations | Tamamlandi | 100% |
| 3 | Conventional Commits + Changelog | Tamamlandi | 100% |
| 4 | GitHub + GitLab Releases | Tamamlandi | 100% |
| 5 | Interactive UI + Hooks + Pipeline | Tamamlandi | 100% |
| 6 | Advanced Features | Baslanmadi | 0% |
| 7 | Testing, CI/CD, Documentation | Baslanmadi | 0% |

**Son Guncelleme:** 2026-02-16
**Aktif Gelistirici:** Claude
**Mevcut Versiyon:** dev (Phase 5 tamamlandi)

---

## Faz 1: Core Foundation

**Durum:** Tamamlandi
**PRD:** `docs/phase_1.md`

### Yapilacaklar

- [x] Go module init (`go mod init`)
- [x] Cobra CLI iskeleti
- [x] Config struct tanimlari
- [x] Config loader (JSON/YAML/TOML)
- [x] Default degerler
- [x] CLI flags -> config merge
- [x] Git tag'den versiyon okuma
- [x] VERSION dosyasindan versiyon okuma
- [x] Semver parse/increment/compare
- [x] Template variable rendering
- [x] Logger (verbose seviyeleri)
- [x] Makefile
- [x] Unit testler
- [x] CalVer struct ve temel implementasyon

### Notlar

- Test coverage: cli=%82.9, config=%87.9, log=%100, version=%86.5
- semver.IncPatch() pre-release'de pre-release'i kaldirir (1.2.3-beta.0 -> 1.2.3), bu dogru semver davranisi
- Viper ile config unmarshaling icin mapstructure tag'leri eklendi
- runGit fonksiyonu test icin mocklanabilir (var olarak tanimli)

---

## Faz 2: Git Operations

**Durum:** Tamamlandi
**PRD:** `docs/phase_2.md`

### Yapilacaklar

- [x] Git runner (komut calistirma)
- [x] Prerequisite checks (branch, clean, upstream, commits)
- [x] Stage + Commit
- [x] Tag olusturma
- [x] Push
- [x] Repo info parse (HTTPS + SSH)
- [x] Basit git log changelog
- [x] Dry-run destegi
- [x] Unit testler

### Notlar

- Test coverage: git=%88.7
- commandExecutor fonksiyon degiskeni ile git komutlari test icin mocklanabilir
- isWriteOperation ile dry-run modunda okuma islemleri calismaya devam eder
- TagExists her zaman gercek git komutu calistirir (dry-run dahil)
- HTTPS ve SSH remote URL formatlari regex ile parse ediliyor

---

## Faz 3: Conventional Commits + Changelog

**Durum:** Tamamlandi
**PRD:** `docs/phase_3.md`

### Yapilacaklar

- [x] Conventional commit parser
- [x] Bump analyzer (major/minor/patch)
- [x] Conventional-changelog formati
- [x] Keep-a-changelog formati
- [x] CHANGELOG.md dosya guncelleme
- [x] Unit testler

### Notlar

- Test coverage: changelog=%93.3
- Regex ile conventional commit parse (type, scope, !, description, body, footers)
- Breaking change algilama: footer (BREAKING CHANGE:) ve bang (feat!) destegi
- Conventional-changelog: Features, Bug Fixes, Performance Improvements, Reverts, BREAKING CHANGES bolumleri
- Keep-a-changelog: Added, Changed, Fixed, Removed bolumleri
- insertAfterHeader ile mevcut CHANGELOG.md icerigini koruyarak prepend

---

## Faz 4: GitHub + GitLab Releases

**Durum:** Tamamlandi
**PRD:** `docs/phase_4.md`

### Yapilacaklar

- [x] Release provider interface
- [x] GitHub client (create, upload, comment)
- [x] GitLab client (create, upload, comment)
- [x] Token yonetimi
- [x] Asset upload (glob)
- [x] GitHub Enterprise destegi
- [x] GitLab CA certificate destegi
- [x] Dry-run destegi
- [x] API mock testleri

### Notlar

- Test coverage: release=%73.7
- Harici SDK kullanilmadi, net/http ile dogrudan REST API cagrisi
- GitHub: CreateRelease, UploadAssets, PostComment, ValidateToken, GHE URL, proxy, makeLatest, autoGenerate, discussionCategory
- GitLab: CreateRelease, UploadAssets (Generic Package + Release Link), PostComment (MR/issue), ValidateToken, CA cert, custom token header
- httptest.NewServer ile mock API testleri
- Asset content type detection: 12+ format (zip, tar.gz, dmg, deb, rpm, exe, sig, vb.)

---

## Faz 5: Interactive UI + Hooks + Pipeline

**Durum:** Tamamlandi
**PRD:** `docs/phase_5.md`

### Yapilacaklar

- [x] Versiyon secim prompt
- [x] Onay prompt'lari
- [x] Spinner animasyonu
- [x] Renkli cikti
- [x] CI ortam algilama
- [x] Hook runner (before/after lifecycle)
- [x] Ana pipeline orchestrator
- [x] Ozel modlar (--changelog, --release-version, --only-version)
- [x] Ozet ciktisi
- [x] Unit testler

### Notlar

- Test coverage: ui=%42.9 (bubbletea interactive models terminal gerektiriyor), hook=%100, runner=%25.3 (pipeline adimlarinda git mock gerekli)
- Bubbletea v1.3.10 ile interaktif terminal UI (selectModel, confirmModel, inputModel)
- Lipgloss v1.1.0 ile renklendirme, NO_COLOR environment variable destegi
- CI algilama: GITHUB_ACTIONS, GITLAB_CI, CIRCLECI, TRAVIS, JENKINS_URL, BITBUCKET_PIPELINE, CODEBUILD_BUILD_ID, TF_BUILD
- NonInteractivePrompter: CI modunda tum prompt'lari otomatik yanitlar
- HookRunner: 12 lifecycle event (before/after: init, bump, release, git:release, github:release, gitlab:release)
- Template variable rendering: ${version}, ${tagName}, ${changelog}, ${releaseUrl}, ${branchName}, ${repo.*}
- Pipeline: init -> prerequisites -> version -> changelog -> git:release -> github:release -> gitlab:release
- Her adimda before/after hook calistirma ve UpdateVars ile degisken guncelleme
- Dry-run tum adimlarda destekleniyor

---

## Faz 6: Advanced Features

**Durum:** Baslanmadi
**PRD:** `docs/phase_6.md`

### Yapilacaklar

- [ ] Bumper: dosyadan versiyon okuma (JSON/YAML/TOML/INI/text)
- [ ] Bumper: dosyaya versiyon yazma
- [ ] Bumper: glob pattern destegi
- [ ] CalVer implementasyonu
- [ ] Pre-release flows
- [ ] --no-increment modu
- [ ] Verbose/debug modlari
- [ ] Unit testler

### Notlar

-

---

## Faz 7: Testing, CI/CD, Documentation

**Durum:** Baslanmadi
**PRD:** `docs/phase_7.md`

### Yapilacaklar

- [ ] Integration testler
- [ ] API mock testleri
- [ ] Coverage %80+ hedefi
- [ ] GitHub Actions CI workflow
- [ ] GitHub Actions Release workflow
- [ ] GoReleaser config
- [ ] Shell completions (bash/zsh/fish)
- [ ] Build info (ldflags)

### Notlar

-

---

## Degisiklik Gecmisi

| Tarih | Gelistirici | Degisiklik |
|-------|------------|------------|
| 2026-02-16 | - | Proje baslatildi, PRD dosyalari olusturuldu |
| 2026-02-16 | Claude | Phase 1 tamamlandi: CLI, config, version, logger, template, tests |
| 2026-02-16 | Claude | Phase 2 tamamlandi: git runner, prerequisites, commit, tag, push, repo info, changelog, tests |
| 2026-02-16 | Claude | Phase 3 tamamlandi: conventional commit parser, bump analyzer, changelog renderers (conventional + keep-a-changelog), file update |
| 2026-02-16 | Claude | Phase 4 tamamlandi: GitHub + GitLab API client, release create, asset upload, comment, token management, GHE/CA cert support |
| 2026-02-16 | Claude | Phase 5 tamamlandi: bubbletea UI, lipgloss colors, spinner, CI detection, hook runner, pipeline orchestrator, dry-run, tests |

---

## Kurallar

1. **Her oturum sonunda bu dosyayi guncelle.**
2. Tamamlanan maddeler `[x]` ile isaretlenir.
3. Yeni eklenen maddeler `[ ]` ile eklenir.
4. Durum alani guncellenir: `Baslanmadi` / `Devam Ediyor` / `Tamamlandi`
5. Ilerleme yuzdesi guncellenir.
6. Notlar bolumune onemli kararlar, engeller veya degisiklikler yazilir.
7. Degisiklik gecmisi tablosuna yeni satirlar eklenir.
