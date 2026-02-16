# Phase 2: Git Operations

> **Hedef:** Git prerequisite checks, commit, tag, push islemleri ve dry-run destegi.

---

## 1. Genel Bakis

Bu faz, release surecinin temel git islemlerini implement eder. Prerequisite kontrolleri yapildiktan sonra sirasiyla: staging, commit, tag, push adimlari yurutulur. Tum adimlar dry-run modunu destekler.

---

## 2. Dosya Yapisi

```
internal/
  git/
    git.go              # Git komutlarini calistiran ana modul
    prerequisites.go    # Prerequisite checks (branch, clean, upstream)
    commit.go           # Staging + commit
    tag.go              # Tag olusturma
    push.go             # Push (tags dahil)
    changelog.go        # Basit git log tabanli changelog
    repo.go             # Repo bilgileri (remote, owner, host parse)
```

---

## 3. Bagimliliklar

| Kutuphane | Kullanim |
|-----------|----------|
| `os/exec` (stdlib) | Git komutlari calistirma |
| `strings`, `regexp` (stdlib) | Remote URL parse |

> **Not:** `go-git` kutuphanesi kullanilmayacak. Tum git islemleri `git` CLI binary'si uzerinden yapilacak. Bu, kullanicinin mevcut git config'ini (gpg signing, hooks, credentials) otomatik olarak kullanmasini saglar.

---

## 4. Git Runner

```go
// internal/git/git.go

type Git struct {
    config  *config.GitConfig
    logger  *log.Logger
    dryRun  bool
    workDir string
}

// NewGit, yeni bir Git instance olusturur.
func NewGit(cfg *config.GitConfig, logger *log.Logger, dryRun bool) *Git

// run, git komutunu calistirir. dryRun=true ise sadece loglar.
func (g *Git) run(args ...string) (string, error)

// runSilent, ciktiyi loglamadan git komutu calistirir.
func (g *Git) runSilent(args ...string) (string, error)
```

---

## 5. Prerequisite Checks

### 5.1 Kontrol Listesi

| Kontrol | Config | Varsayilan | Aciklama |
|---------|--------|------------|----------|
| Branch kontrolu | `git.requireBranch` | `""` (kapali) | Belirtilen branch'te olunmali |
| Temiz working dir | `git.requireCleanWorkingDir` | `true` | Uncommitted degisiklik olmamali |
| Upstream tracking | `git.requireUpstream` | `true` | Remote tracking branch olmali |
| Yeni commit var mi | `git.requireCommits` | `false` | Son tag'den beri commit olmali |

### 5.2 Implementasyon

```go
// internal/git/prerequisites.go

// CheckPrerequisites, tum prerequisite kontrolleri yapar.
// Basarisiz olan ilk kontrolde hata doner.
func (g *Git) CheckPrerequisites() error {
    if err := g.checkBranch(); err != nil {
        return err
    }
    if err := g.checkCleanWorkingDir(); err != nil {
        return err
    }
    if err := g.checkUpstream(); err != nil {
        return err
    }
    if err := g.checkCommits(); err != nil {
        return err
    }
    return nil
}

// checkBranch: git rev-parse --abbrev-ref HEAD
func (g *Git) checkBranch() error

// checkCleanWorkingDir: git status --porcelain
func (g *Git) checkCleanWorkingDir() error

// checkUpstream: git rev-parse --abbrev-ref --symbolic-full-name @{u}
func (g *Git) checkUpstream() error

// checkCommits: git log ${latestTag}..HEAD --oneline
func (g *Git) checkCommits() error
```

---

## 6. Commit

### 6.1 Akis

1. Degisen dosyalari staging'e ekle (`git add . --update` veya `git add .` if addUntrackedFiles)
2. Commit mesajini template'den olustur
3. Commit at (commitArgs dahil)

### 6.2 Implementasyon

```go
// internal/git/commit.go

// Stage, degisiklikleri staging'e ekler.
func (g *Git) Stage() error {
    if g.config.AddUntrackedFiles {
        _, err := g.run("add", ".")
        return err
    }
    _, err := g.run("add", ".", "--update")
    return err
}

// Commit, verilen mesajla commit olusturur.
func (g *Git) Commit(message string) error {
    args := []string{"commit", "--message", message}
    args = append(args, g.config.CommitArgs...)
    _, err := g.run(args...)
    return err
}
```

---

## 7. Tag

### 7.1 Akis

1. Tag adini template'den olustur
2. Annotated tag olustur (tagAnnotation ile)
3. tagArgs ekle

### 7.2 Implementasyon

```go
// internal/git/tag.go

// CreateTag, yeni bir annotated tag olusturur.
func (g *Git) CreateTag(tagName string, annotation string) error {
    args := []string{"tag", "--annotate", "--message", annotation, tagName}
    args = append(args, g.config.TagArgs...)
    _, err := g.run(args...)
    return err
}

// GetLatestTag, en son tag'i doner.
func (g *Git) GetLatestTag() (string, error) {
    if g.config.GetLatestTagFromAllRefs {
        return g.getLatestTagFromAllRefs()
    }
    // git describe --tags --abbrev=0
    return g.runSilent("describe", "--tags", "--abbrev=0")
}

// getLatestTagFromAllRefs, tum ref'lerden tag'leri semver'e gore siralar.
func (g *Git) getLatestTagFromAllRefs() (string, error)

// TagExists, belirtilen tag'in var olup olmadigini kontrol eder.
func (g *Git) TagExists(tagName string) (bool, error)
```

---

## 8. Push

### 8.1 Akis

1. Push repo'yu belirle (default: origin)
2. Push args'lari ekle (default: --follow-tags)
3. Push et

### 8.2 Implementasyon

```go
// internal/git/push.go

// Push, commit ve tag'leri remote'a gonderir.
func (g *Git) Push() error {
    args := []string{"push"}
    args = append(args, g.config.PushArgs...)

    if g.config.PushRepo != "" {
        args = append(args, g.config.PushRepo)
    }

    _, err := g.run(args...)
    return err
}
```

---

## 9. Repo Bilgileri

```go
// internal/git/repo.go

type RepoInfo struct {
    Remote     string // full remote URL
    Protocol   string // "https" or "ssh"
    Host       string // "github.com"
    Owner      string // "emfi"
    Repository string // "release-it-go"
}

// GetRepoInfo, git remote'dan repo bilgilerini parse eder.
// Hem HTTPS hem SSH URL formatlarini destekler:
//   https://github.com/owner/repo.git
//   git@github.com:owner/repo.git
func GetRepoInfo(remoteName string) (*RepoInfo, error)

// GetBranchName, mevcut branch adini doner.
func GetBranchName() (string, error)
```

---

## 10. Basit Changelog (Git Log)

```go
// internal/git/changelog.go

// GenerateChangelog, iki tag arasindaki commit'lerden changelog olusturur.
// Template'deki git log format string'ini kullanir.
func (g *Git) GenerateChangelog(fromTag string, toRef string) (string, error)

// GetCommitsSinceTag, son tag'den bu yana olan commit'leri listeler.
func (g *Git) GetCommitsSinceTag(tag string) ([]string, error)
```

---

## 11. Dry-Run Destegi

Tum git yazma islemleri dry-run modunda sadece loglanir, calistirilmaz:

```go
func (g *Git) run(args ...string) (string, error) {
    cmd := fmt.Sprintf("git %s", strings.Join(args, " "))

    if g.dryRun && isWriteOperation(args) {
        g.logger.DryRun(cmd)
        return "", nil
    }

    g.logger.Debug("run: %s", cmd)
    // actual run...
}

// isWriteOperation, komutun yazma islemi olup olmadigini belirler.
// Read-only komutlar (status, log, describe, rev-parse) dry-run'da da calisir.
func isWriteOperation(args []string) bool {
    readOnlyCommands := map[string]bool{
        "status": true, "log": true, "describe": true,
        "rev-parse": true, "remote": true,
        "diff": true, "show": true,
    }
    if len(args) == 0 {
        return false
    }
    _, isReadOnly := readOnlyCommands[args[0]]
    return !isReadOnly
}
```

---

## 12. Hata Senaryolari

| Senaryo | Hata Mesaji |
|---------|-------------|
| Git kurulu degil | `"git is not installed or not in PATH"` |
| Git repo degil | `"current directory is not a git repository"` |
| Yanlis branch | `"required branch is %s, but current branch is %s"` |
| Dirty working dir | `"working directory is not clean (uncommitted changes exist)"` |
| Upstream yok | `"no upstream configured for current branch"` |
| Yeni commit yok | `"no commits since latest tag %s"` |
| Tag zaten var | `"tag %s already exists"` |
| Push basarisiz | `"failed to push to %s: %w"` |

---

## 13. Kabul Kriterleri

- [ ] `git.requireBranch` kontrolu calisiyor (string ve pattern destegi)
- [ ] `git.requireCleanWorkingDir` kontrolu calisiyor
- [ ] `git.requireUpstream` kontrolu calisiyor
- [ ] `git.requireCommits` kontrolu calisiyor
- [ ] Dosyalar staging'e eklenebiliyor (`addUntrackedFiles` destegi)
- [ ] Commit olusturulabiliyor (commitMessage template + commitArgs)
- [ ] Annotated tag olusturulabiliyor (tagName template + tagAnnotation + tagArgs)
- [ ] Push yapilabiliyor (pushRepo + pushArgs)
- [ ] Remote URL parse edilebiliyor (HTTPS + SSH)
- [ ] Basit git log changelog calisiyor
- [ ] Dry-run modunda yazma islemleri yapilmiyor
- [ ] Dry-run modunda okuma islemleri calismaya devam ediyor
- [ ] Tag cakismasi kontrolu yapiliyor
- [ ] Tum hata senaryolari anlamli mesajlarla handle ediliyor
- [ ] `go test ./internal/git/... -race` basarili
- [ ] Test coverage %70+

---

## 14. Test Senaryolari

### Prerequisites Tests
- Dogru branch'te: hata yok
- Yanlis branch'te: hata
- Temiz working dir: hata yok
- Dirty working dir: hata
- Upstream var: hata yok
- Upstream yok: hata
- Commit var: hata yok
- Commit yok: hata
- Tum kontroller kapali: hata yok

### Commit Tests
- Staging (update only): sadece tracked dosyalar
- Staging (with untracked): tum dosyalar
- Commit mesaji template rendering
- CommitArgs ekleniyor
- Dry-run'da commit yapilmiyor

### Tag Tests
- Tag olusturma (annotated)
- Tag adi template rendering
- TagArgs ekleniyor
- Mevcut tag cakismasi
- Dry-run'da tag olusturulmuyor

### Push Tests
- Default push (origin, --follow-tags)
- Custom pushRepo
- Custom pushArgs
- Dry-run'da push yapilmiyor

### Repo Info Tests
- HTTPS URL parse: `https://github.com/owner/repo.git`
- SSH URL parse: `git@github.com:owner/repo.git`
- GitLab URL parse
- Gecersiz URL
