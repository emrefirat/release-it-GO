# Faz 32: Üçüncü Denetim — Düşük Seviye Temizlik

**Durum:** ✅ Tamamlandı — 2026-11-04

## Özet

Faz 31 denetiminin düşük seviye bulguları (10 madde + npm `infile: false` uyumu). Her madde kod okumasıyla (INI ve JSON-escape ayrıca çalıştırılarak) doğrulandı, 7 commit'te regresyon testleriyle düzeltildi.

| # | Bulgu | Yer | Durum |
|---|-------|-----|-------|
| 1 | INI writer satırı yeniden kuruyor: inline yorum, boşluk ve `\r` kaybı; reader yorumu değere katıyor | `bumper/writer.go`, `reader.go` | ✅ yalnız değer aralığı değişiyor (`splitINIValue`) |
| 2 | `Revert "feat: x"` commit'leri parse edilmiyor (bump yok, Reverts bölümü boş) | `changelog/parser.go` | ✅ `revert` tipi |
| 3 | `hooks install` idempotent değil (`# Generated: <tarih>`); 0755 yeniden uygulanmıyor | `githook/githook.go` | ✅ tarih kaldırıldı, `Chmod` |
| 4 | `version` fallback `v0.0.0-<ts>-<hash>` (tire) formunu kaçırıyor | `cli/buildinfo.go` | ✅ `[.-]` |
| 5 | `RELEASE_CHANGELOG` env limiti (Linux 128 KiB) tüm hook'ları düşürüyor | `hook/hook.go` | ✅ 96 KiB cap + marker |
| 6 | `version.go` ölü kod, bozuk glob | `version/version.go` | ✅ silindi (+ testleri) |
| 7 | Proxy URL şifresi uyarı log'unda | `release/github.go` | ✅ değer loglanmıyor |
| 8 | GitLab `PostComment` `"pr"` → issues | `release/gitlab.go` | ✅ merge_requests (kod korundu, ileride comments için) |
| 9 | JSON fallback HTML-escape (`\u0026`) | `bumper/writer.go` | ✅ `SetEscapeHTML(false)`; anahtar sıralaması map kaynaklı, fallback nadir |
| 10 | Hook `${...}` render'ı map sırasına bağlı; `${changelog}` shell'e tırnaksız giriyor | `hook/hook.go` | ✅ tek geçişli regexp; README'de env var önerisi. Injection tasarımsal (npm ile aynı), tamamen kapatılamaz |
| 11 | npm `conventional-changelog` `infile: false` düşüyor, CHANGELOG.md yazılıyor | `config/compat.go` | ✅ `Infile = ""` |

**Yapılmayan:** CalVer yerel saat (`time.Now()`), mevcut tag'lerle uyum için değiştirilmedi; UTC istenirse ayrı karar.
