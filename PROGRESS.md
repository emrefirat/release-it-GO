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
| 6 | Advanced Features | Tamamlandi | 100% |
| 7 | Testing, CI/CD, Documentation | Tamamlandi | 100% |
| 8 | Init Command & Dual Config | Tamamlandi | 100% |
| 9 | Conventional Commit Linting | Tamamlandi | 100% |
| 10 | UI/Output Iyilestirmesi | Tamamlandi | 100% |
| 11 | Docker Container Destegi | Tamamlandi | 100% |
| 12 | Docker Pre-flight Kontrolleri | Tamamlandi | 100% |
| 13 | Webhook Notification (Slack + Teams) | Tamamlandi | 100% |

**Son Guncelleme:** 2026-02-17
**Aktif Gelistirici:** Claude
**Mevcut Versiyon:** dev (Phase 13 Webhook Notification destegi tamamlandi - production-ready)

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

**Durum:** Tamamlandi
**PRD:** `docs/phase_6.md`

### Yapilacaklar

- [x] Bumper: dosyadan versiyon okuma (JSON/YAML/TOML/INI/text)
- [x] Bumper: dosyaya versiyon yazma
- [x] Bumper: glob pattern destegi
- [x] CalVer runner entegrasyonu
- [x] Pre-release flows (Phase 1'de semver.go'da implement edilmisti)
- [x] --no-increment modu
- [x] --only-version modu
- [x] --changelog ve --release-version CLI modlari
- [x] CalVer + SemVer conflict detection
- [x] Bumper pipeline step (bump adimi)
- [x] Unit testler

### Notlar

- Test coverage: bumper=%87.8
- Bumper: JSON (nested path), YAML, TOML, INI ([section].key), text (consumeWholeFile) destegi
- Bumper: glob pattern (*.json, charts/*/Chart.yaml), prefix (^, ~), dry-run destegi
- CalVer: runner.determineCalVer() ile pipeline'a entegre edildi
- CalVer + pre-release birlikte kullanilamiyor (CLI'da validation)
- Pipeline'a "bump" adimi eklendi: version -> bump -> changelog
- CLI modlari: RunChangelogOnly, RunReleaseVersionOnly, RunOnlyVersion, RunNoIncrement
- RunOnlyVersion: versiyon secimi sonrasi otomatik CI moduna gecer
- RunNoIncrement: versiyon artirmadan changelog ve release gunceller
- Mevcut YAML/TOML dependency'leri kullanildi (go-yaml, go-toml via Viper)
- INI icin stdlib ile basit parser yazildi (harici dependency yok)

---

## Faz 7: Testing, CI/CD, Documentation

**Durum:** Tamamlandi
**PRD:** `docs/phase_7.md`

### Yapilacaklar

- [x] Integration testler
- [x] API mock testleri
- [x] Coverage %80+ hedefi
- [x] GitHub Actions CI workflow
- [x] GitHub Actions Release workflow
- [x] GoReleaser config
- [x] Shell completions (bash/zsh/fish)
- [x] Build info (ldflags)

### Notlar

- Test coverage: bumper=%87.8, changelog=%93.3, cli=%83.0, config=%87.9, git=%86.8, hook=%100, log=%100, release=%86.7, runner=%80.6, ui=%78.6, version=%86.5
- Tum paketler %78+ coverage'a ulasti, toplam %80+ hedefi karsilandi
- 17 integration test: full pipeline, patch/minor/major bump, dry-run, no-tags, changelog-only, release-version-only, disable commit/tag, conventional commit auto-detect, breaking change auto-major, bumper file update, keep-a-changelog, hook execution/failure, config JSON/YAML, no-increment, sequential releases
- Bubbletea model testleri Init/Update/View ile direkt test edildi (terminal gerektirmeden)
- GitLab upload assets, error handling, missing token testleri eklendi
- GitHub Actions CI: Go 1.22/1.23 matrix, test+lint+build, coverage check
- GitHub Actions Release: v* tag'da GoReleaser v2 ile otomatik release
- GoReleaser: linux/darwin/windows x amd64/arm64, ldflags (cli.Version/Commit/Date), nfpms (deb/rpm/apk)
- Shell completions: cobra ile bash/zsh/fish/powershell, cmd.OutOrStdout() ile test edilebilir
- Race condition testleri tum paketlerde basarili

---

## Post-Release: Gercek Ortam Testleri ve Iyilestirmeler

**Durum:** Tamamlandi

### Yapilacaklar

- [x] Gercek GitLab ortaminda release testi (testproject06)
- [x] Guvenlik fix: HTTPS URL'lerden credential stripping (CHANGELOG'a token sizmasini onleme)
- [x] GoReleaser ldflags fix (main.version/commit/date)
- [x] Eski npm release-it config uyumlulugu (normalizeJSON + applyPluginCompat)
  - [x] requireBranch: [] → string donusumu
  - [x] gitlab.assets: {links:[]} → []string donusumu
  - [x] plugins section'dan changelog ayarlarinin map edilmesi
  - [x] npm, versionFile gibi bilinmeyen alanlarin temizlenmesi
- [x] --preRelease shorthand flag'i eklendi (sub-branch prerelease destegi)
- [x] GitLabConfig'e PreRelease alani eklendi
- [x] Gercek GitLab CI/CD pipeline testi (testproject05)
  - [x] Main branch: otomatik release (v1.4.1)
  - [x] Sub-branch: prerelease (v1.5.0-deneme2026.0)
  - [x] SSH ile git push, personal token ile GitLab release API
- [x] Compat testleri (6 test: conventional-changelog, keep-a-changelog, no-plugins, old npm format, requireBranch array, YAML ignored)

### Notlar

- KRITIK GUVENLIK FIX: ParseRepoURL'de HTTPS credential stripping eklendi. Eski halde oauth2:token@host formatindaki URL'ler CHANGELOG compare linklerine siziyordu.
- npm release-it config uyumlulugu: Eski .release-it.json dosyalari (npm, plugins, requireBranch:[], assets:{links:[]}) sorunsuz yukleniyor.
- --preRelease="identifier" shorthand'i: preReleaseId set ediyor + GitHub/GitLab release'i otomatik pre-release olarak isaretliyor.
- GitLab CI pipeline SSH ile push yapiyor (HTTPS token git push icin guvenilir degil), API token ile release olusturuyor.
- Gercek ortam testleri: testproject06 (v0.2.0, v0.3.0, v1.0.0) ve testproject05 (v1.4.1, v1.5.0-deneme2026.0) basarili.

---

## Faz 8: Init Command & Dual Config File Support

**Durum:** Tamamlandi
**PRD:** `docs/phase_8.md`

### Yapilacaklar

- [x] configSearchFiles'a `.release-it-go.*` dosyalarini oncelikli ekle
- [x] Prompter interface'ine generic `Select` metodu ekle (bubbletea + non-interactive)
- [x] `config/writer.go` - WriteConfigJSON (smart defaults omission)
- [x] `config/migrate.go` - DetectLegacyConfig + MigrateLegacyConfig
- [x] `cli/init.go` - Init komutu ve wizard akisi
- [x] Init komutunu root.go'ya kaydet
- [x] Unit testler (writer, migrate, init, prompt select, config priority)
- [x] docs/phase_8.md PRD dokumani

### Notlar

- Native config (.release-it-go.*) legacy config'den (.release-it.*) once aranir
- WriteConfigJSON sadece default'tan farkli alanlari yazar (minimal JSON ciktisi)
- Migration akisi: legacy oku → backup al → normalizeJSON + applyPluginCompat → native yaz
- Init wizard: platform, changelog, git ops, commit msg, tag format, branch secimi
- Mevcut runner_test.go mock prompter'lara Select metodu eklendi (interface uyumlulugu)
- Init wizard'da git push kapatildiginda `requireUpstream` otomatik false yapilir (upstream kontrolu push olmadan anlamsiz)
- `requireCleanWorkingDir` push durumundan bagimsiz olarak HER ZAMAN aktif kalir (commit/tag atarken kirli working dir tehlikeli)
- Init komutuna ozel flag yok, root'tan gelen `--ci` flag'i NonInteractivePrompter ile tum sorulari default yanitlar
- `--ci` modunda mevcut `.release-it-go.json` varsa Confirm("Overwrite?", default=false) → abort eder (guvenli davranis)

---

## Faz 9: Conventional Commit Linting

**Durum:** Tamamlandi
**PRD:** `docs/phase_9.md`

### Yapilacaklar

- [x] `git/changelog.go` - CommitInfo struct + GetCommitsWithHashSinceTag()
- [x] `changelog/lint.go` - LintInput, LintResult, LintCommits() fonksiyonu
- [x] `config/config.go` - RequireConventionalCommits alani
- [x] `runner/runner.go` - checkCommitLint() pipeline adimi + RunCheckCommits() modu
- [x] `cli/root.go` - --check-commits + --ignore-commit-lint flag'leri
- [x] `cli/init.go` - Wizard'a "Require conventional commits?" sorusu
- [x] `changelog/lint_test.go` - LintCommits testleri (8 test)
- [x] `git/changelog_test.go` - GetCommitsWithHashSinceTag testleri (3 test)

### Notlar

- Circular dependency onleme: lint fonksiyonu `changelog` paketinde, `git.CommitInfo` yerine kendi `LintInput` struct'ini kullaniyor
- Runner zaten hem `git` hem `changelog` import ettigi icin donusum runner'da yapiliyor
- Merge commit (`Merge `) ve revert commit (`Revert `) otomatik gecis alir
- commitPattern regex (parser.go'daki) dogrudan yeniden kullanildi
- Pipeline sirasi: init → prerequisites → commitlint → version → ...
- `--check-commits`: bagimsiz lint modu, exit code 1 ile hata donusu
- `--ignore-commit-lint`: RequireConventionalCommits override eder
- Tum testler race detection ile basarili

---

## Faz 10: UI/Output Iyilestirmesi

**Durum:** Tamamlandi

### Yapilacaklar

- [x] `ui/colors.go` - 14 Unicode ikon sabiti (IconSuccess, IconFail, IconVersion, IconTag, IconPush, IconRelease, IconChangelog, IconCommit, IconLint, IconSkip, IconLink, IconRocket, IconDryRun, IconWarning)
- [x] `ui/colors.go` - FormatBold() fonksiyonu
- [x] `log/logger.go` - Print() metodu (slog formatsiz dogrudan cikti)
- [x] `log/logger.go` - Verbose() format degisikligi (slog → `↳` indented dim format)
- [x] `ui/spinner.go` - CI Start() ciktisi (`-` → `⠋` spinner frame)
- [x] `ui/spinner.go` - CI Stop() ciktisi (`OK`/`FAIL` → renkli `✓`/`✗` ikonlari)
- [x] `runner/runner.go` - printBanner() (🚀 release-it-go / 🧪 dry-run banner)
- [x] `runner/runner.go` - Versiyon mesaji (Info → Print + 📦 ikon)
- [x] `runner/runner.go` - Skip mesajlari (Info → Print + ⏭️ ikon)
- [x] `runner/runner.go` - printSummary() lipgloss border box ile yeniden tasarim
- [x] `log/logger_test.go` - Print, Verbose format testleri (4 yeni test)
- [x] Tum testler gecti (`go test ./... -race`)
- [x] `go vet` ve `go fmt` temiz

### Notlar

- Ikonlar Unicode karakter, ANSI kodu degil. `NO_COLOR` sadece lipgloss renklerini kapatir, ikonlar her durumda gorunur
- Logger.Print(): slog formatlamasi olmadan dogrudan stderr'e yazar, kullanici dostu mesajlar icin
- Logger.Verbose(): `    ↳ mesaj` formatinda, indented ve dim renkte (verbose >= 1)
- Logger.Debug(): Mevcut slog formati korundu (degisiklik yok)
- CI spinner: `⠋` frame ile baslama, renkli `✓`/`✗` ile bitme
- printSummary: lipgloss RoundedBorder ile kutu, ikonlu satirlar, duration bilgisi
- printBanner: Run(), RunOnlyVersion(), RunNoIncrement() basina eklendi
- 5 dosya etkilendi: ui/colors.go, log/logger.go, ui/spinner.go, runner/runner.go, log/logger_test.go

---

## Faz 11: Docker Container Destegi

**Durum:** Tamamlandi
**PRD:** `docs/phase_11.md`

### Yapilacaklar

- [x] `docs/phase_11.md` PRD dokumani
- [x] `.dockerignore` build context filtresi
- [x] `Dockerfile` multi-stage build (golang:1.24.3-alpine → alpine:3.21)
- [x] `Makefile` docker-build ve docker-run target'lari
- [x] `PROGRESS.md` Phase 11 guncelleme

### Notlar

- Multi-stage build: builder (golang:1.24.3-alpine) + runtime (alpine:3.21)
- Static binary: CGO_ENABLED=0 GOOS=linux, -trimpath -ldflags="-s -w"
- Runtime paketler: git, openssh-client, ca-certificates
- Non-root user: releaser (UID/GID 1000, build arg ile degistirilebilir)
- git safe.directory '*' ile mount edilen repo'lar icin guvenli erisim
- Build ARG'lar: VERSION, COMMIT, BUILD_DATE, USER_UID, USER_GID
- Tahmini image boyutu: ~30MB
- OCI metadata labels eklendi

---

## Faz 12: Docker Ortami Pre-flight Kontrolleri

**Durum:** Tamamlandi

### Yapilacaklar

- [x] `git/prerequisites.go` - checkGitIdentity() fonksiyonu (user.name/user.email kontrolu)
- [x] `git/prerequisites.go` - CheckPrerequisites() icine identity check eklenmesi
- [x] `git/prerequisites_test.go` - Identity testleri (5 test: commit kapali, ikisi tam, name eksik, email eksik, ikisi eksik)
- [x] `runner/runner.go` - checkTokens() fonksiyonu (GitHub/GitLab token kontrolu)
- [x] `runner/runner.go` - checkPrerequisites() icine token check eklenmesi
- [x] `runner/runner_test.go` - Token testleri (11 test: release kapali, token eksik/set, custom tokenRef, skipChecks, her iki platform)
- [x] Tum testler gecti (`go test ./... -race`)
- [x] `go vet` ve `go build` temiz

### Notlar

- Git identity kontrolu sadece `git.commit: true` ise yapilir (tag-only veya push-only senaryolarda gereksiz)
- Token kontrolu `runner` seviyesinde cunku config bilgisine (GitHub/GitLab ayarlari) erisim gerekiyor
- `skipChecks: true` ile token kontrolu atlanabilir (CI ortaminda farkli auth mekanizmasi kullanildiginda)
- Custom `tokenRef` destegi: kullanici farkli env variable adi kullanabilir
- Hatalar erken (prerequisites asamasinda) verilir, pipeline gec asamada basarisiz olmaz

---

## Faz 13: Webhook Notification Destegi (Slack + Teams)

**Durum:** Tamamlandi
**PRD:** `docs/phase_13.md`

### Yapilacaklar

- [x] `internal/config/config.go` - NotificationConfig + WebhookConfig struct'lari
- [x] `internal/config/defaults.go` - Default notification config (disabled, bos webhooks)
- [x] `internal/notification/notification.go` - Client, SendAll, HTTP POST, resolveURL, renderMessage
- [x] `internal/notification/slack.go` - Slack payload builder ({"text": "..."})
- [x] `internal/notification/teams.go` - Teams MessageCard payload builder
- [x] `internal/notification/notification_test.go` - 13 test, %98+ coverage (httptest mock server)
- [x] `internal/runner/runner.go` - sendNotification() pipeline adimi (tum pipeline'lara eklendi)
- [x] `internal/runner/runner_test.go` - 3 test: disabled, empty webhooks, non-fatal error
- [x] `docs/phase_13.md` - PRD dokumani
- [x] Tum testler gecti (`go test ./... -race`)
- [x] `go vet` ve `go build` temiz

### Notlar

- Notification non-fatal: basarisiz olursa uyari loglanir, pipeline durmaz
- Webhook URL guvenlik icin config'e dogrudan yazilmaz, `urlRef` ile env variable adi belirtilir
- Slack ve Teams icin platform-spesifik default sablonlar mevcut
- Kullanici `messageTemplate` ile ozel sablon tanimlayabilir
- Timeout configurable (default: 30 saniye)
- Dry-run modunda HTTP cagrisi yapilmaz

---

## Bugs

- [x] BUG: Ilk release'de changelog "exit status 128" hatasi (2026-02-16) → `LatestVersion=0.0.0` iken `v0.0.0` tag'i araniyordu ama repo'da boyle bir tag yok. `latestVersionToTag()` helper fonksiyonu eklendi: `0.0.0` veya bos string icin bos doner, bu sayede `GetCommitsSinceTag("")` tum commitleri alir. 3 yer etkilendi: `RunChangelogOnly`, `generateChangelog`, `autoDetectIncrement`.
- [x] BUG: Init wizard commit/tag/push'u tek soru olarak soruyordu (2026-02-16) → Kullanici commit+tag isteyip push istemeyince ikilem yasiyordu. Sorular ayrildi: "Enable git commit and tag?" + "Enable git push?" olarak iki ayri prompt yapildi. Push kapaliyken `requireUpstream` otomatik false.
- [x] BUG: CHANGELOG.md olusturulduktan sonra commit'e dahil edilmiyordu (2026-02-16) → `Stage()` default'ta `git add . --update` ile sadece tracked dosyalari ekliyor. Yeni olusturulan CHANGELOG.md untracked oldugu icin atlaniyordu. Fix: `StageFile()` metodu eklendi, `generateChangelog()` sonunda CHANGELOG.md explicit olarak `git add` ile stage'leniyor.
- [x] BUG: Commit yokken bile release cikiyordu, bos CHANGELOG entry'leri olusuyordu (2026-02-16) → `git.requireCommits` default'u `false` idi, commit olmadan ard arda release atilabiliyordu. Fix: default `true` yapildi. Init wizard'a "Require new commits before release?" sorusu eklendi. Artik son tag'den beri commit yoksa `no commits since latest tag` hatasi verir.
- [x] BUG: "no commits since latest tag" error yerine graceful exit olmali (2026-02-16) → Kullanici commit yoksa hata yerine bilgi mesaji gosterip temiz cikis yapmali. Fix: prerequisites'te error yerine logger.Print + return nil.
- [x] BUG: CI spinner cift satir gosteriyor (Start + Stop) (2026-02-16) → `⠋ Initializing...` + `✓ Initializing` tekrarli. Fix: CI Start() artik bir sey yazmiyor, sadece Stop() sonuc satirini yazar.
- [x] BUG: Init adimi ✗ gosteriyor (basarili olmasina ragmen) (2026-02-16) → GetRepoInfo opsiyonel hata spinner'i erken Stop(false) ile kapatiyor. Fix: Opsiyonel hatalarda spinner kapatilmiyor.
- [x] BUG: printSummary lipgloss kutu gereksiz ve tekrarli (2026-02-16) → Kullanici geri bildirimi: cerceve gereksiz detay. Fix: Duz, minimal cikti formatina gecildi.
- [x] BUG: --preRelease ayni ID ile tekrar calistirildiginda versiyon artmiyor (2026-02-16) → `1.6.0-deneme2.0 → 1.6.0-deneme2.0` ayni versiyon uretiliyor, tag zaten var hatasi. Sebep: `prepatch` increment mevcut pre-release'i dusuruyordu sonra ayni .0 ile basliyordu. Fix: Mevcut versiyon ayni pre-release ID'ye sahipse `"prerelease"` increment kullan (sayi arttirir: `.0 → .1`).
- [x] BUG: --check-commits gecersiz commit type'lari kabul ediyor (2026-02-16) → `fic: deneme commit` gibi gecersiz type'lar conventional commit olarak geciyordu. Sebep: regex `\w+` herhangi bir kelimeyi type olarak kabul ediyordu. Fix: `allowedTypes` map eklendi (Angular preset: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert), type dogrulamasi yapiliyor. Gecersiz type icin "unknown type: fic" sebebi doner. --verbose ile kontrol edilen commitlerin listesi gosteriliyor.

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
| 2026-02-16 | Claude | Phase 6 tamamlandi: bumper (JSON/YAML/TOML/INI/text), CalVer entegrasyonu, CLI modlari, pre-release flows, pipeline bump adimi |
| 2026-02-16 | Claude | Phase 7 tamamlandi: integration tests (17), coverage %80+, CI/CD workflows, GoReleaser, shell completions, build info |
| 2026-02-16 | Claude | Security fix: HTTPS URL credential stripping, GoReleaser ldflags fix |
| 2026-02-16 | Claude | Config compat: npm release-it format uyumlulugu (normalizeJSON, applyPluginCompat) |
| 2026-02-16 | Claude | feat: --preRelease shorthand flag, GitLab PreRelease alani |
| 2026-02-16 | Claude | Gercek ortam testleri: GitLab CI pipeline (main + sub-branch prerelease) basarili |
| 2026-02-16 | Claude | Phase 8 tamamlandi: init command, dual config support, legacy migration, smart config writer |
| 2026-02-16 | Claude | fix: ilk release changelog hatasi (0.0.0 tag bulunamama), init wizard commit/tag/push ayirma |
| 2026-02-16 | Claude | Phase 9 tamamlandi: conventional commit linting, --check-commits, --ignore-commit-lint, pipeline entegrasyonu |
| 2026-02-16 | Claude | Phase 10 tamamlandi: UI/Output iyilestirmesi - ikon sabitleri, FormatBold, Logger.Print(), Verbose dim format, CI spinner ikonlari, banner, printSummary lipgloss box |
| 2026-02-16 | Claude | fix: UI/Output iyilestirmesi v2 - CI spinner cift satir kaldirma, init ✗ bug fix, "no commits" graceful exit, printSummary kutu kaldirma, spinner mesajlari past-tense |
| 2026-02-16 | Claude | fix: pre-release ayni ID ile versiyon artmama hatasi (prepatch → prerelease increment) |
| 2026-02-16 | Claude | fix: commit lint type validation - allowedTypes map ile gecersiz type'lar reddediliyor, --verbose ile commit listesi gosteriliyor |
| 2026-02-16 | Claude | Phase 11 tamamlandi: Docker container destegi - multi-stage Dockerfile, .dockerignore, Makefile docker target'lari |
| 2026-02-16 | Claude | Phase 12 tamamlandi: Docker pre-flight kontrolleri - git identity check, token pre-flight check (GitHub/GitLab) |
| 2026-02-17 | Claude | Phase 13 tamamlandi: Webhook notification destegi - Slack + Teams, non-fatal pipeline adimi, urlRef guvenlik pattern'i, %98+ coverage |
| 2026-02-17 | Claude | feat: init wizard'a "Write CHANGELOG.md file?" sorusu eklendi - changelog etkinken dosya yazmayi kapama imkani |

---

## Kurallar

1. **Her oturum sonunda bu dosyayi guncelle.**
2. Tamamlanan maddeler `[x]` ile isaretlenir.
3. Yeni eklenen maddeler `[ ]` ile eklenir.
4. Durum alani guncellenir: `Baslanmadi` / `Devam Ediyor` / `Tamamlandi`
5. Ilerleme yuzdesi guncellenir.
6. Notlar bolumune onemli kararlar, engeller veya degisiklikler yazilir.
7. Degisiklik gecmisi tablosuna yeni satirlar eklenir.
