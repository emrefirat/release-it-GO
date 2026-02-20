# Test Kurallari

## Zorunluluklar

- **Her yeni fonksiyon icin unit test ZORUNLU.**
- Test coverage minimum %70 (%80 hedef).
- Testler basarisiz olursa commit ATMA.
- Race detection: `go test -race ./...`

## Test Isimlendirme

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Arrange - Test verilerini hazirla
    // Act - Test edilecek fonksiyonu cagir
    // Assert - Sonuclari dogrula
}
```

## Table-Driven Tests (Tercih Edilen)

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got = %v, expected %v", got, tt.expected)
            }
        })
    }
}
```

## Mock Pattern (Git Islemleri)

Bu projede git komutlari `commandExecutor` fonksiyon degiskeni uzerinden calisir. Testlerde mock yapilir:

```go
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    return "mock output", nil
}
```

## Test Komutlari

```bash
make test               # Tum testler (-v -cover -race)
make test-unit          # Sadece internal/ testleri
make test-integration   # Sadece test/integration/ testleri
make coverage           # HTML coverage raporu olustur
```

## Yapi

- Unit testler: Her paketin kendi `*_test.go` dosyalari
- Integration testler: `test/integration/` (gercek git repo olusturur)
- Test fixtures: `test/integration/fixtures/` (config ornekleri)
