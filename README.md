# ☕ Java Version Manager (JVM)

U## 🚀 Funzionalità principali

-   🔍 Elenco JDK da **Adoptium**, **Azul**, **Liberica** e **repository privati**
-   📦 **Download e gestione JDK** con organizzazione automatica in cartelle
-   📋 **Lista versioni installate** con dettagli su dimensioni e stato
-   🧠 Selezione intelligente di una versione per tag (LTS → usata → patch)
-   📊 Visualizzazione tabellare con info su OS, Architettura, link di download
-   ⚡ **Autocompletamento bash** con Tab per tutti i comandi e opzioni
-   🔧 **Strumenti di sistema** per pulizia PATH e manutenzione
-   📄 Banner Figlet all'avvio + spiegazione del comportamento
-   🎛️ Supporto a flag avanzati:
    -   `--provider`, `--all`, `--major-only`, `--latest`, `--jdk`, `--lts-only`
-   🛡️ Supporto a repository privati con configurazione:
    -   Via `~/.jvm/config.json`
    -   Via variabili d'ambiente `JVM_PRIVATE_ENDPOINT` e `JVM_PRIVATE_TOKEN`
-   📦 Comandi ausiliari:
    -   `configure-private`: genera `config.json`
    -   `config-show`: visualizza configurazione attuale
    -   `config-reset`: cancella configurazione privata-platform per JDK multiple su provider pubblici e privati.

## 🔧 Esempi d'uso

```bash
# Comandi completi
jvm remote-list                    # selezione smart per Adoptium
jvm remote-list --provider=azul   # provider alternativo
jvm remote-list --all             # mostra versioni da tutti i provider
jvm remote-list --provider=private  # fetch repository aziendale

jvm configure-private <URL> <TOKEN>  # crea configurazione privata
jvm config-show                     # visualizza configurazione
jvm config-reset                    # rimuovi configurazione

# Comandi abbreviati (più veloci)
jvm rl                            # equivalente a remote-list
jvm rl --provider=azul --jdk=21   # remote-list con parametri
jvm cp <URL> <TOKEN>              # configure-private
jvm cs                            # config-show
jvm cr                            # config-reset

# Help
jvm --help                        # mostra tutti i comandi disponibili
jvm -h                            # alias per --help
```

Manager intelligente per esplorare, filtrare e gestire versioni di OpenJDK da provider pubblici e repository privati — con interfaccia grafica testuale e comportamento smart.

---

## 🚀 Funzionalità principali

-   🔍 Elenco JDK da **Adoptium**, **Azul**, **Liberica** e **repository privati**
-   🧠 Selezione intelligente di una versione per tag (LTS → usata → patch)
-   📊 Visualizzazione tabellare con info su OS, Architettura, link di download
-   📄 Banner Figlet all’avvio + spiegazione del comportamento
-   🎛️ Supporto a flag avanzati:
    -   `--provider`, `--all`, `--major-only`, `--latest`, `--jdk`, `--lts-only`
-   🛡️ Supporto a repository privati con configurazione:
    -   Via `~/.jvm/config.json`
    -   Via variabili d’ambiente `JVM_PRIVATE_ENDPOINT` e `JVM_PRIVATE_TOKEN`
-   📦 Comandi ausiliari:
    -   `configure-private`: genera `config.json`
    -   `config-show`: visualizza configurazione attuale
    -   `config-reset`: cancella configurazione privata

---

## 📦 Installazione

### Windows (.exe globale)

1. Scarica l’installer firmato `jvm-installer.exe`
2. Esegui come amministratore
3. Il tool sarà disponibile in qualsiasi terminale come `jvm`

---

## 🔧 Esempi d’uso

```bash
jvm remote-list                    # selezione smart per Adoptium
jvm remote-list --provider=azul   # provider alternativo
jvm remote-list --all             # mostra versioni da tutti i provider
jvm remote-list --provider=private  # fetch repository aziendale

jvm configure-private <URL> <TOKEN>  # crea configurazione privata
jvm config-show
jvm config-reset
```

---

## 📒 VADEMECUM — Modifiche, Build e Distribuzione

### 🔧 Quando modifichi il codice Go

1. ✅ **Modifica i file** in `cmd/`, `providers/`, `utils/`
2. 🧪 **Testa la CLI** con:
    ```bash
    go run main.go remote-list
    ```
3. **Generazione Installer**
    ```bash
    GOOS=windows GOARCH=amd64 go build -o distribution/jvm.exe main.go
    ```

# 📦 Build automatizzato

Usa build.bat (Windows CMD) o build.sh (Bash) per:

-   ✅ Compilare jvm.exe in distribution/

-   🔐 Firmare con jvm-dev-cert.pfx (autofirmato)

-   📦 Compilare installer .exe via Inno Setup

Esegui dalla root del progetto:

```bash
# Per CMD/PowerShell
build.bat

# Per Bash (Git Bash, WSL, etc.)
./build.sh
```

### ⚠️ Risoluzione problemi di build

Se ricevi errore **"EndUpdateResource failed (110)"** o **"Il file è utilizzato da un altro processo"**:

1. **Chiudi VS Code** completamente
2. **Chiudi tutti i terminali** aperti nella cartella del progetto
3. **Attendi 10 secondi** e riprova il build
4. Se il problema persiste, **riavvia il sistema**

Questo errore si verifica quando Windows mantiene un handle sui file appena compilati.

🧠 Priorità selezione versioni
✅ Versione LTS

📈 Versione più usata

🆕 Patch più recente

📎 Requisiti
Go 1.20+

Inno Setup Compiler

Windows SDK con signtool.exe

Certificato autofirmato .pfx (facoltativo)

🖋️ Creato da Marco Antonio Russo — powered by JVM CLI 💎

### NOTE

-   Installer rotto
-   pulizia path variables
