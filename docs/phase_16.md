# Faz 16: Kritik Bug Düzeltmeleri — URL Parsing + CalVer

## Özet

QA incelemesinde tespit edilen kritik seviye bugların düzeltilmesi. GitLab nested group URL'lerinin desteklenmemesi ve CalVer "yyyy.mm.dd" formatının tamamen bozuk olması.

## Problem 1: GitLab Nested Group URL'leri

`repo.go`'daki regex pattern'leri sadece `host/owner/repo` formatını destekliyor. GitLab'ın yaygın olarak kullanılan nested group yapısı (`host/group/subgroup/repo`) ayrıştırılamıyor.

**Etkilenen regex'ler:**
```go
var httpsPattern = regexp.MustCompile(`^https?://(?:[^@]+@)?([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)
var sshPattern = regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+?)(?:\.git)?$`)
```

**Başarısız olan girdiler:**
- `https://gitlab.com/mygroup/subgroup/myproject.git` → NO MATCH
- `git@gitlab.com:mygroup/subgroup/myproject.git` → NO MATCH

**Beklenen davranış:**
- `Owner = "mygroup/subgroup"`, `Repository = "myproject"`

## Problem 2: CalVer "yyyy.mm.dd" Formatı Bozuk

`calver.go`'daki `parse()` fonksiyonu format-agnostic çalışıyor ve 3. bileşeni her zaman `Minor` olarak yorumluyor. "yyyy.mm.dd" formatında 3. bileşen `Day` olmalı ama `Day` alanı hiç set edilmiyor.

Ayrıca `format()` fonksiyonu "yyyy.mm.dd" case'inde `minor` parametresini tamamen yok sayıyor — aynı gün içinde birden fazla release yapılamıyor.

## Çözüm

### 1. URL Regex Güncellemesi

Regex'leri son `/` öncesindeki tüm path segmentlerini `Owner` olarak yakalaacak şekilde güncelle:
```go
// HTTPS: host / (path segments) / repo
var httpsPattern = regexp.MustCompile(`^https?://(?:[^@]+@)?([^/]+)/(.+)/([^/]+?)(?:\.git)?$`)
// SSH: host : (path segments) / repo
var sshPattern = regexp.MustCompile(`^git@([^:]+):(.+)/([^/]+?)(?:\.git)?$`)
```

### 2. CalVer Format-Aware Parsing

`parse()` fonksiyonuna format parametresi ekle:
- "yyyy.mm.minor" → 3. bileşen = Minor
- "yyyy.mm.dd" → 3. bileşen = Day, Minor = 4. bileşen (varsa)

`format()` fonksiyonunu "yyyy.mm.dd" case'inde minor desteği ekle:
- Minor > 0 ise: `"2024.3.15.1"` formatında çıktı

## Yapılacaklar

- [ ] `internal/git/repo.go` — HTTPS + SSH regex güncelleme (nested group desteği)
- [ ] `internal/git/repo_test.go` — Nested group URL testleri (HTTPS + SSH, 2-3 seviye)
- [ ] `internal/version/calver.go` — `parse()` format-aware parsing, `format()` minor desteği
- [ ] `internal/version/calver_test.go` — "yyyy.mm.dd" format testleri (parse, format, NextVersion)
- [ ] CalVer input validation (month 1-12)
- [ ] Tüm mevcut testlerin geçtiğini doğrula

## Test Senaryoları

### URL Parsing
| Input | Owner | Repository |
|-------|-------|------------|
| `https://gitlab.com/group/repo.git` | `group` | `repo` |
| `https://gitlab.com/group/sub/repo.git` | `group/sub` | `repo` |
| `https://gitlab.com/a/b/c/repo.git` | `a/b/c` | `repo` |
| `git@gitlab.com:group/sub/repo.git` | `group/sub` | `repo` |

### CalVer yyyy.mm.dd
| Input | Parsed Day | Parsed Minor |
|-------|-----------|--------------|
| `2024.3.15` | 15 | 0 |
| `2024.3.15.2` | 15 | 2 |
| `2024.13.5` | ERROR (invalid month) | - |

## Riskler

- URL regex değişikliği mevcut `owner/repo` ayrıştırmayı bozabilir — mevcut testler regresyon koruması sağlar
- CalVer parse değişikliği "yyyy.mm.minor" davranışını etkilememelidir
