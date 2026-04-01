# Faz 20: Git Hook Kurulum Komutu (release-it-go install)

## Özet

`release-it-go install` komutu ile config'deki hook tanımlarını local `.git/hooks/` dizinine otomatik kuran özellik. Kullanıcılar `yarn install` + husky deneyimini Go projelerinde de yaşayacak.

## Problem

npm projelerinde `yarn install` ile husky/lefthook otomatik olarak pre-commit, commit-msg gibi git hook'larını kuruyor. Go projelerinde bu altyapı yok. Geliştiriciler her projede elle hook oluşturmak zorunda.

## Çözüm

Mevcut `hooks` config section'ı genişletilerek git hook'ları da desteklenecek. Yeni config section **açılmayacak** — isim formatı ayrımı yeterli:
- `before:X` / `after:X` → release pipeline hook'ları (mevcut)
- `pre-commit`, `commit-msg`, `pre-push` → git hook'ları (yeni)

## Kullanım

```yaml
hooks:
  # Release pipeline hooks (mevcut)
  "after:bump": ["echo bumped to v${version}"]

  # Git hooks (release-it-go install ile .git/hooks/'a yazılır)
  "pre-commit": ["go fmt ./...", "go vet ./..."]
  "commit-msg": ["./release-it-go --check-commits"]
  "pre-push": ["go test ./..."]
```

```bash
release-it-go install          # Hook'ları kur
release-it-go install --force  # Mevcut hook'ları da üzerine yaz
release-it-go install --remove # Yönetilen hook'ları kaldır
```

## Üretilen Script Formatı

```bash
#!/bin/sh
# Managed by release-it-go — DO NOT EDIT
# Hook: pre-commit

set -e

go fmt ./...
go vet ./...
```

## Desteklenen Git Hook'ları

`pre-commit`, `commit-msg`, `pre-push`, `post-commit`, `post-merge`, `prepare-commit-msg`

## Yapılacaklar

- [ ] `internal/config/config.go` — HooksConfig'e git hook alanları ekle (PreCommit, CommitMsg, PrePush, PostCommit, PostMerge, PrepareCommitMsg)
- [ ] `internal/githook/githook.go` — Installer (Install, Remove, generateScript, isManagedHook, HooksFromConfig)
- [ ] `internal/githook/githook_test.go` — Unit testler (~12 test)
- [ ] `internal/cli/install.go` — Cobra subcommand (--force, --remove)
- [ ] `internal/cli/root.go` — newInstallCommand() kayıt
- [ ] `internal/cli/install_test.go` — CLI testleri (~5 test)
- [ ] `internal/config/writer.go` — fullExampleYAML'a git hook örnekleri
- [ ] Tüm testlerin geçtiğini doğrula (make check)

## Hata Senaryoları

| Senaryo | Davranış |
|---------|----------|
| `.git` yok | Hata: "not a git repository" |
| `.git/hooks/` yok | Otomatik oluştur |
| Mevcut kullanıcı hook'u | `--force` olmadan hata |
| Config'de git hook yok | Bilgi mesajı |
| Yetki hatası | OS hatasını wrap et |

## Kapsam Dışı

- Maven plugin (ayrı Java repo)
- npm package (ayrı repo)
- Binary self-download
