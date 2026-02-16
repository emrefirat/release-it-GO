# Phase 6: Advanced Features

> **Hedef:** Bumper (coklu dosya versiyon guncelleme), calver, pre-release flows, ozel CLI modlari ve verbose/debug modlari.

---

## 1. Genel Bakis

Bu faz, release-it-go'nun ileri seviye ozelliklerini implement eder: birden fazla dosyada versiyon guncelleme (bumper), takvim bazli versiyonlama (calver), pre-release akislari ve ozel CLI modlari.

---

## 2. Dosya Yapisi

```
internal/
  bumper/
    bumper.go              # Bumper ana modulu
    reader.go              # Dosyadan versiyon okuma
    writer.go              # Dosyaya versiyon yazma
    formats.go             # Dosya formatlari (JSON, YAML, TOML, INI, text)
  version/
    calver.go              # Calendar versioning (mevcut dosya genisletilir)
```

---

## 3. Bumper

### 3.1 Desteklenen Dosya Formatlari

| Format | Uzanti | Ornek Yol |
|--------|--------|-----------|
| JSON | `.json` | `package.json` -> `$.version` |
| YAML | `.yaml`, `.yml` | `chart.yaml` -> `$.version` |
| TOML | `.toml` | `pyproject.toml` -> `$.tool.poetry.version` |
| INI | `.ini` | `setup.cfg` -> `[metadata].version` |
| Text | `.txt`, `.VERSION` | Tum dosya = versiyon |

### 3.2 Implementasyon

```go
// internal/bumper/bumper.go

type Bumper struct {
    config *config.BumperConfig
    logger *log.Logger
    dryRun bool
}

// NewBumper, yeni bir bumper instance olusturur.
func NewBumper(cfg *config.BumperConfig, logger *log.Logger, dryRun bool) *Bumper

// ReadVersion, in dosyasindan mevcut versiyonu okur.
// Eger in tanimli degilse bos string doner.
func (b *Bumper) ReadVersion() (string, error)

// WriteVersion, tum out dosyalarina yeni versiyonu yazar.
func (b *Bumper) WriteVersion(version string) error
```

### 3.3 Dosyadan Okuma

```go
// internal/bumper/reader.go

// ReadVersionFromFile, belirtilen dosya ve yoldan versiyonu okur.
func ReadVersionFromFile(file BumperFile) (string, error) {
    data, err := os.ReadFile(file.File)
    if err != nil {
        return "", fmt.Errorf("reading %s: %w", file.File, err)
    }

    format := detectFormat(file)

    switch format {
    case "json":
        return readJSON(data, file.Path)
    case "yaml":
        return readYAML(data, file.Path)
    case "toml":
        return readTOML(data, file.Path)
    case "ini":
        return readINI(data, file.Path)
    case "text":
        return readText(data)
    default:
        return "", fmt.Errorf("unsupported format for %s", file.File)
    }
}
```

### 3.4 Dosyaya Yazma

```go
// internal/bumper/writer.go

// WriteVersionToFile, belirtilen dosyadaki versiyonu gunceller.
func WriteVersionToFile(file BumperFile, version string) error {
    if file.ConsumeWholeFile {
        // Tum dosya icerigi = versiyon
        return os.WriteFile(file.File, []byte(version+"\n"), 0644)
    }

    data, err := os.ReadFile(file.File)
    if err != nil {
        return fmt.Errorf("reading %s: %w", file.File, err)
    }

    // Prefix ekleme
    finalVersion := file.Prefix + version

    format := detectFormat(file)
    updated, err := updateVersion(data, format, file.Path, finalVersion)
    if err != nil {
        return fmt.Errorf("updating version in %s: %w", file.File, err)
    }

    return os.WriteFile(file.File, updated, 0644)
}
```

### 3.5 Format Detection

```go
// detectFormat, dosya uzantisindan veya type config'inden formati belirler.
func detectFormat(file BumperFile) string {
    if file.Type != "" {
        return file.Type
    }
    ext := filepath.Ext(file.File)
    switch ext {
    case ".json":
        return "json"
    case ".yaml", ".yml":
        return "yaml"
    case ".toml":
        return "toml"
    case ".ini", ".cfg":
        return "ini"
    default:
        return "text"
    }
}
```

### 3.6 Glob Pattern Destegi

```go
// Out dizisinde glob pattern kullanilabilir:
// out: ["dist/*.json", "charts/*/Chart.yaml"]

func (b *Bumper) WriteVersion(version string) error {
    for _, outFile := range b.config.Out {
        files, err := resolveGlob(outFile.File)
        if err != nil {
            return err
        }
        for _, f := range files {
            file := outFile // copy
            file.File = f
            if err := WriteVersionToFile(file, version); err != nil {
                return err
            }
        }
    }
    return nil
}
```

---

## 4. Calendar Versioning (CalVer)

### 4.1 Desteklenen Formatlar

| Format | Ornek | Aciklama |
|--------|-------|----------|
| `yy.mm.minor` | `26.2.0` | 2-digit yil, ay, minor |
| `yyyy.mm.minor` | `2026.2.0` | 4-digit yil, ay, minor |
| `yy.mm.micro` | `26.2.0` | micro = minor alias |
| `yyyy.mm.dd` | `2026.2.16` | 4-digit yil, ay, gun |

### 4.2 Implementasyon

```go
// internal/version/calver.go

type CalVer struct {
    config *config.CalVerConfig
}

// NewCalVer, yeni bir CalVer instance olusturur.
func NewCalVer(cfg *config.CalVerConfig) *CalVer

// NextVersion, bir sonraki CalVer versiyonunu hesaplar.
func (cv *CalVer) NextVersion(current string) (string, error) {
    now := time.Now()
    currentParts, err := cv.parse(current)
    if err != nil {
        // Ilk versiyon
        return cv.format(now, 0), nil
    }

    // Takvim bilesenlerini karsilastir
    if cv.calendarChanged(currentParts, now) {
        // Yeni takvim donemi, minor sifirla
        return cv.format(now, 0), nil
    }

    // Ayni donem, minor artir
    return cv.format(now, currentParts.Minor+1), nil
}

// Parse, CalVer string'ini bilesenlere ayirir.
func (cv *CalVer) parse(version string) (*CalVerParts, error)

// Format, bilesenlerden CalVer string olusturur.
func (cv *CalVer) format(t time.Time, minor int) string

type CalVerParts struct {
    Year  int
    Month int
    Day   int
    Minor int
}
```

### 4.3 CalVer + Increment

```go
// Increment config'e gore davranis:
// "calendar": Takvim degistiyse minor=0, ayni kaldiysa minor++
// "calendar.minor": Takvim degistiyse minor=0, ayni kaldiysa minor++
// (Iki mod arasindaki fark: sadece takvim degisiklik formati)

func (cv *CalVer) calendarChanged(parts *CalVerParts, now time.Time) bool {
    switch cv.config.Format {
    case "yy.mm.minor", "yyyy.mm.minor":
        return parts.Year != now.Year() || parts.Month != int(now.Month())
    case "yyyy.mm.dd":
        return parts.Year != now.Year() || parts.Month != int(now.Month()) || parts.Day != now.Day()
    default:
        return parts.Year != now.Year() || parts.Month != int(now.Month())
    }
}
```

---

## 5. Pre-Release Flows

### 5.1 Pre-Release Olusturma

```bash
# Beta release olustur
release-it-go --increment premajor --preReleaseId beta
# 1.2.3 -> 2.0.0-beta.0

release-it-go --increment prerelease --preReleaseId beta
# 2.0.0-beta.0 -> 2.0.0-beta.1

# RC release
release-it-go --increment prerelease --preReleaseId rc
# 2.0.0-beta.1 -> 2.0.0-rc.0
```

### 5.2 Pre-Release Davranislari

- GitHub: `preRelease: true` otomatik set edilir (semver pre-release algisi)
- Tag: pre-release tag'leri `tagExclude` ile filtrelenebilir
- Changelog: pre-release'ler ayri bolumlerde gosterilebilir
- Push: pre-release icin farkli branch'e push edilebilir

### 5.3 Implementasyon

```go
// Pre-release versiyon artirma mantigi
func IncrementVersion(current *semver.Version, incrementType string, preReleaseID string) (*semver.Version, error) {
    switch incrementType {
    case "premajor":
        next := current.IncMajor()
        return addPreRelease(next, preReleaseID, 0)
    case "preminor":
        next := current.IncMinor()
        return addPreRelease(next, preReleaseID, 0)
    case "prepatch":
        next := current.IncPatch()
        return addPreRelease(next, preReleaseID, 0)
    case "prerelease":
        return incrementPreRelease(current, preReleaseID)
    default:
        // major, minor, patch
        return standardIncrement(current, incrementType)
    }
}

func incrementPreRelease(current *semver.Version, preReleaseID string) (*semver.Version, error) {
    pre := current.Prerelease()
    if pre == "" {
        // Henuz pre-release degil, patch + pre-release
        next := current.IncPatch()
        return addPreRelease(next, preReleaseID, 0)
    }

    // Mevcut pre-release'in ID'sini kontrol et
    currentID, currentNum := parsePreRelease(pre)
    if currentID == preReleaseID {
        // Ayni ID, numarayi artir
        return addPreRelease(*current, preReleaseID, currentNum+1)
    }

    // Farkli ID, yeni pre-release serisi
    return addPreRelease(*current, preReleaseID, 0)
}
```

---

## 6. Ozel CLI Modlari

### 6.1 --no-increment

Versiyon artirmadan mevcut release'i gunceller:
- CHANGELOG.md guncelleme
- GitHub/GitLab release guncelleme
- Asset ekleme

```go
func (r *Runner) RunNoIncrement() error {
    // Versiyon artirma adimini atla
    r.ctx.Version = r.ctx.LatestVersion
    // Geri kalan adimlari normal calistir
    return r.runFrom("changelog")
}
```

### 6.2 --only-version

Sadece versiyon prompt'u goster, kalanini otomatik yap:

```go
func (r *Runner) RunOnlyVersion() error {
    version, err := r.ctx.Prompter.SelectVersion(...)
    if err != nil {
        return err
    }
    r.ctx.Version = version
    // Geri kalan adimlari CI modunda calistir (onaysiz)
    r.ctx.IsCI = true
    return r.runFrom("changelog")
}
```

### 6.3 --release-version

Sadece sonraki versiyonu hesapla ve yazdir:

```go
func (r *Runner) RunReleaseVersionOnly() error {
    version, err := r.determineNextVersion()
    if err != nil {
        return err
    }
    fmt.Println(version)
    return nil
}
```

### 6.4 --changelog

Sadece changelog olustur ve yazdir:

```go
func (r *Runner) RunChangelogOnly() error {
    changelog, err := r.generateChangelogContent()
    if err != nil {
        return err
    }
    fmt.Println(changelog)
    return nil
}
```

---

## 7. Verbose/Debug Modlari

### 7.1 Normal Mod (verbose=0)
- Sadece onemli mesajlar
- Spinner ile ilerleme
- Sonuc ozeti

### 7.2 Verbose Mod (-V, verbose=1)
- Hook komutlari ve ciktilari
- Git komutlari
- API endpoint'leri

### 7.3 Debug Mod (-VV, verbose=2)
- Tum internal komutlar
- Config dump
- HTTP request/response headers
- Dosya okuma/yazma detaylari

```go
// Verbose seviyesine gore log filtreleme
func (l *Logger) shouldLog(level int) bool {
    return level <= l.verbose
}
```

---

## 8. Kabul Kriterleri

### Bumper
- [ ] JSON dosyadan versiyon okunuyor (nested path destegi)
- [ ] YAML dosyadan versiyon okunuyor
- [ ] TOML dosyadan versiyon okunuyor
- [ ] INI dosyadan versiyon okunuyor
- [ ] Text dosyadan versiyon okunuyor (consumeWholeFile)
- [ ] JSON dosyaya versiyon yaziliyor
- [ ] YAML dosyaya versiyon yaziliyor
- [ ] TOML dosyaya versiyon yaziliyor
- [ ] INI dosyaya versiyon yaziliyor
- [ ] Text dosyaya versiyon yaziliyor
- [ ] Glob pattern destegi calisiyor
- [ ] Prefix destegi calisiyor (e.g., "^1.2.3")
- [ ] Birden fazla out dosyasi ayni anda guncelleniyor
- [ ] Dry-run modunda dosya yazilmiyor

### CalVer
- [ ] `yy.mm.minor` formati calisiyor
- [ ] `yyyy.mm.minor` formati calisiyor
- [ ] Takvim degistiginde minor sifirlanir
- [ ] Ayni takvim doneminde minor artirilir
- [ ] Ilk versiyon dogru olusturulur (current yokken)
- [ ] CalVer + SemVer birlikte kullanilmiyor (hata)

### Pre-Release
- [ ] premajor increment calisiyor
- [ ] preminor increment calisiyor
- [ ] prepatch increment calisiyor
- [ ] prerelease increment calisiyor (ayni ID: numara artirma)
- [ ] prerelease increment calisiyor (farkli ID: yeni seri)
- [ ] GitHub preRelease otomatik algilama
- [ ] tagExclude ile pre-release filtreleme

### CLI Modlari
- [ ] --no-increment modu calisiyor
- [ ] --only-version modu calisiyor
- [ ] --release-version sadece versiyon yazdiriyor
- [ ] --changelog sadece changelog yazdiriyor
- [ ] Verbose modu (-V) hook/git detayi gosteriyor
- [ ] Debug modu (-VV) tum internal detayi gosteriyor

- [ ] `go test ./internal/bumper/... ./internal/version/... -race` basarili
- [ ] Test coverage %70+

---

## 9. Test Senaryolari

### Bumper Tests
- JSON okuma: `{"version": "1.2.3"}` -> "1.2.3"
- JSON okuma (nested): `{"tool": {"version": "1.2.3"}}` -> "1.2.3"
- YAML okuma: `version: 1.2.3` -> "1.2.3"
- TOML okuma: `[tool.poetry]\nversion = "1.2.3"` -> "1.2.3"
- INI okuma: `[metadata]\nversion = 1.2.3` -> "1.2.3"
- Text okuma: `1.2.3\n` -> "1.2.3"
- JSON yazma: version alani guncellendi
- YAML yazma: version alani guncellendi, diger alanlar korundu
- Glob pattern: `*.json` -> birden fazla dosya
- Prefix: `"^"` -> `"^1.3.0"`
- consumeWholeFile: tum dosya = versiyon
- Dosya bulunamadi -> hata
- Gecersiz format -> hata

### CalVer Tests
- `yy.mm.minor`: 26.1.0 (Subat'ta) -> 26.2.0
- `yy.mm.minor`: 26.2.0 (Subat'ta) -> 26.2.1
- `yyyy.mm.minor`: 2026.2.0 -> 2026.2.1
- Yil degisimi: 25.12.5 (Ocak 2026) -> 26.1.0
- Ilk versiyon: "" -> 26.2.0

### Pre-Release Tests
- 1.2.3 + premajor("beta") -> 2.0.0-beta.0
- 1.2.3 + preminor("alpha") -> 1.3.0-alpha.0
- 1.2.3 + prepatch("rc") -> 1.2.4-rc.0
- 2.0.0-beta.0 + prerelease("beta") -> 2.0.0-beta.1
- 2.0.0-beta.1 + prerelease("rc") -> 2.0.0-rc.0
