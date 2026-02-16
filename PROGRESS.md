# PROGRESS.md - release-it-go Proje Ilerleme Takibi

> Bu dosya, projenin genel ilerlemesini ve her fazin durumunu takip eder.
> Her gelistirme oturumu sonunda guncellenmelidir.

---

## Genel Durum

| Faz | Baslik | Durum | Ilerleme |
|-----|--------|-------|----------|
| 1 | Core Foundation | Baslanmadi | 0% |
| 2 | Git Operations | Baslanmadi | 0% |
| 3 | Conventional Commits + Changelog | Baslanmadi | 0% |
| 4 | GitHub + GitLab Releases | Baslanmadi | 0% |
| 5 | Interactive UI + Hooks + Pipeline | Baslanmadi | 0% |
| 6 | Advanced Features | Baslanmadi | 0% |
| 7 | Testing, CI/CD, Documentation | Baslanmadi | 0% |

**Son Guncelleme:** 2026-02-16
**Aktif Gelistirici:** -
**Mevcut Versiyon:** Henuz baslanmadi

---

## Faz 1: Core Foundation

**Durum:** Baslanmadi
**PRD:** `docs/phase_1.md`

### Yapilacaklar

- [ ] Go module init (`go mod init`)
- [ ] Cobra CLI iskeleti
- [ ] Config struct tanimlari
- [ ] Config loader (JSON/YAML/TOML)
- [ ] Default degerler
- [ ] CLI flags -> config merge
- [ ] Git tag'den versiyon okuma
- [ ] VERSION dosyasindan versiyon okuma
- [ ] Semver parse/increment/compare
- [ ] Template variable rendering
- [ ] Logger (verbose seviyeleri)
- [ ] Makefile
- [ ] Unit testler

### Notlar

-

---

## Faz 2: Git Operations

**Durum:** Baslanmadi
**PRD:** `docs/phase_2.md`

### Yapilacaklar

- [ ] Git runner (komut calistirma)
- [ ] Prerequisite checks (branch, clean, upstream, commits)
- [ ] Stage + Commit
- [ ] Tag olusturma
- [ ] Push
- [ ] Repo info parse (HTTPS + SSH)
- [ ] Basit git log changelog
- [ ] Dry-run destegi
- [ ] Unit testler

### Notlar

-

---

## Faz 3: Conventional Commits + Changelog

**Durum:** Baslanmadi
**PRD:** `docs/phase_3.md`

### Yapilacaklar

- [ ] Conventional commit parser
- [ ] Bump analyzer (major/minor/patch)
- [ ] Conventional-changelog formati
- [ ] Keep-a-changelog formati
- [ ] CHANGELOG.md dosya guncelleme
- [ ] Unit testler

### Notlar

-

---

## Faz 4: GitHub + GitLab Releases

**Durum:** Baslanmadi
**PRD:** `docs/phase_4.md`

### Yapilacaklar

- [ ] Release provider interface
- [ ] GitHub client (create, upload, comment)
- [ ] GitLab client (create, upload, comment)
- [ ] Token yonetimi
- [ ] Asset upload (glob)
- [ ] GitHub Enterprise destegi
- [ ] GitLab CA certificate destegi
- [ ] Dry-run destegi
- [ ] API mock testleri

### Notlar

-

---

## Faz 5: Interactive UI + Hooks + Pipeline

**Durum:** Baslanmadi
**PRD:** `docs/phase_5.md`

### Yapilacaklar

- [ ] Versiyon secim prompt
- [ ] Onay prompt'lari
- [ ] Spinner animasyonu
- [ ] Renkli cikti
- [ ] CI ortam algilama
- [ ] Hook runner (before/after lifecycle)
- [ ] Ana pipeline orchestrator
- [ ] Ozel modlar (--changelog, --release-version, --only-version)
- [ ] Ozet ciktisi
- [ ] Unit testler

### Notlar

-

---

## Faz 6: Advanced Features

**Durum:** Baslanmadi
**PRD:** `docs/phase_6.md`

### Yapilacaklar

- [ ] Bumper: dosyadan versiyon okuma (JSON/YAML/TOML/INI/text)
- [ ] Bumper: dosyaya versiyon yazma
- [ ] Bumper: glob pattern destegi
- [ ] CalVer implementasyonu
- [ ] Pre-release flows
- [ ] --no-increment modu
- [ ] Verbose/debug modlari
- [ ] Unit testler

### Notlar

-

---

## Faz 7: Testing, CI/CD, Documentation

**Durum:** Baslanmadi
**PRD:** `docs/phase_7.md`

### Yapilacaklar

- [ ] Integration testler
- [ ] API mock testleri
- [ ] Coverage %80+ hedefi
- [ ] GitHub Actions CI workflow
- [ ] GitHub Actions Release workflow
- [ ] GoReleaser config
- [ ] Shell completions (bash/zsh/fish)
- [ ] Build info (ldflags)

### Notlar

-

---

## Degisiklik Gecmisi

| Tarih | Gelistirici | Degisiklik |
|-------|------------|------------|
| 2026-02-16 | - | Proje baslatildi, PRD dosyalari olusturuldu |

---

## Kurallar

1. **Her oturum sonunda bu dosyayi guncelle.**
2. Tamamlanan maddeler `[x]` ile isaretlenir.
3. Yeni eklenen maddeler `[ ]` ile eklenir.
4. Durum alani guncellenir: `Baslanmadi` / `Devam Ediyor` / `Tamamlandi`
5. Ilerleme yuzdesi guncellenir.
6. Notlar bolumune onemli kararlar, engeller veya degisiklikler yazilir.
7. Degisiklik gecmisi tablosuna yeni satirlar eklenir.
