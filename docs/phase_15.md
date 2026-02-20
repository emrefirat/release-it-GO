# Faz 15: Branch-Aware Pre-Release Version Detection

## Özet

Pre-release versiyonlama sırasında farklı branch'lerdeki tag'lerin birbirini etkilemesini önleyen branch-aware çözümleme. `--preRelease="deneme"` ile çalışırken sadece mevcut branch'ten erişilebilir tag'lere bakarak seri devamı veya yeni seri başlatma kararı verir.

## Problem

`git describe --tags --abbrev=0` komutu **tüm branch'lerdeki tag'lere** bakar. npm release-it'te versiyon `package.json`'dan okunduğu için her branch kendi serisini doğal olarak koruyordu. release-it-go'da `package.json` yok, versiyon kaynağı git tag'ler.

**Senaryo:** `deneme` branch'inde `--preRelease="deneme"` ile çalışırken, başka branch'lerdeki tag'ler (v2.0.0-beta.0 gibi) en son tag olarak bulunuyor ve `deneme` serisi kırılıyor.

## Çözüm

Pre-release kullanıldığında, `git tag --merged HEAD` ile sadece **mevcut branch'ten erişilebilir** tag'lere bakarak seri devamını veya yeni seri başlatmayı belirlemek.

## Algoritma

```
1. HEAD'den erişilebilir tag'ler arasında *-{preReleaseID}.* pattern'ine uyanı bul
2. HEAD'den erişilebilir en son stable (pre-release'siz) tag'i bul
3. Karar:
   - Pre-release tag bulundu VE base version >= stable → seriyi devam ettir
   - Pre-release tag bulunamadı VEYA base < stable → yeni seri başlat
```

## Desteklenen Senaryolar

### A) Uzun yaşayan branch (seri devam)

```
master:  v1.2.4 ── v1.2.5 ── v2.0.0
              \
deneme:        ── v1.2.5-deneme.0 ── deneme.1 ── (HEAD)
→ deneme tag'leri erişilebilir → v1.2.5-deneme.2
```

### B) Silinen ve yeniden açılan branch (yeni seri)

```
master:  v1.2.4 ── v1.2.5 ── v2.0.0
              \                    \
(eski deneme)  ── deneme.0         (yeni deneme HEAD)
→ eski deneme tag'leri erişilemez → v2.0.1-deneme.0
```

### C) Uzun yaşayan branch, master ilerlemiş (seri devam)

```
master:  v1.2.4 ── v1.2.5 ── v1.3.0 ── v2.0.0
              \
deneme:        ── v1.2.5-deneme.0 ── (HEAD)
→ deneme.0 erişilebilir, base(1.2.5) >= stable(1.2.4) → v1.2.5-deneme.1
```

## Dosyalar

| Dosya | Açıklama |
|-------|----------|
| `internal/git/tag.go` | `GetLatestPreReleaseTagMerged()` + `GetLatestStableTagMerged()` metodları |
| `internal/runner/runner.go` | `resolvePreReleaseBaseTag()` metodu, `determineVersion()` güncellemesi |
| `internal/git/tag_test.go` | Unit testler (16 test: 8 pre-release + 8 stable) |
| `test/integration/release_test.go` | Integration testler (4 test: seri devam, yeni seri, master ilerlemiş, flag'siz) |

## API

### `GetLatestPreReleaseTagMerged(preReleaseID string) (string, error)`

`git tag -l --merged HEAD --sort=-v:refname` çalıştırır, tag'ler arasında `-{preReleaseID}.` pattern'ini arar. TagMatch/TagExclude filtrelerini uygular. Bulursa tag string döner, bulamazsa `("", nil)`.

### `GetLatestStableTagMerged() (string, error)`

Aynı `git tag -l --merged HEAD --sort=-v:refname` çalıştırır, pre-release suffix'i olmayan ilk tag'i döner (version parse edip `Prerelease() == ""` kontrolü). TagMatch/TagExclude filtrelerini uygular.

### `resolvePreReleaseBaseTag(preReleaseID string) (string, error)`

`GetLatestPreReleaseTagMerged()` ve `GetLatestStableTagMerged()` çağırır. Pre-release tag'in base version'ını stable ile karşılaştırır. Seri devam mı, yeni seri mi kararını verir.

## Testler

### Unit Testler (internal/git/tag_test.go)

| Test | Açıklama |
|------|----------|
| finds matching pre-release tag | Doğru pre-release ID'yi buluyor |
| skips non-matching pre-release IDs | Farklı ID'leri atlıyor |
| does not match partial ID (beta vs beta2) | Kısmi eşleşmeyi önlüyor |
| empty preReleaseID returns empty | Boş ID'de boş döner |
| no tags returns empty | Tag yoksa boş döner |
| respects TagMatch filter | TagMatch filtresi çalışıyor |
| respects TagExclude filter | TagExclude filtresi çalışıyor |
| git error returns error | Git hatası error döner |
| finds stable tag skipping pre-release | Pre-release'i atlayıp stable buluyor |
| all pre-release returns empty | Hepsi pre-release ise boş döner |
| first tag is stable | İlk tag stable ise onu döner |
| respects TagMatch/TagExclude | Filtreler stable aramada da çalışıyor |
| invalid version tags are skipped | Parse edilemeyenler atlanıyor |

### Integration Testler (test/integration/release_test.go)

| Test | Beklenen Sonuç |
|------|---------------|
| `PreRelease_BranchAware_ContinueSeries` | v1.2.5-deneme.0 → v1.2.5-deneme.1 |
| `PreRelease_BranchAware_NewSeries` | v2.0.0 → v2.0.1-deneme.0 |
| `PreRelease_BranchAware_MasterAdvanced` | v1.2.5-deneme.0 → v1.2.5-deneme.1 |
| `PreRelease_NoFlag_BehaviorUnchanged` | v1.0.0 → v1.1.0 |
