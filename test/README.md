# Test Suite - Java Version Manager

Questa directory contiene la suite di test completa per Java Version Manager, progettata per garantire la qualità, l'affidabilità e le performance del software.

## 📁 Struttura Test

```
tests/
├── utils_test.go          # Test per funzioni di utilità
├── jdk_validation_test.go # Test per validazione JDK
├── providers_test.go      # Test per provider (Adoptium, Azul, Liberica)
├── config_test.go         # Test per gestione configurazione
└── integration_test.go    # Test di integrazione workflow completo
```

## 🚀 Esecuzione Test

### Windows

```cmd
# Tutti i test
test.bat all

# Solo unit test
test.bat unit

# Test con coverage
test.bat coverage
```

### Linux/macOS

```bash
# Tutti i test
./test.sh all

# Solo unit test
./test.sh unit

# Test con coverage
./test.sh coverage
```

### Comandi Go Diretti

```bash
# Tutti i test
go test ./tests -v

# Test con coverage
go test ./tests -v -cover

# Benchmark
go test ./tests -bench=. -benchmem

# Test brevi (esclude integrazione)
go test ./tests -v -short
```

## 🧪 Tipologie di Test

### Unit Test

-   **utils_test.go**: Test delle funzioni di parsing versioni, validazione directory JDK
-   **jdk_validation_test.go**: Test della logica di validazione installazioni JDK
-   **providers_test.go**: Test dei provider Adoptium, parsing response, strutture dati
-   **config_test.go**: Test gestione configurazione, repository privati, persistenza

### Integration Test

-   **integration_test.go**: Test workflow completo, operazioni end-to-end
-   Simulazione comandi: `init`, `remote-list`, `download`, `list`, `use`
-   Test di concorrenza e gestione errori

### Benchmark Test

-   Performance parsing versioni
-   Validazione directory JDK
-   Operazioni di configurazione
-   Workflow completo

## 📊 Coverage Report

I test includono analisi di coverage completa:

```bash
# Genera report HTML
test.bat coverage  # Windows
./test.sh coverage # Linux/macOS
```

Il report viene generato in `coverage/coverage.html` e include:

-   Percentuale coverage per file
-   Linee di codice coperte/non coperte
-   Funzioni testate/non testate

## ✅ Test Categories

### Funzioni di Utilità

-   [x] `ParseVersionNumber()` - Parsing versioni Java
-   [x] `IsValidJDKDirectory()` - Validazione struttura JDK
-   [x] `GetJVMVersionsDirectory()` - Path directory versioni
-   [x] `FindSingleJDKInstallation()` - Ricerca installazioni

### Validazione JDK

-   [x] Struttura directory corretta (`bin/`, `lib/`)
-   [x] Presenza `java.exe`
-   [x] JDK multipli e disambiguazione
-   [x] JDK corrotti o incompleti

### Provider

-   [x] Adoptium response parsing
-   [x] Strutture `AdoptiumResponse`, `RecommendedEntry`
-   [x] Parsing versioni complesse (`17.0.8+7`, `1.8.0_452`)
-   [x] Selezione versioni raccomandate (LTS preference)

### Configurazione

-   [x] Creazione, lettura, aggiornamento configurazione
-   [x] Repository privati (endpoint, token)
-   [x] Validazione configurazione
-   [x] Persistenza attraverso restart

### Workflow Integrazione

-   [x] Init → Remote-list → Download → List → Use
-   [x] Gestione errori (versioni inesistenti, JDK corrotti)
-   [x] Operazioni concorrenti
-   [x] Performance end-to-end

## 🔧 Mock e Simulazioni

I test utilizzano mock per:

-   **FileSystem**: Directory temporanee per ogni test
-   **Network**: Response provider simulati senza chiamate reali
-   **Registry**: Simulazione modifiche JAVA_HOME senza UAC
-   **Download**: Creazione JDK mock senza download effettivi

## 📈 Metriche Performance

I benchmark misurano:

-   **Parsing Speed**: Versioni al secondo
-   **Validation Speed**: Directory JDK validate al secondo
-   **Config Operations**: Operazioni configurazione al secondo
-   **Memory Usage**: Allocazioni e garbage collection

Target performance:

-   Parsing: >100k versioni/sec
-   Validation: >1k directory/sec
-   Workflow completo: <1 secondo

## 🐛 Test di Errore

Test specifici per gestione errori:

-   Versioni malformate o inesistenti
-   Directory JDK corrotte
-   Configurazione invalida
-   Operazioni senza privilegi
-   Network timeouts (simulati)

## 🔄 Continuous Integration

I test sono progettati per CI/CD:

-   **Deterministici**: Stesso input = stesso output
-   **Isolati**: Nessuna dipendenza tra test
-   **Veloci**: Unit test <5 secondi, integration <30 secondi
-   **Cross-platform**: Windows, Linux, macOS

## 📝 Come Aggiungere Test

### Nuovo Unit Test

```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := NewFeature(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Nuovo Benchmark

```go
func BenchmarkNewFeature(b *testing.B) {
    input := "test-input"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NewFeature(input)
    }
}
```

## 🎯 Best Practices

1. **Test Naming**: `TestFunctionName` per unit, `TestIntegrationWorkflow` per integration
2. **Table Tests**: Usa struct slice per test multipli
3. **Setup/Teardown**: Usa `t.TempDir()` per isolation
4. **Error Testing**: Test sia success che failure cases
5. **Documentation**: Commenta test complessi
6. **Mock Data**: Usa dati realistici ma controllabili

## 📞 Supporto

Per problemi con i test:

1. Verifica che Go sia installato (`go version`)
2. Esegui `go mod tidy` per dependencies
3. Usa `test.bat verbose` per output dettagliato
4. Controlla logs in `coverage/` per dettagli coverage

---

**Obiettivo**: 100% test coverage per funzioni critiche, >90% overall coverage.
