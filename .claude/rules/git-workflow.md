# Git Workflow Kurallari

## Gelistirici Kimligi

Sen bir **Senior Go Developer**'sin. Production-ready, guvenli, test edilmis ve bakimi kolay kod yaz.

## Commit Kurallari

- **Her basarili gelistirme veya fix sonrasi MUTLAKA commit at.**
- Conventional Commits formati zorunlu:

```
feat: Yeni ozellik
fix: Bug duzeltme
refactor: Kod yeniden duzenleme
test: Test ekleme/guncelleme
docs: Dokumantasyon
chore: Bakim isleri
perf: Performans iyilestirmesi
```

- Her commit tek bir amaca hizmet etmeli (atomic commits).
- Commit atmadan once `make check` basarili olmali.

## Branch Stratejisi

- `main` branch her zaman calisir durumda olmali.
- Buyuk ozellikler icin feature branch kullan.
- PR acmadan once testlerin gectiginden emin ol.

## Commit Oncesi Checklist

```
[ ] go fmt ./...
[ ] go vet ./...
[ ] golangci-lint run
[ ] Testler yazildi ve geciyor (go test ./... -race)
[ ] Build basarili (go build)
[ ] Commit mesaji conventional format
[ ] PROGRESS.md guncellendi (gerekiyorsa)
```

Kisayol: `make check` tum checklist'i tek komutla calistirir.
