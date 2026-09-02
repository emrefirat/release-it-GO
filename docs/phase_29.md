# Faz 29: Test Hijyeni ve Dağıtım/CI Güveni (eski Faz 24 kapsamı dahil)

## Özet

İkinci denetimin test-altyapısı bulguları ile ilk denetimden beri bekleyen dağıtım/CI (Faz 24) işlerinin birleşimi. Faz 24 hiç oluşturulmadığı için numaralandırma 23→25 atlamıştı; kapsamı burada devralınır.

## Kapsam ve Durum

### Test hijyeni

| # | İş | Durum |
|---|----|-------|
| 1 | `cli/root_test.go` `Execute --ci` tuzağı: temp dizin + `--dry-run` + gerçek assert (Faz 27.4'te öne alınabilir) | ⬜ |
| 2 | Entegrasyon testlerinde git global-config izolasyonu (`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`, `init.defaultBranch`) — geliştiricinin `commit.gpgsign`/`core.hooksPath` sızmasın | ⬜ |
| 3 | `t.Chdir` (Go 1.26) ile `os.Chdir` + yutulan hatalar temizliği; `commitCounter` global | ⬜ |
| 4 | commit-msg hook uçtan uca: binary derle, `hooks install`, gerçek `git commit` → `--check-msg "$1"` | ⬜ |
| 5 | Pozisyonel arg'lar cobra `Execute` üzerinden gerçek repoda | ⬜ |
| 6 | `RELEASE_*` env gerçek hook'ta; retry GitLab + notification istemci seviyesinde; bumper koruma pipeline üzerinden (bayt-bayt) | ⬜ |
| 7 | Tautolojik/bayat testler: `reasonDescription` (üretimde çağrılmıyor), `HasExpectedFlags` eksikleri, "tag already exists" mesaj assert'i | ⬜ |
| 8 | CI'da coverage kapısı (%70 min) | ⬜ |

### Dağıtım / CI (eski Faz 24)

| # | İş | Durum |
|---|----|-------|
| 9 | Modül adı `github.com/emrefirat/release-it-GO` → `go install` gerçek olsun; README/Dockerfile placeholder'ları | ⬜ |
| 10 | CI security job (govulncheck) yeniden açılır, sürüm pinli | ⬜ |
| 11 | `release.yml`: test koşmadan tag atmasın; tek `git push --atomic origin HEAD "$TAG"` | ⬜ |
| 12 | `.gitignore` `/release-it-go` çapası (şu an `cmd/release-it-go/` altındaki yeni dosyaları gizliyor); `/deneme4/`, `coverage.*` | ⬜ |
| 13 | goreleaser deprecated `format:` → `formats:`; nfpm `license` + `git` bağımlılığı; `before.hooks`'tan `go mod tidy` | ⬜ |
| 14 | `Makefile`: `check` sırası (vuln test'i kapılamasın), `docker-run` `GIT_USER_*` env | ⬜ |
| 15 | dependabot (gomod + github-actions); CI OS matrisi (macos/windows build+unit) | ⬜ |
| 16 | CHANGELOG'un araçla yeniden üretimi (v0.2.0/v0.3.0 girdileri, compare ref'leri) | ⬜ |
