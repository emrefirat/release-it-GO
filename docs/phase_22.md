# Faz 22: Atomic Git Push Default

## Özet

`git push` çağrısına `--atomic` flag'inin varsayılan olarak eklenmesi. Paralel CI yarış senaryolarında **orphan tag** durumunu engeller (commit reddedilirken tag remote'a iletilmesini önler).

## Problem

Mevcut `Push()` davranışı `git push --follow-tags origin` çalıştırıyor. Bu komut **atomic değil**; refs (branch + tag) ayrı ayrı pushlanıyor:

```
* [new tag]         0.10.0 -> 0.10.0    ← TAG REMOTE'A GİTTİ
! [rejected]        master -> master (fetch first)  ← BRANCH REDDEDİLDİ
```

### Veri Bütünlüğü Sonucu

1. **Aynı run**: Tag remote'da, master local'de — tutarsız state
2. **Sonraki run**: Fresh checkout → master HEAD'de 0.10.0 commit'i yok → `git describe --tags --abbrev=0` 0.9.0 döner → release-it-go 0.10.0 hesaplar → `TagExists("0.10.0")` orphan tag'i bulur → **"tag already exists"** hatası

Bu paralel CI ortamlarında (Jenkins multi-job, GitLab CI parallel pipelines, GitHub Actions matrix) gerçek bir veri bütünlüğü riskidir.

## Çözüm

`PushArgs` default değerine `--atomic` eklenir:

```go
// internal/config/defaults.go
PushArgs: []string{"--follow-tags", "--atomic"}  // önceki: []string{"--follow-tags"}
```

`git push --atomic` davranışı:
- Tüm ref'ler **tek transaction** olarak pushlanır
- Bir ref reddedilirse **tümü reddedilir** (rollback)
- Orphan tag durumu fizik olarak imkânsız hale gelir

## Hedefler

- Default davranış paralel CI'da güvenli olur
- Var olan kullanıcı override'ları etkilenmez (`pushArgs: [...]` özelleştirenler)
- Test ve lint başına regresyon yok
- Mevcut paket kapsamları korunur (`config`, `git` paketleri %85+ hedefinde)

## Kapsam Dışı

- `git push --atomic` server desteği yoksa workaround (kullanıcı config override'la halleder)
- Push başarısız olursa local commit/tag rollback'i (büyük scope, ayrı faz)
- `--atomic` özelliğini opt-in flag ile kapatma (config override zaten yeterli)

## Yapılacaklar

### Bölüm 1 — Test First (CLAUDE.md Rule #5)

- [ ] `internal/git/push_test.go` — `TestPush_Default` testine `--atomic` beklentisi ekle (TDD: önce kırılan testi yaz)
- [ ] `internal/git/push_test.go` — `TestPush_DefaultAtomicEnabled` yeni test (explicit doğrulama)
- [ ] `internal/git/push_test.go` — `TestPush_CustomArgsOverridesDefault` mevcut override davranışının korunduğunu doğrula
- [ ] `internal/config/config_test.go` — default değer testini güncelle (`--follow-tags`, `--atomic`)
- [ ] Test çalıştır, kırıldıklarını gör

### Bölüm 2 — Default Değişikliği

- [ ] `internal/config/defaults.go` — `PushArgs` default'una `--atomic` ekle
- [ ] Test'leri yeşil hale getir

### Bölüm 3 — Doc Senkronizasyonu

- [ ] `internal/config/writer.go` — `fullExampleYAML` içindeki `pushArgs` örneğini güncelle
- [ ] `README.md` — config reference'ında `pushArgs` örneği + atomic açıklaması
- [ ] `TROUBLESHOOTING.md` — "git server does not support --atomic push" senaryosu + override örneği
- [ ] `CHANGELOG.md` — Unreleased bölümüne giriş

### Bölüm 4 — Phase Kapanışı

- [ ] `PROGRESS.md` — Phase 22 satırı + Notes + Change History
- [ ] `docs/phase_22.md` — Status: Complete

## Hata Senaryoları ve Beklenen Davranış

| Senaryo | Davranış |
|---------|----------|
| Modern git server (atomic destekli) | Atomic push, orphan tag oluşmaz |
| Eski git server (atomic desteksiz) | `fatal: the receiving end does not support --atomic push` — kullanıcı net hata mesajı görür |
| Kullanıcı `pushArgs: ["--follow-tags"]` override etmiş | Default override edilir, eski davranış korunur |
| `pushArgs: []` (push args yok) | `git push origin` — `--atomic` da yok, kullanıcının kararı |
| Push tek ref ise (branch yok, sadece tag) | `--atomic` no-op, atomic semantiği zaten geçerli |

## Risk Analizi

| Risk | Olasılık | Etki | Mitigation |
|------|----------|------|------------|
| Server `--atomic` desteklemez | Düşük (%1, 2015 öncesi server'lar) | Push başarısız | Config override 1 satır, net hata mesajı |
| Test failure (mevcut testler kırılır) | Yüksek (bilerek, TDD gereği) | — | Tests güncellenecek (Bölüm 1) |
| npm release-it compat kırılır | Yok | — | `pushArgs` config alanı her iki tarafta zaten var |
| Mevcut kullanıcı override'ları | Yok | — | Default değişikliği override mantığını etkilemez |
| Atomic ile dry-run uyumsuzluğu | Yok | — | Dry-run zaten çıktı üretmiyor, etki yok |

## Başarı Kriterleri

- [ ] `make check` temiz (fmt + vet + lint + vuln + test + build)
- [ ] `go test ./... -race` tüm paketler yeşil
- [ ] `internal/git` paket kapsamı %85+ korunur
- [ ] `internal/config` paket kapsamı %85+ korunur
- [ ] Default config'in `pushArgs` değeri `["--follow-tags", "--atomic"]`
- [ ] PROGRESS.md güncel
- [ ] Atomic commit chain (test-first → kod → docs → PROGRESS)

## Commit Stratejisi

| Sıra | Commit | İçerik |
|------|--------|--------|
| 1 | `docs(phase-22): add PRD for atomic git push default` | Bu dosya |
| 2 | `test(git): expect --atomic in default push args` | Bölüm 1 — TDD red phase |
| 3 | `fix(config): add --atomic to default git.pushArgs` | Bölüm 2 — TDD green phase |
| 4 | `docs: document --atomic default and override workaround` | Bölüm 3 |
| 5 | `docs(progress): mark Phase 22 atomic push completion` | Bölüm 4 |

Her commit öncesi `make check`.
