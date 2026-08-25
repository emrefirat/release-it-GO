# Faz 26: Sağlamlaştırma (Denetim Orta/Yapısal Bulgular)

## Özet

2026-08-25 denetiminin orta öncelikli ve yapısal bulgularının kapatılması: hata sözleşmeleri, eşzamanlılık, platform bağımsızlığı, veri güvenliği (bumper) ve HTTP dayanıklılığı.

## Kapsam ve Durum

| # | İş | Denetim ref | Durum |
|---|----|-------------|-------|
| 1 | `git.ErrNoCommits` sentinel hatası (`errors.Is`) — paketler arası string eşleştirme kalksın | M7 | ✅ |
| 2 | Git exec sarmalayıcılarının birleşmesi: stderr + kök hata birlikte (`%s: %s: %w`) | M6 | ✅ |
| 3 | Spinner data race (`s.done` kilitsiz okuma; Stop→Start yarışı) | M8/A18 | ✅ |
| 4 | Windows lifecycle hook desteği (`sh -c` → `cmd /C` seçimi) + hook değişkenlerinin env olarak da ihracı | H4/M2 | ✅ |
| 5 | Tag glob eşleştirme: `matchGlob` → gerçek glob (`path.Match`) + pre-release'i stable üstüne koyan `--sort=-v:refname` yerine semver sıralaması | M11/A16/A17 | ✅ |
| 6 | Hedefli staging: release commit'i yalnız bumper/changelog dosyalarını içersin (`git add .` süpürmesi kalksın) | M10 | ✅ |
| 7 | Ortak HTTP retry helper'ı (exponential backoff, `Retry-After`, 429/5xx/bağlantı hataları) — GitHub/GitLab/notification | F11 | ✅ |
| 8 | Format koruyan bumper: doğrula → hedefli metin değişimi → yeniden parse ile doğrula; başarısızsa mevcut tam-yazım + uyarı | A2 | ✅ |
| 9 | Push başarısızlığında kurtarma yönergesi (oluşan commit/tag adlarıyla) — npm de rollback yapmıyor; `--no-increment` akışı (Faz 25.4) kurtarma yoludur | H2 | ✅ |

## Kapsam Dışı / Ertelenen

- bubbletea v2 / lipgloss v2 migrasyonu — ayrı değerlendirme.
- GHE asset upload (`upload_url` kullanımı, F5) ve `ReleaseProvider` arayüz revizyonu — API kırılımı gerektirir, ayrı faz.
- Ölü config alanları envanteri (A15/F18) — bağla-veya-sil kararları ürün kararı gerektirir.

## Prensipler

- Küçük, bağımsız commit'ler; her biri test-first.
- Davranış değişiklikleri opt-out veya davranış-koruyan (bumper'da otomatik fallback).
- `make check` her commit öncesi uçtan uca yeşil.

## Tamamlanma Notları (2026-08-26)

- **1-2**: `git.ErrNoCommits` (`errors.Is`); `run()` artık `cmd: stderr: %w`, `runSilent()` `stderr: %w` — iki sarmalayıcı da hem kullanıcı metnini hem kök hatayı taşıyor. Runner'daki string eşleştirme kalktı (mock'lardaki sahte-hata tuzağı da temizlendi).
- **3**: Spinner animatörü kendi `done` kanalını closure'la yakalıyor; `Start` önceki animatörü kapatıyor; ticker'lı select `Stop`'ta anında çıkıyor. `-race` kanıt testi eklendi.
- **4**: `shellCommandFor(GOOS, COMSPEC, cmd)` — Windows'ta `cmd.exe /C`; hook env'ine `RELEASE_*` değişkenleri (envKey: camelCase/dotted → SNAKE) eklendi.
- **5**: `matchGlob` → `path.Match` (`*` çıplak deseni legacy her-şey anlamını koruyor); `getLatestTagFromAllRefs` semver karşılaştırmasıyla en yükseği seçiyor (parse edilemeyenlerde git sırasına düşüş).
- **6**: `WriteVersionFiles` değişen dosyaları döndürüyor; `stageReleaseFiles` yalnız bunları + changelog'u stage'liyor. `addUntrackedFiles` eski tam süpürmeyi koruyor. Entegrasyon testi: alakasız yerel düzenleme release commit'ine girmiyor ve working tree'de kalıyor.
- **7**: `internal/httputil.Do` — 429/502/503/504'te exponential backoff (Retry-After öncelikli), bağlantı hatasında yalnız idempotent metodlarda replay (POST'un akıbeti bilinmeden tekrarı duplicate release riski). GitHub/GitLab `doRequest` ve notification `sendOne` bu yoldan geçiyor; upload akışları (dosya gövdesi replay edilemez) bilinçli kapsam dışı.
- **8**: `updateVersionTargeted` — hedefli metin değişimi + yeniden parse + tam ağaç karşılaştırmasıyla İSPAT; belirsiz/ispatlanamayan durumda eski tam-marshal fallback'i (değer yine güncellenir). JSON key sırası/girinti/`&`, YAML yorumları, TOML düzeni artık release diff'inde bozulmuyor. Nested aynı-isimli anahtar testi ağaç karşılaştırmasının yanlış konumu elediğini kanıtlıyor.
- **9**: Push hatası artık `--no-increment` kurtarma akışını adıyla öneriyor (Faz 25.4 o akışı çalışır kılmıştı).
