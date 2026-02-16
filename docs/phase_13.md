# Faz 13: Webhook Notification Desteği (Slack + Teams)

## Özet

Release tamamlandıktan sonra ekibi bilgilendirmek için Slack ve Microsoft Teams webhook notification desteği. Pipeline'ın son adımı olarak çalışır, non-fatal yapıdadır (başarısız olursa uyarı verir, pipeline durmaz).

## Gereksinimler

### Fonksiyonel

- Slack incoming webhook desteği
- Microsoft Teams MessageCard webhook desteği
- Birden fazla webhook endpoint'e aynı anda gönderi
- Özelleştirilebilir mesaj şablonu (`${version}`, `${changelog}`, `${releaseUrl}` vb.)
- Platform-spesifik default şablonlar
- Dry-run modunda HTTP çağrısı yapılmaz

### Güvenlik

- Webhook URL'leri doğrudan config'e yazılmaz
- `urlRef` ile environment variable adı belirtilir (tokenRef pattern'i ile tutarlı)

### Hata Yönetimi

- Notification hataları non-fatal: pipeline durmaz, uyarı loglanır
- Birden fazla webhook varsa, biri başarısız olsa da diğerleri denenir
- Configurable timeout (default: 30 saniye)

## Konfigürasyon

```json
{
  "notification": {
    "enabled": true,
    "webhooks": [
      {
        "type": "slack",
        "urlRef": "SLACK_WEBHOOK_URL"
      },
      {
        "type": "teams",
        "urlRef": "TEAMS_WEBHOOK_URL",
        "messageTemplate": "🎉 ${repo.repository} v${version} yayınlandı!\n${releaseUrl}",
        "timeout": 15
      }
    ]
  }
}
```

## Template Değişkenleri

| Değişken | Açıklama |
|----------|----------|
| `${version}` | Yeni versiyon (örn: 1.2.0) |
| `${latestVersion}` | Önceki versiyon |
| `${tagName}` | Git tag adı (örn: v1.2.0) |
| `${changelog}` | Changelog içeriği |
| `${releaseUrl}` | GitHub/GitLab release URL'i |
| `${branchName}` | Branch adı |
| `${repo.repository}` | Repository adı |
| `${repo.owner}` | Repository sahibi |
| `${repo.host}` | Repository host'u |

## Pipeline Sırası

```
init → prerequisites → commitlint → version → bump → changelog → git:release → github:release → gitlab:release → notification
```

## Payload Formatları

### Slack
```json
{"text": "🚀 *my-repo* v1.2.0 released!\nhttps://github.com/..."}
```

### Teams (MessageCard)
```json
{
  "@type": "MessageCard",
  "summary": "🚀 my-repo v1.2.0 released!\nhttps://github.com/...",
  "text": "🚀 my-repo v1.2.0 released!\nhttps://github.com/..."
}
```

## Dosyalar

| Dosya | Açıklama |
|-------|----------|
| `internal/config/config.go` | NotificationConfig, WebhookConfig struct'ları |
| `internal/config/defaults.go` | Default notification config |
| `internal/notification/notification.go` | Client, SendAll, HTTP POST mantığı |
| `internal/notification/slack.go` | Slack payload builder |
| `internal/notification/teams.go` | Teams MessageCard builder |
| `internal/notification/notification_test.go` | Unit testler (%98+ coverage) |
| `internal/runner/runner.go` | sendNotification pipeline adımı |
| `internal/runner/runner_test.go` | Runner notification testleri |
