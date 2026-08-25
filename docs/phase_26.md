# Faz 26: Sağlamlaştırma (Denetim Orta/Yapısal Bulgular)

## Özet

2026-08-25 denetiminin orta öncelikli ve yapısal bulgularının kapatılması: hata sözleşmeleri, eşzamanlılık, platform bağımsızlığı, veri güvenliği (bumper) ve HTTP dayanıklılığı.

## Kapsam ve Durum

| # | İş | Denetim ref | Durum |
|---|----|-------------|-------|
| 1 | `git.ErrNoCommits` sentinel hatası (`errors.Is`) — paketler arası string eşleştirme kalksın | M7 | ⬜ |
| 2 | Git exec sarmalayıcılarının birleşmesi: stderr + kök hata birlikte (`%s: %s: %w`) | M6 | ⬜ |
| 3 | Spinner data race (`s.done` kilitsiz okuma; Stop→Start yarışı) | M8/A18 | ⬜ |
| 4 | Windows lifecycle hook desteği (`sh -c` → `cmd /C` seçimi) + hook değişkenlerinin env olarak da ihracı | H4/M2 | ⬜ |
| 5 | Tag glob eşleştirme: `matchGlob` → gerçek glob (`path.Match`) + pre-release'i stable üstüne koyan `--sort=-v:refname` yerine semver sıralaması | M11/A16/A17 | ⬜ |
| 6 | Hedefli staging: release commit'i yalnız bumper/changelog dosyalarını içersin (`git add .` süpürmesi kalksın) | M10 | ⬜ |
| 7 | Ortak HTTP retry helper'ı (exponential backoff, `Retry-After`, 429/5xx/bağlantı hataları) — GitHub/GitLab/notification | F11 | ⬜ |
| 8 | Format koruyan bumper: doğrula → hedefli metin değişimi → yeniden parse ile doğrula; başarısızsa mevcut tam-yazım + uyarı | A2 | ⬜ |
| 9 | Push başarısızlığında kurtarma yönergesi (oluşan commit/tag adlarıyla) — npm de rollback yapmıyor; `--no-increment` akışı (Faz 25.4) kurtarma yoludur | H2 | ⬜ |

## Kapsam Dışı / Ertelenen

- bubbletea v2 / lipgloss v2 migrasyonu — ayrı değerlendirme.
- GHE asset upload (`upload_url` kullanımı, F5) ve `ReleaseProvider` arayüz revizyonu — API kırılımı gerektirir, ayrı faz.
- Ölü config alanları envanteri (A15/F18) — bağla-veya-sil kararları ürün kararı gerektirir.

## Prensipler

- Küçük, bağımsız commit'ler; her biri test-first.
- Davranış değişiklikleri opt-out veya davranış-koruyan (bumper'da otomatik fallback).
- `make check` her commit öncesi uçtan uca yeşil.
