# Faz 28: Parametre Doğrulama ve Ölü Alan Temizliği

**Durum:** ✅ Tamamlandı — 2026-09-02 (`2ff14a9`, `fb08e2d`, `5ff9951`)

## Özet

İkinci denetimin parametre-geçerliliği envanteri: her CLI flag ve config alanı izlendi, 16 ölü/etkisiz alan ve bir dizi sessiz-geçersiz değer tespit edildi. Hedef: config yüklenirken **doğrulama** (`config.Validate()`), bilinmeyen anahtar tespiti ve her ölü alan için bağla-veya-sil kararının uygulanması.

## Kapsam ve Durum

### Doğrulama (LoadConfig sonrası `Validate()`)

| # | Kural | Sessiz hata senaryosu | Durum |
|---|-------|------------------------|-------|
| 1 | Bilinmeyen config anahtarı → hata (npm'in bilinen anahtarları `npm`, `plugins`, `versionFile`, `changelogFile` TÜM formatlarda önce ayıklanır) | `hooks.preCommit`, `github.relase: true` sessizce yutuluyor | ✅ `2ff14a9` |
| 2 | Runner'ın tetiklediği tüm adımlar için hook alanı: `before/after:prerequisites`, `commitlint`, `version`, `changelog`, `notification` | Config'e yazılınca düşüyor; README "her adım" diyor | ✅ `2ff14a9` |
| 3 | `git.tagName` `${version}` içermeli | İkinci release'te commit oluşup tag'de patlıyor | ✅ `2ff14a9` |
| 4 | `increment` anahtar kelime veya geçerli semver (config dosyasında ve `-i`'de, pozisyonelle aynı kural) | `-i bogus` prerequisites sonrası hata | ✅ `2ff14a9` (config) / `fb08e2d` (`-i`) |
| 5 | `preReleaseId` `^[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$` | `beta!` geç aşamada hata | ✅ `2ff14a9` |
| 6 | `calver.format` ∈ {yy.mm.minor, yyyy.mm.minor, yyyy.mm.dd} | Tanınmayan değer sessizce `yy.mm.minor` | ✅ `2ff14a9` |
| 7 | `github.host` / `gitlab.origin` şema kuralları (host şemasız, origin şemalı) | `https://https://…`, "unsupported protocol scheme" release adımında | ✅ `2ff14a9` |
| 8 | `github.timeout`, `notification.webhooks[].timeout` ≥ 0 | `-5` → süresiz bekleme | ✅ `2ff14a9` |
| 9 | `notification.webhooks[].type` ∈ {slack, teams}; `bumper.*.type` ∈ {json, yaml, toml, ini, text} | Release sonrası uyarı / bump adımında hata | ✅ `2ff14a9` |
| 10 | Mod flag çakışmaları hata: `--check-msg`+`--dry-run`, `--changelog`+`--release-version`, `minor --no-increment`, `--only-version --no-increment` | Sessiz if-sırası önceliği | ✅ `fb08e2d` |
| 11 | `--dry-run` `init` ve `hooks install`'da uygulanır | Dosya yazıyor, `core.hooksPath` set ediyor | ✅ `fb08e2d` |
| 12 | Legacy YAML/TOML config'lere npm uyumluluğu (normalize + plugin compat) — JSON'la aynı | `requireBranch: [...]` YAML'da sert hata; plugins yok sayılıyor | ✅ `2ff14a9` |

### Ölü alan kararları

| Alan | Karar | Gerekçe | Durum |
|------|-------|---------|-------|
| `git.changelog` (+ `Git.GenerateChangelog`) | SİL | Conventional renderer var; npm dosyalarında normalize ile ayıkla | ✅ `5ff9951` |
| `git.commitsPath` | BAĞLA | npm monorepo kaldıracı; 3 git log çağrısına `-- <path>` | ✅ `5ff9951` |
| `github.releaseNotes`, `gitlab.releaseNotes` | SİL | npm'de shell komutu; injection yüzeyi; changelog notları üretiyor | ✅ `5ff9951` |
| `github.web` | SİL | npm'deki anlamı farklı; akış yok | ✅ `5ff9951` |
| `github.comments.*` | Ayrı faz (bağla) — o zamana dek doküman/defaults'tan kaldır | `PostComment` var, `Closes #N` çıkarımı yok | ✅ `5ff9951` |
| `github.makeLatest` | DÜZELT (Faz 27.16) | — | ✅ `b2781a9` (Faz 27) |
| `gitlab.preRelease` | SİL (+ root.go yazımı) | GitLab'da kavram yok | ✅ `5ff9951` |
| `gitlab.useGenericPackageRepositoryForAssets` | BAĞLA (`false` → project uploads API) | npm varsayılanı false | ✅ `5ff9951` |
| `changelog.preset` | DOĞRULA | {angular, conventionalcommits} kabul, diğerine hata | ✅ `5ff9951` |
| `changelog.addUnreleased`, `keepUnreleased` | SİL | Hiç bağlanmamış, talep yok | ✅ `5ff9951` |
| `changelog.addVersionUrl` | BAĞLA (default true) | `conventional.go:21` tek satır | ✅ `5ff9951` |
| `calver.increment`, `calver.fallbackIncrement` | SİL | `NextVersion` sabit davranış | ✅ `5ff9951` |
| `bumper.*.versionPrefix` | SİL (prefix'e birleştir) | İki prefix alanı, biri ölü | ✅ `5ff9951` |

## QA Test Planı

Her doğrulama kuralı için tablo-güdümlü geçerli/geçersiz değer testi; hata mesajı alan adını ve beklenen biçimi içermeli; `LoadConfig` seviyesinde (dosyadan) ve `Validate()` biriminde. Bilinmeyen anahtar için JSON/YAML/TOML üçünde de test. Silinen alanlar için "eski config yüklenmeye devam eder (uyarıyla)" geriye uyumluluk testi.
