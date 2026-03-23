# Faz 18: Config Uyumluluk ve Edge Case Düzeltmeleri

## Özet

QA incelemesinde tespit edilen config yükleme, npm uyumluluk katmanı ve edge case sorunlarının düzeltilmesi.

## Problem 1: LoadConfigFromBytes JSON Normalization Atlanıyor

`loader.go`'daki `LoadConfigFromBytes()` fonksiyonu doğrudan `json.Unmarshal()` çağırıyor — `normalizeJSON()` ve `applyPluginCompat()` atlanıyor. Legacy npm release-it config'lerinde array tipindeki alanlar (örn: `requireBranch: ["main", "master"]`) string'e dönüştürülmeden parse'a gönderiliyor ve hata alınıyor.

## Problem 2: Plugin Ayarları Kullanıcı Config'ini Override Ediyor

`compat.go`'daki `applyConventionalChangelogPlugin()` kullanıcının explicit olarak set ettiği `changelog.enabled` değerini plugin ayarıyla koşulsuz olarak ezebiliyor. Beklenen davranış: kullanıcı config'i plugin'den öncelikli olmalı.

## Problem 3: normalizeJSON Sessiz Hata

`normalizeJSON()` parse edilemeyen JSON'u sessizce orijinal haliyle döndürüyor. Bu, hatalı config'lerin kısmen yüklenmesine ve beklenmeyen davranışa yol açabilir.

## Problem 4: SemVer Boş PreReleaseID

`IncrementVersion()` boş `preReleaseID` ile çağrıldığında `"1.0.0-0.0"` gibi semantik olarak anlamsız versiyonlar üretebiliyor.

## Çözüm

### 1. LoadConfigFromBytes'a Normalization Ekle
JSON formatında `normalizeJSON()` çağrısı ekle.

### 2. Plugin Override Koşullu Yap
Plugin ayarı sadece kullanıcı explicit set etmemişse uygulansın.

### 3. normalizeJSON Hata Logging
Parse hatası durumunda uyarı logla (mevcut fallback davranışı koru).

### 4. Boş PreReleaseID Validasyonu
Boş preReleaseID ile pre-release increment istendiğinde hata dön veya mantıklı default kullan.

## Yapılacaklar

- [ ] `internal/config/loader.go` — LoadConfigFromBytes'a normalizeJSON ekle (JSON format)
- [ ] `internal/config/loader_test.go` — Array → string normalization testi (LoadConfigFromBytes)
- [ ] `internal/config/compat.go` — Plugin override'ı koşullu yap
- [ ] `internal/config/compat_test.go` — Plugin + explicit config çakışma testi
- [ ] `internal/config/compat.go` — normalizeJSON hatasında log uyarısı
- [ ] `internal/version/semver.go` — Boş preReleaseID validasyonu
- [ ] `internal/version/semver_test.go` — Boş preReleaseID testleri
- [ ] Tüm mevcut testlerin geçtiğini doğrula

## Test Senaryoları

### LoadConfigFromBytes Normalization
```go
// Array → string dönüşümü
input := `{"git": {"requireBranch": ["main", "master"]}}`
// Beklenen: cfg.Git.RequireBranch == "main" (ilk eleman)
```

### Plugin Override
```json
{
  "changelog": {"enabled": true},
  "plugins": {"@release-it/conventional-changelog": {"changelog": false}}
}
```
Beklenen: `changelog.enabled = true` (kullanıcı config öncelikli)

### Boş PreReleaseID
```go
IncrementVersion(v, "premajor", "")
// Beklenen: error veya default ID ("0" veya "rc")
```

## Riskler

- normalizeJSON davranış değişikliği mevcut uyumluluk katmanını etkileyebilir
- Plugin override değişikliği npm migration davranışını değiştirir
- PreReleaseID validasyonu mevcut kullanımları kırabilir — dikkatli kontrol gerekli
