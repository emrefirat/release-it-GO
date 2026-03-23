# Faz 17: Pipeline Sağlamlık İyileştirmeleri

## Özet

QA incelemesinde tespit edilen pipeline orchestration sorunlarının düzeltilmesi. Hook çalışma sırası, eksik unit testler ve SSH port desteği.

## Problem 1: noCommits + After-Hook Sırası

`runner.go` Run() döngüsünde `noCommits` kontrolü after-hook'lardan **önce** yapılıyor. Bu, noCommits durumunda after-hook'ların çalışmaması anlamına geliyor. Bu davranış doğru olabilir (release iptal edildiğinde hook çalışmasın) ama belgelenmemiş ve test edilmemiş.

```go
if err := step.fn(); err != nil { ... }

if r.ctx.noCommits {
    return nil  // ← after-hook çalışmadan çıkıyor
}

r.ctx.UpdateVars()
r.ctx.HookRunner.RunHooks("after:" + step.name)
```

## Problem 2: latestVersionToTag Unit Test Eksikliği

`latestVersionToTag` fonksiyonu changelog, commitlint ve autoDetectIncrement tarafından kullanılıyor ama hiçbir unit testi yok. Edge case'ler:
- Boş version + template
- "v" prefix'li version + "v${version}" template
- Template'siz kullanım

## Problem 3: SSH Non-Standard Port

`repo.go` SSH regex'i port bilgisini ayrıştırmıyor. `git@host:22:owner/repo.git` gibi URL'lerde port owner'a dahil oluyor.

## Problem 4: Boş Release Notes

Changelog disabled iken GitHub/GitLab release boş description ile oluşturuluyor. Kullanıcı farkında olmayabilir.

## Çözüm

### 1. noCommits Davranışını Belgele ve Test Et
- Mevcut davranışı document et (after-hook çalışmaz)
- Yorum ekle
- Unit test ekle

### 2. latestVersionToTag Testleri
Tüm edge case'ler için unit testler yaz.

### 3. SSH Port Desteği
SSH regex'i `git@host:port:owner/repo` formatını destekleyecek şekilde güncelle. Ya da port kısmını tespit edip owner'dan ayır.

### 4. Boş Release Notes Uyarısı
Changelog boş iken release oluşturulacaksa verbose log ile kullanıcıyı bilgilendir.

## Yapılacaklar

- [ ] `internal/runner/runner.go` — noCommits + after-hook davranışına yorum ekle
- [ ] `internal/runner/runner_test.go` — noCommits senaryosunda after-hook çalışmama testi
- [ ] `internal/runner/runner_test.go` — latestVersionToTag unit testleri (6+ senaryo)
- [ ] `internal/git/repo.go` — SSH port desteği (regex veya post-parse strip)
- [ ] `internal/git/repo_test.go` — SSH port testleri
- [ ] `internal/runner/runner.go` — Boş release notes verbose uyarısı
- [ ] Tüm mevcut testlerin geçtiğini doğrula

## Test Senaryoları

### latestVersionToTag
| latestVersion | tagNameTemplate | Beklenen |
|---------------|-----------------|----------|
| `""` | `"v${version}"` | `""` |
| `"0.0.0"` | `"v${version}"` | `""` |
| `"1.0.0"` | `"v${version}"` | `"v1.0.0"` |
| `"v1.0.0"` | `"v${version}"` | `"v1.0.0"` |
| `"1.0.0"` | `"${version}"` | `"1.0.0"` |
| `"1.0.0"` | `""` | `"1.0.0"` |
| `"v2.0.0"` | `"release-${version}"` | `"release-2.0.0"` |

### SSH Port Parsing
| Input | Host | Owner | Repo |
|-------|------|-------|------|
| `git@host:owner/repo.git` | `host` | `owner` | `repo` |
| `git@host:22:owner/repo.git` | `host` | `owner` | `repo` |

## Riskler

- SSH regex değişikliği mevcut URL ayrıştırmayı etkileyebilir
- noCommits davranış değişikliği mevcut hook kullanıcılarını etkileyebilir
