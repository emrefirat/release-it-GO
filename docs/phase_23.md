# Faz 23: Kritik Doğruluk ve Güvenlik Yamaları (Denetim P0)

## Özet

2026-08-25 kapsamlı kod denetiminde (4 paralel inceleme: pipeline çekirdeği, HTTP/güvenlik, domain mantığı + npm paritesi, test/CI/build) tespit edilen **kritik** bulguların düzeltilmesi. Ortak tema: *gönderildiği haliyle varsayılanlar* hiç test edilmemişti — projenin kendi config'i GitHub release yolunu kullanmadığı için (`github.release: false`) iki "kutudan çıktığı haliyle bozuk" varsayılan fark edilmemişti.

## Kapsam ve Durum

| # | Bulgu | Önem | Durum |
|---|-------|------|-------|
| 1 | `hooks install` config'den çıkarılan hook'u temizlemiyor (kullanıcı raporu) | P0 | ✅ `1141847` |
| 2 | Varsayılan `github.host: "api.github.com"` → Enterprise çözümlemesi → tüm GitHub API 404 | KRİTİK | ✅ Bu faz |
| 3 | GitLab `secure` sıfır değeri → `InsecureSkipVerify=true` (varsayılanda TLS doğrulaması kapalı) | KRİTİK | ✅ Bu faz |
| 4 | Geçersiz CA dosyası boş root pool kuruyor (tüm bağlantılar opak x509 hatasıyla düşer) | ORTA (3'e bağlı) | ✅ Bu faz |
| 5 | `BREAKING CHANGE:` footer'ları işlenmiyor (git log yalnız `%s` çekiyor) → major bump çalışmıyor | KRİTİK | ✅ Bu faz |
| 6 | `before:release` / `after:release` hook'ları hiç tetiklenmiyor | YÜKSEK | ✅ Bu faz |
| 7 | Config `ci`/`dry-run`/`verbose` alanları unset flag'lerce eziliyor (`Flags().Changed` yok) | YÜKSEK | ✅ Bu faz |
| 8 | Webhook gizli URL'leri hata mesajlarına sızıyor (`*url.Error`) | YÜKSEK | ✅ Bu faz |
| 9 | npm compat `requireBranch: [...]` → `"main,master"` stringi hiçbir dala eşleşmiyor | YÜKSEK | ✅ Bu faz |
| 10 | Go toolchain 1.26.3 → 1.26.7 (7 stdlib zafiyeti; `make check` vuln adımında kırılıyordu) | YÜKSEK | ✅ Bu faz |

Bumper format kaybı (KRİTİK ama büyük iş) Faz 26'ya; dağıtım/CI bulguları (go install imkânsız, CI security job, CHANGELOG bayatlığı, release.yml orphan-tag yarışı) Faz 24'e planlandı.

## Bu Fazda Yapılan Değişiklikler (madde 2-4)

### GitHub host çözümlemesi

```go
// internal/release/github.go — "api.github.com" artık public API alias'ı
if host == "" || host == "github.com" || host == "api.github.com" {
    return "https://api.github.com"
}
```

- `internal/config/defaults.go`: `Host: "api.github.com"` → `"github.com"`
- Alias, `host: "api.github.com"` yazılı mevcut kullanıcı config'lerini de (kendi eski full-example çıktımız dahil) düzeltir.

### GitLab TLS varsayılanı

- `internal/config/defaults.go`: `Secure: true` eklendi — sıfır değer artık güvenli.
- `secure: false` **açık opt-out** olarak korunuyor (self-signed, CA'sız instance'lar).
- Davranış değişikliği opt-out'tur (Faz 22 `--atomic` modeliyle aynı): TROUBLESHOOTING'e `x509: certificate signed by unknown authority` girdisi eklendi.
- Geçersiz PEM içeren CA dosyası artık boş pool kurmuyor; uyarı + sistem köklerine düşüş.

### BREAKING CHANGE footer işleme (madde 5)

- `internal/git/changelog.go`: yeni `GetFullCommitsSinceTag` — `--pretty=format:%h%x1f%B%x1e` (ASCII unit/record separator) ile hash + **tam mesaj** çekiyor; `%s` yalnız konu satırını getirdiğinden gövdedeki footer'lar parser'a hiç ulaşmıyordu.
- `internal/runner/runner.go`: yeni `commitsSinceLatestRelease()` helper'ı — üç çağrı noktası (`RunChangelogOnly`, `autoDetectIncrement`, `generateChangelog`) tek kaynağa indi; raw-tag fallback'i artık **hepsinde tutarlı** (önceden yalnız changelog'da vardı: tag format geçişinde auto-increment sessizce hep "patch" dönüyordu — denetim M9 bunu da kapattı).
- `internal/changelog/parser.go`: çok satırlı footer değerleri (git trailer devam satırları) artık destekleniyor — önceden tek devam satırı TÜM footer bölümünü düşürüyordu (denetim A14; A1 düzeltilince tetiklenecek gizli bug).
- Yan kazanım: commit hash'leri artık pipeline'da taşındığından changelog girdilerinde `(abc1234)` hash/link ilk kez gerçekten render ediliyor (denetim A10).
- Ölü kod temizliği: üretimde kullanımı kalmayan `GetCommitsSinceTag` ve 3 testi silindi.
- Uçtan uca kanıt: `feat: overhaul API` + `BREAKING CHANGE:` gövdeli commit → `1.0.0 → 2.0.0`, changelog'da `### BREAKING CHANGES` bölümü + hash.

### Regresyon testleri (gönderildiği haliyle varsayılanlar)

- `TestNewGitHubClient_DefaultConfig_UsesPublicAPIBaseURL` — `DefaultConfig()`'ten kurulan istemcinin baseURL'i `https://api.github.com` olmalı.
- `TestGitLabClient_CreateHTTPClient_DefaultConfig_VerifiesTLS` — `DefaultConfig()` ile `InsecureSkipVerify=false`.
- `TestGitLabClient_CreateHTTPClient_ValidCAPEM_InstallsRootPool` — gerçek (üretilmiş self-signed) sertifikayla RootCAs kurulumu; önceki testler geçersiz PEM kullandığından bu yol hiç test edilmemişti.
- `TestGitLabClient_CreateHTTPClient_InvalidCAPEM_FallsBackToSystemRoots`.

## Hedefler

- Varsayılan config'le GitHub release uçtan uca çalışır (404 yok).
- Varsayılan config'de token asla doğrulanmamış TLS üzerinden gitmez.
- Mevcut açık kullanıcı tercihleri (`secure: false`, GHE `host`) davranış değiştirmez.
- Her düzeltme test-first (kırmızı → yeşil) ilerler; "shipped defaults" regresyon testleri kalıcıdır.

## Kapsam Dışı

- GHE asset upload (`upload_url` kullanılmıyor) — denetim F5, Faz 25/26.
- HTTP retry/backoff ve ortak HTTP helper — Faz 26.
- `github.web`, `gitlab.preRelease` gibi ölü config alanları — Faz 25.

## Tamamlanma Notları (madde 6-10)

- **Release hook'ları**: üç pipeline döngüsü ortak `runSteps()` helper'ına toplandı (denetimin mimari önerisi); `before:release` git:release adımından (kendi before hook'undan da) önce, `after:release` tüm adımlar bittikten sonra tetikleniyor. "No commits" graceful çıkışında release hook'ları bilinçli olarak atlanıyor. Entegrasyon testleri sıralamayı ve atlama davranışını kanıtlıyor.
- **Flag önceliği**: `buildFlagOverrides(cmd)` yalnızca `Flags().Changed()` olan bool/count flag'leri için pointer geçiriyor; config dosyasındaki `ci`/`dry-run`/`verbose` artık korunuyor.
- **Webhook sızıntısı**: `sendOne` `*url.Error`'ı açıp yalnızca alt nedeni raporluyor; hata metni tip + urlRef adı taşıyor, URL asla. Regresyon testi "SECRET asla hata metninde görünmez" assert'i içeriyor.
- **requireBranch**: virgülle ayrılmış çoklu desen, herhangi biri eşleşirse geçer; `path.Match` ile platform-bağımsız glob (denetim L5 de kapandı).
- **Toolchain**: go.mod `go 1.26.7`, Dockerfile `golang:1.26.7-alpine`; CI `go-version-file: go.mod` kullandığından otomatik uyum. `make check` vuln adımı ilk kez yeşil.
