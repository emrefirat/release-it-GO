# Phase 11: Docker Container Desteği

## Özet

Bu faz, release-it-go'nun Docker container içinde çalışabilmesini sağlar. Kullanıcılar Go kurulumu veya binary derleme gerektirmeden, proje dizinini mount ederek doğrudan release işlemlerini yapabilir. CI/CD pipeline'larında container image olarak kullanılabilir.

## Özellikler

### 1. Multi-stage Dockerfile

- **Builder stage:** `golang:1.24.3-alpine` ile static binary derleme
- **Runtime stage:** `alpine:3.21` minimal base image (~30MB toplam)
- `CGO_ENABLED=0 GOOS=linux` ile static binary (musl/glibc bağımlılığı yok)
- `-trimpath -ldflags="-s -w"` ile küçültülmüş binary
- Build ARG'lar: `VERSION`, `COMMIT`, `BUILD_DATE`, `USER_UID`, `USER_GID`

### 2. Runtime Ortamı

- **Git:** Container içinde git işlemleri için
- **OpenSSH client:** Private repo erişimi ve SSH agent forwarding
- **CA certificates:** HTTPS bağlantıları için
- **Non-root user:** `releaser` (UID/GID 1000, build arg ile değiştirilebilir)
- **Safe directory:** `git config --global --add safe.directory '*'` (mount edilen repo'lar için)

### 3. .dockerignore

- Gereksiz dosyaların build context'ten çıkarılması
- Hızlı build ve küçük context boyutu

### 4. Makefile Entegrasyonu

- `make docker-build` ve `make docker-run` target'ları
- Version/commit/date bilgisi otomatik enjeksiyon

## Kullanım Örnekleri

### Temel Kullanım

```bash
# Build
docker build -t release-it-go .

# Dry-run
docker run --rm -v $(pwd):/workspace release-it-go --dry-run

# CI modu
docker run --rm -v $(pwd):/workspace -e GITHUB_TOKEN=$GITHUB_TOKEN release-it-go --ci

# İnteraktif mod
docker run --rm -it -v $(pwd):/workspace -e GITHUB_TOKEN=$GITHUB_TOKEN release-it-go
```

### SSH Agent Forwarding (Private Repo)

```bash
docker run --rm -v $(pwd):/workspace \
  -v $SSH_AUTH_SOCK:/ssh-agent -e SSH_AUTH_SOCK=/ssh-agent \
  -e GITHUB_TOKEN=$GITHUB_TOKEN release-it-go --ci
```

### Custom UID/GID (Linux Permission Fix)

```bash
docker build --build-arg USER_UID=$(id -u) --build-arg USER_GID=$(id -g) -t release-it-go:local .
```

## CI/CD Entegrasyon Örnekleri

### GitHub Actions

```yaml
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Release
        run: |
          docker run --rm \
            -v ${{ github.workspace }}:/workspace \
            -e GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }} \
            release-it-go --ci
```

### GitLab CI

```yaml
release:
  image: release-it-go:latest
  script:
    - release-it-go --ci
  variables:
    GITLAB_TOKEN: $CI_JOB_TOKEN
  only:
    - main
```

## Yeni Dosyalar

| Dosya | Açıklama |
|-------|----------|
| `Dockerfile` | Multi-stage build (builder + runtime) |
| `.dockerignore` | Build context optimizasyonu |

## Değiştirilen Dosyalar

| Dosya | Değişiklik |
|-------|-----------|
| `Makefile` | `docker-build` ve `docker-run` target'ları eklendi |

## Build ARG'lar

| ARG | Default | Açıklama |
|-----|---------|----------|
| `VERSION` | `dev` | Binary versiyon bilgisi |
| `COMMIT` | `none` | Git commit hash |
| `BUILD_DATE` | `unknown` | Build tarihi |
| `USER_UID` | `1000` | Container user UID |
| `USER_GID` | `1000` | Container user GID |

## Tahmini Image Boyutu

| Katman | Boyut |
|--------|-------|
| alpine:3.21 base | ~7MB |
| git + openssh-client + ca-certificates | ~15MB |
| release-it-go binary | ~8MB |
| **Toplam** | **~30MB** |

## Doğrulama

```bash
# Build
docker build -t release-it-go:test .

# Version kontrolü
docker run --rm release-it-go:test version

# Help çıktısı
docker run --rm release-it-go:test --help

# Non-root user kontrolü
docker run --rm --entrypoint sh release-it-go:test -c "whoami"
# → releaser

# Image boyutu kontrolü
docker images release-it-go:test
# → ~30MB bekleniyor

# Gerçek repo ile dry-run
docker run --rm -v $(pwd):/workspace release-it-go:test --ci --dry-run
```
