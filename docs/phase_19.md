# Faz 19: Test Kapsamı Güçlendirme

## Özet

QA test kalitesi incelemesinde tespit edilen eksiklerin giderilmesi. Unit test coverage artırımı, integration test senaryoları eklenmesi ve assertion helper'larının güçlendirilmesi.

## Hedefler

- runner coverage: 78% → 85%+
- release coverage: 86% → 90%+
- git coverage: 89% → 93%+
- version coverage: 89% → 93%+
- Integration test senaryoları: 22 → 30+

## Bölüm 1: Unit Test Eksikleri

### 1.1 release — HTTP Client Config (50% → 85%+)
- [ ] `github.go` createHTTPClient: proxy parse testi (valid + invalid URL)
- [ ] `github.go` createHTTPClient: timeout konfigürasyonu (default + custom)
- [ ] `gitlab.go` createHTTPClient: CA cert yükleme (valid + missing + unreadable)
- [ ] `gitlab.go` createHTTPClient: InsecureSkipVerify (true/false)
- [ ] `gitlab.go` createHTTPClient: CertificateAuthorityFileRef env var

### 1.2 git — Prerequisites Error Paths (53% → 80%+)
- [ ] `prerequisites.go` CheckPrerequisites: git not installed
- [ ] `prerequisites.go` checkBranch: invalid glob pattern fallback
- [ ] `prerequisites.go` checkCleanWorkingDir: git status command error
- [ ] `prerequisites.go` checkCommits: git log error after tag found

### 1.3 version — Edge Cases (64-75% → 85%+)
- [ ] `version.go` getLatestTagFromAllRefs: all tags filtered by match/exclude
- [ ] `version.go` DetectVersion: empty VERSION file
- [ ] `semver.go` parsePreRelease: non-numeric suffix ("beta.alpha")
- [ ] `semver.go` addPreRelease: SetPrerelease failure

### 1.4 bumper — Malformed Input (66-75% → 85%+)
- [ ] `yaml.go` readYAML: malformed YAML error
- [ ] `toml.go` readTOML: malformed TOML error
- [ ] `bumper.go` extractNestedValue: non-map intermediate traversal

### 1.5 changelog — Format Edge Cases (75-84% → 90%+)
- [ ] `conventional.go` formatConventionalEntry: commit with scope + empty hash
- [ ] `conventional.go` formatConventionalEntry: commit without scope + empty hash
- [ ] `parser.go` parseFooterLine: footer with `#` separator
- [ ] `conventional.go` groupBySection: unknown commit type handling

## Bölüm 2: Integration Test Senaryoları

### 2.1 Kritik Eksik Senaryolar
- [ ] CalVer release (yy.mm.minor format, same month increment)
- [ ] Tag format: `${version}` (v prefix olmadan)
- [ ] Tag format: `release-${version}` (custom prefix)
- [ ] Hata: tag already exists
- [ ] Hata: dirty working directory (requireCleanWorkingDir=true)
- [ ] Hata: no commits since last tag (requireCommits=true)
- [ ] Changelog disabled — tag oluşturulur ama CHANGELOG.md oluşturulmaz

### 2.2 Assertion Helper Genişletme
- [ ] `assertCommitExists(t, dir, message)` — commit mesajı ile doğrulama
- [ ] `assertFileContains(t, dir, file, content)` — dosya içerik doğrulama
- [ ] `assertFileNotExists(t, dir, file)` — dosya yokluğu doğrulama
- [ ] `assertWorkingDirClean(t, dir)` — git status temiz doğrulama
- [ ] `assertTagCount(t, dir, count)` — tag sayısı doğrulama

## Test Senaryoları

### CalVer Integration
```
Commits: feat: initial → run with CalVer yy.mm.minor
Expected: tag = 26.3.0 (current year.month.0)
Second run: tag = 26.3.1 (minor increment)
```

### Tag Format Variations
```
Config: tagName = "${version}", commits: feat: new feature
Expected: tag = "1.1.0" (no v prefix)
Git log 1.1.0..HEAD should work for next release
```

### Error: Tag Already Exists
```
Pre-create tag v1.1.0, add feat commit
Expected: error "tag v1.1.0 already exists", no commit created
```

## Riskler

- CalVer testleri tarih-bağımlı — `time.Now()` kullanılıyor, yıl/ay değişiminde kırılabilir
- Tag format testlerinde mevcut `latestVersionToTag` fix'ine bağımlılık
