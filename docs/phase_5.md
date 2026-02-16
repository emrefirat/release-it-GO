# Phase 5: Interactive UI + Hooks + Pipeline

> **Hedef:** Terminal prompts, spinner, renklendirme, CI detection, hook sistemi ve ana runner pipeline orchestration.

---

## 1. Genel Bakis

Bu faz, kullanici deneyimini (UX) olusturur: interaktif sorular, renkli cikti, spinner animasyonlari ve CI ortam algilama. Ayrica hook sistemi (before/after lifecycle) ve tum adimlari orkestre eden ana pipeline implement edilir.

---

## 2. Dosya Yapisi

```
internal/
  ui/
    prompt.go              # Interaktif sorular (versiyon secimi, onay)
    spinner.go             # Spinner animasyonu
    colors.go              # Renkli cikti yardimcilari
    ci.go                  # CI ortam algilama
  hook/
    hook.go                # Hook calistirma (before/after lifecycle)
    template.go            # Hook template variable'lari
  runner/
    runner.go              # Ana pipeline orchestrator
    pipeline.go            # Pipeline adimlari ve sirasi
    context.go             # Release context (paylasilan state)
```

---

## 3. Bagimliliklar

| Kutuphane | Versiyon | Kullanim |
|-----------|----------|----------|
| `github.com/charmbracelet/bubbletea` | v1.2+ | Terminal UI framework |
| `github.com/charmbracelet/lipgloss` | v1.0+ | Styling/renklendirme |
| `github.com/charmbracelet/bubbles` | v0.20+ | Hazir UI bilesenler (spinner, text input) |
| `os/exec` (stdlib) | - | Hook komut calistirma |

---

## 4. Interactive UI

### 4.1 Versiyon Secimi Prompt

Interaktif modda kullaniciya versiyon secenekleri sunulur:

```
? Select increment (next version):

  patch (1.2.4)
  minor (1.3.0)
  major (2.0.0)
> Recommended: minor (1.3.0) [based on conventional commits]
  Other, specify...
```

### 4.2 Onay Prompt'lari

Her onemli adim icin onay sorulur:

```
? Commit (chore: release v1.3.0)? (Y/n)
? Tag (v1.3.0)? (Y/n)
? Push? (Y/n)
? Create a release on GitHub (Release 1.3.0)? (Y/n)
```

### 4.3 Implementasyon

```go
// internal/ui/prompt.go

type Prompter interface {
    // SelectVersion, kullanicidan versiyon secmesini ister.
    SelectVersion(current string, recommended string, options []VersionOption) (string, error)

    // Confirm, evet/hayir sorusu sorar.
    Confirm(message string, defaultYes bool) (bool, error)

    // Input, serbest metin girdisi ister.
    Input(message string, defaultValue string) (string, error)
}

type VersionOption struct {
    Label       string // "patch (1.2.4)"
    Version     string // "1.2.4"
    Recommended bool
}

// InteractivePrompter, bubbletea ile terminal UI.
type InteractivePrompter struct{}

// NonInteractivePrompter, CI modunda tum sorulara otomatik cevap verir.
type NonInteractivePrompter struct{}
```

### 4.4 CI Modunda Davranis

CI modunda (`--ci` veya CI ortam algilama):
- Tum onay sorularina otomatik "evet"
- Versiyon secimi: conventional commits'e gore otomatik
- Manual versiyon secimi gerekmiyorsa sessiz calisma

---

## 5. Spinner

```go
// internal/ui/spinner.go

type Spinner struct {
    message string
    active  bool
}

// Start, spinner'i baslatir.
func (s *Spinner) Start(message string)

// Stop, spinner'i durdurur ve sonuc gosterir.
func (s *Spinner) Stop(success bool)

// Update, spinner mesajini gunceller.
func (s *Spinner) Update(message string)
```

Spinner ornek ciktisi:
```
- Checking prerequisites...
  OK Checking prerequisites
- Creating tag v1.3.0...
  OK Creating tag v1.3.0
- Pushing to origin...
  OK Pushing to origin
- Creating release on GitHub...
  OK Creating release on GitHub
```

---

## 6. Renkli Cikti

```go
// internal/ui/colors.go

var (
    StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // yesil
    StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // sari
    StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // kirmizi
    StyleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))  // mavi
    StyleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gri
    StyleBold    = lipgloss.NewStyle().Bold(true)
)

// FormatSuccess, basarili mesaj formatlar.
func FormatSuccess(msg string) string

// FormatError, hata mesaji formatlar.
func FormatError(msg string) string

// FormatDryRun, dry-run mesaji formatlar.
func FormatDryRun(msg string) string
```

---

## 7. CI Ortam Algilama

```go
// internal/ui/ci.go

// IsCI, CI ortaminda calisilip calisilmadigini kontrol eder.
// Bilinen CI ortam degiskenleri:
//   CI, CONTINUOUS_INTEGRATION, BUILD_NUMBER,
//   GITHUB_ACTIONS, GITLAB_CI, CIRCLECI, TRAVIS,
//   JENKINS_URL, BITBUCKET_PIPELINE, CODEBUILD_BUILD_ID
func IsCI() bool {
    ciEnvVars := []string{
        "CI",
        "CONTINUOUS_INTEGRATION",
        "BUILD_NUMBER",
        "GITHUB_ACTIONS",
        "GITLAB_CI",
        "CIRCLECI",
        "TRAVIS",
        "JENKINS_URL",
        "BITBUCKET_PIPELINE",
        "CODEBUILD_BUILD_ID",
        "TF_BUILD",
    }
    for _, envVar := range ciEnvVars {
        if os.Getenv(envVar) != "" {
            return true
        }
    }
    return false
}

// DetectCIProvider, CI saglayicisini tespit eder.
func DetectCIProvider() string // "github-actions", "gitlab-ci", "circle-ci", etc.
```

---

## 8. Hook Sistemi

### 8.1 Lifecycle Hooks

Hook'lar su yasamdongusu adimlarinda calistirilir:

```
before:init -> init -> after:init
before:bump -> bump -> after:bump
before:git:release -> git:release -> after:git:release
before:github:release -> github:release -> after:github:release
before:gitlab:release -> gitlab:release -> after:gitlab:release
before:release -> release -> after:release
```

### 8.2 Implementasyon

```go
// internal/hook/hook.go

type HookRunner struct {
    config *config.HooksConfig
    logger *log.Logger
    dryRun bool
    vars   map[string]string // template variables
}

// NewHookRunner, yeni bir hook runner olusturur.
func NewHookRunner(cfg *config.HooksConfig, logger *log.Logger, dryRun bool) *HookRunner

// RunHooks, belirtilen lifecycle adimi icin tanimli hook'lari calistirir.
// Hook'lar sirasiyla calistirilir. Biri basarisiz olursa durur.
func (h *HookRunner) RunHooks(lifecycle string) error

// SetVars, template variable'larini gunceller.
func (h *HookRunner) SetVars(vars map[string]string)
```

### 8.3 Hook Calistirma

```go
// runCommand, tek bir shell komutunu calistirir.
func (h *HookRunner) runCommand(cmd string) error {
    if h.dryRun {
        h.logger.DryRun("hook: %s", cmd)
        return nil
    }

    // Template variable'larini replace et
    rendered := renderTemplate(cmd, h.vars)

    h.logger.Verbose("hook: %s", rendered)

    // Shell uzerinden calistir
    c := execCommand("sh", "-c", rendered)
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    return c.Run()
}
```

### 8.4 Hook'larda Kullanilabilir Variable'lar

| Lifecycle | Kullanilabilir Variable'lar |
|-----------|---------------------------|
| `before:init` | (sinirli: sadece config degerleri) |
| `after:init` | `latestVersion`, `name`, `repo.*`, `branchName` |
| `before:bump` | + `version` (yeni versiyon) |
| `after:bump` | + `version`, `latestVersion` |
| `before:release` | + `changelog` |
| `after:release` | + `releaseUrl` |

---

## 9. Ana Pipeline (Runner)

### 9.1 Pipeline Adimlari

```
1. Init
   - Config yukle
   - CI ortam kontrol
   - Logger olustur

2. Prerequisites
   - Git kontrolleri (branch, clean, upstream, commits)
   - Token kontrolleri (GitHub/GitLab)

3. Version Bump
   - Mevcut versiyonu al (git tag / VERSION)
   - Commit'leri analiz et (conventional commits)
   - Yeni versiyonu belirle (auto veya interaktif)

4. Changelog
   - Changelog olustur
   - CHANGELOG.md guncelle

5. Bumper
   - Dosyalardaki versiyonlari guncelle

6. Git Release
   - Stage + Commit
   - Tag
   - Push

7. GitHub/GitLab Release
   - Release olustur
   - Asset upload
   - Comments

8. Finalize
   - Ozet goster
   - Cleanup
```

### 9.2 Release Context

```go
// internal/runner/context.go

type ReleaseContext struct {
    Config         *config.Config
    Logger         *log.Logger
    Git            *git.Git
    Prompter       ui.Prompter
    HookRunner     *hook.HookRunner

    // State (pipeline boyunca paylasilan)
    LatestVersion  string
    Version        string
    TagName        string
    Changelog      string
    ReleaseURL     string
    RepoInfo       *git.RepoInfo
    BranchName     string
    IsDryRun       bool
    IsCI           bool

    // Template variables (hook'lar ve config template'leri icin)
    Vars           map[string]string
}

// UpdateVars, template variable'larini mevcut state'e gore gunceller.
func (ctx *ReleaseContext) UpdateVars()
```

### 9.3 Runner Implementasyon

```go
// internal/runner/runner.go

type Runner struct {
    ctx *ReleaseContext
}

// NewRunner, yeni bir runner olusturur.
func NewRunner(cfg *config.Config) (*Runner, error)

// Run, tum release pipeline'ini calistirir.
func (r *Runner) Run() error {
    steps := []struct {
        name string
        fn   func() error
    }{
        {"init", r.init},
        {"prerequisites", r.checkPrerequisites},
        {"version", r.determineVersion},
        {"changelog", r.generateChangelog},
        {"bumper", r.updateFiles},
        {"git", r.gitRelease},
        {"github", r.githubRelease},
        {"gitlab", r.gitlabRelease},
        {"finalize", r.finalize},
    }

    for _, step := range steps {
        r.ctx.HookRunner.RunHooks("before:" + step.name)

        if err := step.fn(); err != nil {
            return fmt.Errorf("%s: %w", step.name, err)
        }

        r.ctx.HookRunner.RunHooks("after:" + step.name)
    }

    return nil
}
```

### 9.4 Ozel Modlar

```go
// --changelog: Sadece changelog olustur ve goster
func (r *Runner) RunChangelogOnly() error

// --release-version: Sadece sonraki versiyonu goster
func (r *Runner) RunReleaseVersionOnly() error

// --only-version: Sadece versiyon sor, kalanini otomatik yap
func (r *Runner) RunOnlyVersion() error

// --no-increment: Mevcut versiyonda release guncelle
func (r *Runner) RunNoIncrement() error
```

---

## 10. Pipeline Hata Yonetimi

```go
// Pipeline adimi basarisiz olursa:
// 1. Hata mesaji goster
// 2. Hook'lar calistirilmis mi kontrol et
// 3. Rollback yap (mumkunse)
// 4. Non-zero exit code ile cik

// Rollback destegi (gelecekte):
// - Tag silme (eger push yapilmamissa)
// - Commit revert (eger push yapilmamissa)
// Not: Push yapildiktan sonra rollback tehlikeli, yapilmaz
```

---

## 11. Ozet Ciktisi

Pipeline tamamlandiginda ozet gosterilir:

```
release-it-go v1.3.0

  Changelog: CHANGELOG.md updated
  Committed: chore: release v1.3.0
  Tagged: v1.3.0
  Pushed: origin (main)
  GitHub Release: https://github.com/owner/repo/releases/tag/v1.3.0

Done! (in 4.2s)
```

Dry-run modunda:
```
release-it-go v1.3.0 (dry-run)

  [dry-run] Changelog: CHANGELOG.md would be updated
  [dry-run] Would commit: chore: release v1.3.0
  [dry-run] Would tag: v1.3.0
  [dry-run] Would push to: origin (main)
  [dry-run] Would create GitHub release: Release 1.3.0

Done! (dry-run, no changes made)
```

---

## 12. Kabul Kriterleri

### UI
- [ ] Versiyon secim prompt'u calisiyor (interaktif mod)
- [ ] Onay prompt'lari calisiyor (Y/n)
- [ ] Serbest metin girdisi calisiyor
- [ ] CI modunda prompt'lar otomatik cevaplanir
- [ ] Spinner animasyonu calisiyor
- [ ] Renkli cikti calisiyor (success/error/warning/info/dim)
- [ ] CI ortam algilama calisiyor (tum bilinen CI ortamlari)
- [ ] NO_COLOR env variable destegi (renkleri kapatma)

### Hooks
- [ ] before/after lifecycle hook'lari calisiyor
- [ ] Hook'larda template variable'lar replace ediliyor
- [ ] Hook basarisiz olursa pipeline duruyor
- [ ] Dry-run modunda hook'lar calistirilmiyor (sadece log)
- [ ] Hook'lar sirayla calistiriliyor

### Pipeline
- [ ] Tum adimlar dogru sirada calistiriliyor
- [ ] Her adim oncesi/sonrasi hook'lar calistiriliyor
- [ ] --changelog modu calisiyor
- [ ] --release-version modu calisiyor
- [ ] --only-version modu calisiyor
- [ ] --no-increment modu calisiyor
- [ ] Hata durumunda anlamli mesaj ve non-zero exit
- [ ] Ozet ciktisi dogru
- [ ] Dry-run ozet ciktisi dogru
- [ ] Pipeline suresi gosteriliyor
- [ ] `go test ./internal/ui/... ./internal/hook/... ./internal/runner/... -race` basarili
- [ ] Test coverage %70+

---

## 13. Test Senaryolari

### UI Tests
- CI algilama: GITHUB_ACTIONS=true -> isCI=true
- CI algilama: hicbir env yok -> isCI=false
- NonInteractivePrompter: Confirm -> her zaman true
- NonInteractivePrompter: SelectVersion -> recommended secilir

### Hook Tests
- Basit komut calistirma
- Template variable rendering
- Birden fazla hook sirayla calistirma
- Hook hata verdiginde durma
- Dry-run: calistirilmama
- Bos hook listesi: hata yok

### Pipeline Tests
- Tam pipeline (mock git + mock release)
- Pipeline adim sirasi dogrulama
- Hook calistirma sirasi dogrulama
- Hata durumunda durma
- --changelog modu
- --release-version modu
- Dry-run pipeline
