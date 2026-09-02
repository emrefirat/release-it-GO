# Faz 30: Doküman Senkronu

## Özet

İkinci denetim 29 **yanlış** doküman ifadesi ve Faz 23–26'da eklenen davranışlar için eksik bölümler tespit etti. Öncelik: yanlış olanlar (eksik olanlardan önce).

## Yanlış İfadeler (öncelikli)

| # | Yer | Yanlış | Doğru | Durum |
|---|-----|--------|-------|-------|
| 1 | `writer.go` fullExampleYAML, `install.go` help | Hook'lar `.git/hooks/`'a kurulur; commit-msg örneği `--check-commits` | `.hooks/` + `core.hooksPath`; `--check-msg "$1"` (Faz 27'de öne alınabilir) | ⬜ |
| 2 | README `go install release-it-go/...`, `github.com/user/release-it-go` | Çalışmaz / placeholder | Faz 29.9 sonrası gerçek yol | ⬜ |
| 3 | README CalVer `26.02.0` | Kod `26.9.0` üretiyor | Örnek ya da kod (ürün kararı) | ⬜ |
| 4 | README `${latestTag}` değişkeni | Yok | `${branchName}`, `${repo.host}`, `${repo.protocol}`, `RELEASE_*` | ⬜ |
| 5 | README hook shell "sh -c" | Windows `%COMSPEC% /C` | İkisi | ⬜ |
| 6 | README `--check-commits` örnek çıktısı | Eski format | Yeni format + `-V` davranışı | ⬜ |
| 7 | README `addUntrackedFiles` açıklaması, "tag adlarında değişken" | Staging semantiği / yalnız `${version}` | Faz 27.14 sonucuna göre | ⬜ |
| 8 | TROUBLESHOOTING Docker `GIT_AUTHOR_*` | Entrypoint `GIT_USER_*` bekliyor | `GIT_USER_NAME/EMAIL` | ⬜ |
| 9 | TROUBLESHOOTING "tag already exists" (local or remote), `--no-increment` "rarely needed" | Yalnız local; kurtarma yolu | Yeni davranış + kurtarma | ⬜ |
| 10 | CONTRIBUTING "Go 1.26.3+", "CI aynı kontrolleri koşar", "kendini release ediyor" | 1.26.7; CI'da vuln yok; v0.2.0/v0.3.0 o akıştan değil | Gerçek durum | ⬜ |
| 11 | ARCHITECTURE "4xx/5xx → pipeline durur", "rate limit yok", hook listesi, sürüm algoritması, renderer'lar | Retry, before/after:release, çıkarım, explicit version, `httputil` | Güncel akış | ⬜ |
| 12 | CLAUDE.md "Phase 20", "tek pipelineStep slice", iki commandExecutor; PROGRESS "v0.1.3"; ARCHITECTURE/DECISIONS tarihleri | Faz 26; üç adım listesi + runSteps; üçüncü seam `execCommand` | Güncel | ⬜ |

## Eksik Bölümler

README: pozisyonel sürüm (`[increment]`), `--check-msg` ve commit-msg hook, `hooks install|remove` (`.hooks/`, `--force`, prune, `core.hooksPath`), hook anahtar adlandırma (kebab-case git hook'ları vs camelCase config vs `before:`/`after:`), `RELEASE_*` env (11 değişken), v-önek çıkarımı, `requireBranch` virgüllü desenler, `certificateAuthorityFileRef`, retry semantiği, bumper format koruma, bağımsız onaylar, üst-düzey `ci/dry-run/verbose` anahtarları, Docker kullanımı. DECISIONS: 4 ADR (retry politikası, bumper ispat-tabanlı splice, TLS varsayılanı, hook shell/anahtar adlandırma). ARCHITECTURE/CLAUDE: `internal/httputil`.
