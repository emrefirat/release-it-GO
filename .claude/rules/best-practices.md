# Go Best Practices ve Proje Ozel Kurallar

## Dependency Management

- Yeni dependency eklemeden once stdlib ile cozulup cozulemeyecegini degerlendir.
- `go mod tidy` her dependency degisikliginden sonra calistir.
- `go mod verify` ile checksum dogrula.
- Indirect dependency sayisini minimize et.

## Interface Tasarimi

- Interface'leri kullanan tarafta tanimla (consumer side), uretici tarafta degil.
- Kucuk interface'ler tercih et (1-3 metod). Buyuk interface'ler bolunmeli.
- `io.Reader`, `io.Writer` gibi stdlib interface'lerini kullan mumkunse.

```go
// DOGRU - Kucuk, odakli interface
type Prompter interface {
    Confirm(msg string, def bool) (bool, error)
    SelectVersion(current, recommended string, options []VersionOption) (string, error)
}

// YANLIS - God interface
type Everything interface {
    // 20+ metod...
}
```

## Context ve Cancellation

- Uzun suren islemlerde `context.Context` kabul et.
- Goroutine baslatirken her zaman context ile iptal mekanizmasi kur.
- HTTP client'larda timeout belirle.

## Concurrency

- Channel'i mutex'a tercih et (mumkunse).
- Goroutine leak'i onlemek icin `defer cancel()` pattern'i kullan.
- `sync.WaitGroup` ile goroutine yasam dongusunu yonet.
- `go test -race ./...` ile race condition kontrol et.

## API ve HTTP Client Kurallari

- HTTP client'larda timeout zorunlu (default: 30s).
- Retry mantigi icin exponential backoff kullan.
- Response body'yi her zaman `defer resp.Body.Close()` ile kapat.
- TLS sertifika dogrulama atlama (InsecureSkipVerify) sadece test ortaminda.

## Config Yonetimi

- Default degerler `config/defaults.go`'da merkezi tanimlanir.
- Config struct'lari `config/config.go`'da.
- Yeni config alani eklerken: struct + default + JSON/YAML tag + test.
- Config dosya formatlari: JSON, YAML, TOML (Viper ile auto-detect).

## CLI Flag Kurallari

- Yeni flag eklerken `internal/cli/root.go`'da tanimla.
- Flag isimleri kebab-case: `--pre-release`, `--dry-run`.
- Boolean flag'ler icin `--no-` prefix'i ile disable: `--no-git.push`.
- Her flag icin aciklayici `Usage` string'i zorunlu.

## Pipeline Adimi Ekleme

Yeni pipeline adimi eklemek icin:
1. `runner.go`'da `pipelineStep` struct'ina ekle
2. Adim fonksiyonunu implement et (spinner + error handling)
3. Before/after hook destegi otomatik gelir
4. Dry-run destegi zorunlu
5. Integration test ekle

## Changelog ve Commit Analizi

- Commit parse: `changelog/parser.go` (Angular preset)
- Bump analiz: `changelog/analyzer.go` (feat→minor, fix→patch, !→major)
- Yeni commit type eklerken `allowedTypes` map'ini guncelle

## Docker Best Practices

- Multi-stage build kullan (builder + runtime).
- Runtime image'da sadece gerekli dependency'ler (git, ca-certificates).
- Binary her zaman static derle: `CGO_ENABLED=0`.
- Non-root user ile calistir.
- `.dockerignore` guncel tut.

## Dosya Islemleri

- Dosya yollari icin `filepath.Join` kullan (platform-independent).
- Gecici dosyalar icin `os.CreateTemp` veya `t.TempDir()` (testlerde).
- Dosya izinleri: 0644 (dosya), 0755 (dizin, executable).
- Dosya acildiktan sonra `defer f.Close()`.

## Hata Mesajlari

- Kullaniciya yonelik mesajlar aciklayici ve aksiyon onerir olmali.
- Internal error detail'leri kullaniciya gosterme.
- Verbose modda (-v) daha fazla detay goster.
- Error zincirinde `%w` ile wrap et, root cause'a ulasilabilsin.
