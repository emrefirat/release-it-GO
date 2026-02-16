# release-it-go - Gelistirici Kurallari ve Proje Rehberi

> Bu dosya, projede calisan tum gelistiricilerin uymas gereken kurallari ve proje mimarisini icerir.

---

## 0. Proje Tanimi

**release-it-go**, [release-it](https://github.com/release-it/release-it) npm aracinin Go ile yeniden yazilmis halidir. npm/Node.js bagimliligi olmadan Git tag + GitHub/GitLab release + changelog olusturma islemlerini yapar.

### Mimari Kararlar

| Karar | Secim |
|-------|-------|
| Uygulama adi | `release-it-go` |
| npm/paket yayinlama | Yok - sadece Git tag + Release |
| Plugin sistemi | Harici plugin yok, hepsi built-in |
| Config formatlari | JSON + YAML + TOML |
| Built-in ozellikler | conventional-changelog, keep-a-changelog, bumper, calver |
| GitHub/GitLab | Ikisi de destekleniyor |
| Terminal UI | Renkli interaktif (bubbletea) |
| Dagitim | GitHub Releases + `go install` |
| Versiyon kaynagi | Git tag (birincil) + VERSION dosyasi (ikincil) |
| Conventional commits | Tam destek (Angular preset) |
| Go minimum | 1.22+ |

### Proje Yapisi

```
release-it-go/
  cmd/release-it-go/main.go     # Uygulama giris noktasi
  internal/
    cli/                         # Cobra CLI
    config/                      # Config yuklenme ve struct'lar
    version/                     # Semver + CalVer
    git/                         # Git islemleri
    changelog/                   # Conventional commits + changelog
    release/                     # GitHub + GitLab API client
    bumper/                      # Coklu dosya versiyon guncelleme
    hook/                        # Lifecycle hooks
    ui/                          # Terminal UI (prompt, spinner, colors)
    runner/                      # Pipeline orchestrator
    log/                         # Logger
  test/integration/              # Integration testler
  docs/                          # Faz PRD dosyalari
  PROGRESS.md                    # Ilerleme takibi
```

### Gelistirme Fazlari

Proje 7 faz halinde gelistirilir. Detaylar `docs/phase_*.md` dosyalarinda:

1. **Core Foundation** - CLI, config, versiyon tespiti
2. **Git Operations** - Prerequisite checks, commit, tag, push
3. **Conventional Commits + Changelog** - Parser, analyzer, CHANGELOG.md
4. **GitHub + GitLab Releases** - API entegrasyonu, asset upload
5. **Interactive UI + Hooks + Pipeline** - Prompts, hooks, runner
6. **Advanced Features** - Bumper, calver, pre-release flows
7. **Testing, CI/CD, Documentation** - Integration test, GoReleaser

### PROGRESS.md Zorunlulugu

- **Her gelistirme oturumu sonunda `PROGRESS.md` dosyasini guncelle.**
- Tamamlanan maddeleri `[x]` ile isaretle.
- Ilerleme yuzdesini guncelle.
- Notlar bolumune onemli kararlar ve engelleri yaz.

---

## 1. Gelistirici Kimligi

- **Sen bir Senior Go Developer'sin.** Bu seviyede kod kalitesi, mimari kararlar ve best practice'ler bekleniyor.
- Kod yazarken "calisiyor" yeterli degil, "dogru calisiyor, guvenli, test edilmis ve bakimi kolay" olmali.
- Her zaman production-ready kod yaz.

---

## 2. Git ve Versiyon Kontrolu

### Commit Kurallari

- **Her basarili gelistirme veya fix sonrasi MUTLAKA commit at.**
- Commit mesajlari aciklayici ve Conventional Commits formatinda olmali:

```
feat: Yeni ozellik eklendi
fix: Bug duzeltildi
refactor: Kod yeniden duzenlendi
test: Test eklendi/guncellendi
docs: Dokumantasyon guncellendi
chore: Bakim isleri
perf: Performans iyilestirmesi
```

- Her commit tek bir amaca hizmet etmeli (atomic commits).
- Commit atmadan once `go build` ve testlerin basarili oldugundan emin ol.

### Branch Stratejisi

- `main` branch her zaman calisir durumda olmali.
- Buyuk ozellikler icin feature branch kullan.
- PR acmadan once testlerin gectiginden emin ol.

---

## 3. Test Gereksinimleri

### Unit Test Zorunlulugu

- **Her yeni fonksiyon icin unit test ZORUNLU.**
- Test coverage minimum %70 olmali (%80 hedef).
- Test dosyalari `*_test.go` formatinda olmali.

### Test Yazim Kurallari

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Arrange - Test verilerini hazirla
    // Act - Test edilecek fonksiyonu cagir
    // Assert - Sonuclari dogrula
}
```

### Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got = %v, expected %v", got, tt.expected)
            }
        })
    }
}
```

### Test Calistirma

```bash
go test ./... -v
go test ./... -cover
go test ./... -race
```

- Testler basarisiz olursa commit ATMA, once duzelt.

---

## 4. Kod Kalitesi

### Go Best Practices

```bash
go fmt ./...
go vet ./...
golangci-lint run
```

### Error Handling

```go
// YANLIS - Error'u ignore etme
result, _ := someFunction()

// DOGRU - Her error'u handle et
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}
```

### Naming Conventions

| Tur | Format | Ornek |
|-----|--------|-------|
| Package | kucuk harf, tek kelime | `config`, `git`, `release` |
| Exported | PascalCase | `CreateRelease`, `UserID` |
| Unexported | camelCase | `generateSlug`, `validateInput` |
| Constants | PascalCase veya SCREAMING_SNAKE | `MaxRetries`, `DEFAULT_TIMEOUT` |
| Interfaces | -er suffix | `Reader`, `Prompter`, `ReleaseProvider` |
| Acronyms | Tutarli buyuk/kucuk | `HTTPServer`, `userID` |

### Kod Organizasyonu

- Dosya basina tek sorumluluk.
- Fonksiyonlar 50 satiri gecmemeli (mumkunse).
- Nested if'lerden kacin, early return kullan.
- Magic number kullanma, constant tanimla.

### Early Return Pattern

```go
// DOGRU - Early return
func process(data *Data) error {
    if data == nil {
        return errors.New("nil data")
    }
    if !data.IsValid() {
        return errors.New("invalid data")
    }
    // actual logic
    return nil
}
```

---

## 5. Guvenlik Kurallari

### ASLA Yapma

- Sifreleri/secret'lari plain text saklama
- Kullanici girdisini dogrulamadan kullanma
- Sensitive bilgileri loglama
- Hardcoded credentials
- Secret'lari koda gomme
- Panic'i error handling yerine kullanma

### Her Zaman Yap

- Input validation uygula
- Environment variable kullan (ozellikle token'lar icin)
- Error mesajlarinda internal detay verme
- Dependency'leri guncel tut

---

## 6. Error Handling

### Error Wrapping

```go
if err != nil {
    return fmt.Errorf("failed to process user %s: %w", userID, err)
}
```

### Custom Error Types

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized")
)
```

### Error Logging

```go
// Sadece error'u olusturan yerde logla, her yerde degil
// En ust seviyede logla (main veya runner)
```

---

## 7. Performance

- Buyuk slice'lar icin capacity belirt: `make([]Item, 0, len(data))`
- Gereksiz allocation'lardan kacin.
- Goroutine leak'lerden kacin - context kullan.
- Race detection: `go test -race ./...`

---

## 8. Logging

- `log/slog` stdlib kullan (Go 1.22+).
- Structured logging tercih et.
- Sensitive veri loglama (sifre, token, PII).
- Log level'i yapilandirabilir yap (normal/verbose/debug).

---

## 9. Yapilmamasi Gerekenler

### Kod
- `panic` kullanma (sadece gercekten kurtarilamaz durumlar icin)
- Global state kullanma
- `init()` fonksiyonlarinda side effect
- Circular dependency olusturma
- God objects/functions
- Deep nesting (3+ seviye)
- Error'lari ignore etme (`_`)

### Surec
- Test yazmadan commit atma
- Review almadan main'e push
- Breaking change'i dokumante etmeden yapma
- PROGRESS.md guncellemeden oturumu bitirme

---

## 10. Bug Takibi

- Kullanicidan gelen veya tespit edilen her bug `PROGRESS.md` dosyasindaki **Bugs** bolumune eklenir.
- Format: `- [ ] BUG: Kisa aciklama (tarih)` (acik) veya `- [x] BUG: Kisa aciklama (tarih) → cozum ozeti` (kapali)
- Bug fix commit'i atildiktan sonra ilgili satir `[x]` ile isaretlenir ve cozum ozeti eklenir.
- Edge case'ler ve ilk release gibi ozel senaryolar mutlaka test edilmeli.

---

## 11. Her Commit Oncesi Checklist

```
[ ] Kod formatlandi (go fmt)
[ ] Static analysis gecti (go vet)
[ ] Lint hatalari yok (golangci-lint)
[ ] Testler yazildi
[ ] Testler geciyor (go test)
[ ] Race condition yok (go test -race)
[ ] Build basarili (go build)
[ ] Commit mesaji anlamli ve conventional
[ ] PROGRESS.md guncellendi (gerekiyorsa)
```

---

## 12. Kullanilacak Araclar

```bash
go fmt ./...
go vet ./...
golangci-lint run
go test ./... -v -cover -race
go build -o bin/release-it-go ./cmd/release-it-go
go mod tidy
go mod verify
```

---

**Bu kurallar tartismaya acik degil, zorunludur.**
