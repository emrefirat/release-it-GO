# Ilerleme Takibi ve Bug Yonetimi

## PROGRESS.md Zorunlulugu

- **Her gelistirme oturumu sonunda `PROGRESS.md` dosyasini guncelle.**
- Tamamlanan maddeleri `[x]` ile isaretle.
- Ilerleme yuzdesini guncelle.
- Notlar bolumune onemli kararlar ve engelleri yaz.
- Son guncelleme tarihini guncelle.

## Yeni Faz Eklerken

1. `docs/phase_N.md` PRD dosyasi olustur (mevcut phase dosyalarinin formatini takip et).
2. `PROGRESS.md`'deki faz tablosuna yeni satir ekle.
3. Faz detay bolumunu ekle (yapilacaklar, notlar).
4. Degisiklik gecmisine yeni satir ekle.

## Bug Takibi

- Tespit edilen her bug `PROGRESS.md` Bugs bolumune eklenir.
- Acik bug: `- [ ] BUG: Kisa aciklama (tarih)`
- Kapali bug: `- [x] BUG: Kisa aciklama (tarih) → cozum ozeti`
- Bug fix commit'i sonrasi satir `[x]` ile isaretlenir.
- Edge case'ler ve ilk release senaryolari mutlaka test edilmeli.

## Oturum Sonu Sureci

1. Yapilan degisiklikleri PROGRESS.md'ye yaz.
2. Bug varsa Bugs bolumune ekle.
3. Degisiklik gecmisine tarih + ozet ekle.
4. Commit at (conventional format).
