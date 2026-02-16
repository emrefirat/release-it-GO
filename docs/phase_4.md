# Phase 4: GitHub + GitLab Releases

> **Hedef:** GitHub ve GitLab API entegrasyonu, release olusturma, asset upload, issue/PR/MR comment ve auto-generate notes.

---

## 1. Genel Bakis

Bu faz, GitHub ve GitLab platformlarinda release olusturma, binary/asset upload etme ve ilgili issue/PR'lara comment birakma ozelliklerini implement eder.

---

## 2. Dosya Yapisi

```
internal/
  release/
    github.go             # GitHub API client
    gitlab.go             # GitLab API client
    release.go            # Release interface + ortak tipler
    assets.go             # Asset upload (glob resolve + upload)
    comments.go           # Issue/PR/MR comment
```

---

## 3. Bagimliliklar

| Kutuphane | Kullanim |
|-----------|----------|
| `net/http` (stdlib) | HTTP istekleri |
| `encoding/json` (stdlib) | JSON encode/decode |
| `mime/multipart` (stdlib) | Asset upload |
| `path/filepath` (stdlib) | Glob patterns |

> **Not:** Harici GitHub/GitLab SDK kullanilmayacak. REST API dogrudan `net/http` ile cagrilacak. Bu, bagimliligi minimumda tutar ve her iki platformu tek bir yaklasimla desteklemeyi saglar.

---

## 4. Release Interface

```go
// internal/release/release.go

type ReleaseProvider interface {
    // CreateRelease, yeni bir release olusturur.
    CreateRelease(opts ReleaseOptions) (*ReleaseResult, error)

    // UploadAssets, release'e asset'leri yukler.
    UploadAssets(releaseID string, assets []string) error

    // PostComment, issue/PR/MR'a comment birakir.
    PostComment(target CommentTarget, message string) error

    // ValidateToken, API token'inin gecerli oldugunu kontrol eder.
    ValidateToken() error
}

type ReleaseOptions struct {
    TagName         string
    ReleaseName     string
    ReleaseNotes    string
    Draft           bool
    PreRelease      bool
    MakeLatest      bool
    AutoGenerate    bool   // GitHub auto-generate notes
    DiscussionCategory string // GitHub discussions
}

type ReleaseResult struct {
    ID         string
    URL        string // release web URL
    UploadURL  string // asset upload URL
}

type CommentTarget struct {
    Type   string // "issue", "pr", "mr"
    Number int
}
```

---

## 5. GitHub Client

### 5.1 API Endpoints

| Islem | Method | Endpoint |
|-------|--------|----------|
| Create Release | POST | `/repos/{owner}/{repo}/releases` |
| Upload Asset | POST | `{upload_url}?name={name}` |
| List Issues (for comments) | GET | `/repos/{owner}/{repo}/issues` |
| Comment on Issue | POST | `/repos/{owner}/{repo}/issues/{number}/comments` |
| Comment on PR | POST | `/repos/{owner}/{repo}/issues/{number}/comments` |
| Validate Token | GET | `/user` |

### 5.2 Implementasyon

```go
// internal/release/github.go

type GitHubClient struct {
    config   *config.GitHubConfig
    repoInfo *git.RepoInfo
    logger   *log.Logger
    dryRun   bool
    client   *http.Client
    baseURL  string // default: "https://api.github.com"
    token    string
}

// NewGitHubClient, yeni bir GitHub client olusturur.
// Token, config.TokenRef environment variable'indan okunur.
func NewGitHubClient(cfg *config.GitHubConfig, repoInfo *git.RepoInfo, logger *log.Logger, dryRun bool) (*GitHubClient, error)

func (c *GitHubClient) CreateRelease(opts ReleaseOptions) (*ReleaseResult, error)
func (c *GitHubClient) UploadAssets(releaseID string, assets []string) error
func (c *GitHubClient) PostComment(target CommentTarget, message string) error
func (c *GitHubClient) ValidateToken() error
```

### 5.3 GitHub Enterprise Destegi

```go
// host config'e gore base URL belirleme
func (c *GitHubClient) resolveBaseURL(host string) string {
    if host == "api.github.com" || host == "" {
        return "https://api.github.com"
    }
    // GitHub Enterprise: https://{host}/api/v3
    return fmt.Sprintf("https://%s/api/v3", host)
}
```

### 5.4 Create Release Request Body

```go
type githubCreateReleaseRequest struct {
    TagName                string `json:"tag_name"`
    Name                   string `json:"name"`
    Body                   string `json:"body"`
    Draft                  bool   `json:"draft"`
    Prerelease             bool   `json:"prerelease"`
    MakeLatest             string `json:"make_latest,omitempty"`
    GenerateReleaseNotes   bool   `json:"generate_release_notes,omitempty"`
    DiscussionCategoryName string `json:"discussion_category_name,omitempty"`
}

type githubCreateReleaseResponse struct {
    ID        int    `json:"id"`
    HTMLURL   string `json:"html_url"`
    UploadURL string `json:"upload_url"`
}
```

---

## 6. GitLab Client

### 6.1 API Endpoints

| Islem | Method | Endpoint |
|-------|--------|----------|
| Create Release | POST | `/api/v4/projects/{id}/releases` |
| Upload Asset (Generic Package) | PUT | `/api/v4/projects/{id}/packages/generic/{pkg}/{version}/{filename}` |
| Create Release Link | POST | `/api/v4/projects/{id}/releases/{tag}/assets/links` |
| Comment on MR | POST | `/api/v4/projects/{id}/merge_requests/{mr_iid}/notes` |
| Comment on Issue | POST | `/api/v4/projects/{id}/issues/{issue_iid}/notes` |
| Validate Token | GET | `/api/v4/user` |

### 6.2 Implementasyon

```go
// internal/release/gitlab.go

type GitLabClient struct {
    config    *config.GitLabConfig
    repoInfo  *git.RepoInfo
    logger    *log.Logger
    dryRun    bool
    client    *http.Client
    baseURL   string
    token     string
    projectID string // URL-encoded "owner/repo"
}

// NewGitLabClient, yeni bir GitLab client olusturur.
func NewGitLabClient(cfg *config.GitLabConfig, repoInfo *git.RepoInfo, logger *log.Logger, dryRun bool) (*GitLabClient, error)

func (c *GitLabClient) CreateRelease(opts ReleaseOptions) (*ReleaseResult, error)
func (c *GitLabClient) UploadAssets(releaseID string, assets []string) error
func (c *GitLabClient) PostComment(target CommentTarget, message string) error
func (c *GitLabClient) ValidateToken() error
```

### 6.3 Create Release Request Body

```go
type gitlabCreateReleaseRequest struct {
    Name        string   `json:"name"`
    TagName     string   `json:"tag_name"`
    Description string   `json:"description"`
    Milestones  []string `json:"milestones,omitempty"`
}
```

### 6.4 GitLab Asset Upload Akisi

GitLab'da asset upload iki adimlidir:
1. Dosyayi Generic Package Repository'ye yukle
2. Release'e asset link olustur

```go
func (c *GitLabClient) UploadAssets(releaseID string, assets []string) error {
    for _, assetPath := range assets {
        // Step 1: Upload to Generic Package Repository
        packageURL, err := c.uploadToGenericPackage(assetPath)
        if err != nil {
            return fmt.Errorf("uploading asset %s: %w", assetPath, err)
        }

        // Step 2: Create release link
        err = c.createReleaseLink(releaseID, filepath.Base(assetPath), packageURL)
        if err != nil {
            return fmt.Errorf("creating release link for %s: %w", assetPath, err)
        }
    }
    return nil
}
```

---

## 7. Asset Upload

```go
// internal/release/assets.go

// ResolveAssets, glob pattern'lerinden dosya listesi olusturur.
func ResolveAssets(patterns []string) ([]string, error) {
    var files []string
    for _, pattern := range patterns {
        matches, err := filepath.Glob(pattern)
        if err != nil {
            return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
        }
        files = append(files, matches...)
    }
    return files, nil
}

// DetectContentType, dosyanin MIME type'ini belirler.
func DetectContentType(filePath string) string
```

---

## 8. Token Yonetimi

```go
// Token okuma sirasi:
// 1. Config'deki tokenRef'den env variable oku
// 2. Bulunamazsa hata dondur (skipChecks=true degilse)

func getToken(tokenRef string, skipChecks bool) (string, error) {
    token := os.Getenv(tokenRef)
    if token == "" && !skipChecks {
        return "", fmt.Errorf("environment variable %s is not set; "+
            "create a token and set it as %s", tokenRef, tokenRef)
    }
    return token, nil
}
```

---

## 9. Proxy Destegi (GitHub)

```go
func (c *GitHubClient) createHTTPClient() *http.Client {
    transport := &http.Transport{}
    if c.config.Proxy != "" {
        proxyURL, _ := url.Parse(c.config.Proxy)
        transport.Proxy = http.ProxyURL(proxyURL)
    }
    return &http.Client{
        Transport: transport,
        Timeout:   time.Duration(c.config.Timeout) * time.Second,
    }
}
```

---

## 10. TLS/Certificate Destegi (GitLab)

```go
func (c *GitLabClient) createHTTPClient() *http.Client {
    tlsConfig := &tls.Config{}

    if c.config.CertificateAuthorityFile != "" {
        caCert, err := os.ReadFile(c.config.CertificateAuthorityFile)
        if err == nil {
            caCertPool := x509.NewCertPool()
            caCertPool.AppendCertsFromPEM(caCert)
            tlsConfig.RootCAs = caCertPool
        }
    }

    if !c.config.Secure {
        tlsConfig.InsecureSkipVerify = true
    }

    return &http.Client{
        Transport: &http.Transport{TLSClientConfig: tlsConfig},
    }
}
```

---

## 11. Release Notes Kaynaklari

Release notes su kaynaklardan gelir (oncelik sirasiyla):

1. **autoGenerate** (GitHub only): GitHub'in kendi auto-generate ozelligi
2. **releaseNotes** (shell command): Config'deki komut calistirilir, stdout kullanilir
3. **releaseNotes** (template): Template string, variable'lar replace edilir
4. **changelog**: Phase 3'te olusturulan changelog icerigi
5. **Basit git log**: Phase 2'deki basit changelog

```go
// getReleaseNotes, oncelik sirasina gore release notes belirler.
func getReleaseNotes(cfg *config.Config, changelog string, vars map[string]string) (string, error)
```

---

## 12. Dry-Run Destegi

Tum API cagrilari dry-run modunda loglanir, calistirilmaz:

```go
func (c *GitHubClient) CreateRelease(opts ReleaseOptions) (*ReleaseResult, error) {
    if c.dryRun {
        c.logger.DryRun("POST %s/repos/%s/%s/releases", c.baseURL, c.repoInfo.Owner, c.repoInfo.Repository)
        c.logger.DryRun("  tag_name: %s", opts.TagName)
        c.logger.DryRun("  name: %s", opts.ReleaseName)
        return &ReleaseResult{URL: "(dry-run)"}, nil
    }
    // actual API call...
}
```

---

## 13. Hata Senaryolari

| Senaryo | Hata Mesaji |
|---------|-------------|
| Token yok | `"environment variable GITHUB_TOKEN is not set"` |
| Gecersiz token | `"GitHub token is invalid (HTTP 401)"` |
| Repo bulunamadi | `"repository %s/%s not found (HTTP 404)"` |
| Rate limit | `"GitHub API rate limit exceeded, retry after %s"` |
| Asset bulunamadi | `"asset file not found: %s"` |
| Upload basarisiz | `"failed to upload asset %s: %w"` |
| Tag yok | `"tag %s does not exist on remote"` |
| Network hatasi | `"failed to connect to %s: %w"` |

---

## 14. Kabul Kriterleri

- [ ] GitHub token env variable'dan okunuyor
- [ ] GitHub token validation calisiyor
- [ ] GitHub release olusturuluyor (tag, name, notes, draft, preRelease)
- [ ] GitHub Enterprise URL destegi calisiyor
- [ ] GitHub asset upload calisiyor (glob pattern destegi)
- [ ] GitHub issue/PR comment calisiyor (template destegi)
- [ ] GitHub auto-generate notes calisiyor
- [ ] GitHub discussion category destegi
- [ ] GitHub makeLatest secenegi calisiyor
- [ ] GitLab token env variable'dan okunuyor
- [ ] GitLab release olusturuluyor
- [ ] GitLab asset upload calisiyor (Generic Package Repository)
- [ ] GitLab milestone association calisiyor
- [ ] GitLab MR/issue comment calisiyor
- [ ] GitLab custom CA certificate destegi
- [ ] Proxy destegi calisiyor
- [ ] Dry-run modunda API cagrisi yapilmiyor
- [ ] Tum hata senaryolari anlamli mesajlarla handle ediliyor
- [ ] Rate limiting handle ediliyor (retry veya bilgilendirme)
- [ ] `go test ./internal/release/... -race` basarili
- [ ] Test coverage %70+

---

## 15. Test Senaryolari

> API testleri `httptest.NewServer` ile mock HTTP server kullanarak yapilacak.

### GitHub Tests
- Token validation: gecerli -> basarili
- Token validation: gecersiz -> 401 hatasi
- Token validation: skipChecks -> atlanir
- Create release: tam parametreler
- Create release: draft, preRelease
- Create release: autoGenerate
- Upload asset: tek dosya
- Upload asset: glob pattern -> birden fazla dosya
- Upload asset: dosya bulunamadi -> hata
- Post comment: issue
- Post comment: PR
- GitHub Enterprise URL: dogru base URL
- Proxy: transport'a ekleniyor
- Dry-run: API cagrisi yok

### GitLab Tests
- Token validation (Private-Token header)
- Create release: tam parametreler
- Create release: milestones
- Upload asset: Generic Package -> Release Link (iki asama)
- Post comment: MR
- Post comment: Issue
- Custom CA certificate
- Dry-run: API cagrisi yok

### Asset Tests
- Glob resolve: `dist/*.zip` -> dosya listesi
- Glob resolve: bos pattern -> bos liste
- Glob resolve: gecersiz pattern -> hata
- Content type detection: .zip, .tar.gz, .dmg

### Release Notes Tests
- autoGenerate onceligi
- Shell command calistirma
- Template rendering
- Fallback to changelog
