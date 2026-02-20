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
- Dependency'leri guncel tut (`make vuln` ile govulncheck)
- HTTPS URL'lerde credential stripping uygula

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

### Docker Guvenligi
- Non-root user (`releaser:1000`)
- Static binary (CGO_ENABLED=0)
- Minimal base image (alpine)
- Git identity env var zorunlulugu (release islemleri icin)
