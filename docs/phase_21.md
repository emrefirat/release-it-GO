# Faz 21: P0 Test Kapsam Boşluklarının Kapatılması

## Özet

QA denetiminde tespit edilen **P0 seviye** (kritik) test boşluklarının kapatılması. Amaç: CLAUDE.md'de belirtilen minimum %70 paket kapsamı ve kritik paketler için %85+ hedefine uyumun sağlanması. Bu faz **yalnızca `_test.go` dosyalarına** dokunur; üretim kodu değişmez.

## Problem

Mevcut QA denetiminde aşağıdaki kritik fonksiyonlar %0 veya çok düşük coverage ile tespit edildi:

| Fonksiyon | Dosya | Kapsam | Etki |
|-----------|-------|--------|------|
| `runCheckMsg` | `cli/root.go:246` | 0% | Phase 20 yeni özelliği, hiç test yok |
| `runHooksInstall` | `cli/install.go:59` | 0% | Hook kurulum mantığı test edilmemiş |
| `runHooksRemove` | `cli/install.go:82` | 0% | Managed header kontrolü + silme test edilmemiş |
| `runInit` | `cli/init.go:39` | 0% | Wizard akışı + migration test edilmemiş |
| `fileExists` | `cli/root.go:336` | 0% | `--check-msg` input mode ayrımı için kritik |
| `reasonDescription` | `cli/root.go:324` | 0% | Lint hata açıklamaları |
| `githubRelease` | `runner/runner.go:832` | 28.6% | Pipeline'ın GitHub adımı entegrasyonu |
| `gitlabRelease` | `runner/runner.go:901` | 28.6% | Pipeline'ın GitLab adımı entegrasyonu |

Paket bazında mevcut durum:
- `internal/cli` — **%58.8** (CLAUDE.md min %70 ihlali)
- `internal/runner` — **%78.1** (kritik paket hedefi %85+ altında)

## Hedefler

- `internal/cli` coverage: %58.8 → **%75+**
- `internal/runner` coverage: %78.1 → **%85+**
- Yukarıdaki tablo fonksiyonlarının hepsi → **%70+**
- Hiçbir üretim kodu değişikliği yok (sadece `_test.go` eklemeleri)
- Mevcut coverage regresyonu yok (her paket eski değerin altına düşmez)

## Kapsam Dışı

- P1 integration test genişletmeleri (Phase 22'ye bırakıldı)
- P2 regresyon testleri (PROGRESS.md bug loglarından) (Phase 23'e bırakıldı)
- P3 fuzz/benchmark/golden file/TUI testleri (Phase 24+ nice-to-have)
- Cobra komutlarının `Execute()` entry point testi (cmd/release-it-go smoke test)
- InteractivePrompter (Bubbletea) %0 → `teatest` gerekli, kapsam dışı

## Yapılacaklar

### Bölüm 1: `cli/root.go` yardımcıları (~50 satır)

- [ ] `fileExists`: var olan dosya → true; olmayan → false; symlink; boş string
- [ ] `reasonDescription`: `changelog.LintResult` her `Reason` değeri için beklenen açıklama
- [ ] Table-driven format

**Dosya:** `internal/cli/root_test.go` (mevcut dosyaya ekleme)

### Bölüm 2: `runCheckMsg` (~120 satır)

- [ ] 3 input modu:
  - Geçerli dosya path → dosyadan oku, conventional → exit 0
  - Geçerli dosya path → dosyadan oku, geçersiz → exit 1
  - `-` (stdin) → stdin'den oku (os.Stdin stub veya `cmd.SetIn`)
  - Direkt string → argüman olarak conventional mesaj
  - Direkt string → argüman olarak geçersiz mesaj
- [ ] Boş input hatası
- [ ] Dosya okuma hatası (okunamayan path)
- [ ] Verbose mod (`-V`) ile çıktı formatı doğrulama

**Dosya:** `internal/cli/root_test.go`

**Dikkat:** `runCheckMsg` içinde `os.Stdin` kullanılıyor olabilir. Eğer test edilemiyorsa:
1. Durum tespit et, sorarak onay al
2. Üretim kodunu küçük dokunuşla stub-able hale getir (`io.Reader` parametresi) → **ayrı commit, üretim kodu değişikliği olarak işaretle**
3. Veya: `cmd.SetIn()` cobra mekanizmasıyla test et (tercih edilen)

### Bölüm 3: `runHooksInstall` / `runHooksRemove` (~150 satır)

- [ ] `t.TempDir()` + `git init` helper (mevcut `test/integration/helpers_test.go` pattern'ini taklit et)
- [ ] Install senaryoları:
  - Hiç git hook config yok → bilgi mesajı + çıkış
  - Tek hook (pre-commit) → `.hooks/pre-commit` oluşur, managed header içerir, executable (0755)
  - Çoklu hook → hepsi yazılır
  - `core.hooksPath` set edilir
  - Mevcut managed hook → üzerine yazar (silent)
  - Mevcut non-managed hook → hata ("use --force")
  - `--force` flag → non-managed hook üzerine yazar
- [ ] Remove senaryoları:
  - Managed hook → silinir
  - Non-managed hook → dokunulmaz (silent skip)
  - Son managed hook silindiğinde `core.hooksPath` unset edilir
- [ ] Git repo değilse (`.git` yok) → hata

**Dosya:** `internal/cli/install_test.go` (mevcut 3 test'e ek)

### Bölüm 4: `runInit` (~200 satır)

- [ ] NonInteractivePrompter (CI mode) ile default wizard akışı → `.release-it-go.yaml` yazılır
- [ ] Format seçimi JSON → `.release-it-go.json` yazılır
- [ ] Mevcut native config var + CI mode → overwrite confirm (false default) → abort
- [ ] Legacy `.release-it.json` var → migration önerilir → backup `.bak` oluşur → native yazılır
- [ ] `--full-example` → `.release-it-go-full.yaml` yazılır, yorum satırları içerir
- [ ] Wizard'ın ForceFields ile explicit yazdığı alanlar doğrulanır (ör. `commit: true` default olsa bile yazılır)
- [ ] Format değişimi (JSON var, wizard YAML seçimi) → eski `.bak` olarak yeniden adlandırılır

**Dosya:** `internal/cli/init_test.go` (mevcut 13 test'e ek; overlap olanları atla)

### Bölüm 5: `runner.githubRelease` ve `runner.gitlabRelease` (~200 satır)

- [ ] `httptest.NewServer` ile fake GitHub/GitLab API (mevcut `release/github_test.go` pattern'ini runner'a taşı)
- [ ] GitHub senaryoları:
  - `release: false` → skip mesajı, hiç HTTP call yok
  - `release: true` + geçerli token → release oluşur, `ctx.ReleaseURL` set edilir
  - `release: true` + token yok → prerequisites'de yakalanmış olmalı; runner'a düşerse hata
  - Dry-run + `release: true` → HTTP call yok, DryRun log
  - Asset upload başarılı
  - API 4xx hatası → pipeline durur, error wrap kontrol
- [ ] GitLab için paralel senaryolar (skip, başarı, dry-run, hata)
- [ ] `github.host` ile fake server URL'ini override (GitHub Enterprise pattern)
- [ ] `gitlab.origin` ile fake server URL'ini override

**Dosya:** `internal/runner/runner_test.go` (mevcut test'e ek)

**Dikkat:** Runner `release.NewGitHubClient()` ile client oluşturuyor. Test için:
1. `github.host` + `gitlab.origin` config field'ları via `httptest.Server.URL` yönlendirmek yeterliyse → kolay
2. Değilse, client factory'sini test hook ile mock etmek gerekir → üretim kodu değişikliği → onay al

## Commit Stratejisi

Her bölüm **ayrı commit**. Sıra en kolaydan zora:

| Sıra | Commit Mesajı | Bölüm | Tahmini Satır |
|------|---------------|-------|---------------|
| 1 | `test(cli): cover fileExists and reasonDescription helpers` | 1 | ~50 |
| 2 | `test(cli): cover runCheckMsg file, stdin and string inputs` | 2 | ~120 |
| 3 | `test(cli): cover hooks install and remove commands` | 3 | ~150 |
| 4 | `test(cli): cover runInit wizard flow and legacy migration` | 4 | ~200 |
| 5 | `test(runner): cover github and gitlab release step integration` | 5 | ~200 |
| 6 | `docs(progress): Phase 21 test coverage completion` | PROGRESS güncelleme | ~30 |

Her commit öncesi: `make check` ✓

## Riskler

- **R1 — `runCheckMsg` `os.Stdin` bağımlılığı**: Eğer cobra `cmd.SetIn()` ile test edilemiyorsa üretim kodu değişikliği gerekir. **Mitigation**: Önce kodu oku, stub-able değilse kullanıcıya sor; onay almadan üretim kodu değiştirme.
- **R2 — Runner test için client factory**: Release client instantiation runner içinde hard-coded. Eğer `github.host` override ile test edilemiyorsa dependency injection refactor gerekir. **Mitigation**: Önce mevcut config field'larının yeterli olup olmadığını doğrula; gerekirse ayrı refactor commit'i + kullanıcı onayı.
- **R3 — Init wizard Prompter mock**: Çoklu `Select`/`Confirm` çağrısı sıralı cevap gerektirir. Mevcut mock pattern'i `internal/runner/runner_test.go`'dan taklit et.
- **R4 — Integration vs unit sınırı**: `runHooksInstall` gerçek git komutu çalıştırıyor. `t.TempDir()` + gerçek git yeterli; mock gerekmez. `internal/cli/install_test.go`'un hâlâ unit test sayılabilmesi için fazla ağırlaşmamasına dikkat.
- **R5 — CI ortam bağımlılığı**: `test/integration/helpers_test.go` git user.name/email config ayarı yapıyor. `install_test.go` de aynı setup'ı yapmalı.

## Başarı Kriterleri

- [ ] `make check` temiz
- [ ] `go test ./internal/cli/ -cover` → %75+
- [ ] `go test ./internal/runner/ -cover` → %85+
- [ ] Tüm yeni testler `-race` ile geçer
- [ ] Mevcut coverage hiçbir pakette düşmemiş
- [ ] Üretim kodu değişikliği: 0 (veya kullanıcı onaylı minimal dokunuş)
- [ ] PROGRESS.md güncel
- [ ] 5-6 atomic commit (conventional format)

## Sonraki Fazlar

- **Faz 22**: P1 — Integration test genişletme (GitHub/GitLab/webhook/hooks komutu E2E, httptest tabanlı)
- **Faz 23**: P2 — PROGRESS.md bug loglarından regresyon testleri
- **Faz 24+**: P3 — Fuzz testleri, benchmark'lar, golden file framework, TUI testleri (teatest)
