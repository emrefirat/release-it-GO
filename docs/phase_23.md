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
| 5 | `BREAKING CHANGE:` footer'ları işlenmiyor (git log yalnız `%s` çekiyor) → major bump çalışmıyor | KRİTİK | ⬜ |
| 6 | `before:release` / `after:release` hook'ları hiç tetiklenmiyor | YÜKSEK | ⬜ |
| 7 | Config `ci`/`dry-run`/`verbose` alanları unset flag'lerce eziliyor (`Flags().Changed` yok) | YÜKSEK | ⬜ |
| 8 | Webhook gizli URL'leri hata mesajlarına sızıyor (`*url.Error`) | YÜKSEK | ⬜ |
| 9 | npm compat `requireBranch: [...]` → `"main,master"` stringi hiçbir dala eşleşmiyor | YÜKSEK | ⬜ |
| 10 | Go toolchain 1.26.3 → ≥1.26.6 (7 stdlib zafiyeti; `make check` vuln adımında kırılıyor) | YÜKSEK | ⬜ |

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
