# Phase 9: Conventional Commit Linting

## Ozet

Bu faz, release oncesinde son tag'den bu yana olan commit'lerin conventional commit standardina uygunlugunu kontrol eden bir lint mekanizmasi ekler. Uyumsuz commit varsa release engellenir. Bagimsiz olarak da calistirilabilir.

## Ozellikler

### 1. Commit Lint Kontrolu

- Son tag'den bu yana tum commit'ler conventional commit formatina karsi kontrol edilir
- `config.git.requireConventionalCommits` (bool, default: false) ile aktif edilir
- Pipeline'da `prerequisites` ile `version` adimlari arasinda calisir

### 2. Istisna Kurallari

- `Merge ` ile baslayan commit'ler otomatik gecer (merge commit)
- `Revert ` ile baslayan commit'ler otomatik gecer (revert commit)
- Mevcut `commitPattern` regex'i kullanilir (type(scope)!: description)

### 3. CLI Flag'leri

- `--check-commits`: Sadece lint calistirir, release yapmaz. Hata varsa exit code 1 doner
- `--ignore-commit-lint`: Lint kontrolunu atlar (requireConventionalCommits override)

### 4. Hata Ciktisi

```
Commit lint failed:
  abc1234    fix some bug                              ← not in conventional commit format
  def5678    update readme                             ← not in conventional commit format

  2 of 5 commits are not conventional.
  Use --ignore-commit-lint to bypass.
```

## Yeni Dosyalar

| Dosya | Aciklama |
|-------|----------|
| `internal/changelog/lint.go` | LintInput, LintResult struct'lari ve LintCommits() fonksiyonu |
| `internal/changelog/lint_test.go` | 8 test: conventional, non-conventional, merge, revert, mixed, empty, scoped/breaking, whitespace |

## Degistirilen Dosyalar

| Dosya | Degisiklik |
|-------|-----------|
| `internal/git/changelog.go` | CommitInfo struct + GetCommitsWithHashSinceTag() metodu |
| `internal/git/changelog_test.go` | GetCommitsWithHashSinceTag icin 3 test eklendi |
| `internal/config/config.go` | GitConfig'e RequireConventionalCommits alani eklendi |
| `internal/runner/runner.go` | checkCommitLint() pipeline adimi, RunCheckCommits() modu, formatLintError() yardimci fonksiyonu |
| `internal/cli/root.go` | --check-commits ve --ignore-commit-lint flag'leri |
| `internal/cli/init.go` | Wizard'a "Require conventional commits?" sorusu eklendi |

## Mimari Kararlar

### Circular Dependency Onleme

- `changelog` paketi `git` paketini import eder
- `git` paketi `changelog`'u import edemez (dongusel)
- Lint fonksiyonu `changelog` paketinde kalir ama `git.CommitInfo` yerine kendi `LintInput` struct'ini kullanir
- `runner` zaten her iki paketi de import eder, donusum orada yapilir

### Pipeline Sirasi

```
init → prerequisites → commitlint → version → bump → changelog → git:release → github:release → gitlab:release
```

## Kullanim

```bash
# Config ile aktif etme (.release-it-go.json)
{
  "git": {
    "requireConventionalCommits": true
  }
}

# Normal release (lint aktifse otomatik calisir)
release-it-go --ci

# Sadece lint kontrol
release-it-go --check-commits

# Lint'i atlayarak release
release-it-go --ci --ignore-commit-lint
```
