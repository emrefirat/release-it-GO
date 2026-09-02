# Faz 27: Kararlılık Düzeltmeleri (İkinci Denetim, 2026-09-02)

## Özet

Faz 23–26 sonrası yapılan ikinci tur denetimin (4 paralel inceleme: pipeline/git kararlılığı, parametre geçerliliği, HTTP/bumper/changelog, doküman/test/CI) **kararlılık** bulgularının kapatılması. Ortak tema: ana yol (`Run()`, varsayılan config, `${version}`/`v${version}` şablonu) sağlam; ikincil giriş noktaları (`--only-version`, `--no-increment`), şablon çeşitliliği ve bazı Faz 25/26 değişikliklerinin eksik kalan yarıları zayıf.

## Kapsam ve Durum

| # | Bulgu | Önem | Kanıt | Durum |
|---|-------|------|-------|-------|
| 1 | Bumper, JSON/YAML/TOML/INI dışı her dosyayı (README.md dahil) tek satır sürüme indirgeyip commit'liyor; `consumeWholeFile: false` no-op | KRİTİK | `bumper/writer.go` `FormatText` dalı; canlı: README → `1.1.0` | ✅ `8894f68` |
| 2 | `--no-increment` kurtarma akışı varsayılan changelog'la kırık: boş duplicate bölüm + ikinci release commit + "tag on different commit" | KRİTİK | canlı doğrulandı; Faz 25.4 testi changelog'u kapatarak yazılmış | ✅ `fcdca5a` |
| 3 | `--only-version`/`--no-increment` prerequisites, commitlint ve token doğrulamasını atlıyor → eksik token commit+tag+push SONRASI fark ediliyor | KRİTİK | `runner.go` RunOnlyVersion/RunNoIncrement adım listeleri | ✅ `fcdca5a` |
| 4 | Test tuzağı: `cli/root_test.go`'daki 6 `Execute --ci` testi ana repoda gerçek release yapabiliyor (dry-run yok, hata yutuluyor) | KRİTİK (test hijyeni) | tracked `internal/cli/CHANGELOG.md` bu yoldan gelmiş | ✅ `1105592` |
| 5 | `${version}`/`v${version}` dışı tag şablonları (`release-${version}`) ikinci release'te "invalid version" | YÜKSEK | `ParseVersion("release-1.0.0")` hata; runner soymuyor | ✅ `2df6ee5` |
| 6 | `getLatestTagFromAllRefs: true` + v'li repo + default şablon → "tag yok" sanılıp `0.1.0` atılıyor (çıkarım yalnız describe yolunda) | YÜKSEK | `tag.go` GetLatestTag erken dönüş | ✅ `2df6ee5` |
| 7 | Açık `tagMatch` eşleşmeyince "süreklilik" fallback'i dışlanan tag'i döndürüyor | ORTA | `tag.go:53-61` | ✅ `2df6ee5` |
| 8 | `-i prerelease/prepatch/…` + `--preRelease` → `preprerelease` hatası (menünün kendi sunduğu anahtarlar) | YÜKSEK | `determineSemVer` `"pre"+increment` koşulsuz | ✅ `2df6ee5` |
| 9 | Açık sürüm ≤ mevcut sürüm sessizce kabul (`release-it-go 0.5.0` 1.0.0 üstüne) | ORTA | `determineSemVer` explicit dalı | ✅ `2df6ee5` |
| 10 | `bumper.in` sürümü + tag yok → changelog adımı "ambiguous argument" ile düşüyor | YÜKSEK | `commitsSinceLatestRelease` fallback GetLatestTag'e bağlı | ✅ `8894f68` |
| 11 | `--only-version`'da `before:bump` hook'u `${version}`'ı boş görüyor (UpdateVars çağrılmıyor) | YÜKSEK | RunOnlyVersion determineVersion'ı doğrudan çağırıyor | ✅ `fcdca5a` |
| 12 | Retry helper 502/504'te POST'u tekrarlıyor → gateway arkasında duplicate release / 422 | YÜKSEK | `httputil/retry.go` retryableStatuses tüm metodlara | ✅ `b2781a9` |
| 13 | Changelog compare linkleri çıplak sürümden kuruluyor; v-önek çıkarımı sonrası tam o repolarda 404 | YÜKSEK | `conventional.go:21-25`, `Options`'ta tag adı yok | ✅ `b2781a9` |
| 14 | Hedefli staging hook'ların ürettiği tracked değişiklikleri (dist/, lockfile) release commit'inin dışında bırakıyor; npm `git add -u` yapar | YÜKSEK | `stageReleaseFiles` yalnız BumpedFiles | ✅ `1cbca70` |
| 15 | GitHub/GitLab istemcileri `HTTPS_PROXY`/`NO_PROXY` ortam değişkenlerini yok sayıyor (notification sayıyor) | ORTA | `&http.Transport{}` sıfır değeri Proxy=nil | ✅ `b2781a9` |
| 16 | `github.makeLatest: false` no-op (`omitempty` + yalnız "true" gönderiliyor) | ORTA | `github.go:40,135` | ✅ `b2781a9` |
| 17 | Remote parse edilemeyince / origin yokken GitHub/GitLab release sessizce atlanıp "Done" basılıyor; `pushRepo` RepoInfo'da kullanılmıyor | ORTA | `init()` Verbose log, `githubRelease` nil dönüş | ✅ `1cbca70` |
| 18 | `autoDetectIncrement` git hatasını logsuz yutup "patch" diyor | ORTA | `runner.go` autoDetectIncrement | ✅ `1cbca70` |
| 19 | `--changelog` / `--release-version` TTY'de interaktif sürüm prompt'u açıyor (script modları) | ORTA | `root.go` CI set etmiyor | ✅ `fcdca5a` |

## QA Test Planı (her madde için)

- **Pozitif + negatif + sınır:** her düzeltme için "doğru davranış", "hata mesajı ve çıkış kodu", "önceki/sonraki repo durumu" (tag listesi, HEAD, working tree) assert edilir — yalnız `err == nil` yeterli değildir.
- **Kırmızı önce:** her test önce mevcut kodda başarısız olmalı; geçen bir test "tautoloji" şüphesiyle gözden geçirilir.
- **Gerçek repo entegrasyonu:** 1, 2, 3, 5, 6, 10, 11, 14 için `test/integration` senaryoları (mock değil).
- **Mevcut testlerin gözden geçirilmesi:** Faz 25.4 `NoIncrement_ExistingTagAtHead` (changelog kapatarak bug'ı maskeliyor) → varsayılan changelog ile yeniden yazılır; `root_test.go` Execute tuzağı → temp dizin + `--dry-run` + assert; `runner_test` "tag already exists" testi mesajı assert etmeli.
- **Parite referansı:** npm release-it davranışı (Context7) — 9, 12, 14, 16 için.

## Kapsam Dışı

Parametre doğrulama ve ölü alanlar → Faz 28; test altyapısı ve dağıtım → Faz 29; dokümanlar → Faz 30.

## Tamamlanma Notları (2026-09-02)

- **Test tuzağı (4)** ilk kapatıldı: altı `Execute` testi geçici repoda gerçek assert'lerle yeniden yazıldı; pozisyonel arg, `--changelog`, `--release-version` cobra üzerinden uçtan uca test ediliyor.
- **Bumper (1, 10)**: `WriteVersionToFileFrom`/`WriteVersionFiles(current, new)` — metin hedefleri yerinde düzenleniyor; mevcut sürüm bilinmiyor/yoksa commit'ten ÖNCE hata (`consumeWholeFile` opt-in). Eski birim testi bug'ı "beklenen davranış" olarak kodluyordu, yeniden yazıldı. Tag'siz repoda `bumper.in` sürümü → tüm commit'lere düşüş.
- **Giriş noktaları (2, 3, 11, 19)**: `runPipeline(versionStep)` — `Run`/`RunOnlyVersion`/`RunNoIncrement` aynı adım listesinden geçiyor (prerequisites, commitlint, token). Boş changelog bölümü artık yazılmıyor; kurtarma entegrasyon testi **varsayılan** changelog'la geçiyor. `--changelog`/`--release-version` CI modunda.
- **Tag şablonları (5–9)**: `git.VersionFromTag`; all-refs + türetilmiş filtre → ham en yüksek tag'e düşüş; açık `tagMatch` eşleşmezse hata; `pre*` anahtar kelimeleri aynen; açık sürüm mevcut sürümden büyük olmalı.
- **HTTP (12, 15, 16) + changelog (13)**: POST yalnız 429/503'te tekrar; `DefaultTransport.Clone()` ile proxy env; `make_latest` her zaman açık; compare linkleri tag adıyla.
- **Staging (14)**: `git add . --update` (npm kuralı) geri geldi — Faz 26.6 hedefli staging hook çıktılarını dışarıda bırakıyordu; `addUntrackedFiles` ek olarak yeni dosyaları alır.
- **Remote (17), autoDetect (18)**: platform release'i açıkken remote parse edilemezse prerequisites hatası; `pushRepo` RepoInfo kaynağı; trailing slash kabul; auto-detect git hatası `Warn`.

QA notu: 19 maddenin tamamı önce kırmızı testle doğrulandı; iki mevcut test (bumper text overwrite, targeted staging) hatalı davranışı kodladığı için yeni sözleşmeye göre yeniden yazıldı; entegrasyon testleri repo durumunu (HEAD, tag listesi, working tree, dosya baytları) assert ediyor.
