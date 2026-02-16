# Phase 7: Testing, CI/CD, Documentation, Distribution

> **Hedef:** Kapsamli test suite, GitHub Actions CI/CD, GoReleaser entegrasyonu, shell completions ve dokumantasyon.

---

## 1. Genel Bakis

Bu faz, projenin production-ready hale gelmesi icin gerekli tum altyapiyi olusturur: test coverage'i %80+'a cikarma, CI/CD pipeline, otomatik binary dagitimi ve kullanici dokumantasyonu.

---

## 2. Dosya Yapisi

```
.github/
  workflows/
    ci.yml                 # CI pipeline (test, lint, build)
    release.yml            # Release pipeline (GoReleaser)
test/
  integration/
    release_test.go        # End-to-end integration testleri
    fixtures/              # Test fixture'lari (config dosyalari, git repo'lar)
      config_json.json
      config_yaml.yaml
      config_toml.toml
      CHANGELOG_existing.md
.goreleaser.yaml           # GoReleaser config
completions/
  release-it-go.bash       # Bash completion
  release-it-go.zsh        # Zsh completion
  release-it-go.fish       # Fish completion
```

---

## 3. Test Stratejisi

### 3.1 Test Katmanlari

| Katman | Kapsam | Araclar |
|--------|--------|---------|
| Unit Test | Tek fonksiyon/method | `testing` stdlib |
| Integration Test | Birden fazla modulu birlikte | `testing` + `os/exec` + temp dirs |
| API Mock Test | GitHub/GitLab API | `httptest.NewServer` |
| E2E Test | Tam pipeline | Gercek git repo + mock API |

### 3.2 Test Coverage Hedefleri

| Paket | Minimum Coverage |
|-------|-----------------|
| `internal/config` | %80 |
| `internal/version` | %90 |
| `internal/changelog` | %85 |
| `internal/git` | %75 |
| `internal/release` | %75 |
| `internal/bumper` | %80 |
| `internal/hook` | %80 |
| `internal/ui` | %60 (UI testi zor) |
| `internal/runner` | %70 |
| **Toplam** | **%80+** |

### 3.3 Integration Test Yapisi

```go
// test/integration/release_test.go

func TestFullReleasePipeline(t *testing.T) {
    // 1. Gecici git repo olustur
    dir := t.TempDir()
    initGitRepo(t, dir)
    createCommits(t, dir, []string{
        "feat: add user authentication",
        "fix: resolve login timeout",
    })

    // 2. Mock GitHub API server
    server := httptest.NewServer(mockGitHubHandler(t))
    defer server.Close()

    // 3. Config olustur
    cfg := config.DefaultConfig()
    cfg.GitHub.Release = true
    cfg.GitHub.Host = server.URL

    // 4. Pipeline calistir
    runner := runner.NewRunner(cfg)
    err := runner.Run()

    // 5. Dogrulama
    assert.NoError(t, err)
    assertTagExists(t, dir, "1.1.0")
    assertChangelogUpdated(t, dir)
}
```

### 3.4 Test Helper'lar

```go
// test/integration/helpers.go

// initGitRepo, gecici bir git repo olusturur ve initial commit atar.
func initGitRepo(t *testing.T, dir string)

// createCommits, belirtilen mesajlarla commit'ler olusturur.
func createCommits(t *testing.T, dir string, messages []string)

// createTag, belirtilen isimle tag olusturur.
func createTag(t *testing.T, dir string, tag string)

// assertTagExists, tag'in var oldugunu dogrular.
func assertTagExists(t *testing.T, dir string, tag string)

// assertChangelogUpdated, CHANGELOG.md'nin guncellendigini dogrular.
func assertChangelogUpdated(t *testing.T, dir string)

// mockGitHubHandler, GitHub API mock handler doner.
func mockGitHubHandler(t *testing.T) http.Handler

// mockGitLabHandler, GitLab API mock handler doner.
func mockGitLabHandler(t *testing.T) http.Handler
```

---

## 4. CI/CD Pipeline

### 4.1 CI Workflow (ci.yml)

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

      - name: Check coverage
        run: |
          go tool cover -func=coverage.txt
          COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage is below 80%"
            exit 1
          fi

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: go build -o release-it-go-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/release-it-go
```

### 4.2 Release Workflow (release.yml)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 5. GoReleaser Config

```yaml
# .goreleaser.yaml
version: 2

project_name: release-it-go

before:
  hooks:
    - go mod tidy
    - go test ./... -race

builds:
  - main: ./cmd/release-it-go
    binary: release-it-go
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

brews:
  - repository:
      owner: emfi
      name: homebrew-tap
    homepage: "https://github.com/emfi/release-it-go"
    description: "Release automation tool for Git projects"
    install: |
      bin.install "release-it-go"

nfpms:
  - package_name: release-it-go
    homepage: "https://github.com/emfi/release-it-go"
    maintainer: "emfi"
    description: "Release automation tool for Git projects"
    formats:
      - deb
      - rpm
      - apk
```

---

## 6. Shell Completions

### 6.1 Cobra ile Otomatik Olusturma

```go
// cmd/release-it-go/main.go (veya completion.go)

func newCompletionCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "completion [bash|zsh|fish|powershell]",
        Short: "Generate shell completion scripts",
        ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            switch args[0] {
            case "bash":
                return cmd.Root().GenBashCompletion(os.Stdout)
            case "zsh":
                return cmd.Root().GenZshCompletion(os.Stdout)
            case "fish":
                return cmd.Root().GenFishCompletion(os.Stdout, true)
            case "powershell":
                return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
            default:
                return fmt.Errorf("unsupported shell: %s", args[0])
            }
        },
    }
}
```

### 6.2 Kurulum Talimatlari

```bash
# Bash
release-it-go completion bash > /etc/bash_completion.d/release-it-go

# Zsh
release-it-go completion zsh > "${fpath[1]}/_release-it-go"

# Fish
release-it-go completion fish > ~/.config/fish/completions/release-it-go.fish
```

---

## 7. Build Info

```go
// cmd/release-it-go/main.go

// ldflags ile set edilir
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func printVersion() {
    fmt.Printf("release-it-go %s (commit: %s, built: %s)\n", version, commit, date)
}
```

---

## 8. Dagitim Kanallari

| Kanal | Komut |
|-------|-------|
| Go install | `go install github.com/emfi/release-it-go/cmd/release-it-go@latest` |
| GitHub Releases | Binary download |
| Homebrew | `brew install emfi/tap/release-it-go` |
| apt/yum | `.deb` / `.rpm` paketleri |

---

## 9. Kabul Kriterleri

### Testing
- [ ] Tum unit testler geciyor
- [ ] Integration testler geciyor (gercek git repo + mock API)
- [ ] Race condition testi geciyor (`go test -race`)
- [ ] Toplam test coverage %80+
- [ ] Her paket kendi minimum coverage hedefini karsilar
- [ ] Test fixture'lari (config dosyalari) mevcut ve dogru

### CI/CD
- [ ] CI workflow calisiyor (test + lint + build)
- [ ] Multi-platform build basarili (linux/darwin/windows x amd64/arm64)
- [ ] GoReleaser ile release olusturuluyor
- [ ] Tag push'ta otomatik release tetikleniyor
- [ ] Coverage raporlama calisiyor

### Distribution
- [ ] `go install` ile kurulabiliyor
- [ ] Binary'ler GitHub Releases'ta mevcut
- [ ] Shell completions calisiyor (bash/zsh/fish)
- [ ] `release-it-go version` dogru build info gosteriyor

### Documentation
- [ ] Tum public fonksiyonlarda GoDoc comment var
- [ ] Paket aciklamalari mevcut

---

## 10. Test Senaryolari

### Integration Tests
- Tam release pipeline: commit -> tag -> push -> GitHub release
- Dry-run pipeline: hicbir degisiklik yapilmadi
- CI mode pipeline: prompt yok, otomatik
- Config dosyasindan okuma (JSON, YAML, TOML)
- Conventional commits -> otomatik minor bump
- Breaking change -> otomatik major bump
- CHANGELOG.md guncelleme (mevcut dosya + yeni dosya)
- Bumper: birden fazla dosya guncelleme
- CalVer: takvim degisikligi
- Pre-release akisi
- Hook calistirma (basarili + basarisiz)
- --changelog modu
- --release-version modu

### API Mock Tests
- GitHub: create release -> 201
- GitHub: create release -> 401 (token hatasi)
- GitHub: create release -> 422 (tag yok)
- GitHub: upload asset -> basarili
- GitHub: rate limit -> 429
- GitLab: create release -> basarili
- GitLab: upload asset -> basarili (2 adim)

### Build Tests
- Cross-compilation basarili (GOOS/GOARCH)
- ldflags ile version bilgisi dogru
- Binary boyutu makul (< 20MB)
