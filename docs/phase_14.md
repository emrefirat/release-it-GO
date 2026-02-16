# Phase 14: YAML Config Yazma + Init Format Secimi

## Ozet

`init --full-example` komutu yorumlu YAML uretir, init wizard'inda kullanici format secer (JSON/YAML), migration da secilen formatta cikti verebilir.

## Degisiklikler

### 1. YAML Config Yazma (`internal/config/writer.go`)

- `WriteConfigYAML(cfg, path)` — minimal config'i YAML olarak yazar
- `WriteFullExampleYAML(path)` — yorumlu YAML referans dosyasi olusturur
- `fullExampleYAML` sabiti — tum opsiyonlar `#` yorum satirlariyla aciklanmis
- `fullExampleJSON` ve `WriteFullExampleJSON` kaldirildi

### 2. Format Secimi (`internal/cli/init.go`)

- Init wizard'a ilk soru olarak format secimi eklendi (JSON / YAML)
- `--full-example` artik `.release-it-go-full.yaml` uretiyor
- Wizard, secilen formatta config yaziyor

### 3. Migration Destegi (`internal/config/migrate.go`)

- `NativeConfigFileForFormat(format)` — formata gore dosya adi dondurur
- `NativeConfigFileYAML` sabiti (`.release-it-go.yaml`)
- `DetectNativeConfigAny()` — hem JSON hem YAML native config tespiti
- `MigrateLegacyConfigTo(path, format)` — secilen formatta migration

### 4. Testler

- YAML yazma testleri (default, non-default, full example, loadable)
- Init wizard YAML format testi
- Mevcut testler format select sorusuyla guncellendi

## Dosyalar

| Dosya | Islem |
|-------|-------|
| `internal/config/writer.go` | `WriteConfigYAML`, `WriteFullExampleYAML`, `fullExampleYAML` |
| `internal/config/writer_test.go` | YAML testleri eklendi, JSON full example testleri guncellendi |
| `internal/config/migrate.go` | `NativeConfigFileForFormat`, `DetectNativeConfigAny`, `MigrateLegacyConfigTo` |
| `internal/cli/init.go` | Format secim sorusu, YAML yazma, full example YAML |
| `internal/cli/init_test.go` | YAML format testi, mevcut testlere format select ekleme |
| `docs/phase_14.md` | Bu dosya |
| `PROGRESS.md` | Faz 14 guncelleme |
