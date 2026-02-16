# Phase 3: Conventional Commits + Changelog

> **Hedef:** Commit parser, otomatik versiyon artirma, Angular preset ve CHANGELOG.md olusturma (conventional-changelog + keep-a-changelog formatlari).

---

## 1. Genel Bakis

Bu faz, conventional commits standardina uygun commit'leri parse eder, otomatik olarak versiyon bump tipini belirler ve CHANGELOG.md dosyasini iki farkli formatta olusturabilir.

---

## 2. Dosya Yapisi

```
internal/
  changelog/
    parser.go              # Conventional commit parser
    analyzer.go            # Bump tipi analizi (major/minor/patch)
    generator.go           # Changelog olusturma orchestrator
    conventional.go        # Conventional-changelog formatinda cikti
    keepachangelog.go      # Keep-a-changelog formatinda cikti
    types.go               # Ortak tipler (Commit, ChangelogEntry, vb.)
```

---

## 3. Conventional Commit Format

### 3.1 Commit Yapisi

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 3.2 Desteklenen Tipler (Angular Preset)

| Tip | Aciklama | Changelog Bolumu | Bump |
|-----|----------|-------------------|------|
| `feat` | Yeni ozellik | Features | minor |
| `fix` | Bug duzeltme | Bug Fixes | patch |
| `docs` | Dokumantasyon | - (changelog'a eklenmez) | - |
| `style` | Formatting | - | - |
| `refactor` | Kod duzenleme | - | - |
| `perf` | Performans | Performance Improvements | patch |
| `test` | Test | - | - |
| `build` | Build sistemi | - | - |
| `ci` | CI config | - | - |
| `chore` | Bakim isleri | - | - |
| `revert` | Geri alma | Reverts | patch |

### 3.3 Breaking Change

Breaking change su sekillerde belirtilir:
- Footer'da: `BREAKING CHANGE: description`
- Tip'te unlem: `feat!: description` veya `feat(scope)!: description`
- Breaking change **her zaman major bump** tetikler

---

## 4. Commit Parser

```go
// internal/changelog/types.go

type Commit struct {
    Hash            string
    Type            string   // feat, fix, docs, etc.
    Scope           string   // optional scope
    Description     string   // commit description
    Body            string   // optional body
    Footers         []Footer
    BreakingChange  bool
    BreakingMessage string   // BREAKING CHANGE aciklamasi
    Raw             string   // orijinal commit mesaji
}

type Footer struct {
    Token string // "BREAKING CHANGE", "Closes", "Refs", etc.
    Value string
}

// internal/changelog/parser.go

// ParseCommit, tek bir commit mesajini parse eder.
func ParseCommit(raw string, hash string) (*Commit, error)

// ParseCommits, birden fazla commit'i parse eder.
// Conventional commit formatina uymayan commit'ler atlanir (hata degil).
func ParseCommits(rawCommits []RawCommit) []*Commit

type RawCommit struct {
    Hash    string
    Message string
}
```

### 4.1 Regex Pattern

```go
// Conventional commit regex
var commitPattern = regexp.MustCompile(
    `^(?P<type>\w+)(?:\((?P<scope>[^)]+)\))?(?P<breaking>!)?:\s+(?P<description>.+)$`,
)
```

---

## 5. Bump Analyzer

```go
// internal/changelog/analyzer.go

type BumpType int

const (
    BumpNone  BumpType = iota
    BumpPatch
    BumpMinor
    BumpMajor
)

func (b BumpType) String() string // "patch", "minor", "major", ""

// AnalyzeBump, commit'leri analiz ederek onerilen bump tipini belirler.
// Kurallar:
//   - Breaking change -> major
//   - feat -> minor
//   - fix, perf, revert -> patch
//   - Hicbiri -> none
// En yuksek seviyeli bump doner.
func AnalyzeBump(commits []*Commit) BumpType

// AnalyzeBumpWithConfig, preset'e gore custom kurallar uygular.
func AnalyzeBumpWithConfig(commits []*Commit, preset string) BumpType
```

---

## 6. Changelog Generator

### 6.1 Conventional-Changelog Format

```markdown
# Changelog

## [1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0) (2026-02-16)

### Features

* **auth:** add OAuth2 support ([abc1234](https://github.com/owner/repo/commit/abc1234))
* implement dark mode ([def5678](https://github.com/owner/repo/commit/def5678))

### Bug Fixes

* **api:** fix rate limiting issue ([fed4321](https://github.com/owner/repo/commit/fed4321))

### BREAKING CHANGES

* **auth:** OAuth1 support has been removed
```

### 6.2 Keep-a-Changelog Format

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-02-16

### Added
- OAuth2 support for authentication
- Dark mode implementation

### Fixed
- Rate limiting issue in API module

### Changed
- Updated dependency versions
```

### 6.3 Implementasyon

```go
// internal/changelog/generator.go

type ChangelogOptions struct {
    Preset         string // "angular" (default)
    Infile         string // "CHANGELOG.md"
    Header         string // "# Changelog"
    KeepAChangelog bool   // false = conventional format
    AddUnreleased  bool
    KeepUnreleased bool
    AddVersionUrl  bool
    RepoInfo       *git.RepoInfo
}

// GenerateChangelog, commit'lerden changelog icerigi olusturur.
// Sadece yeni versiyon bolumunu doner (dosya yazma yapmaz).
func GenerateChangelog(commits []*Commit, version string, prevVersion string, opts ChangelogOptions) (string, error)

// UpdateChangelogFile, mevcut CHANGELOG.md dosyasina yeni versiyonu ekler.
// Dosya yoksa yeni olusturur.
func UpdateChangelogFile(filePath string, newContent string, header string) error
```

```go
// internal/changelog/conventional.go

// RenderConventional, conventional-changelog formatinda markdown olusturur.
func RenderConventional(commits []*Commit, version string, prevVersion string, repoInfo *git.RepoInfo) string
```

```go
// internal/changelog/keepachangelog.go

// RenderKeepAChangelog, keep-a-changelog formatinda markdown olusturur.
func RenderKeepAChangelog(commits []*Commit, version string, date string) string
```

---

## 7. Commit Tipi ve Changelog Bolumu Eslesmesi

### Conventional Format

| Commit Type | Bolum Basligi |
|-------------|---------------|
| `feat` | Features |
| `fix` | Bug Fixes |
| `perf` | Performance Improvements |
| `revert` | Reverts |
| Breaking changes | BREAKING CHANGES |

### Keep-a-Changelog Format

| Commit Type | Bolum Basligi |
|-------------|---------------|
| `feat` | Added |
| `fix` | Fixed |
| `perf`, `refactor` | Changed |
| `revert` | Removed |
| Deprecated | Deprecated |
| Security | Security |

---

## 8. Git Log to Commit Okuma

```go
// Git'den commit listesi cekme formati
// git log v1.1.0..HEAD --pretty=format:"%H|||%s|||%b" --no-merges

const gitLogFormat = "%H|||%s|||%b"
const gitLogSeparator = "|||"

// GetCommitsSince, belirtilen tag'den bu yana olan commit'leri doner.
func GetCommitsSince(tag string) ([]RawCommit, error)
```

---

## 9. Dosya Yazma Stratejisi

### CHANGELOG.md Guncelleme

1. Mevcut dosyayi oku
2. Header'i bul (default: `# Changelog`)
3. Header'dan sonra yeni versiyon bolumunu ekle
4. Geri kalanini koru
5. Dosyaya yaz

```go
// insertAfterHeader, header'dan sonra yeni icerik ekler.
func insertAfterHeader(existing string, header string, newSection string) string {
    idx := strings.Index(existing, header)
    if idx == -1 {
        // Header yoksa basa ekle
        return header + "\n\n" + newSection + "\n\n" + existing
    }
    headerEnd := idx + len(header)
    return existing[:headerEnd] + "\n\n" + newSection + existing[headerEnd:]
}
```

---

## 10. Kabul Kriterleri

- [ ] Conventional commit format'i dogru parse ediliyor (type, scope, description, body, footer)
- [ ] Breaking change algilaniyor (footer ve `!` syntax)
- [ ] Angular preset commit tipleri taniniyor
- [ ] Otomatik bump tipi dogru belirleniyor:
  - Breaking change -> major
  - feat -> minor
  - fix -> patch
  - Karisik commit'lerde en yuksek seviye seciliyor
- [ ] Conventional-changelog formatinda markdown olusturuluyor
- [ ] Keep-a-changelog formatinda markdown olusturuluyor
- [ ] CHANGELOG.md dosyasi guncelleniyor (prepend, mevcut icerik korunuyor)
- [ ] CHANGELOG.md yoksa yeni olusturuluyor
- [ ] Commit hash'leri ve compare URL'leri dogru
- [ ] Scope'lu commit'ler dogru gosteriliyor
- [ ] Tarih formati dogru (YYYY-MM-DD)
- [ ] `addUnreleased` secenegi calisiyor
- [ ] `addVersionUrl` compare URL'leri dogru
- [ ] Non-conventional commit'ler atlaniyor (hata vermiyor)
- [ ] `go test ./internal/changelog/... -race` basarili
- [ ] Test coverage %80+

---

## 11. Test Senaryolari

### Parser Tests
- Basit commit: `feat: add login` -> type=feat, desc="add login"
- Scope'lu: `fix(api): rate limit` -> type=fix, scope=api
- Breaking change (footer): `BREAKING CHANGE: removed X`
- Breaking change (!): `feat!: new API`
- Body'li commit
- Multi-line body
- Footer'li: `Closes #123`
- Gecersiz format: `random message` -> nil (hata degil)
- Bos mesaj

### Analyzer Tests
- Sadece feat -> minor
- Sadece fix -> patch
- feat + fix -> minor
- Breaking change + feat -> major
- Hic conventional commit yok -> none
- Sadece docs/chore -> none

### Conventional Format Tests
- Tek feat commit -> Features bolumu
- Karma commit'ler -> dogru bolumleme
- Breaking change bolumu
- Compare URL dogrulugu
- Tarih formati

### Keep-a-Changelog Format Tests
- feat -> Added bolumu
- fix -> Fixed bolumu
- Unreleased bolumu
- Version URL'leri

### File Update Tests
- Yeni dosya olusturma
- Mevcut dosyaya prepend
- Header korunmasi
- Unreleased bolumu guncelleme
