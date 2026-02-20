# Guvenlik Kurallari

## ASLA Yapma

- Sifreleri/secret'lari plain text saklama veya koda gomme
- Kullanici girdisini dogrulamadan kullanma
- Sensitive bilgileri loglama (sifre, token, PII)
- Hardcoded credentials
- Panic'i error handling yerine kullanma
- Error mesajlarinda internal detay verme

## Her Zaman Yap

- Input validation uygula
- Token'lar icin environment variable kullan (`tokenRef` pattern)
- Dependency'leri guncel tut
- HTTPS URL'lerde credential stripping uygula

## Guvenlik Araclari

Commit oncesi `make check` calistirilmali (govulncheck dahil). Ek olarak:

```bash
# Dependency vulnerability taramasi (Makefile'da `make vuln`)
govulncheck ./...

# Statik guvenlik analizi - guvenlik suphesi veya yeni ozellik eklendiginde calistir
gosec ./...

# Genel statik analiz (Makefile'da `make lint`)
golangci-lint run
```

### Ne Zaman gosec Kullanilmali
- Yeni HTTP client, dosya islemleri veya exec.Command kullanan kod eklendiginde
- External input isleme mantigi degistiginde
- Review sirasinda guvenlik endisesi olustugunda
- Release oncesi son kontrol olarak

## Projede Kullanilan Guvenlik Paternleri

### Token Yonetimi
Token'lar asla config'e yazilmaz. `tokenRef` ile env var adi belirtilir:
```json
{
  "github": {
    "tokenRef": "GITHUB_TOKEN"
  }
}
```

### Webhook URL Guvenligi
Webhook URL'leri `urlRef` ile env var uzerinden alinir, config'e direkt yazilmaz.

### Command Execution Guvenligi
Git komutlari `exec.Command("git", args...)` ile calistirilir. Shell uzerinden (`sh -c`) gecilmez, bu command injection riskini onler.

### Docker Guvenligi
- Non-root user (`releaser:1000`)
- Static binary (CGO_ENABLED=0)
- Minimal base image (alpine)
- Git identity env var zorunlulugu (release islemleri icin)
