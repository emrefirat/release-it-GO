# Phase 8: Init Command & Dual Config File Support

## Ozet

Bu faz, release-it-go'ya interaktif `init` komutu ve ikili config dosyasi destegi ekler.

## Ozellikler

### 1. Dual Config Dosya Destegi

- `.release-it-go.json` (native) dosyalari oncelikli aranir
- `.release-it.json` (legacy/npm uyumlu) dosyalari geri donus olarak desteklenir
- Arama sirasi: `.release-it-go.*` > `.release-it.*`
- Tum formatlar desteklenir: JSON, YAML, TOML

### 2. Init Command (`release-it-go init`)

Interaktif wizard ile sifirdan config olusturma:

- **Platform secimi:** GitHub / GitLab / Sadece Git tag
- **Changelog formati:** Conventional Changelog / Keep a Changelog / Yok
- **Git islemleri:** Commit, tag, push aktif/pasif
- **Commit mesaji sablonu:** Kullanici girisli (varsayilan: `chore(release): release v${version}`)
- **Tag formati:** Kullanici girisli (varsayilan: `v${version}`)
- **Gerekli branch:** Kullanici girisli (varsayilan: `main`)

### 3. Legacy Config Migration

- Mevcut `.release-it.json` dosyasi algilandiginda otomatik migrate teklifi
- Backup olusturma (`.release-it.json.bak`)
- normalizeJSON + applyPluginCompat ile npm alanlari temizleme
- Temiz `.release-it-go.json` ciktisi

### 4. Smart Config Writer

- Sadece default'tan farkli alanlari JSON'a yazar
- `json.MarshalIndent` ile 2-space indent
- Dosya izinleri: 0644

## Yeni Dosyalar

| Dosya | Aciklama |
|-------|----------|
| `internal/config/writer.go` | WriteConfigJSON - config'i JSON olarak yazar |
| `internal/config/migrate.go` | DetectLegacyConfig, MigrateLegacyConfig |
| `internal/cli/init.go` | Init komutu ve wizard akisi |
| `internal/config/writer_test.go` | Writer testleri |
| `internal/config/migrate_test.go` | Migration testleri |
| `internal/cli/init_test.go` | Init command testleri |

## Degistirilen Dosyalar

| Dosya | Degisiklik |
|-------|-----------|
| `internal/config/loader.go` | configSearchFiles'a `.release-it-go.*` eklendi |
| `internal/ui/prompt.go` | Prompter interface'ine `Select` metodu eklendi |
| `internal/cli/root.go` | Init komutu kaydedildi |
| `internal/runner/runner_test.go` | Mock prompter'lara `Select` metodu eklendi |
| `internal/ui/prompt_test.go` | Select testleri eklendi |
| `internal/config/config_test.go` | Config oncelik testleri eklendi |

## Kullanim

```bash
# Yeni config olustur (interaktif wizard)
release-it-go init

# CI modunda (default degerlerle)
release-it-go init --ci

# Mevcut .release-it.json otomatik migrate edilir
# → .release-it.json.bak (backup)
# → .release-it-go.json (native)
```
