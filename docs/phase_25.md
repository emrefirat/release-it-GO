# Faz 25: npm Parite Paketi

## Özet

2026-08-25 denetiminin parite analizinde (Part B) belirlenen, npm release-it'ten göçen bir kullanıcının ilk on dakikada çarptığı boşlukların kapatılması. Hedef: belgelenmiş npm kullanım biçimlerinin release-it-go'da da aynı sonucu vermesi.

## Kapsam ve Durum

| # | Özellik | npm davranışı | Mevcut durum | Durum |
|---|---------|---------------|--------------|-------|
| 1 | Pozisyonel increment/sürüm (`release-it-go minor`, `release-it-go 1.5.0`) + `-i 1.5.0` | Birincil belgelenmiş kullanım | Arg'lar sessizce yutuluyor; `-i 1.5.0` "unsupported increment type" | ✅ |
| 2 | Açık increment'te prompt açılmaması | Açık `-i` asla sormaz | Auto-detect ile çakışınca soruyor (A19) | ✅ |
| 3 | v-önek çıkarımı (son tag `v` ile başlıyorsa `tagName` → `v${version}`) | GitBase.js çıkarımı | Hardcoded `${version}`; v'li repolarda yeni tag'ler öneksiz | ✅ |
| 4 | `--no-increment` kurtarma akışı (mevcut tag'le release adımlarını tekrar koşma) | Belgelenmiş kurtarma yolu | `tag already exists` hatasıyla ölüyor (M4/A8) | ✅ |
| 5 | git şablon değişkenleri (`${branchName}`, `${latestVersion}`, `${repo.*}` vb. commitMessage/tagName/tagAnnotation/releaseName içinde) | Destekli | Yalnız `${version}` işleniyor | ✅ |
| 6 | İnteraktif `--preRelease` seçenekleri (prepatch/preminor/premajor menüsü) | Pre-release varyantları sunulur | Menü düz patch/minor/major; kimlik düşüyor (H5/A5) | ✅ |
| 7 | Commit/tag/push onaylarının bağımsızlığı (birine hayır demek diğerlerini atlamasın) | Her adım ayrı sorulur | "Commit?" reddi tag+push'u da iptal ediyor (M1/A7) | ✅ |
| 8 | Migration'ın tüm bölümleri yazması (hooks/bumper/calver/notification) | — | `toConfigMap` yalnız 4/8 bölüm; hooks sessizce kayboluyor (F4) | ✅ |

## Kapsam Dışı (Faz 24 / 26 / sonrası)

- Genel dot-notation flag override'ları (`--git.tagName=...`) — ayrı tasarım gerektirir.
- `github.comments`, `releaseNotes` komutu, `package.json` config kaynağı, changelog önizlemesi, `git.commitsPath` — sonraki parite turu.
- Ölü config alanları envanterinin bağla-veya-sil taraması — Faz 26 sonrası temizlik.

## Prensipler

- Her madde test-first (kırmızı → yeşil); davranış değişiklikleri CHANGELOG Unreleased'e işlenir.
- npm semantiği kaynak: release-it dokümantasyonu (denetim Part B'de Context7 ile doğrulandı).
- Mevcut açık kullanıcı config'leri davranış değiştirmez (çıkarımlar yalnız varsayılan değerlerde devreye girer).

## Tamamlanma Notları (2026-08-26)

1. **Pozisyonel/açık sürüm**: `resolveIncrementArg` (cli) + `version.IsIncrementType`; geçersiz arg net hatayla reddediliyor (`cobra.MaximumNArgs(1)`). `determineSemVer` açık semver'i (`-i 1.5.0`, `release-it-go 1.5.0`, `v` önekli dahil) doğrudan hedef sürüm olarak kullanıyor.
2. **Prompt**: açık increment (pozisyonel/-i/config) artık asla sormuyor; eski değer-eşitliği koşulu (ikinci autoDetect çağrısıyla) kaldırıldı (A19+L1).
3. **v-önek çıkarımı**: `inferTagNameFormat` — default şablonda, son ham tag v'li ise `v${version}`. Loader `git.tagname` dosyada yazılıysa `TagNameExplicit` işaretliyor; açık şablon çıkarımı devre dışı bırakıyor (bilinçli v-kaldırma geçişleri korunur).
4. **--no-increment kurtarma**: `createReleaseTag` HEAD'i gösteren mevcut tag'i atlıyor (`Tag ... already exists at HEAD — skipping`); farklı commit'teki tag net hata. `TagPointsAtHead` git metodu eklendi.
5. **Şablon değişkenleri**: `renderReleaseTemplate` (tek geçişli regex) — commitMessage/tagAnnotation/github+gitlab releaseName'de `${branchName}`, `${latestVersion}`, `${repo.*}` vb.; bilinmeyen değişken literal kalır. `tagName` bilinçli olarak `${version}`-only (geçmiş tag aramalarının simetrisi bozulmasın).
6. **preRelease menüsü**: `buildVersionOptions` pre modda prepatch/preminor/premajor (+ seri devamı `prerelease` ilk sırada, önerili) sunuyor; kimlik asla düşmüyor.
7. **Onay bağımsızlığı**: `confirmStep` — commit/tag/push ayrı ayrı; hayır yalnız o adımı atlar, prompt hatası (Ctrl+C) release'i durdurur.
8. **Migration**: `toConfigMap` 8 bölüm + üst-düzey skalarlar; hooks/notification round-trip testi.
