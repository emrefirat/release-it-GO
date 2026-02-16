# Phase 1: Core Foundation

> **Hedef:** Go module, CLI iskeleti, config yükleme, versiyon tespiti ve temel altyapı.

---

## 1. Genel Bakış

Bu faz, `release-it-go` projesinin temel yapı taşlarını oluşturur. CLI uygulaması cobra ile kurulacak, config sistemi viper ile yönetilecek ve versiyon tespiti git tag + VERSION dosyasından yapılacak.

---

## 2. Proje Yapısı

```
release-it-go/
├── cmd/
│   └── release-it-go/
│       └── main.go                 # Uygulama giriş noktası
├── internal/
│   ├── cli/
│   │   └── root.go                 # Cobra root command + subcommands
│   ├── config/
│   │   ├── config.go               # Config struct tanımları
│   │   ├── loader.go               # Config dosyası yükleme (JSON/YAML/TOML)
│   │   ├── defaults.go             # Default değerler
│   │   └── merge.go                # Config merge (file + CLI flags)
│   ├── version/
│   │   ├── version.go              # Versiyon tespiti (git tag + VERSION)
│   │   ├── semver.go               # Semver parse/increment
│   │   └── calver.go               # CalVer parse/increment (struct tanımı)
│   └── log/
│       └── logger.go               # Structured logger (slog wrapper)
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.MD
├── PROGRESS.md
└── docs/
    └── phase_1.md (bu dosya)
```

---

## 3. Bağımlılıklar

| Kütüphane | Versiyon | Kullanım |
|-----------|----------|----------|
| `github.com/spf13/cobra` | v1.8+ | CLI framework |
| `github.com/spf13/viper` | v1.18+ | Config yönetimi |
| `github.com/Masterminds/semver/v3` | v3.2+ | Semver parse/compare |
| `log/slog` (stdlib) | Go 1.22+ | Structured logging |

---

## 4. Config Struct Tanımları

### 4.1 Ana Config

```go
// internal/config/config.go

type Config struct {
    // Git operations
    Git GitConfig `json:"git" yaml:"git" toml:"git"`

    // GitHub release
    GitHub GitHubConfig `json:"github" yaml:"github" toml:"github"`

    // GitLab release
    GitLab GitLabConfig `json:"gitlab" yaml:"gitlab" toml:"gitlab"`

    // Hooks (lifecycle)
    Hooks HooksConfig `json:"hooks" yaml:"hooks" toml:"hooks"`

    // Changelog
    Changelog ChangelogConfig `json:"changelog" yaml:"changelog" toml:"changelog"`

    // Bumper (çoklu dosya versiyon güncelleme)
    Bumper BumperConfig `json:"bumper" yaml:"bumper" toml:"bumper"`

    // CalVer
    CalVer CalVerConfig `json:"calver" yaml:"calver" toml:"calver"`

    // CI mode
    CI bool `json:"ci" yaml:"ci" toml:"ci"`

    // Dry-run mode
    DryRun bool `json:"dry-run" yaml:"dry-run" toml:"dry-run"`

    // Verbose level (0=normal, 1=verbose, 2=debug)
    Verbose int `json:"verbose" yaml:"verbose" toml:"verbose"`

    // Increment type override: major, minor, patch, premajor, preminor, prepatch, prerelease
    Increment string `json:"increment" yaml:"increment" toml:"increment"`

    // Pre-release identifier (e.g., "beta", "alpha")
    PreReleaseID string `json:"preReleaseId" yaml:"preReleaseId" toml:"preReleaseId"`
}
```

### 4.2 Git Config

```go
type GitConfig struct {
    // Commit step
    Commit        bool     `json:"commit" yaml:"commit" toml:"commit"`               // default: true
    CommitMessage string   `json:"commitMessage" yaml:"commitMessage" toml:"commitMessage"` // default: "chore: release v${version}"
    CommitArgs    []string `json:"commitArgs" yaml:"commitArgs" toml:"commitArgs"`

    // Tag step
    Tag           bool     `json:"tag" yaml:"tag" toml:"tag"`                         // default: true
    TagName       string   `json:"tagName" yaml:"tagName" toml:"tagName"`             // default: "${version}"
    TagMatch      string   `json:"tagMatch" yaml:"tagMatch" toml:"tagMatch"`           // glob pattern for latest tag
    TagExclude    string   `json:"tagExclude" yaml:"tagExclude" toml:"tagExclude"`
    TagAnnotation string   `json:"tagAnnotation" yaml:"tagAnnotation" toml:"tagAnnotation"` // default: "Release ${version}"
    TagArgs       []string `json:"tagArgs" yaml:"tagArgs" toml:"tagArgs"`

    // Push step
    Push     bool     `json:"push" yaml:"push" toml:"push"`                           // default: true
    PushArgs []string `json:"pushArgs" yaml:"pushArgs" toml:"pushArgs"`               // default: ["--follow-tags"]
    PushRepo string   `json:"pushRepo" yaml:"pushRepo" toml:"pushRepo"`               // default: "origin"

    // Prerequisite checks
    RequireBranch         string `json:"requireBranch" yaml:"requireBranch" toml:"requireBranch"` // branch name or pattern
    RequireCleanWorkingDir bool  `json:"requireCleanWorkingDir" yaml:"requireCleanWorkingDir" toml:"requireCleanWorkingDir"` // default: true
    RequireUpstream       bool   `json:"requireUpstream" yaml:"requireUpstream" toml:"requireUpstream"` // default: true
    RequireCommits        bool   `json:"requireCommits" yaml:"requireCommits" toml:"requireCommits"`

    // Changelog from git log
    Changelog string `json:"changelog" yaml:"changelog" toml:"changelog"` // git log command template

    // Advanced
    GetLatestTagFromAllRefs bool   `json:"getLatestTagFromAllRefs" yaml:"getLatestTagFromAllRefs" toml:"getLatestTagFromAllRefs"`
    CommitsPath             string `json:"commitsPath" yaml:"commitsPath" toml:"commitsPath"`
    AddUntrackedFiles       bool   `json:"addUntrackedFiles" yaml:"addUntrackedFiles" toml:"addUntrackedFiles"`
}
```

### 4.3 GitHub Config

```go
type GitHubConfig struct {
    Release    bool   `json:"release" yaml:"release" toml:"release"`           // default: false
    ReleaseName string `json:"releaseName" yaml:"releaseName" toml:"releaseName"` // default: "Release ${version}"
    ReleaseNotes string `json:"releaseNotes" yaml:"releaseNotes" toml:"releaseNotes"` // shell command or template
    Draft      bool   `json:"draft" yaml:"draft" toml:"draft"`
    PreRelease bool   `json:"preRelease" yaml:"preRelease" toml:"preRelease"`     // auto-detect from semver
    MakeLatest bool   `json:"makeLatest" yaml:"makeLatest" toml:"makeLatest"`     // default: true
    AutoGenerate bool `json:"autoGenerate" yaml:"autoGenerate" toml:"autoGenerate"`
    Assets     []string `json:"assets" yaml:"assets" toml:"assets"`               // glob patterns
    Host       string `json:"host" yaml:"host" toml:"host"`                       // default: "api.github.com"
    TokenRef   string `json:"tokenRef" yaml:"tokenRef" toml:"tokenRef"`           // default: "GITHUB_TOKEN"
    Timeout    int    `json:"timeout" yaml:"timeout" toml:"timeout"`
    Proxy      string `json:"proxy" yaml:"proxy" toml:"proxy"`
    SkipChecks bool   `json:"skipChecks" yaml:"skipChecks" toml:"skipChecks"`
    Web        bool   `json:"web" yaml:"web" toml:"web"`

    // Comments on issues/PRs
    Comments GitHubCommentsConfig `json:"comments" yaml:"comments" toml:"comments"`

    // Discussion category
    DiscussionCategoryName string `json:"discussionCategoryName" yaml:"discussionCategoryName" toml:"discussionCategoryName"`
}

type GitHubCommentsConfig struct {
    Submit bool   `json:"submit" yaml:"submit" toml:"submit"`
    Issue  string `json:"issue" yaml:"issue" toml:"issue"`     // template
    PR     string `json:"pr" yaml:"pr" toml:"pr"`               // template
}
```

### 4.4 GitLab Config

```go
type GitLabConfig struct {
    Release      bool     `json:"release" yaml:"release" toml:"release"`
    ReleaseName  string   `json:"releaseName" yaml:"releaseName" toml:"releaseName"`
    ReleaseNotes string   `json:"releaseNotes" yaml:"releaseNotes" toml:"releaseNotes"`
    Milestones   []string `json:"milestones" yaml:"milestones" toml:"milestones"`
    Assets       []string `json:"assets" yaml:"assets" toml:"assets"`
    TokenRef     string   `json:"tokenRef" yaml:"tokenRef" toml:"tokenRef"`           // default: "GITLAB_TOKEN"
    TokenHeader  string   `json:"tokenHeader" yaml:"tokenHeader" toml:"tokenHeader"`   // default: "Private-Token"
    Origin       string   `json:"origin" yaml:"origin" toml:"origin"`
    SkipChecks   bool     `json:"skipChecks" yaml:"skipChecks" toml:"skipChecks"`
    CertificateAuthorityFile    string `json:"certificateAuthorityFile" yaml:"certificateAuthorityFile" toml:"certificateAuthorityFile"`
    CertificateAuthorityFileRef string `json:"certificateAuthorityFileRef" yaml:"certificateAuthorityFileRef" toml:"certificateAuthorityFileRef"`
    Secure       bool     `json:"secure" yaml:"secure" toml:"secure"`
    UseGenericPackageRepositoryForAssets bool `json:"useGenericPackageRepositoryForAssets" yaml:"useGenericPackageRepositoryForAssets" toml:"useGenericPackageRepositoryForAssets"`
}
```

### 4.5 Hooks Config

```go
type HooksConfig struct {
    BeforeInit    []string `json:"before:init" yaml:"before:init" toml:"before:init"`
    AfterInit     []string `json:"after:init" yaml:"after:init" toml:"after:init"`
    BeforeBump    []string `json:"before:bump" yaml:"before:bump" toml:"before:bump"`
    AfterBump     []string `json:"after:bump" yaml:"after:bump" toml:"after:bump"`
    BeforeRelease []string `json:"before:release" yaml:"before:release" toml:"before:release"`
    AfterRelease  []string `json:"after:release" yaml:"after:release" toml:"after:release"`

    // Git-specific hooks
    BeforeGitRelease []string `json:"before:git:release" yaml:"before:git:release" toml:"before:git:release"`
    AfterGitRelease  []string `json:"after:git:release" yaml:"after:git:release" toml:"after:git:release"`

    // GitHub-specific hooks
    BeforeGitHubRelease []string `json:"before:github:release" yaml:"before:github:release" toml:"before:github:release"`
    AfterGitHubRelease  []string `json:"after:github:release" yaml:"after:github:release" toml:"after:github:release"`

    // GitLab-specific hooks
    BeforeGitLabRelease []string `json:"before:gitlab:release" yaml:"before:gitlab:release" toml:"before:gitlab:release"`
    AfterGitLabRelease  []string `json:"after:gitlab:release" yaml:"after:gitlab:release" toml:"after:gitlab:release"`
}
```

### 4.6 Changelog Config

```go
type ChangelogConfig struct {
    // Conventional changelog
    Enabled bool   `json:"enabled" yaml:"enabled" toml:"enabled"`         // default: true
    Preset  string `json:"preset" yaml:"preset" toml:"preset"`           // default: "angular"
    Infile  string `json:"infile" yaml:"infile" toml:"infile"`           // default: "CHANGELOG.md"
    Header  string `json:"header" yaml:"header" toml:"header"`           // default: "# Changelog"

    // Keep-a-changelog format
    KeepAChangelog bool `json:"keepAChangelog" yaml:"keepAChangelog" toml:"keepAChangelog"` // default: false
    AddUnreleased  bool `json:"addUnreleased" yaml:"addUnreleased" toml:"addUnreleased"`
    KeepUnreleased bool `json:"keepUnreleased" yaml:"keepUnreleased" toml:"keepUnreleased"`
    AddVersionUrl  bool `json:"addVersionUrl" yaml:"addVersionUrl" toml:"addVersionUrl"`
}
```

### 4.7 Bumper Config

```go
type BumperConfig struct {
    Enabled bool           `json:"enabled" yaml:"enabled" toml:"enabled"`
    In      *BumperFile    `json:"in" yaml:"in" toml:"in"`       // version source file
    Out     []BumperFile   `json:"out" yaml:"out" toml:"out"`     // version target files
}

type BumperFile struct {
    File            string `json:"file" yaml:"file" toml:"file"`
    Path            string `json:"path" yaml:"path" toml:"path"`               // default: "version"
    Type            string `json:"type" yaml:"type" toml:"type"`               // json, yaml, toml, ini, text
    Prefix          string `json:"prefix" yaml:"prefix" toml:"prefix"`
    VersionPrefix   string `json:"versionPrefix" yaml:"versionPrefix" toml:"versionPrefix"`
    ConsumeWholeFile bool  `json:"consumeWholeFile" yaml:"consumeWholeFile" toml:"consumeWholeFile"`
}
```

### 4.8 CalVer Config

```go
type CalVerConfig struct {
    Enabled           bool   `json:"enabled" yaml:"enabled" toml:"enabled"`
    Format            string `json:"format" yaml:"format" toml:"format"`             // default: "yy.mm.minor"
    Increment         string `json:"increment" yaml:"increment" toml:"increment"`     // "calendar" or "calendar.minor"
    FallbackIncrement string `json:"fallbackIncrement" yaml:"fallbackIncrement" toml:"fallbackIncrement"` // default: "minor"
}
```

---

## 5. Default Değerler

```go
// internal/config/defaults.go

func DefaultConfig() *Config {
    return &Config{
        Git: GitConfig{
            Commit:                 true,
            CommitMessage:          "chore: release v${version}",
            Tag:                    true,
            TagName:                "${version}",
            TagAnnotation:          "Release ${version}",
            Push:                   true,
            PushArgs:               []string{"--follow-tags"},
            PushRepo:               "origin",
            RequireCleanWorkingDir: true,
            RequireUpstream:        true,
            Changelog:              "git log --pretty=format:\"* %s (%h)\" ${latestTag}...HEAD",
        },
        GitHub: GitHubConfig{
            Release:     false,
            ReleaseName: "Release ${version}",
            MakeLatest:  true,
            Host:        "api.github.com",
            TokenRef:    "GITHUB_TOKEN",
            Comments: GitHubCommentsConfig{
                Issue: ":rocket: _This issue has been resolved in v${version}._",
                PR:    ":rocket: _This pull request is included in v${version}._",
            },
        },
        GitLab: GitLabConfig{
            Release:     false,
            ReleaseName: "Release ${version}",
            TokenRef:    "GITLAB_TOKEN",
            TokenHeader: "Private-Token",
            CertificateAuthorityFileRef: "CI_SERVER_TLS_CA_FILE",
        },
        Changelog: ChangelogConfig{
            Enabled: true,
            Preset:  "angular",
            Infile:  "CHANGELOG.md",
            Header:  "# Changelog",
        },
        CalVer: CalVerConfig{
            Format:            "yy.mm.minor",
            Increment:         "calendar",
            FallbackIncrement: "minor",
        },
    }
}
```

---

## 6. Config Dosyası Yükleme

### 6.1 Desteklenen Dosya Adları (aranma sırası)

1. `.release-it.json`
2. `.release-it.yaml` / `.release-it.yml`
3. `.release-it.toml`
4. `package.json` (içindeki `release-it` anahtarı - geriye uyumluluk)

### 6.2 Config Loader

```go
// internal/config/loader.go

// LoadConfig, config dosyasını arar ve yükler.
// Öncelik: CLI flags > Config file > Defaults
func LoadConfig(configPath string) (*Config, error) {
    cfg := DefaultConfig()

    if configPath != "" {
        // Belirtilen dosyayı yükle
        return loadFromFile(cfg, configPath)
    }

    // Otomatik arama sırası
    searchFiles := []string{
        ".release-it.json",
        ".release-it.yaml",
        ".release-it.yml",
        ".release-it.toml",
    }

    for _, f := range searchFiles {
        if fileExists(f) {
            return loadFromFile(cfg, f)
        }
    }

    // Config dosyası bulunamazsa default değerlerle devam et
    return cfg, nil
}
```

### 6.3 CLI Flag → Config Merge

Viper ile CLI flag'leri config'e merge edilecek. Öncelik sırası:
1. CLI flags (en yüksek)
2. Environment variables
3. Config file
4. Default values

---

## 7. Versiyon Tespiti

### 7.1 Versiyon Kaynakları (öncelik sırası)

1. **Git tag** (birincil): `git describe --tags --abbrev=0`
2. **VERSION dosyası** (ikincil): Proje kökündeki `VERSION` dosyası
3. **Bumper in** (opsiyonel): Config'de `bumper.in` tanımlıysa o dosya
4. **Fallback**: `0.0.0`

### 7.2 Semver İşlemleri

```go
// internal/version/semver.go

// ParseVersion, versiyon string'ini parse eder.
func ParseVersion(v string) (*semver.Version, error)

// IncrementVersion, versiyon tipine göre artırır.
// incrementType: "major", "minor", "patch", "premajor", "preminor", "prepatch", "prerelease"
func IncrementVersion(current *semver.Version, incrementType string, preReleaseID string) (*semver.Version, error)

// CompareVersions, iki versiyonu karşılaştırır.
func CompareVersions(a, b string) (int, error)
```

### 7.3 Git Tag'den Versiyon Okuma

```go
// internal/version/version.go

// GetLatestTagVersion, git tag'lerden en son versiyonu bulur.
// tagMatch ve tagExclude pattern'leri destekler.
func GetLatestTagVersion(opts VersionOptions) (string, error)

// GetVersionFromFile, VERSION dosyasından versiyon okur.
func GetVersionFromFile(filePath string) (string, error)

type VersionOptions struct {
    TagMatch              string
    TagExclude            string
    GetFromAllRefs        bool
}
```

---

## 8. CLI Yapısı (Cobra)

### 8.1 Komutlar

```
release-it-go                    # Ana release komutu (default)
release-it-go --dry-run          # Dry-run modu
release-it-go --ci               # CI modu (non-interactive)
release-it-go --changelog        # Sadece changelog göster
release-it-go --release-version  # Sadece sonraki versiyonu göster
release-it-go --increment major  # Versiyon tipi belirt
release-it-go version            # Uygulama versiyonunu göster
```

### 8.2 CLI Flags

| Flag | Kısaltma | Tip | Default | Açıklama |
|------|----------|-----|---------|----------|
| `--config` | `-c` | string | "" | Config dosyası yolu |
| `--dry-run` | `-d` | bool | false | Dry-run modu |
| `--ci` | | bool | false | CI modu |
| `--increment` | `-i` | string | "" | major/minor/patch/pre* |
| `--preReleaseId` | | string | "" | Pre-release tanımlayıcı |
| `--verbose` | `-V` | count | 0 | -V=1, -VV=2 |
| `--changelog` | | bool | false | Sadece changelog göster |
| `--release-version` | | bool | false | Sadece versiyon göster |
| `--only-version` | | bool | false | Sadece versiyon sor |
| `--no-increment` | | bool | false | Versiyon artırma |
| `--no-git.commit` | | bool | false | Commit atma |
| `--no-git.tag` | | bool | false | Tag oluşturma |
| `--no-git.push` | | bool | false | Push yapma |

---

## 9. Logger

```go
// internal/log/logger.go

type Logger struct {
    slogger *slog.Logger
    verbose int // 0=normal, 1=verbose, 2=debug
    dryRun  bool
}

// NewLogger, yeni bir logger oluşturur.
func NewLogger(verbose int, dryRun bool) *Logger

// Info, normal mesajlar için.
func (l *Logger) Info(msg string, args ...any)

// Verbose, -V flag'i ile görünen mesajlar.
func (l *Logger) Verbose(msg string, args ...any)

// Debug, -VV flag'i ile görünen mesajlar.
func (l *Logger) Debug(msg string, args ...any)

// DryRun, dry-run mesajları için (prefix: "[dry-run]").
func (l *Logger) DryRun(msg string, args ...any)

// Warn, uyarı mesajları.
func (l *Logger) Warn(msg string, args ...any)

// Error, hata mesajları.
func (l *Logger) Error(msg string, args ...any)
```

---

## 10. Template Variables

Aşağıdaki template değişkenleri config string'lerinde kullanılabilir:

| Değişken | Açıklama |
|----------|----------|
| `${version}` | Yeni versiyon |
| `${latestVersion}` | Önceki versiyon |
| `${latestTag}` | Önceki git tag |
| `${changelog}` | Oluşturulan changelog |
| `${name}` | Proje adı |
| `${repo.remote}` | Remote URL |
| `${repo.protocol}` | Git protocol |
| `${repo.host}` | Repository host |
| `${repo.owner}` | Repository owner |
| `${repo.repository}` | Repository adı |
| `${branchName}` | Branch adı |
| `${releaseUrl}` | Release URL (oluşturulduktan sonra) |

```go
// internal/config/template.go

// RenderTemplate, template string'indeki ${var} ifadelerini değerlerle değiştirir.
func RenderTemplate(tmpl string, vars map[string]string) string
```

---

## 11. Makefile

```makefile
.PHONY: build test lint fmt clean all

BINARY_NAME=release-it-go
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/release-it-go

test:
	go test ./... -v -cover -race

lint:
	golangci-lint run

fmt:
	go fmt ./...
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

all: fmt lint test build
```

---

## 12. Kabul Kriterleri

- [ ] `go mod init github.com/emfi/release-it-go` başarılı
- [ ] `cobra` ile CLI iskeleti oluşturuldu, `release-it-go --help` çalışıyor
- [ ] `release-it-go version` uygulama versiyonunu gösteriyor
- [ ] Config dosyası (.json/.yaml/.toml) yüklenebiliyor
- [ ] Default config değerleri doğru set ediliyor
- [ ] CLI flags config değerlerini override edebiliyor
- [ ] Git tag'den en son versiyon okunabiliyor
- [ ] VERSION dosyasından versiyon okunabiliyor
- [ ] Semver parse/increment/compare çalışıyor
- [ ] Template variable rendering çalışıyor (`${version}` → `1.2.3`)
- [ ] Logger verbose seviyeleri doğru çalışıyor
- [ ] Dry-run modunda `[dry-run]` prefix'i görünüyor
- [ ] `go build ./...` başarılı
- [ ] `go test ./... -race` başarılı
- [ ] Test coverage %70+
- [ ] `go vet ./...` hata yok

---

## 13. Test Senaryoları

### Config Tests
- Default config'in tüm alanları doğru
- JSON config dosyası yüklenebiliyor
- YAML config dosyası yüklenebiliyor
- TOML config dosyası yüklenebiliyor
- Olmayan config dosyasında default kullanılıyor
- CLI flags config'i override ediyor
- Bozuk config dosyasında anlamlı hata

### Version Tests
- Semver parse: "1.2.3", "v1.2.3", "1.2.3-beta.1"
- Increment: patch → 1.2.4, minor → 1.3.0, major → 2.0.0
- Pre-release: "1.2.3" + prerelease("beta") → "1.2.4-beta.0"
- Geçersiz semver string'de hata
- VERSION dosyası okuma (var/yok/boş)

### Template Tests
- Basit değişken: `${version}` → `1.2.3`
- Birden fazla değişken: `Release ${version} for ${name}`
- Olmayan değişken: `${unknown}` → `${unknown}` (değişmez)
- Boş template: `""` → `""`

### Logger Tests
- Verbose=0: sadece Info/Warn/Error görünür
- Verbose=1: Verbose mesajlar da görünür
- Verbose=2: Debug mesajlar da görünür
- DryRun=true: prefix eklenir
