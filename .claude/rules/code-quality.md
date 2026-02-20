# Kod Kalitesi Kurallari

## Naming Conventions

| Tur | Format | Ornek |
|-----|--------|-------|
| Package | kucuk harf, tek kelime | `config`, `git`, `release` |
| Exported | PascalCase | `CreateRelease`, `UserID` |
| Unexported | camelCase | `generateSlug`, `validateInput` |
| Constants | PascalCase veya SCREAMING_SNAKE | `MaxRetries`, `DEFAULT_TIMEOUT` |
| Interfaces | -er suffix | `Reader`, `Prompter`, `ReleaseProvider` |
| Acronyms | Tutarli buyuk/kucuk | `HTTPServer`, `userID` |

## Error Handling

```go
// YANLIS - Error'u ignore etme
result, _ := someFunction()

// DOGRU - Her error'u handle et
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}
```

- `%w` ile error wrapping kullan.
- Sadece error'u olusturan yerde logla (en ust seviyede: main veya runner).
- `panic` kullanma (sadece kurtarilamaz durumlar icin).

## Custom Error Types

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

## Kod Organizasyonu

- Dosya basina tek sorumluluk.
- Fonksiyonlar 50 satiri gecmemeli (mumkunse).
- Deep nesting yapma (3+ seviye), early return kullan.
- Magic number kullanma, constant tanimla.

## Early Return Pattern

```go
func process(data *Data) error {
    if data == nil {
        return errors.New("nil data")
    }
    if !data.IsValid() {
        return errors.New("invalid data")
    }
    // actual logic
    return nil
}
```

## Yapilmamasi Gerekenler

- `panic` kullanma
- Global state kullanma
- `init()` fonksiyonlarinda side effect
- Circular dependency olusturma
- God objects/functions
- Error'lari ignore etme (`_`)

## Logging

- `log/slog` stdlib kullan.
- Structured logging tercih et.
- Log level: normal / verbose / debug (yapilandirabilir).
- Sensitive veri loglama (sifre, token, PII).

## Performance

- Buyuk slice'lar icin capacity belirt: `make([]Item, 0, len(data))`
- Gereksiz allocation'lardan kacin.
- Goroutine leak'lerden kacin - context kullan.
